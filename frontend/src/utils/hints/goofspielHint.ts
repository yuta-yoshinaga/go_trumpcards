import type { GoofspielResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Goofspiel, or null when no
 * suggestion is available.
 *
 * Every bid is a judgement call — the whole game is guessing what the opponent
 * will spend — so no advice here is ever more than moderate.
 */
export function getGoofspielHint(state: GoofspielResponse): HintResult | null {
  const hint = state.hint;
  if (!hint || hint.cardIndex === undefined) return null;

  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
