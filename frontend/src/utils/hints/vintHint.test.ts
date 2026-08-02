import { describe, expect, it } from 'vitest';
import type { Card, VintResponse } from '../../types/card';
import { getVintHint } from './vintHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[] };

function base({ hand = [card('SPADE', 9), card('HEART', 4)], ...overrides }: Partial<VintResponse> & Extra = {}) {
  return {
    players: [
      { id: 0, isHuman: true, team: 0, cardCount: hand.length, cards: hand },
      { id: 1, isHuman: false, team: 1, cardCount: 13, cards: [] },
      { id: 2, isHuman: false, team: 0, cardCount: 13, cards: [] },
      { id: 3, isHuman: false, team: 1, cardCount: 13, cards: [] },
    ],
    phase: 1,
    handNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    bids: [],
    highBid: null,
    declarerIdx: 1,
    trumpSuit: 1,
    trick: [],
    validPlays: [0, 1],
    trickLeaderIdx: 0,
    trickNumber: 1,
    teamTricks: [0, 0],
    below: [0, 0],
    above: [0, 0],
    gamesWon: [0, 0],
    lastResult: null,
    gameEndFlag: false,
    message: '',
    ...overrides,
  } as unknown as VintResponse;
}

describe('getVintHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getVintHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet outside the play phase', () => {
    expect(getVintHint(base({ phase: 0 }))).toBeNull();
  });

  it('stays quiet when another seat is on turn', () => {
    expect(getVintHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });

  // フォローは強制。合法手が 1 枚に潰れる。
  it('names the only legal card', () => {
    expect(getVintHint(base({ validPlays: [1] }))).toEqual({
      targetAction: 'card-1',
      reason: 'frontendHint.vintForced',
      confidence: 'strong',
    });
  });

  // **札 0 も強制手になりうる。**真偽値で見ると先頭だけ落ちる。
  it('keeps a forced play on card index 0', () => {
    expect(getVintHint(base({ validPlays: [0] }))?.targetAction).toBe('card-0');
  });

  it('tells the declarer they are on the declaring side', () => {
    expect(getVintHint(base({ declarerIdx: 0 }))?.reason).toBe('frontendHint.vintDeclaring');
  });

  // **相方は対角の席。**席 2 が宣言者なら席 0 も宣言側。
  it('treats the declarer partner as the declaring side', () => {
    expect(getVintHint(base({ declarerIdx: 2 }))?.reason).toBe('frontendHint.vintDeclaring');
  });

  it('tells a defender they are defending', () => {
    expect(getVintHint(base({ declarerIdx: 1 }))?.reason).toBe('frontendHint.vintDefending');
  });

  // 宣言者が未決 (-1) のあいだは相方も決まらない。
  it('does not call an undecided declarer a partner', () => {
    expect(getVintHint(base({ declarerIdx: -1 }))?.reason).toBe('frontendHint.vintDefending');
  });

  it('stays quiet when the server offers no legal play', () => {
    expect(getVintHint(base({ validPlays: [] }))).toBeNull();
  });

  it('stays quiet without a visible hand', () => {
    expect(getVintHint(base({ hand: [] }))).toBeNull();
  });
});
