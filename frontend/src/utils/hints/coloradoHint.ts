import type { ColoradoResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Colorado phase: game clear. */
const PHASE_GAME_CLEAR = 1;

/**
 * Returns a Colorado frontend hint derived from the backend hint, or null.
 *
 * The backend falls back to "bury the waste card somewhere" once nothing else
 * is available. That move is always legal, so surfacing it as a highlight would
 * light up the board on almost every turn — only the moves that make progress
 * are worth pointing at.
 */
export function getColoradoHint(state: ColoradoResponse): HintResult | null {
  if (state.phase >= PHASE_GAME_CLEAR) return null;
  if (!state.hint) return null;

  const { fromZone, fromIdx, toZone, toIdx } = state.hint;

  if (toZone === 'foundation') {
    if (fromZone === 'waste') {
      return {
        targetAction: `waste-to-f${toIdx}`,
        reason: 'frontendHint.coloradoWaste',
        confidence: 'strong',
      };
    }
    return {
      targetAction: `t${fromIdx}-to-f${toIdx}`,
      reason: 'frontendHint.coloradoTableau',
      confidence: 'strong',
    };
  }

  if (fromZone === 'stock' && toZone === 'tableau') {
    return {
      targetAction: `stock-to-t${toIdx}`,
      reason: 'frontendHint.coloradoFillGap',
      confidence: 'moderate',
    };
  }

  if (fromZone === 'stock') {
    return { targetAction: 'draw', reason: 'frontendHint.coloradoDraw', confidence: 'moderate' };
  }

  return null;
}
