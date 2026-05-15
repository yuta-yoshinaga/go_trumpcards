import { describe, expect, it } from 'vitest';
import type { Card, MightyResponse } from '../../types/card';
import { MightyPhase } from '../../types/phases';
import { getMightyHint } from './mightyHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<MightyResponse> = {}): MightyResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 10,
        cards: [card('SPADE', 14), card('HEART', 13), card('CLOVER', 5)],
        bid: -1,
        bidNoTrump: false,
        isDeclarer: false,
        isPartner: false,
        partnerRevealed: false,
        pointCards: 0,
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
        bidNoTrump: false,
        isDeclarer: false,
        isPartner: false,
        partnerRevealed: false,
        pointCards: 0,
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
        bidNoTrump: false,
        isDeclarer: false,
        isPartner: false,
        partnerRevealed: false,
        pointCards: 0,
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
        bidNoTrump: false,
        isDeclarer: false,
        isPartner: false,
        partnerRevealed: false,
        pointCards: 0,
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
        bidNoTrump: false,
        isDeclarer: false,
        isPartner: false,
        partnerRevealed: false,
        pointCards: 0,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
    ],
    phase: MightyPhase.PLAY,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    currentTrick: [],
    trumpSuit: 1,
    partnerCard: null,
    declarerIdx: 0,
    partnerIdx: -1,
    partnerRevealed: false,
    highestBid: 13,
    highestBidder: 0,
    winningBidNoTrump: false,
    kitty: [],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 0, minBid: 13, noTrumpExtra: 1, pointLimit: 100 },
    ...overrides,
  };
}

