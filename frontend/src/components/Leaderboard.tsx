import { LeaderboardEntry } from '../types';

type Props = {
  title: string;
  entries: LeaderboardEntry[];
  highlight?: string;
};

export function Leaderboard({ title, entries, highlight }: Props) {
  return (
    <div className="glass rounded-2xl p-6">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-white">{title}</h3>
        <span className="text-xs text-white/60">Top {entries.length}</span>
      </div>
      <div className="space-y-3">
        {entries.map((row, idx) => (
          <div
            key={row.id}
            className={`flex items-center justify-between rounded-xl px-3 py-2 border border-white/5 bg-white/5 ${
              row.id === highlight ? 'border-neon/60 shadow-glow' : ''
            }`}
          >
            <div className="flex items-center gap-3">
              <span className="text-sm font-mono text-white/60">#{idx + 1}</span>
              <div>
                <p className="text-sm font-semibold text-white">{row.name}</p>
                <p className="text-xs text-white/60">ELO {row.currentElo}</p>
              </div>
            </div>
            <div className="text-xs text-white/60">W {row.wins} · L {row.losses}</div>
          </div>
        ))}
        {entries.length === 0 && <p className="text-sm text-white/60">No data yet</p>}
      </div>
    </div>
  );
}
