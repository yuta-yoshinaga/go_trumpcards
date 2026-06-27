import type { Card } from '../types/card';

/**
 * Chinese Poker foul detection, ported 1:1 from the Go domain so the client-side
 * SET_HANDS warning agrees with the server's authoritative `cpValidateHands`.
 *
 * A "foul" is any arrangement that violates back >= middle >= front. The server
 * evaluates the two 5-card rows (middle, back) with a category-only ranker plus a
 * high-card tiebreak, and the 3-card front row with a dedicated 3-card ranker that
 * is mapped onto the 5-card rank space before comparison. This module mirrors each
 * of those helpers exactly (`internal/domain/hand_eval.go`,
 * `internal/domain/ChinesePoker.go`, `internal/domain/HoldemPlayer.go`).
 */

// 5-card poker hand rank constants (mirror internal/domain/poker_hand_rank.go).
const POKER_HIGH_CARD = 0;
const POKER_ONE_PAIR = 1;
const POKER_TWO_PAIR = 2;
const POKER_THREE_OF_A_KIND = 3;
const POKER_STRAIGHT = 4;
const POKER_FLUSH = 5;
const POKER_FULL_HOUSE = 6;
const POKER_FOUR_OF_A_KIND = 7;
const POKER_STRAIGHT_FLUSH = 8;
const POKER_ROYAL_FLUSH = 9;

// 3-card poker hand rank constants (mirror internal/domain/hand_eval.go).
// Note: in 3-card poker, Three of a Kind ranks ABOVE Straight.
const THREE_HIGH_CARD = 1;
const THREE_PAIR = 2;
const THREE_FLUSH = 3;
const THREE_STRAIGHT = 4;
const THREE_OF_A_KIND = 5;
const THREE_STRAIGHT_FLUSH = 6;

function sortedValuesAsc(cards: readonly Card[]): number[] {
  return cards.map((c) => c.value).sort((a, b) => a - b);
}

/** Mirror of `checkStraightValues` (sorted ascending; wheel and broadway count). */
function checkStraightValues(sorted: readonly number[]): boolean {
  if (sorted[0] === 1 && sorted[1] === 2 && sorted[2] === 3 && sorted[3] === 4 && sorted[4] === 5) {
    return true;
  }
  if (sorted[0] === 1 && sorted[1] === 10 && sorted[2] === 11 && sorted[3] === 12 && sorted[4] === 13) {
    return true;
  }
  for (let i = 1; i < sorted.length; i += 1) {
    if (sorted[i] !== sorted[i - 1] + 1) return false;
  }
  return true;
}

function checkRoyalStraightValues(sorted: readonly number[]): boolean {
  return (
    sorted.length === 5 &&
    sorted[0] === 1 &&
    sorted[1] === 10 &&
    sorted[2] === 11 &&
    sorted[3] === 12 &&
    sorted[4] === 13
  );
}

/** Mirror of `evalFiveCardHand`: returns a POKER_* category (no joker/five-of-a-kind). */
export function cpEvalFiveCardHand(cards: readonly Card[]): number {
  if (cards.length !== 5) return POKER_HIGH_CARD;
  const values = sortedValuesAsc(cards);
  const designs = cards.map((c) => c.design);

  const isFlush = designs.every((d) => d === designs[0]);
  const isStraight = checkStraightValues(values);

  const valueCounts = new Map<number, number>();
  for (const v of values) valueCounts.set(v, (valueCounts.get(v) ?? 0) + 1);
  const counts = Array.from(valueCounts.values()).sort((a, b) => b - a);

  if (isFlush && isStraight) {
    return checkRoyalStraightValues(values) ? POKER_ROYAL_FLUSH : POKER_STRAIGHT_FLUSH;
  }
  if (counts[0] === 4) return POKER_FOUR_OF_A_KIND;
  if (counts.length >= 2 && counts[0] === 3 && counts[1] === 2) return POKER_FULL_HOUSE;
  if (isFlush) return POKER_FLUSH;
  if (isStraight) return POKER_STRAIGHT;
  if (counts[0] === 3) return POKER_THREE_OF_A_KIND;
  if (counts.length >= 2 && counts[0] === 2 && counts[1] === 2) return POKER_TWO_PAIR;
  if (counts[0] === 2) return POKER_ONE_PAIR;
  return POKER_HIGH_CARD;
}

/** Mirror of `checkThreeCardStraight` (A-2-3 and Q-K-A wrap, plus normal runs). */
function checkThreeCardStraight(sorted: readonly number[]): boolean {
  if (sorted.length !== 3) return false;
  if (sorted[0] === 1 && sorted[1] === 2 && sorted[2] === 3) return true;
  if (sorted[0] === 1 && sorted[1] === 12 && sorted[2] === 13) return true;
  return sorted[1] === sorted[0] + 1 && sorted[2] === sorted[1] + 1;
}

