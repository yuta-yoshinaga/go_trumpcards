import { describe, expect, it } from 'vitest';
import type { CitadelResponse } from '../../../types/card';
import { formatCitadelState } from './citadelFormatter';

function makeState(overrides?: Partial<CitadelResponse>): CitadelResponse {
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

describe('formatCitadelState', () => {
  it('renders header and empty board', () => {
    const result = formatCitadelState(makeState());
    expect(result).toContain('Beleaguered Castle');
    expect(result).toContain('foundation:');
    expect(result).toContain('t0: [empty]');
  });

  it('renders cards in tableau', () => {
    const result = formatCitadelState(
      makeState({
        tableau: [[{ card: { design: 'SPADE', value: 1 }, faceUp: true }], ...Array.from({ length: 7 }, () => [])],
      }),
    );
    expect(result).toContain('[0]');
  });

  it('renders foundation top card', () => {
    const result = formatCitadelState(
      makeState({
        foundation: [[{ design: 'HEART', value: 1 }], [], [], []],
      }),
    );
    expect(result).toContain('foundation:');
  });

  it('shows hint when present', () => {
    const result = formatCitadelState(
      makeState({
        hint: { fromCol: 0, cardIndex: 1, toZone: 'tableau', toCol: 2 },
        messageCode: 'citadel.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('tableau2');
  });

  // **頼んでいないヒントは CLI に出さない。**#4483 以降 Output() もヒントを載せる
  // ので、state.hint だけを見ると毎手 HINT が印字される。
  it('does not print a passive hint carried on an ordinary response', () => {
    const result = formatCitadelState(
      makeState({
        hint: { fromCol: 0, cardIndex: 1, toZone: 'tableau', toCol: 2 },
        messageCode: 'citadel.playing',
      }),
    );
    expect(result).not.toContain('HINT');
  });

  it('shows stalemate message', () => {
    const result = formatCitadelState(makeState({ isStalemate: true }));
    expect(result).toContain('Stalemate');
  });

  it('shows congrats on win phase', () => {
    const result = formatCitadelState(makeState({ phase: 1 }));
    expect(result).toContain('Congratulations');
  });
});
