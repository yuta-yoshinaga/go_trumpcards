import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import {
  evaluateIndianRummyDeclare,
  INDIAN_RUMMY_DEADWOOD_CAP,
  indianRummyCardPoints,
  indianRummyDeadwoodScore,
  indianRummyHasPureSequence,
  indianRummyIsWild,
  indianRummyValidateDeclaration,
} from './indianRummyDeclare';

/** Terse card constructor. */
const c = (design: Card['design'], value: number): Card => ({ design, value }) as Card;

// The expected verdicts below were cross-checked against the Go domain functions
// (IndianRummyValidateDeclaration / IndianRummyHasPureSequence / IndianRummyDeadwoodScore).

/** Valid: two pure runs (♠3-5, ♥7-9) + set of 10s + set of Ks. */
const validHand: Card[] = [
  c('SPADE', 3),
  c('SPADE', 4),
  c('SPADE', 5),
  c('HEART', 7),
  c('HEART', 8),
  c('HEART', 9),
  c('DIAMOND', 10),
  c('CLOVER', 10),
  c('SPADE', 10),
  c('DIAMOND', 13),
  c('CLOVER', 13),
  c('HEART', 13),
  c('SPADE', 13),
];

/** Invalid: only sets, no sequence at all -> no pure sequence. */
const noSequenceHand: Card[] = [
  c('DIAMOND', 2),
  c('CLOVER', 2),
  c('SPADE', 2),
  c('DIAMOND', 6),
  c('CLOVER', 6),
  c('SPADE', 6),
  c('DIAMOND', 10),
  c('CLOVER', 10),
  c('SPADE', 10),
  c('DIAMOND', 13),
  c('CLOVER', 13),
  c('HEART', 13),
  c('SPADE', 13),
];

/** Invalid: one pure run but the remaining 10 cards cannot all be melded. */
const uncoveredHand: Card[] = [
  c('SPADE', 3),
  c('SPADE', 4),
  c('SPADE', 5),
  c('HEART', 7),
  c('DIAMOND', 9),
  c('CLOVER', 11),
  c('DIAMOND', 2),
  c('CLOVER', 13),
  c('SPADE', 8),
  c('HEART', 6),
  c('DIAMOND', 4),
  c('HEART', 12),
  c('SPADE', 1),
];

/** Valid with wildRank=5: two pure runs + set of Ks (with a wild 5) + set of 2s. */
const wildHand: Card[] = [
  c('SPADE', 6),
  c('SPADE', 7),
  c('SPADE', 8),
  c('HEART', 9),
  c('HEART', 10),
  c('HEART', 11),
  c('DIAMOND', 13),
  c('CLOVER', 13),
  c('DIAMOND', 5), // wild (rank 5)
  c('DIAMOND', 2),
  c('CLOVER', 2),
  c('SPADE', 2),
  c('HEART', 2),
];

/** Valid using an Ace-high run (Q-K-A of spades). */
const aceHighHand: Card[] = [
  c('SPADE', 12),
  c('SPADE', 13),
  c('SPADE', 1),
  c('HEART', 7),
  c('HEART', 8),
  c('HEART', 9),
  c('DIAMOND', 10),
  c('CLOVER', 10),
  c('SPADE', 10),
  c('DIAMOND', 4),
  c('CLOVER', 4),
  c('HEART', 4),
  c('SPADE', 4),
];

/** Valid using a printed joker as the wild card in the set of 10s. */
const jokerHand: Card[] = [
  c('SPADE', 3),
  c('SPADE', 4),
  c('SPADE', 5),
  c('HEART', 7),
  c('HEART', 8),
  c('HEART', 9),
  c('DIAMOND', 10),
  c('CLOVER', 10),
  c('JOKER', 0),
  c('DIAMOND', 13),
  c('CLOVER', 13),
  c('HEART', 13),
  c('SPADE', 13),
];

describe('indianRummyIsWild', () => {
  it('treats printed jokers as wild', () => {
    expect(indianRummyIsWild(c('JOKER', 0), 0)).toBe(true);
  });
  it('treats the round wild rank as wild', () => {
    expect(indianRummyIsWild(c('SPADE', 5), 5)).toBe(true);
    expect(indianRummyIsWild(c('SPADE', 6), 5)).toBe(false);
  });
  it('does not treat any rank as wild when wildRank is 0', () => {
    expect(indianRummyIsWild(c('SPADE', 5), 0)).toBe(false);
  });
});

describe('indianRummyCardPoints', () => {
  it('scores wilds as 0, Ace and high cards as 10, pips as face value', () => {
    expect(indianRummyCardPoints(c('JOKER', 0), 0)).toBe(0);
    expect(indianRummyCardPoints(c('SPADE', 5), 5)).toBe(0); // wild rank
    expect(indianRummyCardPoints(c('SPADE', 1), 0)).toBe(10); // Ace
    expect(indianRummyCardPoints(c('SPADE', 7), 0)).toBe(7);
    expect(indianRummyCardPoints(c('SPADE', 10), 0)).toBe(10);
    expect(indianRummyCardPoints(c('SPADE', 13), 0)).toBe(10);
  });
});