/** Mirror of `evalThreeCardHand`: returns a THREE_* rank. */
export function cpEvalThreeCardHand(cards: readonly Card[]): number {
  if (cards.length !== 3) return THREE_HIGH_CARD;
  const values = sortedValuesAsc(cards);
  const designs = cards.map((c) => c.design);

  const isFlush = designs[0] === designs[1] && designs[1] === designs[2];
  const isStraight = checkThreeCardStraight(values);

  const valueCounts = new Map<number, number>();
  for (const v of values) valueCounts.set(v, (valueCounts.get(v) ?? 0) + 1);

  if (isFlush && isStraight) return THREE_STRAIGHT_FLUSH;
  for (const cnt of valueCounts.values()) {
    if (cnt === 3) return THREE_OF_A_KIND;
  }
  if (isStraight) return THREE_STRAIGHT;
  if (isFlush) return THREE_FLUSH;
  for (const cnt of valueCounts.values()) {
    if (cnt === 2) return THREE_PAIR;
  }
  return THREE_HIGH_CARD;
}

/** Mirror of `threeCardHandHighValues` / `cpFiveCardHighValues`: raw values, Ace=14, desc. */
function highValuesDesc(cards: readonly Card[]): number[] {
  return cards.map((c) => (c.value === 1 ? 14 : c.value)).sort((a, b) => b - a);
}

/** Mirror of `isWheelHand`: 5-card A-2-3-4-5. */
function isWheelHand(cards: readonly Card[]): boolean {
  if (cards.length !== 5) return false;
  const v = sortedValuesAsc(cards);
  return v[0] === 1 && v[1] === 2 && v[2] === 3 && v[3] === 4 && v[4] === 5;
}

/** Mirror of `tieBreakValues`: unique values sorted by (frequency desc, value desc). */
function tieBreakValues(vals: readonly number[]): number[] {
  const freq = new Map<number, number>();
  for (const v of vals) freq.set(v, (freq.get(v) ?? 0) + 1);
  const unique = Array.from(freq.keys());
  unique.sort((a, b) => {
    const fa = freq.get(a) ?? 0;
    const fb = freq.get(b) ?? 0;
    if (fa !== fb) return fb - fa;
    return b - a;
  });
  return unique;
}

/** Mirror of `compareHighCardsSlice`: a>b → 1, a<b → -1, equal → 0. */
function compareHighCardsSlice(a: readonly Card[], b: readonly Card[]): number {
  if (a.length === 0 || b.length === 0) return 0;
  const aWheel = isWheelHand(a);
  const bWheel = isWheelHand(b);
  const aVals = a.map((c) => (c.value === 1 && !aWheel ? 14 : c.value));
  const bVals = b.map((c) => (c.value === 1 && !bWheel ? 14 : c.value));
  const aTB = tieBreakValues(aVals);
  const bTB = tieBreakValues(bVals);
  for (let i = 0; i < aTB.length && i < bTB.length; i += 1) {
    if (aTB[i] > bTB[i]) return 1;
    if (aTB[i] < bTB[i]) return -1;
  }
  return 0;
}

/** Mirror of `cpMapThreeCardToFiveCardRank`. */
function mapThreeCardToFiveCardRank(threeCardRank: number): number {
  switch (threeCardRank) {
    case THREE_HIGH_CARD:
      return POKER_HIGH_CARD;
    case THREE_PAIR:
      return POKER_ONE_PAIR;
    case THREE_FLUSH:
      return POKER_TWO_PAIR;
    case THREE_STRAIGHT:
      return POKER_TWO_PAIR;
    case THREE_OF_A_KIND:
      return POKER_THREE_OF_A_KIND;
    case THREE_STRAIGHT_FLUSH:
      return POKER_STRAIGHT;
    default:
      return POKER_HIGH_CARD;
  }
}

/** Mirror of `cpFrontNotStrongerThanMiddle`: true when front does NOT beat middle. */
function frontNotStrongerThanMiddle(front: readonly Card[], middleRank: number, middle: readonly Card[]): boolean {
  const cpFront = mapThreeCardToFiveCardRank(cpEvalThreeCardHand(front));
  if (cpFront < middleRank) return true;
  if (cpFront > middleRank) return false;
  const frontVals = highValuesDesc(front);
  const middleVals = highValuesDesc(middle);
  for (let i = 0; i < frontVals.length && i < middleVals.length; i += 1) {
    if (frontVals[i] > middleVals[i]) return false;
    if (frontVals[i] < middleVals[i]) return true;
  }
  return true;
}

/**
 * Returns `true` when the given arrangement is a foul (violates back >= middle >= front).
 * Mirrors the server's `cpValidateHands` (negated). Rows of the wrong length are
 * treated as non-foul (incomplete) so the warning only fires on a complete layout.
 */
export function chinesePokerIsFoul(front: readonly Card[], middle: readonly Card[], back: readonly Card[]): boolean {
  if (front.length !== 3 || middle.length !== 5 || back.length !== 5) return false;

  const middleRank = cpEvalFiveCardHand(middle);
  const backRank = cpEvalFiveCardHand(back);

  if (backRank < middleRank) return true;
  if (backRank === middleRank && compareHighCardsSlice(back, middle) < 0) return true;

  return !frontNotStrongerThanMiddle(front, middleRank, middle);
}
