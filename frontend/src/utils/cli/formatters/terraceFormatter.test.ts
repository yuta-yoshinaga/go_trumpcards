import { describe, expect, it } from 'vitest';
import type { TerraceResponse } from '../../../types/card';
import { formatTerraceState } from './terraceFormatter';

function makeState(overrides?: Partial<TerraceResponse>): TerraceResponse {
  return {
    reserve: [],
    tableau: Array.from({ length: 9 }, () => []),
    foundation: Array.from({ length: 8 }, () => []),
    stockCount: 84,
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

describe('formatTerraceState', () => {
  it('renders header, piles and an empty board', () => {
    const result = formatTerraceState(makeState());
    expect(result).toContain('Terrace');
    expect(result).toContain('base rank: 5');
    expect(result).toContain('foundations:');
    expect(result).toContain('stock: 84');
    expect(result).toContain('t0: [empty]');
    expect(result).toContain('t8: [empty]');
  });

  it('prompts for the base rank when it is unset', () => {
    const result = formatTerraceState(makeState({ awaitingBaseRank: true, baseRank: 0 }));
    expect(result).toContain('not set yet');
    expect(result).not.toContain('base rank: 0');
  });

  // The terrace never refills, so its depth and its restriction are status.
  it('shows the terrace depth and that it is foundations-only', () => {
    const result = formatTerraceState(
      makeState({
        reserve: [
          { design: 'SPADE', value: 3 },
          { design: 'HEART', value: 9 },
        ],
      }),
    );
    expect(result).toContain('(2, foundations only)');
  });

  it('renders the waste top', () => {
    expect(formatTerraceState(makeState({ waste: [{ design: 'DIAMOND', value: 4 }] }))).toContain('waste:');
  });

  it('renders cards in a pile', () => {
    const result = formatTerraceState(
      makeState({ tableau: [[{ design: 'SPADE', value: 9 }], ...Array.from({ length: 8 }, () => [])] }),
    );
    expect(result).toContain('[0]');
  });

  it('shows a tableau hint with its pile', () => {
    const result = formatTerraceState(
      makeState({ hint: { fromZone: 'tableau', fromIdx: 3, toZone: 'foundation', toIdx: 2 } }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('t3');
    expect(result).toContain('foundation2');
  });

  it('renders a draw hint without indices', () => {
    const result = formatTerraceState(
      makeState({ hint: { fromZone: 'stock', fromIdx: -1, toZone: 'waste', toIdx: -1 } }),
    );
    expect(result).toContain('stock → waste');
  });

  it('shows stalemate message', () => {
    expect(formatTerraceState(makeState({ isStalemate: true }))).toContain('Stalemate');
  });

  it('shows the server message', () => {
    expect(formatTerraceState(makeState({ message: 'nope' }))).toContain('nope');
  });

  it('shows congrats on win phase', () => {
    expect(formatTerraceState(makeState({ phase: 1 }))).toContain('Congratulations');
  });
});
