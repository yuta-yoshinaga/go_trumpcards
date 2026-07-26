import type { BigTwoResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns null — Big Two's optimal play depends on complex hand evaluation
 * and opponent modeling that the client does not perform.
 * Keeping an explicit stub so the game is registered in {@link hooks/useGameHint.useGameHint | useGameHint}
 * and the hint toggle UI stays consistent across the suite.
 */
export function getBigTwoHint(_state: BigTwoResponse): HintResult | null {
  return null;
}
