import type { Card, GongZhuResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { GongZhuPhase } from '../../types/phases';

/** Suit design for hearts. */
const HEART = 'HEART';
/** Suit design for spades. */
const SPADE = 'SPADE';
/** Suit design for diamonds. */
const DIAMOND = 'DIAMOND';
/** Queen value (♠Q = pig). */
const QUEEN = 12;
/** Jack value (♦J = sheep). */
const JACK = 11;

/** Returns a frontend HintResult for Gong Zhu, or null if no suggestion. */
export function getGongZhuHint(state: GongZhuResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;

  if (state.phase === GongZhuPhase.EXPOSE) {
    return getExposeHint(human.cards);
  }

  if (state.phase === GongZhuPhase.PLAY && state.currentPlayerIdx === state.players.findIndex((p) => p.isHuman)) {
    return getPlayHint(human.cards, state);
  }

  return null;
}

/** Hint for the exposure phase: exposing the ♦J doubles its value. */
function getExposeHint(cards: Card[]): HintResult {
  const hasSheep = cards.some((c) => c.design === DIAMOND && c.value === JACK);
  if (hasSheep) {
    return { targetAction: 'expose', reason: 'hint.exposeSheep', confidence: 'moderate' };
  }
  return { targetAction: 'expose', reason: 'hint.exposeNone', confidence: 'moderate' };
}

/** Hint for the play phase: follow suit, chase the sheep, dump the pig. */
function getPlayHint(cards: Card[], state: GongZhuResponse): HintResult {
  const trick = state.currentTrick;

  if (trick.length === 0) {
    return { targetAction: 'play', reason: 'hint.leadLowest', confidence: 'moderate' };
  }

  const ledSuit = trick[0].card.design;
  const suitCards = cards.filter((c) => c.design === ledSuit);
  const trickHasSheep = trick.some((tc) => tc.card.design === DIAMOND && tc.card.value === JACK);

  if (suitCards.length > 0) {
    if (trickHasSheep) {
      return { targetAction: 'play', reason: 'hint.chaseSheep', confidence: 'strong' };
    }
    return { targetAction: 'play', reason: 'hint.followSuit', confidence: 'strong' };
  }

  // Void in led suit: dump the pig, then high hearts.
  const hasPig = cards.some((c) => c.design === SPADE && c.value === QUEEN);
  if (hasPig) {
    return { targetAction: 'play', reason: 'hint.dumpPig', confidence: 'strong' };
  }
  const hasHearts = cards.some((c) => c.design === HEART);
  if (hasHearts) {
    return { targetAction: 'play', reason: 'hint.dumpHearts', confidence: 'moderate' };
  }
  return { targetAction: 'play', reason: 'hint.playHighest', confidence: 'moderate' };
}
