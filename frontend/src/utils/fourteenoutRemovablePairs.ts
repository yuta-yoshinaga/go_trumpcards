import type { FourteenOutBoardCell } from '../types/card';

/** The sum every removable pair must reach. Sync: `domain.FourteenOutTargetSum`. */
export const FOURTEEN_OUT_TARGET_SUM = 14;

/** The exposed card of a column, or `null` when the column is cleared. */
export function fourteenOutTail(col: FourteenOutBoardCell[] | undefined): FourteenOutBoardCell['card'] {
  if (!col || col.length === 0) return null;
  return col[col.length - 1]?.card ?? null;
}

/**
 * Whether the tails of two columns can be removed together.
 *
 * Sync: `FourteenOut.Remove` in `internal/domain/FourteenOut.go`.
 *
 * **合計 14 だけが規則。**スートは見ず、列が隣り合っている必要もない。
 * クローン元の Monte Carlo は「同ランク」かつ「8方向で隣接」を要求するので、
 * その判定を残すと合法な手を出せず、非合法な手を出せてしまう。
 *
 * K と A が組めるのも特例ではなく 13+1=14 の結果。同ランクで組めるのは 7+7 だけ。
 */
export function fourteenOutCanRemove(columns: FourteenOutBoardCell[][], c1: number, c2: number): boolean {
  if (c1 === c2) return false;
  const a = fourteenOutTail(columns[c1]);
  const b = fourteenOutTail(columns[c2]);
  if (!a || !b) return false;
  return a.value + b.value === FOURTEEN_OUT_TARGET_SUM;
}

/**
 * Counts the pairs of column tails that currently sum to 14.
 *
 * Each unordered pair is counted once: the inner loop starts at `c1 + 1`,
 * matching the domain's `forEachRemovablePair` scan.
 *
 * @param columns - The 12 Fourteen Out columns (ragged; a cleared column is empty).
 * @returns The number of distinct removable pairs.
 */
export function countRemovablePairs(columns: FourteenOutBoardCell[][]): number {
  let count = 0;
  for (let c1 = 0; c1 < columns.length; c1++) {
    if (!fourteenOutTail(columns[c1])) continue;
    for (let c2 = c1 + 1; c2 < columns.length; c2++) {
      if (fourteenOutCanRemove(columns, c1, c2)) count++;
    }
  }
  return count;
}

/** The set of columns whose tail can pair with the tail of `from`. */
export function fourteenOutPartners(columns: FourteenOutBoardCell[][], from: number): Set<number> {
  const out = new Set<number>();
  for (let c = 0; c < columns.length; c++) {
    if (fourteenOutCanRemove(columns, from, c)) out.add(c);
  }
  return out;
}
