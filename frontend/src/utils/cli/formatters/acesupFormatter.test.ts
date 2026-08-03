import { describe, expect, it } from 'vitest';
import type { AcesUpCard, AcesUpResponse, Card, CardDesign } from '../../../types/card';
import { formatAcesUpState } from './acesupFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const auCard = (c: Card, top = false): AcesUpCard => ({ card: c, top, removable: false, movable: false });

const base: AcesUpResponse = {
  columns: [[auCard(card('SPADE', 5), true)], [], [auCard(card('HEART', 2)), auCard(card('HEART', 9), true)], []],
  stockCount: 44,
  discardCount: 4,
  phase: 0,
  moveCount: 3,
  canUndo: true,
  isStalemate: false,
  message: '',
};

describe('formatAcesUpState', () => {
  it('renders header, stock and discard counts', () => {
    const out = formatAcesUpState(base);
    expect(out).toContain('Aces Up');
    expect(out).toContain('stock: 44');
    expect(out).toContain('discard: 4');
    expect(out).toContain('moves: 3');
  });

  it('marks empty columns and top cards', () => {
    const out = formatAcesUpState(base);
    expect(out).toContain('(empty)');
    // top card wrapped in brackets
    expect(out).toMatch(/\[.*\]/);
  });

  it('renders stalemate and hint lines', () => {
    const out = formatAcesUpState({
      ...base,
      isStalemate: true,
      hint: { type: 'move', col: 2 },
      messageCode: 'acesup.hintAvailable',
    });
    expect(out).toContain('Stalemate');
    expect(out).toContain('HINT: move col 2');
  });

  it('renders draw hint without a column', () => {
    const out = formatAcesUpState({ ...base, hint: { type: 'draw', col: -1 }, messageCode: 'acesup.hintAvailable' });
    expect(out).toContain('HINT: deal');
  });

  it('renders win message and arbitrary message', () => {
    const out = formatAcesUpState({ ...base, phase: 1, message: 'done' });
    expect(out).toContain('Congratulations');
    expect(out).toContain('done');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const state = { ...base, hint: { type: 'move', col: 2 } } satisfies AcesUpResponse;
    expect(formatAcesUpState({ ...state, messageCode: 'acesup.hintAvailable' })).toContain('HINT:');
    expect(formatAcesUpState({ ...state, messageCode: 'acesup.playing' })).not.toContain('HINT:');
  });
});
