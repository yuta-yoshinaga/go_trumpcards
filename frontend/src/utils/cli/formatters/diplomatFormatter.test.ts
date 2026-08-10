import { describe, expect, it } from 'vitest';
import type { DiplomatResponse } from '../../../types/card';
import { formatDiplomatState } from './diplomatFormatter';

function makeState(overrides?: Partial<DiplomatResponse>): DiplomatResponse {
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

describe('formatDiplomatState', () => {
  it('renders header, piles and an empty board', () => {
    const result = formatDiplomatState(makeState());
    expect(result).toContain('Diplomat');
    expect(result).toContain('foundations:');
    expect(result).toContain('stock: 96');
    expect(result).toContain('t0: [empty]');
    expect(result).toContain('t7: [empty]');
  });

  // An empty pile behaves differently here, so it says where a card may come from.
  it('says an empty pile takes stock or waste only', () => {
    expect(formatDiplomatState(makeState())).toContain('stock or waste only');
  });

  it('renders cards in a pile', () => {
    const result = formatDiplomatState(
      makeState({ tableau: [[{ design: 'SPADE', value: 9 }], ...Array.from({ length: 7 }, () => [])] }),
    );
    expect(result).toContain('[0]');
  });

  it('renders the waste top', () => {
    expect(formatDiplomatState(makeState({ waste: [{ design: 'DIAMOND', value: 4 }] }))).toContain('waste:');
  });

  it('shows a tableau hint with its pile', () => {
    const result = formatDiplomatState(
      makeState({
        hint: { fromZone: 'tableau', fromIdx: 3, toZone: 'foundation', toIdx: 2 },
        messageCode: 'diplomat.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('t3');
    expect(result).toContain('foundation2');
  });

  // **頼んでいないヒントは CLI に出さない。**#4483 以降 Output() もヒントを載せる
  // ので、state.hint だけを見ると毎手 HINT が印字される。
  it('does not print a passive hint carried on an ordinary response', () => {
    const result = formatDiplomatState(
      makeState({
        hint: { fromZone: 'tableau', fromIdx: 3, toZone: 'foundation', toIdx: 2 },
        messageCode: 'diplomat.playing',
      }),
    );
    expect(result).not.toContain('HINT:');
  });

  it('renders a draw hint without indices', () => {
    const result = formatDiplomatState(
      makeState({
        hint: { fromZone: 'stock', fromIdx: -1, toZone: 'waste', toIdx: -1 },
        messageCode: 'diplomat.hintAvailable',
      }),
    );
    expect(result).toContain('stock → waste');
  });

  it('shows stalemate message', () => {
    expect(formatDiplomatState(makeState({ isStalemate: true }))).toContain('Stalemate');
  });

  it('shows the server message', () => {
    expect(formatDiplomatState(makeState({ message: 'nope' }))).toContain('nope');
  });

  it('shows congrats on win phase', () => {
    expect(formatDiplomatState(makeState({ phase: 1 }))).toContain('Congratulations');
  });
});
