import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, SlyFoxResponse } from '../../../types/card';
import { formatSlyFoxState } from './slyfoxFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeState(overrides: Partial<SlyFoxResponse> = {}): SlyFoxResponse {
  return {
    tableau: [[card('SPADE', 1)], [], [card('HEART', 9), card('CLOVER', 4)]],
    foundation: [[card('SPADE', 1)], [], [], [], [card('SPADE', 13)], [], [], []],
    foundationAscending: [true, true, true, true, false, false, false, false],
    stockCount: 71,
    dealtThisCycle: 20,
    dealCycle: 20,
    reserveLocked: false,
    phase: 0,
    moveCount: 13,
    canUndo: true,
    message: '',
    ...overrides,
  };
}

describe('formatSlyFoxState', () => {
  it('marks which foundations build up and which build down', () => {
    const out = formatSlyFoxState(makeState());
    const line = out.split('\n').find((l) => l.startsWith('foundation:'));
    expect(line).toBeDefined();
    expect(line).toContain('↑');
    expect(line).toContain('↓');
  });

  it('shows the stock count', () => {
    expect(formatSlyFoxState(makeState())).toContain('stock: 71');
  });

  // **周の進みは盤から読めない。**書かないと、リザーブが送れない理由が分からない。
  it('says how many more cards open the reserve while it is locked', () => {
    const out = formatSlyFoxState(makeState({ reserveLocked: true, dealtThisCycle: 13, dealCycle: 20 }));
    expect(out).toContain('13/20 dealt');
    expect(out).toContain('7 more');
    expect(out).not.toContain('the reserve is open');
  });

  // 負のコントロール: 開いていれば開いていると言う。片方しか出ないと、
  // 上のテストは「常にロック文言」でも通る。
  it('says the reserve is open once the round is dealt out', () => {
    expect(formatSlyFoxState(makeState({ reserveLocked: false }))).toContain('the reserve is open');
  });

  it('lists every pile with its depth', () => {
    const out = formatSlyFoxState(makeState());
    expect(out).toContain('slot0:');
    expect(out).toContain('slot1: [  ] (0)');
    expect(out).toContain('slot2:');
    expect(out).toContain('(2)');
  });

  it('shows the move count and any message', () => {
    const out = formatSlyFoxState(makeState({ message: 'no moves left' }));
    expect(out).toContain('moves: 13');
    expect(out).toContain('no moves left');
  });
});
