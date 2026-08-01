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
        messageCode: 'canfield.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('foundation2');
  });

  it('shows congrats on win phase', () => {
    const result = formatCanfieldState(makeState({ phase: 1 }));
    expect(result).toContain('Congratulations');
  });

  it('renders tableau cards with indices', () => {
    const result = formatCanfieldState(
      makeState({
        tableau: [[{ card: { design: 'CLOVER', value: 5 } }], [], [], []],
      }),
    );
    expect(result).toContain('[0]');
  });

  it('renders foundation top cards', () => {
    const result = formatCanfieldState(
      makeState({
        foundation: [[{ design: 'HEART', value: 1 }], [], [], []],
      }),
    );
    expect(result).toContain('foundation:');
  });

  it('renders reserve top card when present', () => {
    const result = formatCanfieldState(
      makeState({
        reserve: [{ design: 'SPADE', value: 9 }],
      }),
    );
    expect(result).toContain('reserve:');
  });

  it('renders message when present', () => {
    const result = formatCanfieldState(makeState({ message: 'No moves available' }));
    expect(result).toContain('No moves available');
  });

  it('handles foundation hint with negative toCol', () => {
    const result = formatCanfieldState(
      makeState({
        hint: { fromZone: 'waste', fromCol: -1, cardIndex: 0, toZone: 'foundation', toCol: -1 },
        messageCode: 'canfield.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('→ foundation');
  });
});
