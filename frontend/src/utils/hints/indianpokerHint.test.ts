import { describe, expect, it } from 'vitest';
import type { IndianPokerResponse } from '../../types/card';
import { IndianPokerPhase } from '../../types/phases';
import { getIndianPokerHint } from './indianpokerHint';

function makePlayer(overrides: Partial<IndianPokerResponse['players'][0]> = {}) {
  return {
    id: 0,
    isHuman: true,
    card: null,
    chips: 1000,
    currentBet: 0,
    folded: false,
    allIn: false,
    cardRank: 7,
    playStyleName: '',
    ...overrides,
  };
}

function makeState(overrides: Partial<IndianPokerResponse> = {}): IndianPokerResponse {
  return {
    estimatedStrength: 50,
    players: [
      makePlayer(),
      makePlayer({ id: 1, isHuman: false, cardRank: 7 }),
      makePlayer({ id: 2, isHuman: false, cardRank: 7 }),
    ],
    pot: 100,
    sidePots: [],
    dealerIdx: 0,
    currentTurn: 0,
    phase: IndianPokerPhase.BETTING,
    gameEndFlag: false,
    lastBet: 0,
    minRaise: 0,
    bettingLimit: 0,
    raiseCount: 0,
    maxBetAmount: 0,
    roundResults: [],
    cpuActions: [],
    handCount: 1,
    ante: 10,
    message: '',
    ...overrides,
  };
}

describe('getIndianPokerHint', () => {
  it('returns null in INIT phase', () => {
    expect(getIndianPokerHint(makeState({ phase: IndianPokerPhase.INIT }))).toBeNull();
  });

  it('returns null in ANTE phase', () => {
    expect(getIndianPokerHint(makeState({ phase: IndianPokerPhase.ANTE }))).toBeNull();
  });

  it('returns null in SHOWDOWN phase', () => {
    expect(getIndianPokerHint(makeState({ phase: IndianPokerPhase.SHOWDOWN }))).toBeNull();
  });

  it('returns null in END phase', () => {
    expect(getIndianPokerHint(makeState({ phase: IndianPokerPhase.END }))).toBeNull();
  });

  it('returns null when no human player', () => {
    const state = makeState({
      players: [makePlayer({ isHuman: false }), makePlayer({ id: 1, isHuman: false })],
    });
    expect(getIndianPokerHint(state)).toBeNull();
  });

  it('returns null when human is folded', () => {
    const state = makeState();
    state.players[0].folded = true;
    expect(getIndianPokerHint(state)).toBeNull();
  });

  it('returns null when human is all-in', () => {
    const state = makeState();
    state.players[0].allIn = true;
    expect(getIndianPokerHint(state)).toBeNull();
  });

  it('returns null when all opponents are folded', () => {
    const state = makeState();
    state.players[1].folded = true;
    state.players[2].folded = true;
    expect(getIndianPokerHint(state)).toBeNull();
  });

  it('suggests fold when opponents have high card ranks', () => {
    const state = makeState({
      players: [
        makePlayer(),
        makePlayer({ id: 1, isHuman: false, cardRank: 10 }),
        makePlayer({ id: 2, isHuman: false, cardRank: 12 }),
      ],
    });
    const result = getIndianPokerHint(state);
    expect(result?.targetAction).toBe('fold');
    expect(result?.reason).toBe('hint.opponentsStrong');
  });

  it('suggests raise when opponents have low card ranks', () => {
    const state = makeState({
      players: [
        makePlayer(),
        makePlayer({ id: 1, isHuman: false, cardRank: 3 }),
        makePlayer({ id: 2, isHuman: false, cardRank: 4 }),
      ],
    });
    const result = getIndianPokerHint(state);
    expect(result?.targetAction).toBe('raise');
    expect(result?.reason).toBe('hint.opponentsWeak');
  });

  it('suggests call when opponents have medium card ranks', () => {
    const state = makeState({
      players: [
        makePlayer(),
        makePlayer({ id: 1, isHuman: false, cardRank: 6 }),
        makePlayer({ id: 2, isHuman: false, cardRank: 8 }),
      ],
    });
    const result = getIndianPokerHint(state);
    expect(result?.targetAction).toBe('call');
    expect(result?.reason).toBe('hint.uncertain');
  });
});
