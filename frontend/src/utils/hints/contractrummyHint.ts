import type { ContractRummyResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a hint for the current Contract Rummy state, or null if no hint is
 * available. Currently a stub: Contract Rummy's optimal play involves
 * planning many turns ahead toward a multi-meld contract, which is hard to
 * surface as a single-step hint without overwhelming the player. Future work:
 * highlight cards that fit the current round's contract, or suggest the best
 * discard to deny opponents.
 */
export function getContractRummyHint(_state: ContractRummyResponse | null): HintResult | null {
  return null;
}
