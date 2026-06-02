import type { OsmosisResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Osmosis phase: game clear. */
const PHASE_GAME_CLEAR = 1;

/** Returns an Osmosis frontend hint derived from the backend hint, or null. */
export function getOsmosisHint(state: OsmosisResponse): HintResult | null {
  if (state.phase >= PHASE_GAME_CLEAR) return null;
  if (!state.hint) return null;

  return {
    targetAction: 'moveToFoundation',
    reason: 'frontendHint.moveToFoundation',
    confidence: 'strong',
  };
}
