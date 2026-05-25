import type { Card, CassinoBuild } from '../types/card';

/** Cassino face cards (J/Q/K) take by exact-rank match only — they never contribute to numeric sums. */
export function isCassinoFaceCard(card: Card): boolean {
  return card.value >= 11 && card.value <= 13;
}

/** A suggested one-click action computed from the player's current selection. */
export type CassinoSuggestion =
  | { type: 'take'; reason: 'sum' | 'faceMatch' | 'buildMatch'; value: number }
  | { type: 'build'; declaredValue: number };

/**
 * Given the player's current selection (hand card + selected table indices + selected build indices),
 * return the action that selection implies, if any.
 *
 * - Take by numeric sum: sum of selected table card values equals the played card's value (face cards excluded).
 * - Take by face match: every selected table card has the same rank as the played face card.
 * - Take by build match: the player has selected one or more builds whose declared value matches the played card.
 * - Build: hand card + table sum equals a value N where the player holds another N in hand (2-10, N > hand value).
 *
 * Returns the strongest inferred action (take is preferred over build when both apply).
 */
export function suggestCassinoAction(args: {
  handCard: Card;
  hand: Card[];
  handIndex: number;
  selectedTableCards: Card[];
  selectedBuilds: CassinoBuild[];
}): CassinoSuggestion | null {
  const { handCard, hand, handIndex, selectedTableCards, selectedBuilds } = args;

  if (isCassinoFaceCard(handCard)) {
    // Face take: every selected table card matches rank, no builds selected.
    if (selectedBuilds.length === 0 && selectedTableCards.length > 0) {
      const allMatch = selectedTableCards.every((c) => c.value === handCard.value);
      if (allMatch) return { type: 'take', reason: 'faceMatch', value: handCard.value };
    }
    return null;
  }

  // Numeric hand cards cannot capture face cards — guard before every numeric path
  // (including build-match) so a mixed table selection never silently passes the sum check.
  if (selectedTableCards.some(isCassinoFaceCard)) return null;

  // Build-match take: selected builds all share the played card's value.
  if (selectedBuilds.length > 0) {
    const allMatch = selectedBuilds.every((b) => b.value === handCard.value);
    const tableSum = sumNonFace(selectedTableCards);
    const tableOk = selectedTableCards.length === 0 || tableSum === handCard.value;
    if (allMatch && tableOk) {
      return { type: 'take', reason: 'buildMatch', value: handCard.value };
    }
    return null;
  }

  if (selectedTableCards.length === 0) return null;

  const tableSum = sumNonFace(selectedTableCards);
  if (tableSum === handCard.value) {
    return { type: 'take', reason: 'sum', value: handCard.value };
  }

  const combined = handCard.value + tableSum;
  if (combined >= 2 && combined <= 10 && combined > handCard.value) {
    const hasCapture = hand.some((c, i) => i !== handIndex && !isCassinoFaceCard(c) && c.value === combined);
    if (hasCapture) {
      return { type: 'build', declaredValue: combined };
    }
  }

  return null;
}

function sumNonFace(cards: Card[]): number {
  return cards.reduce((acc, c) => (isCassinoFaceCard(c) ? acc : acc + c.value), 0);
}
