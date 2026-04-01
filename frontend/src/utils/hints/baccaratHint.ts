import type { BaccaratResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { BaccaratBetType, BaccaratPhase } from '../../types/phases';

/**
 * Returns a hint for Baccarat during the bet phase.
 * Banker bet has the lowest house edge (1.06%) vs Player (1.24%) vs Tie (14.36%).
 * If the user already selected Banker, no hint is shown (already optimal).
 * If the user selected Tie, warns about the high house edge.
 */
export function getBaccaratHint(state: BaccaratResponse): HintResult | null {
  if (state.phase !== BaccaratPhase.BET) return null;

  if (state.betType === BaccaratBetType.BANKER) return null;

  if (state.betType === BaccaratBetType.TIE) {
    return {
      targetAction: 'banker',
      reason: 'hintReason.avoidTie',
      confidence: 'strong',
    };
  }

  return {
    targetAction: 'banker',
    reason: 'hintReason.bankerBestOdds',
    confidence: 'strong',
  };
}
