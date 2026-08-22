import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { DRAMAHA_HOLE_CARDS, dramahaBestFive, dramahaHands } from './dramahaBestFive';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('dramahaBestFive (the Omaha half)', () => {
  it('returns null without enough cards', () => {
    expect(dramahaBestFive([c('SPADE', 1)], [c('HEART', 2), c('HEART', 3), c('HEART', 4)])).toBeNull();
    expect(dramahaBestFive([c('SPADE', 1), c('SPADE', 2)], [c('HEART', 3), c('HEART', 4)])).toBeNull();
  });

  it('always picks exactly 2 hole + 3 board indices', () => {
    const best = dramahaBestFive(
      [c('SPADE', 1), c('HEART', 13), c('DIAMOND', 7), c('CLOVER', 2), c('HEART', 4)],
      [c('SPADE', 10), c('HEART', 11), c('DIAMOND', 12), c('CLOVER', 3), c('SPADE', 5)],
    );
    expect(best).not.toBeNull();
    expect(best?.holeIdx).toHaveLength(2);
    expect(best?.boardIdx).toHaveLength(3);
  });

  it('uses the two hole cards that complete a flush under the must-use-2 rule', () => {
    // Hole has two spades (idx 0,1); board has three spades (idx 0,1,2) → spade flush.
    const best = dramahaBestFive(
      [c('SPADE', 9), c('SPADE', 4), c('HEART', 2), c('DIAMOND', 6), c('CLOVER', 10)],
      [c('SPADE', 13), c('SPADE', 11), c('SPADE', 7), c('CLOVER', 3), c('DIAMOND', 8)],
    );
    expect(best?.holeIdx).toEqual([0, 1]); // the two spades
    expect(best?.boardIdx).toEqual([0, 1, 2]); // the three board spades
  });

  it('cannot use 3+ hole cards even when they would form a better hand', () => {
    // Three hole spades + two board spades would be a flush if 3 hole were allowed;
    // under Dramaha only 2 hole count, so trip nines beats the illegal flush.
    const best = dramahaBestFive(
      [c('SPADE', 2), c('SPADE', 3), c('SPADE', 4), c('HEART', 9), c('CLOVER', 13)],
      [c('SPADE', 5), c('SPADE', 6), c('HEART', 9), c('DIAMOND', 9), c('CLOVER', 12)],
    );
    expect(best?.holeIdx).toContain(3); // the H9 hole card
    expect(best?.holeIdx).toHaveLength(2);
    expect(best?.boardIdx).toHaveLength(3);
  });
});

describe('dramahaHands (both halves of the split)', () => {
  /** Five hole cards that are a pair of fours and nothing more. */
  const pairOfFours = [c('CLOVER', 4), c('DIAMOND', 4), c('HEART', 9), c('SPADE', 2), c('CLOVER', 7)];

  it('names the Omaha half from 2 hole + 3 board', () => {
    // Hole aces (idx 0,1) plus three board cards → three aces with the board ace.
    const { omaha } = dramahaHands(
      [c('SPADE', 1), c('HEART', 1), c('CLOVER', 7), c('DIAMOND', 3), c('SPADE', 8)],
      [c('DIAMOND', 1), c('CLOVER', 5), c('HEART', 12)],
    );
    expect(omaha?.key).toBe('threeOfAKind');
    expect(omaha?.holeIdx).toEqual([0, 1]);
    expect(omaha?.boardIdx).toHaveLength(3);
  });

  it('reads the draw half from the five hole cards ONLY, never the board', () => {
    // The hole is a plain pair of fours. The board is four more spades: read
    // together with the two hole spades that would be a flush, and the Omaha
    // half indeed becomes one. The draw half must stay a pair.
    const spadeBoard = [c('SPADE', 13), c('SPADE', 11), c('SPADE', 6), c('SPADE', 3), c('HEART', 10)];
    const spadeHole = [c('SPADE', 4), c('SPADE', 8), c('DIAMOND', 4), c('HEART', 9), c('CLOVER', 2)];

    const { omaha, draw } = dramahaHands(spadeHole, spadeBoard);

    expect(omaha?.key).toBe('flush');
    expect(draw?.key).toBe('onePair');
    // The draw half is the five hole cards, in order, and no board card.
    expect(draw?.holeIdx).toEqual([0, 1, 2, 3, 4]);
    expect(draw?.boardIdx).toEqual([]);
  });

  it('decides the draw half before there is any board at all', () => {
    const { omaha, draw } = dramahaHands(pairOfFours, []);
    expect(omaha).toBeNull();
    expect(draw?.key).toBe('onePair');
  });

  it('is unaffected by which board it is handed', () => {
    const straightBoard = [c('SPADE', 5), c('HEART', 6), c('DIAMOND', 7)];
    expect(dramahaHands(pairOfFours, straightBoard).draw?.key).toBe('onePair');
    expect(dramahaHands(pairOfFours, []).draw?.key).toBe('onePair');
  });

  it('always qualifies — a busted holding still ranks as high card', () => {
    const { draw } = dramahaHands([c('SPADE', 2), c('HEART', 5), c('DIAMOND', 9), c('CLOVER', 11), c('SPADE', 13)], []);
    expect(draw?.key).toBe('highCard');
  });

  it('has no draw half while the seat is not holding five cards', () => {
    expect(dramahaHands([c('SPADE', 2), c('HEART', 5)], []).draw).toBeNull();
    expect(dramahaHands([], []).draw).toBeNull();
  });

  it('deals five hole cards, not Omaha four', () => {
    expect(DRAMAHA_HOLE_CARDS).toBe(5);
  });
});
