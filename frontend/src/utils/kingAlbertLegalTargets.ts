import type { Card } from '../types/card';
import type { KingAlbertTableauCard } from '../types/games/kingalbert';

/** Where the selected card may legally go. */
export interface KingAlbertLegalTargets {
  /** Tableau column indices that accept the card. */
  tableau: Set<number>;
  /** Foundation indices that accept the card. */
  foundation: Set<number>;
}

const BLACK = new Set(['SPADE', 'CLOVER']);

const isBlack = (card: Card): boolean => BLACK.has(card.design);

/**
 * Legal destinations for the selected card.
 *
 * Sync: `KingAlbert.canPlaceOnTableau` / `canPlaceOnFoundation`
 * (`internal/domain/KingAlbert.go`).
 *
 * **タブローは交互の色で1つ下がるときだけ。**姉妹の包囲された城は色を見ないので、
 * そちらの規則を流用すると置けない列まで光る。空き列にはどのカードでも置ける。
 *
 * 選択中の札が乗っている列を除く必要は無い。列の一番上と自分自身を比べることになり、
 * `value === value - 1` は必ず偽になるので、その列は元から候補にならない。
 */
export function kingAlbertLegalTargets(
  tableau: KingAlbertTableauCard[][],
  foundation: Card[][],
  card: Card | null | undefined,
): KingAlbertLegalTargets {
  const result: KingAlbertLegalTargets = { tableau: new Set(), foundation: new Set() };
  if (!card) return result;

  tableau.forEach((col, idx) => {
    if (col.length === 0) {
      result.tableau.add(idx);
      return;
    }
    const top = col[col.length - 1]?.card;
    if (top && card.value === top.value - 1 && isBlack(card) !== isBlack(top)) {
      result.tableau.add(idx);
    }
  });

  foundation.forEach((pile, idx) => {
    if (pile.length === 0) {
      if (card.value === 1) result.foundation.add(idx);
      return;
    }
    const top = pile[pile.length - 1];
    if (top && card.design === top.design && card.value === top.value + 1) {
      result.foundation.add(idx);
    }
  });

  return result;
}