describe('indianRummyValidateDeclaration - valid hands pass', () => {
  it('accepts two pure runs plus two sets', () => {
    expect(indianRummyValidateDeclaration(validHand, 0)).toBe(true);
  });
  it('accepts a set completed with the round wild rank', () => {
    expect(indianRummyValidateDeclaration(wildHand, 5)).toBe(true);
  });
  it('accepts an Ace-high sequence', () => {
    expect(indianRummyValidateDeclaration(aceHighHand, 0)).toBe(true);
  });
  it('accepts a set completed with a printed joker', () => {
    expect(indianRummyValidateDeclaration(jokerHand, 0)).toBe(true);
  });
});

describe('indianRummyValidateDeclaration - invalid hands fail', () => {
  it('rejects a hand with no pure sequence (only sets)', () => {
    expect(indianRummyValidateDeclaration(noSequenceHand, 0)).toBe(false);
  });
  it('rejects a hand with uncovered (unmelded) cards', () => {
    expect(indianRummyValidateDeclaration(uncoveredHand, 0)).toBe(false);
  });
  it('rejects a hand that is not exactly 13 cards', () => {
    expect(indianRummyValidateDeclaration(validHand.slice(0, 12), 0)).toBe(false);
    expect(indianRummyValidateDeclaration([...validHand, c('CLOVER', 9)], 0)).toBe(false);
  });
});

describe('indianRummyHasPureSequence', () => {
  it('detects a pure sequence', () => {
    expect(indianRummyHasPureSequence(validHand, 0)).toBe(true);
  });
  it('returns false when there is no pure sequence', () => {
    expect(indianRummyHasPureSequence(noSequenceHand, 0)).toBe(false);
  });
});

describe('indianRummyDeadwoodScore', () => {
  it('returns 0 for a fully melded valid hand', () => {
    expect(indianRummyDeadwoodScore(validHand, 0)).toBe(0);
  });
  it('returns the full cap when there is no pure sequence', () => {
    expect(indianRummyDeadwoodScore(noSequenceHand, 0)).toBe(INDIAN_RUMMY_DEADWOOD_CAP);
  });
  it('returns the minimum deadwood when a pure sequence exists', () => {
    // Cross-checked against the Go domain: 76 points remain unmelded.
    expect(indianRummyDeadwoodScore(uncoveredHand, 0)).toBe(76);
  });
});

describe('evaluateIndianRummyDeclare', () => {
  it('reports a valid hand with no penalty', () => {
    const p = evaluateIndianRummyDeclare(validHand, 0);
    expect(p.valid).toBe(true);
    expect(p.hasPureSequence).toBe(true);
    expect(p.unmeldedCount).toBe(0);
    expect(p.penalty).toBe(0);
  });

  it('flags a missing pure sequence with the full penalty', () => {
    const p = evaluateIndianRummyDeclare(noSequenceHand, 0);
    expect(p.valid).toBe(false);
    expect(p.hasPureSequence).toBe(false);
    expect(p.penalty).toBe(INDIAN_RUMMY_DEADWOOD_CAP);
  });

  it('flags unmelded cards with a positive count and points', () => {
    const p = evaluateIndianRummyDeclare(uncoveredHand, 0);
    expect(p.valid).toBe(false);
    expect(p.hasPureSequence).toBe(true);
    expect(p.unmeldedCount).toBeGreaterThan(0);
    expect(p.unmeldedPoints).toBe(76);
    expect(p.penalty).toBe(INDIAN_RUMMY_DEADWOOD_CAP);
  });
});

// #5501: 手札に出るのは合計のデッドウッドだけで、点数基準そのものはどこにも
// 書かれていなかった。**A も 10 点**という、ジンラミー系に慣れたプレイヤーほど
// 意外に感じる仕様なので、画面に凡例を出した。その凡例が主張する内容を、
// 実装から機械的に確かめる — 関数を変えたら凡例が嘘になる。
describe('indianRummyCardPoints backs the on-screen legend', () => {
  const WILD = 5;
  const card = (value: number, design = 'SPADE'): Card => ({ design, value }) as Card;

  it('scores A, 10, J, Q and K as ten', () => {
    for (const v of [1, 10, 11, 12, 13]) {
      expect(indianRummyCardPoints(card(v), WILD)).toBe(10);
    }
  });

  it('scores 2 through 9 at face value', () => {
    for (const v of [2, 3, 4, 6, 7, 8, 9]) {
      expect(indianRummyCardPoints(card(v), WILD)).toBe(v);
    }
  });

  it('scores a wild at zero', () => {
    expect(indianRummyCardPoints(card(WILD), WILD)).toBe(0);
  });
});
