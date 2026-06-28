import type { KlondikeTableauCard } from '../types/card';

/**
 * Columns onto which the selected Wasp card may legally move: a non-empty
 * column whose face-up top card is the same suit and exactly one rank higher.
 * The source column is excluded; empty columns accept any card and are handled
 * by the page's own placement UI.
 *
 * @param tableau - All tableau columns.
 * @param source - The selected card's column and index (either may be undefined).
 * @returns The set of legal target column indices.
 */
export function waspLegalTargets(
  tableau: KlondikeTableauCard[][],
  source: { col?: number; cardIndex?: number },
): Set<number> {
  const result = new Set<number>();
  if (source.col === undefined || source.cardIndex === undefined) return result;
  const srcCol = source.col;
  const srcCard = tableau[srcCol]?.[source.cardIndex]?.card;
  if (!srcCard) return result;
  tableau.forEach((col, idx) => {
    if (idx === srcCol || col.length === 0) return;
    const top = col[col.length - 1];
    if (top?.faceUp && top.card && top.card.design === srcCard.design && top.card.value === srcCard.value + 1) {
      result.add(idx);
    }
  });
  return result;
}
