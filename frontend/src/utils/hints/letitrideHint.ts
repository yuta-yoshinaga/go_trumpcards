import type { Card, LetItRideResponse } from '../../types/card';
import { isMaskedCard } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { LetItRidePhase } from '../../types/phases';

/** Minimum card value for a qualifying pair (10). */
const TENS_VALUE = 10;

/**
 * Returns a Let It Ride hint for the decision phases.
 * Basic strategy: Let It Ride with three of a kind or better,
 * a pair of 10s or better, or three to a royal flush.
 * In SECOND_DECISION the first revealed community card is included in the evaluation.
 * Otherwise, pull.
 *
 * Note: `state.handRank` is 0 during decision phases (the server only computes it at
 * resolution), so hand quality is evaluated directly from the available cards.
 */
export function getLetitrideHint(state: LetItRideResponse): HintResult | null {
  if (state.phase !== LetItRidePhase.FIRST_DECISION && state.phase !== LetItRidePhase.SECOND_DECISION) {
    return null;
  }
  if (!state.playerHand || state.playerHand.length === 0) return null;

  // Build available cards: player hand + any revealed community cards.
  const revealedCommunity = state.communityCards.filter((c): c is Card => !isMaskedCard(c));
  const availableCards: Card[] = [...state.playerHand, ...revealedCommunity];

  // Three of a kind or better — always let it ride
  if (hasThreeOfAKindOrBetter(availableCards)) {
    return { targetAction: 'letitride', reason: 'hint.strongHand', confidence: 'strong' };
  }

  // Pair of 10s or better
  if (hasPairTensOrBetter(availableCards)) {
    return { targetAction: 'letitride', reason: 'hint.pairTensOrBetter', confidence: 'moderate' };
  }

  // Three or more to a royal flush (10, J, Q, K, A of same suit)
  if (hasRoyalFlushDraw(availableCards)) {
    return { targetAction: 'letitride', reason: 'hint.threeToRoyalFlush', confidence: 'moderate' };
  }

  return { targetAction: 'pull', reason: 'hint.weakHand', confidence: 'moderate' };
}

/** Check if the available cards include three of a kind or better. */
function hasThreeOfAKindOrBetter(cards: Card[]): boolean {
  const valueCounts = new Map<number, number>();
  for (const c of cards) {
    valueCounts.set(c.value, (valueCounts.get(c.value) ?? 0) + 1);
  }
  for (const count of valueCounts.values()) {
    if (count >= 3) return true;
  }
  return false;
}

/** Check if the available cards contain a pair of 10s or better. */
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

/** Check if at least 3 of the available cards are the same suit and all royal values (A/10/J/Q/K). */
function hasRoyalFlushDraw(cards: Card[]): boolean {
  const royalValues = new Set([1, 10, 11, 12, 13]);
  const royalCards = cards.filter((c) => royalValues.has(c.value));
  if (royalCards.length < 3) return false;
  const bySuit = new Map<string, number>();
  for (const c of royalCards) {
    bySuit.set(c.design, (bySuit.get(c.design) ?? 0) + 1);
  }
  for (const count of bySuit.values()) {
    if (count >= 3) return true;
  }
  return false;
}
