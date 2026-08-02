import { describe, expect, it } from 'vitest';
import type { Card, CrazyEightsResponse } from '../../types/card';
import { CrazyEightsPhase } from '../../types/phases';
import { getCrazyEightsHint } from './crazyeightsHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<CrazyEightsResponse> = {}): CrazyEightsResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 3,
        cards: [card('HEART', 5), card('SPADE', 10), card('DIAMOND', 3)],
        roundScore: 0,
        cumulativeScore: 0,
      },
      { id: 1, isHuman: false, cardCount: 3, cards: [], roundScore: 0, cumulativeScore: 0 },
    ],
    phase: CrazyEightsPhase.PLAY,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: card('HEART', 7),
    drawPileCount: 20,
    chosenSuit: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: { cpuDifficulty: 0, pointLimit: 100 },
    ...overrides,
  };
}

describe('getCrazyEightsHint', () => {
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getCrazyEightsHint(state)).toBeNull();
  });

  it('returns null when game has ended', () => {
    expect(getCrazyEightsHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when it is not human turn', () => {
    expect(getCrazyEightsHint(makeState({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('returns null in round end phase', () => {
    expect(getCrazyEightsHint(makeState({ phase: CrazyEightsPhase.ROUND_END }))).toBeNull();
  });

  it('suggests playing matching suit card', () => {
    const state = makeState();
    // discardTop is HEART 7, human has HEART 5 (matching suit)
    const result = getCrazyEightsHint(state);
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.playMatchingSuit');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests playing matching value card', () => {
    const state = makeState({ discardTop: card('CLOVER', 10) });
    // human has SPADE 10 (matching value)
    const result = getCrazyEightsHint(state);
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.playMatchingValue');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests saving eight when other plays available', () => {
    const state = makeState();
    state.players[0].cards = [card('HEART', 8), card('HEART', 5), card('SPADE', 2)];
    const result = getCrazyEightsHint(state);
    expect(result?.reason).toBe('hint.saveEight');
    expect(result?.confidence).toBe('moderate');
  });

  it('suggests playing eight when only option', () => {
    const state = makeState();
    state.players[0].cards = [card('HEART', 8), card('SPADE', 2), card('DIAMOND', 4)];
    // discardTop is HEART 7, only HEART 8 matches
    const result = getCrazyEightsHint(state);
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.playEight');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests drawing when no playable cards', () => {
    const state = makeState({ discardTop: card('CLOVER', 7) });
    state.players[0].cards = [card('HEART', 5), card('SPADE', 10), card('DIAMOND', 3)];
    const result = getCrazyEightsHint(state);
    expect(result?.targetAction).toBe('draw');
    expect(result?.reason).toBe('hint.drawCard');
    expect(result?.confidence).toBe('moderate');
  });

  it('suggests best suit in choose suit phase with strong confidence', () => {
    const state = makeState({ phase: CrazyEightsPhase.CHOOSE_SUIT });
    state.players[0].cards = [card('HEART', 5), card('HEART', 9), card('SPADE', 2)];
    const result = getCrazyEightsHint(state);
    expect(result?.targetAction).toBe('chooseSuit');
    expect(result?.reason).toBe('hint.chooseMostSuit');
    expect(result?.confidence).toBe('strong');
  });

  it('returns moderate confidence in choose suit phase when only 8s remain', () => {
    const state = makeState({ phase: CrazyEightsPhase.CHOOSE_SUIT });
    state.players[0].cards = [card('HEART', 8), card('SPADE', 8)];
    const result = getCrazyEightsHint(state);
    expect(result?.targetAction).toBe('chooseSuit');
    expect(result?.confidence).toBe('moderate');
  });

  it('handles chosen suit override in play phase', () => {
    const state = makeState({ discardTop: card('CLOVER', 8), chosenSuit: 1 });
    // chosenSuit 1 = SPADE, human has SPADE 10
    state.players[0].cards = [card('SPADE', 10), card('DIAMOND', 3)];
    const result = getCrazyEightsHint(state);
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.playMatchingSuit');
    expect(result?.confidence).toBe('strong');
  });
});

// **8 の後は指定スートだけが通る。**ドメインは chosenSuit > 0 のとき
// `card.GetDesign() == g.chosenSuit` だけを見て、場札のランクは見ない。
// ランク一致を門番していないと、出せない札を strong で勧める (#4598)。
describe('getCrazyEightsHint with a called suit', () => {
  const calledDiamond = {
    chosenSuit: 4,
    discardTop: card('HEART', 9),
  };

  it('does not offer a rank match that the called suit forbids', () => {
    const state = makeState({
      ...calledDiamond,
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 2,
          cards: [card('SPADE', 9), card('CLOVER', 2)],
          roundScore: 0,
          cumulativeScore: 0,
        },
        { id: 1, isHuman: false, cardCount: 3, cards: [], roundScore: 0, cumulativeScore: 0 },
      ],
    });
    expect(getCrazyEightsHint(state)).toEqual({
      targetAction: 'draw',
      reason: 'hint.drawCard',
      confidence: 'moderate',
    });
  });

  it('does not tell the player to save the only card they can play', () => {
    const state = makeState({
      ...calledDiamond,
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 2,
          cards: [card('SPADE', 9), card('CLOVER', 8)],
          roundScore: 0,
          cumulativeScore: 0,
        },
        { id: 1, isHuman: false, cardCount: 3, cards: [], roundScore: 0, cumulativeScore: 0 },
      ],
    });
    expect(getCrazyEightsHint(state)?.reason).toBe('hint.playEight');
  });

  it('still offers a card of the called suit', () => {
    const state = makeState({
      ...calledDiamond,
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 2,
          cards: [card('DIAMOND', 2), card('SPADE', 3)],
          roundScore: 0,
          cumulativeScore: 0,
        },
        { id: 1, isHuman: false, cardCount: 3, cards: [], roundScore: 0, cumulativeScore: 0 },
      ],
    });
    expect(getCrazyEightsHint(state)?.reason).toBe('hint.playMatchingSuit');
  });
});
