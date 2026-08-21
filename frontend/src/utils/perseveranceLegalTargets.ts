import type { Card, PerseveranceTableauCard } from '../types/card';

/** Where the selected card may legally go. */
export interface PerseveranceLegalTargets {
  /** Tableau column indices that accept the card. */
  tableau: Set<number>;
  /** Foundation indices that accept the card. */
  foundation: Set<number>;
}

/**
 * Legal destinations for the selected card.
 *
 * Sync: `Perseverance.canPlaceOnTableau` / `canPlaceOnFoundation`.
 *
 * **タブローも同スート。**ランクが1つ下がり、かつスートが一致するときだけ。
 * ファンデーションは同スートで1つ上がるときだけ。
 *
 * **クローン元の Baker\'s Dozen はここでスートを見ない。**その版をそのまま残すと
 * ♠8 を ♥9 に置けるように光り、サーバに弾かれる手を勧めることになる。
 *
 * **空き列には置けない。**Perseverance は空き列を埋められないので、空の列を
 * 移動先に含めてはいけない (他のソリティアと違うところ)。
 */
export function perseveranceLegalTargets(
  tableau: PerseveranceTableauCard[][],
  foundation: Card[][],
  card: Card | null | undefined,
): PerseveranceLegalTargets {
  const result: PerseveranceLegalTargets = { tableau: new Set(), foundation: new Set() };
  if (!card) return result;

  tableau.forEach((col, idx) => {
    // **空き列には置けない** (Perseverance も Baker's Dozen も空き列は埋められない)。空の列には
    // トップ札が無いので `top` の存在確認がそのままその条件になる。専用の
    // `col.length === 0` を足しても常に同じ結果になる枝が増えるだけだった。
    const top = col[col.length - 1]?.card;
    if (top && card.design === top.design && card.value === top.value - 1) result.tableau.add(idx);
  });

  foundation.forEach((pile, idx) => {
    // **Perseverance では組札は空にならない。**A 4 枚は配る前から乗っている。
    // それでも分岐を残すのは、サーバが壊れた state を返したときに落ちないため。
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
