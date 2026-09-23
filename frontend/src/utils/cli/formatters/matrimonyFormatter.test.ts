import { describe, expect, it } from 'vitest';
import type { MatrimonyResponse } from '../../../types/games/matrimony';
import { formatMatrimonyState } from './matrimonyFormatter';

function makeState(overrides?: Partial<MatrimonyResponse>): MatrimonyResponse {
  return {
    tableau: Array.from({ length: 16 }, () => null),
    foundation: Array.from({ length: 4 }, () => []),
    stockCount: 88,
    redealCount: 0,
    waste: [],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatMatrimonyState', () => {
  it('renders header, piles and an empty board', () => {
    const result = formatMatrimonyState(makeState());
    expect(result).toContain('Matrimony');
    expect(result).toContain('foundations:');
    expect(result).toContain('stock: 88');
    // Sixteen slots, four to a row.
    expect(result).toContain('[0]');
    expect(result).toContain('[15]');
  });

  it('renders a card in a slot', () => {
    const tableau: (typeof makeState extends never ? never : { design: string; value: number } | null)[] = Array.from(
      { length: 16 },
      () => null,
    );
    tableau[0] = { design: 'SPADE', value: 9 };
    const result = formatMatrimonyState(makeState({ tableau: tableau as never }));
    expect(result).toContain('[0]');
    // An empty slot renders as a gap rather than vanishing.
    expect(result).toContain('[1]');
  });

  it('renders the waste top', () => {
    expect(formatMatrimonyState(makeState({ waste: [{ design: 'DIAMOND', value: 4 }] }))).toContain('waste:');
  });

  it('shows a tableau hint with its pile', () => {
    const result = formatMatrimonyState(
      makeState({
        hint: { fromZone: 'tableau', fromIdx: 3, toZone: 'foundation', toIdx: 2 },
        messageCode: 'matrimony.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('t3');
    expect(result).toContain('foundation2');
  });

  // **頼んでいないヒントは CLI に出さない。**#4483 以降 Output() もヒントを載せる
  // ので、state.hint だけを見ると毎手 HINT が印字される。
  it('does not print a passive hint carried on an ordinary response', () => {
    const result = formatMatrimonyState(
      makeState({
        hint: { fromZone: 'tableau', fromIdx: 3, toZone: 'foundation', toIdx: 2 },
        messageCode: 'matrimony.playing',
      }),
    );
    expect(result).not.toContain('HINT:');
  });

  it('renders a draw hint without indices', () => {
    const result = formatMatrimonyState(
      makeState({
        hint: { fromZone: 'stock', fromIdx: -1, toZone: 'waste', toIdx: -1 },
        messageCode: 'matrimony.hintAvailable',
      }),
    );
    expect(result).toContain('stock → waste');
  });

  it('shows stalemate message', () => {
    expect(formatMatrimonyState(makeState({ isStalemate: true }))).toContain('Stalemate');
  });

  it('shows the server message', () => {
    expect(formatMatrimonyState(makeState({ message: 'nope' }))).toContain('nope');
  });

  it('shows congrats on win phase', () => {
    expect(formatMatrimonyState(makeState({ phase: 1 }))).toContain('Congratulations');
  });
});
