import type { Card, OhHellResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { OhHellPhase } from '../../types/phases';

/** Maps numeric trump suit to card design string. */
const SUIT_NUM_TO_DESIGN: Readonly<Record<number, Card['design']>> = {
  1: 'SPADE',
  2: 'CLOVER',
  3: 'HEART',
  4: 'DIAMOND',
};

/** High card threshold for bid estimation. */
const HIGH_CARD_VALUE = 11;
/** Minimum high cards for strong bid confidence. */
const STRONG_BID_THRESHOLD = 3;

/** Returns a frontend HintResult for Oh Hell, or null if no suggestion. */
export function getOhHellHint(state: OhHellResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);

  if (state.phase === OhHellPhase.BID) {
    if (state.bidPlayerIdx !== humanIdx) return null;
    return getBidHint(human.cards, state);
  }

  if (state.phase === OhHellPhase.PLAY) {
    if (state.currentPlayerIdx !== humanIdx) return null;
    return getPlayHint(human.cards, state);
  }

  return null;
}

/** Estimate bid from high cards and trump count. */
function getBidHint(cards: Card[], state: OhHellResponse): HintResult {
  const trumpDesign = SUIT_NUM_TO_DESIGN[state.trumpSuit];
  const highCards = cards.filter((c) => c.value >= HIGH_CARD_VALUE).length;
  const trumpCount = cards.filter((c) => c.design === trumpDesign).length;
  let estimatedTricks = Math.max(0, Math.round(highCards * 0.6 + trumpCount * 0.4));

  // Adjust if estimated bid equals restricted bid
  if (state.restrictedBid >= 0 && estimatedTricks === state.restrictedBid) {
    estimatedTricks = estimatedTricks > 0 ? estimatedTricks - 1 : estimatedTricks + 1;
  }

  const confidence = highCards >= STRONG_BID_THRESHOLD ? 'strong' : 'moderate';
  return { targetAction: `bid:${estimatedTricks}`, reason: 'hint.bidEstimate', confidence };
}

/** Hint for play phase: follow suit, trump, or lead. */
function getPlayHint(cards: Card[], state: OhHellResponse): HintResult {
  const trumpDesign = SUIT_NUM_TO_DESIGN[state.trumpSuit];
  const trick = state.currentTrick;

  // Leading the trick
  if (trick.length === 0) {
    return { targetAction: 'play', reason: 'hint.leadStrategic', confidence: 'moderate' };
  }

  // Following suit
  const ledSuit = trick[0].card.design;
  const suitCards = cards.filter((c) => c.design === ledSuit);

  if (suitCards.length > 0) {
    return { targetAction: 'play', reason: 'hint.followSuit', confidence: 'strong' };
  }

  // Void in led suit: trump or discard
  const hasTrump = cards.some((c) => c.design === trumpDesign);
  if (hasTrump) {
    return { targetAction: 'play', reason: 'hint.trumpWithCard', confidence: 'moderate' };
  }

  return { targetAction: 'play', reason: 'hint.discardLowest', confidence: 'moderate' };
}
