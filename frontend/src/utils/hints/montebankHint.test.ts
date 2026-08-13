import { describe, expect, it } from 'vitest';
import type { Card, MonteBankResponse } from '../../types/card';
import { MonteBankPhase } from '../../types/phases';
import { getMontebankHint } from './montebankHint';

const card = (design: string, value: number): Card => ({ design, value }) as Card;

const entry = (suitCount: number) =>
  ({
    card: card('SPADE', 1),
    suitCount,
    remainingOfSuit: 10 - suitCount,
    isEven: suitCount === 1,
    isPicked: false,
  }) as MonteBankResponse['layout'][number];

const state = (over: Partial<MonteBankResponse> = {}) =>
  ({
    phase: MonteBankPhase.BET,
    layout: [entry(2), entry(2), entry(1), entry(1)],
    pick: -1,
    bet: 0,
    result: 0,
    payout: 0,
    chips: 1000,
    roundNumber: 1,
    remainingCards: 35,
    gameEndFlag: false,
    payoutMultiplier: 3,
    message: '',
    ...over,
  }) as MonteBankResponse;

describe('getMontebankHint', () => {
  it('終局・決着後・場札なしでは助言しない', () => {
    expect(getMontebankHint(state({ gameEndFlag: true }))).toBeNull();
    expect(getMontebankHint(state({ phase: MonteBankPhase.RESULT }))).toBeNull();
    expect(getMontebankHint(state({ layout: [] }))).toBeNull();
  });

  // **1 枚だけのスートがあるなら、それが唯一の互角の賭け。**
  it('互角の札があればそう言う', () => {
    const hint = getMontebankHint(state());
    expect(hint?.reason).toBe('frontendHint.monteBankLoneSuit');
    expect(hint?.confidence).toBe('strong');
  });

  // **どれも重複しているときに「良い手がある」と言わない。**
  it('どれも重複していればそう言う', () => {
    const hint = getMontebankHint(state({ layout: [entry(2), entry(2), entry(2), entry(2)] }));
    expect(hint?.reason).toBe('frontendHint.monteBankAllDuplicated');
    expect(hint?.confidence).toBe('moderate');
  });

  it('4枚同スートでも助言は返す', () => {
    const hint = getMontebankHint(state({ layout: [entry(4), entry(4), entry(4), entry(4)] }));
    expect(hint?.reason).toBe('frontendHint.monteBankAllDuplicated');
  });
});
