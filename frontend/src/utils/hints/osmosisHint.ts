import type { OsmosisResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Osmosis phase: game clear. */
const PHASE_GAME_CLEAR = 1;

/**
 * Returns an Osmosis frontend hint derived from the backend hint, branching by source zone to indicate where cards seep from.
 */
export function getOsmosisHint(state: OsmosisResponse): HintResult | null {
  if (state.phase >= PHASE_GAME_CLEAR) return null;
  if (!state.hint) return null;

  const { hint } = state;
  if (hint.fromZone === 'reserve') {
    return {
      targetAction: 'moveToFoundation',
      reason: 'frontendHint.osmosisFromReserve',
      reasonParams: { col: hint.fromCol, foundation: hint.toCol },
      targetPos: hint.toCol,
      confidence: 'strong',
    };
  }

  if (hint.fromZone === 'waste') {
    return {
      targetAction: 'moveToFoundation',
      reason: 'frontendHint.osmosisFromWaste',
      reasonParams: { foundation: hint.toCol },
      targetPos: hint.toCol,
      confidence: 'strong',
    };
  }

  return {
    targetAction: 'moveToFoundation',
    reason: 'frontendHint.moveToFoundation',
    targetPos: hint.toCol,
    confidence: 'strong',
  };
}
