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

/**
 * Whether `cardIndex` in `col` starts a movable group.
 *
 * Sync: `Perseverance.runStarts` / `isRun`.
 *
 * **並びを一括で動かせるのが Perseverance の看板ルール**なので、掴める札は上札
 * だけではない。そこから上が同スート降順に連続していれば、埋もれた札からでも
 * 掴める。並びは必ず上札から続くので、cardIndex から上を見れば足りる。
 *
 * クローン元の Baker's Dozen は 1 枚ずつしか動かせず、`isTop` だけで判定していた。
 * その判定を残すと、ドメインが受け取れる手を UI が永久に出せない。
 */
export function perseveranceStartsRun(col: PerseveranceTableauCard[], cardIndex: number): boolean {
  if (cardIndex < 0 || cardIndex >= col.length) return false;
  for (let i = cardIndex; i + 1 < col.length; i++) {
    const upper = col[i]?.card;
    const lower = col[i + 1]?.card;
    if (!upper || !lower) return false;
    if (upper.design !== lower.design || lower.value !== upper.value - 1) return false;
  }
  return true;
}
