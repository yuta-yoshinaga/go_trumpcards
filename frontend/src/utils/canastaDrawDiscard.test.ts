import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { canastaDrawDiscardProblem, canastaIsBlack3, canastaIsWild } from './canastaDrawDiscard';

const c = (design: CardDesign, value: number): Card => ({ design, value });
const top = c('SPADE', 9);

describe('canastaIsWild', () => {
  it('treats jokers and twos as wild', () => {
    expect(canastaIsWild(c('JOKER', 0))).toBe(true);
    expect(canastaIsWild(c('HEART', 2))).toBe(true);
  });

  it('treats every other card as natural', () => {
    expect(canastaIsWild(c('HEART', 3))).toBe(false);
    expect(canastaIsWild(c('SPADE', 9))).toBe(false);
  });
});

describe('canastaDrawDiscardProblem', () => {
  it('asks for two cards when none are chosen', () => {
    expect(canastaDrawDiscardProblem([], top)).toBe('selectTwo');
  });

  it('asks for one more when only one is chosen', () => {
    expect(canastaDrawDiscardProblem([c('HEART', 9)], top)).toBe('selectOneMore');
  });

  it('rejects more than two', () => {
    expect(canastaDrawDiscardProblem([c('HEART', 9), c('CLOVER', 9), c('DIAMOND', 9)], top)).toBe('tooMany');
  });

  it('accepts a natural pair of the discard top rank', () => {
    expect(canastaDrawDiscardProblem([c('HEART', 9), c('CLOVER', 9)], top)).toBeNull();
  });

  it('rejects a wild in the pair, which the old count-only check let through', () => {
    expect(canastaDrawDiscardProblem([c('HEART', 9), c('JOKER', 0)], top)).toBe('wildInPair');
    expect(canastaDrawDiscardProblem([c('HEART', 2), c('CLOVER', 9)], top)).toBe('wildInPair');
  });

  it('rejects a pair that does not match the discard top', () => {
    expect(canastaDrawDiscardProblem([c('HEART', 8), c('CLOVER', 8)], top)).toBe('rankMismatch');
    // One matching card is not enough — both must.
    expect(canastaDrawDiscardProblem([c('HEART', 9), c('CLOVER', 8)], top)).toBe('rankMismatch');
  });

  it('rejects taking from an empty pile', () => {
    expect(canastaDrawDiscardProblem([c('HEART', 9), c('CLOVER', 9)], null)).toBe('pileEmpty');
  });
});

describe('canastaIsBlack3', () => {
  it('is true only for the spade and club threes', () => {
    expect([canastaIsBlack3(c('SPADE', 3)), canastaIsBlack3(c('CLOVER', 3))]).toEqual([true, true]);
    expect([canastaIsBlack3(c('HEART', 3)), canastaIsBlack3(c('SPADE', 4))]).toEqual([false, false]);
  });
});

describe('canastaDrawDiscardProblem with a black three on top', () => {
  it('refuses the take outright, even for a matching natural pair', () => {
    // Two black threes pass both the wild check (a three is not wild) and the
    // rank check (3 === 3), so without this guard the button would light up for
    // a request the server always rejects.
    expect(canastaDrawDiscardProblem([c('SPADE', 3), c('CLOVER', 3)], c('SPADE', 3))).toBe('blackThreeTop');
  });

  it('still allows a red three on top, which is not blocked', () => {
    expect(canastaDrawDiscardProblem([c('SPADE', 3), c('CLOVER', 3)], c('HEART', 3))).toBeNull();
  });
});

// #5502: ドメインは黒3の直後にワイルドトップも弾くのに、ここには無かった。
// 「取れます」と見せてからサーバに拒否される。
describe('canastaDrawDiscardProblem wild top', () => {
  it('refuses a joker on top', () => {
    expect(canastaDrawDiscardProblem([c('HEART', 9), c('CLOVER', 9)], c('JOKER', 0))).toBe('wildTop');
  });

  it('refuses a two on top', () => {
    expect(canastaDrawDiscardProblem([c('HEART', 2), c('CLOVER', 2)], c('SPADE', 2))).toBe('wildTop');
  });

  // 黒3の判定はワイルドより先。3 はワイルドではないので順序が入れ替わっても
  // 結果は同じだが、ドメインと同じ順序に揃えておく。
  it('still reports a black three top first', () => {
    expect(canastaDrawDiscardProblem([c('HEART', 3), c('DIAMOND', 3)], c('SPADE', 3))).toBe('blackThreeTop');
  });
});
