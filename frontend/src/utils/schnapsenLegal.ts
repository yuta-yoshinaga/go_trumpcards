/**
 * Legal-move guide helpers for the Schnapsen / Sixty-Six page.
 *
 * The backend already computes which hand indices are legal to play right now
 * (`validPlays`, from `Schnapsen.GetValidPlayIndices`). The frontend only needs
 * to decide when to *highlight* them: exclusively during the human's turn in
 * the second phase (stock exhausted, must-follow rules). In the first phase any
 * card is legal, so no guide ring is shown to avoid a visual regression.
 */

/**
 * Compute the set of hand indices to highlight as legal-to-play.
 *
 * Returns an empty set outside the human's turn or during phase 1, where every
 * card is legal and highlighting would be noise. During phase 2 it returns the
 * backend-provided `validPlays` as a set for O(1) membership checks in render.
 *
 * This is purely additive: illegal cards are never disabled here — the backend
 * still validates every play — so the ring only guides, it never blocks clicks.
 *
 * @param isEndgame - True once the stock is exhausted (phase 2, must-follow).
 * @param isHumanTurn - True when it is the human player's turn to play.
 * @param validPlays - Backend-computed legal hand indices (may be undefined).
 * @returns Set of hand indices to ring-highlight (empty in phase 1 / off-turn).
 */
export function computeSchnapsenLegalRing(
  isEndgame: boolean,
  isHumanTurn: boolean,
  validPlays: readonly number[] | undefined,
): Set<number> {
  if (!isEndgame || !isHumanTurn) return new Set<number>();
  return new Set(validPlays ?? []);
}
