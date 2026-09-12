import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { TrexContract } from '../types/phases';
import { TREX_PENALTIES, trexCardPenalty, trexIsPenaltyCard } from './trexPenaltyCards';

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

describe('trexIsPenaltyCard', () => {
  it('marks only the king of hearts under King of Hearts', () => {
    expect(trexIsPenaltyCard(c('HEART', 13), TrexContract.KING_OF_HEARTS)).toBe(true);
    expect(trexIsPenaltyCard(c('SPADE', 13), TrexContract.KING_OF_HEARTS)).toBe(false);
    expect(trexIsPenaltyCard(c('HEART', 12), TrexContract.KING_OF_HEARTS)).toBe(false);
  });

  it('marks every diamond under Diamonds', () => {
    expect(trexIsPenaltyCard(c('DIAMOND', 2), TrexContract.DIAMONDS)).toBe(true);
    expect(trexIsPenaltyCard(c('DIAMOND', 13), TrexContract.DIAMONDS)).toBe(true);
    expect(trexIsPenaltyCard(c('HEART', 13), TrexContract.DIAMONDS)).toBe(false);
  });

  // **クイーンはスートを問わない。**♥Q だけを見ると 3 枚見落とす。
  it('marks every queen under Queens, of any suit', () => {
    for (const d of ['SPADE', 'CLOVER', 'HEART', 'DIAMOND'] as CardDesign[]) {
      expect(trexIsPenaltyCard(c(d, 12), TrexContract.QUEENS)).toBe(true);
    }
    expect(trexIsPenaltyCard(c('HEART', 13), TrexContract.QUEENS)).toBe(false);
  });

  // **トリック契約では個々の札に失点は無い。**トリックそのものが失点なので、
  // 札を赤くすると嘘になる。
  it('marks nothing under the trick-counting and Trix contracts', () => {
    for (const contract of [TrexContract.TRICKS, TrexContract.DOMINOES, TrexContract.NONE]) {
      expect(trexIsPenaltyCard(c('HEART', 13), contract)).toBe(false);
      expect(trexIsPenaltyCard(c('DIAMOND', 5), contract)).toBe(false);
      expect(trexIsPenaltyCard(c('SPADE', 12), contract)).toBe(false);
    }
  });

  it('is false for a missing card', () => {
    expect(trexIsPenaltyCard(null, TrexContract.DIAMONDS)).toBe(false);
    expect(trexIsPenaltyCard(undefined, TrexContract.DIAMONDS)).toBe(false);
  });
});
