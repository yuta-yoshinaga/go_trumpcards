import { describe, expect, it } from 'vitest';
import type { Card, RistikontraResponse } from '../../types/card';
import { getRistikontraHint } from './ristikontraHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[] };

function base({
  hand = [card('SPADE', 5), card('HEART', 9)],
  ...overrides
}: Partial<RistikontraResponse> & Extra = {}) {
  return {
    players: [
      { id: 0, isHuman: true, cardCount: hand.length, cards: hand, capturedCount: 0, pistiPoints: 0 },
      { id: 1, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, pistiPoints: 0 },
    ],
    currentTurn: 0,
    pile: [card('CLOVER', 9)],
    pileTop: card('CLOVER', 9),
    pileCount: 1,
    lastCaptureIdx: -1,
    counterRank: 0,
    gameEndFlag: false,
    phase: 'play',
    remainingDeck: 20,
    winners: [],
    message: '',
    ...overrides,
  } as unknown as RistikontraResponse;
}

describe('getRistikontraHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getRistikontraHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet between rounds', () => {
    expect(getRistikontraHint(base({ phase: 'roundEnd' }))).toBeNull();
  });

  it('stays quiet when the opponent is on turn', () => {
    expect(getRistikontraHint(base({ currentTurn: 1 }))).toBeNull();
  });

  // **1 枚の場に同ランクを重ねると Pişti。**ボーナスが付く唯一の形。

  // 場が厚ければ同じ手でも Pişti にはならない。
  it('calls the same match an ordinary capture on a deeper pile', () => {
    expect(getRistikontraHint(base({ pileCount: 3 }))?.reason).toBe('frontendHint.ristikontraCapture');
  });

  // **札 0 も取り手になりうる。**真偽値で見ると先頭だけ落ちる。
  it('keeps a capture on card index 0', () => {
    const hand = [card('SPADE', 9), card('HEART', 5)];
    expect(getRistikontraHint(base({ hand }))?.targetAction).toBe('card-0');
  });

  // **ジャックは場があるときに使う。**

  it('lays the lowest card when nothing captures', () => {
    const hand = [card('SPADE', 5), card('HEART', 3)];
    expect(getRistikontraHint(base({ hand, pileTop: card('CLOVER', 8), pileCount: 2 }))).toEqual({
      targetAction: 'card-1',
      reason: 'frontendHint.ristikontraLayLow',
      confidence: 'moderate',
    });
  });

  // ジャックしか無ければ出すしかない。

  // 場が 1 枚でも、同じ数字を持っていなければ Pişti にはならない。

  it('stays quiet without a visible hand', () => {
    expect(getRistikontraHint(base({ hand: [] }))).toBeNull();
  });

  // --- 打ち返し (risti-kontra) ---------------------------------------------
  // 削除したのはクローン元 (ピシュティ) のジャック/Pişti ボーナスを見る 9 本。
  // このゲームに無い規則なので、代わりに実際の規則を見るテストを置く。

  it('steals the bundle when a counter is open and the rank is held', () => {
    // 席 1 が 9 で捕獲した直後。手札の 9 を出せば束ごと奪える。
    const h = getRistikontraHint(
      base({
        hand: [card('SPADE', 5), card('DIAMOND', 9)],
        counterRank: 9,
        lastCaptureIdx: 1,
        pileTop: null,
        pile: [],
        pileCount: 0,
      }),
    );
    expect(h).toEqual({
      targetAction: 'card-1',
      reason: 'frontendHint.ristikontraCounter',
      confidence: 'strong',
    });
  });

  it('prefers the counter over an ordinary capture', () => {
    // 場のトップ (5) も取れるが、束ごと奪える 9 のほうが振れ幅が大きい。
    const h = getRistikontraHint(
      base({
        hand: [card('SPADE', 5), card('DIAMOND', 9)],
        counterRank: 9,
        pileTop: card('CLOVER', 5),
        pile: [card('CLOVER', 5)],
        pileCount: 1,
      }),
    );
    expect(h?.targetAction).toBe('card-1');
    expect(h?.reason).toBe('frontendHint.ristikontraCounter');
  });

  it('does not claim a counter when the window is closed', () => {
    // counterRank が 0 = 直前に別のランクが出たので、もう奪えない。
    const h = getRistikontraHint(
      base({
        hand: [card('SPADE', 5), card('DIAMOND', 9)],
        counterRank: 0,
        pileTop: card('CLOVER', 9),
      }),
    );
    expect(h?.reason).toBe('frontendHint.ristikontraCapture');
  });

  it('does not claim a counter without the rank in hand', () => {
    const h = getRistikontraHint(
      base({
        hand: [card('SPADE', 5), card('DIAMOND', 3)],
        counterRank: 9,
        pileTop: null,
        pile: [],
        pileCount: 0,
      }),
    );
    expect(h?.reason).toBe('frontendHint.ristikontraLayLow');
  });

  // --- ジャックはただの札 ---------------------------------------------------

  it('treats a Jack as an ordinary card, not a wild sweep', () => {
    // ピシュティならジャックで場を総取りできるので strong な助言になる。
    // ここでは取れないので、一番小さい札を捨てる助言になるはず。
    const h = getRistikontraHint(
      base({
        hand: [card('SPADE', 11), card('DIAMOND', 4)],
        pileTop: card('CLOVER', 7),
        pile: [card('CLOVER', 7), card('HEART', 2)],
        pileCount: 2,
      }),
    );
    expect(h?.reason).toBe('frontendHint.ristikontraLayLow');
    expect(h?.targetAction).toBe('card-1'); // 4 のほう
  });

  it('is willing to discard a Jack when it is the lowest useful card', () => {
    // ピシュティはジャックを温存するので、ここで 11 を選ぶことはない。
    const h = getRistikontraHint(
      base({
        hand: [card('SPADE', 11), card('DIAMOND', 13)],
        pileTop: card('CLOVER', 7),
        pile: [card('CLOVER', 7)],
        pileCount: 1,
      }),
    );
    expect(h?.targetAction).toBe('card-0'); // ジャックを避けない
  });
});
