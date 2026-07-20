import { describe, expect, it } from 'vitest';
import type { Card, JassResponse } from '../../types/card';
import { JassPhase } from '../../types/phases';
import { getJassHint } from './jassHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<JassResponse> = {}): JassResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 9,
        cards: [card('SPADE', 11), card('HEART', 10), card('CLOVER', 6)],
        team: 0,
        trickCount: 0,
      },
      { id: 1, isHuman: false, cardCount: 9, cards: [], team: 1, trickCount: 0 },
      { id: 2, isHuman: false, cardCount: 9, cards: [], team: 0, trickCount: 0 },
      { id: 3, isHuman: false, cardCount: 9, cards: [], team: 1, trickCount: 0 },
    ],
    phase: JassPhase.PLAY,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    forehandIdx: 0,
    trumpSuit: 1,
    schieben: false,
    makerTeam: 0,
    makerPlayerIdx: 0,
    currentTrick: [],
    lastTrick: [],
    lastTrickWinner: -1,
    teamScores: [0, 0],
    roundPoints: [0, 0],
    roundWeisPoints: [0, 0],
    roundStockPoints: [0, 0],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 0, targetScore: 1000, lastTrickBonus: 5, enableWeis: true },
    ...overrides,
  };
}

describe('getJassHint', () => {
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getJassHint(state)).toBeNull();
  });

  it('returns null when human has no cards', () => {
    const state = makeState();
    state.players[0].cards = [];
    expect(getJassHint(state)).toBeNull();
  });

  it('returns null in TRICK_END phase', () => {
    expect(getJassHint(makeState({ phase: JassPhase.TRICK_END }))).toBeNull();
  });

  it('returns null in ROUND_END phase', () => {
    expect(getJassHint(makeState({ phase: JassPhase.ROUND_END }))).toBeNull();
  });

  describe('BID_TRUMP phase', () => {
    it('returns null when not human bid turn', () => {
      const state = makeState({ phase: JassPhase.BID_TRUMP, bidPlayerIdx: 1 });
      expect(getJassHint(state)).toBeNull();
    });

    it('suggests strategic trump with a strong suit', () => {
      const state = makeState({ phase: JassPhase.BID_TRUMP, bidPlayerIdx: 0 });
      state.players[0].cards = [
        card('CLOVER', 11), // J=14
        card('CLOVER', 9), // 9=10
        card('CLOVER', 1), // A=7
      ];
      const hint = getJassHint(state);
      expect(hint?.targetAction).toBe('callTrump');
      expect(hint?.reason).toBe('hint.strategicTrump');
      expect(hint?.confidence).toBe('strong');
    });

    it('suggests schieben with a weak hand', () => {
      const state = makeState({ phase: JassPhase.BID_TRUMP, bidPlayerIdx: 0 });
      state.players[0].cards = [card('HEART', 8), card('CLOVER', 7), card('SPADE', 6)];
      const hint = getJassHint(state);
      expect(hint?.targetAction).toBe('schieben');
      expect(hint?.reason).toBe('hint.schiebenRecommended');
    });
  });

  describe('BID_PARTNER phase', () => {
    it('returns null when not human bid turn', () => {
      const state = makeState({ phase: JassPhase.BID_PARTNER, bidPlayerIdx: 1 });
      expect(getJassHint(state)).toBeNull();
    });

    it('always suggests choosing a trump (no schieben)', () => {
      const state = makeState({ phase: JassPhase.BID_PARTNER, bidPlayerIdx: 0 });
      state.players[0].cards = [card('HEART', 8), card('CLOVER', 7), card('SPADE', 6)];
      const hint = getJassHint(state);
      expect(hint?.targetAction).toBe('callTrump');
      expect(hint?.reason).toBe('hint.strategicTrump');
    });
  });

  describe('PLAY phase', () => {
    it('returns null when not human turn', () => {
      const state = makeState({ phase: JassPhase.PLAY, currentPlayerIdx: 1 });
      expect(getJassHint(state)).toBeNull();
    });

    it('suggests lead with trump when human has trump and trick empty', () => {
      const state = makeState({ phase: JassPhase.PLAY, currentPlayerIdx: 0, trumpSuit: 1 });
      state.players[0].cards = [card('SPADE', 11), card('HEART', 10)];
      const hint = getJassHint(state);
      expect(hint?.targetAction).toBe('play');
      expect(hint?.reason).toBe('hint.leadTrump');
    });

    it('suggests lead strong when no trump and trick empty', () => {
      const state = makeState({ phase: JassPhase.PLAY, currentPlayerIdx: 0, trumpSuit: 1 });
      state.players[0].cards = [card('HEART', 1), card('CLOVER', 13)];
      const hint = getJassHint(state);
      expect(hint?.reason).toBe('hint.leadStrong');
    });

    it('suggests follow suit when trick has led suit and human has it', () => {
      const state = makeState({
        phase: JassPhase.PLAY,
        currentPlayerIdx: 0,
        trumpSuit: 1,
        currentTrick: [{ playerIdx: 3, card: card('CLOVER', 1) }],
      });
      state.players[0].cards = [card('CLOVER', 13), card('HEART', 10)];
      const hint = getJassHint(state);
      expect(hint?.reason).toBe('hint.followSuit');
      expect(hint?.confidence).toBe('strong');
    });

    it('suggests trump cut when void of led suit but has trump', () => {
      const state = makeState({
        phase: JassPhase.PLAY,
        currentPlayerIdx: 0,
        trumpSuit: 1,
        currentTrick: [{ playerIdx: 3, card: card('HEART', 1) }],
      });
      state.players[0].cards = [card('SPADE', 9), card('CLOVER', 13)];
      const hint = getJassHint(state);
      expect(hint?.reason).toBe('hint.trumpCut');
    });

    it('suggests discard weak when void of led and no trump', () => {
      const state = makeState({
        phase: JassPhase.PLAY,
        currentPlayerIdx: 0,
        trumpSuit: 1,
        currentTrick: [{ playerIdx: 3, card: card('HEART', 1) }],
      });
      state.players[0].cards = [card('CLOVER', 8), card('DIAMOND', 7)];
      const hint = getJassHint(state);
      expect(hint?.reason).toBe('hint.discardWeak');
    });
  });
});
