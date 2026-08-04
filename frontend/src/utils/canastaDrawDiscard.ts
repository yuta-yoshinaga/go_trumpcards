import type { Card } from '../types/card';

/** i18n key suffix explaining why a pair cannot take the discard pile, or null when it can. */
export type CanastaDrawDiscardProblem =
  | 'selectTwo'
  | 'selectOneMore'
  | 'tooMany'
  | 'pileEmpty'
  | 'wildInPair'
  | 'rankMismatch';

/**
 * Whether a card counts as wild in Canasta — a joker or any two.
 * Mirrors `CanastaIsWild` in `internal/domain/Canasta.go`.
 * @param card - The card to test.
 * @returns Whether it is wild.
 */
export function canastaIsWild(card: Card): boolean {
  return card.design === 'JOKER' || card.value === 2;
}

/**
 * Why the selected cards cannot take the discard pile, or null when they can.
 *
 * `PlayerDrawFromDiscard` requires exactly two cards, both natural, both of the
 * discard top's rank — always, not only while the pile is frozen. Checking only
 * the count made the button look ready for a selection the server would reject.
 *
 * The initial-meld minimum is deliberately not checked here: it depends on
 * server-side state the page does not model, so the server stays the authority
 * on it.
 * @param selected - The cards the player has selected, in selection order.
 * @param discardTop - The card on top of the discard pile.
 * @returns The problem key, or null when the pair is acceptable.
 */
export function canastaDrawDiscardProblem(
  selected: readonly Card[],
  discardTop: Card | null | undefined,
): CanastaDrawDiscardProblem | null {
  if (selected.length > 2) return 'tooMany';
  if (selected.length === 0) return 'selectTwo';
  if (selected.length === 1) return 'selectOneMore';
  if (!discardTop) return 'pileEmpty';
  const [a, b] = selected as [Card, Card];
  if (canastaIsWild(a) || canastaIsWild(b)) return 'wildInPair';
  if (a.value !== discardTop.value || b.value !== discardTop.value) return 'rankMismatch';
  return null;
}
