import { describe, expect, it } from 'vitest';
import type { PerseveranceResponse } from '../../../types/card';
import { formatPerseveranceState } from './perseveranceFormatter';

function makeState(overrides?: Partial<PerseveranceResponse>): PerseveranceResponse {
  return {
    tableau: Array.from({ length: 13 }, () => []),
    foundation: [[], [], [], []],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    redealsLeft: 2,
    message: '',
    ...overrides,
  };
}

describe('formatPerseveranceState', () => {
  it('renders header and empty board', () => {
    const result = formatPerseveranceState(makeState());
    expect(result).toContain('Perseverance');
    expect(result).toContain('foundation:');
    expect(result).toContain('t0: [empty]');
  });

  it('renders cards in tableau', () => {
    const result = formatPerseveranceState(
      makeState({
        tableau: [[{ card: { design: 'SPADE', value: 1 }, faceUp: true }], ...Array.from({ length: 12 }, () => [])],
      }),
    );
    expect(result).toContain('[0]');
  });

  it('renders foundation top card', () => {
    const result = formatPerseveranceState(
      makeState({
        foundation: [[{ design: 'HEART', value: 1 }], [], [], []],
      }),
    );
    expect(result).toContain('foundation:');
  });

  it('shows hint when present', () => {
    const result = formatPerseveranceState(
      makeState({
        hint: { fromCol: 0, cardIndex: 1, toZone: 'tableau', toCol: 2 },
        messageCode: 'perseverance.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('tableau2');
  });

  it('shows stalemate message', () => {
    const result = formatPerseveranceState(makeState({ isStalemate: true }));
    expect(result).toContain('Stalemate');
  });

  it('shows congrats on win phase', () => {
    const result = formatPerseveranceState(makeState({ phase: 1 }));
    expect(result).toContain('Congratulations');
  });
});
