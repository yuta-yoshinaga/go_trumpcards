import { describe, expect, it } from 'vitest';
import { FIVEHUNDRED_MISERE_VALUE, FIVEHUNDRED_OPEN_MISERE_VALUE, fivehundredBidValue } from './fivehundredBidValue';

describe('fivehundredBidValue (Avondale schedule)', () => {
  it('scores suit bids by base + 100 per trick above 6', () => {
    // ♠=40, ♣=60, ♦=80, ♥=100 base; +100 per extra trick.
    expect(fivehundredBidValue(6, 1)).toBe(40); // 6♠
    expect(fivehundredBidValue(7, 1)).toBe(140); // 7♠
    expect(fivehundredBidValue(10, 1)).toBe(440); // 10♠
    expect(fivehundredBidValue(6, 2)).toBe(60); // 6♣
    expect(fivehundredBidValue(6, 4)).toBe(80); // 6♦
    expect(fivehundredBidValue(6, 3)).toBe(100); // 6♥
    expect(fivehundredBidValue(10, 3)).toBe(500); // 10♥
  });

  it('scores no-trump (suit -1) from base 120', () => {
    expect(fivehundredBidValue(6, -1)).toBe(120); // 6NT
    expect(fivehundredBidValue(10, -1)).toBe(520); // 10NT
  });

  it('exposes the fixed Misère and Open Misère values', () => {
    expect(FIVEHUNDRED_MISERE_VALUE).toBe(250);
    expect(FIVEHUNDRED_OPEN_MISERE_VALUE).toBe(520);
  });
});
