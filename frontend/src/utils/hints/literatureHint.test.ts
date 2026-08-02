import { describe, expect, it } from 'vitest';
import type { Card, LiteratureResponse } from '../../types/card';
import { getLiteratureHint } from './literatureHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

const LOW_SPADES = [2, 3, 4, 5, 6, 7].map((v) => card('SPADE', v));
const LOW_HEARTS = [2, 3, 4, 5, 6, 7].map((v) => card('HEART', v));

type Extra = { hand?: Card[] };

function base({ hand = [card('SPADE', 3)], ...overrides }: Partial<LiteratureResponse> & Extra = {}) {
  return {
    players: [
      { id: 0, isHuman: true, team: 0, cardCount: hand.length, cards: hand, isCurrentTurn: true },
      { id: 1, isHuman: false, team: 1, cardCount: 8, cards: [], isCurrentTurn: false },
    ],
    phase: 0,
    currentPlayerIdx: 0,
    halfSuits: [0, 0],
    halfSuitCards: [LOW_SPADES, LOW_HEARTS],
    asks: [],
    claims: [],
    lastAsk: null,
    lastClaim: null,
    teamHalfSuits: [0, 0],
    cancelledCount: 0,
    openCount: 2,
    winThreshold: 5,
    halfSuitCnt: 8,
    gameEndFlag: false,
    winnerTeam: -1,
    message: '',
    ...overrides,
  } as unknown as LiteratureResponse;
}

describe('getLiteratureHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getLiteratureHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet when another seat is on turn', () => {
    expect(getLiteratureHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });

  // **持っているハーフスートしか聞けない。**
  it('offers an ask while a held half-suit is still open', () => {
    expect(getLiteratureHint(base({ hand: [card('SPADE', 3)] }))).toEqual({
      targetAction: 'ask',
      reason: 'frontendHint.literatureAskHeldSuit',
      confidence: 'moderate',
    });
  });

  // 持っている札のハーフスートが既に決着済みなら、聞ける先がない。
  it('falls back to a claim when every held half-suit is decided', () => {
    const s = base({ hand: [card('SPADE', 3)], halfSuits: [1, 0] });
    expect(getLiteratureHint(s)).toEqual({
      targetAction: 'claim',
      reason: 'frontendHint.literatureMustClaim',
      confidence: 'moderate',
    });
  });

  // **他のハーフスートが空いていても、そこの札を持っていなければ聞けない。**
  it('does not offer an ask for an open half-suit the player holds nothing of', () => {
    const s = base({ hand: [card('SPADE', 3)], halfSuits: [1, 0] });
    expect(getLiteratureHint(s)?.targetAction).toBe('claim');
  });

  it('offers an ask on the second half-suit when that is the held one', () => {
    const s = base({ hand: [card('HEART', 5)], halfSuits: [1, 0] });
    expect(getLiteratureHint(s)?.targetAction).toBe('ask');
  });

  // 取り消し済み (3) も未確定ではない。
  it('treats a cancelled half-suit as decided', () => {
    const s = base({ hand: [card('SPADE', 3)], halfSuits: [3, 0] });
    expect(getLiteratureHint(s)?.targetAction).toBe('claim');
  });

  it('stays quiet without a visible hand', () => {
    expect(getLiteratureHint(base({ hand: [] }))).toBeNull();
  });
});
