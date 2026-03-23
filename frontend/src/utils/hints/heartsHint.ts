import type { Card, HeartsResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { HeartsPhase } from '../../types/phases';

/** Suit design for hearts. */
const HEART = 'HEART';
/** Suit design for spades. */
const SPADE = 'SPADE';
/** Queen value. */
const QUEEN = 12;

/** Returns a frontend HintResult for Hearts, or null if no suggestion. */
export function getHeartsHint(state: HeartsResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;

  if (state.phase === HeartsPhase.PASS) {
    return getPassHint(human.cards);
  }

  if (state.phase === HeartsPhase.PLAY && state.currentPlayerIdx === state.players.findIndex((p) => p.isHuman)) {
    return getPlayHint(human.cards, state);
  }

  return null;
}

/** Hint for pass phase: pass high hearts and Queen of Spades. */
function getPassHint(cards: Card[]): HintResult {
  const hasQueenOfSpades = cards.some((c) => c.design === SPADE && c.value === QUEEN);
  const hasHighHearts = cards.some((c) => c.design === HEART && c.value >= 10);

  if (hasQueenOfSpades) {
    return { targetAction: 'pass', reason: 'hint.passQueenOfSpades', confidence: 'strong' };
  }
  if (hasHighHearts) {
    return { targetAction: 'pass', reason: 'hint.passHighHearts', confidence: 'strong' };
  }
  return { targetAction: 'pass', reason: 'hint.passHighCards', confidence: 'moderate' };
}

/** Hint for play phase: follow suit, avoid penalties. */
function getPlayHint(cards: Card[], state: HeartsResponse): HintResult {
  const trick = state.currentTrick;

  // Leading the trick
  if (trick.length === 0) {
    if (!state.heartsBroken) {
      return { targetAction: 'play', reason: 'hint.leadNonHeart', confidence: 'strong' };
    }
    return { targetAction: 'play', reason: 'hint.leadLowest', confidence: 'moderate' };
  }

  // Following suit
  const ledSuit = trick[0].card.design;
  const suitCards = cards.filter((c) => c.design === ledSuit);

  if (suitCards.length > 0) {
    return { targetAction: 'play', reason: 'hint.followSuit', confidence: 'strong' };
  }

  // Void in led suit: dump penalty cards
  const hasQueenOfSpades = cards.some((c) => c.design === SPADE && c.value === QUEEN);
  if (hasQueenOfSpades) {
    return { targetAction: 'play', reason: 'hint.dumpQueenOfSpades', confidence: 'strong' };
  }

  const hasHearts = cards.some((c) => c.design === HEART);
  if (hasHearts) {
    return { targetAction: 'play', reason: 'hint.dumpHearts', confidence: 'moderate' };
  }

  return { targetAction: 'play', reason: 'hint.playHighest', confidence: 'moderate' };
}
