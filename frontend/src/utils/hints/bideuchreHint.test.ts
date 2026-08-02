import { describe, expect, it } from 'vitest';
import type { BidEuchreResponse, Card } from '../../types/card';
import { getBidEuchreHint } from './bideuchreHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[] };

function base({ hand = [card('SPADE', 9), card('HEART', 4)], ...overrides }: Partial<BidEuchreResponse> & Extra = {}) {
  return {
    players: [
      { id: 0, isHuman: true, team: 0, cardCount: hand.length, cards: hand },
      { id: 1, isHuman: false, team: 1, cardCount: 6, cards: [] },
    ],
    phase: 2,
    handNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 1,
    bids: [],
    highBid: null,
    declarerIdx: 0,
    trump: 0,
    trumpSuit: 1,
    trumpChosen: true,
    trick: [],
    validPlays: [0, 1],
    trickLeaderIdx: 0,
    trickNumber: 1,
    teamTricks: [0, 0],
    scores: [0, 0],
    lastResult: null,
    gameTarget: 32,
    gameEndFlag: false,
    message: '',
    ...overrides,
  } as unknown as BidEuchreResponse;
}

describe('getBidEuchreHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getBidEuchreHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet outside the play phase', () => {
    expect(getBidEuchreHint(base({ phase: 0 }))).toBeNull();
  });

  it('stays quiet when another seat is on turn', () => {
    expect(getBidEuchreHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('names the only legal card', () => {
    expect(getBidEuchreHint(base({ validPlays: [1] }))).toEqual({
      targetAction: 'card-1',
      reason: 'frontendHint.bideuchreForced',
      confidence: 'strong',
    });
  });

  // **札 0 も強制手になりうる。**真偽値で見ると先頭だけ落ちる。
  it('keeps a forced play on card index 0', () => {
    expect(getBidEuchreHint(base({ validPlays: [0] }))?.targetAction).toBe('card-0');
  });

  // **no-trump low は序列が逆。**9 が最強になる。
  it('warns that the ranking is reversed under no-trump low', () => {
    expect(getBidEuchreHint(base({ trump: 5, trumpSuit: 0 }))?.reason).toBe('frontendHint.bideuchreLowRanking');
  });

  // **trumpSuit だけでは 2 つの no-trump を見分けられない。**
  it('does not treat no-trump high as reversed', () => {
    expect(getBidEuchreHint(base({ trump: 4, trumpSuit: 0 }))?.reason).toBe('frontendHint.bideuchreChoose');
  });

  it('says nothing special under a suit declaration', () => {
    expect(getBidEuchreHint(base({ trump: 0, trumpSuit: 1 }))?.reason).toBe('frontendHint.bideuchreChoose');
  });

  it('stays quiet when the server offers no legal play', () => {
    expect(getBidEuchreHint(base({ validPlays: [] }))).toBeNull();
  });

  it('stays quiet without a visible hand', () => {
    expect(getBidEuchreHint(base({ hand: [] }))).toBeNull();
  });
});
