import { describe, expect, it } from 'vitest';
import type { AgnesResponse } from '../../../types/card';
import { formatAgnesState } from './agnesFormatter';

function makeState(overrides?: Partial<AgnesResponse>): AgnesResponse {
  return {
    tableau: [[], [], [], [], [], [], []],
    stockCount: 0,
    foundation: [[], [], [], []],
    baseRank: 1,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    message: '',
    ...overrides,
  };
}

describe('formatAgnesState', () => {
  it('renders header and empty board', () => {
    const result = formatAgnesState(makeState());
    expect(result).toContain('Agnes Sorel');
    expect(result).toContain('stock: 0');
    expect(result).toContain('foundation:');
    expect(result).toContain('t0: [empty]');
  });

  it('renders face-up tableau cards with indices', () => {
    const result = formatAgnesState(
      makeState({
        tableau: [[{ card: { design: 'CLOVER', value: 5 }, faceUp: true }], [], [], [], [], [], []],
      }),
    );
    expect(result).toContain('[0]');
    expect(result).not.toContain('[0]??');
  });

  it('renders face-down tableau cards as ??', () => {
    const result = formatAgnesState(
      makeState({
        tableau: [[{ card: { design: 'CLOVER', value: 5 }, faceUp: false }], [], [], [], [], [], []],
      }),
    );
    expect(result).toContain('??');
  });

  it('renders face-down tableau card with null card as ??', () => {
    const result = formatAgnesState(
      makeState({
        tableau: [[{ card: null, faceUp: false }], [], [], [], [], [], []],
      }),
    );
    expect(result).toContain('??');
  });

  it('shows hint when present', () => {
    const result = formatAgnesState(
      makeState({
        hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 1, toZone: 'foundation', toCol: 2 },
        messageCode: 'agnes.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('foundation2');
  });

  it('handles hint with negative toCol', () => {
    const result = formatAgnesState(
      makeState({
        hint: { fromZone: 'tableau', fromCol: -1, cardIndex: 0, toZone: 'foundation', toCol: -1 },
        messageCode: 'agnes.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('→ foundation');
  });

  it('shows congrats on win phase', () => {
    const result = formatAgnesState(makeState({ phase: 1 }));
    expect(result).toContain('Congratulations');
  });

  it('renders foundation top cards', () => {
    const result = formatAgnesState(
      makeState({
        foundation: [[{ design: 'HEART', value: 1 }], [], [], []],
      }),
    );
    expect(result).toContain('foundation:');
  });

  it('renders message when present', () => {
    const result = formatAgnesState(makeState({ message: 'No moves available' }));
    expect(result).toContain('No moves available');
  });
});
