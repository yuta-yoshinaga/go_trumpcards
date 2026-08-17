import type { Card } from '../types/card';

/** i18n key suffix explaining why a pair cannot take the discard pile, or null when it can. */
export type CanastaDrawDiscardProblem =
  | 'selectTwo'
  | 'selectOneMore'
  | 'tooMany'
  | 'pileEmpty'
  | 'blackThreeTop'
  | 'wildTop'
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
 * Whether a card is a black three, which blocks taking the pile outright.
 * Mirrors `CanastaIsBlack3` in `internal/domain/Canasta.go`.
 * @param card - The card to test.
 * @returns Whether it is a black three.
 */
export function canastaIsBlack3(card: Card): boolean {
  return card.value === 3 && (card.design === 'SPADE' || card.design === 'CLOVER');
}

/**
 * Why the selected cards cannot take the discard pile, or null when they can.
 *
 * `PlayerDrawFromDiscard` refuses a black three or a wild on top outright, then requires
 * exactly two cards, both natural, both of the discard top's rank — always, not
 * only while the pile is frozen. Checking only
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
  // Checked before the pair, as the domain does: a black three on top blocks the
  // take outright, and two black threes in hand would otherwise pass the rank
  // check (3 === 3) since a black three is not wild.
  if (canastaIsBlack3(discardTop)) return 'blackThreeTop';
  // ワイルドがトップでも取れない。PlayerDrawFromDiscard は黒3の直後にこれも弾くが、
  // ここには無かったので「取れます」と見せてからサーバに拒否されていた (#5502)。
  if (canastaIsWild(discardTop)) return 'wildTop';
  const [a, b] = selected as [Card, Card];
  if (canastaIsWild(a) || canastaIsWild(b)) return 'wildInPair';
  if (a.value !== discardTop.value || b.value !== discardTop.value) return 'rankMismatch';
  return null;
}
