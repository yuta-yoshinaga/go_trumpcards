import { describe, expect, it } from 'vitest';
import type { NapoleonsSquareResponse } from '../../../types/card';
import { formatNapoleonsSquareState } from './napoleonssquareFormatter';

function makeState(overrides?: Partial<NapoleonsSquareResponse>): NapoleonsSquareResponse {
  return {
    tableau: Array.from({ length: 12 }, () => []),
    stockCount: 48,
    waste: [],
    foundation: Array.from({ length: 8 }, () => []),
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatNapoleonsSquareState', () => {
  it('renders header, stock and an empty board', () => {
    const result = formatNapoleonsSquareState(makeState());
    expect(result).toContain("Napoleon's Square");
    expect(result).toContain('foundations:');
    expect(result).toContain('stock: 48');
    expect(result).toContain('t0: [empty]');
    expect(result).toContain('t11: [empty]');
  });

  it('renders the waste top card', () => {
    const result = formatNapoleonsSquareState(makeState({ waste: [{ design: 'HEART', value: 7 }] }));
    expect(result).toContain('waste:');
    expect(result).toContain('(1)');
  });

  it('renders cards in the tableau', () => {
    const result = formatNapoleonsSquareState(
      makeState({
        tableau: [[{ card: { design: 'SPADE', value: 9 }, faceUp: true }], ...Array.from({ length: 11 }, () => [])],
      }),
    );
    expect(result).toContain('[0]');
  });

  it('shows a tableau hint with the run head', () => {
    const result = formatNapoleonsSquareState(
      makeState({ hint: { fromZone: 'tableau', fromCol: 3, cardIndex: 1, toZone: 'tableau', toCol: 7 } }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('t3[1]');
    expect(result).toContain('tableau7');
  });

  it('shows a stock hint without a destination index', () => {
    const result = formatNapoleonsSquareState(
      makeState({ hint: { fromZone: 'stock', fromCol: -1, cardIndex: -1, toZone: 'waste', toCol: -1 } }),
    );
    expect(result).toContain('stock');
    expect(result).toContain('waste');
  });

  it('shows stalemate message', () => {
    expect(formatNapoleonsSquareState(makeState({ isStalemate: true }))).toContain('Stalemate');
  });

  it('shows congrats on win phase', () => {
    expect(formatNapoleonsSquareState(makeState({ phase: 1 }))).toContain('Congratulations');
  });

  // A null card can reach the formatter from a partially-restored KV snapshot;
  // it must render a placeholder rather than throw.
  it('renders a placeholder for a missing card', () => {
    const result = formatNapoleonsSquareState(
      makeState({
        tableau: [[{ card: null, faceUp: true }], ...Array.from({ length: 11 }, () => [])],
      }),
    );
    expect(result).toContain('[?]');
  });

  it('renders a waste-to-foundation hint', () => {
    const result = formatNapoleonsSquareState(
      makeState({ hint: { fromZone: 'waste', fromCol: -1, cardIndex: -1, toZone: 'foundation', toCol: 2 } }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('waste');
    expect(result).toContain('foundation2');
  });
});
