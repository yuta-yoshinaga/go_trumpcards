import { badgeInfoColors } from '../styles/badgeStyles';

/** Props for the BidProgressBar component. */
export interface BidProgressBarProps {
  /** Bid amount declared by the player. Negative values render nothing. */
  bid: number;
  /** Tricks the player has actually won so far this round. */
  tricksWon: number;
  /** Test id for the outer container. */
  testId?: string;
  /** ARIA label for the progress bar. */
  ariaLabel?: string;
}

/**
 * Visualizes Call Break bid progress as a segmented bar. Each segment
 * corresponds to one trick toward the bid. State tinting:
 * - Yellow while under bid (working toward target),
 * - Green when bid is exactly met,
 * - Green + blue "+N" pill when over bid (overtricks add fractional score),
 * - Red when bid is 0 and any trick is taken (auto-fail).
 *
 * Returns null for negative/unset bids.
 */
export function BidProgressBar({ bid, tricksWon, testId, ariaLabel }: BidProgressBarProps) {
  if (bid < 0) return null;

  const isNil = bid === 0;
  const nilBroken = isNil && tricksWon > 0;
  const filledSegments = isNil ? 0 : Math.min(bid, tricksWon);
  const totalSegments = isNil ? 1 : bid;
  const overTricks = isNil ? tricksWon : Math.max(0, tricksWon - bid);
  const isMet = !isNil && tricksWon >= bid;
  // aria-valuenow tracks the *broken* nil-bid bar (which fills its single
  // segment red) as fully advanced — otherwise screen readers would read 0/1
  // while the eye sees a full red bar.
  const ariaValueMax = bid > 0 ? bid : 1;
  const ariaValueNow = nilBroken ? 1 : Math.min(filledSegments, ariaValueMax);

  // Conditional segment color: full class strings per branch to keep Tailwind
  // class-string parsers happy and the per-state classes greppable.
  const getSegmentClass = (idx: number): string => {
    const filled = idx < filledSegments || nilBroken;
    if (!filled) return 'h-1.5 flex-1 rounded-sm bg-white/15';
    if (nilBroken) return 'h-1.5 flex-1 rounded-sm bg-ds-error';
    if (isMet) return 'h-1.5 flex-1 rounded-sm bg-ds-success motion-safe:animate-pulse';
    return 'h-1.5 flex-1 rounded-sm bg-ds-warning';
  };

  return (
    <div
      data-testid={testId ?? 'bid-progress-bar'}
      className="mt-0.5 flex items-center gap-1"
      role="progressbar"
      aria-label={ariaLabel}
      aria-valuemin={0}
      aria-valuemax={ariaValueMax}
      aria-valuenow={ariaValueNow}
    >
      <div className="flex flex-1 gap-0.5">
        {Array.from({ length: totalSegments }, (_, idx) => (
          <div key={`bid-seg-${idx.toString()}`} className={getSegmentClass(idx)} />
        ))}
      </div>
      {overTricks > 0 && !nilBroken && (
        <span className={`rounded-full px-1.5 py-0 text-[10px] font-semibold leading-tight ${badgeInfoColors}`}>
          +{overTricks}
        </span>
      )}
    </div>
  );
}
