import { describe, expect, it } from 'vitest';
import type { PontoonResponse } from '../../types/card';
import { PontoonPhase } from '../../types/phases';
import { getPontoonHint } from './pontoonHint';

type Overrides = Partial<PontoonResponse> & { total?: number; cardCount?: number };

function base({ total = 12, cardCount = 2, ...overrides }: Overrides = {}): PontoonResponse {
  return {
    seats: [
      {
        name: 'You',
        isCpu: false,
        hands: [{ cards: Array.from({ length: cardCount }, () => null), bet: 10, total, rank: 1, hidden: false }],
      },
    ],
    bankerIdx: 1,
    isHumanBanker: false,
    chips: 100,
    activeSeat: 0,
    activeHand: 0,
    nextBanker: -1,
    lastResult: '',
    phase: PontoonPhase.PLAYER_TURN,
    canStick: false,
    canTwist: true,
    canBuy: true,
    canSplit: false,
    message: '',
    ...overrides,
  } as PontoonResponse;
}

describe('getPontoonHint', () => {
  it('stays quiet outside the player turn', () => {
    expect(getPontoonHint(base({ phase: PontoonPhase.BET }))).toBeNull();
  });

  it('stays quiet when another seat is acting', () => {
    expect(getPontoonHint(base({ activeSeat: 1 }))).toBeNull();
  });

  it('stays quiet while the hand is hidden', () => {
    const s = base();
    s.seats[0].hands[0].hidden = true;
    expect(getPontoonHint(s)).toBeNull();
  });

  // 11 以下はどの札を引いてもバーストしない。賭け金を足す価値がある局面。
  it('buys on a total that cannot bust', () => {
    expect(getPontoonHint(base({ total: 11 }))).toEqual({
      targetAction: 'buy',
      reason: 'frontendHint.pontoonBuySafe',
      confidence: 'strong',
    });
  });

  // **買えないときは無料の 1 枚。**押せない手を勧めない。
  it('twists instead when buying is closed', () => {
    expect(getPontoonHint(base({ total: 11, canBuy: false }))?.targetAction).toBe('twist');
  });

  // 15 未満は宣言できないので、選択の余地なく引く。
  it('twists below the sticking minimum', () => {
    expect(getPontoonHint(base({ total: 14 }))).toEqual({
      targetAction: 'twist',
      reason: 'frontendHint.pontoonMustDraw',
      confidence: 'strong',
    });
  });

  it('sticks on a strong total', () => {
    expect(getPontoonHint(base({ total: 18, canStick: true }))).toEqual({
      targetAction: 'stick',
      reason: 'frontendHint.pontoonStickStrong',
      confidence: 'strong',
    });
  });

  it('sticks on a borderline total', () => {
    expect(getPontoonHint(base({ total: 16, canStick: true }))?.reason).toBe('frontendHint.pontoonStickBorderline');
  });

  // **4 枚で 15-16 は引く。**5 枚trick はポイント手に勝つので、
  // ここだけは境界の総計でも引く価値がある。
  it('chases the five-card trick from four cards', () => {
    expect(getPontoonHint(base({ total: 16, canStick: true, cardCount: 4 }))).toEqual({
      targetAction: 'twist',
      reason: 'frontendHint.pontoonFiveCard',
      confidence: 'moderate',
    });
  });

  it('splits when the server offers it', () => {
    expect(getPontoonHint(base({ total: 12, canSplit: true }))).toEqual({
      targetAction: 'split',
      reason: 'frontendHint.pontoonSplit',
      confidence: 'moderate',
    });
  });

  // 15 以上あるのにスティックが閉じている（分割後など）。宣言できない以上
  // 「止まれ」とは言えない。
  it('stays quiet on a stickable total when sticking is closed', () => {
    expect(getPontoonHint(base({ total: 18, canStick: false, canTwist: false }))).toBeNull();
  });

  it('stays quiet when nothing is legal', () => {
    expect(getPontoonHint(base({ canStick: false, canTwist: false, canBuy: false, canSplit: false }))).toBeNull();
  });
});
