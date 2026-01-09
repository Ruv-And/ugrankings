export type Artist = {
  id: string;
  name: string;
  spotifyUrl: string;
  previewUrl?: string;
  currentElo: number;
  genres: string[];
  wins: number;
  losses: number;
};

export type MatchupResponse = {
  artists: Artist[];
};

export type LeaderboardEntry = {
  id: string;
  name: string;
  currentElo: number;
  wins: number;
  losses: number;
};
