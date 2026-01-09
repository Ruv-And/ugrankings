import { Artist, LeaderboardEntry, MatchupResponse } from './types';

const API_BASE = import.meta.env.VITE_API_BASE ?? '';

export async function fetchMatchup(signal?: AbortSignal): Promise<MatchupResponse> {
  const res = await fetch(`${API_BASE}/api/matchup`, { signal });
  if (!res.ok) throw new Error('Failed to load matchup');
  return res.json();
}

export async function submitVote(winnerId: string, loserId: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/vote`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ winner_id: winnerId, loser_id: loserId })
  });
  if (!res.ok) throw new Error('Vote failed');
}

export async function fetchLeaderboard(type: string): Promise<LeaderboardEntry[]> {
  const res = await fetch(`${API_BASE}/api/leaderboard?type=${type}`);
  if (!res.ok) throw new Error('Failed to load leaderboard');
  return res.json();
}

export async function fetchTrending(period: string): Promise<LeaderboardEntry[]> {
  const res = await fetch(`${API_BASE}/api/trending?period=${period}`);
  if (!res.ok) throw new Error('Failed to load trending');
  return res.json();
}

export type Matchup = {
  artists: [Artist, Artist];
};
