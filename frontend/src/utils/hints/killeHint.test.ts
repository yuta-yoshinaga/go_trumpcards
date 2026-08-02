import { describe, expect, it } from 'vitest';
import type { Card, KilleResponse } from '../../types/card';
import { getKilleHint } from './killeHint';

const card = (value: number): Card => ({ design: 'SPADE', value });

type Extra = { strength?: number; hasCard?: boolean; isOut?: boolean };

function base({ strength = 5, hasCard = true, isOut = false, ...overrides }: Partial<KilleResponse> & Extra = {}) {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        card: hasCard ? card(strength) : null,
        strength,
        chips: 100,
        reentries: 0,
        reentryCost: 0,
        canReenter: true,
        isOut,
        knock: '',
      },
      {
        id: 1,
        isHuman: false,
        card: null,
        strength: 0,
        chips: 100,
        reentries: 0,
        reentryCost: 0,
        canReenter: true,
        isOut: false,
        knock: '',
      },
    ],
    phase: 0,
    roundNumber: 1,
    currentPlayerIdx: 0,
    dealerIdx: 1,
    stockCount: 20,
    pot: 30,
    events: [],
    loserIdxs: [],
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  } as unknown as KilleResponse;
}

describe('getKilleHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getKilleHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet outside the exchange phase', () => {
    expect(getKilleHint(base({ phase: 1 }))).toBeNull();
  });

  it('stays quiet when another seat is on turn', () => {
    expect(getKilleHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('stays quiet for a seat that is out', () => {
    expect(getKilleHint(base({ isOut: true }))).toBeNull();
  });

  it('swaps a weak card', () => {
    expect(getKilleHint(base({ strength: 4 }))).toEqual({
      targetAction: 'exchange',
      reason: 'frontendHint.killeSwapLow',
      confidence: 'moderate',
    });
  });

  it('keeps a strong card', () => {
    expect(getKilleHint(base({ strength: 15 }))).toEqual({
      targetAction: 'satisfied',
      reason: 'frontendHint.killeKeepHigh',
      confidence: 'moderate',
    });
  });

  // 境界: 11 以上で残す。
  it('treats the threshold itself as worth keeping', () => {
    expect(getKilleHint(base({ strength: 11 }))?.targetAction).toBe('satisfied');
    expect(getKilleHint(base({ strength: 10 }))?.targetAction).toBe('exchange');
  });

  // **親は山と交換する。**隣に断られない席なので行き先が違う。
  it('names the stock when the dealer swaps', () => {
    expect(getKilleHint(base({ strength: 4, dealerIdx: 0 }))?.reason).toBe('frontendHint.killeSwapStock');
  });

  // **strength は札の額面ではない。**交換で渡った Harlequin は 0 になる。
  it('reads the effective strength rather than the printed rank', () => {
    const s = base({ strength: 0 });
    s.players[0].card = card(21);
    expect(getKilleHint(s)).toBeNull();
  });

  it('stays quiet while the card is face down', () => {
    expect(getKilleHint(base({ hasCard: false }))).toBeNull();
  });
});
