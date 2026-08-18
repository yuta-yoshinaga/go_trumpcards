import { describe, expect, it } from 'vitest';
import type { BristolResponse } from '../../../types/card';
import { formatBristolState } from './bristolFormatter';

function makeState(overrides: Partial<BristolResponse> = {}): BristolResponse {
  return {
    tableau: [
      [{ design: 'SPADE', value: 8 }],
      [{ design: 'HEART', value: 9 }],
      [{ design: 'CLOVER', value: 4 }],
      [{ design: 'DIAMOND', value: 10 }],
      [{ design: 'SPADE', value: 3 }],
      [{ design: 'HEART', value: 6 }],
      [{ design: 'CLOVER', value: 2 }],
      [],
    ],
    fan: [[{ design: 'HEART', value: 4 }], [], []],
    stockCount: 28,
    foundation: [[], [], [], []],
    legalTargets: {},
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    undoToEscape: 0,
    message: '',
    ...overrides,
  };
}

describe('formatBristolState', () => {
  it('includes Bristol header', () => {
    const output = formatBristolState(makeState());
    expect(output).toContain('Bristol');
  });

  it('formats foundation piles with cards', () => {
    const output = formatBristolState(makeState({ foundation: [[{ design: 'SPADE', value: 1 }], [], [], []] }));
    expect(output).toContain('foundation0:');
    expect(output).toContain('(1)');
  });

  it('formats empty foundation piles with placeholder', () => {
    const output = formatBristolState(makeState());
    expect(output).toContain('[  ]');
  });

  it('formats non-empty tableau columns', () => {
    const output = formatBristolState(makeState());
    expect(output).toContain('tableau0:');
  });

  it('formats empty tableau columns', () => {
    const output = formatBristolState(makeState());
    expect(output).toContain('[empty]');
  });

  it('formats fan pile with cards', () => {
    const output = formatBristolState(makeState());
    expect(output).toContain('fan0:');
    expect(output).toContain('(1)');
  });

  it('formats empty fan pile', () => {
    const output = formatBristolState(makeState());
    expect(output).toMatch(/fan[12]: \[empty\]/);
  });

  it('shows stock count and move count', () => {
    const output = formatBristolState(makeState({ stockCount: 25, moveCount: 3 }));
    expect(output).toContain('stock: 25');
    expect(output).toContain('moves: 3');
  });

  it('shows undo:yes when canUndo is true', () => {
    const output = formatBristolState(makeState({ canUndo: true }));
    expect(output).toContain('undo:yes');
  });

  it('shows undo:no when canUndo is false', () => {
    const output = formatBristolState(makeState({ canUndo: false }));
    expect(output).toContain('undo:no');
  });

  it('shows tableau-to-foundation hint', () => {
    const output = formatBristolState(
      makeState({
        hint: { fromZone: 'tableau', fromCol: 2, toZone: 'foundation', toCol: 0 },
        messageCode: 'bristol.hintAvailable',
      }),
    );
    expect(output).toContain('HINT: tableau2 → foundation0');
  });

  // **頼んでいないヒントは CLI に出さない。**#4483 以降 Output() もヒントを載せる
  // ので、state.hint だけを見ると毎手 HINT が印字される。
  it('does not print a passive hint carried on an ordinary response', () => {
    const result = formatBristolState(
      makeState({
        hint: { fromZone: 'tableau', fromCol: 2, toZone: 'foundation', toCol: 0 },
        messageCode: 'bristol.playing',
      }),
    );
    expect(result).not.toContain('HINT');
  });

  it('shows fan-to-tableau hint', () => {
    const output = formatBristolState(
      makeState({
        hint: { fromZone: 'fan', fromCol: 1, toZone: 'tableau', toCol: 3 },
        messageCode: 'bristol.hintAvailable',
      }),
    );
    expect(output).toContain('HINT: fan1 → tableau3');
  });

  it('omits hint line when no hint is present', () => {
    const output = formatBristolState(makeState());
    expect(output).not.toContain('HINT:');
  });

  it('shows message when set', () => {
    const output = formatBristolState(makeState({ message: 'Some message' }));
    expect(output).toContain('Some message');
  });

  it('shows win message for phase 1', () => {
    const output = formatBristolState(makeState({ phase: 1 }));
    expect(output).toContain('Congratulations! You win!');
  });

  it('does not show win message for phase 0', () => {
    const output = formatBristolState(makeState());
    expect(output).not.toContain('Congratulations!');
  });
});
