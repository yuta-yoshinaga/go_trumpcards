/** Confidence level of a hint recommendation. */
export type HintConfidence = 'strong' | 'moderate';

/** Result of a hint evaluation for a game action. */
export interface HintResult {
  /** Action identifier matching a button's data-hint-action attribute. */
  targetAction: string;
  /** i18n key for the reasoning text. */
  reason: string;
  /** Confidence level of the recommendation. */
  confidence: HintConfidence;
  /**
   * Board position the advice points at, when the hint knows one.
   *
   * A hint that says "swap with your highest card" has already worked out
   * *which* card it means; dropping the index leaves the player to redo that
   * search by eye (#4887). Pages use it to highlight the slot.
   */
  targetPos?: number;
}
