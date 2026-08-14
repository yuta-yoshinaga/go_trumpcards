import { describe, expect, it } from 'vitest';
import type { BhabhiResponse, Card } from '../../../types/card';
import { formatBhabhiState } from './bhabhiFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 2,
  cards: id === 0 ? [card('HEART', 10), card('SPADE', 1)] : [],
  rank: -1,
  pickups: 0,
  ...over,
});

const state = (over: Partial<BhabhiResponse> = {}): BhabhiResponse =>
  ({
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    trickNumber: 4,
    leadSuit: 0,
    pile: [],
    lastPickupIdx: -1,
    lastPickupSize: 0,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    validPlays: [0],
    aliveCount: 4,
    gameEndFlag: false,
    bhabhiIdx: -1,
    stalemate: false,
    stalemateTricks: 300,
    config: { playerCnt: 4 },
    message: '',
    ...over,
  }) as unknown as BhabhiResponse;

describe('formatBhabhiState', () => {
  it('reports loading for a null state', () => {
    expect(formatBhabhiState(null)).toBe('Loading...');
  });

  it('shows the trick, table size and how many are still in', () => {
    const out = formatBhabhiState(state());
    expect(out).toContain('trick 5');
    expect(out).toContain('4 players');
    expect(out).toContain('4 still in');
    expect(out).toContain('PLAY');
  });

  // **勝者ではなく敗者を決めるゲーム。** 目的を毎回書く。
  it('states that the game names a loser', () => {
    expect(formatBhabhiState(state())).toContain('the last player still holding cards is the Bhabhi');
  });

  // リードは未開始と確定の両側を踏む。
  it('says the table is empty before anyone leads', () => {
    expect(formatBhabhiState(state())).toContain('the table is empty');
  });

  it('names the led suit and the pile size once someone leads', () => {
    const out = formatBhabhiState(
      state({ leadSuit: 3, pile: [{ playerIdx: 1, card: card('HEART', 5) }] } as Partial<BhabhiResponse>),
    );
    expect(out).toContain('led: ♥ (1 on the table');
    expect(out).toContain('pile: ');
  });

  it('falls back to ? for an unknown led suit code', () => {
    expect(formatBhabhiState(state({ leadSuit: 9, pile: [] }))).toContain('led: ?');
  });

  // **順位は上がった順であって強さではない。** 残っている席と区別が付くこと。
  it('distinguishes seats that are out from seats still holding cards', () => {
    const out = formatBhabhiState(
      state({ players: [seat(0), seat(1, { rank: 1, cardCount: 0 }), seat(2), seat(3)] } as Partial<BhabhiResponse>),
    );
    expect(out).toContain('out, 1 to finish');
    expect(out).toContain('2 cards');
  });

  it('shows how often each seat has picked up', () => {
    expect(formatBhabhiState(state({ players: [seat(0, { pickups: 3 }), seat(1), seat(2), seat(3)] }))).toContain(
      '3 pickups',
    );
  });

  it('marks the playable cards in your hand', () => {
    expect(formatBhabhiState(state())).toMatch(/your hand: \[0\].*\*/);
  });

  it('says the hand is empty rather than printing nothing', () => {
    const out = formatBhabhiState(
      state({ players: [{ ...seat(0), cards: [] }, seat(1), seat(2), seat(3)] } as Partial<BhabhiResponse>),
    );
    expect(out).toContain('your hand: (empty)');
  });

  it('omits the hand block when no seat is human', () => {
    expect(formatBhabhiState(state({ players: [seat(1), seat(2), seat(3)] }))).not.toContain('your hand:');
  });

  // **直前の引き取りは盤面に痕跡が残らない。**
  it('reports the last pickup while the game runs', () => {
    const out = formatBhabhiState(state({ lastPickupIdx: 2, lastPickupSize: 5 }));
    expect(out).toContain('picked up 5 cards');
  });

  it('drops the pickup line once the game is over', () => {
    const out = formatBhabhiState(state({ lastPickupIdx: 2, lastPickupSize: 5, gameEndFlag: true, bhabhiIdx: 1 }));
    expect(out).not.toContain('picked up 5 cards');
  });

  // 終わり方は 2 通りあり、別の文言になる。
  it('names the Bhabhi on a normal finish', () => {
    const out = formatBhabhiState(state({ phase: 1, gameEndFlag: true, bhabhiIdx: 2 }));
    expect(out).toContain('is the Bhabhi');
    expect(out).not.toContain('deadlocked');
  });

  it('says so when the game was cut short as deadlocked', () => {
    const out = formatBhabhiState(
      state({ phase: 1, gameEndFlag: true, bhabhiIdx: 1, stalemate: true, trickNumber: 300 }),
    );
    expect(out).toContain('deadlocked after 300 tricks');
    expect(out).toContain('holds the most cards');
  });

  it('copes with an undecided Bhabhi', () => {
    expect(formatBhabhiState(state({ gameEndFlag: true, bhabhiIdx: -1 }))).toContain('? is the Bhabhi');
  });

  it('echoes a server message', () => {
    expect(formatBhabhiState(state({ message: 'must follow the led suit' }))).toContain('must follow the led suit');
  });
});
