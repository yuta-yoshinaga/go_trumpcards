import type { Card, OmiResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { OmiPhase } from '../../types/phases';

/** Maps numeric trump suit to card design string. */
const SUIT_NUM_TO_DESIGN: Readonly<Record<number, Card['design']>> = {
  1: 'SPADE',
  2: 'CLOVER',
  3: 'HEART',
  4: 'DIAMOND',
};

/** Minimum trump cards to recommend leading with trump. */
const STRONG_TRUMP_THRESHOLD = 2;
/** Minimum suit count to suggest calling with strong confidence. */
const STRONG_SUIT_THRESHOLD = 3;

/** Returns a frontend HintResult for Omi, or null if no suggestion. */
export function getOmiHint(state: OmiResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);

  if (state.phase === OmiPhase.CALL_TRUMP) {
    if (state.bidPlayerIdx !== humanIdx) return null;
    return getCallTrumpHint(human.cards);
  }

  if (state.phase === OmiPhase.PLAY) {
    if (state.currentPlayerIdx !== humanIdx) return null;
    return getPlayHint(human.cards, state);
  }

  return null;
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

/** Hint for play phase: follow suit, lead, or trump cut. */
function getPlayHint(cards: Card[], state: OmiResponse): HintResult {
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

  // Void in led suit: trump cut if possible (Omi allows any card when void)
  const hasTrump = cards.some((c) => c.design === trumpDesign);
  if (hasTrump) {
    return { targetAction: 'play', reason: 'hint.trumpCut', confidence: 'moderate' };
  }

  return { targetAction: 'play', reason: 'hint.leadOffSuit', confidence: 'moderate' };
}
