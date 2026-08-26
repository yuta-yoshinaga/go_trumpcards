import { describe, expect, it } from 'vitest';
import type { Card, FollowTheQueenResponse } from '../../types/card';
import { FollowTheQueenPhase } from '../../types/phases';
import { getFollowTheQueenHint } from './followthequeenHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hole?: Card[]; door?: Card[]; currentBet?: number };

function base({
  hole = [card('SPADE', 4), card('HEART', 6)],
  door = [card('DIAMOND', 9)],
  currentBet = 0,
  ...overrides
}: Partial<FollowTheQueenResponse> & Extra = {}) {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        holeCards: hole,
        doorCards: door,
        chips: 500,
        currentBet,
        folded: false,
        allIn: false,
        handRank: 0,
        handName: '',
        bestHand: [],
        playStyleName: '',
        totalHands: 0,
        vpip: 0,
        pfr: 0,
        threeBet: 0,
        af: '',
      },
      {
        id: 1,
        isHuman: false,
        holeCards: [],
        doorCards: [card('CLOVER', 12)],
        chips: 500,
        currentBet: 0,
        folded: false,
        allIn: false,
        handRank: 0,
        handName: '',
        bestHand: [],
        playStyleName: '',
        totalHands: 0,
        vpip: 0,
        pfr: 0,
        threeBet: 0,
        af: '',
      },
    ],
    communityCard: null,
    pot: 30,
    sidePots: [],
    dealerIdx: 0,
    currentTurn: 0,
    phase: FollowTheQueenPhase.FOURTH_STREET,
    gameEndFlag: false,
    lastBet: 0,
    minRaise: 10,
    bettingLimit: 0,
    raiseCount: 0,
    maxBetAmount: 0,
    roundResults: [],
    cpuActions: [],
    handCount: 1,
    ante: 1,
    bringIn: 5,
    smallBet: 10,
    bigBet: 20,
    tournamentMode: false,
    anteLevelHands: 0,
    anteMultiplier: 100,
    tableSize: 2,
    bringInPlayerIdx: 1,
    rebuyAvailable: false,
    addonAvailable: false,
    rebuyCounts: [],
    addonUsed: [],
    rebuyEnabled: false,
    addonEnabled: false,
    rebuyMaxCount: 0,
    rebuyChips: 0,
    wildRank: 0,
    ...overrides,
  } as unknown as FollowTheQueenResponse;
}

