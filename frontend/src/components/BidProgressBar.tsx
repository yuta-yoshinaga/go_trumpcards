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
  const segmentClass = nilBroken ? 'bg-ds-error' : isMet ? 'bg-ds-success motion-safe:animate-pulse' : 'bg-ds-warning';

  return (
    <div
      data-testid={testId ?? 'bid-progress-bar'}
      className="mt-0.5 flex items-center gap-1"
      role="progressbar"
      aria-label={ariaLabel}
      aria-valuemin={0}
      aria-valuemax={bid > 0 ? bid : 1}
      aria-valuenow={Math.min(filledSegments, bid > 0 ? bid : 1)}
    >
      <div className="flex flex-1 gap-0.5">
        {Array.from({ length: totalSegments }, (_, idx) => (
          <div
            key={`bid-seg-${idx.toString()}`}
            className={`h-1.5 flex-1 rounded-sm ${idx < filledSegments || nilBroken ? segmentClass : 'bg-white/15'}`}
          />
        ))}
      </div>
      {overTricks > 0 && !nilBroken && (
        <span className="rounded-full bg-ds-info/30 px-1.5 py-0 text-[10px] font-semibold leading-tight text-ds-info">
          +{overTricks}
        </span>
      )}
    </div>
  );
}
