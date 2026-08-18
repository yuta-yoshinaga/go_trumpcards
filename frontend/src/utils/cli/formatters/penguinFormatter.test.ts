import { describe, expect, it } from 'vitest';
import type { PenguinResponse } from '../../../types/card';
import { formatPenguinState } from './penguinFormatter';

function makeState(overrides?: Partial<PenguinResponse>): PenguinResponse {
  return {
    tableau: Array.from({ length: 7 }, () => []),
    freeCells: [null, null, null, null, null, null, null],
    foundation: [[], [], [], []],
    baseRank: 5,
    maxMovableCards: 8,
    maxMovableCardsToEmptyColumn: 4,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatPenguinState', () => {
  it('renders header and empty board', () => {
    const result = formatPenguinState(makeState());
    expect(result).toContain('Penguin');
    expect(result).toContain('baseRank: 5');
    expect(result).toContain('foundation:');
    expect(result).toContain('col0: [empty]');
  });

  it('renders baseRank with face card label', () => {
    const result = formatPenguinState(makeState({ baseRank: 1 }));
    expect(result).toContain('baseRank: A');
  });

  it('renders baseRank King label', () => {
    const result = formatPenguinState(makeState({ baseRank: 13 }));
    expect(result).toContain('baseRank: K');
  });

  it('renders baseRank Jack label', () => {
    const result = formatPenguinState(makeState({ baseRank: 11 }));
    expect(result).toContain('baseRank: J');
  });

  it('renders baseRank Queen label', () => {
    const result = formatPenguinState(makeState({ baseRank: 12 }));
    expect(result).toContain('baseRank: Q');
  });

  it('renders cards in tableau', () => {
    const result = formatPenguinState(
      makeState({
        tableau: [[{ design: 'SPADE', value: 1 }], ...Array.from({ length: 6 }, () => [])],
      }),
    );
    expect(result).toContain('[0]');
  });

  it('renders free cells with cards', () => {
    const result = formatPenguinState(
      makeState({
        freeCells: [{ design: 'SPADE', value: 5 }, null, null, null, null, null, null],
      }),
    );
    expect(result).toContain('cells:');
    expect(result).toContain('[c0]');
  });

  it('renders foundation top card', () => {
    const result = formatPenguinState(
      makeState({
        foundation: [[{ design: 'HEART', value: 5 }], [], [], []],
      }),
    );
    expect(result).toContain('foundation:');
  });

  it('shows hint when present', () => {
    const result = formatPenguinState(
      makeState({
        hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: -1 },
        messageCode: 'penguin.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
  });

  it('shows stalemate message', () => {
    const result = formatPenguinState(makeState({ isStalemate: true }));
    expect(result).toContain('Stalemate');
  });

  it('shows congrats on win phase', () => {
    const result = formatPenguinState(makeState({ phase: 1 }));
    expect(result).toContain('Congratulations');
  });

  it('shows move count', () => {
    const result = formatPenguinState(makeState({ moveCount: 42 }));
    expect(result).toContain('moves: 42');
  });

  it('shows message when present', () => {
    const result = formatPenguinState(makeState({ message: 'test message' }));
    expect(result).toContain('test message');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 1 };
    expect(formatPenguinState(makeState({ hint, messageCode: 'penguin.hintAvailable' }))).toContain('HINT:');
    expect(formatPenguinState(makeState({ hint, messageCode: 'penguin.playing' }))).not.toContain('HINT:');
  });
});
