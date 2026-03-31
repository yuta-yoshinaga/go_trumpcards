import type { Card, NapoleonResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { NapoleonPhase } from '../../types/phases';

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
/** Minimum suit count for strong trump declaration. */
const STRONG_SUIT_THRESHOLD = 3;

/** Returns a frontend HintResult for Napoleon, or null if no suggestion. */
export function getNapoleonHint(state: NapoleonResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);

  if (state.phase === NapoleonPhase.BID) {
    if (state.bidPlayerIdx !== humanIdx) return null;
    return getBidHint(human.cards);
  }

  if (state.phase === NapoleonPhase.TRUMP_DECLARATION) {
    if (state.napoleonIdx !== humanIdx) return null;
    return getTrumpDeclarationHint(human.cards);
  }

  if (state.phase === NapoleonPhase.KITTY_EXCHANGE) {
    if (state.napoleonIdx !== humanIdx) return null;
    return getKittyExchangeHint(human.cards, state.trumpSuit);
  }

  if (state.phase === NapoleonPhase.PLAY) {
    if (state.currentPlayerIdx !== humanIdx) return null;
    return getPlayHint(human.cards, state);
  }

  return null;
}

/** Estimate bid from high cards and suit distribution. */
function getBidHint(cards: Card[]): HintResult {
  const highCards = cards.filter((c) => c.value >= HIGH_CARD_VALUE).length;
  const estimatedTricks = Math.min(17, 12 + Math.round(highCards * 0.5));
  const confidence = highCards >= STRONG_BID_THRESHOLD ? 'strong' : 'moderate';
  const reason = confidence === 'strong' ? 'hint.bidStrong' : 'hint.bidModerate';
  return { targetAction: `bid:${estimatedTricks}`, reason, confidence };
}

/** Suggest the suit with the most high cards as trump. */
function getTrumpDeclarationHint(cards: Card[]): HintResult {
  const suitCounts: Record<string, number> = {};
  for (const c of cards) {
    if (c.design === 'JOKER') continue;
    suitCounts[c.design] = (suitCounts[c.design] ?? 0) + 1;
  }
  const bestCount = Math.max(...Object.values(suitCounts));
  const confidence = bestCount >= STRONG_SUIT_THRESHOLD ? 'strong' : 'moderate';
  return { targetAction: 'declareTrump', reason: 'hint.declareTrump', confidence };
}

/** Suggest discarding weakest non-trump card during kitty exchange. */
function getKittyExchangeHint(_cards: Card[], _trumpSuit: number): HintResult {
  return { targetAction: 'discard', reason: 'hint.discardWeakest', confidence: 'strong' };
}

/** Hint for play phase: follow suit, trump cut, lead, or joker. */
function getPlayHint(cards: Card[], state: NapoleonResponse): HintResult {
  const trumpDesign = SUIT_NUM_TO_DESIGN[state.trumpSuit];
  const trick = state.currentTrick;

  // Leading the trick
  if (trick.length === 0) {
    const hasHighCard = cards.some((c) => c.value >= HIGH_CARD_VALUE);
    if (hasHighCard) {
      return { targetAction: 'play', reason: 'hint.leadStrong', confidence: 'strong' };
    }
    return { targetAction: 'play', reason: 'hint.leadLow', confidence: 'moderate' };
  }

  // Following suit
  const ledSuit = trick[0].card.design;
  const suitCards = cards.filter((c) => c.design === ledSuit);

  if (suitCards.length > 0) {
    return { targetAction: 'play', reason: 'hint.followSuit', confidence: 'strong' };
  }

  // Void in led suit: check for joker, trump, or discard
  const hasJoker = cards.some((c) => c.design === 'JOKER');
  if (hasJoker) {
    return { targetAction: 'play', reason: 'hint.playJoker', confidence: 'strong' };
  }

  const hasTrump = cards.some((c) => c.design === trumpDesign);
  if (hasTrump) {
    return { targetAction: 'play', reason: 'hint.trumpCut', confidence: 'moderate' };
  }

  return { targetAction: 'play', reason: 'hint.discardLow', confidence: 'moderate' };
}
