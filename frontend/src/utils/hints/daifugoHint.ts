import type { Card, DaifugoResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Minimum cards of the same value to trigger a revolution. */
const REVOLUTION_COUNT = 4;

/** Returns a frontend HintResult for Daifugo, or null if no suggestion. */
export function getDaifugoHint(state: DaifugoResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.isFinished) return null;
  if (state.gameEndFlag) return null;
  if (state.pendingAction !== 'none') return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (state.currentTurn !== humanIdx) return null;

  // Free turn (empty table or last player was self)
  if (state.tableCards.length === 0) {
    return getFreeTurnHint(human.cards);
  }

  return getFollowHint(human.cards, state);
}

/** Hint when it's a free turn (no cards on table). */
function getFreeTurnHint(cards: Card[]): HintResult {
  if (hasRevolution(cards)) {
    return { targetAction: 'play', reason: 'hint.revolutionChance', confidence: 'strong' };
  }

  return { targetAction: 'play', reason: 'hint.playLowest', confidence: 'moderate' };
}

/** Hint when following an existing play. */
function getFollowHint(cards: Card[], state: DaifugoResponse): HintResult {
  const tableValue = state.tableCards[0]?.value ?? 0;
  const revolution = state.revolutionActive;

  const hasStronger = cards.some((c) => isStrongerThan(c.value, tableValue, revolution));

  if (hasStronger) {
    return { targetAction: 'play', reason: 'hint.playStronger', confidence: 'strong' };
  }

  return { targetAction: 'pass', reason: 'hint.passNoPlay', confidence: 'moderate' };
}

/** Check if a card value beats the table value, considering revolution. */
function isStrongerThan(cardValue: number, tableValue: number, revolution: boolean): boolean {
  if (revolution) {
    return cardValue < tableValue;
  }
  return cardValue > tableValue;
}

/** Check if hand contains 4+ cards of the same value (revolution potential). */
function hasRevolution(cards: Card[]): boolean {
  const valueCounts = new Map<number, number>();
  for (const c of cards) {
    valueCounts.set(c.value, (valueCounts.get(c.value) ?? 0) + 1);
  }
  return [...valueCounts.values()].some((count) => count >= REVOLUTION_COUNT);
}
