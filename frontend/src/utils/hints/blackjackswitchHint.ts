import type { BlackJackSwitchResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { BlackJackSwitchPhase } from '../../types/phases';

/** ディーラーが 17 で止まる境目。ソフト 17 は引く (`BJSwitchDealerHitsSoft17`)。 */
const DEALER_STANDS = 17;

/** これ以上引くとバーストしやすい上限。 */
const HIT_BELOW = 17;

/** ディーラーの弱い見せ札 (2-6)。 */
const DEALER_WEAK_MAX = 6;

/**
 * Returns a frontend {@link HintResult} for Blackjack Switch, or null when no
 * suggestion is available.
 *
 * There is no server-side GetHint, so this is derived from the two scores.
 *
 * Two things separate this from ordinary blackjack, and both push the player
 * toward *more* aggression, not less:
 *
 * - **Switching is free.** The second card of each hand may be swapped once,
 *   right after the deal. Nothing is staked on it, so the only question is
 *   whether the swap improves the pair — and the hint says to take it whenever
 *   it does.
 * - **A dealer 22 is a push, not a bust** (`BlackJackSwitch.go:33`). That is the
 *   price paid for the switch, and it means a marginal hand is worth *less* than
 *   in ordinary blackjack: the dealer's worst outcome no longer pays you. So the
 *   hint leans to standing on the usual 12-16-against-a-weak-upcard spots rather
 *   than hoping the dealer breaks.
 */
export function getBlackjackswitchHint(state: BlackJackSwitchResponse): HintResult | null {
  if (state.phase === BlackJackSwitchPhase.BET) {
    return state.chips <= 0
      ? null
      : { targetAction: 'bet', reason: 'frontendHint.blackjackswitchBet', confidence: 'moderate' };
  }

  if (state.phase === BlackJackSwitchPhase.SWITCH) {
    return switchImproves(state)
      ? { targetAction: 'switch', reason: 'frontendHint.blackjackswitchSwitch', confidence: 'moderate' }
      : { targetAction: 'keep', reason: 'frontendHint.blackjackswitchKeep', confidence: 'moderate' };
  }

  if (state.phase !== BlackJackSwitchPhase.ACTION) return null;

  const hand = state.hands[state.currentHandIdx];
  if (!hand || hand.stood || hand.busted) return null;

  if (hand.score >= HIT_BELOW) {
    return { targetAction: 'stand', reason: 'frontendHint.blackjackswitchStand', confidence: 'moderate' };
  }

  // **22 がプッシュなので、ディーラーの崩れを当てにできない。**弱い見せ札でも
  // 通常より強く「引く」側に寄る。
  const upcard = dealerUpcard(state);
  if (hand.score >= 12 && upcard > 0 && upcard <= DEALER_WEAK_MAX) {
    return { targetAction: 'stand', reason: 'frontendHint.blackjackswitchStandPush22', confidence: 'moderate' };
  }

  return { targetAction: 'hit', reason: 'frontendHint.blackjackswitchHit', confidence: 'moderate' };
}

/** ディーラーの見せ札の値。A は 11 として数える。伏せ札しか無ければ 0。 */
function dealerUpcard(state: BlackJackSwitchResponse): number {
  const up = state.dealerCards.find((c) => c !== null);
  if (!up) return 0;
  return up.value === 1 ? 11 : Math.min(up.value, 10);
}

/**
 * 入れ替えると 2 ハンドの合計が良くなるか。
 *
 * 判定は素朴に「両手が 17 以上になる方を選ぶ」。片方を強くして片方を潰す入れ替えは
 * 合計では得しないので、**両手が使える形かどうか**で見る。
 */
function switchImproves(state: BlackJackSwitchResponse): boolean {
  const [a, b] = state.hands;
  if (!a || !b) return false;
  const now = usable(a.score) + usable(b.score);
  // 2 枚目を入れ替えた後のスコアは (合計 - 自分の 2 枚目 + 相手の 2 枚目)。
  const aSecond = cardValue(a.cards[1]);
  const bSecond = cardValue(b.cards[1]);
  const after = usable(a.score - aSecond + bSecond) + usable(b.score - bSecond + aSecond);
  return after > now;
}

/** 17 以上 21 以下なら 1、それ以外は 0。 */
function usable(score: number): number {
  return score >= DEALER_STANDS && score <= 21 ? 1 : 0;
}

/** 札の値。A は 11、絵札は 10。null は 0。 */
function cardValue(c: { value: number } | null | undefined): number {
  if (!c) return 0;
  if (c.value === 1) return 11;
  return Math.min(c.value, 10);
}
