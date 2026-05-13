import type { PiquetResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { PiquetPhase } from '../../types/phases';

/** Returns a Piquet frontend hint or null.
 *
 * In Piquet the backend supplies the recommended card index or discard list.
 * The frontend hint here just routes the player attention to the right action.
 */
export function getPiquetHint(state: PiquetResponse): HintResult | null {
  if (!state.hint) return null;
  if (state.phase === PiquetPhase.PLAY && state.hint.cardIndex != null) {
    return { targetAction: 'play.card', reason: 'piquet.hint.play', confidence: 'moderate' };
  }
  if (state.phase === PiquetPhase.EXCHANGE && state.hint.discardIndices != null) {
    return { targetAction: 'play.exchange', reason: 'piquet.hint.discard', confidence: 'moderate' };
  }
  return null;
}
