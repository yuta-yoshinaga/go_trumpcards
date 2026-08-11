import { describe, expect, it } from 'vitest';
import type { Card, IsraeliWhistResponse } from '../../../types/card';
import { formatIsraeliWhistState } from './israeliwhistFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 2,
  cards: id === 0 ? [card('HEART', 11), card('SPADE', 1)] : [],
  auctionBid: -1,
  auctionSuit: 0,
  passed: false,
  bid: -1,
  trickCount: 0,
  roundScore: 0,
  totalScore: 0,
  ...over,
});

const state = (over: Partial<IsraeliWhistResponse> = {}): IsraeliWhistResponse =>
  ({
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 2,
    roundNumber: 2,
    trickNumber: 3,
    trumpSuit: 3,
    declarerIdx: 1,
    highBid: 7,
    highSuit: 3,
    minimumBid: 0,
    restrictedBid: -1,
    currentPlayerIdx: 0,
    auctionPlayerIdx: 0,
    bidPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 2,
    currentTrick: [],
    validPlays: [0],
    gameEndFlag: false,
    winnerIdx: -1,
    config: { rounds: 4 },
    message: '',
    ...over,
  }) as unknown as IsraeliWhistResponse;

describe('formatIsraeliWhistState', () => {
  it('reports loading for a null state', () => {
    expect(formatIsraeliWhistState(null)).toBe('Loading...');
  });

  it('shows the round, trick and trump', () => {
    const out = formatIsraeliWhistState(state());
    expect(out).toContain('round 2/4');
    expect(out).toContain('trick 4/13');
    expect(out).toContain('trump: ♥ (won with 7)');
  });

  // **得点表は常時出る。** 2 乗と全員一致の倍率は盤面から読めない。
  it('always states the scoring', () => {
    const out = formatIsraeliWhistState(state());
    expect(out).toContain('exact +(call^2 + 10)');
    expect(out).toContain('all-exact or all-miss doubles');
  });

  // オークション中と決着後で行が入れ替わる。両側を踏む。
  it('shows the standing bid while the auction runs', () => {
    const out = formatIsraeliWhistState(state({ phase: 0, trumpSuit: 0, highBid: 6, highSuit: 1 }));
    expect(out).toContain('auction: high bid 6 ♠');
    expect(out).not.toContain('trump: ');
  });

  // **押せない宣言は出る前に言う。** 両方向とその不在を踏む。
  it('announces the quota and the barred call only when they apply', () => {
    expect(formatIsraeliWhistState(state({ minimumBid: 9 }))).toContain('call at least 9');
    expect(formatIsraeliWhistState(state({ restrictedBid: 4 }))).toContain('4 is barred');

    const plain = formatIsraeliWhistState(state());
    expect(plain).not.toContain('call at least');
    expect(plain).not.toContain('barred');
  });

  // 1 段階目の立場が席ごとに出る。3 種すべて踏む。
  it('labels each seat with its auction standing', () => {
    const out = formatIsraeliWhistState(
      state({
        declarerIdx: 0,
        players: [seat(0, { auctionBid: 7, bid: 8 }), seat(1, { passed: true }), seat(2), seat(3)],
      } as Partial<IsraeliWhistResponse>),
    );
    expect(out).toContain('[won 7]');
    expect(out).toContain('[passed]');
    expect(out).toContain('[bidding]');
    expect(out).toContain('call 8');
    expect(out).toContain('no call');
  });

  it('marks legal cards in the hand', () => {
    const out = formatIsraeliWhistState(state({ validPlays: [1] }));
    expect(out).toMatch(/\[1\]\S+\*/);
    expect(out).not.toMatch(/\[0\]\S+\*/);
  });

  it('shows the current trick when one is under way', () => {
    const out = formatIsraeliWhistState(
      state({ currentTrick: [{ playerIdx: 1, card: card('SPADE', 13) }] } as Partial<IsraeliWhistResponse>),
    );
    expect(out).toContain('trick:');
  });

  it.each([
    [0, 'wins'],
    [-1, 'tie'],
  ])('reports outcome %s at game end', (winnerIdx, expected) => {
    expect(formatIsraeliWhistState(state({ gameEndFlag: true, winnerIdx }))).toContain(expected);
  });

  it('appends the server message', () => {
    expect(formatIsraeliWhistState(state({ message: 'hello' }))).toContain('hello');
  });
});
