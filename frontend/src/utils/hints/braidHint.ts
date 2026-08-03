import type { BraidResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { BraidPhase } from '../../types/phases';

/** Returns a Braid frontend hint or null.
 *
 * A braid-field move gets its own reason because those four slots are the only
 * thing that consumes the braid -- letting them clog is what loses the game, so
 * the advice is stronger than an ordinary foundation move. */
export function getBraidHint(state: BraidResponse): HintResult | null {
  if (state.phase !== BraidPhase.PLAYING) return null;
  // Nothing else is legal until the direction is fixed, so it outranks even a
  // stalemate readout.
  if (state.awaitingDirection) {
    return { targetAction: 'play.chooseDirection', reason: 'hintReason.chooseDirection', confidence: 'strong' };
  }
  if (state.isStalemate) return null;
  if (!state.hint) return null;
  if (state.hint.fromZone === 'field') {
    return { targetAction: 'play.foundation', reason: 'hintReason.fromBraidField', confidence: 'strong' };
  }
  if (state.hint.toZone === 'foundation') {
    return { targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' };
  }
  if (state.hint.fromZone === 'stock') {
    return { targetAction: 'play.draw', reason: 'hintReason.draw', confidence: 'moderate' };
  }
  return { targetAction: 'play.helper', reason: 'hintReason.toHelper', confidence: 'moderate' };
}
