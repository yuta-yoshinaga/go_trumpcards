import { describe, expect, it } from 'vitest';
import type { Card, PrsiResponse } from '../../types/card';
import { PrsiPhase } from '../../types/phases';
import { getPrsiHint } from './prsiHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<PrsiResponse> = {}): PrsiResponse {
  return {
    players: [
      { id: 0, isHuman: true, cardCount: 3, cards: [card('HEART', 10), card('SPADE', 8), card('DIAMOND', 13)] },
      { id: 1, isHuman: false, cardCount: 3, cards: [] },
      { id: 2, isHuman: false, cardCount: 3, cards: [] },
      { id: 3, isHuman: false, cardCount: 3, cards: [] },
    ],
    phase: PrsiPhase.PLAY,
    currentPlayerIdx: 0,
    discardTop: card('HEART', 9),
    drawPileCount: 20,
    penaltyDrawCount: 0,
    pendingSkips: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: { cpuDifficulty: 1 },
    ...overrides,
  };
}

describe('getPrsiHint', () => {
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getPrsiHint(state)).toBeNull();
  });

  it('returns null when game has ended', () => {
    expect(getPrsiHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when it is not the human turn', () => {
    expect(getPrsiHint(makeState({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('returns null in the game end phase', () => {
    expect(getPrsiHint(makeState({ phase: PrsiPhase.GAME_END }))).toBeNull();
  });

  it('returns null when human has no cards', () => {
    const state = makeState();
    state.players[0].cards = [];
    expect(getPrsiHint(state)).toBeNull();
  });

  it('suggests playing a matching-suit card', () => {
    // discardTop ♥9; human has ♥10 (matching suit, plain card)
    const result = getPrsiHint(makeState());
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.playMatchingSuit');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests playing a matching-value card', () => {
    const state = makeState({ discardTop: card('CLOVER', 10) });
    // human has ♥10 (matching value)
    const result = getPrsiHint(state);
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.playMatchingValue');
  });

  it('suggests drawing when no card matches', () => {
    const state = makeState({ discardTop: card('CLOVER', 6) });
    state.players[0].cards = [card('HEART', 10), card('SPADE', 8), card('DIAMOND', 13)];
    const result = getPrsiHint(state);
    expect(result?.targetAction).toBe('draw');
    expect(result?.reason).toBe('hint.drawCard');
  });

  it('suggests stacking a 7 under an active penalty', () => {
    const state = makeState({ penaltyDrawCount: 2 });
    state.players[0].cards = [card('SPADE', 7), card('HEART', 10)];
    const result = getPrsiHint(state);
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.stackSeven');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests drawing the penalty when no 7 is available', () => {
    const state = makeState({ penaltyDrawCount: 4 });
    state.players[0].cards = [card('HEART', 10), card('SPADE', 8)];
    const result = getPrsiHint(state);
    expect(result?.targetAction).toBe('draw');
    expect(result?.reason).toBe('hint.drawPenalty');
  });

  it('prefers a plain card over an action card when both match', () => {
    const state = makeState({ discardTop: card('HEART', 1) });
    // matches: ♥10 (plain, suit) and ♠A (action, value). Prefer the plain card.
    state.players[0].cards = [card('SPADE', 1), card('HEART', 10)];
    const result = getPrsiHint(state);
    expect(result?.reason).toBe('hint.playMatchingSuit');
  });

  it('falls back to an action card when only action cards match', () => {
    const state = makeState({ discardTop: card('HEART', 7) });
    state.players[0].cards = [card('SPADE', 7), card('DIAMOND', 13)];
    const result = getPrsiHint(state);
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.playActionCard');
  });

  it('handles a null discard top with a moderate play hint', () => {
    const result = getPrsiHint(makeState({ discardTop: null }));
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.playMatchingSuit');
    expect(result?.confidence).toBe('moderate');
  });
});
