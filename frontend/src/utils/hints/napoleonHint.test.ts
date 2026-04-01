import { describe, expect, it } from 'vitest';
import type { Card, NapoleonResponse } from '../../types/card';
import { NapoleonPhase } from '../../types/phases';
import { getNapoleonHint } from './napoleonHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<NapoleonResponse> = {}): NapoleonResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 10,
        cards: [card('SPADE', 14), card('HEART', 13), card('CLOVER', 5)],
        bid: -1,
        isNapoleon: false,
        isAdjutant: false,
        adjutantRevealed: false,
        pictureCards: 0,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 10,
        cards: [],
        bid: -1,
        isNapoleon: false,
        isAdjutant: false,
        adjutantRevealed: false,
        pictureCards: 0,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
      {
        id: 2,
        isHuman: false,
        cardCount: 10,
        cards: [],
        bid: -1,
        isNapoleon: false,
        isAdjutant: false,
        adjutantRevealed: false,
        pictureCards: 0,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
      {
        id: 3,
        isHuman: false,
        cardCount: 10,
        cards: [],
        bid: -1,
        isNapoleon: false,
        isAdjutant: false,
        adjutantRevealed: false,
        pictureCards: 0,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
      {
        id: 4,
        isHuman: false,
        cardCount: 10,
        cards: [],
        bid: -1,
        isNapoleon: false,
        isAdjutant: false,
        adjutantRevealed: false,
        pictureCards: 0,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
    ],
    phase: NapoleonPhase.PLAY,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    currentTrick: [],
    trumpSuit: 1,
    adjutantCard: null,
    napoleonIdx: 0,
    adjutantIdx: -1,
    adjutantRevealed: false,
    highestBid: 12,
    highestBidder: 0,
    kitty: [],
    gameEndFlag: false,
    winnerTeam: -1,
    message: '',
    config: { cpuDifficulty: 0, minBid: 12, pointLimit: 100 },
    ...overrides,
  };
}

