import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { TrexContract } from '../types/phases';
import { TREX_PENALTIES, trexCardPenalty } from './trexPenaltyCards';

const c = (design: CardDesign, value: number): Card => ({ design, value });

describe('TREX_PENALTIES', () => {
  it('has the expected penalty constants', () => {
    expect(TREX_PENALTIES.kingOfHearts).toBe(-75);
    expect(TREX_PENALTIES.diamond).toBe(-10);
    expect(TREX_PENALTIES.queen).toBe(-25);
  });
});

describe('trexCardPenalty', () => {
  it('returns -75 only for the king of hearts under King of Hearts', () => {
    expect(trexCardPenalty(c('HEART', 13), TrexContract.KING_OF_HEARTS)).toBe(-75);
    expect(trexCardPenalty(c('SPADE', 13), TrexContract.KING_OF_HEARTS)).toBe(0);
    expect(trexCardPenalty(c('HEART', 12), TrexContract.KING_OF_HEARTS)).toBe(0);
  });

  it('returns -10 for every diamond under Diamonds', () => {
    expect(trexCardPenalty(c('DIAMOND', 2), TrexContract.DIAMONDS)).toBe(-10);
    expect(trexCardPenalty(c('DIAMOND', 13), TrexContract.DIAMONDS)).toBe(-10);
    expect(trexCardPenalty(c('HEART', 13), TrexContract.DIAMONDS)).toBe(0);
  });

  it('returns -25 for every queen under Queens, of any suit', () => {
    for (const d of ['SPADE', 'CLOVER', 'HEART', 'DIAMOND'] as CardDesign[]) {
      expect(trexCardPenalty(c(d, 12), TrexContract.QUEENS)).toBe(-25);
    }
    expect(trexCardPenalty(c('HEART', 13), TrexContract.QUEENS)).toBe(0);
  });

  it('returns 0 under the trick-counting, Dominoes (Trix), and None contracts', () => {
    for (const contract of [TrexContract.TRICKS, TrexContract.DOMINOES, TrexContract.NONE]) {
      expect(trexCardPenalty(c('HEART', 13), contract)).toBe(0);
      expect(trexCardPenalty(c('DIAMOND', 5), contract)).toBe(0);
      expect(trexCardPenalty(c('SPADE', 12), contract)).toBe(0);
    }
  });

  it('returns 0 for missing card', () => {
    expect(trexCardPenalty(null, TrexContract.DIAMONDS)).toBe(0);
    expect(trexCardPenalty(undefined, TrexContract.DIAMONDS)).toBe(0);
  });
});
