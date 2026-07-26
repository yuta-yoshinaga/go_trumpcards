import type { PresidentResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns null — President's optimal play depends on information-set reasoning
 * (opponent hands, partnership dynamics) that the client does not model.
 * Keeping an explicit stub so the game is registered in {@link hooks/useGameHint.useGameHint | useGameHint}
 * and the hint toggle UI stays consistent across the suite.
 */
export function getPresidentHint(_state: PresidentResponse): HintResult | null {
  return null;
}
