import type { WaspResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Wasp phase: game clear. */
const PHASE_GAME_CLEAR = 1;

/** Returns a Wasp frontend hint derived from the backend hint, or null. */
export function getWaspHint(state: WaspResponse): HintResult | null {
  if (state.phase >= PHASE_GAME_CLEAR) return null;
  if (!state.hint) return null;

  const isDeal = state.hint.fromCol < 0;
  return {
    targetAction: isDeal ? 'dealStock' : 'moveToTableau',
    reason: isDeal ? 'frontendHint.dealStock' : 'frontendHint.moveToTableau',
    confidence: 'moderate',
  };
}