describe('getFollowTheQueenHint', () => {
  it('stays quiet once the hand is over', () => {
    expect(getFollowTheQueenHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet at showdown, where there is nothing to decide', () => {
    expect(getFollowTheQueenHint(base({ phase: FollowTheQueenPhase.SHOWDOWN }))).toBeNull();
  });

  it('stays quiet when another seat is to act', () => {
    expect(getFollowTheQueenHint(base({ currentTurn: 1 }))).toBeNull();
  });

  it('stays quiet for a folded seat', () => {
    const s = base();
    s.players[0].folded = true;
    expect(getFollowTheQueenHint(s)).toBeNull();
  });

  it('stays quiet for an all-in seat, which has nothing left to choose', () => {
    const s = base();
    s.players[0].allIn = true;
    expect(getFollowTheQueenHint(s)).toBeNull();
  });

  it('does not read the forced bring-in as strength', () => {
    const s = base({ phase: FollowTheQueenPhase.THIRD_STREET, bringInPlayerIdx: 0 });
    expect(getFollowTheQueenHint(s)?.reason).toBe('frontendHint.followthequeenBringIn');
  });

  it('says nothing special about the bring-in seat on later streets', () => {
    // ブリングインは 3rd street だけの強制。以降は普通の判断に戻る。
    const s = base({ phase: FollowTheQueenPhase.FOURTH_STREET, bringInPlayerIdx: 0 });
    expect(getFollowTheQueenHint(s)?.reason).not.toBe('frontendHint.followthequeenBringIn');
  });

  it('raises on a pair', () => {
    const s = base({ hole: [card('SPADE', 9), card('HEART', 4)], door: [card('DIAMOND', 9)] });
    expect(getFollowTheQueenHint(s)?.reason).toBe('frontendHint.followthequeenRaisePair');
  });

  it('finds a pair across the hole and door cards, not only within one', () => {
    // 三枚目は 12 (Q) にしないこと —— このゲームでは常時ワイルドなので、
    // ペア分岐に届く前にワイルド分岐が返ってしまう。
    const s = base({ hole: [card('SPADE', 3)], door: [card('DIAMOND', 3), card('CLOVER', 8)] });
    expect(getFollowTheQueenHint(s)?.reason).toBe('frontendHint.followthequeenRaisePair');
  });

  it('checks for free when nothing is owed', () => {
    expect(getFollowTheQueenHint(base({ lastBet: 0 }))?.reason).toBe('frontendHint.followthequeenCheckFree');
  });

  it('treats a matched bet as nothing owed', () => {
    // 既に同額を払っていれば負債は 0。差し引かないとフォールドを勧めてしまう。
    const s = base({ lastBet: 20, currentBet: 20 });
    expect(getFollowTheQueenHint(s)?.reason).toBe('frontendHint.followthequeenCheckFree');
  });

  it('calls a bet with a high card', () => {
    const s = base({ lastBet: 20, hole: [card('SPADE', 1), card('HEART', 4)], door: [card('DIAMOND', 6)] });
    expect(getFollowTheQueenHint(s)?.reason).toBe('frontendHint.followthequeenCallHigh');
  });

  it('folds a bet with nothing', () => {
    const s = base({ lastBet: 20, hole: [card('SPADE', 4), card('HEART', 6)], door: [card('DIAMOND', 9)] });
    expect(getFollowTheQueenHint(s)?.reason).toBe('frontendHint.followthequeenFoldWeak');
  });

  it('stays quiet without any cards', () => {
    expect(getFollowTheQueenHint(base({ hole: [], door: [] }))).toBeNull();
  });
});

// **ヒントがワイルドを数えているか。**同じ3枚が、ワイルドが動いた途端に
// 「降りろ」から「レイズ」に変わる。数えなければ手の強さを2段階見誤る。
describe('getFollowTheQueenHint wild cards', () => {
  const junk = { hole: [card('SPADE', 3), card('HEART', 6)], door: [card('DIAMOND', 9)] };

  it('folds the same three cards when nothing is wild', () => {
    expect(getFollowTheQueenHint(base({ lastBet: 20, wildRank: 0, ...junk }))?.reason).toBe(
      'frontendHint.followthequeenFoldWeak',
    );
  });

  it('raises once one of them is wild', () => {
    expect(getFollowTheQueenHint(base({ lastBet: 20, wildRank: 9, ...junk }))?.reason).toBe(
      'frontendHint.followthequeenRaiseWildOne',
    );
  });

  it('counts a queen as wild with no wild rank set', () => {
    const s = base({
      lastBet: 20,
      wildRank: 0,
      hole: [card('SPADE', 3), card('HEART', 6)],
      door: [card('DIAMOND', 12)],
    });
    expect(getFollowTheQueenHint(s)?.reason).toBe('frontendHint.followthequeenRaiseWildOne');
  });

  it('says trips-or-better on two wilds', () => {
    const s = base({
      lastBet: 20,
      wildRank: 9,
      hole: [card('SPADE', 9), card('HEART', 6)],
      door: [card('DIAMOND', 9)],
    });
    expect(getFollowTheQueenHint(s)?.reason).toBe('frontendHint.followthequeenRaiseWildTwo');
  });

  it('prefers the wild reading over the pair reading', () => {
    // 9-9 は素のスタッドならワンペア。9 がワイルドなら実際はスリーカード以上。
    const s = base({
      lastBet: 20,
      wildRank: 9,
      hole: [card('SPADE', 9), card('HEART', 9)],
      door: [card('DIAMOND', 4)],
    });
    expect(getFollowTheQueenHint(s)?.reason).not.toBe('frontendHint.followthequeenRaisePair');
  });
});

// **押す手は Raise とは限らない。** `BettingControls` は負債があるときだけ
// Call+Raise を、無いときは Bet+Check を描く。負債 0 で 'raise' を指すのは
// 存在しないボタンを指すことで、#4643 で下のコール分岐について直したのと
// 同じ誤りを、あとから足した攻めの分岐で繰り返していた。
describe('getFollowTheQueenHint push action', () => {
  const strong = {
    wildRank: 9,
    hole: [card('SPADE', 9), card('HEART', 6)],
    door: [card('DIAMOND', 9)],
  };

  it('says bet, not raise, when nothing is owed', () => {
    const hint = getFollowTheQueenHint(base({ lastBet: 0, ...strong }));
    expect(hint?.targetAction).toBe('bet');
  });

  it('says raise when there is an outstanding bet', () => {
    const hint = getFollowTheQueenHint(base({ lastBet: 20, ...strong }));
    expect(hint?.targetAction).toBe('raise');
  });

  it('applies the same rule to the plain pair branch', () => {
    const paired = { wildRank: 0, hole: [card('SPADE', 4), card('HEART', 4)], door: [card('DIAMOND', 8)] };
    expect(getFollowTheQueenHint(base({ lastBet: 0, ...paired }))?.targetAction).toBe('bet');
    expect(getFollowTheQueenHint(base({ lastBet: 20, ...paired }))?.targetAction).toBe('raise');
  });
});
