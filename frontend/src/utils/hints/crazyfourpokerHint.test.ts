import { describe, expect, it } from 'vitest';
import type { CrazyFourPokerResponse } from '../../types/card';
import { CrazyFourPokerPhase } from '../../types/phases';
import { getCrazyfourpokerHint } from './crazyfourpokerHint';

const base = {
  phase: CrazyFourPokerPhase.BET,
  gameEndFlag: false,
  hasAcesOrBetter: false,
  playerQualifies: true,
  maxMultiplier: 1,
} as unknown as CrazyFourPokerResponse;

const at = (over: Partial<CrazyFourPokerResponse>) => ({ ...base, ...over }) as CrazyFourPokerResponse;

describe('getCrazyfourpokerHint', () => {
  it('賭けフェーズと決着後は助言しない', () => {
    expect(getCrazyfourpokerHint(at({ phase: CrazyFourPokerPhase.BET }))).toBeNull();
    expect(getCrazyfourpokerHint(at({ phase: CrazyFourPokerPhase.RESULT }))).toBeNull();
  });

  it('終局後は助言しない', () => {
    expect(getCrazyfourpokerHint(at({ phase: CrazyFourPokerPhase.DECIDE, gameEndFlag: true }))).toBeNull();
  });

  it('エースのペア以上なら上げる助言', () => {
    const hint = getCrazyfourpokerHint(
      at({ phase: CrazyFourPokerPhase.DECIDE, hasAcesOrBetter: true, maxMultiplier: 3 }),
    );
    expect(hint?.targetAction).toBe('raise');
    expect(hint?.reason).toBe('frontendHint.crazyFourPokerAces');
    expect(hint?.confidence).toBe('strong');
  });

  it('エース未満なら同額で乗る助言', () => {
    const hint = getCrazyfourpokerHint(
      at({ phase: CrazyFourPokerPhase.DECIDE, hasAcesOrBetter: false, maxMultiplier: 1 }),
    );
    expect(hint?.targetAction).toBe('play');
    expect(hint?.reason).toBe('frontendHint.crazyFourPokerMinimum');
  });

  // **降りるのも定石の一部。** CUI 側のヒントは fold を出すので、Web だけ出さないと
  // 同じ局面で違う助言になる。判定はサーバの `playerQualifies` に従う。
  it('キングにも届かなければ降りる助言', () => {
    const hint = getCrazyfourpokerHint(
      at({ phase: CrazyFourPokerPhase.DECIDE, hasAcesOrBetter: false, playerQualifies: false }),
    );
    expect(hint?.targetAction).toBe('fold');
    expect(hint?.reason).toBe('frontendHint.crazyFourPokerFold');
  });

  // **playerQualifies を自分で計算し直さない。** キング以上の判定はドメインの規則。
  it('札の中身ではなく playerQualifies に従う', () => {
    const weak = getCrazyfourpokerHint(at({ phase: CrazyFourPokerPhase.DECIDE, playerQualifies: false }));
    const ok = getCrazyfourpokerHint(at({ phase: CrazyFourPokerPhase.DECIDE, playerQualifies: true }));
    expect(weak?.targetAction).toBe('fold');
    expect(ok?.targetAction).toBe('play');
  });

  // **助言側で倍率を決め直さない。**
  //
  // 3 倍が使えるかはドメインの規則 (`maxMultiplier`) で、ここで手役から計算すると
  // ゲームの本体である規則が 2 か所に増える。
  it('maxMultiplier を無視して手役から判定し直さない', () => {
    // サーバが「上限 1」と言っているのに hasAcesOrBetter だけ true という
    // 矛盾した状態でも、助言は hasAcesOrBetter に従うだけで倍率を作らない。
    const hint = getCrazyfourpokerHint(
      at({ phase: CrazyFourPokerPhase.DECIDE, hasAcesOrBetter: true, maxMultiplier: 1 }),
    );
    expect(hint).not.toBeNull();
    expect(hint).not.toHaveProperty('multiplier');
  });
});
