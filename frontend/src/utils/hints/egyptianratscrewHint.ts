import type { EgyptianRatscrewResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Returns an Egyptian Ratscrew frontend hint or null. */
export function getEgyptianRatscrewHint(state: EgyptianRatscrewResponse): HintResult | null {
  if (state.gameEndFlag) return null;
  if (state.isSlappable) {
    return { targetAction: 'slap', reason: 'hint.slap', confidence: 'strong' };
  }
  if (state.isHumanTurn) {
    return { targetAction: 'step', reason: 'hint.flipCard', confidence: 'moderate' };
  }
  return null;
}
