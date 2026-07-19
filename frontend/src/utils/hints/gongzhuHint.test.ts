import { describe, expect, it } from 'vitest';
import type { Card, GongZhuResponse } from '../../types/card';
import { GongZhuPhase } from '../../types/phases';
import { getGongZhuHint } from './gongzhuHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<GongZhuResponse> = {}): GongZhuResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 3,
        cards: [card('CLOVER', 3), card('DIAMOND', 5), card('HEART', 10)],
        capturedPointCards: [],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 3,
        cards: [],
        capturedPointCards: [],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
      {
        id: 2,
        isHuman: false,
        cardCount: 3,
        cards: [],
        capturedPointCards: [],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
      {
        id: 3,
        isHuman: false,
        cardCount: 3,
        cards: [],
        capturedPointCards: [],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
    ],
    phase: GongZhuPhase.PLAY,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    currentTrick: [],
    heartsBroken: false,
    exposed: { pig: false, sheep: false, ace: false, doubler: false },
    exposableIndices: [],
    gameEndFlag: false,
    winnerIdx: -1,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 0, pointLimit: 1000 },
    ...overrides,
  };
}

describe('getGongZhuHint', () => {
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getGongZhuHint(state)).toBeNull();
  });

  it('returns null when human has no cards', () => {
    const state = makeState();
    state.players[0].cards = [];
    expect(getGongZhuHint(state)).toBeNull();
  });

  it('returns null in TRICK_END phase', () => {
    expect(getGongZhuHint(makeState({ phase: GongZhuPhase.TRICK_END }))).toBeNull();
  });

  it('returns null when not current player turn', () => {
    expect(getGongZhuHint(makeState({ currentPlayerIdx: 2 }))).toBeNull();
  });

  // Expose phase
  it('suggests exposing the sheep when held', () => {
    const state = makeState({ phase: GongZhuPhase.EXPOSE });
    state.players[0].cards = [card('DIAMOND', 11), card('CLOVER', 3)];
    expect(getGongZhuHint(state)?.reason).toBe('hint.exposeSheep');
  });

  it('suggests exposing nothing when no sheep', () => {
    const state = makeState({ phase: GongZhuPhase.EXPOSE });
    state.players[0].cards = [card('CLOVER', 3), card('DIAMOND', 5)];
    expect(getGongZhuHint(state)?.reason).toBe('hint.exposeNone');
  });

  // Play phase - leading
  it('suggests leading lowest when leading', () => {
    expect(getGongZhuHint(makeState({ currentTrick: [] }))?.reason).toBe('hint.leadLowest');
  });

  // Play phase - following
  it('suggests following suit when possible', () => {
    const state = makeState({ currentTrick: [{ playerIdx: 1, card: card('CLOVER', 7) }] });
    expect(getGongZhuHint(state)?.reason).toBe('hint.followSuit');
  });

  it('suggests chasing the sheep when the sheep is in the trick and can follow', () => {
    const state = makeState({
      currentTrick: [
        { playerIdx: 1, card: card('DIAMOND', 4) },
        { playerIdx: 2, card: card('DIAMOND', 11) },
      ],
    });
    state.players[0].cards = [card('DIAMOND', 13), card('HEART', 3)];
    expect(getGongZhuHint(state)?.reason).toBe('hint.chaseSheep');
  });

  it('suggests dumping the pig when void in led suit', () => {
    const state = makeState({ currentTrick: [{ playerIdx: 1, card: card('CLOVER', 7) }] });
    state.players[0].cards = [card('SPADE', 12), card('HEART', 3), card('DIAMOND', 5)];
    expect(getGongZhuHint(state)?.reason).toBe('hint.dumpPig');
  });

  it('suggests dumping hearts when void and no pig', () => {
    const state = makeState({ currentTrick: [{ playerIdx: 1, card: card('CLOVER', 7) }] });
    state.players[0].cards = [card('HEART', 13), card('DIAMOND', 5), card('SPADE', 3)];
    expect(getGongZhuHint(state)?.reason).toBe('hint.dumpHearts');
  });

  it('suggests playing highest when void with no penalty cards', () => {
    const state = makeState({ currentTrick: [{ playerIdx: 1, card: card('CLOVER', 7) }] });
    state.players[0].cards = [card('DIAMOND', 13), card('SPADE', 3)];
    expect(getGongZhuHint(state)?.reason).toBe('hint.playHighest');
  });
});
