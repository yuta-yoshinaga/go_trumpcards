import { describe, expect, it } from 'vitest';
import type { Card, EuchreResponse } from '../../types/card';
import { EuchrePhase } from '../../types/phases';
import { getEuchreHint } from './euchreHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<EuchreResponse> = {}): EuchreResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 5,
        cards: [card('SPADE', 14), card('HEART', 10), card('CLOVER', 5)],
        team: 0,
        trickCount: 0,
      },
      { id: 1, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
      { id: 2, isHuman: false, cardCount: 5, cards: [], team: 0, trickCount: 0 },
      { id: 3, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
    ],
    phase: EuchrePhase.PLAY,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    trumpSuit: 1,
    faceUpCard: card('SPADE', 11),
    makerTeam: 0,
    goingAlone: false,
    goingAlonePlayerIdx: -1,
    currentTrick: [],
    teamScores: [0, 0],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 0, pointLimit: 10 },
    ...overrides,
  };
}

describe('getEuchreHint', () => {
  // Null/guard conditions
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getEuchreHint(state)).toBeNull();
  });

  it('returns null when human has no cards', () => {
    const state = makeState();
    state.players[0].cards = [];
    expect(getEuchreHint(state)).toBeNull();
  });

  it('returns null in TRICK_END phase', () => {
    expect(getEuchreHint(makeState({ phase: EuchrePhase.TRICK_END }))).toBeNull();
  });

  it('returns null in ROUND_END phase', () => {
    expect(getEuchreHint(makeState({ phase: EuchrePhase.ROUND_END }))).toBeNull();
  });

  it('returns null in GAME_END phase', () => {
    expect(getEuchreHint(makeState({ phase: EuchrePhase.GAME_END }))).toBeNull();
  });

  // Pick-up phase
  it('returns null in PICK_UP phase when not human bid turn', () => {
    expect(getEuchreHint(makeState({ phase: EuchrePhase.PICK_UP, bidPlayerIdx: 1 }))).toBeNull();
  });

  it('suggests order up (strong) when hand has 2+ trump-suit cards', () => {
    const state = makeState({ phase: EuchrePhase.PICK_UP, bidPlayerIdx: 0, faceUpCard: card('SPADE', 11) });
    state.players[0].cards = [card('SPADE', 14), card('SPADE', 12), card('HEART', 5)];
    const result = getEuchreHint(state);
    expect(result?.reason).toBe('hint.orderUpStrong');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests pass (moderate) when hand has few trump-suit cards', () => {
    const state = makeState({ phase: EuchrePhase.PICK_UP, bidPlayerIdx: 0, faceUpCard: card('SPADE', 11) });
    state.players[0].cards = [card('HEART', 5), card('DIAMOND', 7), card('CLOVER', 3)];
    const result = getEuchreHint(state);
    expect(result?.reason).toBe('hint.passWeak');
    expect(result?.confidence).toBe('moderate');
  });

  // Call trump phase
  it('returns null in CALL_TRUMP phase when not human bid turn', () => {
    expect(getEuchreHint(makeState({ phase: EuchrePhase.CALL_TRUMP, bidPlayerIdx: 2 }))).toBeNull();
  });

  it('suggests calling suit with most cards (strong if 3+)', () => {
    const state = makeState({ phase: EuchrePhase.CALL_TRUMP, bidPlayerIdx: 0 });
    state.players[0].cards = [
      card('HEART', 14),
      card('HEART', 13),
      card('HEART', 10),
      card('CLOVER', 5),
      card('DIAMOND', 3),
    ];
    const result = getEuchreHint(state);
    expect(result?.reason).toBe('hint.callStrongSuit');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests calling suit with moderate confidence if count < 3', () => {
    const state = makeState({ phase: EuchrePhase.CALL_TRUMP, bidPlayerIdx: 0 });
    state.players[0].cards = [
      card('HEART', 14),
      card('HEART', 10),
      card('CLOVER', 5),
      card('DIAMOND', 3),
      card('SPADE', 2),
    ];
    const result = getEuchreHint(state);
    expect(result?.reason).toBe('hint.callStrongSuit');
    expect(result?.confidence).toBe('moderate');
  });

  // Discard phase
  it('returns null in DISCARD phase when human is not dealer', () => {
    expect(getEuchreHint(makeState({ phase: EuchrePhase.DISCARD, dealerIdx: 2 }))).toBeNull();
  });

  it('suggests discarding weakest non-trump card', () => {
    const state = makeState({ phase: EuchrePhase.DISCARD, dealerIdx: 0, trumpSuit: 1 });
    state.players[0].cards = [card('SPADE', 14), card('HEART', 3), card('CLOVER', 5)];
    const result = getEuchreHint(state);
    expect(result?.reason).toBe('hint.discardWeakest');
    expect(result?.confidence).toBe('strong');
  });

  // Play phase - leading
  it('returns null in PLAY phase when not human turn', () => {
    expect(getEuchreHint(makeState({ phase: EuchrePhase.PLAY, currentPlayerIdx: 2 }))).toBeNull();
  });

  it('suggests leading with trump when have many trump cards', () => {
    const state = makeState({ phase: EuchrePhase.PLAY, currentPlayerIdx: 0, trumpSuit: 1 });
    state.players[0].cards = [card('SPADE', 14), card('SPADE', 12), card('HEART', 5)];
    const result = getEuchreHint(state);
    expect(result?.reason).toBe('hint.leadTrump');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests leading off-suit when few trump cards', () => {
    const state = makeState({ phase: EuchrePhase.PLAY, currentPlayerIdx: 0, trumpSuit: 1 });
    state.players[0].cards = [card('HEART', 5), card('DIAMOND', 7), card('CLOVER', 3)];
    const result = getEuchreHint(state);
    expect(result?.reason).toBe('hint.leadOffSuit');
    expect(result?.confidence).toBe('moderate');
  });

  // Play phase - following
  it('suggests following suit when have cards of led suit', () => {
    const state = makeState({
      phase: EuchrePhase.PLAY,
      currentPlayerIdx: 0,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('HEART', 10), card('SPADE', 14), card('CLOVER', 5)];
    const result = getEuchreHint(state);
    expect(result?.reason).toBe('hint.followSuit');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests trump cut when void in led suit and have trump', () => {
    const state = makeState({
      phase: EuchrePhase.PLAY,
      currentPlayerIdx: 0,
      trumpSuit: 1,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('SPADE', 14), card('CLOVER', 5)];
    const result = getEuchreHint(state);
    expect(result?.reason).toBe('hint.trumpCut');
    expect(result?.confidence).toBe('moderate');
  });

  it('suggests discarding weakest when void and no trump', () => {
    const state = makeState({
      phase: EuchrePhase.PLAY,
      currentPlayerIdx: 0,
      trumpSuit: 1,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('DIAMOND', 3), card('CLOVER', 5)];
    const result = getEuchreHint(state);
    expect(result?.reason).toBe('hint.discardWeakest');
    expect(result?.confidence).toBe('moderate');
  });
});
