import { describe, expect, it } from 'vitest';
import type { BraidResponse, Card, CardDesign } from '../../../types/card';
import { formatBraidState } from './braidFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeState(overrides?: Partial<BraidResponse>): BraidResponse {
  return {
    braid: [],
    fields: Array.from({ length: 4 }, () => null),
    helpers: Array.from({ length: 8 }, () => null),
    foundation: Array.from({ length: 8 }, () => []),
    stockCount: 71,
    waste: [],
    baseRank: 5,
    direction: 1,
    awaitingDirection: false,
    redealsLeft: 2,
    canRedeal: false,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatBraidState', () => {
  it('renders the header, slots and an empty board', () => {
    const result = formatBraidState(makeState());
    expect(result).toContain('Braid');
    expect(result).toContain('base rank: 5 (up, in suit)');
    expect(result).toContain('foundations:');
    expect(result).toContain('braid: [  ] (0, tail only)');
    expect(result).toContain('stock: 71 (2 redeal(s) left)');
    expect(result).toContain('fd0:[  ]');
    expect(result).toContain('hp7:[  ]');
  });

  it('leads with the direction prompt while it is unset', () => {
    expect(formatBraidState(makeState({ awaitingDirection: true }))).toContain('direction not set yet');
  });

  it('labels a descending game', () => {
    expect(formatBraidState(makeState({ direction: 2 }))).toContain('(down, in suit)');
  });

  // An empty slot must keep its position, or the printed index stops matching
  // the index a hint refers to.
  it('keeps an empty slot in place', () => {
    const fields = [card('SPADE', 3), null, card('HEART', 7), null];
    const result = formatBraidState(makeState({ fields }));
    expect(result).toContain('fd1:[  ]');
    expect(result).toContain('fd3:[  ]');
    expect(result).toMatch(/fd0:\S+ fd1:\[ {2}\] fd2:\S+ fd3:\[ {2}\]/);
  });

  it('renders the braid tail, waste and move count', () => {
    const result = formatBraidState(
      makeState({
        braid: [card('SPADE', 2), card('HEART', 9)],
        waste: [card('CLOVER', 4)],
        moveCount: 7,
        canUndo: true,
      }),
    );
    expect(result).toContain('(2, tail only)');
    expect(result).toContain('moves: 7  undo:yes');
  });

  it('renders a hint with slot indices', () => {
    const result = formatBraidState(
      makeState({ hint: { fromZone: 'field', fromIdx: 2, toZone: 'foundation', toIdx: 1 } }),
    );
    expect(result).toContain('HINT: field2 → foundation1');
  });

  it('renders a hint without indices', () => {
    const result = formatBraidState(
      makeState({ hint: { fromZone: 'stock', fromIdx: -1, toZone: 'waste', toIdx: -1 } }),
    );
    expect(result).toContain('HINT: stock → waste');
  });

  it('renders stalemate, message and the win line', () => {
    const result = formatBraidState(makeState({ isStalemate: true, message: 'oops', phase: 1 }));
    expect(result).toContain('Stalemate');
    expect(result).toContain('oops');
    expect(result).toContain('Congratulations!');
  });
});
