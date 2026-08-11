import { describe, expect, it } from 'vitest';
import type { RoyalCotillionResponse } from '../../../types/card';
import { formatRoyalCotillionState } from './royalcotillionFormatter';

function makeState(overrides?: Partial<RoyalCotillionResponse>): RoyalCotillionResponse {
  return {
    tableau: Array.from({ length: 16 }, () => null),
    reserve: Array.from({ length: 4 }, () => []),
    foundationOdd: [true, true, true, true, false, false, false, false],
    foundation: Array.from({ length: 8 }, () => []),
    stockCount: 76,
    waste: [],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatRoyalCotillionState', () => {
  it('renders header, piles and an empty board', () => {
    const result = formatRoyalCotillionState(makeState());
    expect(result).toContain('Royal Cotillion');
    expect(result).toContain('foundations:');
    expect(result).toContain('stock: 76');
    // Sixteen slots, four to a row, plus the four reserve piles.
    expect(result).toContain('[0]');
    expect(result).toContain('[15]');
    expect(result).toContain('r0: [empty] (never refilled)');
    expect(result).toContain('r3: [empty] (never refilled)');
  });

  // An empty pile behaves differently here, so it says where a card may come from.
  it('says an empty pile takes stock or waste only', () => {
    expect(formatRoyalCotillionState(makeState())).toContain('never refilled');
  });

  it('renders a card in a slot', () => {
    const tableau: (typeof makeState extends never ? never : { design: string; value: number } | null)[] = Array.from(
      { length: 16 },
      () => null,
    );
    tableau[0] = { design: 'SPADE', value: 9 };
    const result = formatRoyalCotillionState(makeState({ tableau: tableau as never }));
    expect(result).toContain('[0]');
    // An empty slot renders as a gap rather than vanishing.
    expect(result).toContain('[1]');
  });

  // Both series are marked, or a pile's next rank cannot be worked out.
  it('marks the Ace-start and deuce-start series', () => {
    const result = formatRoyalCotillionState(makeState());
    expect(result).toContain('A:');
    expect(result).toContain('2:');
  });

  it('renders the waste top', () => {
    expect(formatRoyalCotillionState(makeState({ waste: [{ design: 'DIAMOND', value: 4 }] }))).toContain('waste:');
  });

  it('shows a tableau hint with its pile', () => {
    const result = formatRoyalCotillionState(
      makeState({
        hint: { fromZone: 'tableau', fromIdx: 3, toZone: 'foundation', toIdx: 2 },
        messageCode: 'royalcotillion.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('t3');
    expect(result).toContain('foundation2');
  });

  // **頼んでいないヒントは CLI に出さない。**#4483 以降 Output() もヒントを載せる
  // ので、state.hint だけを見ると毎手 HINT が印字される。
  it('does not print a passive hint carried on an ordinary response', () => {
    const result = formatRoyalCotillionState(
      makeState({
        hint: { fromZone: 'tableau', fromIdx: 3, toZone: 'foundation', toIdx: 2 },
        messageCode: 'royalcotillion.playing',
      }),
    );
    expect(result).not.toContain('HINT:');
  });

  it('renders a draw hint without indices', () => {
    const result = formatRoyalCotillionState(
      makeState({
        hint: { fromZone: 'stock', fromIdx: -1, toZone: 'waste', toIdx: -1 },
        messageCode: 'royalcotillion.hintAvailable',
      }),
    );
    expect(result).toContain('stock → waste');
  });

  it('shows stalemate message', () => {
    expect(formatRoyalCotillionState(makeState({ isStalemate: true }))).toContain('Stalemate');
  });

  it('shows the server message', () => {
    expect(formatRoyalCotillionState(makeState({ message: 'nope' }))).toContain('nope');
  });

  it('shows congrats on win phase', () => {
    expect(formatRoyalCotillionState(makeState({ phase: 1 }))).toContain('Congratulations');
  });
});
