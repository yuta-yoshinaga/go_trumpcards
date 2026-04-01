import type { Card, CribbageResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { CribbagePhase } from '../../types/phases';

/** Returns a frontend HintResult for Cribbage, or null if no suggestion. */
export function getCribbageHint(state: CribbageResponse): HintResult | null {
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (humanIdx === -1) return null;
  const human = state.players[humanIdx];
  if (human.cards.length === 0) return null;
  if (state.gameEndFlag) return null;

  if (state.phase === CribbagePhase.DISCARD) {
    if (state.currentPlayerIdx !== humanIdx) return null;
    const isDealer = state.dealerIdx === humanIdx;
    return getDiscardHint(human.cards, isDealer);
  }

  if (state.phase === CribbagePhase.PEGGING) {
    if (state.currentPlayerIdx !== humanIdx) return null;
    return getPeggingHint(human.cards, state.pegCount);
  }

  return null;
}

/** Hint for the discard phase: suggest keeping high-scoring combinations. */
function getDiscardHint(hand: Card[], isDealer: boolean): HintResult {
  if (hand.length <= 4) {
    return { targetAction: 'discard', reason: 'hint.discardAny', confidence: 'moderate' };
  }

  const hasFiveOrTen = hand.some((c) => pegValue(c) === 5 || pegValue(c) === 10);

  if (isDealer) {
    if (hasFiveOrTen) {
      return { targetAction: 'discard', reason: 'hint.discardToCribDealer', confidence: 'moderate' };
    }
    return { targetAction: 'discard', reason: 'hint.keepBestHand', confidence: 'moderate' };
  }

  // Non-dealer: avoid giving good cards to opponent's crib
  if (hasFiveOrTen) {
    return { targetAction: 'discard', reason: 'hint.keepFivesAndTens', confidence: 'moderate' };
  }

  return { targetAction: 'discard', reason: 'hint.keepBestHand', confidence: 'moderate' };
}

/** Hint for the pegging phase: suggest best card to play or Go. */
function getPeggingHint(hand: Card[], pegCount: number): HintResult {
  const playable = hand.filter((c) => pegCount + pegValue(c) <= 31);

  if (playable.length === 0) {
    return { targetAction: 'go', reason: 'hint.mustGo', confidence: 'strong' };
  }

  // Check if we can hit exactly 15
  const hitsFifteen = playable.find((c) => pegCount + pegValue(c) === 15);
  if (hitsFifteen) {
    return { targetAction: 'peg', reason: 'hint.pegFifteen', confidence: 'strong' };
  }

  // Check if we can hit exactly 31
  const hitsThirtyOne = playable.find((c) => pegCount + pegValue(c) === 31);
  if (hitsThirtyOne) {
    return { targetAction: 'peg', reason: 'hint.pegThirtyOne', confidence: 'strong' };
  }

  // Prefer playing a card that avoids leaving count at 5 or 21 (opponent can score 15/31)
  const safe = playable.filter((c) => {
    const newCount = pegCount + pegValue(c);
    return newCount !== 5 && newCount !== 21;
  });

  if (safe.length > 0) {
    return { targetAction: 'peg', reason: 'hint.pegSafe', confidence: 'moderate' };
  }

  return { targetAction: 'peg', reason: 'hint.pegPlay', confidence: 'moderate' };
}

/** Peg value of a card (face cards = 10, ace = 1). */
function pegValue(card: Card): number {
  if (card.value >= 10) return 10;
  return card.value;
}
