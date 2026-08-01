import { describe, expect, it } from 'vitest';
import type { KingAlbertResponse } from '../../../types/card';
import { formatKingAlbertState } from './kingalbertFormatter';

function makeState(overrides?: Partial<KingAlbertResponse>): KingAlbertResponse {
  return {
    tableau: Array.from({ length: 9 }, () => []),
    reserve: Array.from({ length: 7 }, () => null),
    foundation: [[], [], [], []],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatKingAlbertState', () => {
  it('renders header and empty board', () => {
    const result = formatKingAlbertState(makeState());
    expect(result).toContain('King Albert');
    expect(result).toContain('foundation:');
    expect(result).toContain('reserve:');
    expect(result).toContain('t0: [empty]');
  });

  it('renders cards in tableau', () => {
    const result = formatKingAlbertState(
      makeState({
        tableau: [[{ card: { design: 'SPADE', value: 1 }, faceUp: true }], ...Array.from({ length: 8 }, () => [])],
      }),
    );
    expect(result).toContain('[0]');
  });

  it('renders reserve cards', () => {
    const result = formatKingAlbertState(
      makeState({
        reserve: [{ design: 'HEART', value: 7 }, null, null, null, null, null, null],
      }),
    );
    expect(result).toContain('reserve:');
  });

  it('renders foundation top card', () => {
    const result = formatKingAlbertState(
      makeState({
        foundation: [[{ design: 'HEART', value: 1 }], [], [], []],
      }),
    );
    expect(result).toContain('foundation:');
  });

  it('shows hint from tableau when present', () => {
    const result = formatKingAlbertState(
      makeState({
        hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 1, toZone: 'tableau', toCol: 2 },
        messageCode: 'kingalbert.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('t0[1]');
    expect(result).toContain('tableau2');
  });

  // **頼んでいないヒントは CLI に出さない。**#4483 以降 Output() もヒントを載せる。
  it('does not print a passive hint carried on an ordinary response', () => {
    const result = formatKingAlbertState(
      makeState({
        hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 1, toZone: 'tableau', toCol: 2 },
        messageCode: 'kingalbert.playing',
      }),
    );
    expect(result).not.toContain('HINT');
  });

  it('shows hint from reserve when present', () => {
    const result = formatKingAlbertState(
      makeState({
        hint: { fromZone: 'reserve', fromCol: 3, cardIndex: 0, toZone: 'foundation', toCol: 1 },
        messageCode: 'kingalbert.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('r3');
  });

  it('shows stalemate message', () => {
    const result = formatKingAlbertState(makeState({ isStalemate: true }));
    expect(result).toContain('Stalemate');
  });

  it('shows congrats on win phase', () => {
    const result = formatKingAlbertState(makeState({ phase: 1 }));
    expect(result).toContain('Congratulations');
  });
});