describe('getMightyHint', () => {
  // Null/guard conditions
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getMightyHint(state)).toBeNull();
  });

  it('returns null when human has no cards', () => {
    const state = makeState();
    state.players[0].cards = [];
    expect(getMightyHint(state)).toBeNull();
  });

  it('returns null in TRICK_END phase', () => {
    expect(getMightyHint(makeState({ phase: MightyPhase.TRICK_END }))).toBeNull();
  });

  it('returns null in ROUND_END phase', () => {
    expect(getMightyHint(makeState({ phase: MightyPhase.ROUND_END }))).toBeNull();
  });

  it('returns null in GAME_END phase', () => {
    expect(getMightyHint(makeState({ phase: MightyPhase.GAME_END }))).toBeNull();
  });

  // Bid phase
  it('returns null in BID phase when not human bid turn', () => {
    expect(getMightyHint(makeState({ phase: MightyPhase.BID, bidPlayerIdx: 1 }))).toBeNull();
  });

  it('returns strong bid hint for many high cards', () => {
    const state = makeState({ phase: MightyPhase.BID, bidPlayerIdx: 0 });
    state.players[0].cards = [card('SPADE', 13), card('HEART', 13), card('DIAMOND', 12), card('CLOVER', 11)];
    const result = getMightyHint(state);
    expect(result?.reason).toBe('hint.bidStrong');
    expect(result?.confidence).toBe('strong');
  });

  it('returns moderate bid hint for weak cards', () => {
    const state = makeState({ phase: MightyPhase.BID, bidPlayerIdx: 0 });
    state.players[0].cards = [card('CLOVER', 3), card('DIAMOND', 5), card('HEART', 7)];
    const result = getMightyHint(state);
    expect(result?.reason).toBe('hint.bidModerate');
    expect(result?.confidence).toBe('moderate');
  });

  it('returns noTrump hint when holding both Mighty and Joker', () => {
    const state = makeState({ phase: MightyPhase.BID, bidPlayerIdx: 0 });
    state.players[0].cards = [card('SPADE', 1), card('JOKER', 0), card('HEART', 3)];
    const result = getMightyHint(state);
    expect(result?.reason).toBe('hint.bidNoTrump');
    expect(result?.confidence).toBe('strong');
  });

  // Trump-and-friend declaration phase
  it('returns null in TRUMP_AND_FRIEND when human is not declarer', () => {
    expect(getMightyHint(makeState({ phase: MightyPhase.TRUMP_AND_FRIEND, declarerIdx: 2 }))).toBeNull();
  });

  it('suggests trump declaration based on longest suit', () => {
    const state = makeState({ phase: MightyPhase.TRUMP_AND_FRIEND, declarerIdx: 0 });
    state.players[0].cards = [
      card('HEART', 14),
      card('HEART', 13),
      card('HEART', 10),
      card('HEART', 5),
      card('SPADE', 5),
    ];
    const result = getMightyHint(state);
    expect(result?.reason).toBe('hint.declareTrump');
    expect(result?.confidence).toBe('strong');
  });

  it('declareTrump moderate confidence when no strong suit', () => {
    const state = makeState({ phase: MightyPhase.TRUMP_AND_FRIEND, declarerIdx: 0 });
    state.players[0].cards = [card('HEART', 5), card('SPADE', 5), card('CLOVER', 5)];
    const result = getMightyHint(state);
    expect(result?.reason).toBe('hint.declareTrump');
    expect(result?.confidence).toBe('moderate');
  });

  // Kitty exchange phase
  it('returns null in KITTY_EXCHANGE when human is not declarer', () => {
    expect(getMightyHint(makeState({ phase: MightyPhase.KITTY_EXCHANGE, declarerIdx: 3 }))).toBeNull();
  });

  it('suggests discarding weakest in kitty exchange', () => {
    const state = makeState({ phase: MightyPhase.KITTY_EXCHANGE, declarerIdx: 0 });
    const result = getMightyHint(state);
    expect(result?.reason).toBe('hint.discardWeakest');
    expect(result?.confidence).toBe('strong');
  });

  // Play phase
  it('returns null in PLAY phase when not human turn', () => {
    expect(getMightyHint(makeState({ phase: MightyPhase.PLAY, currentPlayerIdx: 2 }))).toBeNull();
  });

  it('suggests leading the joker when holding one', () => {
    const state = makeState({ phase: MightyPhase.PLAY, currentPlayerIdx: 0 });
    state.players[0].cards = [card('JOKER', 0), card('HEART', 3)];
    const result = getMightyHint(state);
    expect(result?.reason).toBe('hint.leadJoker');
  });

  it('suggests leading with strong card when no joker', () => {
    const state = makeState({ phase: MightyPhase.PLAY, currentPlayerIdx: 0 });
    state.players[0].cards = [card('SPADE', 13), card('HEART', 5), card('CLOVER', 3)];
    const result = getMightyHint(state);
    expect(result?.reason).toBe('hint.leadStrong');
  });

  it('suggests leading low when no point/high cards and no joker', () => {
    const state = makeState({ phase: MightyPhase.PLAY, currentPlayerIdx: 0 });
    state.players[0].cards = [card('HEART', 3), card('CLOVER', 4), card('DIAMOND', 5)];
    const result = getMightyHint(state);
    expect(result?.reason).toBe('hint.leadLow');
    expect(result?.confidence).toBe('moderate');
  });

  it('suggests following suit when holding led suit', () => {
    const state = makeState({
      phase: MightyPhase.PLAY,
      currentPlayerIdx: 0,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('HEART', 10), card('SPADE', 13)];
    const result = getMightyHint(state);
    expect(result?.reason).toBe('hint.followSuit');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests trump cut when void and have trump', () => {
    const state = makeState({
      phase: MightyPhase.PLAY,
      currentPlayerIdx: 0,
      trumpSuit: 1,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('SPADE', 8), card('CLOVER', 5)];
    const result = getMightyHint(state);
    expect(result?.reason).toBe('hint.trumpCut');
    expect(result?.confidence).toBe('moderate');
  });

  it('suggests playing joker when void and have joker', () => {
    const state = makeState({
      phase: MightyPhase.PLAY,
      currentPlayerIdx: 0,
      trumpSuit: 1,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('JOKER', 0), card('CLOVER', 5)];
    const result = getMightyHint(state);
    expect(result?.reason).toBe('hint.playJoker');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests playing Mighty (♠A) when void and trump != spades', () => {
    const state = makeState({
      phase: MightyPhase.PLAY,
      currentPlayerIdx: 0,
      trumpSuit: 3, // hearts trump
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('SPADE', 1), card('CLOVER', 5)];
    // Held ♠A and have led suit (HEART) → because we lack HEART here, void check triggers
    // Wait: only SPADE and CLOVER, no HEART. Voided. Trump=hearts. SPADE A is Mighty.
    const result = getMightyHint(state);
    expect(result?.reason).toBe('hint.playMighty');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests discarding low when void and no trump/joker/mighty', () => {
    const state = makeState({
      phase: MightyPhase.PLAY,
      currentPlayerIdx: 0,
      trumpSuit: 1, // spades trump
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('DIAMOND', 3), card('CLOVER', 5)];
    const result = getMightyHint(state);
    expect(result?.reason).toBe('hint.discardLow');
    expect(result?.confidence).toBe('moderate');
  });
});
