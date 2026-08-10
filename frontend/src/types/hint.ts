/** Confidence level of a hint recommendation. */
export type HintConfidence = 'strong' | 'moderate';

/** Result of a hint evaluation for a game action. */
export interface HintResult {
  /** Action identifier matching a button's data-hint-action attribute. */
  targetAction: string;
  /** i18n key for the reasoning text. */
  reason: string;
  /**
   * Interpolation values for {@link reason}.
   *
   * A hint that already knows *where* the move goes should say so: Nertz's
   * reason was a fixed "pick a destination" string while the CUI printed
   * "Nertz → Foundation 2" from the very same fields (#4885).
   */
  reasonParams?: Record<string, string | number>;
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
  /**
   * Board positions the advice points at, when it means several at once.
   *
   * Anaconda's backend already works out *which* cards to pass or keep; the
   * tooltip said "pass 3 cards (weak hand)" and left the player to pick them
   * (#4851).
   */
  targetIndices?: number[];
}
