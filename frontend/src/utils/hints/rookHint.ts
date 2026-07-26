import type { RookResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns null — optimal Rook (ルーク) play depends on bidding evaluation, nest
 * inference, trump-color choice, and partnership signalling that the client
 * does not model. The authoritative hint is computed server-side (see
 * internal/domain/Rook.go GetHint, surfaced via the `hint` command). This
 * explicit stub keeps the game registered in {@link hooks/useGameHint.useGameHint | useGameHint} so the
 * hint-toggle UI stays consistent across the suite.
 */
export function getRookHint(_state: RookResponse): HintResult | null {
  return null;
}
