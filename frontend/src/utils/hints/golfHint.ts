import type { GolfResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { GolfPhase } from '../../types/phases';

/** True when two Golf card values are adjacent (±1, with K-A wraparound). */
function isAdjacentRank(v1: number, v2: number): boolean {
  const diff = Math.abs(v1 - v2);
  return diff === 1 || diff === 12;
}

/** Returns a frontend HintResult for Golf Solitaire, or null when nothing is suggestable. */
export function getGolfHint(state: GolfResponse): HintResult | null {
  if (state.phase !== GolfPhase.PLAYING) return null;

  const wasteTop = state.waste.at(-1);
  let removable = 0;
  if (wasteTop) {
    for (const column of state.layout) {
      const top = column.find((c) => c.exposed && !c.removed && c.card !== null);
      if (top?.card && isAdjacentRank(top.card.value, wasteTop.value)) {
        removable++;
      }
    }
  }

  if (removable > 1) {
    return { targetAction: 'remove', reason: 'frontendHint.multipleRemovable', confidence: 'strong' };
  }
  if (removable === 1) {
    return { targetAction: 'remove', reason: 'frontendHint.canRemove', confidence: 'strong' };
  }
  if (state.stockCount > 0) {
    return { targetAction: 'draw', reason: 'frontendHint.drawFromStock', confidence: 'moderate' };
  }
  return null;
}
