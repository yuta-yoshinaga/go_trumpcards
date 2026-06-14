import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { omahaBestFive } from './omahaBestFive';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('omahaBestFive', () => {
  it('returns null without enough cards', () => {
    expect(omahaBestFive([c('SPADE', 1)], [c('HEART', 2), c('HEART', 3), c('HEART', 4)])).toBeNull();
    expect(omahaBestFive([c('SPADE', 1), c('SPADE', 2)], [c('HEART', 3), c('HEART', 4)])).toBeNull();
  });

  it('always picks exactly 2 hole + 3 board indices', () => {
    const best = omahaBestFive(
      [c('SPADE', 1), c('HEART', 13), c('DIAMOND', 7), c('CLOVER', 2)],
      [c('SPADE', 10), c('HEART', 11), c('DIAMOND', 12), c('CLOVER', 3), c('SPADE', 5)],
    );
    expect(best).not.toBeNull();
    expect(best?.holeIdx).toHaveLength(2);
    expect(best?.boardIdx).toHaveLength(3);
  });

  it('uses the two hole cards that complete a flush under the must-use-2 rule', () => {
    // Hole has two spades (idx 0,1); board has three spades (idx 0,1,2) → spade flush.
    const best = omahaBestFive(
      [c('SPADE', 9), c('SPADE', 4), c('HEART', 2), c('DIAMOND', 6)],
      [c('SPADE', 13), c('SPADE', 11), c('SPADE', 7), c('CLOVER', 3), c('DIAMOND', 8)],
    );
    expect(best?.holeIdx).toEqual([0, 1]); // the two spades
    expect(best?.boardIdx).toEqual([0, 1, 2]); // the three board spades
  });

  it('cannot use 3+ hole cards even when they would form a better hand', () => {
    // Three hole spades + two board spades would be a flush if 3 hole were allowed;
    // under Omaha only 2 hole count, so a board-heavy pair beats the illegal flush.
    const best = omahaBestFive(
      [c('SPADE', 2), c('SPADE', 3), c('SPADE', 4), c('HEART', 9)],
      [c('SPADE', 5), c('SPADE', 6), c('HEART', 9), c('DIAMOND', 9), c('CLOVER', 12)],
    );
    // Best legal hand uses the H9 hole + the three nines on board → trip nines.
    expect(best?.holeIdx).toContain(3); // the H9 hole card
    expect(best?.holeIdx).toHaveLength(2);
    expect(best?.boardIdx).toHaveLength(3);
  });

  it('works with a 5-card (Big O) hole', () => {
    const best = omahaBestFive(
      [c('SPADE', 1), c('SPADE', 13), c('HEART', 2), c('DIAMOND', 6), c('CLOVER', 9)],
      [c('SPADE', 10), c('SPADE', 11), c('SPADE', 12), c('CLOVER', 3), c('DIAMOND', 8)],
    );
    // Royal-ish spade flush: A,K hole + 10,J,Q board spades.
    expect(best?.holeIdx).toEqual([0, 1]);
    expect(best?.boardIdx).toEqual([0, 1, 2]);
  });
});