describe('getNapoleonHint', () => {
  // Null/guard conditions
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getNapoleonHint(state)).toBeNull();
  });

  it('returns null when human has no cards', () => {
    const state = makeState();
    state.players[0].cards = [];
    expect(getNapoleonHint(state)).toBeNull();
  });

  it('returns null in TRICK_END phase', () => {
    expect(getNapoleonHint(makeState({ phase: NapoleonPhase.TRICK_END }))).toBeNull();
  });

  it('returns null in ROUND_END phase', () => {
    expect(getNapoleonHint(makeState({ phase: NapoleonPhase.ROUND_END }))).toBeNull();
  });

  it('returns null in GAME_END phase', () => {
    expect(getNapoleonHint(makeState({ phase: NapoleonPhase.GAME_END }))).toBeNull();
  });

  // Bid phase
  it('returns null in BID phase when not human bid turn', () => {
    expect(getNapoleonHint(makeState({ phase: NapoleonPhase.BID, bidPlayerIdx: 1 }))).toBeNull();
  });

  it('returns bid hint with strong confidence for many high cards', () => {
    const state = makeState({ phase: NapoleonPhase.BID, bidPlayerIdx: 0 });
    state.players[0].cards = [card('SPADE', 14), card('HEART', 13), card('DIAMOND', 12), card('CLOVER', 11)];
    const result = getNapoleonHint(state);
    expect(result?.targetAction).toBe('bid:14');
    expect(result?.reason).toBe('hint.bidStrong');
    expect(result?.confidence).toBe('strong');
  });

  it('returns bid hint with moderate confidence for few high cards', () => {
    const state = makeState({ phase: NapoleonPhase.BID, bidPlayerIdx: 0 });
    state.players[0].cards = [card('CLOVER', 3), card('DIAMOND', 5), card('HEART', 7)];
    const result = getNapoleonHint(state);
    expect(result?.targetAction).toBe('bid:12');
    expect(result?.reason).toBe('hint.bidModerate');
    expect(result?.confidence).toBe('moderate');
  });

  // Trump declaration phase
  it('returns null in TRUMP_DECLARATION when human is not napoleon', () => {
    expect(getNapoleonHint(makeState({ phase: NapoleonPhase.TRUMP_DECLARATION, napoleonIdx: 2 }))).toBeNull();
  });

  it('suggests trump suit based on strongest suit', () => {
    const state = makeState({ phase: NapoleonPhase.TRUMP_DECLARATION, napoleonIdx: 0 });
    state.players[0].cards = [card('HEART', 14), card('HEART', 13), card('HEART', 10), card('SPADE', 5)];
    const result = getNapoleonHint(state);
    expect(result?.reason).toBe('hint.declareTrump');
    expect(result?.confidence).toBe('strong');
  });

  // Kitty exchange phase
  it('returns null in KITTY_EXCHANGE when human is not napoleon', () => {
    expect(getNapoleonHint(makeState({ phase: NapoleonPhase.KITTY_EXCHANGE, napoleonIdx: 3 }))).toBeNull();
  });

  it('suggests discarding weakest non-trump in kitty exchange', () => {
    const state = makeState({ phase: NapoleonPhase.KITTY_EXCHANGE, napoleonIdx: 0, trumpSuit: 1 });
    state.players[0].cards = [card('SPADE', 14), card('HEART', 3), card('CLOVER', 5)];
    const result = getNapoleonHint(state);
    expect(result?.reason).toBe('hint.discardWeakest');
    expect(result?.confidence).toBe('strong');
  });

  // Play phase
  it('returns null in PLAY phase when not human turn', () => {
    expect(getNapoleonHint(makeState({ phase: NapoleonPhase.PLAY, currentPlayerIdx: 2 }))).toBeNull();
  });

  it('suggests leading with strong card', () => {
    const state = makeState({ phase: NapoleonPhase.PLAY, currentPlayerIdx: 0 });
    state.players[0].cards = [card('SPADE', 14), card('HEART', 5), card('CLOVER', 3)];
    const result = getNapoleonHint(state);
    expect(result?.reason).toBe('hint.leadStrong');
  });

  it('suggests leading low when no high cards', () => {
    const state = makeState({ phase: NapoleonPhase.PLAY, currentPlayerIdx: 0 });
    state.players[0].cards = [card('HEART', 3), card('CLOVER', 4), card('DIAMOND', 5)];
    const result = getNapoleonHint(state);
    expect(result?.reason).toBe('hint.leadLow');
    expect(result?.confidence).toBe('moderate');
  });

  it('suggests following suit', () => {
    const state = makeState({
      phase: NapoleonPhase.PLAY,
      currentPlayerIdx: 0,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('HEART', 10), card('SPADE', 14)];
    const result = getNapoleonHint(state);
    expect(result?.reason).toBe('hint.followSuit');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests trump cut when void in led suit', () => {
    const state = makeState({
      phase: NapoleonPhase.PLAY,
      currentPlayerIdx: 0,
      trumpSuit: 1,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('SPADE', 14), card('CLOVER', 5)];
    const result = getNapoleonHint(state);
    expect(result?.reason).toBe('hint.trumpCut');
    expect(result?.confidence).toBe('moderate');
  });

  it('suggests playing joker when void and have joker', () => {
    const state = makeState({
      phase: NapoleonPhase.PLAY,
      currentPlayerIdx: 0,
      trumpSuit: 1,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('JOKER', 0), card('CLOVER', 5)];
    const result = getNapoleonHint(state);
    expect(result?.reason).toBe('hint.playJoker');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests discarding low when void with no trump and no joker', () => {
    const state = makeState({
      phase: NapoleonPhase.PLAY,
      currentPlayerIdx: 0,
      trumpSuit: 1,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('DIAMOND', 3), card('CLOVER', 5)];
    const result = getNapoleonHint(state);
    expect(result?.reason).toBe('hint.discardLow');
    expect(result?.confidence).toBe('moderate');
  });
});
