import type { Card, FourCardPokerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { FourCardPokerPhase } from '../../types/phases';

/**
 * 4 枚役の順位 (sync: internal/domain/four_card_hand_eval.go:9)。
 *
 * **フラッシュがストレートより上。**4 枚だと階段の方が揃いやすいため、5 枚役の
 * 常識とは逆になる。ここを写し間違えると、フラッシュを持ちながら降りる助言になる。
 */
export const FOUR_CARD_RANK = {
  HIGH_CARD: 1,
  PAIR: 2,
  TWO_PAIR: 3,
  STRAIGHT: 4,
  FLUSH: 5,
  THREE: 6,
  STRAIGHT_FLUSH: 7,
  FOUR: 8,
} as const;

const RANK_HIGH_CARD = 1;
const RANK_PAIR = 2;
const RANK_TWO_PAIR = 3;
const RANK_STRAIGHT = 4;
const RANK_FLUSH = 5;
const RANK_THREE = 6;
const RANK_STRAIGHT_FLUSH = 7;
const RANK_FOUR = 8;

/** プレイベットはアンティの 1〜3 倍 (sync: `FourCardPokerMinPlayMul` / `Max`)。 */
const MAX_MULTIPLIER = 3;

/** これ以上なら最大倍率で押す。 */
const PUSH_FROM = RANK_THREE;

/** これ以下なら降りる。 */
const FOLD_BELOW = RANK_PAIR;

/**
 * Returns a frontend {@link HintResult} for Four Card Poker, or null when no
 * suggestion is available.
 *
 * **`playerHandRank` is not read.** It is only filled by `updateBestHands`,
 * which runs on fold and on resolve (`FourCardPoker.go:204`) — never while the
 * action is still open. Reading it here would make every branch dead in
 * production while unit tests that set the field went on passing, which is
 * exactly the trap found in review on #4622. The hand is evaluated from the five
 * cards the player can see instead.
 *
 * The decision is the play multiplier: fold, or bet one to three times the ante
 * (`FourCardPokerMinPlayMul`/`Max`). Since the dealer shows one card and gets a
 * free extra card, a marginal hand is worth less than it looks, so the hint only
 * pushes to the maximum on a made hand of three of a kind or better.
 *
 * The ranking has one thing worth stating: **a flush beats a straight here**.
 * With four cards a straight is the easier hand to make, so the usual five-card
 * order is inverted (`four_card_hand_eval.go:9`).
 */
export function getFourCardPokerHint(state: FourCardPokerResponse): HintResult | null {
  if (state.phase === FourCardPokerPhase.BET) {
    return state.chips <= 0
      ? null
      : { targetAction: 'bet', reason: 'frontendHint.fourcardpokerBet', confidence: 'moderate' };
  }

  if (state.phase !== FourCardPokerPhase.ACTION) return null;
  if (state.playerHand.length === 0) return null;

  const rank = bestRank(state.playerHand);

  if (rank >= PUSH_FROM) {
    return {
      targetAction: `play-${MAX_MULTIPLIER}`,
      reason: 'frontendHint.fourcardpokerPlayMax',
      confidence: 'moderate',
    };
  }
  if (rank <= FOLD_BELOW) {
    return { targetAction: 'fold', reason: 'frontendHint.fourcardpokerFold', confidence: 'moderate' };
  }
  return { targetAction: 'play-1', reason: 'frontendHint.fourcardpokerPlayMin', confidence: 'moderate' };
}

/**
 * 5 枚から選べる 4 枚の最良ランク。
 *
 * **export しているのはテストのため。**フラッシュとストレートの順序を入れ替えても
 * どちらも「最小ベット」の側に落ちるので、`getFourCardPokerHint` の戻り値だけでは
 * 区別できない —— 実際その負のコントロールが素通りした。順位そのものを見る。
 */
export function bestRank(hand: Card[]): number {
  if (hand.length <= 4) return evalFour(hand);
  let best = RANK_HIGH_CARD;
  for (let skip = 0; skip < hand.length; skip += 1) {
    const r = evalFour(hand.filter((_, i) => i !== skip));
    if (r > best) best = r;
  }
  return best;
}

/** 4 枚役のランク。`evalFourCardHand` と同じ順序。 */
function evalFour(cards: Card[]): number {
  const counts = new Map<number, number>();
  for (const c of cards) counts.set(c.value, (counts.get(c.value) ?? 0) + 1);
  const groups = [...counts.values()].sort((a, b) => b - a);
  const flush = cards.every((c) => c.design === cards[0].design);
  const straight = isStraight(cards);

  if (groups[0] === 4) return RANK_FOUR;
  if (straight && flush) return RANK_STRAIGHT_FLUSH;
  if (groups[0] === 3) return RANK_THREE;
  // **フラッシュがストレートより上。**5 枚役の順序で書くと逆になる。
  if (flush) return RANK_FLUSH;
  if (straight) return RANK_STRAIGHT;
  if (groups[0] === 2 && groups[1] === 2) return RANK_TWO_PAIR;
  if (groups[0] === 2) return RANK_PAIR;
  return RANK_HIGH_CARD;
}

/** 連続する 4 ランクか。A は高くも低くも使える。 */
function isStraight(cards: Card[]): boolean {
  if (cards.length !== 4) return false;
  const vs = [...new Set(cards.map((c) => c.value))];
  if (vs.length !== 4) return false;
  const asc = [...vs].sort((a, b) => a - b);
  const run = (xs: number[]) => xs.every((v, i) => i === 0 || v === xs[i - 1] + 1);
  if (run(asc)) return true;
  // A を 14 として見た並び (J-Q-K-A)。
  const high = [...vs.map((v) => (v === 1 ? 14 : v))].sort((a, b) => a - b);
  return run(high);
}
