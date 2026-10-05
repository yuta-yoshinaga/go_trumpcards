import type { SirTommyResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** SirTommy phase: game clear. */
const PHASE_GAME_CLEAR = 1;

/** Returns a SirTommy frontend hint derived from the backend hint, or null. */
export function getSirTommyHint(state: SirTommyResponse): HintResult | null {
  if (state.phase >= PHASE_GAME_CLEAR) return null;
  if (!state.hint) return null;

  if (state.hint.fromZone === 'stock') {
    if (state.hint.toZone === 'waste') {
      return {
        targetAction: `stock-to-waste${state.hint.wasteIdx}`,
        reason: 'frontendHint.sirtommyStockToWaste',
        confidence: 'strong',
      };
    }
    return {
      targetAction: `stock-to-f${state.hint.foundationIdx}`,
      reason: 'frontendHint.sirtommyStock',
      confidence: 'strong',
    };
  }
  return {
    targetAction: `waste${state.hint.wasteIdx}-to-f${state.hint.foundationIdx}`,
    reason: 'frontendHint.sirtommyWaste',
    confidence: 'strong',
  };
}
