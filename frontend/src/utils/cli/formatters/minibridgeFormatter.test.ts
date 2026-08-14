import { describe, expect, it } from 'vitest';
import type { Card, MinibridgeResponse } from '../../../types/card';
import { formatMinibridgeState } from './minibridgeFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 2,
  cards: id === 0 ? [card('HEART', 1), card('SPADE', 8)] : [],
  hcp: 10,
  team: id % 2,
  trickCount: 0,
  ...over,
});

const state = (over: Partial<MinibridgeResponse> = {}): MinibridgeResponse =>
  ({
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 1,
    roundNumber: 2,
    trickNumber: 3,
    contractLevel: 2,
    contractSuit: 3,
    requiredTricks: 8,
    declarerIdx: 0,
    dummyIdx: 2,
    dummyHand: [],
    lastMade: false,
    lastTricks: 0,
    teamScores: [110, 50],
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 0,
    currentTrick: [],
    validPlays: [0],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { rounds: 4 },
    message: '',
    ...over,
  }) as unknown as MinibridgeResponse;

describe('formatMinibridgeState', () => {
  it('reports loading for a null state', () => {
    expect(formatMinibridgeState(null)).toBe('Loading...');
  });

  it('shows the deal, trick and phase', () => {
    const out = formatMinibridgeState(state());
    expect(out).toContain('deal 2/4');
    expect(out).toContain('trick 4/13');
    expect(out).toContain('PLAY');
  });

  // **競りが無いこと自体が規則。**
  it('states that there is no auction', () => {
    const out = formatMinibridgeState(state());
    expect(out).toContain('no auction');
    expect(out).toContain('40 in total');
  });

  // 契約は未決定と確定の両側を踏む。
  it('shows the contract and the tricks it needs', () => {
    expect(formatMinibridgeState(state({ contractLevel: 0, contractSuit: 0, requiredTricks: 0 }))).toContain(
      'contract: not yet chosen',
    );

    const out = formatMinibridgeState(state());
    expect(out).toContain('contract: 2♥');
    expect(out).toContain('needs 8 tricks');
  });

  // **ノートランプは NT と書く。** 数字の 0 では読めない。
  it('writes no-trump as NT', () => {
    expect(formatMinibridgeState(state({ contractSuit: 0 }))).toContain('contract: 2NT');
  });

  // **HCP は 4 席ぶん出て、合計は 40。**
  it("shows every seat's HCP", () => {
    const out = formatMinibridgeState(state());
    const shown = [...out.matchAll(/HCP (\d+)/g)].map((m) => Number(m[1]));
    expect(shown).toHaveLength(4);
    expect(shown.reduce((a, b) => a + b, 0)).toBe(40);
  });

  it('marks the declarer and the dummy', () => {
    const out = formatMinibridgeState(state());
    expect(out).toContain('[declarer]');
    expect(out).toContain('[dummy]');
  });

  it('shows the running totals', () => {
    expect(formatMinibridgeState(state())).toContain('your pair 110 | their pair 50');
  });

  it('marks the seat on turn', () => {
    expect(formatMinibridgeState(state())).toMatch(/^>\S/m);
  });

  it('marks the playable cards in your hand', () => {
    const out = formatMinibridgeState(state());
    expect(out).toMatch(/\[0\]\S+\*/);
    expect(out).not.toMatch(/\[1\]\S+\*/);
  });

  // **ダミーは契約が決まってから公開される。**
  it("shows the dummy's hand only once it is revealed", () => {
    expect(formatMinibridgeState(state())).not.toContain("dummy's hand");
    expect(formatMinibridgeState(state({ dummyHand: [card('SPADE', 1)] }))).toContain("dummy's hand");
  });

  it('shows the current trick when there is one', () => {
    const out = formatMinibridgeState(
      state({ currentTrick: [{ playerIdx: 1, card: card('SPADE', 13) }] } as Partial<MinibridgeResponse>),
    );
    expect(out).toContain('trick:');
  });

  it('reports the deal result', () => {
    expect(formatMinibridgeState(state({ phase: 2, lastMade: true, lastTricks: 9 }))).toContain(
      'contract made: 9 of 8 tricks',
    );
    expect(formatMinibridgeState(state({ phase: 2, lastMade: false, lastTricks: 6 }))).toContain(
      'contract down: 6 of 8 tricks',
    );
  });

  it('reports the winning pair, and a tie', () => {
    expect(formatMinibridgeState(state({ phase: 3, gameEndFlag: true, winnerTeam: 0 }))).toContain('your pair wins');
    expect(formatMinibridgeState(state({ phase: 3, gameEndFlag: true, winnerTeam: 1 }))).toContain('their pair wins');
    expect(formatMinibridgeState(state({ phase: 3, gameEndFlag: true, winnerTeam: -1 }))).toContain('game over — tie');
  });

  it('shows the server message', () => {
    expect(formatMinibridgeState(state({ message: 'boom' }))).toContain('boom');
  });
});
