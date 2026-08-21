import type { SlyFoxResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Sly Fox phase: game clear. */
const PHASE_GAME_CLEAR = 1;

/**
 * Returns a Sly Fox frontend hint derived from the backend hint, or null.
 *
 * There is no waste: a `stock` hint means "deal the next card", and its `toIdx`
 * names the reserve slot that costs the least to bury — or a foundation, when
 * the dealt card can go straight there without spending one of the twenty.
 */
export function getSlyFoxHint(state: SlyFoxResponse): HintResult | null {
  if (state.phase >= PHASE_GAME_CLEAR) return null;
  if (!state.hint) return null;

  const { fromZone, fromIdx, toZone, toIdx } = state.hint;

  if (fromZone === 'stock') {
    // 組札へ直接送れる札は 20 枚に数えないので、常にこちらが得。
    if (toZone === 'foundation') {
      return {
        targetAction: `deal-to-f${toIdx}`,
        reason: 'frontendHint.slyFoxDealToFoundation',
        confidence: 'strong',
      };
    }
    return {
      targetAction: `deal-to-t${toIdx}`,
      reason: 'frontendHint.slyFoxDealToSlot',
      confidence: 'moderate',
    };
  }

  if (toZone === 'foundation') {
    return {
      targetAction: `t${fromIdx}-to-f${toIdx}`,
      reason: 'frontendHint.slyFoxTableau',
      confidence: 'strong',
    };
  }

  return null;
}
