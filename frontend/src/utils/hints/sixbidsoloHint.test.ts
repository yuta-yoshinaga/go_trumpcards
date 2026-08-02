import { describe, expect, it } from 'vitest';
import type { Card, SixBidSoloResponse } from '../../types/card';
import { getSixBidSoloHint } from './sixbidsoloHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[] };

function base({ hand = [card('SPADE', 9), card('HEART', 4)], ...overrides }: Partial<SixBidSoloResponse> & Extra = {}) {
  return {
    players: [
      { id: 0, isHuman: true, cardCount: hand.length, cards: hand },
      { id: 1, isHuman: false, cardCount: 11, cards: [] },
      { id: 2, isHuman: false, cardCount: 11, cards: [] },
    ],
    phase: 2,
    handNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 2,
    bids: [],
    highBid: { kind: 1, player: 0 },
    declarerIdx: 0,
    trumpSuit: 1,
    declared: true,
    calledCard: null,
    spreadOpen: false,
    widow: [],
    widowSize: 3,
    trick: [],
    validPlays: [0, 1],
    trickLeaderIdx: 0,
    trickNumber: 1,
    lastResult: null,
    bidTargets: [],
    totalPoints: 120,
    baseTarget: 60,
    gameEndFlag: false,
    message: '',
    ...overrides,
  } as unknown as SixBidSoloResponse;
}

describe('getSixBidSoloHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getSixBidSoloHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet outside the play phase', () => {
    expect(getSixBidSoloHint(base({ phase: 0 }))).toBeNull();
  });

  it('stays quiet when another seat is on turn', () => {
    expect(getSixBidSoloHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('names the only legal card', () => {
    expect(getSixBidSoloHint(base({ validPlays: [1] }))).toEqual({
      targetAction: 'card-1',
      reason: 'frontendHint.sixbidsoloForced',
      confidence: 'strong',
    });
  });

  // **札 0 も強制手になりうる。**真偽値で見ると先頭だけ落ちる。
  it('keeps a forced play on card index 0', () => {
    expect(getSixBidSoloHint(base({ validPlays: [0] }))?.targetAction).toBe('card-0');
  });

  // **ミゼールは取らないほうが勝ち。**宣言者と守る側で真逆になる。
  it('tells the misere declarer to duck', () => {
    const s = base({ highBid: { kind: 3, player: 0 }, declarerIdx: 0 });
    expect(getSixBidSoloHint(s)?.reason).toBe('frontendHint.sixbidsoloMisereDuck');
  });

  it('tells a defender to force tricks onto the misere declarer', () => {
    const s = base({ highBid: { kind: 3, player: 1 }, declarerIdx: 1 });
    expect(getSixBidSoloHint(s)?.reason).toBe('frontendHint.sixbidsoloMisereForce');
  });

  // スプレッドミゼール (5) も同じ扱い。
  it('treats a spread misere the same way', () => {
    const s = base({ highBid: { kind: 5, player: 0 }, declarerIdx: 0 });
    expect(getSixBidSoloHint(s)?.reason).toBe('frontendHint.sixbidsoloMisereDuck');
  });

  it('says nothing special under a points bid', () => {
    expect(getSixBidSoloHint(base({ highBid: { kind: 1, player: 0 } }))?.reason).toBe('frontendHint.sixbidsoloChoose');
  });

  // **trumpSuit だけでは 2 つのミゼールを見分けられない。**
  it('does not infer a misere from a zero trump suit', () => {
    const s = base({ highBid: { kind: 4, player: 0 }, trumpSuit: 0 });
    expect(getSixBidSoloHint(s)?.reason).toBe('frontendHint.sixbidsoloChoose');
  });

  it('stays quiet without a standing bid', () => {
    expect(getSixBidSoloHint(base({ highBid: null }))?.reason).toBe('frontendHint.sixbidsoloChoose');
  });

  it('stays quiet when the server offers no legal play', () => {
    expect(getSixBidSoloHint(base({ validPlays: [] }))).toBeNull();
  });

  it('stays quiet without a visible hand', () => {
    expect(getSixBidSoloHint(base({ hand: [] }))).toBeNull();
  });
});
