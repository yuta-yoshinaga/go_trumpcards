import { describe, expect, it } from 'vitest';
import type { Card, KaiserResponse } from '../../types/card';
import { getKaiserHint } from './kaiserHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[] };

function base({ hand = [card('SPADE', 9), card('HEART', 8)], ...overrides }: Partial<KaiserResponse> & Extra = {}) {
  return {
    players: [
      { id: 0, isHuman: true, team: 0, cardCount: hand.length, cards: hand, isDealer: false },
      { id: 1, isHuman: false, team: 1, cardCount: 8, cards: [], isDealer: true },
    ],
    phase: 2,
    handNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 1,
    bids: [],
    highBid: null,
    declarerIdx: 0,
    trumpSuit: 1,
    contract: 0,
    kittySize: 0,
    trick: [],
    trickLeaderIdx: 0,
    trickNumber: 1,
    validPlays: [0, 1],
    teamHandPoints: [0, 0],
    teamScores: [0, 0],
    heartFiveBy: -1,
    spadeThreeBy: -1,
    gameEndFlag: false,
    message: '',
    config: { cpuDifficulty: 1, targetScore: 52 },
    ...overrides,
  } as KaiserResponse;
}

describe('getKaiserHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getKaiserHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet when another seat is on turn', () => {
    expect(getKaiserHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('stays quiet outside the play phase', () => {
    expect(getKaiserHint(base({ phase: 0 }))).toBeNull();
  });

  // **♠3 は −3 点。**取られるより先に、取られない場面で出したい。
  it('warns while the spade three is still in hand', () => {
    const hand = [card('SPADE', 3), card('HEART', 8)];
    expect(getKaiserHint(base({ hand, validPlays: [0, 1] }))).toEqual({
      targetAction: 'card-0',
      reason: 'frontendHint.kaiserDumpSpadeThree',
      confidence: 'moderate',
    });
  });

  // **♥5 は +5 点。**自分から出すと相手に取られうるので急がない。
  it('does not push the heart five out early', () => {
    const hand = [card('HEART', 5), card('CLOVER', 8)];
    expect(getKaiserHint(base({ hand, validPlays: [0, 1] }))?.targetAction).not.toBe('card-0');
  });

  // フォローは強制。合法手が 1 枚に潰れる。
  it('names the only legal card', () => {
    expect(getKaiserHint(base({ validPlays: [1] }))).toEqual({
      targetAction: 'card-1',
      reason: 'frontendHint.kaiserForced',
      confidence: 'strong',
    });
  });

  // **札 0 も強制手になりうる。**真偽値で見ると先頭だけ落ちる。
  it('keeps a forced play on card index 0', () => {
    expect(getKaiserHint(base({ validPlays: [0] }))?.targetAction).toBe('card-0');
  });

  it('points at the legal plays when there is a choice', () => {
    expect(getKaiserHint(base({ validPlays: [0, 1] }))).toEqual({
      targetAction: 'card-0',
      reason: 'frontendHint.kaiserChoose',
      confidence: 'moderate',
    });
  });

  // ♠3 が合法でなければ勧めない。
  it('does not name the spade three when it cannot be played', () => {
    const hand = [card('SPADE', 3), card('HEART', 8)];
    expect(getKaiserHint(base({ hand, validPlays: [1] }))?.targetAction).toBe('card-1');
  });

  // 出せる札が ♥5 だけなら、避けようがないので出す。
  it('plays the heart five when it is the only choice left', () => {
    const hand = [card('HEART', 5), card('HEART', 5)];
    expect(getKaiserHint(base({ hand, validPlays: [0, 1] }))?.targetAction).toBe('card-0');
  });

  it('stays quiet when the server offers no legal play', () => {
    expect(getKaiserHint(base({ validPlays: [] }))).toBeNull();
  });

  it('stays quiet without a visible hand', () => {
    expect(getKaiserHint(base({ hand: [] }))).toBeNull();
  });
});
