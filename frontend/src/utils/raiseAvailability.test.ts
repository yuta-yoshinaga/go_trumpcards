import { describe, expect, it } from 'vitest';
import { raiseAvailability, raiseCost } from './raiseAvailability';

const base = { raiseCount: 0, maxRaises: 3, chips: 100, currentBet: 10, roundBet: 0, ante: 5 };

describe('raiseCost', () => {
  it('is what is still owed plus the ante', () => {
    expect(raiseCost({ currentBet: 10, roundBet: 4, ante: 5 })).toBe(11);
  });

  // **すでに賭けた分が現在のベットを超えることがある。**負の need にすると
  // アンテより安く見積もってしまう。
  it('floors what is owed at zero rather than going negative', () => {
    expect(raiseCost({ currentBet: 10, roundBet: 25, ante: 5 })).toBe(5);
  });
});

describe('raiseAvailability', () => {
  it('is open while under the cap and able to pay', () => {
    expect(raiseAvailability(base)).toBe('open');
  });

  it('reports the cap once the round allowance is spent', () => {
    expect(raiseAvailability({ ...base, raiseCount: 3 })).toBe('cap');
    expect(raiseAvailability({ ...base, raiseCount: 4 })).toBe('cap');
  });

  // **枚数不足も理由として区別する。**上限に達していないのに押せないとき、
  // 「レイズ 1/3回」とだけ出すと「まだできる」と読めてしまう。
  it('reports the chip shortage while still under the cap', () => {
    expect(raiseAvailability({ ...base, raiseCount: 1, chips: 14 })).toBe('chips');
  });

  it('is open at exactly the cost', () => {
    expect(raiseAvailability({ ...base, chips: 15 })).toBe('open');
  });

  // 両方成り立つときは上限を先に言う。ラウンド内では回復しないほう。
  it('prefers the cap when both hold', () => {
    expect(raiseAvailability({ ...base, raiseCount: 3, chips: 0 })).toBe('cap');
  });
});
