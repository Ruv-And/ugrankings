package main

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

type Artist struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	SpotifyURL string   `json:"spotifyUrl"`
	PreviewURL string   `json:"previewUrl,omitempty"`
	CurrentElo int      `json:"currentElo"`
	Genres     []string `json:"genres"`
	Wins       int      `json:"wins"`
	Losses     int      `json:"losses"`
}

type CassandraStore struct {
	session *gocql.Session
	kafka   *KafkaProducer
	mu      sync.RWMutex
}

func newCassandraStore(session *gocql.Session, kafka *KafkaProducer) *CassandraStore {
	store := &CassandraStore{
		session: session,
		kafka:   kafka,
	}

	// Seed initial data if empty
	store.seedData()
	return store
}

func (s *CassandraStore) seedData() {
	var count int
	if err := s.session.Query("SELECT COUNT(*) FROM artists").Scan(&count); err != nil || count > 0 {
		return
	}

	seedArtists := []Artist{
		{ID: uuid.New().String(), Name: "Nia Noir", SpotifyURL: "https://open.spotify.com/artist/a1", CurrentElo: 1500, Genres: []string{"alt-rap", "neo-soul"}},
		{ID: uuid.New().String(), Name: "Basement Prophet", SpotifyURL: "https://open.spotify.com/artist/a2", CurrentElo: 1480, Genres: []string{"underground", "boom-bap"}},
		{ID: uuid.New().String(), Name: "Lo-Fi Griot", SpotifyURL: "https://open.spotify.com/artist/a3", CurrentElo: 1520, Genres: []string{"lofi", "storytelling"}},
		{ID: uuid.New().String(), Name: "Vanta", SpotifyURL: "https://open.spotify.com/artist/a4", CurrentElo: 1495, Genres: []string{"trap", "darkwave"}},
		{ID: uuid.New().String(), Name: "Southside Sage", SpotifyURL: "https://open.spotify.com/artist/a5", CurrentElo: 1510, Genres: []string{"southern", "grit"}},
	}

	for _, artist := range seedArtists {
		id, _ := uuid.Parse(artist.ID)
		s.session.Query(`
			INSERT INTO artists (artist_id, name, spotify_url, current_elo, total_votes, wins, losses, genres, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, id, artist.Name, artist.SpotifyURL, artist.CurrentElo, 0, 0, 0, artist.Genres, time.Now(), time.Now()).Exec()
	}
}

func (s *CassandraStore) matchup() ([2]Artist, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var artists []Artist
	iter := s.session.Query("SELECT artist_id, name, spotify_url, current_elo, wins, losses, genres FROM artists").Iter()

	var id gocql.UUID
	var name, spotifyURL string
	var currentElo, wins, losses int
	var genres []string

	for iter.Scan(&id, &name, &spotifyURL, &currentElo, &wins, &losses, &genres) {
		artists = append(artists, Artist{
			ID:         id.String(),
			Name:       name,
			SpotifyURL: spotifyURL,
			CurrentElo: currentElo,
			Wins:       wins,
			Losses:     losses,
			Genres:     genres,
		})
	}

	if err := iter.Close(); err != nil {
		return [2]Artist{}, err
	}

	if len(artists) < 2 {
		return [2]Artist{}, nil
	}

	// Smart pairing by ELO
	rand.Seed(time.Now().UnixNano())
	first := artists[rand.Intn(len(artists))]

	var second Artist
	bestDelta := math.MaxInt
	for _, a := range artists {
		if a.ID == first.ID {
			continue
		}
		delta := int(math.Abs(float64(a.CurrentElo - first.CurrentElo)))
		if delta < bestDelta {
			bestDelta = delta
			second = a
		}
	}

	if second.ID == "" {
		second = artists[(rand.Intn(len(artists)-1)+1)%len(artists)]
	}

	// Randomize order
	if rand.Intn(2) == 0 {
		return [2]Artist{first, second}, nil
	}
	return [2]Artist{second, first}, nil
}

func (s *CassandraStore) vote(winnerID, loserID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get current ELOs for event
	var winnerElo, loserElo int
	winnerUUID, _ := uuid.Parse(winnerID)
	loserUUID, _ := uuid.Parse(loserID)

	if err := s.session.Query("SELECT current_elo FROM artists WHERE artist_id = ?", winnerUUID).Scan(&winnerElo); err != nil {
		return err
	}
	if err := s.session.Query("SELECT current_elo FROM artists WHERE artist_id = ?", loserUUID).Scan(&loserElo); err != nil {
		return err
	}

	// Store vote in Cassandra
	voteID := gocql.TimeUUID()
	today := time.Now().Truncate(24 * time.Hour)

	if err := s.session.Query(`
		INSERT INTO votes (vote_date, vote_id, winner_id, loser_id, timestamp)
		VALUES (?, ?, ?, ?, ?)
	`, today, voteID, winnerUUID, loserUUID, time.Now()).Exec(); err != nil {
		return err
	}

	// Publish to Kafka for async ELO processing
	event := VoteEvent{
		VoteID:          voteID.String(),
		WinnerID:        winnerID,
		LoserID:         loserID,
		Timestamp:       time.Now(),
		WinnerEloBefore: winnerElo,
		LoserEloBefore:  loserElo,
	}

	return s.kafka.PublishVote(event)
}

func (s *CassandraStore) ProcessVoteEvent(event VoteEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	winnerUUID, _ := uuid.Parse(event.WinnerID)
	loserUUID, _ := uuid.Parse(event.LoserID)

	// Calculate new ELOs
	newWinnerElo, newLoserElo := calculateElo(event.WinnerEloBefore, event.LoserEloBefore, 32)

	// Update both artists atomically
	batch := s.session.NewBatch(gocql.LoggedBatch)

	batch.Query(`
		UPDATE artists SET current_elo = ?, wins = wins + 1, total_votes = total_votes + 1, updated_at = ?
		WHERE artist_id = ?
	`, newWinnerElo, time.Now(), winnerUUID)

	batch.Query(`
		UPDATE artists SET current_elo = ?, losses = losses + 1, total_votes = total_votes + 1, updated_at = ?
		WHERE artist_id = ?
	`, newLoserElo, time.Now(), loserUUID)

	return s.session.ExecuteBatch(batch)
}

func (s *CassandraStore) leaderboard(kind string) ([]Artist, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := "SELECT artist_id, name, current_elo, wins, losses FROM leaderboard_by_elo"
	if kind == "velocity" {
		// For velocity, order by win difference
		query = "SELECT artist_id, name, current_elo, wins, losses FROM artists"
	}

	iter := s.session.Query(query).Iter()
	var artists []Artist
	var id gocql.UUID
	var name string
	var currentElo, wins, losses int

	for iter.Scan(&id, &name, &currentElo, &wins, &losses) {
		artists = append(artists, Artist{
			ID:         id.String(),
			Name:       name,
			CurrentElo: currentElo,
			Wins:       wins,
			Losses:     losses,
		})
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	// Limit to top 20
	if len(artists) > 20 {
		artists = artists[:20]
	}

	return artists, nil
}

func expectedScore(ratingA, ratingB int) float64 {
	return 1.0 / (1.0 + math.Pow(10, float64(ratingB-ratingA)/400.0))
}

func calculateElo(winnerElo, loserElo, k int) (int, int) {
	expectWin := expectedScore(winnerElo, loserElo)
	expectLose := expectedScore(loserElo, winnerElo)

	newWinnerElo := int(float64(winnerElo) + float64(k)*(1.0-expectWin))
	newLoserElo := int(float64(loserElo) + float64(k)*(0.0-expectLose))

	return newWinnerElo, newLoserElo
}
