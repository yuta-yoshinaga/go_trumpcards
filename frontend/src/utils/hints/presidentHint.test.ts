import { describe, expect, it } from 'vitest';
import type { Card, PresidentResponse } from '../../types/card';
import { getPresidentHint } from './presidentHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[] };

function base({ hand = [card('SPADE', 5), card('HEART', 9)], ...overrides }: Partial<PresidentResponse> & Extra = {}) {
  return {
    players: [
      { id: 0, isHuman: true, isFinished: false, rank: 0, cardCount: hand.length, cards: hand },
      { id: 1, isHuman: false, isFinished: false, rank: 0, cardCount: 5, cards: [] },
    ],
    currentTurn: 0,
    tableCards: [],
    lastPlayPlayerIdx: -1,
    gameEndFlag: false,
    revolutionActive: false,
    config: { revolutionEnabled: true, cardExchangeEnabled: true, passFieldFlushEnabled: true, cpuDifficulty: 1 },
    exchangeActions: [],
    cpuActions: [],
    humanAction: null,
    ...overrides,
  } as PresidentResponse;
}

describe('getPresidentHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getPresidentHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it("stays quiet on another seat's turn", () => {
    expect(getPresidentHint(base({ currentTurn: 1 }))).toBeNull();
  });

  it('stays quiet for a player who has gone out', () => {
    const s = base();
    s.players[0].isFinished = true;
    expect(getPresidentHint(s)).toBeNull();
  });

  it('leads the weakest card onto an empty table', () => {
    // 5 と 9 なら 5 が弱い。
    expect(getPresidentHint(base())?.targetAction).toBe('card-0');
  });

  it('ranks the two above the ace, which is above the king', () => {
    // 通常時は 2 が最強 (15)、A が 14。K(13) が一番弱い。
    const s = base({ hand: [card('SPADE', 2), card('HEART', 1), card('CLOVER', 13)] });
    expect(getPresidentHint(s)?.targetAction).toBe('card-2');
  });

  it('inverts the order under a revolution', () => {
    // 革命中は `18 - strength`。2(15) → 3、3 → 15 なので 2 が一番弱い。
    const s = base({ hand: [card('SPADE', 2), card('HEART', 3)], revolutionActive: true });
    expect(getPresidentHint(s)?.targetAction).toBe('card-0');

    // 同じ手を通常時に出すと逆に 3 が弱い。
    const normal = base({ hand: [card('SPADE', 2), card('HEART', 3)] });
    expect(getPresidentHint(normal)?.targetAction).toBe('card-1');
  });

  it('names the revolution in its reason so the inversion is visible', () => {
    const s = base({ revolutionActive: true });
    expect(getPresidentHint(s)?.reason).toBe('frontendHint.presidentLeadRevolution');
    expect(getPresidentHint(base())?.reason).toBe('frontendHint.presidentLead');
  });

  it('plays the weakest card that still beats the table', () => {
    // 場は 7。手札 5/9/J のうち 7 を超える一番弱いのは 9。
    const s = base({
      hand: [card('SPADE', 5), card('HEART', 9), card('CLOVER', 11)],
      tableCards: [card('DIAMOND', 7)],
    });
    expect(getPresidentHint(s)?.targetAction).toBe('card-1');
  });

  it('passes when nothing in hand beats the table', () => {
    const s = base({ hand: [card('SPADE', 5), card('HEART', 6)], tableCards: [card('DIAMOND', 2)] });
    expect(getPresidentHint(s)?.targetAction).toBe('pass');
  });

  it('passes under a revolution when the inverted order leaves nothing', () => {
    // 革命中、場が 3 (最強) なら何も超えられない。
    const s = base({
      hand: [card('SPADE', 5), card('HEART', 9)],
      tableCards: [card('DIAMOND', 3)],
      revolutionActive: true,
    });
    expect(getPresidentHint(s)?.targetAction).toBe('pass');
  });

  it('stays quiet without a hand', () => {
    expect(getPresidentHint(base({ hand: [] }))).toBeNull();
  });
});
