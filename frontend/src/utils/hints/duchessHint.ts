import type { DuchessResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { DuchessPhase } from '../../types/phases';

/** Returns a Duchess frontend hint or null.
 *
 * Choosing the base rank comes first: until it is set nothing else is legal, so
 * that case is surfaced as its own reason rather than as a reserve move. After
 * that the backend's own ranking is translated, with reserve moves called out
 * separately because emptying the reserve is what unlocks the empty columns. */
export function getDuchessHint(state: DuchessResponse): HintResult | null {
  if (state.phase !== DuchessPhase.PLAYING) return null;
  if (state.awaitingBaseRank) {
    return { targetAction: 'play.chooseBase', reason: 'hintReason.chooseBase', confidence: 'strong' };
  }
  if (state.isStalemate) return null;
  if (!state.hint) return null;
  if (state.hint.toZone === 'foundation') {
    return { targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' };
  }
  if (state.hint.fromZone === 'stock') {
    return { targetAction: 'play.draw', reason: 'hintReason.draw', confidence: 'moderate' };
  }
  if (state.hint.fromZone === 'reserve') {
    return { targetAction: 'play.tableau', reason: 'hintReason.fromReserve', confidence: 'strong' };
  }
  return { targetAction: 'play.tableau', reason: 'hintReason.toTableau', confidence: 'strong' };
}
