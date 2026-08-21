import { describe, expect, it } from 'vitest';
import type { WhiteheadResponse } from '../../../types/card';
import { formatWhiteheadState } from './whiteheadFormatter';

function makeState(overrides?: Partial<WhiteheadResponse>): WhiteheadResponse {
  return {
    tableau: [
      [{ card: { design: 'SPADE', value: 13 }, faceUp: true }],
      [
        { card: { design: 'HEART', value: 5 }, faceUp: false },
        { card: { design: 'CLOVER', value: 3 }, faceUp: true },
      ],
      [],
      [],
      [],
      [],
      [],
    ],
    stockCount: 24,
    waste: [{ design: 'DIAMOND', value: 7 }],
    foundation: [[], [], [], []],
    phase: 1,
    moveCount: 5,
    drawCount: 3,
    canUndo: true,
    isStalemate: false,
    score: 10,
    scoringMode: 0,
    message: '',
    ...overrides,
  };
}

describe('formatWhiteheadState', () => {
  it('formats basic state', () => {
    const output = formatWhiteheadState(makeState());
    expect(output).toContain('Whitehead');
    expect(output).toContain('Stock: 24');
    expect(output).toContain('moves: 5');
    expect(output).toContain('col0:');
  });

  it('shows foundation cards', () => {
    const output = formatWhiteheadState(makeState({ foundation: [[{ design: 'SPADE', value: 1 }], [], [], []] }));
    expect(output).toContain('Foundation:');
  });

  it('shows face-down cards as [?]', () => {
    const output = formatWhiteheadState(makeState());
    expect(output).toContain('[?]');
  });

  it('shows stalemate', () => {
    const output = formatWhiteheadState(makeState({ isStalemate: true }));
    expect(output).toContain('Stalemate');
  });

  it('shows win message', () => {
    const output = formatWhiteheadState(makeState({ phase: 2 }));
    expect(output).toContain('Congratulations');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { fromZone: 'tableau', fromCol: 1, cardIndex: 0, toZone: 'foundation', toCol: 2 };
    expect(formatWhiteheadState(makeState({ hint, messageCode: 'whitehead.hintAvailable' }))).toContain('HINT:');
    expect(formatWhiteheadState(makeState({ hint, messageCode: 'whitehead.playing' }))).not.toContain('HINT:');
  });
});
