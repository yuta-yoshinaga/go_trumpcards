import type { FiveHundredResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns null — optimal 500 (Five Hundred) play depends on bidding evaluation,
 * bower/joker trump tracking, kitty inference, and partnership signalling that
 * the client does not model. The authoritative hint is computed server-side
 * (see internal/domain/FiveHundred.go GetHint, surfaced via the `hint` command).
 * This explicit stub keeps the game registered in {@link hooks/useGameHint.useGameHint | useGameHint} so the
 * hint-toggle UI stays consistent across the suite.
 */
export function getFiveHundredHint(_state: FiveHundredResponse): HintResult | null {
  return null;
}
