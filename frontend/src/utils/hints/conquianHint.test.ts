import { describe, expect, it } from 'vitest';
import type { Card, ConquianResponse } from '../../types/card';
import { ConquianPhase } from '../../types/phases';
import { getConquianHint } from './conquianHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[] };

function base({ hand = [card('SPADE', 3), card('HEART', 11)], ...overrides }: Partial<ConquianResponse> & Extra = {}) {
  return {
    players: [
      { id: 0, isHuman: true, cardCount: hand.length, cards: hand, melds: [], wins: 0 },
      { id: 1, isHuman: false, cardCount: 10, cards: [], melds: [], wins: 0 },
    ],
    phase: ConquianPhase.MELD,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: card('CLOVER', 12),
    drawPileCount: 20,
    gameEndFlag: false,
    winnerIdx: -1,
    roundWinnerIdx: -1,
    tookDiscard: false,
    message: '',
    config: { cpuDifficulty: 1 },
    ...overrides,
  } as ConquianResponse;
}

describe('getConquianHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getConquianHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet when the opponent is on turn', () => {
    expect(getConquianHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('stays quiet between rounds', () => {
    expect(getConquianHint(base({ phase: ConquianPhase.ROUND_END }))).toBeNull();
  });

  // **同じランクが手札にあれば拾う。**セットに近づく。
  it('takes a discard that matches a rank in hand', () => {
    const hand = [card('SPADE', 7), card('HEART', 2)];
    const s = base({ phase: ConquianPhase.DRAW, hand, discardTop: card('CLOVER', 7) });
    expect(getConquianHint(s)).toEqual({
      targetAction: 'takeDiscard',
      reason: 'frontendHint.conquianTakeDiscard',
      confidence: 'moderate',
    });
  });

  // **同じスートで隣り合っていれば拾う。**ランに近づく。
  it('takes a discard that sits next to a card of the same suit', () => {
    const hand = [card('CLOVER', 6), card('HEART', 2)];
    const s = base({ phase: ConquianPhase.DRAW, hand, discardTop: card('CLOVER', 7) });
    expect(getConquianHint(s)?.targetAction).toBe('takeDiscard');
  });

  // **8/9/10 を抜いた 40 枚デッキなので、7 と J は隣接する。**生の value を
  // 引き算すると 4 離れて見え、繋がっていないと誤判定する (#4614 と同じ形)。
  it('treats a same-suit seven and jack as adjacent', () => {
    const hand = [card('CLOVER', 7), card('HEART', 2)];
    const s = base({ phase: ConquianPhase.DRAW, hand, discardTop: card('CLOVER', 11) });
    expect(getConquianHint(s)?.targetAction).toBe('takeDiscard');
  });

  it('does not treat a same-suit six and jack as adjacent', () => {
    const hand = [card('CLOVER', 6), card('HEART', 2)];
    const s = base({ phase: ConquianPhase.DRAW, hand, discardTop: card('CLOVER', 11) });
    expect(getConquianHint(s)?.targetAction).toBe('drawStock');
  });

  it('draws from the stock when the discard connects with nothing', () => {
    const hand = [card('SPADE', 2), card('HEART', 5)];
    const s = base({ phase: ConquianPhase.DRAW, hand, discardTop: card('CLOVER', 12) });
    expect(getConquianHint(s)).toEqual({
      targetAction: 'drawStock',
      reason: 'frontendHint.conquianDrawStock',
      confidence: 'moderate',
    });
  });

  it('draws from the stock when there is nothing to take', () => {
    const s = base({ phase: ConquianPhase.DRAW, discardTop: null });
    expect(getConquianHint(s)?.targetAction).toBe('drawStock');
  });

  // **繋がっていない札のうち一番重いものを捨てる。**点は残った札の合計。
  it('discards the heaviest card that connects with nothing', () => {
    const hand = [card('SPADE', 3), card('SPADE', 4), card('HEART', 13), card('DIAMOND', 2)];
    expect(getConquianHint(base({ hand }))).toEqual({
      targetAction: 'card-2',
      reason: 'frontendHint.conquianDiscardHeavy',
      confidence: 'moderate',
    });
  });

  // **札 0 も捨て札になりうる。**真偽値で見ると先頭だけ落ちる。
  it('keeps a discard suggestion on card index 0', () => {
    const hand = [card('HEART', 13), card('SPADE', 3), card('SPADE', 4)];
    expect(getConquianHint(base({ hand }))?.targetAction).toBe('card-0');
  });

  // 全部が繋がっているなら崩す先はどれでも同じではない。一番重い札を出す。
  it('falls back to the heaviest card when everything connects', () => {
    const hand = [card('SPADE', 3), card('SPADE', 4), card('SPADE', 5)];
    expect(getConquianHint(base({ hand }))?.targetAction).toBe('card-2');
  });

  // **8/9/10 はこのデッキに無い。**万一届いても隣接扱いしないことを固定する。
  // ドメインの conquianRankPosition も同じく 0 を返す。
  it('never treats a rank outside the forty-card deck as adjacent', () => {
    const hand = [card('CLOVER', 9), card('HEART', 2)];
    const s = base({ phase: ConquianPhase.DRAW, hand, discardTop: card('CLOVER', 10) });
    expect(getConquianHint(s)?.targetAction).toBe('drawStock');
  });

  it('stays quiet without a visible hand', () => {
    expect(getConquianHint(base({ hand: [] }))).toBeNull();
  });
});
