import { describe, expect, it } from 'vitest';
import type { CrazyQuiltResponse } from '../../../types/card';
import { formatCrazyQuiltState } from './crazyquiltFormatter';

function makeState(overrides?: Partial<CrazyQuiltResponse>): CrazyQuiltResponse {
  return {
    quilt: Array.from({ length: 64 }, () => null),
    available: Array.from({ length: 64 }, () => false),
    foundationAscending: [true, true, true, true, false, false, false, false],
    redealsLeft: 1,
    foundation: Array.from({ length: 8 }, () => []),
    stockCount: 32,
    waste: [],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatCrazyQuiltState', () => {
  it('renders header, piles and an empty board', () => {
    const result = formatCrazyQuiltState(makeState());
    expect(result).toContain('Crazy Quilt');
    expect(result).toContain('foundations:');
    expect(result).toContain('stock: 32');
  });

  // An empty pile behaves differently here, so it says where a card may come from.
  it('shows the redeal count and both foundation directions', () => {
    const out = formatCrazyQuiltState(makeState());
    expect(out).toContain('redeals: 1');
    expect(out).toContain('↑');
    expect(out).toContain('↓');
  });

  it('marks the takeable cards and draws the empties', () => {
    const quilt = Array.from({ length: 64 }, () => null as { design: string; value: number } | null);
    quilt[0] = { design: 'SPADE', value: 9 };
    quilt[1] = { design: 'HEART', value: 3 };
    const available = Array.from({ length: 64 }, (_, i) => i === 0);
    const result = formatCrazyQuiltState(makeState({ quilt: quilt as never, available }));
    expect(result).toContain('*');
    expect(result).toContain('* = takeable');
    // An emptied cell is drawn rather than skipped, so the grid stays aligned.
    expect(result).toContain('.');
  });

  it('renders the waste top', () => {
    expect(formatCrazyQuiltState(makeState({ waste: [{ design: 'DIAMOND', value: 4 }] }))).toContain('waste:');
  });

  it('shows a tableau hint with its pile', () => {
    const result = formatCrazyQuiltState(
      makeState({
        hint: { fromZone: 'tableau', fromIdx: 3, toZone: 'foundation', toIdx: 2 },
        messageCode: 'crazyquilt.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('t3');
    expect(result).toContain('foundation2');
  });

  // **頼んでいないヒントは CLI に出さない。**#4483 以降 Output() もヒントを載せる
  // ので、state.hint だけを見ると毎手 HINT が印字される。
  it('does not print a passive hint carried on an ordinary response', () => {
    const result = formatCrazyQuiltState(
      makeState({
        hint: { fromZone: 'tableau', fromIdx: 3, toZone: 'foundation', toIdx: 2 },
        messageCode: 'crazyquilt.playing',
      }),
    );
    expect(result).not.toContain('HINT:');
  });

  it('renders a draw hint without indices', () => {
    const result = formatCrazyQuiltState(
      makeState({
        hint: { fromZone: 'stock', fromIdx: -1, toZone: 'waste', toIdx: -1 },
        messageCode: 'crazyquilt.hintAvailable',
      }),
    );
    expect(result).toContain('stock → waste');
  });

  it('shows stalemate message', () => {
    expect(formatCrazyQuiltState(makeState({ isStalemate: true }))).toContain('Stalemate');
  });

  it('shows the server message', () => {
    expect(formatCrazyQuiltState(makeState({ message: 'nope' }))).toContain('nope');
  });

  it('shows congrats on win phase', () => {
    expect(formatCrazyQuiltState(makeState({ phase: 1 }))).toContain('Congratulations');
  });
});
