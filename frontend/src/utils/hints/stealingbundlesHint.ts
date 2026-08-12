import type { StealingBundlesResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Stealing Bundles, or null when no
 * suggestion is available.
 *
 * Taking a whole bundle swings the count by its full size, so that advice is
 * the confident one; capturing off the table and trailing are judgement calls.
 */
export function getStealingBundlesHint(state: StealingBundlesResponse): HintResult | null {
  const hint = state.hint;
  if (!hint || hint.cardIndex === undefined) return null;

  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: hint.reason === 'stealingbundlesSteal' ? 'strong' : 'moderate',
  };
}
