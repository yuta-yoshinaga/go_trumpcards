import { describe, expect, it } from 'vitest';
import type { GrandfathersClockResponse } from '../../../types/card';
import { formatGrandfathersClockState } from './grandfathersclockFormatter';

function makeState(overrides?: Partial<GrandfathersClockResponse>): GrandfathersClockResponse {
  return {
    tableau: Array.from({ length: 8 }, () => []),
    foundation: Array.from({ length: 12 }, (_, i) => ({ cards: [], targetRank: i + 1, complete: false })),
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatGrandfathersClockState', () => {
  it('renders header and every face with its target', () => {
    const result = formatGrandfathersClockState(makeState());
    expect(result).toContain("Grandfather's Clock");
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
    expect(formatGrandfathersClockState(makeState({ foundation }))).toContain('done');
  });

  it('renders cards in the tableau', () => {
    const result = formatGrandfathersClockState(
      makeState({
        tableau: [[{ card: { design: 'SPADE', value: 9 }, faceUp: true }], ...Array.from({ length: 7 }, () => [])],
      }),
    );
    expect(result).toContain('[0]');
  });

  it('renders a placeholder for a missing card', () => {
    const result = formatGrandfathersClockState(
      makeState({ tableau: [[{ card: null, faceUp: true }], ...Array.from({ length: 7 }, () => [])] }),
    );
    expect(result).toContain('[?]');
  });

  it('shows a clock-face hint', () => {
    const result = formatGrandfathersClockState(makeState({ hint: { fromCol: 3, toZone: 'foundation', toIdx: 7 } }));
    expect(result).toContain('HINT');
    expect(result).toContain('f7');
  });

  it('shows a tableau hint', () => {
    const result = formatGrandfathersClockState(makeState({ hint: { fromCol: 3, toZone: 'tableau', toIdx: 5 } }));
    expect(result).toContain('t5');
  });

  it('shows stalemate message', () => {
    expect(formatGrandfathersClockState(makeState({ isStalemate: true }))).toContain('Stalemate');
  });

  it('shows congrats on win phase', () => {
    expect(formatGrandfathersClockState(makeState({ phase: 1 }))).toContain('Congratulations');
  });
});
