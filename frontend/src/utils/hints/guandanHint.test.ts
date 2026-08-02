import { describe, expect, it } from 'vitest';
import type { Card, GuandanResponse } from '../../types/card';
import { getGuandanHint } from './guandanHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[]; partnerRank?: number };

function base({
  hand = [card('SPADE', 5), card('HEART', 8)],
  partnerRank = 0,
  ...overrides
}: Partial<GuandanResponse> & Extra = {}) {
  return {
    players: [
      { id: 0, isHuman: true, team: 0, cardCount: hand.length, cards: hand, finishedRank: 0 },
      { id: 1, isHuman: false, team: 1, cardCount: 27, cards: [], finishedRank: 0 },
      { id: 2, isHuman: false, team: 0, cardCount: 27, cards: [], finishedRank: partnerRank },
      { id: 3, isHuman: false, team: 1, cardCount: 27, cards: [], finishedRank: 0 },
    ],
    phase: 1,
    handNumber: 1,
    currentPlayerIdx: 0,
    level: 2,
    teamLevels: [2, 2],
    declarerTeam: 0,
    lastCombo: null,
    lastPlayerIdx: -1,
    finished: [],
    tributes: [],
    tributeCancelled: false,
    lastResult: null,
    minLevel: 2,
    maxLevel: 14,
    advanceFirstSecond: 4,
    advanceFirstThird: 2,
    advanceFirstFourth: 1,
    gameEndFlag: false,
    winnerTeam: -1,
    message: '',
    ...overrides,
  } as unknown as GuandanResponse;
}

describe('getGuandanHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getGuandanHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet outside the play phase', () => {
    expect(getGuandanHint(base({ phase: 0 }))).toBeNull();
  });

  it('stays quiet when another seat is on turn', () => {
    expect(getGuandanHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });

  // **相方が上がっていると、勝敗ではなく順位でレベルが決まる。**
  it('flags a partner who has already gone out', () => {
    expect(getGuandanHint(base({ partnerRank: 1 }))).toEqual({
      targetAction: 'play',
      reason: 'frontendHint.guandanPartnerOut',
      confidence: 'moderate',
    });
  });

  // **相方は対角の席。**相手が上がっても順位の話にはしない。
  it('does not mistake an opponent going out for the partner', () => {
    const s = base();
    s.players[1].finishedRank = 1;
    expect(getGuandanHint(s)?.reason).not.toBe('frontendHint.guandanPartnerOut');
  });

  // **レベル札はエースより強く、赤 2 枚はワイルド。**
  it('points out level cards in hand', () => {
    const s = base({ hand: [card('SPADE', 2), card('HEART', 8)], level: 2 });
    expect(getGuandanHint(s)?.reason).toBe('frontendHint.guandanLevelCards');
  });

  it('does not call an ordinary rank a level card', () => {
    const s = base({ hand: [card('SPADE', 5), card('HEART', 8)], level: 2 });
    expect(getGuandanHint(s)?.reason).not.toBe('frontendHint.guandanLevelCards');
  });

  it('offers a free lead when nothing is on the table', () => {
    expect(getGuandanHint(base({ lastCombo: null }))?.reason).toBe('frontendHint.guandanFreeLead');
  });

  it('suggests weighing a pass against a live combo', () => {
    const s = base({
      lastCombo: { kind: 1, rank: 9, size: 1, cards: [card('CLOVER', 9)] } as GuandanResponse['lastCombo'],
    });
    expect(getGuandanHint(s)?.targetAction).toBe('pass');
  });

  it('stays quiet without a visible hand', () => {
    expect(getGuandanHint(base({ hand: [] }))).toBeNull();
  });
});
