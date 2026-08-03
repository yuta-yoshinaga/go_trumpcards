import { describe, expect, it } from 'vitest';
import type { Card, KalookiResponse } from '../../types/card';
import { getKalookiHint } from './kalookiHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[]; hasOpened?: boolean };

function base({
  hand = [card('SPADE', 3), card('HEART', 11)],
  hasOpened = true,
  ...overrides
}: Partial<KalookiResponse> & Extra = {}) {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: hand.length,
        cards: hand,
        melds: [],
        hasOpened,
        roundScore: 0,
        cumulativeScore: 0,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 13,
        cards: [],
        melds: [],
        hasOpened: false,
        roundScore: 0,
        cumulativeScore: 0,
      },
    ],
    phase: 1,
    openingThreshold: 51,
    currentPlayerIdx: 0,
    discardTop: card('CLOVER', 9),
    drawPileCount: 20,
    gameEndFlag: false,
    winnerIdx: -1,
    roundWinnerIdx: -1,
    message: '',
    config: { cpuDifficulty: 1, targetScore: 150 },
    ...overrides,
  } as KalookiResponse;
}

describe('getKalookiHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getKalookiHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet when another seat is on turn', () => {
    expect(getKalookiHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('stays quiet between rounds', () => {
    expect(getKalookiHint(base({ phase: 2 }))).toBeNull();
  });

  it('takes a discard that matches a rank in hand', () => {
    const hand = [card('SPADE', 9), card('HEART', 2)];
    const s = base({ phase: 0, hand, discardTop: card('CLOVER', 9) });
    expect(getKalookiHint(s)).toEqual({
      targetAction: 'takeDiscard',
      reason: 'frontendHint.kalookiTakeDiscard',
      confidence: 'moderate',
    });
  });

  it('draws from the stock when the discard connects with nothing', () => {
    const hand = [card('SPADE', 2), card('HEART', 5)];
    const s = base({ phase: 0, hand, discardTop: card('CLOVER', 9) });
    expect(getKalookiHint(s)?.targetAction).toBe('drawStock');
  });

  it('draws from the stock when there is nothing to take', () => {
    expect(getKalookiHint(base({ phase: 0, discardTop: null }))?.targetAction).toBe('drawStock');
  });

  // **開く前は崩さない。**規定点を一度に満たす必要があり、重い札こそ材料になる。
  it('tells an unopened player to build the opening meld', () => {
    expect(getKalookiHint(base({ hasOpened: false }))).toEqual({
      targetAction: 'meld',
      reason: 'frontendHint.kalookiOpenFirst',
      confidence: 'moderate',
    });
  });

  it('discards the heaviest loose card once opened', () => {
    const hand = [card('SPADE', 3), card('SPADE', 4), card('HEART', 11), card('DIAMOND', 2)];
    expect(getKalookiHint(base({ hand }))).toEqual({
      targetAction: 'card-2',
      reason: 'frontendHint.kalookiDiscardHeavy',
      confidence: 'moderate',
    });
  });

  // **札 0 も捨て札になりうる。**真偽値で見ると先頭だけ落ちる。
  it('keeps a discard suggestion on card index 0', () => {
    const hand = [card('HEART', 12), card('SPADE', 3), card('SPADE', 4)];
    expect(getKalookiHint(base({ hand }))?.targetAction).toBe('card-0');
  });

  it('falls back to the heaviest card when everything connects', () => {
    const hand = [card('SPADE', 3), card('SPADE', 4), card('SPADE', 5)];
    expect(getKalookiHint(base({ hand }))?.targetAction).toBe('card-2');
  });

  it('stays quiet without a visible hand', () => {
    expect(getKalookiHint(base({ hand: [] }))).toBeNull();
  });

  it('treats the ace as adjacent to the king, not only to the two', () => {
    // ドメインは Ace-high のランを認める。生の値で引くと A(1) と K(13) が
    // 12 離れて見えるので、K を拾う理由を見落とす。
    const s = base({
      phase: 0,
      hand: [card('SPADE', 1), card('HEART', 5)],
      discardTop: card('SPADE', 13),
    });
    expect(getKalookiHint(s)?.targetAction).toBe('takeDiscard');
  });

  it('still requires the same suit for an ace-king neighbour', () => {
    const s = base({
      phase: 0,
      hand: [card('SPADE', 1), card('HEART', 5)],
      discardTop: card('CLOVER', 13),
    });
    expect(getKalookiHint(s)?.targetAction).toBe('drawStock');
  });
});
