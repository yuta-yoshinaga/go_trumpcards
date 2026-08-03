import { describe, expect, it } from 'vitest';
import type { CongressResponse } from '../../../types/card';
import { formatCongressState } from './congressFormatter';

function makeState(overrides?: Partial<CongressResponse>): CongressResponse {
  return {
    tableau: Array.from({ length: 8 }, () => []),
    foundation: Array.from({ length: 8 }, () => []),
    stockCount: 96,
    waste: [],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatCongressState', () => {
  it('renders header, piles and an empty board', () => {
    const result = formatCongressState(makeState());
    expect(result).toContain('Congress');
    expect(result).toContain('foundations:');
    expect(result).toContain('stock: 96');
    expect(result).toContain('t0: [empty]');
    expect(result).toContain('t7: [empty]');
  });

  // An empty pile behaves differently here, so it says where a card may come from.
  it('says an empty pile takes stock or waste only', () => {
    expect(formatCongressState(makeState())).toContain('stock or waste only');
  });

  it('renders cards in a pile', () => {
    const result = formatCongressState(
      makeState({ tableau: [[{ design: 'SPADE', value: 9 }], ...Array.from({ length: 7 }, () => [])] }),
    );
    expect(result).toContain('[0]');
  });

  it('renders the waste top', () => {
    expect(formatCongressState(makeState({ waste: [{ design: 'DIAMOND', value: 4 }] }))).toContain('waste:');
  });

  it('shows a tableau hint with its pile', () => {
    const result = formatCongressState(
      makeState({
        hint: { fromZone: 'tableau', fromIdx: 3, toZone: 'foundation', toIdx: 2 },
        messageCode: 'congress.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('t3');
    expect(result).toContain('foundation2');
  });

  // **頼んでいないヒントは CLI に出さない。**#4483 以降 Output() もヒントを載せる
  // ので、state.hint だけを見ると毎手 HINT が印字される。
  it('does not print a passive hint carried on an ordinary response', () => {
    const result = formatCongressState(
      makeState({
        hint: { fromZone: 'tableau', fromIdx: 3, toZone: 'foundation', toIdx: 2 },
        messageCode: 'congress.playing',
      }),
    );
    expect(result).not.toContain('HINT:');
  });

  it('renders a draw hint without indices', () => {
    const result = formatCongressState(
      makeState({
        hint: { fromZone: 'stock', fromIdx: -1, toZone: 'waste', toIdx: -1 },
        messageCode: 'congress.hintAvailable',
      }),
    );
    expect(result).toContain('stock → waste');
  });

  it('shows stalemate message', () => {
    expect(formatCongressState(makeState({ isStalemate: true }))).toContain('Stalemate');
  });

  it('shows the server message', () => {
    expect(formatCongressState(makeState({ message: 'nope' }))).toContain('nope');
  });

  it('shows congrats on win phase', () => {
    expect(formatCongressState(makeState({ phase: 1 }))).toContain('Congratulations');
  });
});
