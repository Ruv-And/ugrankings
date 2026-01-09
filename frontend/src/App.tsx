import { useState, useEffect } from 'react';
import { fetchLeaderboard, fetchMatchup, submitVote } from './api';
import { Artist, LeaderboardEntry } from './types';

type Tab = 'rankings' | 'match' | 'suggest';

export default function App() {
  const [activeTab, setActiveTab] = useState<Tab>('rankings');
  
  return (
    <div className="min-h-screen bg-black text-white">
      <div className="max-w-4xl mx-auto p-6">
        <header className="mb-8">
          <h1 className="text-3xl font-bold mb-2">Underground Rap Rankings</h1>
          <p className="text-gray-400">Community-driven artist rankings</p>
        </header>

        <nav className="border-b border-gray-700 mb-6">
          <div className="flex space-x-0">
            {[
              { id: 'rankings' as Tab, label: 'Rankings' },
              { id: 'match' as Tab, label: 'Vote' },
              { id: 'suggest' as Tab, label: 'Suggest' }
            ].map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`tab ${activeTab === tab.id ? 'tab-active' : 'tab-inactive'}`}
              >
                {tab.label}
              </button>
            ))}
          </div>
        </nav>

        {activeTab === 'rankings' && <RankingsTab />}
        {activeTab === 'match' && <MatchTab />}
        {activeTab === 'suggest' && <SuggestTab />}
      </div>
    </div>
  );
}

function RankingsTab() {
  const [rankings, setRankings] = useState<LeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchLeaderboard('overall')
      .then(setRankings)
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div>Loading rankings...</div>;

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Top Artists</h2>
      <div className="space-y-3">
        {rankings.map((artist, index) => (
          <div key={artist.id} className="card flex justify-between items-center">
            <div className="flex items-center space-x-4">
              <span className="text-gray-400 w-8">#{index + 1}</span>
              <div>
                <h3 className="font-medium">{artist.name}</h3>
                <p className="text-sm text-gray-400">ELO: {artist.currentElo}</p>
              </div>
            </div>
            <div className="text-right">
              <p className="text-sm">Wins: {artist.wins}</p>
              <p className="text-sm text-gray-400">Losses: {artist.losses}</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function MatchTab() {
  const [matchup, setMatchup] = useState<Artist[]>([]);
  const [loading, setLoading] = useState(false);
  const [voting, setVoting] = useState(false);

  const loadMatchup = async () => {
    setLoading(true);
    try {
      const result = await fetchMatchup();
      setMatchup(result.artists);
    } catch (err) {
      console.error('Failed to load matchup:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleVote = async (winnerId: string) => {
    if (matchup.length < 2 || voting) return;
    
    const loserId = matchup.find(a => a.id !== winnerId)?.id;
    if (!loserId) return;

    setVoting(true);
    try {
      await submitVote(winnerId, loserId);
      await loadMatchup(); // Load next matchup
    } catch (err) {
      console.error('Vote failed:', err);
    } finally {
      setVoting(false);
    }
  };

  useEffect(() => {
    loadMatchup();
  }, []);

  if (loading) return <div>Loading matchup...</div>;

  if (matchup.length < 2) {
    return (
      <div>
        <h2 className="text-xl font-semibold mb-4">Vote</h2>
        <div className="card">
          <p>No matchup available. <button onClick={loadMatchup} className="btn ml-2">Retry</button></p>
        </div>
      </div>
    );
  }

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Who's Better?</h2>
      <div className="grid md:grid-cols-2 gap-6 mb-6">
        {matchup.map((artist) => (
          <div key={artist.id} className="card">
            <h3 className="text-lg font-medium mb-2">{artist.name}</h3>
            <p className="text-gray-400 mb-2">ELO: {artist.currentElo}</p>
            <p className="text-sm text-gray-400 mb-4">
              {artist.wins}W - {artist.losses}L
            </p>
            {artist.genres.length > 0 && (
              <div className="flex flex-wrap gap-1 mb-4">
                {artist.genres.map((genre) => (
                  <span key={genre} className="text-xs bg-gray-800 px-2 py-1 rounded">
                    {genre}
                  </span>
                ))}
              </div>
            )}
            <button
              onClick={() => handleVote(artist.id)}
              disabled={voting}
              className="btn w-full"
            >
              {voting ? 'Voting...' : 'Vote for this artist'}
            </button>
          </div>
        ))}
      </div>
      <div className="text-center">
        <button onClick={loadMatchup} className="btn-secondary" disabled={loading}>
          Skip this matchup
        </button>
      </div>
    </div>
  );
}

function SuggestTab() {
  const [artistName, setArtistName] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [message, setMessage] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!artistName.trim()) return;

    setSubmitting(true);
    setMessage('');

    try {
      // For now, just simulate submission since backend doesn't have this endpoint
      await new Promise(resolve => setTimeout(resolve, 1000));
      setMessage(`Thanks for suggesting "${artistName}"!`);
      setArtistName('');
    } catch (err) {
      setMessage('Failed to submit suggestion');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Suggest New Artist</h2>
      <div className="card max-w-md">
        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label className="block text-sm font-medium mb-2">
              Artist Name
            </label>
            <input
              type="text"
              value={artistName}
              onChange={(e) => setArtistName(e.target.value)}
              placeholder="Enter artist name"
              className="w-full px-3 py-2 bg-gray-800 border border-gray-600 rounded text-white placeholder-gray-400 focus:outline-none focus:border-white"
              required
            />
          </div>
          <button
            type="submit"
            disabled={submitting || !artistName.trim()}
            className="btn w-full"
          >
            {submitting ? 'Submitting...' : 'Suggest Artist'}
          </button>
        </form>
        {message && (
          <p className="mt-3 text-sm text-gray-400">{message}</p>
        )}
      </div>
    </div>
  );
}
