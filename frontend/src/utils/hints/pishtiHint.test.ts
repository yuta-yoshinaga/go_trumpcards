import { describe, expect, it } from 'vitest';
import type { Card, PishtiResponse } from '../../types/card';
import { getPishtiHint } from './pishtiHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[] };

function base({ hand = [card('SPADE', 5), card('HEART', 9)], ...overrides }: Partial<PishtiResponse> & Extra = {}) {
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
    gameEndFlag: false,
    phase: 'play',
    remainingDeck: 20,
    winners: [],
    message: '',
    ...overrides,
  } as unknown as PishtiResponse;
}

describe('getPishtiHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getPishtiHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet between rounds', () => {
    expect(getPishtiHint(base({ phase: 'roundEnd' }))).toBeNull();
  });

  it('stays quiet when the opponent is on turn', () => {
    expect(getPishtiHint(base({ currentTurn: 1 }))).toBeNull();
  });

  // **1 枚の場に同ランクを重ねると Pişti。**ボーナスが付く唯一の形。
  it('names the pisti when exactly one card is showing', () => {
    expect(getPishtiHint(base({ pileCount: 1 }))).toEqual({
      targetAction: 'card-1',
      reason: 'frontendHint.pishtiPisti',
      confidence: 'strong',
    });
  });

  // 場が厚ければ同じ手でも Pişti にはならない。
  it('calls the same match an ordinary capture on a deeper pile', () => {
    expect(getPishtiHint(base({ pileCount: 3 }))?.reason).toBe('frontendHint.pishtiCapture');
  });

  // **札 0 も取り手になりうる。**真偽値で見ると先頭だけ落ちる。
  it('keeps a capture on card index 0', () => {
    const hand = [card('SPADE', 9), card('HEART', 5)];
    expect(getPishtiHint(base({ hand }))?.targetAction).toBe('card-0');
  });

  // **ジャックは場があるときに使う。**
  it('sweeps with a jack when the pile is worth taking', () => {
    const hand = [card('SPADE', 11), card('HEART', 3)];
    expect(getPishtiHint(base({ hand, pileCount: 4, pileTop: card('CLOVER', 8) }))).toEqual({
      targetAction: 'card-0',
      reason: 'frontendHint.pishtiJackSweep',
      confidence: 'moderate',
    });
  });

  it('does not throw a jack at an empty pile', () => {
    const hand = [card('SPADE', 11), card('HEART', 3)];
    const s = base({ hand, pile: [], pileTop: null, pileCount: 0 });
    expect(getPishtiHint(s)?.targetAction).toBe('card-1');
  });

  it('lays the lowest card when nothing captures', () => {
    const hand = [card('SPADE', 5), card('HEART', 3)];
    expect(getPishtiHint(base({ hand, pileTop: card('CLOVER', 8), pileCount: 2 }))).toEqual({
      targetAction: 'card-1',
      reason: 'frontendHint.pishtiLayLow',
      confidence: 'moderate',
    });
  });

  // ジャックしか無ければ出すしかない。
  it('plays a jack when it is the only card left', () => {
    const hand = [card('SPADE', 11)];
    const s = base({ hand, pile: [], pileTop: null, pileCount: 0 });
    expect(getPishtiHint(s)?.targetAction).toBe('card-0');
  });

  // 場が 1 枚でも、同じ数字を持っていなければ Pişti にはならない。
  it('does not claim a pisti without a matching card', () => {
    const hand = [card('SPADE', 5), card('HEART', 3)];
    const s = base({ hand, pileCount: 1, pileTop: card('CLOVER', 8) });
    expect(getPishtiHint(s)?.reason).toBe('frontendHint.pishtiLayLow');
  });

  it('stays quiet without a visible hand', () => {
    expect(getPishtiHint(base({ hand: [] }))).toBeNull();
  });
});
