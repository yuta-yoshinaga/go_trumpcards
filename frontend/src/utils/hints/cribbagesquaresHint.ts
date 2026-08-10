import type { CribbageSquaresResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { CribbageSquaresPhase } from '../../types/phases';

/**
 * Returns a Cribbage Squares placement hint, or null when the game is over or
 * there is no card in hand.
 *
 * The server already works out the best cell using the real cribbage scorer,
 * so this forwards that answer rather than re-deriving a weaker one in the
 * page. `targetAction` is `cell-<row>-<col>`, matching `data-hint-action` on
 * the grid buttons.
 */
export function getCribbageSquaresHint(state: CribbageSquaresResponse): HintResult | null {
  if (state.phase !== CribbageSquaresPhase.PLAYING) return null;
  if (!state.currentCard) return null;
  if (!state.hint) return null;

  return {
    targetAction: `cell-${state.hint.row}-${state.hint.col}`,
    reason: state.hint.synergy ? 'hint.placeSynergy' : 'hint.placeAny',
    confidence: state.hint.score >= 6 ? 'strong' : 'moderate',
  };
}
