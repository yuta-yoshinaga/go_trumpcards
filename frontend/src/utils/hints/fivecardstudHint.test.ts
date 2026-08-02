import { describe, expect, it } from 'vitest';
import type { FiveCardStudResponse } from '../../types/card';
import { FiveCardStudPhase } from '../../types/phases';
import { getFiveCardStudHint } from './fivecardstudHint';

type Extra = { handRank?: number; currentBet?: number; folded?: boolean; chips?: number };

function base({
  handRank = 0,
  currentBet = 0,
  folded = false,
  chips = 200,
  ...overrides
}: Partial<FiveCardStudResponse> & Extra = {}) {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        holeCards: [],
        doorCards: [],
        chips,
        currentBet,
        folded,
        allIn: false,
        handRank,
        handName: '',
        bestHand: [],
        playStyleName: '',
        totalHands: 0,
        vpip: 0,
      },
    ],
    communityCard: null,
    pot: 30,
    sidePots: [],
    dealerIdx: 1,
    currentTurn: 0,
    phase: FiveCardStudPhase.THIRD_STREET,
    gameEndFlag: false,
    lastBet: 0,
    minRaise: 10,
    bettingLimit: 0,
    raiseCount: 0,
    maxBetAmount: 100,
    roundResults: [],
    cpuActions: [],
    handCount: 1,
    ante: 5,
    bringIn: 5,
    smallBet: 10,
    bigBet: 20,
    message: '',
    ...overrides,
  } as FiveCardStudResponse;
}

describe('getFiveCardStudHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getFiveCardStudHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet when another seat is on turn', () => {
    expect(getFiveCardStudHint(base({ currentTurn: 1 }))).toBeNull();
  });

  it('stays quiet after folding', () => {
    expect(getFiveCardStudHint(base({ folded: true }))).toBeNull();
  });

  it('stays quiet outside a betting street', () => {
    expect(getFiveCardStudHint(base({ phase: FiveCardStudPhase.SHOWDOWN }))).toBeNull();
  });

  // **ワンペア以上は賭けに行く。**handRank 1 = One Pair。
  it('raises on a made hand', () => {
    expect(getFiveCardStudHint(base({ handRank: 1 }))).toEqual({
      targetAction: 'raise',
      reason: 'frontendHint.fivecardstudRaiseMade',
      confidence: 'moderate',
    });
  });

  // ハイカードで、ただで見られるなら見る。
  it('checks a weak hand when nothing is owed', () => {
    expect(getFiveCardStudHint(base({ handRank: 0, lastBet: 0 }))).toEqual({
      targetAction: 'check',
      reason: 'frontendHint.fivecardstudCheckFree',
      confidence: 'moderate',
    });
  });

  // 払う必要があるならハイカードは降りる。
  it('folds a weak hand facing a bet', () => {
    expect(getFiveCardStudHint(base({ handRank: 0, lastBet: 20, currentBet: 0 }))).toEqual({
      targetAction: 'fold',
      reason: 'frontendHint.fivecardstudFoldWeak',
      confidence: 'moderate',
    });
  });

  // **既に払い込んでいる分は差し引く。**同額まで出していれば負債はない。
  it('treats an already-matched bet as nothing owed', () => {
    expect(getFiveCardStudHint(base({ handRank: 0, lastBet: 20, currentBet: 20 }))?.targetAction).toBe('check');
  });

  // 上限に達していれば上げられない。押せない手を勧めない。
  it('calls instead of raising when no raise is possible', () => {
    const s = base({ handRank: 1, lastBet: 20, currentBet: 0, maxBetAmount: 0 });
    expect(getFiveCardStudHint(s)).toEqual({
      targetAction: 'call',
      reason: 'frontendHint.fivecardstudCallMade',
      confidence: 'moderate',
    });
  });

  // 上限に達していて、かつ払う必要もない局面。
  it('checks a made hand when it can neither raise nor owes anything', () => {
    const s = base({ handRank: 2, lastBet: 0, maxBetAmount: 0 });
    expect(getFiveCardStudHint(s)).toEqual({
      targetAction: 'check',
      reason: 'frontendHint.fivecardstudCheckMade',
      confidence: 'moderate',
    });
  });

  it('stays quiet without a human seat', () => {
    const s = base();
    s.players = [];
    expect(getFiveCardStudHint(s)).toBeNull();
  });
});
