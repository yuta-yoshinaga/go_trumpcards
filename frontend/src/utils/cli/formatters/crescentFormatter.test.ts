import { describe, expect, it } from 'vitest';
import type { CrescentResponse } from '../../../types/card';
import { formatCrescentState } from './crescentFormatter';

const baseState: CrescentResponse = {
  tableau: [
    [{ card: { design: 'SPADE', value: 13 }, faceUp: true }],
    [],
    [
      { card: { design: 'HEART', value: 5 }, faceUp: true },
      { card: { design: 'HEART', value: 4 }, faceUp: true },
    ],
  ],
  foundation: [[{ design: 'SPADE', value: 1 }], [], [], [], [{ design: 'CLOVER', value: 13 }], [], [], []],
  redealsRemaining: 2,
  phase: 0,
  moveCount: 3,
  canUndo: true,
  isStalemate: false,
  message: '',
};

describe('formatCrescentState', () => {
  it('renders ascending and descending foundations separately', () => {
    const out = formatCrescentState(baseState);
    expect(out).toContain('Crescent');
    // Ascending pile 0 holds the Ace of spades; descending pile 4 holds the King.
    expect(out).toMatch(/asc:.*A/);
    expect(out).toMatch(/desc:.*K/);
  });

  it('renders tableau columns, empties, and the redeal count', () => {
    const out = formatCrescentState(baseState);
    expect(out).toContain('t0: [0]');
    expect(out).toContain('t1: [empty]');
    expect(out).toContain('redeals: 2');
  });

  it('renders a positional hint', () => {
    const out = formatCrescentState({
      ...baseState,
      hint: { fromCol: 2, toZone: 'foundation', toCol: 4, redeal: false },
    });
    expect(out).toContain('HINT: t2 → foundation4');
  });

  it('renders a redeal hint', () => {
    const out = formatCrescentState({ ...baseState, hint: { fromCol: -1, toZone: '', toCol: -1, redeal: true } });
    expect(out).toContain('HINT: redeal');
  });

  it('renders stalemate and win banners', () => {
    expect(formatCrescentState({ ...baseState, isStalemate: true })).toContain('Stalemate');
    expect(formatCrescentState({ ...baseState, phase: 1 })).toContain('Congratulations');
  });
});
