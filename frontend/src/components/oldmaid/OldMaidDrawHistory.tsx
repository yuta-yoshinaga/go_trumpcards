import { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import type { DrawHistoryEntry, OldMaidPlayerData } from '../../types/card';
import { findPlayerName } from '../../utils/playerUtils';

/** Colored chip used as a player icon in the timeline. Deterministic by index. */
function PlayerChip({ name, idx }: { name: string; idx: number }) {
  // Tailwind classes (NOT a template literal) so the JIT compiler can pick them up.
  const palette = ['bg-blue-600', 'bg-emerald-600', 'bg-amber-600', 'bg-purple-600', 'bg-pink-600', 'bg-cyan-600'];
  const colorClass = palette[idx % palette.length];
  return (
    <span
      className={`inline-flex items-center justify-center rounded-full text-white text-[10px] font-bold px-1.5 py-0.5 min-w-[2.25rem] ${colorClass}`}
      data-player-idx={idx}
    >
      {name}
    </span>
  );
}

/**
 * Renders a graphical timeline of draw history entries for Old Maid (#1889).
 * Each entry shows `[drawer chip] → [target chip]` with a discard burst when a
 * pair was discarded. When the target is flagged as a "suspect" (the player
 * suspected of holding the Old Maid), the arrow turns red so the player can
 * track at a glance who the dangerous card has moved between.
 */
export function OldMaidDrawHistory({
  entries,
  players,
  suspectPins,
}: {
  entries: DrawHistoryEntry[];
  players: OldMaidPlayerData[];
  /** Player ids currently pinned as suspects. Optional for backwards compat. */
  suspectPins?: ReadonlySet<number>;
}) {
  const { t } = useTranslation('oldmaid');
  const scrollRef = useRef<HTMLDivElement | null>(null);
  // biome-ignore lint/correctness/useExhaustiveDependencies: entries triggers scroll on new history entries
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [entries]);

  if (entries.length === 0) return null;

  return (
    <div className="bg-black/50 rounded-lg my-2 p-2" data-testid="draw-history-timeline">
      <div className="text-ds-text-primary font-bold text-xs mb-1">{t('history.title')}</div>
      <div ref={scrollRef} className="max-h-[140px] overflow-y-auto text-xs text-game-text-muted space-y-1">
        {entries.map((entry, i) => {
          const fromPlayer = players[entry.drawPlayerIdx];
          const targetPlayer = players[entry.drawFromIdx];
          const fromName = findPlayerName(players, entry.drawPlayerIdx);
          const targetName = findPlayerName(players, entry.drawFromIdx);
          const targetIsSuspect = suspectPins?.has(targetPlayer?.id ?? -1) ?? false;
          const arrowClass = targetIsSuspect ? 'text-ds-error font-bold' : 'text-ds-text-muted';
          return (
            <div
              key={i}
              className="flex flex-wrap items-center gap-1.5"
              data-testid="draw-history-entry"
              data-suspect-target={targetIsSuspect ? 'true' : 'false'}
            >
              <span className="tabular-nums text-ds-text-muted">{i + 1}.</span>
              <PlayerChip name={fromName} idx={fromPlayer?.id ?? entry.drawPlayerIdx} />
              <span aria-hidden="true" className={arrowClass}>
                ➔
              </span>
              <PlayerChip name={targetName} idx={targetPlayer?.id ?? entry.drawFromIdx} />
              {entry.discardedPairs > 0 && (
                <span
                  className="text-ds-warning text-sm"
                  aria-label={t('history.discarded', { count: entry.discardedPairs })}
                  data-testid="discard-burst"
                >
                  💥
                </span>
              )}
              {entry.drawerFinished && (
                <span className="text-ds-success text-[10px]">{t('history.finished', { name: fromName })}</span>
              )}
              {entry.targetFinished && (
                <span className="text-ds-success text-[10px]">{t('history.finished', { name: targetName })}</span>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
