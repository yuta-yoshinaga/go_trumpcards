import type { CruelResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Cruel phase: game clear. */
const PHASE_GAME_CLEAR = 1;

/** Returns a Cruel frontend hint derived from the backend hint, or null. */
export function getCruelHint(state: CruelResponse): HintResult | null {
  if (state.phase >= PHASE_GAME_CLEAR) return null;
  // If the game flagged a stalemate, surface a Shift recommendation even when
  // the backend hint payload is empty — Shift is the only escape.
  if (!state.hint) {
    if (state.isStalemate) {
      return {
        targetAction: 'shift',
        reason: 'frontendHint.shiftRecommended',
        confidence: 'moderate',
      };
    }
    return null;
  }

  const isFoundation = state.hint.toZone === 'foundation';
  return {
    targetAction: isFoundation ? 'moveToFoundation' : 'moveToTableau',
    reason: isFoundation ? 'frontendHint.moveToFoundation' : 'frontendHint.moveToTableau',
    confidence: isFoundation ? 'strong' : 'moderate',
  };
}
