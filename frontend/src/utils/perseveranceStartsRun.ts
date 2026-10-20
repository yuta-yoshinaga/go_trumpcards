import type { PerseveranceTableauCard } from '../types/card';

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
