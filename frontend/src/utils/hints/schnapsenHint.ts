import type { SchnapsenResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns null — optimal Schnapsen play depends on marriage timing, trump
 * conservation, stock-depletion tracking, and the must-follow second phase that
 * the client does not model. The authoritative hint is computed server-side
 * (see internal/domain/Schnapsen.go GetHint, surfaced via the `hint` command).
 * This explicit stub keeps the game registered in {@link hooks/useGameHint.useGameHint | useGameHint} so the
 * hint-toggle UI stays consistent across the suite.
 */
export function getSchnapsenHint(_state: SchnapsenResponse): HintResult | null {
  return null;
}
