/**
 * Utilities shared across solitaire-style games (Klondike, Spider, Yukon, Forty Thieves).
 */

/** Minimal tableau card shape — any solitaire card with a face-up flag. */
export interface TableauCardLike {
  faceUp: boolean;
}

/**
 * Returns true when every card in the tableau is face-up. This is the standard
 * prerequisite for a safe "auto-complete" run on games that use a face-down stock
 * (Klondike, Spider, Yukon, Forty Thieves) — once no hidden cards remain, the
 * remaining cards can always be sent to the foundation.
 */
export function isTableauAllFaceUp(tableau: readonly (readonly TableauCardLike[])[]): boolean {
  for (const col of tableau) {
    for (const cell of col) {
      if (!cell.faceUp) return false;
    }
  }
  return true;
}

/** A selected tableau card: the column and the index within that column. */
export interface TableauSelection {
  col: number;
  cardIndex: number;
}

/**
 * Determines whether the tableau card at (`col`, `cardIdx`) belongs to the
 * "block" that would be lifted together when the card at `selected` is picked
 * up. In Yukon a face-up card moves with every card sitting on top of it
 * regardless of suit/rank order, so the block spans from the selected index
 * down to the column bottom. Returns `false` when nothing is selected, or the
 * candidate is in a different column, or the candidate sits above the selection.
 *
 * This powers the tap-to-preview block highlight on touch devices, where hover
 * is unavailable (#3152).
 */
export function isInMoveBlock(selected: TableauSelection | null, col: number, cardIdx: number): boolean {
  return selected !== null && selected.col === col && cardIdx >= selected.cardIndex;
}
