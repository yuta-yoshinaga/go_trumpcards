import type { Card } from '../types/card';
import type { DoubleKlondikeTableauCard } from '../types/games/doubleklondike';
import { isRedSuitDesign } from './cardAlt';

/** King, the only rank an empty Double Klondike column accepts. */
const KING = 13;

/**
 * Whether `card` may be stacked on tableau column `toCol`: an empty column takes
 * a King only, otherwise the move must alternate colour and descend by one.
 * Mirrors `canPlaceOnTableau` in `internal/domain/DoubleKlondike.go`.
 * @param card - The card being moved.
 * @param tableau - All tableau columns.
 * @param toCol - Index of the destination column.
 * @returns Whether the move is legal.
 */
export function doubleKlondikeCanPlaceOnTableau(
  card: Card,
  tableau: readonly (readonly DoubleKlondikeTableauCard[])[],
  toCol: number,
): boolean {
  const pile = tableau[toCol];
  if (!pile) return false;
  if (pile.length === 0) return card.value === KING;
  const top = pile[pile.length - 1];
  if (!top?.faceUp || !top.card) return false;
  return isRedSuitDesign(card.design) !== isRedSuitDesign(top.card.design) && card.value === top.card.value - 1;
}

/**
 * Whether `card` may go onto foundation pile `fIdx`: an empty pile takes an Ace,
 * otherwise it builds up in the same suit. Two decks mean eight piles, so a card
 * can be legal on more than one — each index is asked separately, exactly as
 * `canPlaceOnFoundation` in `internal/domain/DoubleKlondike.go` does.
 * @param card - The card being moved.
 * @param foundation - All foundation piles.
 * @param fIdx - Index of the destination pile.
 * @returns Whether the move is legal.
 */
export function doubleKlondikeCanPlaceOnFoundation(
  card: Card,
  foundation: readonly (readonly Card[])[],
  fIdx: number,
): boolean {
  const pile = foundation[fIdx];
  if (!pile) return false;
  if (pile.length === 0) return card.value === 1;
  const top = pile[pile.length - 1];
  return top !== undefined && card.design === top.design && card.value === top.value + 1;
}
