import { describe, expect, it } from 'vitest';
import { computeIndianPokerEquity } from './indianPokerEquity';

describe('computeIndianPokerEquity', () => {
  it('returns 1 when there are no opponents', () => {
    expect(computeIndianPokerEquity([])).toBe(1);
  });

  // **K は最強ではない。**ドメインがエースを 14 にリマップするので、K の上に
  // A が 4 枚ある (#4690)。以前ここは 0 を期待しており、バグを固定していた。
  it('returns the ace count when the only opponent shows a King', () => {
    expect(computeIndianPokerEquity([13])).toBeCloseTo(4 / 51, 10);
  });

  it('returns a high equity when the only opponent shows a 2', () => {
    // Max opp = 2; ranks above = 3..14 の 12 ランク × 4 = 48; remaining = 51。
    const eq = computeIndianPokerEquity([2]);
    expect(eq).toBeCloseTo(48 / 51, 6);
  });

  it('subtracts visible cards in the above-max range', () => {
    // Three opponents: 10, 11, 12. Max = 12. Above = K と A で 8 枚。
    // 場に見えている above-max は無いので 8。Remaining = 52 - 3 = 49。
    expect(computeIndianPokerEquity([10, 11, 12])).toBeCloseTo(8 / 49, 6);
  });

  it('ignores invalid rank values', () => {
    // **14 はもう有効値。**範囲外は 1 未満と 15 以上 (エースは 14 に来るので
    // 1 は決して現れない)。max = 5; above = 6..14 の 9 ランク × 4 = 36。
    expect(computeIndianPokerEquity([1, 15, NaN, 5])).toBeCloseTo(36 / 51, 6);
  });

  it('returns 0 when every remaining card is at or below the opponent max', () => {
    // 2 人とも A(14) を見せている = これを超える札は無い。
    expect(computeIndianPokerEquity([14, 14])).toBe(0);
  });

  // **ドメインはエースを 14 にリマップしている (indianPokerCardRank)。**
  // ところがここは 1..13 で弾いていたので、相手のエースが「無効値」として
  // 計算から丸ごと外れ、最も危険な場面ほど勝率を高く見せていた (#4690)。
  it('counts an opponent ace instead of discarding it', () => {
    // 相手がエース(14)1枚。これを超える札は存在しないので勝率 0。
    expect(computeIndianPokerEquity([14])).toBe(0);
  });

  it('does not let an ace inflate the equity of a lower opponent', () => {
    // K(13) と A(14)。A を無視すると「K 超え = A の4枚」で 4/50 になるが、
    // 正しくは A を超える札が無いので 0。
    expect(computeIndianPokerEquity([13, 14])).toBe(0);
  });

  // **14 を勝ち筋としても数える。**フィルタだけ直してループ上限が 13 のままだと、
  // 相手が K のとき「A で勝てる」が数えられず勝率 0 になる。
  it('counts aces as winning cards when the opponent shows a king', () => {
    // 相手 K(13) 1枚 → 勝てるのは A の4枚。残り 51 枚。
    expect(computeIndianPokerEquity([13])).toBeCloseTo(4 / 51, 10);
  });
});
