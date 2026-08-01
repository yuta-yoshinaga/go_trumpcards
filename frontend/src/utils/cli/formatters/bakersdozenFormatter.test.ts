import { describe, expect, it } from 'vitest';
import type { BakersDozenResponse } from '../../../types/card';
import { formatBakersDozenState } from './bakersdozenFormatter';

function makeState(overrides?: Partial<BakersDozenResponse>): BakersDozenResponse {
  return {
    tableau: Array.from({ length: 13 }, () => []),
    foundation: [[], [], [], []],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatBakersDozenState', () => {
  it('renders header and empty board', () => {
    const result = formatBakersDozenState(makeState());
    expect(result).toContain("Baker's Dozen");
    expect(result).toContain('foundation:');
    expect(result).toContain('t0: [empty]');
  });

  it('renders cards in tableau', () => {
    const result = formatBakersDozenState(
      makeState({
        tableau: [[{ card: { design: 'SPADE', value: 1 }, faceUp: true }], ...Array.from({ length: 12 }, () => [])],
      }),
    );
    expect(result).toContain('[0]');
  });

  it('renders foundation top card', () => {
    const result = formatBakersDozenState(
      makeState({
        foundation: [[{ design: 'HEART', value: 1 }], [], [], []],
      }),
    );
    expect(result).toContain('foundation:');
  });

  it('shows hint when present', () => {
    const result = formatBakersDozenState(
      makeState({
        hint: { fromCol: 0, cardIndex: 1, toZone: 'tableau', toCol: 2 },
        messageCode: 'bakersdozen.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('tableau2');
  });

  it('shows stalemate message', () => {
    const result = formatBakersDozenState(makeState({ isStalemate: true }));
    expect(result).toContain('Stalemate');
  });

  it('shows congrats on win phase', () => {
    const result = formatBakersDozenState(makeState({ phase: 1 }));
    expect(result).toContain('Congratulations');
  });
});
