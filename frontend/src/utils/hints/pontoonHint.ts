import type { PontoonResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { PontoonPhase } from '../../types/phases';

/** これ以下ならどの札を引いてもバーストしない。 */
const SAFE_TO_DRAW = 11;

/** 宣言できる最低の総計。Pontoon は 15 未満でスティックできない。 */
const STICK_MINIMUM = 15;

/** ここから上は引かずに勝負する。 */
const STRONG_TOTAL = 17;

/** 5 枚トリックまであと 1 枚。 */
const FIVE_CARD_REACH = 4;

/**
 * Returns a frontend {@link HintResult} for Pontoon, or null when no suggestion
 * is available.
 *
 * Pontoon exposes no hint from the backend, so this is computed from the
 * visible state. It never invents legality: `canStick` / `canTwist` / `canBuy` /
 * `canSplit` come from the server precisely because the 15-minimum and the
 * no-buy-after-twist rule would otherwise be implemented twice and drift, so
 * the hint only ever names an action the server has already allowed.
 */
export function getPontoonHint(state: PontoonResponse): HintResult | null {
  if (state.phase !== PontoonPhase.PLAYER_TURN) return null;

  const seat = state.seats[0];
  // 席 0 が人間。他の席が動いている間は助言することがない。
  if (!seat || seat.isCpu || state.activeSeat !== 0) return null;

  const hand = seat.hands[state.activeHand];
  // 伏せられている手は total が 0 で届く。読んで助言すると嘘になる。
  if (!hand || hand.hidden) return null;

  if (state.canSplit) {
    return { targetAction: 'split', reason: 'frontendHint.pontoonSplit', confidence: 'moderate' };
  }

  const total = hand.total;

  if (total <= SAFE_TO_DRAW) {
    // **押せない手を勧めない。**ツイスト後は買えないので canBuy を見る。
    return state.canBuy
      ? { targetAction: 'buy', reason: 'frontendHint.pontoonBuySafe', confidence: 'strong' }
      : { targetAction: 'twist', reason: 'frontendHint.pontoonDrawSafe', confidence: 'strong' };
  }

  if (total < STICK_MINIMUM) {
    // 15 未満は宣言できない。引く以外の道がない。
    return state.canTwist
      ? { targetAction: 'twist', reason: 'frontendHint.pontoonMustDraw', confidence: 'strong' }
      : null;
  }

  // **4 枚での 15-16 だけは引く。**5 枚トリックはポイント手に勝つので、
  // 境界の総計を捨ててでも狙う価値がある。
  if (hand.cards.length === FIVE_CARD_REACH && total < STRONG_TOTAL && state.canTwist) {
    return { targetAction: 'twist', reason: 'frontendHint.pontoonFiveCard', confidence: 'moderate' };
  }

  if (!state.canStick) return null;
  return total >= STRONG_TOTAL
    ? { targetAction: 'stick', reason: 'frontendHint.pontoonStickStrong', confidence: 'strong' }
    : { targetAction: 'stick', reason: 'frontendHint.pontoonStickBorderline', confidence: 'moderate' };
}
