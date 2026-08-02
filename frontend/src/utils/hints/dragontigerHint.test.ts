import { describe, expect, it } from 'vitest';
import type { DragonTigerResponse } from '../../types/card';
import { DragonTigerBetType, DragonTigerPhase } from '../../types/phases';
import { getDragontigerHint } from './dragontigerHint';

function base(overrides: Partial<DragonTigerResponse> = {}) {
  return {
    phase: DragonTigerPhase.BET,
    chips: 100,
    betAmount: 10,
    betType: DragonTigerBetType.DRAGON,
    result: 0,
    payout: 0,
    history: [],
    message: '',
    ...overrides,
  } as unknown as DragonTigerResponse;
}

describe('getDragontigerHint', () => {
  it('stays quiet outside the betting phase', () => {
    expect(getDragontigerHint(base({ phase: DragonTigerPhase.END }))).toBeNull();
  });

  // **タイは 8:1 だが約 13 分の 1。**期待値で負ける。
  it('warns about the tie odds', () => {
    expect(getDragontigerHint(base({ betType: DragonTigerBetType.TIE }))).toEqual({
      targetAction: 'bet',
      reason: 'frontendHint.dragontigerTieOdds',
      confidence: 'moderate',
    });
  });

  it('treats a dragon stake as the even-money bet', () => {
    expect(getDragontigerHint(base({ betType: DragonTigerBetType.DRAGON }))?.reason).toBe(
      'frontendHint.dragontigerEvenMoney',
    );
  });

  // **Dragon と Tiger は同じ扱い。**片方だけ別扱いにしない。
  it('treats a tiger stake the same as a dragon stake', () => {
    expect(getDragontigerHint(base({ betType: DragonTigerBetType.TIGER }))?.reason).toBe(
      'frontendHint.dragontigerEvenMoney',
    );
  });
});
