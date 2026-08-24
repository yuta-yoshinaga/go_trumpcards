import { useTranslation } from 'react-i18next';
import type { SevenCardStudPlayerData, SevenCardStudResult } from '../types/card';
import { cardAlt } from '../utils/cardAlt';
import { findPlayerName } from '../utils/playerUtils';

/** Props for {@link StudChicagoSplit}. */
export interface StudChicagoSplitProps {
  results: SevenCardStudResult[] | undefined;
  players: SevenCardStudPlayerData[];
}

/**
 * Renders the high / spade breakdown of a Chicago showdown.
 *
 * Without this the pot silently halves and nothing on screen explains why. The
 * spade half is decided by a **single face-down card**, so the card itself is
 * named rather than left implicit — that one card is the whole reason the other
 * half went where it did.
 */
export function StudChicagoSplit({ results, players }: StudChicagoSplitProps) {
  const { t } = useTranslation('chicago');
  if (!results || results.length === 0) return null;

  // wonSpade is the spade half alone; wonAmount is high + spade, so the high
  // half is the difference. Reading wonAmount as "the high" would double-count
  // a scoop.
  const hiWinners = results.flatMap((r) => {
    const hi = r.wonAmount - (r.wonSpade ?? 0);
    return hi > 0 ? [{ name: findPlayerName(players, r.playerIdx), amount: hi }] : [];
  });
  const spadeWinners = results.flatMap((r) =>
    r.wonSpade
      ? [
          {
            name: findPlayerName(players, r.playerIdx),
            amount: r.wonSpade,
            card: r.spadeCard ? cardAlt(r.spadeCard) : '',
          },
        ]
      : [],
  );
  if (hiWinners.length === 0 && spadeWinners.length === 0) return null;

  // A scoop is one player taking BOTH halves -- the best outcome in the game,
  // so it gets its own line rather than being read off two badges.
  const scoopers = results.flatMap((r) =>
    r.wonSpade && r.wonAmount - r.wonSpade > 0
      ? [{ name: findPlayerName(players, r.playerIdx), total: r.wonAmount }]
      : [],
  );

  return (
    <div className="mb-2 text-center text-sm" data-testid="studchicago-split">
      <div className="mb-1 text-ds-text-muted">{t('split.title')}</div>
      {scoopers.length > 0 && (
        <div className="mb-1.5 flex flex-wrap justify-center gap-2" role="status">
          {scoopers.map((s) => (
            <span
              key={`scoop-${s.name}`}
              data-testid="studchicago-scoop-badge"
              className="inline-block rounded-full border border-ds-accent bg-ds-accent px-3 py-0.5 font-bold text-ds-text-on-accent"
            >
              {t('split.scoop', { name: s.name, total: s.total })}
            </span>
          ))}
        </div>
      )}
      <div className="flex flex-wrap justify-center gap-2">
        {hiWinners.map((w) => (
          <span
            key={`hi-${w.name}`}
            data-testid="studchicago-hi-badge"
            className="inline-block rounded border border-ds-success bg-ds-surface px-2 py-0.5 text-ds-success"
          >
            {t('split.hi')}: {t('split.winner', { name: w.name, amount: w.amount })}
          </span>
        ))}
        {spadeWinners.map((w) => (
          <span
            key={`spade-${w.name}`}
            data-testid="studchicago-spade-badge"
            className="inline-block rounded border border-ds-info bg-ds-surface px-2 py-0.5 text-ds-info"
          >
            {t('split.spade')}: {t('split.winner', { name: w.name, amount: w.amount })}
            {w.card && ` (${w.card})`}
          </span>
        ))}
      </div>
      {spadeWinners.length === 0 && hiWinners.length > 0 && (
        <div className="mt-1 text-xs text-ds-text-muted" data-testid="studchicago-hi-takes-all">
          {t('split.hiTakesAll')}
        </div>
      )}
    </div>
  );
}
