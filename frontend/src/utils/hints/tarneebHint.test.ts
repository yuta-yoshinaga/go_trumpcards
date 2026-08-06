import { describe, expect, it } from 'vitest';
import type { Card, TarneebResponse } from '../../types/card';
import { TarneebPhase } from '../../types/phases';
import { getTarneebHint } from './tarneebHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makePlayer(id: number, isHuman: boolean, team: number, cards: Card[] = []) {
  return {
    id,
    isHuman,
    team,
    cardCount: cards.length || 13,
    cards,
    bid: -1,
    roundScore: 0,
    cumulativeScore: 0,
    trickCount: 0,
  };
}

function makeState(overrides: Partial<TarneebResponse> = {}): TarneebResponse {
  const humanCards = overrides.players?.[0]?.cards ?? [card('SPADE', 13), card('HEART', 1), card('CLOVER', 5)];
  return {
    players: [
      makePlayer(0, true, 0, humanCards),
      makePlayer(1, false, 1),
      makePlayer(2, false, 0),
      makePlayer(3, false, 1),
    ],
    teamScores: [0, 0],
    phase: TarneebPhase.BID,
    roundNumber: 1,
    trickNumber: 0,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    bidWinnerIdx: -1,
    highestBid: 0,
    trumpSuit: 0,
    redealCount: 0,
    dealerIdx: 3,
    currentTrick: [],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: -1,
    validPlayIndices: [],
    message: '',
    config: { cpuDifficulty: 1, pointLimit: 31, minBid: 7 },
    ...overrides,
  };
}

describe('getTarneebHint', () => {
  it('returns null when human has no cards', () => {
    const state = makeState({
      players: [makePlayer(0, true, 0, []), makePlayer(1, false, 1), makePlayer(2, false, 0), makePlayer(3, false, 1)],
    });
    expect(getTarneebHint(state)).toBeNull();
  });

  it('returns null when not human bid turn', () => {
    const state = makeState({ bidPlayerIdx: 1 });
    expect(getTarneebHint(state)).toBeNull();
  });

  it('suggests pass when hand is weak', () => {
    const state = makeState({
      players: [
        makePlayer(0, true, 0, [card('SPADE', 2), card('CLOVER', 3), card('HEART', 4)]),
        makePlayer(1, false, 1),
        makePlayer(2, false, 0),
        makePlayer(3, false, 1),
      ],
    });
    const hint = getTarneebHint(state);
    expect(hint?.targetAction).toBe('bid:0');
  });

  it('suggests a bid when hand has many high cards', () => {
    const state = makeState({
      players: [
        makePlayer(0, true, 0, [
          card('SPADE', 1),
          card('SPADE', 13),
          card('SPADE', 12),
          card('SPADE', 11),
          card('SPADE', 10),
          card('HEART', 1),
          card('HEART', 13),
          card('DIAMOND', 12),
        ]),
        makePlayer(1, false, 1),
        makePlayer(2, false, 0),
        makePlayer(3, false, 1),
      ],
    });
    const hint = getTarneebHint(state);
    expect(hint?.targetAction).toMatch(/^bid:\d+$/);
  });

  it('suggests trump = longest suit when bid winner is human', () => {
    const state = makeState({
      phase: TarneebPhase.TRUMP_DECLARATION,
      bidWinnerIdx: 0,
      players: [
        makePlayer(0, true, 0, [
          card('HEART', 13),
          card('HEART', 12),
          card('HEART', 11),
          card('HEART', 10),
          card('HEART', 9),
          card('SPADE', 2),
        ]),
        makePlayer(1, false, 1),
        makePlayer(2, false, 0),
        makePlayer(3, false, 1),
      ],
    });
    const hint = getTarneebHint(state);
    expect(hint?.targetAction).toBe('trump:HEART');
  });

  it('returns null for trump phase if human is not bid winner', () => {
    const state = makeState({ phase: TarneebPhase.TRUMP_DECLARATION, bidWinnerIdx: 1 });
    expect(getTarneebHint(state)).toBeNull();
  });

  it('suggests follow-suit when in play phase with led suit available', () => {
    const state = makeState({
      phase: TarneebPhase.PLAY,
      trumpSuit: 1,
      currentTrick: [{ playerIdx: 3, card: card('HEART', 5) }],
      players: [
        makePlayer(0, true, 0, [card('HEART', 9), card('SPADE', 13)]),
        makePlayer(1, false, 1),
        makePlayer(2, false, 0),
        makePlayer(3, false, 1),
      ],
    });
    const hint = getTarneebHint(state);
    expect(hint?.reason).toBe('hint.followSuit');
  });

  it('suggests trump cut when void of led suit but has trump', () => {
    const state = makeState({
      phase: TarneebPhase.PLAY,
      trumpSuit: 1,
      currentTrick: [{ playerIdx: 3, card: card('HEART', 5) }],
      players: [
        makePlayer(0, true, 0, [card('CLOVER', 2), card('SPADE', 13)]),
        makePlayer(1, false, 1),
        makePlayer(2, false, 0),
        makePlayer(3, false, 1),
      ],
    });
    const hint = getTarneebHint(state);
    expect(hint?.reason).toBe('hint.trumpCut');
  });

  it('suggests discard when void of both led suit and trump', () => {
    const state = makeState({
      phase: TarneebPhase.PLAY,
      trumpSuit: 1,
      currentTrick: [{ playerIdx: 3, card: card('HEART', 5) }],
      players: [
        makePlayer(0, true, 0, [card('CLOVER', 2), card('DIAMOND', 4)]),
        makePlayer(1, false, 1),
        makePlayer(2, false, 0),
        makePlayer(3, false, 1),
      ],
    });
    const hint = getTarneebHint(state);
    expect(hint?.reason).toBe('hint.discardLowest');
  });

  it('returns null in unrecognised phase', () => {
    const state = makeState({ phase: TarneebPhase.GAME_END });
    expect(getTarneebHint(state)).toBeNull();
  });
});
