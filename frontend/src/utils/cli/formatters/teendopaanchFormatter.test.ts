import { describe, expect, it } from 'vitest';
import type { Card, TeenDoPaanchResponse } from '../../../types/card';
import { formatTeenDoPaanchState } from './teendopaanchFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 2,
  cards: id === 0 ? [card('HEART', 1), card('SPADE', 8)] : [],
  target: [5, 3, 2][id] ?? 0,
  trickCount: 0,
  met: 0,
  ...over,
});

const state = (over: Partial<TeenDoPaanchResponse> = {}): TeenDoPaanchResponse =>
  ({
    players: [seat(0), seat(1), seat(2)],
    phase: 1,
    roundNumber: 2,
    trickNumber: 3,
    trumpSuit: 3,
    fivePlayerIdx: 0,
    lastExchange: 0,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    currentTrick: [],
    validPlays: [0],
    gameEndFlag: false,
    winnerIdx: -1,
    config: { rounds: 3 },
    message: '',
    ...over,
  }) as unknown as TeenDoPaanchResponse;

describe('formatTeenDoPaanchState', () => {
  it('reports loading for a null state', () => {
    expect(formatTeenDoPaanchState(null)).toBe('Loading...');
  });

  it('shows the round, trick and phase', () => {
    const out = formatTeenDoPaanchState(state());
    expect(out).toContain('round 2/3');
    expect(out).toContain('trick 4/10');
    expect(out).toContain('PLAY');
  });

  // **ノルマは宣言ではなく割り当て。** 規則そのものを毎回書く。
  it('states that targets are assigned, not bid', () => {
    const out = formatTeenDoPaanchState(state());
    expect(out).toContain('assigned');
    expect(out).toContain('3/2/5');
  });

  // 切り札は未宣言と確定の両側を踏む。
  it('says how trump gets chosen while it is undeclared', () => {
    expect(formatTeenDoPaanchState(state({ trumpSuit: 0 }))).toContain('undeclared (the 5-target seat');
  });

  it('names the suit once trump is declared', () => {
    expect(formatTeenDoPaanchState(state())).toContain('trump: ♥');
    expect(formatTeenDoPaanchState(state({ trumpSuit: 9 }))).toContain('trump: ?');
  });

  it('marks the seat that declares trump', () => {
    expect(formatTeenDoPaanchState(state({ fivePlayerIdx: 1 }))).toContain('[trump]');
  });

  // **あと何トリック要るかが読めないと打ち方が決まらない。**
  it('shows each seat target, tricks taken and rounds met', () => {
    const out = formatTeenDoPaanchState(
      state({ players: [seat(0, { trickCount: 4, met: 2 }), seat(1), seat(2)] } as Partial<TeenDoPaanchResponse>),
    );
    expect(out).toContain('target 5, took 4');
    expect(out).toContain('met 2');
  });

  // **前ラウンドの札のやり取りは盤面に痕跡が残らない。**
  it('reports the exchange only when cards actually moved', () => {
    expect(formatTeenDoPaanchState(state())).not.toContain('changed hands');
    expect(formatTeenDoPaanchState(state({ lastExchange: 2 }))).toContain('2 cards changed hands');
  });

  it('marks the playable cards in your hand', () => {
    expect(formatTeenDoPaanchState(state())).toMatch(/your hand: \[0\].*\*/);
  });

  it('says the hand is empty rather than printing nothing', () => {
    const out = formatTeenDoPaanchState(
      state({ players: [{ ...seat(0), cards: [] }, seat(1), seat(2)] } as Partial<TeenDoPaanchResponse>),
    );
    expect(out).toContain('your hand: (empty)');
  });

  it('omits the hand block when no seat is human', () => {
    expect(formatTeenDoPaanchState(state({ players: [seat(1), seat(2)] }))).not.toContain('your hand:');
  });

  it('shows the current trick when one is in progress', () => {
    expect(formatTeenDoPaanchState(state({ currentTrick: [{ playerIdx: 1, card: card('SPADE', 9) }] }))).toContain(
      'trick: ',
    );
  });

  it.each([
    [0, 'wins on targets met'],
    [1, 'wins on targets met'],
    [-1, 'tie'],
  ])('reports the game result for winnerIdx %s', (winnerIdx, expected) => {
    expect(formatTeenDoPaanchState(state({ phase: 3, gameEndFlag: true, winnerIdx }))).toContain(expected);
  });

  it('falls back to the raw phase for an unknown value', () => {
    expect(formatTeenDoPaanchState(state({ phase: 99 }))).toContain('99');
  });

  it('echoes a server message', () => {
    expect(formatTeenDoPaanchState(state({ message: 'must follow the led suit' }))).toContain(
      'must follow the led suit',
    );
  });
});
