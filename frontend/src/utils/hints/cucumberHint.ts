import type { CucumberResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Cucumber, or null when no
 * suggestion is available.
 *
 * When you cannot beat the trick the play is forced, so that advice is
 * certain; which high card to spend is a judgement call.
 */
export function getCucumberHint(state: CucumberResponse): HintResult | null {
  const hint = state.hint;
  if (!hint || hint.cardIndex === undefined) return null;

  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: hint.reason === 'cucumberForced' ? 'strong' : 'moderate',
  };
}
