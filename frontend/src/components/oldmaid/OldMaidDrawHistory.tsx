import { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import type { DrawHistoryEntry, OldMaidPlayerData } from '../../types/card';
import { findPlayerName } from '../../utils/playerUtils';

/** Renders a scrollable timeline of draw history entries for Old Maid. */
export function OldMaidDrawHistory({
  entries,
  players,
}: {
  entries: DrawHistoryEntry[];
  players: OldMaidPlayerData[];
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
      <div className="text-white font-bold text-xs mb-1">{t('history.title')}</div>
      <div ref={scrollRef} className="max-h-[120px] overflow-y-auto text-xs text-game-text-muted">
        {entries.map((entry, i) => {
          const from = findPlayerName(players, entry.drawPlayerIdx);
          const target = findPlayerName(players, entry.drawFromIdx);
          let line = `${i + 1}. ${t('history.entry', { from, target })}`;
          if (entry.discardedPairs > 0) line += ` ${t('history.discarded', { count: entry.discardedPairs })}`;
          if (entry.drawerFinished) line += ` ${t('history.finished', { name: from })}`;
          if (entry.targetFinished) line += ` ${t('history.finished', { name: target })}`;
          return (
            // biome-ignore lint/suspicious/noArrayIndexKey: history entries are append-only with stable order
            <div key={i}>{line}</div>
          );
        })}
      </div>
    </div>
  );
}
