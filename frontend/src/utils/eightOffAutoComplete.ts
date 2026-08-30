import type { Card } from '../types/card';
import { eightOffFoundationTarget } from './eightOffFoundationTarget';

/**
 * Whether running auto-complete would actually move a card.
 *
 * Sync: `EightOff.AutoComplete` (`internal/domain/EightOff.go:411-456`), which sweeps the
 * free cells and then each tableau's end card onto the foundations, repeating until nothing
 * moves. This answers the same question for the first pass — if nothing is placeable now,
 * the whole loop is a no-op — and reuses {@link eightOffFoundationTarget} so the placement
 * rule is stated once.
 *
 * **The sister-game shortcut does not transfer.** BeleagueredCastle and StreetsAndAlleys
 * gate on `foundation.some(p => p.length > 1)` because their `Reset` seeds all four aces
 * onto the foundations before dealing (`internal/domain/BeleagueredCastle.go:104-112`), so
 * there "length > 1" reads as "built past the ace". Eight Off starts with empty foundations
 * (`EightOff.go:97`), so both that test and `length > 0` report "not ready" in states where
 * the sweep would happily move an ace off a free cell or a column end.
 */
export function eightOffAutoCompleteReady(
  freeCells: readonly (Card | null)[],
  tableau: readonly (readonly (Card | null)[])[],
  foundation: Card[][],
): boolean {
  for (const card of freeCells) {
    if (card && eightOffFoundationTarget(card, foundation)) return true;
  }
  for (const col of tableau) {
    const end = col[col.length - 1];
    if (end && eightOffFoundationTarget(end, foundation)) return true;
  }
  return false;
}
