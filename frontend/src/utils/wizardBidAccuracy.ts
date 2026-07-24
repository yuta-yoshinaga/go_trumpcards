/** Bid-vs-actual outcome bucket for one player's just-completed round. */
export type WizardBidOutcome = 'made' | 'over' | 'under';

/** One player's bid-vs-actual accuracy summary for the completed round. */
export interface WizardBidAccuracyEntry {
  name: string;
  bid: number;
  trickCount: number;
  /** Actual tricks minus bid: 0 = exact, positive = overshot, negative = undershot. */
  delta: number;
  outcome: WizardBidOutcome;
}

/** Minimal player shape needed to compute a bid-accuracy entry. */
export interface WizardBidAccuracyPlayer {
  name: string;
  bid: number;
  trickCount: number;
}

/**
 * Summarizes how each player's actual tricks won compares to their bid for the
 * just-completed Wizard round. Players who never placed a bid (`bid < 0`) are
 * skipped, since there is no contract to measure against.
 *
 * @param players - Players carrying the round's `bid` and `trickCount`.
 * @returns One entry per bidding player: the signed `delta` (tricks − bid) and
 *   whether they made their contract exactly, overshot, or undershot it.
 */
export function wizardBidAccuracy(players: readonly WizardBidAccuracyPlayer[]): WizardBidAccuracyEntry[] {
  return players
    .filter((p) => p.bid >= 0)
    .map((p) => {
      const delta = p.trickCount - p.bid;
      const outcome: WizardBidOutcome = delta === 0 ? 'made' : delta > 0 ? 'over' : 'under';
      return { name: p.name, bid: p.bid, trickCount: p.trickCount, delta, outcome };
    });
}
