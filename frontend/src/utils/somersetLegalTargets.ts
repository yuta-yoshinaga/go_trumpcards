import type { Card } from '../types/card';
import type { SomersetTableauCard } from '../types/games/somerset';

/** Where the selected card may legally go. */
export interface SomersetLegalTargets {
  /** Tableau column indices that accept the card. */
  tableau: Set<number>;
  /** Foundation indices that accept the card. */
  foundation: Set<number>;
}

/**
 * Legal destinations for the selected card.
 *
 * Sync: `Somerset.canPlaceOnTableau` / `canPlaceOnFoundation`.
 *
 * **タブローはスートを見ない。**ランクが1つ下がるかどうかだけ。ファンデーションは
 * 逆に同スートで1つ上がるときだけ。この2つを取り違えると、置けない列を光らせる。
 *
 * **空き列にはどのカードでも置ける。**姉妹の Baker's Dozen は空き列を埋められない
 * ので、そちらの規則を流用すると実際には打てる手を落とす (#4799)。
 */
export function somersetLegalTargets(
  tableau: SomersetTableauCard[][],
  foundation: Card[][],
  card: Card | null | undefined,
): SomersetLegalTargets {
  const result: SomersetLegalTargets = { tableau: new Set(), foundation: new Set() };
  if (!card) return result;

  tableau.forEach((col, idx) => {
    if (col.length === 0) {
      result.tableau.add(idx);
      return;
    }
    const top = col[col.length - 1]?.card;
    if (top && card.value === top.value - 1) result.tableau.add(idx);
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
