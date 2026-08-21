import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, FourteenOutBoardCell } from '../types/card';
import {
  countRemovablePairs,
  FOURTEEN_OUT_TARGET_SUM,
  fourteenOutCanRemove,
  fourteenOutPartners,
  fourteenOutTail,
} from './fourteenoutRemovablePairs';

const card = (design: CardDesign, value: number): Card => ({ design, value });

/** 12 列の盤を作る。指定した列だけ中身を持つ。 */
function board(...cols: number[][]): FourteenOutBoardCell[][] {
  return Array.from({ length: 12 }, (_, i) =>
    (cols[i] ?? []).map((v) => ({ card: card('SPADE', v) }) as FourteenOutBoardCell),
  );
}

describe('fourteenOutTail', () => {
  it('returns the last card of a column', () => {
    expect(fourteenOutTail(board([3, 9])[0])?.value).toBe(9);
  });

  it('returns null for a cleared or missing column', () => {
    expect(fourteenOutTail([])).toBeNull();
    expect(fourteenOutTail(undefined)).toBeNull();
  });
});

describe('fourteenOutCanRemove', () => {
  // **合計 14 が唯一の規則。**7 通りすべてが通ること。
  it.each([
    ['K and A', 13, 1],
    ['Q and 2', 12, 2],
    ['J and 3', 11, 3],
    ['10 and 4', 10, 4],
    ['9 and 5', 9, 5],
    ['8 and 6', 8, 6],
    ['7 and 7', 7, 7],
  ])('accepts %s', (_name, a, b) => {
    expect(fourteenOutCanRemove(board([a], [b]), 0, 1)).toBe(true);
  });

  // 負のコントロール: 境界の両側。
  it.each([
    ['sums to 13', 9, 4],
    ['sums to 15', 9, 6],
    ['two kings sum to 26', 13, 13],
    ['king with a two sums to 15', 13, 2],
    ['same rank that is not 7-7', 5, 5],
  ])('refuses a pair that %s', (_name, a, b) => {
    expect(fourteenOutCanRemove(board([a], [b]), 0, 1)).toBe(false);
  });

  // **クローン元は隣接セルしか組めない。**離れた列でも通ること。
  it('pairs distant columns', () => {
    const cols = Array.from({ length: 12 }, () => [] as FourteenOutBoardCell[]);
    cols[0] = [{ card: card('SPADE', 9) }];
    cols[11] = [{ card: card('HEART', 5) }];
    expect(fourteenOutCanRemove(cols, 0, 11)).toBe(true);
  });

  // **クローン元は同ランクを要求する。**スート違いでも通ること。
  it('ignores suit', () => {
    const cols = Array.from({ length: 12 }, () => [] as FourteenOutBoardCell[]);
    cols[0] = [{ card: card('SPADE', 9) }];
    cols[1] = [{ card: card('DIAMOND', 5) }];
    expect(fourteenOutCanRemove(cols, 0, 1)).toBe(true);
  });

  // **末尾しか見ない。**埋もれた札は合計が合っても使えない。
  it('only looks at the tail', () => {
    // 列0 = [9, 2]。末尾は 2 なので 5 とは組めない (2+5=7)。
    expect(fourteenOutCanRemove(board([9, 2], [5]), 0, 1)).toBe(false);
    expect(fourteenOutCanRemove(board([9], [5]), 0, 1)).toBe(true);
  });

  it('refuses the same column and a cleared column', () => {
    expect(fourteenOutCanRemove(board([7], [7]), 0, 0)).toBe(false);
    expect(fourteenOutCanRemove(board([7]), 0, 1)).toBe(false);
  });

  it('exposes the target sum as a constant', () => {
    expect(FOURTEEN_OUT_TARGET_SUM).toBe(14);
  });
});

describe('countRemovablePairs', () => {
  it('returns 0 for a cleared board', () => {
    expect(countRemovablePairs(board())).toBe(0);
  });

  // 同じ組を 2 度数えない。7 が 3 枚なら 3 通り。
  it('counts each unordered pair once', () => {
    expect(countRemovablePairs(board([7], [7], [7]))).toBe(3);
  });

  it('counts only pairs that reach 14', () => {
    // 9 / 5 / 4 → 9+5 だけ。
    expect(countRemovablePairs(board([9], [5], [4]))).toBe(1);
  });

  it('returns 0 when nothing pairs', () => {
    expect(countRemovablePairs(board([2], [3], [4]))).toBe(0);
  });
});

describe('fourteenOutPartners', () => {
  it('lists every column whose tail completes the pair', () => {
    expect([...fourteenOutPartners(board([9], [5], [4], [5]), 0)]).toEqual([1, 3]);
  });

  it('never lists the source column', () => {
    expect(fourteenOutPartners(board([7], [7]), 0).has(0)).toBe(false);
  });

  it('is empty when no partner exists', () => {
    expect(fourteenOutPartners(board([2], [3]), 0).size).toBe(0);
  });
});
