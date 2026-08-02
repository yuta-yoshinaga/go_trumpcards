import { describe, expect, it } from 'vitest';
import type { Card, ChinchonResponse } from '../../types/card';
import { ChinchonPhase } from '../../types/phases';
import { getChinchonHint } from './chinchonHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[] };

function base({ hand = [card('SPADE', 3), card('HEART', 11)], ...overrides }: Partial<ChinchonResponse> & Extra = {}) {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: hand.length,
        cards: hand,
        roundScore: 0,
        cumulativeScore: 0,
        eliminated: false,
      },
      { id: 1, isHuman: false, cardCount: 7, cards: [], roundScore: 0, cumulativeScore: 0, eliminated: false },
    ],
    phase: ChinchonPhase.DISCARD,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: card('CLOVER', 9),
    drawPileCount: 20,
    gameEndFlag: false,
    winnerIdx: -1,
    knockerIdx: -1,
    knockerMelds: [],
    message: '',
    config: { cpuDifficulty: 1, targetScore: 100 },
    ...overrides,
  } as ChinchonResponse;
}

describe('getChinchonHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getChinchonHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet when the opponent is on turn', () => {
    expect(getChinchonHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('stays quiet between rounds', () => {
    expect(getChinchonHint(base({ phase: ChinchonPhase.ROUND_END }))).toBeNull();
  });

  // **同じランクが手札にあれば拾う。**セットに近づく。
  it('takes a discard that matches a rank in hand', () => {
    const hand = [card('SPADE', 9), card('HEART', 2)];
    const s = base({ phase: ChinchonPhase.DRAW, hand, discardTop: card('CLOVER', 9) });
    expect(getChinchonHint(s)).toEqual({
      targetAction: 'takeDiscard',
      reason: 'frontendHint.chinchonTakeDiscard',
      confidence: 'moderate',
    });
  });

  // **同じスートで隣り合っていれば拾う。**ランに近づく。
  it('takes a discard that sits next to a card of the same suit', () => {
    const hand = [card('CLOVER', 8), card('HEART', 2)];
    const s = base({ phase: ChinchonPhase.DRAW, hand, discardTop: card('CLOVER', 9) });
    expect(getChinchonHint(s)?.targetAction).toBe('takeDiscard');
  });

  it('draws from the stock when the discard connects with nothing', () => {
    const hand = [card('SPADE', 2), card('HEART', 5)];
    const s = base({ phase: ChinchonPhase.DRAW, hand, discardTop: card('CLOVER', 9) });
    expect(getChinchonHint(s)).toEqual({
      targetAction: 'drawStock',
      reason: 'frontendHint.chinchonDrawStock',
      confidence: 'moderate',
    });
  });

  it('draws from the stock when there is nothing to take', () => {
    const s = base({ phase: ChinchonPhase.DRAW, discardTop: null });
    expect(getChinchonHint(s)?.targetAction).toBe('drawStock');
  });

  // **繋がっていない札のうち一番重いものを捨てる。**点は残った札の合計。
  it('discards the heaviest card that connects with nothing', () => {
    const hand = [card('SPADE', 3), card('SPADE', 4), card('HEART', 11), card('DIAMOND', 2)];
    expect(getChinchonHint(base({ hand }))).toEqual({
      targetAction: 'card-2',
      reason: 'frontendHint.chinchonDiscardHeavy',
      confidence: 'moderate',
    });
  });

  // **札 0 も捨て札になりうる。**真偽値で見ると先頭だけ落ちる。
  it('keeps a discard suggestion on card index 0', () => {
    const hand = [card('HEART', 12), card('SPADE', 3), card('SPADE', 4)];
    expect(getChinchonHint(base({ hand }))?.targetAction).toBe('card-0');
  });

  // 全部が繋がっているなら崩す先はどれでも同じではない。一番重い札を出す。
  it('falls back to the heaviest card when everything connects', () => {
    const hand = [card('SPADE', 3), card('SPADE', 4), card('SPADE', 5)];
    expect(getChinchonHint(base({ hand }))?.targetAction).toBe('card-2');
  });

  it('stays quiet without a visible hand', () => {
    expect(getChinchonHint(base({ hand: [] }))).toBeNull();
  });

  it('stays quiet for an eliminated player', () => {
    const s = base();
    s.players[0].eliminated = true;
    expect(getChinchonHint(s)).toBeNull();
  });
});
