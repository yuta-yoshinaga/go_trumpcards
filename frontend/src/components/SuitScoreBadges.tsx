import type { Card } from '../types/card';
import { fiftyOneBestSuit, fiftyOneSuitScores } from '../utils/fiftyOneSuitScores';

const SUITS = ['SPADE', 'CLOVER', 'HEART', 'DIAMOND'] as const;

/** Renders per-suit Fifty-One score badges for a visible hand. */
export function SuitScoreBadges({
  cards,
  ariaLabel,
  listTestId = 'suit-score-badges',
  badgeTestId = (design) => `suit-badge-${design}`,
}: {
  cards: Card[];
  ariaLabel: string;
  listTestId?: string;
  badgeTestId?: (design: (typeof SUITS)[number]) => string;
}) {
  const suitTotals = fiftyOneSuitScores(cards);
  const bestSuit = fiftyOneBestSuit(suitTotals);

  return (
    <ul
      className="flex justify-center gap-1.5 mb-1.5 text-xs flex-wrap list-none p-0 m-0"
      aria-label={ariaLabel}
      data-testid={listTestId}
    >
      {SUITS.map((d) => {
        const isLeader = d === bestSuit && suitTotals[d] > 0;
        const symbol = d === 'SPADE' ? '♠' : d === 'CLOVER' ? '♣' : d === 'HEART' ? '♥' : '♦';
        const isRed = d === 'HEART' || d === 'DIAMOND';
        const classes = isLeader
          ? 'bg-ds-accent text-ds-text-on-accent border-ds-accent'
          : 'bg-ds-surface text-ds-text border-ds-border';
        return (
          <li
            key={d}
            data-testid={badgeTestId(d)}
            className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full border font-medium ${classes}`}
          >
            <span className={isLeader ? '' : isRed ? 'text-ds-error' : ''}>{symbol}</span>
            <span className="tabular-nums">{suitTotals[d]}</span>
          </li>
        );
      })}
    </ul>
  );
}
