import { describe, expect, it } from 'vitest';
import { VINT_DENOM_COUNT, vintBidBeats, vintBidRank, vintLevelHasLegalBid } from './vintBid';

describe('vintBidRank', () => {
  it('ranks a higher level above every denomination of a lower one', () => {
    expect(vintBidRank(0, 4)).toBeGreaterThan(vintBidRank(VINT_DENOM_COUNT - 1, 3));
  });

  it('ranks the denominations ♠ < ♣ < ♦ < ♥ < NT within a level', () => {
    const ranks = [0, 1, 2, 3, 4].map((d) => vintBidRank(d, 3));
    expect(ranks).toEqual([...ranks].sort((a, b) => a - b));
    expect(new Set(ranks).size).toBe(5);
  });
});

describe('vintBidBeats', () => {
  it('allows anything when nobody has bid', () => {
    expect(vintBidBeats(0, 1, null)).toBe(true);
    expect(vintBidBeats(0, 1, undefined)).toBe(true);
  });

  // **同格は通らない。**サーバは `<=` で弾く。
  it('rejects an equal bid, not just a lower one', () => {
    expect(vintBidBeats(2, 4, { denom: 2, level: 4 })).toBe(false);
    expect(vintBidBeats(1, 4, { denom: 2, level: 4 })).toBe(false);
    expect(vintBidBeats(3, 4, { denom: 2, level: 4 })).toBe(true);
  });

  // **同じレベルの上位スートはまだ残る。**「レベルが上でなければ不可」では
  // 出せる宣言を潰しすぎる。
  it('keeps the higher denominations of the standing level open', () => {
    expect(vintBidBeats(4, 4, { denom: 0, level: 4 })).toBe(true);
    expect(vintBidBeats(0, 5, { denom: 4, level: 4 })).toBe(true);
  });
});

describe('vintLevelHasLegalBid', () => {
  it('keeps the standing level open while a higher denomination remains', () => {
    expect(vintLevelHasLegalBid(4, { denom: 0, level: 4 })).toBe(true);
  });

  it('closes the level once its top denomination is taken', () => {
    expect(vintLevelHasLegalBid(4, { denom: VINT_DENOM_COUNT - 1, level: 4 })).toBe(false);
    expect(vintLevelHasLegalBid(3, { denom: 0, level: 4 })).toBe(false);
  });

  it('leaves every level open before the first bid', () => {
    expect(vintLevelHasLegalBid(1, null)).toBe(true);
  });
});
