import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, PerseveranceTableauCard } from '../types/card';
import { perseveranceStartsRun } from './perseveranceStartsRun';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const tc = (design: CardDesign, value: number): PerseveranceTableauCard => ({
  card: card(design, value),
  faceUp: true,
});

describe('perseveranceStartsRun', () => {
  // ♥K ♠9 ♠8 ♠7 — 上3枚が並び、♥K で切れる。
  const col: PerseveranceTableauCard[] = [tc('HEART', 13), tc('SPADE', 9), tc('SPADE', 8), tc('SPADE', 7)];

  it('accepts the top card and every index the run reaches', () => {
    expect(perseveranceStartsRun(col, 3)).toBe(true);
    expect(perseveranceStartsRun(col, 2)).toBe(true);
    expect(perseveranceStartsRun(col, 1)).toBe(true);
  });

  // 負のコントロール: 並びが切れた下の札は掴めない。
  it('refuses an index below the break', () => {
    expect(perseveranceStartsRun(col, 0)).toBe(false);
  });

  it('refuses a group whose suits differ', () => {
    expect(perseveranceStartsRun([tc('DIAMOND', 8), tc('SPADE', 7)], 0)).toBe(false);
  });

  it('refuses a group that skips a rank', () => {
    expect(perseveranceStartsRun([tc('SPADE', 8), tc('SPADE', 6)], 0)).toBe(false);
  });

  it('refuses an out-of-range index', () => {
    expect(perseveranceStartsRun(col, -1)).toBe(false);
    expect(perseveranceStartsRun(col, 99)).toBe(false);
  });
});
