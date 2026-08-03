import { describe, expect, it } from 'vitest';
import { MINCHIATE_PLAYER_COUNT, MINCHIATE_SURPLUS, minchiateTeamOf } from './minchiate';

describe('minchiateTeamOf', () => {
  // 対面同士が組む。ここがずれると味方のトリックを奪う手が正しく見えてしまう。
  it('pairs the opposite seats', () => {
    expect(minchiateTeamOf(0)).toBe(minchiateTeamOf(2));
    expect(minchiateTeamOf(1)).toBe(minchiateTeamOf(3));
    expect(minchiateTeamOf(0)).not.toBe(minchiateTeamOf(1));
  });

  it('produces exactly two teams across every seat', () => {
    const teams = new Set(Array.from({ length: MINCHIATE_PLAYER_COUNT }, (_, i) => minchiateTeamOf(i)));
    expect(teams.size).toBe(2);
  });
});

describe('deck arithmetic', () => {
  // 97 は 4 で割り切れる (12 枚ずつ) が全部は配らない。余りが残ることを固定する。
  it('leaves a surplus the dealer must bury', () => {
    const deck = 97;
    const hand = 21;
    expect(deck - MINCHIATE_PLAYER_COUNT * hand).toBe(MINCHIATE_SURPLUS);
  });
});
