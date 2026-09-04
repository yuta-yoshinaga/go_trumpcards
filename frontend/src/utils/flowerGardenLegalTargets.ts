import type { Card } from '../types/card';
import type { FlowerGardenTableauCard } from '../types/games/flowergarden';

/** Where the selected card may legally go. */
export interface FlowerGardenLegalTargets {
  /** Tableau column indices that accept the card. */
  tableau: Set<number>;
  /** Foundation indices that accept the card. */
  foundation: Set<number>;
}

/**
 * Legal destinations for the selected card.
 *
 * Sync: `FlowerGarden.canPlaceOnTableau` / `canPlaceOnFoundationPile`
 * (`internal/domain/FlowerGarden.go`).
 *
 * - タブローの規則: 値差 -1 のみ（スートも色も見ない。赤黒交互ではない）。空列には任意のカードを置ける。
 * - ファンデーションの規則: 空なら A、そうでなければ同スートで +1。ただし WebController は index を受け取らず
 *   ドメイン側の `findFoundation` で受け入れ可能な最初の山に着地するため、合法なファンデーションとして
 *   光らせるのは実際に着地する最初の1つだけにする。
 */
export function flowerGardenLegalTargets(
  tableau: FlowerGardenTableauCard[][],
  foundation: Card[][],
  card: Card | null | undefined,
): FlowerGardenLegalTargets {
  const result: FlowerGardenLegalTargets = { tableau: new Set(), foundation: new Set() };
  if (!card) return result;

  tableau.forEach((col, idx) => {
    if (col.length === 0) {
      result.tableau.add(idx);
      return;
    }
    const top = col[col.length - 1]?.card;
    if (top && card.value === top.value - 1) {
      result.tableau.add(idx);
    }
  });

  for (let idx = 0; idx < foundation.length; idx++) {
    const pile = foundation[idx];
    if (pile.length === 0) {
      if (card.value === 1) {
        result.foundation.add(idx);
        break;
      }
    } else {
      const top = pile[pile.length - 1];
      if (top && card.design === top.design && card.value === top.value + 1) {
        result.foundation.add(idx);
        break;
      }
    }
  }

  return result;
}
