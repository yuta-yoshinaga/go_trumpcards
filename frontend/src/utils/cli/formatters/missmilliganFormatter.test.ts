import { describe, expect, it } from 'vitest';
import type { MissMilliganResponse } from '../../../types/card';
import { formatMissMilliganState } from './missmilliganFormatter';

function makeState(overrides?: Partial<MissMilliganResponse>): MissMilliganResponse {
  return {
    tableau: Array.from({ length: 8 }, () => []),
    stockCount: 96,
    foundation: Array.from({ length: 8 }, () => []),
    waived: [],
    canWaive: false,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatMissMilliganState', () => {
  it('renders header, stock and an empty board', () => {
    const result = formatMissMilliganState(makeState());
    expect(result).toContain('Miss Milligan');
    expect(result).toContain('foundations:');
    expect(result).toContain('stock: 96');
    expect(result).toContain('t0: [empty]');
    expect(result).toContain('t7: [empty]');
  });

  it('announces that waiving is available', () => {
    expect(formatMissMilliganState(makeState({ stockCount: 0, canWaive: true }))).toContain('waiving available');
  });

  // Holding cards blocks dealing, so it belongs on the status line.
  it('shows the waived cards instead of the availability note', () => {
    const result = formatMissMilliganState(
      makeState({ stockCount: 0, canWaive: false, waived: [{ design: 'HEART', value: 8 }] }),
    );
    expect(result).toContain('waived:');
    expect(result).not.toContain('waiving available');
  });

  it('renders cards in the tableau', () => {
    const result = formatMissMilliganState(
      makeState({
        tableau: [[{ card: { design: 'SPADE', value: 9 }, faceUp: true }], ...Array.from({ length: 7 }, () => [])],
      }),
    );
    expect(result).toContain('[0]');
  });

  it('renders a placeholder for a missing card', () => {
    const result = formatMissMilliganState(
      makeState({ tableau: [[{ card: null, faceUp: true }], ...Array.from({ length: 7 }, () => [])] }),
    );
    expect(result).toContain('[?]');
  });

  it('shows a tableau hint with the run head', () => {
    const result = formatMissMilliganState(
      makeState({ hint: { fromZone: 'tableau', fromCol: 3, cardIndex: 1, toZone: 'tableau', toIdx: 7 } }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('t3[1]');
  });

  // A deal targets every column, so it has no single destination index.
  it('renders a deal hint without a destination index', () => {
    const result = formatMissMilliganState(
      makeState({ hint: { fromZone: 'stock', fromCol: -1, cardIndex: -1, toZone: 'tableau', toIdx: -1 } }),
    );
    expect(result).toContain('stock');
    expect(result).toContain('deal');
  });

  it('shows stalemate message', () => {
    expect(formatMissMilliganState(makeState({ isStalemate: true }))).toContain('Stalemate');
  });

  it('shows congrats on win phase', () => {
    expect(formatMissMilliganState(makeState({ phase: 1 }))).toContain('Congratulations');
  });
});
