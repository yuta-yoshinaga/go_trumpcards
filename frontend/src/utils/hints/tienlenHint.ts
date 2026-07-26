import type { TienLenResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns null — Tien Len's optimal play depends on combination evaluation,
 * chop timing, and opponent modeling that the client does not perform.
 * Keeping an explicit stub so the game is registered in {@link hooks/useGameHint.useGameHint | useGameHint}
 * and the hint toggle UI stays consistent across the suite.
 */
export function getTienLenHint(_state: TienLenResponse): HintResult | null {
  return null;
}
