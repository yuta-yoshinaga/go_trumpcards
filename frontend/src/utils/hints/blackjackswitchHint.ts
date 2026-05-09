import type { BlackJackSwitchResponse as _BlackJackSwitchResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a Blackjack Switch hint. The variant's basic strategy diverges from
 * standard Blackjack in non-trivial ways (the dealer-22-push and 1:1 BJ payout
 * shift many borderline decisions), so this stub returns null until a proper
 * lookup table is built. Keeping the registration consistent across games is
 * required by the new-game checklist (#1669 — checklist item 14).
 */
export function getBlackjackswitchHint(_state: _BlackJackSwitchResponse): HintResult | null {
  return null;
}
