import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, NarcoticCard, NarcoticResponse } from '../../../types/card';
import { formatNarcoticState } from './narcoticFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const auCard = (c: Card, top = false): NarcoticCard => ({ card: c, top, removable: false, movable: false });

const base: NarcoticResponse = {
  columns: [[auCard(card('SPADE', 5), true)], [], [auCard(card('HEART', 2)), auCard(card('HEART', 9), true)], []],
  stockCount: 44,
  discardCount: 4,
  redealCount: 0,
  phase: 0,
  moveCount: 3,
  canUndo: true,
  isStalemate: false,
  message: '',
};

describe('formatNarcoticState', () => {
  it('renders header, stock and discard counts', () => {
    const out = formatNarcoticState(base);
    expect(out).toContain('Narcotic');
    expect(out).toContain('stock: 44');
    expect(out).toContain('discard: 4');
    expect(out).toContain('moves: 3');
  });

  it('marks empty columns and top cards', () => {
    const out = formatNarcoticState(base);
    expect(out).toContain('(empty)');
    // top card wrapped in brackets
    expect(out).toMatch(/\[.*\]/);
  });

  it('renders stalemate and hint lines', () => {
    const out = formatNarcoticState({
      ...base,
      isStalemate: true,
      hint: { type: 'move', col: 2 },
      messageCode: 'narcotic.hintAvailable',
    });
    expect(out).toContain('Stalemate');
    expect(out).toContain('HINT: move col 2');
  });

  it('renders draw hint without a column', () => {
    const out = formatNarcoticState({
      ...base,
      hint: { type: 'draw', col: -1 },
      messageCode: 'narcotic.hintAvailable',
    });
    expect(out).toContain('HINT: deal');
  });

  it('renders win message and arbitrary message', () => {
    const out = formatNarcoticState({ ...base, phase: 1, message: 'done' });
    expect(out).toContain('Congratulations');
    expect(out).toContain('done');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const state = { ...base, hint: { type: 'move', col: 2 } } satisfies NarcoticResponse;
    expect(formatNarcoticState({ ...state, messageCode: 'narcotic.hintAvailable' })).toContain('HINT:');
    expect(formatNarcoticState({ ...state, messageCode: 'narcotic.playing' })).not.toContain('HINT:');
  });
});
