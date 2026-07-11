import { describe, expect, it } from 'vitest';
import type { Card, ZhengResponse } from '../../types/card';
import { getZhengHint } from './zhengHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function player(id: number, isHuman: boolean, cards: Card[], over: Partial<ZhengResponse['players'][number]> = {}) {
  return { id, isHuman, isFinished: false, rank: 0, cardCount: cards.length, cards, ...over };
}

function fixture(overrides: Partial<ZhengResponse> = {}): ZhengResponse {
  return {
    players: [
      player(0, true, [card('SPADE', 3), card('HEART', 7)]),
      player(1, false, []),
      player(2, false, []),
      player(3, false, []),
    ],
    currentTurn: 0,
    tableCards: [],
    tablePlayType: 0,
    lastPlayPlayerIdx: -1,
    gameEndFlag: false,
    cpuActions: [],
    humanAction: null,
    config: { cpuDifficulty: 0 },
    message: '',
    ...overrides,
  };
}

describe('getZhengHint', () => {
  it('returns null after game end, when finished, or off-turn', () => {
    expect(getZhengHint(fixture({ gameEndFlag: true }))).toBeNull();
    expect(getZhengHint(fixture({ currentTurn: 2 }))).toBeNull();
    const finished = fixture();
    finished.players[0].isFinished = true;
    expect(getZhengHint(finished)).toBeNull();
  });

  it('suggests shedding weak cards on a lead', () => {
    const hint = getZhengHint(fixture());
    expect(hint).toEqual({ targetAction: 'play', reason: 'hint.playLow', confidence: 'moderate' });
  });

  it('suggests playing when a higher single exists (suits irrelevant)', () => {
    const state = fixture({ tableCards: [card('SPADE', 5)], tablePlayType: 1 });
    expect(getZhengHint(state)?.targetAction).toBe('play');
  });

  it('suggests passing when the same rank in another suit is all we have', () => {
    const state = fixture({
      players: [
        player(0, true, [card('HEART', 5), card('DIAMOND', 4)]),
        player(1, false, []),
        player(2, false, []),
        player(3, false, []),
      ],
      tableCards: [card('SPADE', 5)],
      tablePlayType: 1,
    });
    expect(getZhengHint(state)?.targetAction).toBe('pass');
  });

  it('treats a joker as a single that beats a 2', () => {
    const state = fixture({
      players: [player(0, true, [card('JOKER', 1)]), player(1, false, []), player(2, false, []), player(3, false, [])],
      tableCards: [card('SPADE', 2)],
      tablePlayType: 1,
    });
    expect(getZhengHint(state)?.targetAction).toBe('play');
  });

  it('never counts a joker toward a pair', () => {
    const state = fixture({
      players: [
        player(0, true, [card('JOKER', 1), card('JOKER', 2)]),
        player(1, false, []),
        player(2, false, []),
        player(3, false, []),
      ],
      tableCards: [card('SPADE', 13), card('HEART', 13)],
      tablePlayType: 2,
    });
    // Two jokers cannot pair — but they form the joker bomb, which beats a pair.
    expect(getZhengHint(state)?.targetAction).toBe('play');
  });

  it('suggests playing a bomb over any non-bomb play', () => {
    const state = fixture({
      players: [
        player(0, true, [card('SPADE', 4), card('HEART', 4), card('CLOVER', 4), card('DIAMOND', 4)]),
        player(1, false, []),
        player(2, false, []),
        player(3, false, []),
      ],
      tableCards: [card('SPADE', 2)],
      tablePlayType: 1,
    });
    expect(getZhengHint(state)?.targetAction).toBe('play');
  });

  it('requires a higher bomb to beat a bomb', () => {
    const weaker = fixture({
      players: [
        player(0, true, [card('SPADE', 4), card('HEART', 4), card('CLOVER', 4), card('DIAMOND', 4)]),
        player(1, false, []),
        player(2, false, []),
        player(3, false, []),
      ],
      tableCards: [card('SPADE', 9), card('HEART', 9), card('CLOVER', 9), card('DIAMOND', 9)],
      tablePlayType: 6,
    });
    expect(getZhengHint(weaker)?.targetAction).toBe('pass');
  });

  it('always passes against the joker bomb', () => {
    const state = fixture({
      players: [
        player(0, true, [card('SPADE', 4), card('HEART', 4), card('CLOVER', 4), card('DIAMOND', 4)]),
        player(1, false, []),
        player(2, false, []),
        player(3, false, []),
      ],
      tableCards: [card('JOKER', 1), card('JOKER', 2)],
      tablePlayType: 7,
    });
    expect(getZhengHint(state)?.targetAction).toBe('pass');
  });

  it('finds a higher straight of the same length', () => {
    const state = fixture({
      players: [
        player(0, true, [card('SPADE', 6), card('HEART', 7), card('CLOVER', 8)]),
        player(1, false, []),
        player(2, false, []),
        player(3, false, []),
      ],
      tableCards: [card('SPADE', 3), card('HEART', 4), card('CLOVER', 5)],
      tablePlayType: 4,
    });
    expect(getZhengHint(state)?.targetAction).toBe('play');
  });

  it('passes when only a shorter straight is available', () => {
    const state = fixture({
      players: [
        player(0, true, [card('SPADE', 6), card('HEART', 7), card('CLOVER', 8)]),
        player(1, false, []),
        player(2, false, []),
        player(3, false, []),
      ],
      tableCards: [card('SPADE', 3), card('HEART', 4), card('CLOVER', 5), card('DIAMOND', 6)],
      tablePlayType: 4,
    });
    expect(getZhengHint(state)?.targetAction).toBe('pass');
  });

  it('finds a higher pair run of the same length', () => {
    const state = fixture({
      players: [
        player(0, true, [
          card('SPADE', 7),
          card('HEART', 7),
          card('SPADE', 8),
          card('HEART', 8),
          card('SPADE', 9),
          card('HEART', 9),
        ]),
        player(1, false, []),
        player(2, false, []),
        player(3, false, []),
      ],
      tableCards: [
        card('SPADE', 4),
        card('HEART', 4),
        card('SPADE', 5),
        card('HEART', 5),
        card('SPADE', 6),
        card('HEART', 6),
      ],
      tablePlayType: 5,
    });
    expect(getZhengHint(state)?.targetAction).toBe('play');
  });
});
