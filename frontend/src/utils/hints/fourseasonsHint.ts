import type { FourSeasonsResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Four Seasons phase: game clear. */
const PHASE_GAME_CLEAR = 1;

/**
 * Returns a Four Seasons frontend hint derived from the backend hint, or null.
 *
 * The backend only ever suggests a move to a foundation — a cross-to-cross
 * shuffle is never forced, so suggesting one would be noise.
 */
export function getFourSeasonsHint(state: FourSeasonsResponse): HintResult | null {
  if (state.phase >= PHASE_GAME_CLEAR) return null;
  if (!state.hint) return null;

  if (state.hint.fromZone === 'waste') {
    return {
      targetAction: `waste-to-f${state.hint.toIdx}`,
      reason: 'frontendHint.fourseasonsWaste',
      confidence: 'strong',
    };
  }
  return {
    targetAction: `t${state.hint.fromIdx}-to-f${state.hint.toIdx}`,
    reason: 'frontendHint.fourseasonsTableau',
    confidence: 'strong',
  };
}
