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

/** Minimal card shape for movable-run detection: suit (`design`) + rank (`value`). */
interface SuitRankCard {
  design: string;
  value: number;
}

/** Tableau card with a nullable face and a face-up flag (Spider column shape). */
interface RunTableauCard {
  card: SuitRankCard | null;
  faceUp: boolean;
}

/**
 * Computes the contiguous same-suit descending run that moves as a single unit when the
 * card at `index` is grabbed in a Spider-style tableau column. Grabbing a card in Spider
 * always takes every card below it, so the grab is only legal when the entire tail
 * `[index .. bottom]` is face-up, all the same suit, and strictly descending by one rank
 * (mirroring the backend's `isValidSpiderSequence`). Returns the indices of that tail when
 * it forms a valid movable run, or an empty array when the card is face-down, missing, or
 * the tail is broken (i.e. not movable, so no highlight should appear).
 */
export function spiderMovableRun(column: readonly RunTableauCard[], index: number): number[] {
  if (index < 0 || index >= column.length) return [];
  const start = column[index];
  if (!start.faceUp || !start.card) return [];
  for (let i = index + 1; i < column.length; i++) {
    const prev = column[i - 1].card;
    const curr = column[i];
    if (!curr.faceUp || !curr.card || !prev) return [];
    if (curr.card.design !== prev.design) return [];
    if (curr.card.value !== prev.value - 1) return [];
  }
  const run: number[] = [];
  for (let i = index; i < column.length; i++) run.push(i);
  return run;
}
