import type { BakersDozenTableauCard, Card } from '../types/card';

/** Where the selected card may legally go. */
export interface BakersDozenLegalTargets {
  /** Tableau column indices that accept the card. */
  tableau: Set<number>;
  /** Foundation indices that accept the card. */
  foundation: Set<number>;
}

/**
 * Legal destinations for the selected card.
 *
 * Sync: `BakersDozen.canPlaceOnTableau` / `canPlaceOnFoundation`.
 *
 * **タブローはスートを見ない。**ランクが1つ下がるかどうかだけ。ファンデーションは
 * 逆に同スートで1つ上がるときだけ。この2つを取り違えると、置けない列を光らせる。
 *
 * **空き列には置けない。**Baker's Dozen は空き列を埋められないので、空の列を
 * 移動先に含めてはいけない (他のソリティアと違うところ)。
 */
export function bakersDozenLegalTargets(
  tableau: BakersDozenTableauCard[][],
  foundation: Card[][],
  card: Card | null | undefined,
): BakersDozenLegalTargets {
  const result: BakersDozenLegalTargets = { tableau: new Set(), foundation: new Set() };
  if (!card) return result;

  tableau.forEach((col, idx) => {
    // **空き列には置けない** (Baker's Dozen は空き列を埋められない)。空の列には
    // トップ札が無いので `top` の存在確認がそのままその条件になる。専用の
    // `col.length === 0` を足しても常に同じ結果になる枝が増えるだけだった。
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
