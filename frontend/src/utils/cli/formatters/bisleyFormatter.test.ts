import { describe, expect, it } from 'vitest';
import type { BisleyResponse } from '../../../types/card';
import { formatBisleyState } from './bisleyFormatter';

function makeState(overrides?: Partial<BisleyResponse>): BisleyResponse {
  return {
    tableau: Array.from({ length: 13 }, () => []),
    aceFoundations: [[], [], [], []],
    kingFoundations: [[], [], [], []],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatBisleyState', () => {
  it('renders header and empty board', () => {
    const result = formatBisleyState(makeState());
    expect(result).toContain('Bisley');
    expect(result).toContain('ascending  (A->K):');
    expect(result).toContain('descending (K->A):');
    expect(result).toContain('t0: [empty]');
    expect(result).toContain('t12: [empty]');
  });

  it('renders cards in tableau', () => {
    const result = formatBisleyState(
      makeState({
        tableau: [[{ card: { design: 'SPADE', value: 2 }, faceUp: true }], ...Array.from({ length: 12 }, () => [])],
      }),
    );
    expect(result).toContain('[0]');
  });

  it('renders both foundation top cards', () => {
    const result = formatBisleyState(
      makeState({
        aceFoundations: [[{ design: 'HEART', value: 1 }], [], [], []],
        kingFoundations: [[{ design: 'HEART', value: 13 }], [], [], []],
      }),
    );
    expect(result).toContain('ascending  (A->K):');
    expect(result).toContain('descending (K->A):');
  });

  it('shows a tableau hint', () => {
    const result = formatBisleyState(
      makeState({ hint: { fromCol: 0, toZone: 'tableau', toIdx: 2 }, messageCode: 'bisley.hintAvailable' }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('t2');
  });

  it('shows a foundation hint', () => {
    const result = formatBisleyState(
      makeState({ hint: { fromCol: 3, toZone: 'king', toIdx: 1 }, messageCode: 'bisley.hintAvailable' }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('king1');
  });

  it('shows stalemate message', () => {
    const result = formatBisleyState(makeState({ isStalemate: true }));
    expect(result).toContain('Stalemate');
  });

  it('shows congrats on win phase', () => {
    const result = formatBisleyState(makeState({ phase: 1 }));
    expect(result).toContain('Congratulations');
  });
});
