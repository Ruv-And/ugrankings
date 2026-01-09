package main

import (
	"log"
	"time"

	"github.com/gocql/gocql"
)

func setupCassandra() (*gocql.Session, error) {
	cluster := gocql.NewCluster("127.0.0.1:9042")
	cluster.Keyspace = "ugrankings"
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 10 * time.Second

	session, err := cluster.CreateSession()
	if err != nil {
		// Try creating keyspace first
		tempCluster := gocql.NewCluster("127.0.0.1:9042")
		tempCluster.Consistency = gocql.Quorum
		tempSession, tempErr := tempCluster.CreateSession()
		if tempErr != nil {
			return nil, err
		}
		defer tempSession.Close()

		if err := tempSession.Query(`
			CREATE KEYSPACE IF NOT EXISTS ugrankings
			WITH REPLICATION = {
				'class': 'SimpleStrategy',
				'replication_factor': 1
			}
		`).Exec(); err != nil {
			return nil, err
		}

		log.Println("Created keyspace ugrankings")

		// Retry with keyspace
		session, err = cluster.CreateSession()
		if err != nil {
			return nil, err
		}
	}

	// Create tables
	if err := createTables(session); err != nil {
		session.Close()
		return nil, err
	}

	log.Println("Connected to Cassandra")
	return session, nil
}

func createTables(session *gocql.Session) error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS artists (
			artist_id uuid PRIMARY KEY,
			name text,
			spotify_url text,
			current_elo int,
			total_votes int,
			wins int,
			losses int,
			genres set<text>,
			created_at timestamp,
			updated_at timestamp
		)`,
		`CREATE TABLE IF NOT EXISTS votes (
			vote_date date,
			vote_id timeuuid,
			winner_id uuid,
			loser_id uuid,
			timestamp timestamp,
			PRIMARY KEY ((vote_date), vote_id)
		) WITH CLUSTERING ORDER BY (vote_id DESC)`,
		`CREATE MATERIALIZED VIEW IF NOT EXISTS leaderboard_by_elo AS
			SELECT artist_id, name, current_elo, total_votes, wins, losses
			FROM artists
			WHERE current_elo IS NOT NULL AND artist_id IS NOT NULL
			PRIMARY KEY (current_elo, artist_id)
			WITH CLUSTERING ORDER BY (current_elo DESC)`,
	}

	for _, query := range tables {
		if err := session.Query(query).Exec(); err != nil {
			log.Printf("Failed to create table: %v", err)
			return err
		}
	}

	log.Println("Created Cassandra tables")
	return nil
}
