import type { Card } from '../types/card';

/**
 * Joker Poker made-hand evaluation, ported 1:1 from the Go domain
 * (`evalWildHand` in `hand_eval_wild.go` + `jokerPokerGetResult` in
 * `VideoPokerVariant.go`). It lets the draw-phase UI show the player's current
 * best paying hand in real time, so a beginner can see whether the five held +
 * dealt cards already qualify for the "Kings or Better" pay table.
 *
 * The Joker (`design === 'JOKER'`) is wild: it is substituted with every one of
 * the 52 standard cards and the best resulting rank is kept, exactly matching
 * the server's enumeration.
 */

/** Rank ints mirror the Go `PokerHand*` constants (High Card = 0 … Five of a Kind = 10). */
const HIGH_CARD = 0;
const ONE_PAIR = 1;
const TWO_PAIR = 2;
const THREE_OF_A_KIND = 3;
const STRAIGHT = 4;
const FLUSH = 5;
const FULL_HOUSE = 6;
const FOUR_OF_A_KIND = 7;
const STRAIGHT_FLUSH = 8;
const ROYAL_FLUSH = 9;
const FIVE_OF_A_KIND = 10;

/** Ace rank value in this codebase (A = 1, J = 11, Q = 12, K = 13). */
const ACE = 1;
/** King rank value. */
const KING = 13;

/** The four standard suits a wild card can be substituted with. */
const SUITS: Card['design'][] = ['SPADE', 'CLOVER', 'HEART', 'DIAMOND'];

/** The player's current best made hand during the Joker Poker draw phase. */
export interface JokerPokerMadeHand {
  /**
   * Paytable row key (matching `videoPokerPayoutRows('jokerpoker')`, e.g.
   * `'fiveOfAKind'`, `'wildRoyalFlush'`, `'kingsOrBetter'`) when the hand pays,
   * or `null` when it is below the Kings-or-Better minimum (no payout).
   */
  rowKey: string | null;
}

/**
 * Evaluate the five current cards and return the paying paytable row key, or a
 * `null` row key when the hand does not reach the Kings-or-Better minimum.
 * Returns `null` for any non-5-card input so callers can hide the readout until
 * a full hand is on the table.
 */
export function evaluateJokerPokerMadeHand(cards: readonly Card[]): JokerPokerMadeHand | null {
  if (cards.length !== 5) return null;
  const { rank, usedWilds } = evalWildHand(cards);
  return { rowKey: jokerPokerRowKey(cards, rank, usedWilds) };
}

/** Map a rank + wild-usage to the Joker Poker paytable row key (or null if it pays nothing). */
function jokerPokerRowKey(cards: readonly Card[], rank: number, usedWilds: boolean): string | null {
  if (rank === ROYAL_FLUSH) return usedWilds ? 'wildRoyalFlush' : 'naturalRoyalFlush';
  switch (rank) {
    case FIVE_OF_A_KIND:
      return 'fiveOfAKind';
    case STRAIGHT_FLUSH:
      return 'straightFlush';
    case FOUR_OF_A_KIND:
      return 'fourOfAKind';
    case FULL_HOUSE:
      return 'fullHouse';
    case FLUSH:
      return 'flush';
    case STRAIGHT:
      return 'straight';
    case THREE_OF_A_KIND:
      return 'threeOfAKind';
    case TWO_PAIR:
      return 'twoPair';
    case ONE_PAIR:
      return isKingsOrBetter(cards) ? 'kingsOrBetter' : null;
    default:
      return null;
  }
}

/** Whether the (non-joker) pair is aces or kings — the Joker Poker pay minimum. */
function isKingsOrBetter(cards: readonly Card[]): boolean {
  const counts = new Map<number, number>();
  for (const c of cards) {
    if (c.design === 'JOKER') continue;
    counts.set(c.value, (counts.get(c.value) ?? 0) + 1);
  }
  for (const [value, count] of counts) {
    if (count >= 2 && (value === ACE || value === KING)) return true;
  }
  return false;
}

/**
 * Evaluate a 5-card hand that may contain wild Jokers, returning the best rank
 * and whether any wild was used to form it. Substitutes each Joker with all 52
 * standard cards and keeps the best rank (mirrors the Go `evalWildHand`).
 */
function evalWildHand(cards: readonly Card[]): { rank: number; usedWilds: boolean } {
  const nonWilds = cards.filter((c) => c.design !== 'JOKER');
  const numWilds = cards.length - nonWilds.length;

  if (numWilds === 0) return { rank: evalFiveCardHand(cards), usedWilds: false };
  // Four or five wilds can always make Five of a Kind.
  if (numWilds >= 4) return { rank: FIVE_OF_A_KIND, usedWilds: true };

  let best = HIGH_CARD;
  const subs: Card[] = [];
  const enumerate = (depth: number): void => {
    if (best === ROYAL_FLUSH) return; // cannot improve beyond a Royal Flush
    if (depth === numWilds) {
      const rank = evalFiveCardHand([...nonWilds, ...subs]);
      if (rank > best) best = rank;
      return;
    }
    for (const suit of SUITS) {
      for (let value = 1; value <= 13; value++) {
        subs[depth] = { design: suit, value };
        enumerate(depth + 1);
        if (best === ROYAL_FLUSH) return;
      }
    }
  };
  enumerate(0);
  return { rank: best, usedWilds: true };
}

/**
 * Evaluate a standard (no-wild) 5-card hand, including Five of a Kind (only
 * reachable once wilds have been substituted). Aces are low (1) for the
 * A-2-3-4-5 wheel and the high anchor of 10-J-Q-K-A.
 */
function evalFiveCardHand(cards: readonly Card[]): number {
  const values = cards.map((c) => c.value).sort((a, b) => a - b);
  const designs = cards.map((c) => c.design);
  const groupSizes = Object.values(countByValue(values)).sort((a, b) => b - a);
  const isFlush = designs.every((d) => d === designs[0]);
  const isStraight = checkStraight(values);
  const isRoyal = isStraight && values[0] === 1 && values[4] === 13 && values[1] === 10;

  if (groupSizes[0] === 5) return FIVE_OF_A_KIND;
  if (isFlush && isRoyal) return ROYAL_FLUSH;
  if (isFlush && isStraight) return STRAIGHT_FLUSH;
  if (groupSizes[0] === 4) return FOUR_OF_A_KIND;
  if (groupSizes[0] === 3 && groupSizes[1] === 2) return FULL_HOUSE;
  if (isFlush) return FLUSH;
  if (isStraight) return STRAIGHT;
  if (groupSizes[0] === 3) return THREE_OF_A_KIND;
  if (groupSizes[0] === 2 && groupSizes[1] === 2) return TWO_PAIR;
  if (groupSizes[0] === 2) return ONE_PAIR;
  return HIGH_CARD;
}

function countByValue(values: readonly number[]): Record<number, number> {
  const out: Record<number, number> = {};
  for (const v of values) out[v] = (out[v] ?? 0) + 1;
  return out;
}

function checkStraight(sortedValues: readonly number[]): boolean {
  if (new Set(sortedValues).size !== 5) return false;
  if (sortedValues[4] - sortedValues[0] === 4) return true;
  // 10-J-Q-K-A sorts to [1, 10, 11, 12, 13].
  return sortedValues[0] === 1 && sortedValues[1] === 10 && sortedValues[4] === 13;
}
