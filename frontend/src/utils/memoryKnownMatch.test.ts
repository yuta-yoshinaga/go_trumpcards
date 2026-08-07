import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import type { MemoryBoardCard } from '../types/games/memory';
import { memoryKnownMatch } from './memoryKnownMatch';

const card = (value: number): Card => ({ design: 'SPADE', value });
const cell = (over: Partial<MemoryBoardCard> = {}): MemoryBoardCard => ({
  card: null,
  faceUp: false,
  taken: false,
  ...over,
});

describe('memoryKnownMatch', () => {
  const seen = new Map<number, Card>([
    [2, card(7)],
    [3, { design: 'HEART', value: 9 }],
  ]);

  it('finds the remembered twin of the single face-up card', () => {
    const board = [cell({ card: card(7), faceUp: true }), cell(), cell(), cell()];
    expect(memoryKnownMatch(board, seen)).toBe(2);
  });

  it('answers null when nothing remembered matches', () => {
    const board = [cell({ card: card(4), faceUp: true }), cell(), cell(), cell()];
    expect(memoryKnownMatch(board, seen)).toBeNull();
  });

  // **表向きが1枚のときだけ答える。**0枚や2枚のときに答えると、実際には
  // 打てない手を指す表示になる。
  it.each([
    ['nothing is face up', [cell(), cell(), cell(), cell()]],
    [
      'two cards are already face up',
      [cell({ card: card(7), faceUp: true }), cell({ card: card(5), faceUp: true }), cell(), cell()],
    ],
  ])('answers null when %s', (_name, board) => {
    expect(memoryKnownMatch(board, seen)).toBeNull();
  });

  it('skips a position that has already been taken off the board', () => {
    const board = [cell({ card: card(7), faceUp: true }), cell(), cell({ taken: true }), cell()];
    expect(memoryKnownMatch(board, seen)).toBeNull();
  });

  it('ignores an empty memory', () => {
    const board = [cell({ card: card(7), faceUp: true }), cell(), cell(), cell()];
    expect(memoryKnownMatch(board, new Map())).toBeNull();
  });
});
