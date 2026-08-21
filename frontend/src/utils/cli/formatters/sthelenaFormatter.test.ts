import { describe, expect, it } from 'vitest';
import type { StHelenaResponse } from '../../../types/card';
import { formatStHelenaState } from './sthelenaFormatter';

const baseState: StHelenaResponse = {
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
  restrictionsActive: true,
  phase: 0,
  moveCount: 3,
  canUndo: true,
  isStalemate: false,
  message: '',
};

describe('formatStHelenaState', () => {
  it('renders ascending and descending foundations separately', () => {
    const out = formatStHelenaState(baseState);
    expect(out).toContain('StHelena');
    // Ascending pile 0 holds the Ace of spades; descending pile 4 holds the King.
    expect(out).toMatch(/asc:.*A/);
    expect(out).toMatch(/desc:.*K/);
  });

  it('renders tableau columns, empties, and the redeal count', () => {
    const out = formatStHelenaState(baseState);
    expect(out).toContain('t0: [0]');
    expect(out).toContain('t1: [empty]');
    expect(out).toContain('redeals: 2');
  });

  it('renders a positional hint', () => {
    const out = formatStHelenaState({
      ...baseState,
      hint: { fromCol: 2, toZone: 'foundation', toCol: 4, redeal: false },
      messageCode: 'sthelena.hintAvailable',
    });
    expect(out).toContain('HINT: t2 → foundation4');
  });

  it('renders a redeal hint', () => {
    const out = formatStHelenaState({
      ...baseState,
      hint: { fromCol: -1, toZone: '', toCol: -1, redeal: true },
      messageCode: 'sthelena.hintAvailable',
    });
    expect(out).toContain('HINT: redeal');
  });

  it('renders stalemate and win banners', () => {
    expect(formatStHelenaState({ ...baseState, isStalemate: true })).toContain('Stalemate');
    expect(formatStHelenaState({ ...baseState, phase: 1 })).toContain('Congratulations');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { fromCol: 2, toZone: 'foundation', toCol: 4, redeal: false };
    expect(formatStHelenaState({ ...baseState, hint, messageCode: 'sthelena.hintAvailable' })).toContain('HINT:');
    expect(formatStHelenaState({ ...baseState, hint, messageCode: 'sthelena.playing' })).not.toContain('HINT:');
  });
});
