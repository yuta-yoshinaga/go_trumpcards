import type { BristolResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Bristol phase: game clear. */
const PHASE_GAME_CLEAR = 1;

/** Returns a Bristol frontend hint derived from the backend hint, or null. */
export function getBristolHint(state: BristolResponse): HintResult | null {
  if (state.phase >= PHASE_GAME_CLEAR) return null;
  if (!state.hint) return null;

  if (state.hint.toZone === 'foundation') {
    return {
      targetAction: 'moveToFoundation',
      reason: 'frontendHint.moveToFoundation',
      confidence: 'strong',
    };
  }
  return {
    targetAction: 'moveToTableau',
    reason: 'frontendHint.moveToTableau',
    confidence: 'moderate',
  };
}
