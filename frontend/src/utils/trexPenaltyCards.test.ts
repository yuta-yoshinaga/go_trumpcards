import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { TrexContract } from '../types/phases';
import { trexIsPenaltyCard } from './trexPenaltyCards';

const c = (design: CardDesign, value: number): Card => ({ design, value });

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
