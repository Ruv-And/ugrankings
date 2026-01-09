import { Artist } from '../types';

type Props = {
  artist: Artist;
  onVote: (id: string) => void;
  loading: boolean;
};

export function MatchupCard({ artist, onVote, loading }: Props) {
  return (
    <div className="glass rounded-2xl p-6 flex flex-col gap-4 text-left h-full justify-between">
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="text-sm uppercase tracking-[0.2em] text-white/60">Artist</p>
          <h3 className="text-2xl font-semibold text-white">{artist.name}</h3>
        </div>
        <span className="text-xs font-mono px-3 py-1 rounded-full bg-white/10 text-neon border border-white/10">{artist.currentElo} ELO</span>
      </div>

      {artist.genres.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {artist.genres.map((g) => (
            <span key={g} className="text-xs px-3 py-1 rounded-full bg-white/5 border border-white/10 text-white/70">
              {g}
            </span>
          ))}
        </div>
      )}

      {artist.previewUrl ? (
        <audio className="w-full" controls src={artist.previewUrl} preload="none" />
      ) : (
        <div className="text-sm text-white/60">No preview available</div>
      )}

      <div className="flex items-center justify-between text-white/70 text-xs">
        <span>Wins: {artist.wins}</span>
        <span>Losses: {artist.losses}</span>
      </div>

      <button
        onClick={() => onVote(artist.id)}
        disabled={loading}
        className="btn-primary w-full text-center"
      >
        {loading ? 'Submitting…' : 'Vote'}
      </button>

      <a
        href={artist.spotifyUrl}
        target="_blank"
        rel="noreferrer"
        className="text-sm text-neon hover:underline"
      >
        Open on Spotify
      </a>
    </div>
  );
}
