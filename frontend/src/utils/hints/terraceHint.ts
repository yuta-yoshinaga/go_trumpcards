import type { TerraceResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { TerracePhase } from '../../types/phases';

/** Returns a Terrace frontend hint or null.
 *
 * A terrace move gets its own reason because the terrace can only ever reach a
 * foundation and is never refilled -- letting it stall is what loses the game,
 * so the advice is stronger than an ordinary foundation move. */
export function getTerraceHint(state: TerraceResponse): HintResult | null {
  if (state.phase !== TerracePhase.PLAYING) return null;
  if (state.awaitingBaseRank) {
    return { targetAction: 'play.chooseBase', reason: 'hintReason.chooseBase', confidence: 'strong' };
  }
  if (state.isStalemate) return null;
  if (!state.hint) return null;
  if (state.hint.fromZone === 'reserve') {
    return { targetAction: 'play.foundation', reason: 'hintReason.fromTerrace', confidence: 'strong' };
  }
  if (state.hint.toZone === 'foundation') {
    return { targetAction: 'play.foundation', reason: 'hintReason.toFoundation', confidence: 'strong' };
  }
  if (state.hint.fromZone === 'stock') {
    return { targetAction: 'play.draw', reason: 'hintReason.draw', confidence: 'moderate' };
  }
  return { targetAction: 'play.tableau', reason: 'hintReason.toTableau', confidence: 'strong' };
}
