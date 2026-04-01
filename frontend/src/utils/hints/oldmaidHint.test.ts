import { describe, expect, it } from 'vitest';
import type { OldMaidResponse } from '../../types/card';
import { getOldMaidHint } from './oldmaidHint';

function makeState(overrides: Partial<OldMaidResponse> = {}): OldMaidResponse {
  return {
    players: [
      { id: 0, isHuman: true, isFinished: false, cardCount: 3, cards: [] },
      { id: 1, isHuman: false, isFinished: false, cardCount: 5, cards: [] },
      { id: 2, isHuman: false, isFinished: false, cardCount: 4, cards: [] },
    ],
    currentTurn: 0,
    nextDrawTargetIdx: 1,
    gameEndFlag: false,
    hasDrawn: false,
    lastDrawPlayerIdx: -1,
    lastDrawFromIdx: -1,
    lastDrawCard: null,
    lastDiscardedPairs: 0,
    cpuActions: [],
    drawHistory: [],
    cpuHighlightedCardIdx: -1,
    removedCard: null,
    mode: 0,
    message: '',
    ...overrides,
  };
}

describe('getOldMaidHint', () => {
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getOldMaidHint(state)).toBeNull();
  });

  it('returns null when game has ended', () => {
    expect(getOldMaidHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when it is not human turn', () => {
    expect(getOldMaidHint(makeState({ currentTurn: 1 }))).toBeNull();
  });

  it('suggests drawing from edges when target has many cards', () => {
    const state = makeState({ nextDrawTargetIdx: 1 });
    state.players[1].cardCount = 5;
    const result = getOldMaidHint(state);
    expect(result?.targetAction).toBe('draw');
    expect(result?.reason).toBe('hint.drawFromEdge');
    expect(result?.confidence).toBe('moderate');
  });

  it('suggests drawing randomly when target has few cards', () => {
    const state = makeState({ nextDrawTargetIdx: 1 });
    state.players[1].cardCount = 2;
    const result = getOldMaidHint(state);
    expect(result?.targetAction).toBe('draw');
    expect(result?.reason).toBe('hint.drawAny');
    expect(result?.confidence).toBe('moderate');
  });

  it('returns null when human is finished', () => {
    const state = makeState();
    state.players[0].isFinished = true;
    expect(getOldMaidHint(state)).toBeNull();
  });
});
