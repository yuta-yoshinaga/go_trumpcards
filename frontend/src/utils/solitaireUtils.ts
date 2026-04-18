/**
 * Utilities shared across solitaire-style games (Klondike, Spider, Yukon, Forty Thieves).
 */

/** Minimal tableau card shape — any solitaire card with a face-up flag. */
interface TableauCardLike {
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
