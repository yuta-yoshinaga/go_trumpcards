import type { Card, JassResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { JassPhase } from '../../types/phases';

/** Maps numeric trump suit to card design string. */
const SUIT_NUM_TO_DESIGN: Readonly<Record<number, Card['design']>> = {
  1: 'SPADE',
  2: 'CLOVER',
  3: 'HEART',
  4: 'DIAMOND',
};

/** Card-value rank score used to estimate trump-suit strength (higher = stronger). */
const TRUMP_VALUE_SCORE: Readonly<Record<number, number>> = {
  11: 14, // J — highest trump (Bauer)
  9: 10, // Nell
  1: 7, // A
  10: 5,
  13: 3, // K
  12: 2, // Q
  8: 0,
  7: 0,
  6: 0,
};

// Thresholds mirror the backend Normal-difficulty CPU heuristic
// (internal/domain/Jass.go cpuSelectTrump / cpuSchieben).
const CALL_TRUMP_THRESHOLD = 22;
const SCHIEBEN_THRESHOLD = 14;

/** Returns a frontend HintResult for Jass, or null if no suggestion. */
export function getJassHint(state: JassResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);

  if (state.phase === JassPhase.BID_TRUMP) {
    if (state.bidPlayerIdx !== humanIdx) return null;
    return getBidTrumpHint(human.cards);
  }

  if (state.phase === JassPhase.BID_PARTNER) {
    if (state.bidPlayerIdx !== humanIdx) return null;
    return getBidPartnerHint();
  }

  if (state.phase === JassPhase.PLAY) {
    if (state.currentPlayerIdx !== humanIdx) return null;
    return getPlayHint(human.cards, state);
  }

  return null;
}

/** Returns the best combined trump score across the four suits. */
function bestSuitScore(cards: Card[]): number {
  let best = 0;
  for (const design of Object.values(SUIT_NUM_TO_DESIGN)) {
    const score = cards
      .filter((c) => c.design === design)
      .reduce((acc, c) => acc + (TRUMP_VALUE_SCORE[c.value] ?? 0), 0);
    if (score > best) best = score;
  }
  return best;
}

/**
 * Hint for the trump-bid phase: choose the strongest suit if it clears the
 * threshold, otherwise schieben (pass the bid to the partner) when the hand is weak.
 */
function getBidTrumpHint(cards: Card[]): HintResult {
  const best = bestSuitScore(cards);
  if (best >= CALL_TRUMP_THRESHOLD) {
    return { targetAction: 'callTrump', reason: 'hint.strategicTrump', confidence: 'strong' };
  }
  if (best <= SCHIEBEN_THRESHOLD) {
    return { targetAction: 'schieben', reason: 'hint.schiebenRecommended', confidence: 'moderate' };
  }
  return { targetAction: 'callTrump', reason: 'hint.strategicTrump', confidence: 'moderate' };
}

/** Hint for the forced partner-bid phase: pick the strongest suit (no schieben allowed). */
function getBidPartnerHint(): HintResult {
  return { targetAction: 'callTrump', reason: 'hint.strategicTrump', confidence: 'moderate' };
}

/** Hint for play phase: follow suit, lead, or trump cut. */
function getPlayHint(cards: Card[], state: JassResponse): HintResult {
  const trumpDesign = SUIT_NUM_TO_DESIGN[state.trumpSuit];
  const trick = state.currentTrick;

  if (trick.length === 0) {
    const hasTrump = cards.some((c) => c.design === trumpDesign);
    if (hasTrump) {
      return { targetAction: 'play', reason: 'hint.leadTrump', confidence: 'moderate' };
    }
    return { targetAction: 'play', reason: 'hint.leadStrong', confidence: 'moderate' };
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

  return { targetAction: 'play', reason: 'hint.discardWeak', confidence: 'moderate' };
}
