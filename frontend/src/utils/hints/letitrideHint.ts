import type { Card, LetItRideResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { LetItRidePhase } from '../../types/phases';

/** PokerHandThreeOfAKind = 3 (sync: internal/domain/PokerPlayer.go). */
const RANK_THREE_OF_A_KIND = 3;

/** PokerHandOnePair = 1 (sync: internal/domain/PokerPlayer.go). */
const RANK_ONE_PAIR = 1;

/** Minimum card value for a qualifying pair (10). */
const TENS_VALUE = 10;

/**
 * Returns a Let It Ride hint for the decision phases.
 * Basic strategy: Let It Ride with three of a kind or better,
 * a pair of 10s or better, or three to a royal flush.
 * Otherwise, pull.
 */
export function getLetitrideHint(state: LetItRideResponse): HintResult | null {
  if (state.phase !== LetItRidePhase.FIRST_DECISION && state.phase !== LetItRidePhase.SECOND_DECISION) {
    return null;
  }
  if (!state.playerHand || state.playerHand.length === 0) return null;

  // Three of a kind or better — always let it ride
  if (state.handRank >= RANK_THREE_OF_A_KIND) {
    return { targetAction: 'letitride', reason: 'hint.strongHand', confidence: 'strong' };
  }

  // Pair of 10s or better
  if (state.handRank >= RANK_ONE_PAIR && hasPairTensOrBetter(state.playerHand)) {
    return { targetAction: 'letitride', reason: 'hint.pairTensOrBetter', confidence: 'moderate' };
  }

  // Three to a royal flush (10, J, Q, K, A of same suit)
  if (hasThreeToRoyalFlush(state.playerHand)) {
    return { targetAction: 'letitride', reason: 'hint.threeToRoyalFlush', confidence: 'moderate' };
  }

  return { targetAction: 'pull', reason: 'hint.weakHand', confidence: 'moderate' };
}

/** Check if the hand contains a pair of 10s or better. */
function hasPairTensOrBetter(cards: Card[]): boolean {
  const valueCounts = new Map<number, number>();
  for (const c of cards) {
    valueCounts.set(c.value, (valueCounts.get(c.value) ?? 0) + 1);
  }
  for (const [value, count] of valueCounts) {
    // Ace (value=1) counts as high
    if (count >= 2 && (value >= TENS_VALUE || value === 1)) {
      return true;
    }
  }
  return false;
}

/** Check if the 3 player cards are three to a royal flush (same suit, all in {1,10,11,12,13}). */
function hasThreeToRoyalFlush(cards: Card[]): boolean {
  if (cards.length < 3) return false;
  const royalValues = new Set([1, 10, 11, 12, 13]);
  const allRoyal = cards.every((c) => royalValues.has(c.value));
  if (!allRoyal) return false;
  const suit = cards[0].design;
  return cards.every((c) => c.design === suit);
}
