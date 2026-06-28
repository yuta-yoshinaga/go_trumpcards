import type { BeggarMyNeighbourResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { BeggarMyNeighbourPhase } from '../../types/phases';

/** Returns a Beggar-My-Neighbour frontend hint or null. */
export function getBeggarMyNeighbourHint(state: BeggarMyNeighbourResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  switch (state.phase) {
    case BeggarMyNeighbourPhase.PAY_PENALTY:
      return { targetAction: 'step', reason: 'hint.payPenalty', confidence: 'strong' };
    case BeggarMyNeighbourPhase.COLLECT:
      return { targetAction: 'step', reason: 'hint.collectPile', confidence: 'strong' };
    default:
      return { targetAction: 'step', reason: 'hint.playCard', confidence: 'strong' };
  }
}
