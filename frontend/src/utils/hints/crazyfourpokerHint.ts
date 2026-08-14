import type { CrazyFourPokerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { CrazyFourPokerPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Crazy 4 Poker, or null when there is
 * nothing to advise.
 *
 * **It does not decide the multiplier itself.** Whether 3x is available is a
 * domain rule, surfaced as `maxMultiplier`; re-deriving it here would put the
 * game's defining rule in a second place. The hint only distinguishes "you have
 * the privilege, use it" from "ordinary hand".
 */
export function getCrazyfourpokerHint(state: CrazyFourPokerResponse): HintResult | null {
  if (state.gameEndFlag || state.phase !== CrazyFourPokerPhase.DECIDE) return null;

  if (state.hasAcesOrBetter) {
    return { targetAction: 'raise', reason: 'frontendHint.crazyFourPokerAces', confidence: 'strong' };
  }
  // **降りるのも定石の一部。** ディーラーもキング以上でしか成立しないので、
  // こちらがそれに届かないなら降りたほうがよい。判定 (`playerQualifies`) は
  // サーバが計算したものを使う ── ここで札から「キング以上か」を引き直すと、
  // ディーラーの成立条件と同じ規則がフロントにもう 1 つできる。
  if (!state.playerQualifies) {
    return { targetAction: 'fold', reason: 'frontendHint.crazyFourPokerFold', confidence: 'moderate' };
  }
  return { targetAction: 'play', reason: 'frontendHint.crazyFourPokerMinimum', confidence: 'moderate' };
}
