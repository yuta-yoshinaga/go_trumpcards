import { describe, expect, it } from 'vitest';
import type { Card, KarnoffelResponse } from '../../types/card';
import { getKarnoffelHint } from './karnoffelHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[] };

function base({ hand = [card('SPADE', 9), card('HEART', 4)], ...overrides }: Partial<KarnoffelResponse> & Extra = {}) {
  return {
    players: [
      { id: 0, isHuman: true, team: 0, cardCount: hand.length, cards: hand },
      { id: 1, isHuman: false, team: 1, cardCount: 5, cards: [] },
      { id: 2, isHuman: false, team: 0, cardCount: 5, cards: [] },
      { id: 3, isHuman: false, team: 1, cardCount: 5, cards: [] },
    ],
    phase: 0,
    handNumber: 1,
    currentPlayerIdx: 0,
    dealerIdx: 3,
    chosenSuit: 1,
    trick: [],
    validPlays: [0, 1],
    trickLeaderIdx: 0,
    trickNumber: 1,
    teamTricks: [0, 0],
    handsWon: [0, 0],
    lastResult: null,
    tricksToWin: 3,
    handSize: 5,
    targetHands: 2,
    gameEndFlag: false,
    winnerTeam: -1,
    message: '',
    ...overrides,
  } as unknown as KarnoffelResponse;
}

describe('getKarnoffelHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getKarnoffelHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet between hands', () => {
    expect(getKarnoffelHint(base({ phase: 1 }))).toBeNull();
  });

  it('stays quiet when another seat is on turn', () => {
    expect(getKarnoffelHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });

  // フォロー義務は無いが、悪魔札は最初のトリックをリードできない。
  it('names the only legal card', () => {
    expect(getKarnoffelHint(base({ validPlays: [1] }))).toEqual({
      targetAction: 'card-1',
      reason: 'frontendHint.karnoffelForced',
      confidence: 'strong',
    });
  });

  // **札 0 も強制手になりうる。**真偽値で見ると先頭だけ落ちる。
  it('keeps a forced play on card index 0', () => {
    expect(getKarnoffelHint(base({ validPlays: [0] }))?.targetAction).toBe('card-0');
  });

  // **5 枚のうち 3 トリックで手が決まる。**あと 1 つは早く来る。
  it('flags the trick that would take the hand', () => {
    expect(getKarnoffelHint(base({ teamTricks: [2, 0] }))?.reason).toBe('frontendHint.karnoffelOneAway');
  });

  // **相手の 2 トリックを自分のものと混同しない。**
  it('reads the tricks of the human team, not the other', () => {
    expect(getKarnoffelHint(base({ teamTricks: [0, 2] }))?.reason).toBe('frontendHint.karnoffelChoose');
  });

  it('says nothing special earlier in the hand', () => {
    expect(getKarnoffelHint(base({ teamTricks: [1, 1] }))?.reason).toBe('frontendHint.karnoffelChoose');
  });

  it('stays quiet when the server offers no legal play', () => {
    expect(getKarnoffelHint(base({ validPlays: [] }))).toBeNull();
  });

  it('stays quiet without a visible hand', () => {
    expect(getKarnoffelHint(base({ hand: [] }))).toBeNull();
  });
});
