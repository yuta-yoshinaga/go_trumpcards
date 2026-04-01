import type { PyramidResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { PyramidPhase } from '../../types/phases';

/** Card value for King (can be removed solo). */
const KING_VALUE = 13;

/** Target sum for a removable pair. */
const TARGET_SUM = 13;

/** Returns a frontend HintResult for Pyramid, or null if no suggestion. */
export function getPyramidHint(state: PyramidResponse): HintResult | null {
  if (state.phase !== PyramidPhase.PLAYING) return null;

  const exposed = getExposedCards(state);

  // Priority 1: Solo King removal
  if (exposed.some((c) => c.value === KING_VALUE)) {
    return { targetAction: 'remove', reason: 'frontendHint.removeKing', confidence: 'strong' };
  }

  // Priority 2: Pyramid pair summing to 13
  if (hasPairSummingTo13(exposed)) {
    return { targetAction: 'remove', reason: 'frontendHint.removePair', confidence: 'strong' };
  }

  // Priority 3: Waste + exposed card summing to 13
  const wasteTop = state.waste.length > 0 ? state.waste[state.waste.length - 1] : null;
  if (wasteTop) {
    if (wasteTop.value === KING_VALUE) {
      return { targetAction: 'remove', reason: 'frontendHint.useWasteKing', confidence: 'strong' };
    }
    if (exposed.some((c) => c.value + wasteTop.value === TARGET_SUM)) {
      return { targetAction: 'remove', reason: 'frontendHint.useWaste', confidence: 'strong' };
    }
  }

  // Priority 4: Draw from stock
  if (state.stockCount > 0) {
    return { targetAction: 'draw', reason: 'frontendHint.drawFromStock', confidence: 'moderate' };
  }

  return null;
}

/** Collect all exposed, non-removed cards with their values. */
function getExposedCards(state: PyramidResponse): { value: number }[] {
  const result: { value: number }[] = [];
  for (const row of state.pyramid) {
    for (const cell of row) {
      if (cell.card && !cell.removed && cell.exposed) {
        result.push({ value: cell.card.value });
      }
    }
  }
  return result;
}

/** Check if any two exposed cards sum to 13. */
function hasPairSummingTo13(cards: { value: number }[]): boolean {
  for (let i = 0; i < cards.length; i++) {
    for (let j = i + 1; j < cards.length; j++) {
      if (cards[i].value + cards[j].value === TARGET_SUM) return true;
    }
  }
  return false;
}
