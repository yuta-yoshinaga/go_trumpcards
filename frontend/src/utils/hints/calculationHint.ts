import type { CalculationResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Calculation phase: game clear. */
const PHASE_GAME_CLEAR = 1;

/** Returns a Calculation frontend hint derived from the backend hint, or null. */
export function getCalculationHint(state: CalculationResponse): HintResult | null {
  if (state.phase >= PHASE_GAME_CLEAR) return null;
  if (!state.hint) return null;

  if (state.hint.fromZone === 'stock') {
    return {
      targetAction: `stock-to-f${state.hint.foundationIdx}`,
      reason: 'frontendHint.calculationStock',
      confidence: 'strong',
    };
  }
  return {
    targetAction: `waste${state.hint.wasteIdx}-to-f${state.hint.foundationIdx}`,
    reason: 'frontendHint.calculationWaste',
    confidence: 'strong',
  };
}
