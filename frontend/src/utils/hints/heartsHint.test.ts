import { describe, expect, it } from 'vitest';
import type { Card, HeartsResponse } from '../../types/card';
import { HeartsPhase } from '../../types/phases';
import { getHeartsHint } from './heartsHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<HeartsResponse> = {}): HeartsResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 3,
        cards: [card('CLOVER', 3), card('DIAMOND', 5), card('HEART', 10)],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        penaltyCards: [],
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 3,
        cards: [],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        penaltyCards: [],
      },
      {
        id: 2,
        isHuman: false,
        cardCount: 3,
        cards: [],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        penaltyCards: [],
      },
      {
        id: 3,
        isHuman: false,
        cardCount: 3,
        cards: [],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        penaltyCards: [],
      },
    ],
    phase: HeartsPhase.PLAY,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    currentTrick: [],
    heartsBroken: false,
    passDirection: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 0, pointLimit: 100, omnibusJD: false },
    ...overrides,
  };
}

describe('getHeartsHint', () => {
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getHeartsHint(state)).toBeNull();
  });

  it('returns null when human has no cards', () => {
    const state = makeState();
    state.players[0].cards = [];
    expect(getHeartsHint(state)).toBeNull();
  });

  it('returns null in TRICK_END phase', () => {
    expect(getHeartsHint(makeState({ phase: HeartsPhase.TRICK_END }))).toBeNull();
  });

  // Pass phase
  it('suggests passing Queen of Spades', () => {
    const state = makeState({ phase: HeartsPhase.PASS });
    state.players[0].cards = [card('SPADE', 12), card('CLOVER', 3), card('DIAMOND', 5)];
    const result = getHeartsHint(state);
    expect(result?.reason).toBe('hint.passQueenOfSpades');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests passing high hearts', () => {
    const state = makeState({ phase: HeartsPhase.PASS });
    state.players[0].cards = [card('HEART', 13), card('CLOVER', 3), card('DIAMOND', 5)];
    const result = getHeartsHint(state);
    expect(result?.reason).toBe('hint.passHighHearts');
  });

  it('suggests passing high cards when no penalty cards', () => {
    const state = makeState({ phase: HeartsPhase.PASS });
    state.players[0].cards = [card('CLOVER', 3), card('DIAMOND', 5), card('SPADE', 7)];
    const result = getHeartsHint(state);
    expect(result?.reason).toBe('hint.passHighCards');
    expect(result?.confidence).toBe('moderate');
  });

  // Play phase - leading
  it('suggests leading non-heart when hearts not broken', () => {
    const result = getHeartsHint(makeState({ currentTrick: [] }));
    expect(result?.reason).toBe('hint.leadNonHeart');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests leading lowest when hearts broken', () => {
    const result = getHeartsHint(makeState({ currentTrick: [], heartsBroken: true }));
    expect(result?.reason).toBe('hint.leadLowest');
  });

  // Play phase - following
  it('suggests following suit when possible', () => {
    const state = makeState({ currentTrick: [{ playerIdx: 1, card: card('CLOVER', 7) }] });
    const result = getHeartsHint(state);
    expect(result?.reason).toBe('hint.followSuit');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests dumping Queen of Spades when void in led suit', () => {
    const state = makeState({ currentTrick: [{ playerIdx: 1, card: card('SPADE', 7) }] });
    state.players[0].cards = [card('SPADE', 12), card('HEART', 3), card('DIAMOND', 5)];
    // Human has no spades except QS but the led suit is spade... let's make them void in led suit
    state.players[0].cards = [card('SPADE', 12), card('HEART', 3), card('DIAMOND', 5)];
    // Wait, spade IS the led suit and they have QS which is a spade card - so they'd follow suit
    // Let me make led suit be clover, and human has no clover but has QS
    const state2 = makeState({ currentTrick: [{ playerIdx: 1, card: card('CLOVER', 7) }] });
    state2.players[0].cards = [card('SPADE', 12), card('HEART', 3), card('DIAMOND', 5)];
    const result = getHeartsHint(state2);
    expect(result?.reason).toBe('hint.dumpQueenOfSpades');
  });

  it('suggests dumping hearts when void and no QS', () => {
    const state = makeState({ currentTrick: [{ playerIdx: 1, card: card('CLOVER', 7) }] });
    state.players[0].cards = [card('HEART', 13), card('DIAMOND', 5), card('SPADE', 3)];
    const result = getHeartsHint(state);
    expect(result?.reason).toBe('hint.dumpHearts');
  });

  it('suggests playing highest when void with no penalty cards', () => {
    const state = makeState({ currentTrick: [{ playerIdx: 1, card: card('CLOVER', 7) }] });
    state.players[0].cards = [card('DIAMOND', 13), card('DIAMOND', 5), card('SPADE', 3)];
    const result = getHeartsHint(state);
    expect(result?.reason).toBe('hint.playHighest');
  });

  it('returns null when not current player turn', () => {
    const state = makeState({ currentPlayerIdx: 2 });
    expect(getHeartsHint(state)).toBeNull();
  });
});
