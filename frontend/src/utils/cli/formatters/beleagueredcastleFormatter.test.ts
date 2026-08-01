import { describe, expect, it } from 'vitest';
import type { BeleagueredCastleResponse } from '../../../types/card';
import { formatBeleagueredCastleState } from './beleagueredcastleFormatter';

function makeState(overrides?: Partial<BeleagueredCastleResponse>): BeleagueredCastleResponse {
  return {
    tableau: Array.from({ length: 8 }, () => []),
    foundation: [[], [], [], []],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatBeleagueredCastleState', () => {
  it('renders header and empty board', () => {
    const result = formatBeleagueredCastleState(makeState());
    expect(result).toContain('Beleaguered Castle');
    expect(result).toContain('foundation:');
    expect(result).toContain('t0: [empty]');
  });

  it('renders cards in tableau', () => {
    const result = formatBeleagueredCastleState(
      makeState({
        tableau: [[{ card: { design: 'SPADE', value: 1 }, faceUp: true }], ...Array.from({ length: 7 }, () => [])],
      }),
    );
    expect(result).toContain('[0]');
  });

  it('renders foundation top card', () => {
    const result = formatBeleagueredCastleState(
      makeState({
        foundation: [[{ design: 'HEART', value: 1 }], [], [], []],
      }),
    );
    expect(result).toContain('foundation:');
  });

  it('shows hint when present', () => {
    const result = formatBeleagueredCastleState(
      makeState({
        hint: { fromCol: 0, cardIndex: 1, toZone: 'tableau', toCol: 2 },
        messageCode: 'beleagueredcastle.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('tableau2');
  });

  it('shows stalemate message', () => {
    const result = formatBeleagueredCastleState(makeState({ isStalemate: true }));
    expect(result).toContain('Stalemate');
  });

  it('shows congrats on win phase', () => {
    const result = formatBeleagueredCastleState(makeState({ phase: 1 }));
    expect(result).toContain('Congratulations');
  });
});
