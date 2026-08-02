import { describe, expect, it } from 'vitest';
import type { Card, FiveCardStudResponse } from '../../types/card';
import { FiveCardStudPhase } from '../../types/phases';
import { getFiveCardStudHint } from './fivecardstudHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

/** A のカード値。1 で届くので数値だけ見ると一番低い札に見える。 */
const ACE_VALUE = 1;

type Extra = { hole?: Card[]; door?: Card[]; currentBet?: number; folded?: boolean };

function base({
  hole = [card('SPADE', 4)],
  door = [card('HEART', 7)],
  currentBet = 0,
  folded = false,
  ...overrides
}: Partial<FiveCardStudResponse> & Extra = {}) {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        holeCards: hole,
        doorCards: door,
        chips: 200,
        currentBet,
        folded,
        allIn: false,
        // ショーダウンまで 0 のまま届く。ヒントはこれを読まない (#4622)。
        handRank: 0,
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
    // 固定リミットなので常に 0。ヒントはこれを読まない (#4622)。
    maxBetAmount: 0,
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

  // **ペアは手札から数える。**handRank はショーダウンまで 0 で届くので使えない。
  it('raises on a pair, with maxBetAmount at its fixed-limit zero', () => {
    const s = base({ hole: [card('SPADE', 9)], door: [card('HEART', 9)] });
    expect(getFiveCardStudHint(s)).toEqual({
      targetAction: 'raise',
      reason: 'frontendHint.fivecardstudRaisePair',
      confidence: 'moderate',
    });
  });

  it('checks a weak hand when nothing is owed', () => {
    const s = base({ hole: [card('SPADE', 4)], door: [card('HEART', 7)], lastBet: 0 });
    expect(getFiveCardStudHint(s)).toEqual({
      targetAction: 'check',
      reason: 'frontendHint.fivecardstudCheckFree',
      confidence: 'moderate',
    });
  });

  it('folds a low hand facing a bet', () => {
    const s = base({ hole: [card('SPADE', 4)], door: [card('HEART', 7)], lastBet: 20 });
    expect(getFiveCardStudHint(s)?.targetAction).toBe('fold');
  });

  // 高い札があれば見に行く価値がある。
  it('calls with a high card facing a bet', () => {
    const s = base({ hole: [card('SPADE', 13)], door: [card('HEART', 7)], lastBet: 20 });
    expect(getFiveCardStudHint(s)).toEqual({
      targetAction: 'call',
      reason: 'frontendHint.fivecardstudCallHigh',
      confidence: 'moderate',
    });
  });

  // **A は 1 で届く。**数値だけで見ると一番低い札に見える。
  it('counts an ace as a high card', () => {
    const s = base({ hole: [card('SPADE', ACE_VALUE)], door: [card('HEART', 7)], lastBet: 20 });
    expect(getFiveCardStudHint(s)?.targetAction).toBe('call');
  });

  // **既に払い込んでいる分は差し引く。**同額まで出していれば負債はない。
  it('treats an already-matched bet as nothing owed', () => {
    const s = base({ hole: [card('SPADE', 4)], door: [card('HEART', 7)], lastBet: 20, currentBet: 20 });
    expect(getFiveCardStudHint(s)?.targetAction).toBe('check');
  });

  it('stays quiet before any card is dealt', () => {
    expect(getFiveCardStudHint(base({ hole: [], door: [] }))).toBeNull();
  });

  it('stays quiet without a human seat', () => {
    const s = base();
    s.players = [];
    expect(getFiveCardStudHint(s)).toBeNull();
  });
});
