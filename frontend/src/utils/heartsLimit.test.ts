import { describe, expect, it } from 'vitest';
import { heartsNearPointLimit } from './heartsLimit';

describe('heartsNearPointLimit', () => {
  // 境界そのものを固定する。HeartsCuiPresenter の
  // `score*100 >= pointLimit*80` と同じ位置で切り替わることを確かめる。
  //
  // (`score >= limit * 0.8` でも上限 1〜2000 の総当たりで差は出なかった。
  //  整数式にしているのは Go 側と字面を揃えて差分を取りやすくするためで、
  //  浮動小数が壊れるからではない。)
  it('matches the CUI threshold exactly at the boundary', () => {
    // 100 点上限 → 80 点ちょうどから強調。
    expect(heartsNearPointLimit(79, 100)).toBe(false);
    expect(heartsNearPointLimit(80, 100)).toBe(true);
    expect(heartsNearPointLimit(81, 100)).toBe(true);
  });

  // 0.8 倍が割り切れない上限。境界の位置が Go と一致することを直接確かめる。
  it('handles limits whose 80% is not an integer', () => {
    // 55 * 0.8 = 44。44*100 = 4400 >= 55*80 = 4400 → 到達。
    expect(heartsNearPointLimit(43, 55)).toBe(false);
    expect(heartsNearPointLimit(44, 55)).toBe(true);

    // 33 * 0.8 = 26.4。27 で初めて超える。
    expect(heartsNearPointLimit(26, 33)).toBe(false);
    expect(heartsNearPointLimit(27, 33)).toBe(true);
  });

  it('never warns when the limit is disabled', () => {
    expect(heartsNearPointLimit(0, 0)).toBe(false);
    expect(heartsNearPointLimit(999, 0)).toBe(false);
    expect(heartsNearPointLimit(999, -1)).toBe(false);
  });

  it('does not warn at the start of a game', () => {
    expect(heartsNearPointLimit(0, 100)).toBe(false);
  });

  it('still warns once the limit is passed', () => {
    expect(heartsNearPointLimit(100, 100)).toBe(true);
    expect(heartsNearPointLimit(150, 100)).toBe(true);
  });
});
