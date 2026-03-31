import type { Card, EuchreResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { EuchrePhase } from '../../types/phases';

/** Maps numeric trump suit to card design string. */
const SUIT_NUM_TO_DESIGN: Readonly<Record<number, Card['design']>> = {
  1: 'SPADE',
  2: 'CLOVER',
  3: 'HEART',
  4: 'DIAMOND',
};

/** Minimum trump cards to recommend order-up or lead with trump. */
const STRONG_TRUMP_THRESHOLD = 2;
/** Minimum suit count to suggest calling with strong confidence. */
const STRONG_SUIT_THRESHOLD = 3;

/** Returns a frontend HintResult for Euchre, or null if no suggestion. */
export function getEuchreHint(state: EuchreResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);

  if (state.phase === EuchrePhase.PICK_UP) {
    if (state.bidPlayerIdx !== humanIdx) return null;
    return getPickUpHint(human.cards, state);
  }

  if (state.phase === EuchrePhase.CALL_TRUMP) {
    if (state.bidPlayerIdx !== humanIdx) return null;
    return getCallTrumpHint(human.cards);
  }

  if (state.phase === EuchrePhase.DISCARD) {
    if (state.dealerIdx !== humanIdx) return null;
    return getDiscardHint(human.cards, state.trumpSuit);
  }

  if (state.phase === EuchrePhase.PLAY) {
    if (state.currentPlayerIdx !== humanIdx) return null;
    return getPlayHint(human.cards, state);
  }

  return null;
}

/** Hint for pick-up phase: order up if hand has trump-suit cards. */
function getPickUpHint(cards: Card[], state: EuchreResponse): HintResult {
  const trumpDesign = state.faceUpCard?.design;
  const trumpCount = trumpDesign ? cards.filter((c) => c.design === trumpDesign).length : 0;

  if (trumpCount >= STRONG_TRUMP_THRESHOLD) {
    return { targetAction: 'orderUp', reason: 'hint.orderUpStrong', confidence: 'strong' };
  }
  return { targetAction: 'pass', reason: 'hint.passWeak', confidence: 'moderate' };
}

/** Hint for call trump phase: suggest suit with most cards. */
function getCallTrumpHint(cards: Card[]): HintResult {
  const suitCounts: Record<string, number> = {};
  for (const c of cards) {
    suitCounts[c.design] = (suitCounts[c.design] ?? 0) + 1;
  }
  const bestCount = Math.max(...Object.values(suitCounts));
  const confidence = bestCount >= STRONG_SUIT_THRESHOLD ? 'strong' : 'moderate';
  return { targetAction: 'callTrump', reason: 'hint.callStrongSuit', confidence };
}

/** Hint for discard phase: discard weakest non-trump card. */
function getDiscardHint(_cards: Card[], _trumpSuit: number): HintResult {
  return { targetAction: 'discard', reason: 'hint.discardWeakest', confidence: 'strong' };
}

/** Hint for play phase: follow suit, lead, or trump cut. */
function getPlayHint(cards: Card[], state: EuchreResponse): HintResult {
  const trumpDesign = SUIT_NUM_TO_DESIGN[state.trumpSuit];
  const trick = state.currentTrick;

  // Leading the trick
  if (trick.length === 0) {
    const trumpCards = cards.filter((c) => c.design === trumpDesign);
    if (trumpCards.length >= STRONG_TRUMP_THRESHOLD) {
      return { targetAction: 'play', reason: 'hint.leadTrump', confidence: 'strong' };
    }
    return { targetAction: 'play', reason: 'hint.leadOffSuit', confidence: 'moderate' };
  }

  // Following suit
  const ledSuit = trick[0].card.design;
  const suitCards = cards.filter((c) => c.design === ledSuit);

  if (suitCards.length > 0) {
    return { targetAction: 'play', reason: 'hint.followSuit', confidence: 'strong' };
  }

  // Void in led suit: trump cut if possible
  const hasTrump = cards.some((c) => c.design === trumpDesign);
  if (hasTrump) {
    return { targetAction: 'play', reason: 'hint.trumpCut', confidence: 'moderate' };
  }

  return { targetAction: 'play', reason: 'hint.discardWeakest', confidence: 'moderate' };
}
