import type { BidWhistResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns null — optimal Bid Whist play depends on bidding evaluation,
 * joker/trump tracking across Uptown/Downtown reversals, kitty inference, and
 * partnership signalling that the client does not model. The authoritative hint
 * is computed server-side (see internal/domain/BidWhist.go GetHint, surfaced via
 * the `hint` command). This explicit stub keeps the game registered in
 * {@link hooks/useGameHint.useGameHint | useGameHint} so the hint-toggle UI stays consistent across the suite.
 */
export function getBidWhistHint(_state: BidWhistResponse): HintResult | null {
  return null;
}
