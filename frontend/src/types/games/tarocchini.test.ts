import { describe, expect, it } from 'vitest';
import { TAROCCHINI_PLAYER_COUNT, TAROCCHINI_SURPLUS, tarocchiniTeamOf } from './tarocchini';

describe('tarocchiniTeamOf', () => {
  // 対面同士が組む。ここがずれると味方のトリックを奪う手が正しく見えてしまう。
  it('pairs the opposite seats', () => {
    expect(tarocchiniTeamOf(0)).toBe(tarocchiniTeamOf(2));
    expect(tarocchiniTeamOf(1)).toBe(tarocchiniTeamOf(3));
    expect(tarocchiniTeamOf(0)).not.toBe(tarocchiniTeamOf(1));
  });

  it('produces exactly two teams across every seat', () => {
    const teams = new Set(Array.from({ length: TAROCCHINI_PLAYER_COUNT }, (_, i) => tarocchiniTeamOf(i)));
    expect(teams.size).toBe(2);
  });
});

describe('deck arithmetic', () => {
  // 62 は 4 で割り切れない。余りが構造として残ることを固定しておく。
  it('leaves a surplus the dealer must bury', () => {
    const deck = 62;
    const hand = 15;
    expect(deck - TAROCCHINI_PLAYER_COUNT * hand).toBe(TAROCCHINI_SURPLUS);
  });
});
