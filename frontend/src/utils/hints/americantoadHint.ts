import type { AmericanToadResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { AmericanToadPhase } from '../../types/phases';

/** Returns an American Toad frontend hint or null.
 *
 * The backend ranks its own suggestions. A reserve move gets its own reason
 * because emptying the reserve is what unlocks the empty columns, and a draw
 * that would spend the single redeal is called out separately -- it is not the
 * same cheap move as an ordinary turn. */
export function getAmericanToadHint(state: AmericanToadResponse): HintResult | null {
  if (state.phase !== AmericanToadPhase.PLAYING) return null;
  if (state.isStalemate) return null;
  if (!state.hint) return null;
  if (state.hint.toZone === 'foundation') {
    return { targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' };
  }
  if (state.hint.fromZone === 'stock') {
    if (state.canRedeal) {
      return { targetAction: 'play.redeal', reason: 'hintReason.redeal', confidence: 'moderate' };
    }
    return { targetAction: 'play.draw', reason: 'hintReason.draw', confidence: 'moderate' };
  }
  if (state.hint.fromZone === 'reserve') {
    return { targetAction: 'play.tableau', reason: 'hintReason.fromReserve', confidence: 'strong' };
  }
  return { targetAction: 'play.tableau', reason: 'hintReason.toTableau', confidence: 'strong' };
}
