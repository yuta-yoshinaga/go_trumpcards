import type { EasthavenResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Easthaven phase: game clear. */
const PHASE_GAME_CLEAR = 1;

/** Returns an Easthaven frontend hint derived from the backend hint, or null. */
export function getEasthavenHint(state: EasthavenResponse): HintResult | null {
  if (state.phase >= PHASE_GAME_CLEAR) return null;
  if (!state.hint) return null;

  const isFoundation = state.hint.toZone === 'foundation';
  return {
    targetAction: isFoundation ? 'moveToFoundation' : 'moveToTableau',
    reason: isFoundation ? 'frontendHint.moveToFoundation' : 'frontendHint.moveToTableau',
    confidence: isFoundation ? 'strong' : 'moderate',
  };
}
