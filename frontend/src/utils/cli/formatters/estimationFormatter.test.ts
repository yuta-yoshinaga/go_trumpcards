import { describe, expect, it } from 'vitest';
import type { Card, EstimationResponse } from '../../../types/card';
import { formatEstimationState } from './estimationFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 2,
  cards: id === 0 ? [card('HEART', 11), card('SPADE', 1)] : [],
  bid: -1,
  callType: 0,
  trickCount: 0,
  roundScore: 0,
  totalScore: 0,
  ...over,
});

const state = (over: Partial<EstimationResponse> = {}): EstimationResponse =>
  ({
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 2,
    roundNumber: 2,
    trickNumber: 3,
    trumpSuit: 3,
    restrictedBid: -1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 2,
    currentTrick: [],
    validPlays: [0],
    gameEndFlag: false,
    winnerIdx: -1,
    config: { rounds: 5 },
    message: '',
    ...over,
  }) as unknown as EstimationResponse;

describe('formatEstimationState', () => {
  it('reports loading for a null state', () => {
    expect(formatEstimationState(null)).toBe('Loading...');
  });

  it('shows the round, trick and trump', () => {
    const out = formatEstimationState(state());
    expect(out).toContain('round 2/5');
    expect(out).toContain('trick 4/13');
    expect(out).toContain('trump: ♥');
  });

  // **得点表は常時出る。** 的中だけが得点になることは盤面から読めない。
  it('always states the scoring', () => {
    const out = formatEstimationState(state());
    expect(out).toContain('exact +(10+call)');
    expect(out).toContain('Dash (0) is ±23');
    expect(out).toContain('Risk doubles');
  });

  it('says trump is undecided before the dealer chooses', () => {
    expect(formatEstimationState(state({ trumpSuit: 0, phase: 0 }))).toContain('trump: undecided');
  });

  // **禁止値は出る前に言う。** 出してから拒否されるのでは遅い。両側を踏む。
  it('announces the barred call only when one applies', () => {
    expect(formatEstimationState(state({ restrictedBid: 4, phase: 1 }))).toContain('4 is barred');
    expect(formatEstimationState(state({ restrictedBid: -1 }))).not.toContain('barred');
  });

  // 3 種類の宣言表示をすべて踏む。
  it('labels each kind of call', () => {
    const out = formatEstimationState(
      state({
        players: [
          seat(0, { bid: 0, callType: 1 }),
          seat(1, { bid: 6, callType: 2 }),
          seat(2, { bid: 3, callType: 0 }),
          seat(3),
        ],
      } as Partial<EstimationResponse>),
    );
    expect(out).toContain('Dash (0)');
    expect(out).toContain('Risk (6)');
    expect(out).toContain('call 3');
    expect(out).toContain('no call');
  });

  it('marks legal cards in the hand', () => {
    const out = formatEstimationState(state({ validPlays: [1] }));
    expect(out).toMatch(/\[1\]\S+\*/);
    expect(out).not.toMatch(/\[0\]\S+\*/);
  });

  it('shows the current trick when one is under way', () => {
    const out = formatEstimationState(
      state({ currentTrick: [{ playerIdx: 1, card: card('SPADE', 13) }] } as Partial<EstimationResponse>),
    );
    expect(out).toContain('trick:');
  });

  it.each([
    [0, 'wins'],
    [-1, 'tie'],
  ])('reports outcome %s at game end', (winnerIdx, expected) => {
    expect(formatEstimationState(state({ gameEndFlag: true, winnerIdx }))).toContain(expected);
  });

  it('appends the server message', () => {
    expect(formatEstimationState(state({ message: 'hello' }))).toContain('hello');
  });
});
