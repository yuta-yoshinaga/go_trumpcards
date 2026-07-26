import { useTranslation } from 'react-i18next';

/** One player's line in a round-result table. */
export interface RoundResultEntry {
  playerIdx: number;
  handName: string;
  kickers?: string;
  wonAmount: number;
  mucked?: boolean;
}

/** Props for {@link RoundResults}. */
export interface RoundResultsProps {
  results: RoundResultEntry[] | undefined;
  players: { isHuman: boolean }[];
}

/**
 * Renders round result summary showing each player's hand and winnings.
 *
 * Also renders a sibling sr-only live region (`role="status"` /
 * `aria-live="polite"`) so screen-reader users hear the showdown breakdown
 * (opponent hands, kickers, chips won) without having to navigate back to
 * the visible table after each hand.
 */
export function RoundResults({ results, players }: RoundResultsProps) {
  const { t } = useTranslation('common');
  if (!results || results.length === 0) return null;

  const announcement = results
    .map((r) => {
      const name = players[r.playerIdx]?.isHuman ? t('player.you') : `CPU ${r.playerIdx}`;
      if (r.mucked) {
        return t('roundResultsAnnouncement.entryMucked', { name });
      }
      const handName = r.handName || '';
      const hasKickers = Boolean(r.kickers);
      const hasWon = r.wonAmount > 0;
      if (hasKickers && hasWon) {
        return t('roundResultsAnnouncement.entryHandWithKickersWon', {
          name,
          handName,
          kickers: r.kickers,
          amount: r.wonAmount,
        });
      }
      if (hasKickers) {
        return t('roundResultsAnnouncement.entryHandWithKickers', {
          name,
          handName,
          kickers: r.kickers,
        });
      }
      if (hasWon) {
        return t('roundResultsAnnouncement.entryHandWon', {
          name,
          handName,
          amount: r.wonAmount,
        });
      }
      return t('roundResultsAnnouncement.entryHand', { name, handName });
    })
    .join(', ');

  return (
    <>
      <div data-testid="round-results-visible" className="bg-black/30 rounded p-2 mb-3 text-white text-xs">
        <div className="font-bold mb-1">{t('label.result')}</div>
        {results.map((r) => (
          <div key={r.playerIdx}>
            {players[r.playerIdx]?.isHuman ? t('player.you') : `CPU ${r.playerIdx}`}
            {r.mucked ? `: ${t('label.mucked')}` : r.handName && `: ${r.handName}`}
            {!r.mucked && r.kickers && ` (${t('label.kicker', { kickers: r.kickers })})`}
            {r.wonAmount > 0 && (
              <span className="text-ds-warning ml-1"> {t('label.chipsWon', { amount: r.wonAmount })}</span>
            )}
          </div>
        ))}
      </div>
      <div role="status" aria-live="polite" aria-atomic="true" className="sr-only">
        {t('roundResultsAnnouncement.message', { details: announcement })}
      </div>
    </>
  );
}
