import { describe, expect, it } from 'vitest';
import type { BostonResponse, Card } from '../../types/card';
import { getBostonHint } from './bostonHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[] };

function base({ hand = [card('SPADE', 9), card('HEART', 4)], ...overrides }: Partial<BostonResponse> & Extra = {}) {
  return {
    players: [
      { id: 0, isHuman: true, cardCount: hand.length, cards: hand },
      { id: 1, isHuman: false, cardCount: 13, cards: [] },
      { id: 2, isHuman: false, cardCount: 13, cards: [] },
      { id: 3, isHuman: false, cardCount: 13, cards: [] },
    ],
    phase: 2,
    handNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    bids: [],
    highBid: null,
    bidOptions: [],
    declarerIdx: 1,
    partnerIdx: -1,
    trumpSuit: 1,
    exposed: false,
    trick: [],
    validPlays: [0, 1],
    trickLeaderIdx: 0,
    trickNumber: 1,
    declarerTricks: 0,
    bidMade: false,
    handSize: 13,
    targetHands: 4,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  } as unknown as BostonResponse;
}

describe('getBostonHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getBostonHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet outside the play phase', () => {
    expect(getBostonHint(base({ phase: 0 }))).toBeNull();
  });

  it('stays quiet when another seat is on turn', () => {
    expect(getBostonHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });

  // フォローは強制。合法手が 1 枚に潰れる。
  it('names the only legal card', () => {
    expect(getBostonHint(base({ validPlays: [1] }))).toEqual({
      targetAction: 'card-1',
      reason: 'frontendHint.bostonForced',
      confidence: 'strong',
    });
  });

  // **札 0 も強制手になりうる。**真偽値で見ると先頭だけ落ちる。
  it('keeps a forced play on card index 0', () => {
    expect(getBostonHint(base({ validPlays: [0] }))?.targetAction).toBe('card-0');
  });

  // **守る側と攻める側で欲しい結果が逆。**宣言側でなければ契約を崩す側。
  it('tells a defender to break the contract', () => {
    expect(getBostonHint(base({ declarerIdx: 1, partnerIdx: 2 }))?.reason).toBe('frontendHint.bostonBreakContract');
  });

  it('tells the declarer to push the contract', () => {
    expect(getBostonHint(base({ declarerIdx: 0, partnerIdx: -1 }))?.reason).toBe('frontendHint.bostonPushContract');
  });

  // **呼ばれた相方のトリックは宣言側に数える。**守る側と混同しない。
  it('treats a called partner as the declaring side', () => {
    expect(getBostonHint(base({ declarerIdx: 1, partnerIdx: 0 }))?.reason).toBe('frontendHint.bostonPushContract');
  });

  it('stays quiet when the server offers no legal play', () => {
    expect(getBostonHint(base({ validPlays: [] }))).toBeNull();
  });

  it('stays quiet without a visible hand', () => {
    expect(getBostonHint(base({ hand: [] }))).toBeNull();
  });
});
