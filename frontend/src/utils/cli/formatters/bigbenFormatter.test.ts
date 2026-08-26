import { describe, expect, it } from 'vitest';
import type { BigBenResponse } from '../../../types/card';
import { formatBigBenState } from './bigbenFormatter';

function makeState(overrides?: Partial<BigBenResponse>): BigBenResponse {
  return {
    tableau: Array.from({ length: 8 }, () => []),
    foundation: Array.from({ length: 12 }, (_, i) => ({ cards: [], targetRank: i + 1, complete: false })),
    phase: 0,
    moveCount: 0,
    stockCount: 52,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatBigBenState', () => {
  it('renders header and every face with its target', () => {
    const result = formatBigBenState(makeState());
    expect(result).toContain('Big Ben');
    expect(result).toContain('f0: [  ] -> 1');
    expect(result).toContain('f11: [  ] -> 12');
    expect(result).toContain('t0: [empty]');
    expect(result).toContain('t7: [empty]');
  });

  it('marks a completed face', () => {
    const foundation = Array.from({ length: 12 }, (_, i) => ({
      cards: i === 0 ? [{ design: 'HEART' as const, value: 1 }] : [],
      targetRank: i + 1,
      complete: i === 0,
    }));
    expect(formatBigBenState(makeState({ foundation }))).toContain('done');
  });

  it('renders cards in the tableau', () => {
    const result = formatBigBenState(
      makeState({
        tableau: [[{ card: { design: 'SPADE', value: 9 }, faceUp: true }], ...Array.from({ length: 7 }, () => [])],
      }),
    );
    expect(result).toContain('[0]');
  });

  it('renders a placeholder for a missing card', () => {
    const result = formatBigBenState(
      makeState({ tableau: [[{ card: null, faceUp: true }], ...Array.from({ length: 7 }, () => [])] }),
    );
    expect(result).toContain('[?]');
  });

  it('shows a clock-face hint', () => {
    const result = formatBigBenState(
      makeState({
        hint: { fromZone: 'tableau', fromCol: 3, toZone: 'foundation', toIdx: 7 },
        messageCode: 'bigben.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('f7');
  });

  // **頼んでいないヒントは CLI に出さない。**#4483 以降 Output() もヒントを載せる。
  it('does not print a passive hint carried on an ordinary response', () => {
    const result = formatBigBenState(
      makeState({
        hint: { fromZone: 'tableau', fromCol: 3, toZone: 'foundation', toIdx: 7 },
        messageCode: 'bigben.playing',
      }),
    );
    expect(result).not.toContain('HINT');
  });

  it('shows a tableau hint', () => {
    const result = formatBigBenState(
      makeState({
        hint: { fromZone: 'tableau', fromCol: 3, toZone: 'tableau', toIdx: 5 },
        messageCode: 'bigben.hintAvailable',
      }),
    );
    expect(result).toContain('t5');
  });

  it('shows stalemate message', () => {
    expect(formatBigBenState(makeState({ isStalemate: true }))).toContain('Stalemate');
  });

  it('shows congrats on win phase', () => {
    expect(formatBigBenState(makeState({ phase: 1 }))).toContain('Congratulations');
  });
});
