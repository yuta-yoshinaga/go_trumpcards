import { useTranslation } from 'react-i18next';
import type { SevenCardStudPlayerData, SevenCardStudResult } from '../types/card';
import { valueName } from '../utils/cardUtils';
import { findPlayerName } from '../utils/playerUtils';

/** Props for {@link StudHiLoSplit}. */
export interface StudHiLoSplitProps {
  results: SevenCardStudResult[] | undefined;
  players: SevenCardStudPlayerData[];
}

/**
 * Renders the Hi/Lo breakdown of a Seven Card Stud Hi-Lo showdown.
 *
 * Without this the pot silently halves and nothing on screen explains why. The
 * "high takes it all" line is the point of eight-or-better, so it is stated
 * explicitly rather than left as the absence of a low badge.
 */
export function StudHiLoSplit({ results, players }: StudHiLoSplitProps) {
  const { t } = useTranslation('sevencardstudhilo');
  if (!results || results.length === 0) return null;

  // wonLow is the low half alone; wonAmount is high + low, so the high half is
  // the difference. Reading wonAmount as "the high" would double-count a scoop.
  const hiWinners = results.flatMap((r) => {
    const hi = r.wonAmount - (r.wonLow ?? 0);
    return hi > 0 ? [{ name: findPlayerName(players, r.playerIdx), amount: hi }] : [];
  });
  const loWinners = results.flatMap((r) =>
    r.wonLow
      ? [
          {
            name: findPlayerName(players, r.playerIdx),
            amount: r.wonLow,
            cards: (r.lowBestHand ?? []).map((card) => valueName(card.value)).join(' '),
          },
        ]
      : [],
  );
  if (hiWinners.length === 0 && loWinners.length === 0) return null;

  // A scoop is one player taking BOTH halves -- the best outcome in the game,
  // so it gets its own line rather than being read off two badges.
  const scoopers = results.flatMap((r) =>
    r.wonLow && r.wonAmount - r.wonLow > 0 ? [{ name: findPlayerName(players, r.playerIdx), total: r.wonAmount }] : [],
  );

  return (
    <div className="mb-2 text-center text-sm" data-testid="studhilo-split">
      <div className="mb-1 text-ds-text-muted">{t('hiLo.title')}</div>
      {scoopers.length > 0 && (
        <div className="mb-1.5 flex flex-wrap justify-center gap-2" role="status">
          {scoopers.map((s) => (
            <span
              key={`scoop-${s.name}`}
              data-testid="studhilo-scoop-badge"
              className="inline-block rounded-full border border-ds-accent bg-ds-accent px-3 py-0.5 font-bold text-ds-text-on-accent"
            >
              {t('hiLo.scoop', { name: s.name, total: s.total })}
            </span>
          ))}
        </div>
      )}
      <div className="flex flex-wrap justify-center gap-2">
        {hiWinners.map((w) => (
          <span
            key={`hi-${w.name}`}
            data-testid="studhilo-hi-badge"
            className="inline-block rounded border border-ds-success bg-ds-surface px-2 py-0.5 text-ds-success"
          >
            {t('hiLo.hi')}: {t('hiLo.winner', { name: w.name, amount: w.amount })}
          </span>
        ))}
        {loWinners.map((w) => (
          <span
            key={`lo-${w.name}`}
            data-testid="studhilo-lo-badge"
            className="inline-block rounded border border-ds-info bg-ds-surface px-2 py-0.5 text-ds-info"
          >
            {t('hiLo.lo')}: {t('hiLo.winner', { name: w.name, amount: w.amount })}
            {w.cards && ` (${w.cards})`}
          </span>
        ))}
      </div>
      {loWinners.length === 0 && hiWinners.length > 0 && (
        <div className="mt-1 text-xs text-ds-text-muted" data-testid="studhilo-hi-takes-all">
          {t('hiLo.hiTakesAll')}
        </div>
      )}
    </div>
  );
}
