import { describe, expect, it } from 'vitest';
import type { Card, RollingStoneResponse } from '../../../types/card';
import { formatRollingStoneState } from './rollingstoneFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 8,
  cards: id === 0 ? [card('SPADE', 9), card('HEART', 10)] : [],
  pickups: 0,
  finishedAt: 0,
  ...over,
});

const state = (over: Partial<RollingStoneResponse> = {}): RollingStoneResponse =>
  ({
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    mustPickUp: false,
    validPlays: [0],
    currentTrick: [],
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    trickNumber: 2,
    lastPickupIdx: -1,
    finishedCnt: 0,
    deckSize: 32,
    discarded: 8,
    gameEndFlag: false,
    winnerIdx: -1,
    config: { playerCnt: 4 },
    message: '',
    ...over,
  }) as unknown as RollingStoneResponse;

describe('formatRollingStoneState', () => {
  it('reports loading for a null state', () => {
    expect(formatRollingStoneState(null)).toBe('Loading...');
  });

  // **デッキ枚数は人数で変わり、抜けた枚数も出す。**
  it('shows the trick, deck size and what is still in play', () => {
    const out = formatRollingStoneState(state());
    expect(out).toContain('trick 3');
    expect(out).toContain('deck 32');
    expect(out).toContain('24 still in play');
    expect(out).toContain('PLAY');
  });

  // **勝利条件が逆さまなのが規則そのもの。**
  it('states the inverted goal every time', () => {
    const out = formatRollingStoneState(state());
    expect(out).toContain('scores nothing');
    expect(out).toContain('run out of cards first');
  });

  it('shows every hand size and pickup count', () => {
    const out = formatRollingStoneState(state({ players: [seat(0, { cardCount: 11, pickups: 2 })] }));
    expect(out).toContain('11 cards, picked up 2x');
  });

  // **引き取った席と上がった席は盤面に痕跡が残らない。**
  it('marks the last pickup and finishers', () => {
    expect(formatRollingStoneState(state())).not.toContain('just picked up');
    expect(formatRollingStoneState(state({ lastPickupIdx: 2 }))).toContain('just picked up');
    expect(formatRollingStoneState(state({ players: [seat(0, { finishedAt: 1, cardCount: 0 })] }))).toContain(
      '[out in 1]',
    );
  });

  it('marks the seat on turn and the playable cards', () => {
    const out = formatRollingStoneState(state());
    expect(out).toMatch(/^>\S/m);
    expect(out).toMatch(/\[0\]\S+\*/);
    expect(out).not.toMatch(/\[1\]\S+\*/);
  });

  // **出せる札が無いことははっきり言う。**
  it('says so when a pickup is forced', () => {
    expect(formatRollingStoneState(state())).not.toContain('cannot follow');
    const out = formatRollingStoneState(
      state({
        mustPickUp: true,
        validPlays: [],
        currentTrick: [{ playerIdx: 1, card: card('CLOVER', 9) }],
        leadSuit: 2,
      } as Partial<RollingStoneResponse>),
    );
    expect(out).toContain('cannot follow');
    expect(out).toContain('1 cards');
    // **追従できなかったスートまで言う** (#5764)。
    expect(out).toContain('cannot follow ♣');
  });

  it('shows the current trick when there is one', () => {
    const out = formatRollingStoneState(
      state({ currentTrick: [{ playerIdx: 1, card: card('SPADE', 13) }] } as Partial<RollingStoneResponse>),
    );
    expect(out).toContain('trick:');
  });

  // **上限で切った局は「上がった」わけではない。** 言い分ける。
  it('distinguishes running out from the stalemate', () => {
    const ranOut = formatRollingStoneState(
      state({ phase: 1, gameEndFlag: true, winnerIdx: 0, players: [seat(0, { cardCount: 0, finishedAt: 1 })] }),
    );
    expect(ranOut).toContain('ran out first');

    const stale = formatRollingStoneState(
      state({ phase: 1, gameEndFlag: true, winnerIdx: 0, players: [seat(0, { cardCount: 3 })] }),
    );
    expect(stale).toContain('nobody ran out');
    expect(stale).toContain('fewest (3)');
  });

  it('shows the server message', () => {
    expect(formatRollingStoneState(state({ message: 'boom' }))).toContain('boom');
  });
});
