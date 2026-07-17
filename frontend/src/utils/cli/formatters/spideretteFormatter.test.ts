import { describe, expect, it } from 'vitest';
import type { SpideretteResponse } from '../../../types/card';
import { formatSpideretteState } from './spideretteFormatter';

const baseState: SpideretteResponse = {
  tableau: [
    [{ card: { design: 'SPADE', value: 13 }, faceUp: true }],
    [],
    [
      { card: null, faceUp: false },
      { card: { design: 'HEART', value: 5 }, faceUp: true },
    ],
    [],
    [],
    [],
    [],
  ],
  stockCount: 20,
  completedSuits: 1,
  score: 40,
  phase: 0,
  moveCount: 3,
  canUndo: true,
  isStalemate: false,
};

describe('formatSpideretteState', () => {
  it('renders the header, stock, and completed count', () => {
    const out = formatSpideretteState(baseState);
    expect(out).toContain('Spiderette');
    expect(out).toContain('stock: 20 | completed: 1/4');
    expect(out).toContain('moves: 3 | score: 40');
  });

  it('renders columns with face-up indices, hidden cards, and empties', () => {
    const out = formatSpideretteState(baseState);
    expect(out).toContain('col0: [0]'); // face-up king
    expect(out).toContain('col1: [empty]');
    expect(out).toContain('[?]'); // face-down card hidden
  });

  it('renders a hint line when present', () => {
    const out = formatSpideretteState({ ...baseState, hint: { fromCol: 0, cardIndex: 0, toCol: 2 } });
    expect(out).toContain('HINT: col0[0] → col2');
  });

  it('renders a stalemate notice and a win banner', () => {
    expect(formatSpideretteState({ ...baseState, isStalemate: true })).toContain('Stalemate');
    expect(formatSpideretteState({ ...baseState, phase: 1 })).toContain('Congratulations');
  });
});
