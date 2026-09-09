import { describe, expect, it } from 'vitest';
import type { WillOTheWispResponse } from '../../../types/games/willothewisp';
import { formatWillOTheWispState } from './willothewispFormatter';

const baseState: WillOTheWispResponse = {
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
  scoring: { start: 500, movePenalty: 1, suitBonus: 100 },
  phase: 0,
  moveCount: 3,
  canUndo: true,
  isStalemate: false,
  message: '',
};

describe('formatWillOTheWispState', () => {
  it('renders the header, stock, and completed count', () => {
    const out = formatWillOTheWispState(baseState);
    expect(out).toContain("Will o' the Wisp");
    expect(out).toContain('stock: 20 | completed: 1/4');
    expect(out).toContain('moves: 3 | score: 40');
  });

  it('renders columns with face-up indices, hidden cards, and empties', () => {
    const out = formatWillOTheWispState(baseState);
    expect(out).toContain('col0: [0]'); // face-up king
    expect(out).toContain('col1: [empty]');
    expect(out).toContain('[?]'); // face-down card hidden
  });

  it('renders a hint line when present', () => {
    const out = formatWillOTheWispState({
      ...baseState,
      hint: { fromCol: 0, cardIndex: 0, toCol: 2 },
      messageCode: 'willothewisp.hintAvailable',
    });
    expect(out).toContain('HINT: col0[0] → col2');
  });

  it('renders a stalemate notice and a win banner', () => {
    expect(formatWillOTheWispState({ ...baseState, isStalemate: true })).toContain('Stalemate');
    expect(formatWillOTheWispState({ ...baseState, phase: 1 })).toContain('Congratulations');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { fromCol: 1, cardIndex: 0, toCol: 3 };
    expect(formatWillOTheWispState({ ...baseState, hint, messageCode: 'willothewisp.hintAvailable' })).toContain(
      'HINT:',
    );
    expect(formatWillOTheWispState({ ...baseState, hint, messageCode: 'willothewisp.playing' })).not.toContain('HINT:');
  });
});
