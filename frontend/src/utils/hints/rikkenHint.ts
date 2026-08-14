import type { RikkenResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { RikkenPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Rikken, or null when there is
 * nothing to advise.
 *
 * The decision that matters is the contract: bidding one you cannot make
 * loses points, and because scoring is zero-sum, defending a contract someone
 * else overreached on scores just as well. In play there is rarely a choice
 * beyond following suit, so the hint names that rule rather than pretending
 * to a read.
 */
export function getRikkenHint(state: RikkenResponse): HintResult | null {
  if (state.gameEndFlag || !state.isHumanTurn) return null;

  if (state.phase === RikkenPhase.BID) {
    return { targetAction: 'bid', reason: 'frontendHint.rikkenBidStrength', confidence: 'moderate' };
  }
  if (state.phase === RikkenPhase.PLAY) {
    return { targetAction: 'play', reason: 'frontendHint.rikkenFollowSuit', confidence: 'moderate' };
  }
  return null;
}
