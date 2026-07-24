import type { Card } from '../types/card';
import { evaluateBestHand, PokerHand, type PokerHandRank } from './pokerSquaresUtils';

const ACE_VALUE = 1;
/** Lowest pair rank that qualifies for the Mississippi Stud pay table (pair of 6s). */
const MIN_PAYING_PAIR = 6;

/** The player's current best made hand during a Mississippi Stud betting street. */
export interface MississippiStudMadeHand {
  /** Best made-hand category reachable from the currently-known cards. */
  rank: PokerHandRank;
  /**
   * Whether the made hand qualifies for the pay table (pair of 6s or higher, or
   * any hand above one pair). Low pairs (2–5) and high card do not qualify.
   */
  paytableEligible: boolean;
}

/**
 * Evaluate the player's current made hand from the known cards (two hole cards
 * plus any revealed community cards) during a Mississippi Stud street.
 *
 * Delegates hand-category detection to the shared {@link evaluateBestHand}
 * evaluator. Returns `null` when fewer than two cards are known or when nothing
 * better than a high card is made, so the caller can keep the board free of
 * clutter before a genuine made hand exists.
 */
export function evaluateMississippiStudMadeHand(cards: readonly Card[]): MississippiStudMadeHand | null {
  if (cards.length < 2) return null;
  const rank = evaluateBestHand(cards);
  if (rank === null || rank === PokerHand.HighCard) return null;
  if (rank > PokerHand.OnePair) return { rank, paytableEligible: true };
  return { rank, paytableEligible: hasPayingPair(cards) };
}

/** Report whether any pair among the cards is a paying pair (aces or 6s+). */
function hasPayingPair(cards: readonly Card[]): boolean {
  const counts = new Map<number, number>();
  for (const c of cards) counts.set(c.value, (counts.get(c.value) ?? 0) + 1);
  for (const [value, count] of counts) {
    if (count >= 2 && (value === ACE_VALUE || value >= MIN_PAYING_PAIR)) return true;
  }
  return false;
}
