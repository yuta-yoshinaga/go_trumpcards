import { describe, expect, it } from 'vitest';
import type { Card, SevensResponse } from '../../types/card';
import { getSevensHint } from './sevensHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<SevensResponse> = {}): SevensResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        isFinished: false,
        rank: 0,
        cardCount: 5,
        passesUsed: 0,
        maxPasses: 3,
        cards: [card('HEART', 6), card('SPADE', 8), card('DIAMOND', 4), card('CLOVER', 2), card('HEART', 10)],
        lastPlayedJoker: false,
      },
      {
        id: 1,
        isHuman: false,
        isFinished: false,
        rank: 0,
        cardCount: 5,
        passesUsed: 0,
        maxPasses: 3,
        cards: [],
        lastPlayedJoker: false,
      },
    ],
    currentTurn: 0,
    // 7 is placed for each suit initially; min=7, max=7
    tableMinVals: [0, 7, 7, 7, 7],
    tableMaxVals: [0, 7, 7, 7, 7],
    tablePlaced: [0, 1 << 7, 1 << 7, 1 << 7, 1 << 7],
    config: {
      tunnelEnabled: false,
      tunnelSkipWidth: 0,
      jokerCount: 0,
      cpuStrategy: 0,
      maxPasses: 3,
      noJokerFinish: false,
      jokerReclaimEnabled: false,
      endStopEnabled: false,
      jokerConsecutiveBanned: false,
    },
    gameEndFlag: false,
    cpuActions: [],
    humanAction: null,
    message: '',
    ...overrides,
  };
}

describe('getSevensHint', () => {
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getSevensHint(state)).toBeNull();
  });

  it('returns null when game has ended', () => {
    expect(getSevensHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when not human turn', () => {
    expect(getSevensHint(makeState({ currentTurn: 1 }))).toBeNull();
  });

  it('suggests playing a card that extends the range', () => {
    const state = makeState();
    // HEART min=7, max=7 → can play 6 or 8
    // Human has HEART 6 → playable
    const result = getSevensHint(state);
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.playExtend');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests pass when no playable cards', () => {
    const state = makeState();
    // No cards adjacent to any range
    state.players[0].cards = [card('HEART', 2), card('SPADE', 12)];
    const result = getSevensHint(state);
    expect(result?.targetAction).toBe('pass');
    expect(result?.reason).toBe('hint.passAvailable');
    expect(result?.confidence).toBe('moderate');
  });

  it('warns about pass limit when passes are running low', () => {
    const state = makeState();
    state.players[0].cards = [card('HEART', 2), card('SPADE', 12)];
    state.players[0].passesUsed = 2;
    state.players[0].maxPasses = 3;
    const result = getSevensHint(state);
    expect(result?.targetAction).toBe('pass');
    expect(result?.reason).toBe('hint.passLimitWarning');
    expect(result?.confidence).toBe('strong');
  });

  it('returns null when human is finished', () => {
    const state = makeState();
    state.players[0].isFinished = true;
    expect(getSevensHint(state)).toBeNull();
  });

  it('detects playable card with tunnel rule (K wraps to A)', () => {
    const state = makeState();
    state.config.tunnelEnabled = true;
    state.tableMinVals = [0, 1, 7, 7, 7];
    state.tableMaxVals = [0, 13, 7, 7, 7];
    // SPADE min=1, max=13 → with tunnel, K(13) wraps to A(1) side
    state.players[0].cards = [card('DIAMOND', 2)];
    // DIAMOND min=7, max=7 → 2 is not adjacent, no tunnel match either (min≠1, max≠13)
    const result = getSevensHint(state);
    expect(result?.targetAction).toBe('pass');
  });

  it('detects playable card with tunnel rule (A side)', () => {
    const state = makeState();
    state.config.tunnelEnabled = true;
    state.tableMinVals = [0, 1, 7, 7, 7];
    state.tableMaxVals = [0, 13, 7, 7, 7];
    state.players[0].cards = [card('SPADE', 13)];
    // SPADE min=1 and card=13 with tunnel → playable
    const result = getSevensHint(state);
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.playExtend');
  });
});
