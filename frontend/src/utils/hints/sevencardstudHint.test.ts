import { describe, expect, it } from 'vitest';
import type { Card, SevenCardStudResponse } from '../../types/card';
import { SevenCardStudPhase } from '../../types/phases';
import { getSevenCardStudHint } from './sevencardstudHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hole?: Card[]; door?: Card[]; currentBet?: number };

function base({
  hole = [card('SPADE', 4), card('HEART', 6)],
  door = [card('DIAMOND', 9)],
  currentBet = 0,
  ...overrides
}: Partial<SevenCardStudResponse> & Extra = {}) {
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
    phase: SevenCardStudPhase.FOURTH_STREET,
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
    ...overrides,
  } as unknown as SevenCardStudResponse;
}

describe('getSevenCardStudHint', () => {
  it('stays quiet once the hand is over', () => {
    expect(getSevenCardStudHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet at showdown, where there is nothing to decide', () => {
    expect(getSevenCardStudHint(base({ phase: SevenCardStudPhase.SHOWDOWN }))).toBeNull();
  });

  it('stays quiet when another seat is to act', () => {
    expect(getSevenCardStudHint(base({ currentTurn: 1 }))).toBeNull();
  });

  it('stays quiet for a folded seat', () => {
    const s = base();
    s.players[0].folded = true;
    expect(getSevenCardStudHint(s)).toBeNull();
  });

  it('stays quiet for an all-in seat, which has nothing left to choose', () => {
    const s = base();
    s.players[0].allIn = true;
    expect(getSevenCardStudHint(s)).toBeNull();
  });

  it('does not read the forced bring-in as strength', () => {
    const s = base({ phase: SevenCardStudPhase.THIRD_STREET, bringInPlayerIdx: 0 });
    expect(getSevenCardStudHint(s)?.reason).toBe('frontendHint.sevencardstudBringIn');
  });

  it('says nothing special about the bring-in seat on later streets', () => {
    // ブリングインは 3rd street だけの強制。以降は普通の判断に戻る。
    const s = base({ phase: SevenCardStudPhase.FOURTH_STREET, bringInPlayerIdx: 0 });
    expect(getSevenCardStudHint(s)?.reason).not.toBe('frontendHint.sevencardstudBringIn');
  });

  it('raises on a pair', () => {
    const s = base({ hole: [card('SPADE', 9), card('HEART', 4)], door: [card('DIAMOND', 9)] });
    expect(getSevenCardStudHint(s)?.reason).toBe('frontendHint.sevencardstudRaisePair');
  });

  it('finds a pair across the hole and door cards, not only within one', () => {
    const s = base({ hole: [card('SPADE', 3)], door: [card('DIAMOND', 3), card('CLOVER', 12)] });
    expect(getSevenCardStudHint(s)?.reason).toBe('frontendHint.sevencardstudRaisePair');
  });

  it('checks for free when nothing is owed', () => {
    expect(getSevenCardStudHint(base({ lastBet: 0 }))?.reason).toBe('frontendHint.sevencardstudCheckFree');
  });

  it('treats a matched bet as nothing owed', () => {
    // 既に同額を払っていれば負債は 0。差し引かないとフォールドを勧めてしまう。
    const s = base({ lastBet: 20, currentBet: 20 });
    expect(getSevenCardStudHint(s)?.reason).toBe('frontendHint.sevencardstudCheckFree');
  });

  it('calls a bet with a high card', () => {
    const s = base({ lastBet: 20, hole: [card('SPADE', 1), card('HEART', 4)], door: [card('DIAMOND', 6)] });
    expect(getSevenCardStudHint(s)?.reason).toBe('frontendHint.sevencardstudCallHigh');
  });

  it('folds a bet with nothing', () => {
    const s = base({ lastBet: 20, hole: [card('SPADE', 4), card('HEART', 6)], door: [card('DIAMOND', 9)] });
    expect(getSevenCardStudHint(s)?.reason).toBe('frontendHint.sevencardstudFoldWeak');
  });

  it('plays a low draw under Hi-Lo that it would fold at high only', () => {
    const low = [card('SPADE', 2), card('HEART', 3)];
    const door = [card('DIAMOND', 5), card('CLOVER', 6), card('SPADE', 7)];
    const hiLo = base({ lastBet: 20, isHiLo: true, hole: low, door });
    expect(hiLo?.isHiLo).toBe(true);
    expect(getSevenCardStudHint(hiLo)?.reason).toBe('frontendHint.sevencardstudPlayLow');

    // **同じ手が高目だけのテーブルではただの弱い手。**分岐が isHiLo を見ている確認。
    const high = base({ lastBet: 20, hole: low, door });
    expect(getSevenCardStudHint(high)?.reason).toBe('frontendHint.sevencardstudFoldWeak');
  });

  it('answers a paired low hand as a pair, before looking at the low', () => {
    // 2-2-3-4-5。ロー判定より先にペア分岐で返る。この順序が `lowCards` に
    // 重複ランクを渡さない根拠なので、順序自体を固定しておく。
    const s = base({
      lastBet: 20,
      isHiLo: true,
      hole: [card('SPADE', 2), card('HEART', 2)],
      door: [card('DIAMOND', 3), card('CLOVER', 4), card('SPADE', 5)],
    });
    expect(getSevenCardStudHint(s)?.reason).toBe('frontendHint.sevencardstudRaisePair');
  });

  it('does not call nine a low card', () => {
    const s = base({
      lastBet: 20,
      isHiLo: true,
      hole: [card('SPADE', 2), card('HEART', 3)],
      door: [card('DIAMOND', 5), card('CLOVER', 6), card('SPADE', 9)],
    });
    expect(getSevenCardStudHint(s)?.reason).not.toBe('frontendHint.sevencardstudPlayLow');
  });

  it('stays quiet without any cards', () => {
    expect(getSevenCardStudHint(base({ hole: [], door: [] }))).toBeNull();
  });
});
