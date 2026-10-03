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
  const roundResults = results ?? [];

  // wonSpade is the spade half alone; wonAmount is high + spade, so the high
  // half is the difference. Reading wonAmount as "the high" would double-count
  // a scoop.
  const hiWinners = roundResults.flatMap((r) => {
    const hi = r.wonAmount - (r.wonSpade ?? 0);
    return hi > 0 ? [{ name: findPlayerName(players, r.playerIdx), amount: hi }] : [];
  });
  const spadeWinners = roundResults.flatMap((r) =>
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

  // A scoop is one player taking BOTH halves -- the best outcome in the game,
  // so it gets its own line rather than being read off two badges.
  const scoopers = roundResults.flatMap((r) =>
    r.wonSpade && r.wonAmount - r.wonSpade > 0
      ? [{ name: findPlayerName(players, r.playerIdx), total: r.wonAmount }]
      : [],
  );

  return (
    <div role="status" aria-live="polite" aria-atomic="true" data-testid="studchicago-live-region">
      {hiWinners.length > 0 || spadeWinners.length > 0 ? (
        <div className="mb-2 text-center text-sm" data-testid="studchicago-split">
          <div className="mb-1 text-ds-text-muted">{t('split.title')}</div>
          {scoopers.length > 0 && (
            <div className="mb-1.5 flex flex-wrap justify-center gap-2">
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
      ) : null}
    </div>
  );
}
