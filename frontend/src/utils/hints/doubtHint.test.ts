import { describe, expect, it } from 'vitest';
import type { Card, DoubtResponse } from '../../types/card';
import { DoubtPhase } from '../../types/phases';
import { getDoubtHint } from './doubtHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<DoubtResponse> = {}): DoubtResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        isFinished: false,
        cardCount: 5,
        cards: [card('HEART', 3), card('HEART', 7), card('SPADE', 3), card('DIAMOND', 10), card('CLOVER', 5)],
      },
      { id: 1, isHuman: false, isFinished: false, cardCount: 5, cards: [] },
      { id: 2, isHuman: false, isFinished: false, cardCount: 5, cards: [] },
    ],
    currentTurn: 0,
    phase: DoubtPhase.PLAY as 0 | 1 | 2,
    tableCardCount: 0,
    lastAction: null,
    cpuDoubters: [],
    cpuActions: [],
    humanAction: null,
    lastDoubtResult: null,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    doubtWindowSec: 5,
    penaltyDrawLimit: 0,
    ...overrides,
  };
}

describe('getDoubtHint', () => {
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getDoubtHint(state)).toBeNull();
  });

  it('returns null when game has ended', () => {
    expect(getDoubtHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when it is not human turn in play phase', () => {
    expect(getDoubtHint(makeState({ currentTurn: 1 }))).toBeNull();
  });

  it('suggests playing truthfully when hand has pairs of same value', () => {
    const state = makeState();
    // Human has two 3s
    const result = getDoubtHint(state);
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.playTruth');
    expect(result?.confidence).toBe('moderate');
  });

  it('suggests bluffing when no pairs available', () => {
    const state = makeState();
    state.players[0].cards = [
      card('HEART', 1),
      card('SPADE', 2),
      card('DIAMOND', 4),
      card('CLOVER', 6),
      card('HEART', 8),
    ];
    const result = getDoubtHint(state);
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.bluffCarefully');
    expect(result?.confidence).toBe('moderate');
  });

  it('suggests doubting when last action has tell in doubt phase', () => {
    const state = makeState({
      phase: DoubtPhase.DOUBT as 0 | 1 | 2,
      lastAction: { playerIdx: 1, claimedValue: 5, cardCount: 2, isBluff: true, hasTell: true },
    });
    const result = getDoubtHint(state);
    expect(result?.targetAction).toBe('doubt');
    expect(result?.reason).toBe('hint.doubtTell');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests skipping when no tell in doubt phase', () => {
    const state = makeState({
      phase: DoubtPhase.DOUBT as 0 | 1 | 2,
      lastAction: { playerIdx: 1, claimedValue: 5, cardCount: 2, isBluff: false },
    });
    const result = getDoubtHint(state);
    expect(result?.targetAction).toBe('skip');
    expect(result?.reason).toBe('hint.skipSafe');
    expect(result?.confidence).toBe('moderate');
  });

  it('returns null in doubt phase when no last action', () => {
    const state = makeState({
      phase: DoubtPhase.DOUBT as 0 | 1 | 2,
      lastAction: null,
    });
    expect(getDoubtHint(state)).toBeNull();
  });

  it('returns null when human is finished', () => {
    const state = makeState();
    state.players[0].isFinished = true;
    expect(getDoubtHint(state)).toBeNull();
  });
});
