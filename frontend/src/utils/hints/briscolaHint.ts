import type { BriscolaResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a Briscola frontend hint derived from the backend hint, or null.
 *
 * The suggestion rides along with every state response (see
 * BriscolaWebPresenter.Output, #4483), and Briscola.GetHint already returns
 * nil unless it is the human's turn to play, so there is nothing to re-check
 * here beyond having a card to point at.
 */
export function getBriscolaHint(state: BriscolaResponse): HintResult | null {
  const hint = state.hint;
  if (!hint || hint.cardIndex === undefined || hint.cardIndex < 0) return null;

  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
