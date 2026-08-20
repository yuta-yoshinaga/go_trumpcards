import { describe, expect, it } from 'vitest';
import type { CardDesign, PitchResponse } from '../../types/card';
import { PitchPhase } from '../../types/phases';
import { getPitchHint } from './pitchHint';

const card = (design: CardDesign, value: number) => ({ design, value });

const baseConfig = { cpuDifficulty: 1, pointLimit: 7 };

const makeState = (override: Partial<PitchResponse> = {}): PitchResponse => ({
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 0,
      cards: [],
      bid: -1,
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    { id: 1, isHuman: false, cardCount: 6, cards: [], bid: -1, roundScore: 0, cumulativeScore: 0, trickCount: 0 },
    { id: 2, isHuman: false, cardCount: 6, cards: [], bid: -1, roundScore: 0, cumulativeScore: 0, trickCount: 0 },
    { id: 3, isHuman: false, cardCount: 6, cards: [], bid: -1, roundScore: 0, cumulativeScore: 0, trickCount: 0 },
  ],
  phase: PitchPhase.BID,
  roundNumber: 1,
  trickNumber: 0,
  dealerIdx: 3,
  currentPlayerIdx: -1,
  bidPlayerIdx: 0,
  currentBid: 0,
  bidWinnerIdx: -1,
  trumpSuit: 0,
  currentTrick: [],
  lastTrick: [],
  lastTrickWinner: -1,
  gameEndFlag: false,
  winnerIdx: -1,
  leadPlayerIdx: -1,
  roundBreakdown: { high: -1, low: -1, jack: -1, game: -1 },
  validPlayIndices: [],
  message: '',
  config: baseConfig,
  ...override,
});

describe('getPitchHint', () => {
  it('returns null when no human cards', () => {
    expect(getPitchHint(makeState())).toBeNull();
  });

  it('suggests pass with weak hand', () => {
    const state = makeState();
    state.players[0].cards = [card('CLOVER', 4), card('DIAMOND', 5), card('HEART', 6)];
    const hint = getPitchHint(state);
    expect(hint?.targetAction).toBe('bid:0');
  });

  it('suggests bid with strong hand', () => {
    const state = makeState();
    state.players[0].cards = [
      card('SPADE', 1),
      card('SPADE', 13),
      card('HEART', 1),
      card('HEART', 12),
      card('DIAMOND', 13),
    ];
    const hint = getPitchHint(state);
    expect(hint?.targetAction).toMatch(/^bid:[2-4]$/);
  });

  it('returns null when not human bid turn', () => {
    const state = makeState({ bidPlayerIdx: 1 });
    state.players[0].cards = [card('SPADE', 1)];
    expect(getPitchHint(state)).toBeNull();
  });

  it('suggests trump-set hint when leading first card of round', () => {
    const state = makeState({
      phase: PitchPhase.PLAY,
      currentPlayerIdx: 0,
      trickNumber: 1,
      currentTrick: [],
      trumpSuit: 0,
    });
    state.players[0].cards = [card('SPADE', 1), card('HEART', 9)];
    const hint = getPitchHint(state);
    expect(hint?.reason).toBe('hint.setTrumpLead');
  });

  it('suggests follow suit if has lead suit', () => {
    const state = makeState({
      phase: PitchPhase.PLAY,
      currentPlayerIdx: 0,
      trickNumber: 2,
      trumpSuit: 1,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 10) }],
    });
    state.players[0].cards = [card('HEART', 9), card('SPADE', 5)];
    const hint = getPitchHint(state);
    expect(hint?.reason).toBe('hint.followSuit');
  });

  it('suggests discard low when void in lead suit', () => {
    const state = makeState({
      phase: PitchPhase.PLAY,
      currentPlayerIdx: 0,
      trickNumber: 2,
      trumpSuit: 1,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 10) }],
    });
    state.players[0].cards = [card('CLOVER', 5), card('DIAMOND', 4)];
    const hint = getPitchHint(state);
    expect(hint?.reason).toBe('hint.discardLow');
  });

  it('returns null in non-bid/play phase', () => {
    const state = makeState({ phase: PitchPhase.TRICK_END });
    state.players[0].cards = [card('SPADE', 1)];
    expect(getPitchHint(state)).toBeNull();
  });
});
