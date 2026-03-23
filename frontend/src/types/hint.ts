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
}
