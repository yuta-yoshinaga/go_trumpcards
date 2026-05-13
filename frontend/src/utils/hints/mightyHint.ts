import type { Card, MightyResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { MightyPhase } from '../../types/phases';

/** Maps numeric trump suit to card design string. */
const SUIT_NUM_TO_DESIGN: Readonly<Record<number, Card['design']>> = {
  1: 'SPADE',
  2: 'CLOVER',
  3: 'HEART',
  4: 'DIAMOND',
};

/** Picture cards (face cards) counted as point cards. */
const POINT_VALUES: ReadonlySet<number> = new Set([1, 10, 11, 12, 13]);
/** High card threshold for bid estimation. */
const HIGH_CARD_VALUE = 11;
/** Minimum high cards for strong bid confidence. */
const STRONG_BID_THRESHOLD = 4;
/** Minimum suit count for strong trump declaration. */
const STRONG_SUIT_THRESHOLD = 4;

/** Returns a frontend HintResult for Mighty, or null if no suggestion. */
export function getMightyHint(state: MightyResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);

  if (state.phase === MightyPhase.BID) {
    if (state.bidPlayerIdx !== humanIdx) return null;
    return getBidHint(human.cards);
  }

  if (state.phase === MightyPhase.TRUMP_AND_FRIEND) {
    if (state.declarerIdx !== humanIdx) return null;
    return getTrumpAndFriendHint(human.cards);
  }

  if (state.phase === MightyPhase.KITTY_EXCHANGE) {
    if (state.declarerIdx !== humanIdx) return null;
    return getKittyExchangeHint();
  }

  if (state.phase === MightyPhase.PLAY) {
    if (state.currentPlayerIdx !== humanIdx) return null;
    return getPlayHint(human.cards, state);
  }

  return null;
}

/** Estimate bid from high cards and joker/mighty holdings. */
function getBidHint(cards: Card[]): HintResult {
  const highCards = cards.filter((c) => c.value >= HIGH_CARD_VALUE).length;
  const hasJoker = cards.some((c) => c.design === 'JOKER');
  const hasMighty = cards.some((c) => c.design === 'SPADE' && c.value === 1);

  if (hasJoker && hasMighty) {
    return { targetAction: 'bidNoTrump', reason: 'hint.bidNoTrump', confidence: 'strong' };
  }
  if (highCards >= STRONG_BID_THRESHOLD) {
    return { targetAction: 'bid:high', reason: 'hint.bidStrong', confidence: 'strong' };
  }
  return { targetAction: 'bid:low', reason: 'hint.bidModerate', confidence: 'moderate' };
}

/** Suggest the longest suit as trump. */
function getTrumpAndFriendHint(cards: Card[]): HintResult {
  const suitCounts: Record<string, number> = {};
  for (const c of cards) {
    if (c.design === 'JOKER') continue;
    suitCounts[c.design] = (suitCounts[c.design] ?? 0) + 1;
  }
  const bestCount = Object.values(suitCounts).reduce((a, b) => (a > b ? a : b), 0);
  const confidence = bestCount >= STRONG_SUIT_THRESHOLD ? 'strong' : 'moderate';
  return { targetAction: 'declareTrump', reason: 'hint.declareTrump', confidence };
}

/** Suggest discarding three weakest non-trump non-point cards. */
function getKittyExchangeHint(): HintResult {
  return { targetAction: 'discard', reason: 'hint.discardWeakest', confidence: 'strong' };
}

/** Hint for play phase. */
function getPlayHint(cards: Card[], state: MightyResponse): HintResult {
  const trumpDesign = SUIT_NUM_TO_DESIGN[state.trumpSuit];
  const trick = state.currentTrick;

  // Leading the trick
  if (trick.length === 0) {
    const hasJoker = cards.some((c) => c.design === 'JOKER');
    if (hasJoker) {
      return { targetAction: 'play', reason: 'hint.leadJoker', confidence: 'moderate' };
    }
    const hasHighCard = cards.some((c) => c.value >= HIGH_CARD_VALUE || POINT_VALUES.has(c.value));
    if (hasHighCard) {
      return { targetAction: 'play', reason: 'hint.leadStrong', confidence: 'strong' };
    }
    return { targetAction: 'play', reason: 'hint.leadLow', confidence: 'moderate' };
  }

  // Following suit
  const ledCard = trick[0].card;
  const ledSuit = ledCard.design;
  const suitCards = cards.filter((c) => c.design === ledSuit);

  // Mighty (♠A or ♦A when spades trump) is a winning shot
  const hasMighty = cards.some(
    (c) =>
      (state.trumpSuit !== 1 && c.design === 'SPADE' && c.value === 1) ||
      (state.trumpSuit === 1 && c.design === 'DIAMOND' && c.value === 1),
  );

  if (suitCards.length > 0) {
    return { targetAction: 'play', reason: 'hint.followSuit', confidence: 'strong' };
  }

  // Void in led suit: check for joker, trump, or discard
  const hasJoker = cards.some((c) => c.design === 'JOKER');
  if (hasJoker) {
    return { targetAction: 'play', reason: 'hint.playJoker', confidence: 'strong' };
  }

  if (hasMighty) {
    return { targetAction: 'play', reason: 'hint.playMighty', confidence: 'strong' };
  }

  const hasTrump = cards.some((c) => c.design === trumpDesign);
  if (hasTrump) {
    return { targetAction: 'play', reason: 'hint.trumpCut', confidence: 'moderate' };
  }

  return { targetAction: 'play', reason: 'hint.discardLow', confidence: 'moderate' };
}
