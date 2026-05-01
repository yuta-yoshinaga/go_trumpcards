import { describe, expect, it } from 'vitest';
import type { CanfieldResponse } from '../../../types/card';
import { formatCanfieldState } from './canfieldFormatter';

function makeState(overrides?: Partial<CanfieldResponse>): CanfieldResponse {
  return {
    tableau: [[], [], [], []],
    reserve: [],
    stockCount: 0,
    waste: [],
    foundation: [[], [], [], []],
    baseRank: 1,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    message: '',
    ...overrides,
  };
}

describe('formatCanfieldState', () => {
  it('renders header and empty board', () => {
    const result = formatCanfieldState(makeState());
    expect(result).toContain('Canfield');
    expect(result).toContain('stock: 0');
    expect(result).toContain('reserve:');
    expect(result).toContain('foundation:');
    expect(result).toContain('t0: [empty]');
  });

  it('renders waste top card', () => {
    const result = formatCanfieldState(makeState({ waste: [{ design: 'SPADE', value: 5 }] }));
    expect(result).toContain('5');
  });

  it('shows hint when present', () => {
    const result = formatCanfieldState(
      makeState({
        hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 1, toZone: 'foundation', toCol: 2 },
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('foundation2');
  });

  it('shows congrats on win phase', () => {
    const result = formatCanfieldState(makeState({ phase: 1 }));
    expect(result).toContain('Congratulations');
  });
});
