import { describe, expect, it } from 'vitest';
import type { Card, MacauResponse } from '../../types/card';
import { MacauPhase } from '../../types/phases';
import { getMacauHint } from './macauHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<MacauResponse> = {}): MacauResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 3,
        cards: [card('HEART', 5), card('SPADE', 10), card('DIAMOND', 3)],
        roundScore: 0,
        cumulativeScore: 0,
        hasDeclared: false,
      },
      { id: 1, isHuman: false, cardCount: 3, cards: [], roundScore: 0, cumulativeScore: 0, hasDeclared: false },
    ],
    phase: MacauPhase.PLAY,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: card('HEART', 7),
    drawPileCount: 20,
    chosenSuit: 0,
    penaltyDrawCount: 0,
    playableIndices: [],
    direction: 1,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: { cpuDifficulty: 0, pointLimit: 100 },
    ...overrides,
  };
}

describe('getMacauHint', () => {
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getMacauHint(state)).toBeNull();
  });

  it('returns null when game has ended', () => {
    expect(getMacauHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when not human turn', () => {
    expect(getMacauHint(makeState({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('returns null in round end phase', () => {
    expect(getMacauHint(makeState({ phase: MacauPhase.ROUND_END }))).toBeNull();
  });

  it('suggests declaring in must-declare phase', () => {
    const result = getMacauHint(makeState({ phase: MacauPhase.MUST_DECLARE }));
    expect(result?.reason).toBe('hint.declareMacau');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests stacking a 2 during a penalty when holding a 2', () => {
    const state = makeState({ penaltyDrawCount: 2 });
    state.players[0].cards = [card('SPADE', 2), card('HEART', 5)];
    const result = getMacauHint(state);
    expect(result?.targetAction).toBe('play');
    expect(result?.reason).toBe('hint.stackTwo');
  });

  it('suggests taking the penalty when no 2 to stack', () => {
    const state = makeState({ penaltyDrawCount: 2 });
    state.players[0].cards = [card('HEART', 5), card('SPADE', 10)];
    const result = getMacauHint(state);
    expect(result?.targetAction).toBe('draw');
    expect(result?.reason).toBe('hint.takePenalty');
  });

  it('suggests playing matching suit card', () => {
    const result = getMacauHint(makeState());
    expect(result?.reason).toBe('hint.playMatchingSuit');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests playing matching value card', () => {
    const result = getMacauHint(makeState({ discardTop: card('CLOVER', 10) }));
    expect(result?.reason).toBe('hint.playMatchingValue');
  });

  it('suggests saving eight when another play is available', () => {
    const state = makeState();
    state.players[0].cards = [card('HEART', 8), card('HEART', 5), card('SPADE', 2)];
    expect(getMacauHint(state)?.reason).toBe('hint.saveEight');
  });

  it('suggests playing eight when only option', () => {
    const state = makeState();
    state.players[0].cards = [card('HEART', 8), card('SPADE', 2), card('DIAMOND', 4)];
    expect(getMacauHint(state)?.reason).toBe('hint.playEight');
  });

  it('suggests drawing when no playable cards', () => {
    const state = makeState({ discardTop: card('CLOVER', 7) });
    state.players[0].cards = [card('HEART', 5), card('SPADE', 10), card('DIAMOND', 3)];
    expect(getMacauHint(state)?.reason).toBe('hint.drawCard');
  });

  it('suggests best suit in choose suit phase', () => {
    const state = makeState({ phase: MacauPhase.CHOOSE_SUIT });
    state.players[0].cards = [card('HEART', 5), card('HEART', 9), card('SPADE', 2)];
    const result = getMacauHint(state);
    expect(result?.targetAction).toBe('chooseSuit');
    expect(result?.confidence).toBe('strong');
  });

  it('moderate confidence in choose suit when only 8s remain', () => {
    const state = makeState({ phase: MacauPhase.CHOOSE_SUIT });
    state.players[0].cards = [card('HEART', 8), card('SPADE', 8)];
    expect(getMacauHint(state)?.confidence).toBe('moderate');
  });

  it('handles chosen suit override in play phase', () => {
    const state = makeState({ discardTop: card('CLOVER', 8), chosenSuit: 1 });
    state.players[0].cards = [card('SPADE', 10), card('DIAMOND', 3)];
    expect(getMacauHint(state)?.reason).toBe('hint.playMatchingSuit');
  });
});

// **8 の後は指定スートだけが通る。**ドメインは chosenSuit > 0 のとき
// `card.GetDesign() == g.chosenSuit` だけを見て、場札のランクは見ない。
// ランク一致を門番していないと、出せない札を strong で勧める (#4598)。
describe('getMacauHint with a called suit', () => {
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
          hasDeclared: false,
        },
        { id: 1, isHuman: false, cardCount: 3, cards: [], roundScore: 0, cumulativeScore: 0, hasDeclared: false },
      ],
    });
    expect(getMacauHint(state)).toEqual({ targetAction: 'draw', reason: 'hint.drawCard', confidence: 'moderate' });
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
          hasDeclared: false,
        },
        { id: 1, isHuman: false, cardCount: 3, cards: [], roundScore: 0, cumulativeScore: 0, hasDeclared: false },
      ],
    });
    expect(getMacauHint(state)?.reason).toBe('hint.playEight');
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
          hasDeclared: false,
        },
        { id: 1, isHuman: false, cardCount: 3, cards: [], roundScore: 0, cumulativeScore: 0, hasDeclared: false },
      ],
    });
    expect(getMacauHint(state)?.reason).toBe('hint.playMatchingSuit');
  });
});
