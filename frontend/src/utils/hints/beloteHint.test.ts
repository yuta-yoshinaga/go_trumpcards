import { describe, expect, it } from 'vitest';
import type { BeloteResponse, Card } from '../../types/card';
import { BelotePhase } from '../../types/phases';
import { getBeloteHint } from './beloteHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<BeloteResponse> = {}): BeloteResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 5,
        cards: [card('SPADE', 11), card('HEART', 10), card('CLOVER', 5)],
        team: 0,
        trickCount: 0,
      },
      { id: 1, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
      { id: 2, isHuman: false, cardCount: 5, cards: [], team: 0, trickCount: 0 },
      { id: 3, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
    ],
    phase: BelotePhase.PLAY,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    trumpSuit: 1,
    faceUpCard: card('SPADE', 11),
    makerTeam: 0,
    makerPlayerIdx: 0,
    currentTrick: [],
    teamScores: [0, 0],
    roundPoints: [0, 0],
    roundBeloteBonus: [0, 0],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 0, targetScore: 1000, dixDeDer: 10, enableBeloteRebelote: true },
    ...overrides,
  };
}

describe('getBeloteHint', () => {
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getBeloteHint(state)).toBeNull();
  });

  it('returns null when human has no cards', () => {
    const state = makeState();
    state.players[0].cards = [];
    expect(getBeloteHint(state)).toBeNull();
  });

  it('returns null in TRICK_END phase', () => {
    expect(getBeloteHint(makeState({ phase: BelotePhase.TRICK_END }))).toBeNull();
  });

  it('returns null in ROUND_END phase', () => {
    expect(getBeloteHint(makeState({ phase: BelotePhase.ROUND_END }))).toBeNull();
  });

  describe('BID_PICK_UP phase', () => {
    it('returns null when not human bid turn', () => {
      const state = makeState({ phase: BelotePhase.BID_PICK_UP, bidPlayerIdx: 1 });
      expect(getBeloteHint(state)).toBeNull();
    });

    it('suggests orderUp with strong trump-suit hand', () => {
      const state = makeState({
        phase: BelotePhase.BID_PICK_UP,
        bidPlayerIdx: 0,
        faceUpCard: card('SPADE', 11),
      });
      state.players[0].cards = [
        card('SPADE', 11), // J=14
        card('SPADE', 9), // 9=10
        card('HEART', 10),
      ];
      const hint = getBeloteHint(state);
      expect(hint?.targetAction).toBe('orderUp');
      expect(hint?.confidence).toBe('strong');
    });

    it('suggests pass with weak trump-suit hand', () => {
      const state = makeState({
        phase: BelotePhase.BID_PICK_UP,
        bidPlayerIdx: 0,
        faceUpCard: card('SPADE', 13),
      });
      state.players[0].cards = [card('CLOVER', 8), card('HEART', 7)];
      const hint = getBeloteHint(state);
      expect(hint?.targetAction).toBe('pass');
    });

    it('returns pass when faceUpCard is null', () => {
      const state = makeState({ phase: BelotePhase.BID_PICK_UP, bidPlayerIdx: 0, faceUpCard: null });
      const hint = getBeloteHint(state);
      expect(hint?.targetAction).toBe('pass');
    });
  });

  describe('BID_CALL_TRUMP phase', () => {
    it('returns null when not human bid turn', () => {
      const state = makeState({ phase: BelotePhase.BID_CALL_TRUMP, bidPlayerIdx: 1 });
      expect(getBeloteHint(state)).toBeNull();
    });

    it('suggests calltrump with strong suit', () => {
      const state = makeState({ phase: BelotePhase.BID_CALL_TRUMP, bidPlayerIdx: 0 });
      state.players[0].cards = [
        card('CLOVER', 11), // J=14
        card('CLOVER', 9), // 9=10
        card('CLOVER', 1), // A=7
      ];
      const hint = getBeloteHint(state);
      expect(hint?.targetAction).toBe('callTrump');
      expect(hint?.confidence).toBe('strong');
    });

    it('suggests pass when no suit reaches threshold', () => {
      const state = makeState({ phase: BelotePhase.BID_CALL_TRUMP, bidPlayerIdx: 0 });
      state.players[0].cards = [card('HEART', 8), card('CLOVER', 7), card('SPADE', 8)];
      const hint = getBeloteHint(state);
      expect(hint?.targetAction).toBe('pass');
    });
  });

  describe('PLAY phase', () => {
    it('returns null when not human turn', () => {
      const state = makeState({ phase: BelotePhase.PLAY, currentPlayerIdx: 1 });
      expect(getBeloteHint(state)).toBeNull();
    });

    it('suggests lead with trump when human has trump and trick empty', () => {
      const state = makeState({ phase: BelotePhase.PLAY, currentPlayerIdx: 0, trumpSuit: 1 });
      state.players[0].cards = [card('SPADE', 11), card('HEART', 10)];
      const hint = getBeloteHint(state);
      expect(hint?.targetAction).toBe('play');
      expect(hint?.reason).toBe('hint.leadTrump');
    });

    it('suggests lead off-suit when no trump and trick empty', () => {
      const state = makeState({ phase: BelotePhase.PLAY, currentPlayerIdx: 0, trumpSuit: 1 });
      state.players[0].cards = [card('HEART', 1), card('CLOVER', 13)];
      const hint = getBeloteHint(state);
      expect(hint?.reason).toBe('hint.leadOffSuit');
    });

    it('suggests follow suit when trick has led suit and human has it', () => {
      const state = makeState({
        phase: BelotePhase.PLAY,
        currentPlayerIdx: 0,
        trumpSuit: 1,
        currentTrick: [{ playerIdx: 3, card: card('CLOVER', 1) }],
      });
      state.players[0].cards = [card('CLOVER', 13), card('HEART', 10)];
      const hint = getBeloteHint(state);
      expect(hint?.reason).toBe('hint.followSuit');
      expect(hint?.confidence).toBe('strong');
    });

    it('suggests trump cut when void of led suit but has trump', () => {
      const state = makeState({
        phase: BelotePhase.PLAY,
        currentPlayerIdx: 0,
        trumpSuit: 1,
        currentTrick: [{ playerIdx: 3, card: card('HEART', 1) }],
      });
      state.players[0].cards = [card('SPADE', 9), card('CLOVER', 13)];
      const hint = getBeloteHint(state);
      expect(hint?.reason).toBe('hint.trumpCut');
    });

    it('suggests discard weakest when void of led and no trump', () => {
      const state = makeState({
        phase: BelotePhase.PLAY,
        currentPlayerIdx: 0,
        trumpSuit: 1,
        currentTrick: [{ playerIdx: 3, card: card('HEART', 1) }],
      });
      state.players[0].cards = [card('CLOVER', 8), card('DIAMOND', 7)];
      const hint = getBeloteHint(state);
      expect(hint?.reason).toBe('hint.discardWeakest');
    });
  });
});
