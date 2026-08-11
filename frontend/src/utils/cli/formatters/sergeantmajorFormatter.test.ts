import { describe, expect, it } from 'vitest';
import type { Card, SergeantMajorResponse } from '../../../types/card';
import { formatSergeantMajorState } from './sergeantmajorFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 2,
  cards: id === 0 ? [card('HEART', 1), card('SPADE', 8)] : [],
  target: [8, 5, 3][id] ?? 0,
  trickCount: 0,
  score: 0,
  ...over,
});

const state = (over: Partial<SergeantMajorResponse> = {}): SergeantMajorResponse =>
  ({
    players: [seat(0), seat(1), seat(2)],
    phase: 2,
    roundNumber: 2,
    trickNumber: 3,
    trumpSuit: 3,
    kittySize: 0,
    discardCount: 4,
    lastExchange: 0,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 0,
    currentTrick: [],
    validPlays: [0],
    gameEndFlag: false,
    winnerIdx: -1,
    config: { rounds: 3 },
    message: '',
    ...over,
  }) as unknown as SergeantMajorResponse;

describe('formatSergeantMajorState', () => {
  it('reports loading for a null state', () => {
    expect(formatSergeantMajorState(null)).toBe('Loading...');
  });

  it('shows the round, trick and phase', () => {
    const out = formatSergeantMajorState(state());
    expect(out).toContain('round 2/3');
    expect(out).toContain('trick 4/16');
    expect(out).toContain('PLAY');
  });

  // **ノルマは席順で決まる。** 規則そのものを毎回書く。
  it('states that targets follow the seats', () => {
    const out = formatSergeantMajorState(state());
    expect(out).toContain('dealer 8, next 5, next 3');
    expect(out).toContain('nobody bids');
  });

  // 切り札は未宣言と確定の両側を踏む。
  it('mentions the kitty while trump is undeclared', () => {
    const out = formatSergeantMajorState(state({ phase: 0, trumpSuit: 0, kittySize: 4 }));
    expect(out).toContain('undeclared');
    expect(out).toContain('4-card kitty');
  });

  it('names the suit and the dealer once trump is declared', () => {
    expect(formatSergeantMajorState(state())).toContain('trump: ♥');
    expect(formatSergeantMajorState(state({ trumpSuit: 9 }))).toContain('trump: ?');
  });

  it('marks the dealer', () => {
    expect(formatSergeantMajorState(state({ dealerIdx: 1 }))).toContain('[dealer]');
  });

  // **あと何トリック要るかが読めないと打ち方が決まらない。**
  it('shows each seat target, tricks taken and running total', () => {
    const out = formatSergeantMajorState(
      state({ players: [seat(0, { trickCount: 6, score: 2 }), seat(1), seat(2)] } as Partial<SergeantMajorResponse>),
    );
    expect(out).toContain('target 8, took 6');
    expect(out).toContain('total 2');
  });

  // **前ラウンドの札のやり取りは盤面に痕跡が残らない。**
  it('reports the exchange only when cards actually moved', () => {
    expect(formatSergeantMajorState(state())).not.toContain('changed hands');
    expect(formatSergeantMajorState(state({ lastExchange: 3 }))).toContain('3 cards changed hands');
  });

  it('marks the playable cards in your hand', () => {
    expect(formatSergeantMajorState(state())).toMatch(/your hand: \[0\].*\*/);
  });

  it('says the hand is empty rather than printing nothing', () => {
    const out = formatSergeantMajorState(
      state({ players: [{ ...seat(0), cards: [] }, seat(1), seat(2)] } as Partial<SergeantMajorResponse>),
    );
    expect(out).toContain('your hand: (empty)');
  });

  it('omits the hand block when no seat is human', () => {
    expect(formatSergeantMajorState(state({ players: [seat(1), seat(2)] }))).not.toContain('your hand:');
  });

  it('shows the current trick when one is in progress', () => {
    expect(formatSergeantMajorState(state({ currentTrick: [{ playerIdx: 1, card: card('SPADE', 9) }] }))).toContain(
      'trick: ',
    );
  });

  it.each([
    [0, 'wins on points'],
    [1, 'wins on points'],
    [-1, 'tie'],
  ])('reports the game result for winnerIdx %s', (winnerIdx, expected) => {
    expect(formatSergeantMajorState(state({ phase: 4, gameEndFlag: true, winnerIdx }))).toContain(expected);
  });

  it('falls back to the raw phase for an unknown value', () => {
    expect(formatSergeantMajorState(state({ phase: 99 }))).toContain('99');
  });

  it('echoes a server message', () => {
    expect(formatSergeantMajorState(state({ message: 'must follow the led suit' }))).toContain(
      'must follow the led suit',
    );
  });
});
