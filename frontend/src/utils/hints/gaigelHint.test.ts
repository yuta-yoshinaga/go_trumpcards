import { describe, expect, it } from 'vitest';
import type { Card, GaigelResponse } from '../../types/card';
import { GaigelPhase } from '../../types/phases';
import { getGaigelHint } from './gaigelHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<GaigelResponse> = {}): GaigelResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 5,
        cards: [card('SPADE', 1), card('HEART', 10), card('CLOVER', 7)],
        team: 0,
        trickCount: 0,
      },
      { id: 1, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
      { id: 2, isHuman: false, cardCount: 5, cards: [], team: 0, trickCount: 0 },
      { id: 3, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
    ],
    phase: GaigelPhase.PLAY,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    dealerIdx: 3,
    trumpSuit: 1,
    stockRemaining: 20,
    isEndgame: false,
    currentTrick: [],
    teamScores: [0, 0],
    roundPoints: [0, 0],
    roundMarriage: [0, 0],
    marriageIndices: [],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 1, targetScore: 101 },
    ...overrides,
  };
}

describe('getGaigelHint', () => {
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getGaigelHint(state)).toBeNull();
  });

  it('returns null when human has no cards', () => {
    const state = makeState();
    state.players[0].cards = [];
    expect(getGaigelHint(state)).toBeNull();
  });

  it('returns null in TRICK_END phase', () => {
    expect(getGaigelHint(makeState({ phase: GaigelPhase.TRICK_END }))).toBeNull();
  });

  it('returns null in ROUND_END phase', () => {
    expect(getGaigelHint(makeState({ phase: GaigelPhase.ROUND_END }))).toBeNull();
  });

  it('returns null when not human turn', () => {
    expect(getGaigelHint(makeState({ currentPlayerIdx: 1 }))).toBeNull();
  });

  describe('marriage', () => {
    it('suggests marriage when leading with a declarable index', () => {
      const state = makeState({ currentTrick: [], marriageIndices: [0] });
      const hint = getGaigelHint(state);
      expect(hint?.targetAction).toBe('marriage');
      expect(hint?.reason).toBe('hint.marriage');
      expect(hint?.confidence).toBe('strong');
    });
  });

  describe('lead', () => {
    it('suggests lead trump when holding trump and trick empty', () => {
      const state = makeState({ trumpSuit: 1 });
      state.players[0].cards = [card('SPADE', 1), card('HEART', 7)];
      const hint = getGaigelHint(state);
      expect(hint?.reason).toBe('hint.leadTrump');
    });

    it('suggests lead value when no trump but a high-value card', () => {
      const state = makeState({ trumpSuit: 1 });
      state.players[0].cards = [card('HEART', 1), card('CLOVER', 7)];
      const hint = getGaigelHint(state);
      expect(hint?.reason).toBe('hint.leadValue');
    });

    it('suggests lead low when no trump and only low cards', () => {
      const state = makeState({ trumpSuit: 1 });
      state.players[0].cards = [card('HEART', 11), card('CLOVER', 7)];
      const hint = getGaigelHint(state);
      expect(hint?.reason).toBe('hint.leadLow');
    });
  });

  describe('follow', () => {
    it('suggests follow win when holding the led suit', () => {
      const state = makeState({
        trumpSuit: 1,
        currentTrick: [{ playerIdx: 3, card: card('CLOVER', 1) }],
      });
      state.players[0].cards = [card('CLOVER', 13), card('HEART', 10)];
      const hint = getGaigelHint(state);
      expect(hint?.reason).toBe('hint.followWin');
      expect(hint?.confidence).toBe('strong');
    });

    it('suggests cut when void of led suit but holding trump', () => {
      const state = makeState({
        trumpSuit: 1,
        currentTrick: [{ playerIdx: 3, card: card('HEART', 1) }],
      });
      state.players[0].cards = [card('SPADE', 13), card('CLOVER', 7)];
      const hint = getGaigelHint(state);
      expect(hint?.reason).toBe('hint.followCut');
    });

    it('suggests dump when void of led suit and no trump', () => {
      const state = makeState({
        trumpSuit: 1,
        currentTrick: [{ playerIdx: 3, card: card('HEART', 1) }],
      });
      state.players[0].cards = [card('CLOVER', 7), card('DIAMOND', 11)];
      const hint = getGaigelHint(state);
      expect(hint?.reason).toBe('hint.followDump');
    });
  });
});
