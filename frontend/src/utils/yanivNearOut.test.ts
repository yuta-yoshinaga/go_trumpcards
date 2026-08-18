import { describe, expect, it } from 'vitest';
import { yanivIsNearOut } from './yanivNearOut';

describe('yanivIsNearOut', () => {
  it('warns past 80% of the limit', () => {
    expect(yanivIsNearOut(161, 200)).toBe(true);
  });

  // **ちょうど 80% は警告しない。**CUI は `>` で比べており、境界の扱いが
  // ずれると同じ点数で片方の画面だけ警告を出す。
  it('stays quiet exactly at 80%', () => {
    expect(yanivIsNearOut(160, 200)).toBe(false);
  });

  it('stays quiet below the threshold', () => {
    expect(yanivIsNearOut(100, 200)).toBe(false);
  });

  // 上限が無い設定では脱落しないので、警告する意味がない。
  it('never warns without a limit', () => {
    expect(yanivIsNearOut(999, 0)).toBe(false);
  });

  // 割り算にすると丸めで境界が動く組み合わせ (整数のままなら動かない)。
  it('keeps the boundary exact for limits that do not divide evenly', () => {
    expect(yanivIsNearOut(84, 105)).toBe(false); // 84/105 = ちょうど 0.8
    expect(yanivIsNearOut(85, 105)).toBe(true);
  });
});
