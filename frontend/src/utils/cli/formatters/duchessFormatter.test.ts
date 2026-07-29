import { describe, expect, it } from 'vitest';
import type { DuchessResponse } from '../../../types/card';
import { formatDuchessState } from './duchessFormatter';

function makeState(overrides?: Partial<DuchessResponse>): DuchessResponse {
  return {
    reserve: Array.from({ length: 4 }, () => []),
    tableau: Array.from({ length: 4 }, () => []),
    foundation: Array.from({ length: 4 }, () => []),
    stockCount: 35,
    waste: [],
    baseRank: 5,
    awaitingBaseRank: false,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatDuchessState', () => {
  it('renders header, stock and an empty board', () => {
    const result = formatDuchessState(makeState());
    expect(result).toContain('Duchess');
    expect(result).toContain('base rank: 5');
    expect(result).toContain('foundations:');
    expect(result).toContain('reserve:');
    expect(result).toContain('stock: 35');
    expect(result).toContain('t0: [empty]');
    expect(result).toContain('t3: [empty]');
  });

  // Until the base rank is set nothing else is legal, so it leads the board.
  it('prompts for the base rank when it is unset', () => {
    const result = formatDuchessState(makeState({ awaitingBaseRank: true, baseRank: 0 }));
    expect(result).toContain('choose one');
    expect(result).not.toContain('base rank: 0');
  });

  it('renders reserve fans with their depth', () => {
    const result = formatDuchessState(
      makeState({
        reserve: [
          [
            { design: 'SPADE', value: 3 },
            { design: 'HEART', value: 9 },
          ],
          [],
          [],
          [],
        ],
      }),
    );
    expect(result).toContain('r0:');
    expect(result).toContain('(2)');
    expect(result).toContain('r1:[  ]');
  });

  it('renders the waste top', () => {
    const result = formatDuchessState(makeState({ waste: [{ design: 'DIAMOND', value: 4 }] }));
    expect(result).toContain('waste:');
  });

  it('renders cards in the tableau', () => {
    const result = formatDuchessState(
      makeState({
        tableau: [[{ card: { design: 'SPADE', value: 9 }, faceUp: true }], ...Array.from({ length: 3 }, () => [])],
      }),
    );
    expect(result).toContain('[0]');
  });

  it('renders a placeholder for a missing card', () => {
    const result = formatDuchessState(
      makeState({ tableau: [[{ card: null, faceUp: true }], ...Array.from({ length: 3 }, () => [])] }),
    );
    expect(result).toContain('[?]');
  });

  it('shows a tableau hint with the run head', () => {
    const result = formatDuchessState(
      makeState({ hint: { fromZone: 'tableau', fromIdx: 3, cardIndex: 1, toZone: 'tableau', toIdx: 2 } }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('t3[1]');
  });

  // A draw has no destination pile index.
  it('renders a draw hint without a destination index', () => {
    const result = formatDuchessState(
      makeState({ hint: { fromZone: 'stock', fromIdx: -1, cardIndex: -1, toZone: 'waste', toIdx: -1 } }),
    );
    expect(result).toContain('stock');
    expect(result).toContain('draw');
  });

  it('shows stalemate message', () => {
    expect(formatDuchessState(makeState({ isStalemate: true }))).toContain('Stalemate');
  });

  it('shows the server message', () => {
    expect(formatDuchessState(makeState({ message: 'nope' }))).toContain('nope');
  });

  it('shows congrats on win phase', () => {
    expect(formatDuchessState(makeState({ phase: 1 }))).toContain('Congratulations');
  });
});
