import type { BeloteResponse, Card } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { BelotePhase } from '../../types/phases';

/** Maps numeric trump suit to card design string. */
const SUIT_NUM_TO_DESIGN: Readonly<Record<number, Card['design']>> = {
  1: 'SPADE',
  2: 'CLOVER',
  3: 'HEART',
  4: 'DIAMOND',
};

/** Card-value rank score used to estimate trump-suit strength (higher = stronger). */
const TRUMP_VALUE_SCORE: Readonly<Record<number, number>> = {
  11: 14, // J — highest trump
  9: 10,
  1: 7, // A
  10: 5,
  13: 3, // K
  12: 2, // Q
  8: 0,
  7: 0,
};

const PICKUP_TAKE_THRESHOLD = 18;
const CALL_TRUMP_THRESHOLD = 22;

/** Returns a frontend HintResult for Belote, or null if no suggestion. */
export function getBeloteHint(state: BeloteResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);

  if (state.phase === BelotePhase.BID_PICK_UP) {
    if (state.bidPlayerIdx !== humanIdx) return null;
    return getPickUpHint(human.cards, state);
  }

  if (state.phase === BelotePhase.BID_CALL_TRUMP) {
    if (state.bidPlayerIdx !== humanIdx) return null;
    return getCallTrumpHint(human.cards);
  }

  if (state.phase === BelotePhase.PLAY) {
    if (state.currentPlayerIdx !== humanIdx) return null;
    return getPlayHint(human.cards, state);
  }

  return null;
}

/** Hint for pick-up phase: take if trump-suit cards have enough strength. */
function getPickUpHint(cards: Card[], state: BeloteResponse): HintResult {
  const trumpDesign = state.faceUpCard?.design;
  if (!trumpDesign) {
    return { targetAction: 'pass', reason: 'hint.passWeak', confidence: 'moderate' };
  }
  const score = cards
    .filter((c) => c.design === trumpDesign)
    .reduce((acc, c) => acc + (TRUMP_VALUE_SCORE[c.value] ?? 0), 0);

  if (score >= PICKUP_TAKE_THRESHOLD) {
    return { targetAction: 'orderUp', reason: 'hint.orderUpStrong', confidence: 'strong' };
  }
  return { targetAction: 'pass', reason: 'hint.passWeak', confidence: 'moderate' };
}

/** Hint for call-trump phase: pick the suit with the highest combined trump score. */
function getCallTrumpHint(cards: Card[]): HintResult {
  let bestScore = 0;
  for (const design of Object.values(SUIT_NUM_TO_DESIGN)) {
    const score = cards
      .filter((c) => c.design === design)
      .reduce((acc, c) => acc + (TRUMP_VALUE_SCORE[c.value] ?? 0), 0);
    if (score > bestScore) bestScore = score;
  }
  if (bestScore >= CALL_TRUMP_THRESHOLD) {
    return { targetAction: 'callTrump', reason: 'hint.callStrongSuit', confidence: 'strong' };
  }
  return { targetAction: 'pass', reason: 'hint.passWeak', confidence: 'moderate' };
}

/** Hint for play phase: follow suit, lead, or trump cut. */
function getPlayHint(cards: Card[], state: BeloteResponse): HintResult {
  const trumpDesign = SUIT_NUM_TO_DESIGN[state.trumpSuit];
  const trick = state.currentTrick;

  if (trick.length === 0) {
    const hasTrump = cards.some((c) => c.design === trumpDesign);
    if (hasTrump) {
      return { targetAction: 'play', reason: 'hint.leadTrump', confidence: 'moderate' };
    }
    return { targetAction: 'play', reason: 'hint.leadOffSuit', confidence: 'moderate' };
  }

  const ledSuit = trick[0].card.design;
  const hasLed = cards.some((c) => c.design === ledSuit);
  if (hasLed) {
    return { targetAction: 'play', reason: 'hint.followSuit', confidence: 'strong' };
  }

  const hasTrump = cards.some((c) => c.design === trumpDesign);
  if (hasTrump) {
    return { targetAction: 'play', reason: 'hint.trumpCut', confidence: 'moderate' };
  }

  return { targetAction: 'play', reason: 'hint.discardWeakest', confidence: 'moderate' };
}
