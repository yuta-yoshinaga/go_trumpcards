import { describe, expect, it } from 'vitest';
import type { Card, OmiResponse } from '../../types/card';
import { OmiPhase } from '../../types/phases';
import { getOmiHint } from './omiHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<OmiResponse> = {}): OmiResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 8,
        cards: [card('SPADE', 14), card('HEART', 10), card('CLOVER', 5)],
        team: 0,
        trickCount: 0,
      },
      { id: 1, isHuman: false, cardCount: 8, cards: [], team: 1, trickCount: 0 },
      { id: 2, isHuman: false, cardCount: 8, cards: [], team: 0, trickCount: 0 },
      { id: 3, isHuman: false, cardCount: 8, cards: [], team: 1, trickCount: 0 },
    ],
    phase: OmiPhase.PLAY,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    trumpCallerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    trumpSuit: 1,
    dealStage: 2,
    faceUpCard: null,
    makerTeam: 0,
    goingAlone: false,
    goingAlonePlayerIdx: -1,
    currentTrick: [],
    teamScores: [0, 0],
    teamTricks: [0, 0],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 0, pointLimit: 10 },
    ...overrides,
  };
}

describe('getOmiHint', () => {
  // Null/guard conditions
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getOmiHint(state)).toBeNull();
  });

  it('returns null when human has no cards', () => {
    const state = makeState();
    state.players[0].cards = [];
    expect(getOmiHint(state)).toBeNull();
  });

  it('returns null in TRICK_END phase', () => {
    expect(getOmiHint(makeState({ phase: OmiPhase.TRICK_END }))).toBeNull();
  });

  it('returns null in ROUND_END phase', () => {
    expect(getOmiHint(makeState({ phase: OmiPhase.ROUND_END }))).toBeNull();
  });

  it('returns null in GAME_END phase', () => {
    expect(getOmiHint(makeState({ phase: OmiPhase.GAME_END }))).toBeNull();
  });

  // Omi has no PICK_UP or DISCARD phases — verify
  it('returns null for any unrecognised phase (no PICK_UP/DISCARD)', () => {
    // Phase 10/11 were Euchre placeholders, now removed
    expect(getOmiHint(makeState({ phase: 10 }))).toBeNull();
    expect(getOmiHint(makeState({ phase: 11 }))).toBeNull();
  });

  // Call trump phase
  it('returns null in CALL_TRUMP phase when not human bid turn', () => {
    expect(getOmiHint(makeState({ phase: OmiPhase.CALL_TRUMP, bidPlayerIdx: 2 }))).toBeNull();
  });

  it('suggests calling suit with most cards (strong if 3+)', () => {
    const state = makeState({ phase: OmiPhase.CALL_TRUMP, bidPlayerIdx: 0 });
    state.players[0].cards = [
      card('HEART', 14),
      card('HEART', 13),
      card('HEART', 10),
      card('CLOVER', 5),
      card('DIAMOND', 3),
    ];
    const result = getOmiHint(state);
    expect(result?.reason).toBe('hint.callStrongSuit');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests calling suit with moderate confidence if count < 3', () => {
    const state = makeState({ phase: OmiPhase.CALL_TRUMP, bidPlayerIdx: 0 });
    state.players[0].cards = [
      card('HEART', 14),
      card('HEART', 10),
      card('CLOVER', 5),
      card('DIAMOND', 3),
      card('SPADE', 2),
    ];
    const result = getOmiHint(state);
    expect(result?.reason).toBe('hint.callStrongSuit');
    expect(result?.confidence).toBe('moderate');
  });

  // Play phase - leading
  it('returns null in PLAY phase when not human turn', () => {
    expect(getOmiHint(makeState({ phase: OmiPhase.PLAY, currentPlayerIdx: 2 }))).toBeNull();
  });

  it('suggests leading with trump when have many trump cards', () => {
    const state = makeState({ phase: OmiPhase.PLAY, currentPlayerIdx: 0, trumpSuit: 1 });
    state.players[0].cards = [card('SPADE', 14), card('SPADE', 12), card('HEART', 5)];
    const result = getOmiHint(state);
    expect(result?.reason).toBe('hint.leadTrump');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests leading off-suit when few trump cards', () => {
    const state = makeState({ phase: OmiPhase.PLAY, currentPlayerIdx: 0, trumpSuit: 1 });
    state.players[0].cards = [card('HEART', 5), card('DIAMOND', 7), card('CLOVER', 3)];
    const result = getOmiHint(state);
    expect(result?.reason).toBe('hint.leadOffSuit');
    expect(result?.confidence).toBe('moderate');
  });

  // Play phase - following (Omi: must follow suit if possible, any card if void)
  it('suggests following suit when have cards of led suit', () => {
    const state = makeState({
      phase: OmiPhase.PLAY,
      currentPlayerIdx: 0,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('HEART', 10), card('SPADE', 14), card('CLOVER', 5)];
    const result = getOmiHint(state);
    expect(result?.reason).toBe('hint.followSuit');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests trump cut when void in led suit and have trump', () => {
    const state = makeState({
      phase: OmiPhase.PLAY,
      currentPlayerIdx: 0,
      trumpSuit: 1,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('SPADE', 14), card('CLOVER', 5)];
    const result = getOmiHint(state);
    expect(result?.reason).toBe('hint.trumpCut');
    expect(result?.confidence).toBe('moderate');
  });

  it('suggests leading off-suit when void and no trump (Omi allows any card when void)', () => {
    const state = makeState({
      phase: OmiPhase.PLAY,
      currentPlayerIdx: 0,
      trumpSuit: 1,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    // No heart (led suit), no spade (trump)
    state.players[0].cards = [card('DIAMOND', 3), card('CLOVER', 5)];
    const result = getOmiHint(state);
    // Any card is legal — recommend off-suit lead
    expect(result?.reason).toBe('hint.leadOffSuit');
    expect(result?.confidence).toBe('moderate');
  });
});
