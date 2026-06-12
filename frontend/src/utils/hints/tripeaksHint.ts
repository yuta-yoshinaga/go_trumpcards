import type { TriPeaksResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { TriPeaksPhase } from '../../types/phases';

/** Maximum card value (King). */
const MAX_VALUE = 13;

/** Returns a frontend HintResult for TriPeaks, or null if no suggestion. */
export function getTriPeaksHint(state: TriPeaksResponse): HintResult | null {
  if (state.phase !== TriPeaksPhase.PLAYING) return null;

  const wasteTop = state.waste.length > 0 ? state.waste[state.waste.length - 1] : null;
  if (!wasteTop) {
    if (state.stockCount > 0) {
      return { targetAction: 'draw', reason: 'frontendHint.drawFromStock', confidence: 'moderate' };
    }
    return null;
  }

  const removableCount = countRemovableCards(state, wasteTop.value);

  if (removableCount > 1) {
    return { targetAction: 'remove', reason: 'frontendHint.multipleRemovable', confidence: 'strong' };
  }
  if (removableCount === 1) {
    return { targetAction: 'remove', reason: 'frontendHint.canRemove', confidence: 'strong' };
  }

  if (state.stockCount > 0) {
    return { targetAction: 'draw', reason: 'frontendHint.drawFromStock', confidence: 'moderate' };
  }

  return null;
}

/** Count exposed cards adjacent in value to the waste top (with King-Ace wrap). */
function countRemovableCards(state: TriPeaksResponse, wasteValue: number): number {
  let count = 0;
  for (const row of state.layout) {
    for (const cell of row) {
      if (cell.card && !cell.removed && cell.exposed) {
        if (isAdjacent(cell.card.value, wasteValue)) count++;
      }
    }
  }
  return count;
}

/** Check if two values are adjacent (with King-Ace wrap-around). */
export function isTriPeaksAdjacent(a: number, b: number): boolean {
  return isAdjacent(a, b);
}

function isAdjacent(a: number, b: number): boolean {
  const diff = Math.abs(a - b);
  return diff === 1 || diff === MAX_VALUE - 1;
}
