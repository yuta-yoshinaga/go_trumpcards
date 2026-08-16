import { describe, expect, it } from 'vitest';
import type { PokerResponse } from '../../types/card';
import { PokerPhase } from '../../types/phases';
import { getPokerHint } from './pokerHint';

function makeState(overrides: Partial<PokerResponse> = {}): PokerResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cards: [],
        chips: 1000,
        currentBet: 0,
        folded: false,
        allIn: false,
        handRank: 0,
        handName: 'High Card',
        exchangeCount: 0,
        playStyleName: '',
      },
    ],
    pot: 0,
    sidePots: [],
    dealerIdx: 0,
    currentTurn: 0,
    phase: PokerPhase.EXCHANGE,
    exchangeRead: false,
    gameEndFlag: false,
    lastBet: 0,
    minRaise: 0,
    ante: 10,
    jokerCount: 0,
    bettingLimit: 0,
    raiseCount: 0,
    maxBetAmount: 0,
    roundResults: [],
    cpuActions: [],
    cpuExchanges: [],
    isLowball: false,
    message: '',
    ...overrides,
  };
}

describe('getPokerHint', () => {
  it('returns null when no human player', () => {
    const state = makeState({ players: [{ ...makeState().players[0], isHuman: false }] });
    expect(getPokerHint(state)).toBeNull();
  });

  it('returns null in INIT phase', () => {
    expect(getPokerHint(makeState({ phase: PokerPhase.INIT }))).toBeNull();
  });

  it('returns null in END phase', () => {
    expect(getPokerHint(makeState({ phase: PokerPhase.END }))).toBeNull();
  });

  // Exchange phase
  it('suggests exchange for weak hand (high card)', () => {
    const result = getPokerHint(makeState());
    expect(result?.targetAction).toBe('exchange');
    expect(result?.confidence).toBe('moderate');
  });

  it('suggests exchange for one pair', () => {
    const state = makeState();
    state.players[0].handRank = 1;
    const result = getPokerHint(state);
    expect(result?.targetAction).toBe('exchange');
  });

  it('suggests stand for two pair', () => {
    const state = makeState();
    state.players[0].handRank = 2;
    const result = getPokerHint(state);
    expect(result?.targetAction).toBe('stand');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests stand for three of a kind', () => {
    const state = makeState();
    state.players[0].handRank = 3;
    expect(getPokerHint(state)?.targetAction).toBe('stand');
  });

  // Betting phases
  it('suggests fold for high card in DEAL phase', () => {
    const state = makeState({ phase: PokerPhase.DEAL });
    state.players[0].handRank = 0;
    const result = getPokerHint(state);
    expect(result?.targetAction).toBe('fold');
  });

  it('suggests call for one pair in SECOND_BET phase', () => {
    const state = makeState({ phase: PokerPhase.SECOND_BET });
    state.players[0].handRank = 1;
    const result = getPokerHint(state);
    expect(result?.targetAction).toBe('call');
  });

  it('suggests raise for three of a kind', () => {
    const state = makeState({ phase: PokerPhase.DEAL });
    state.players[0].handRank = 3;
    const result = getPokerHint(state);
    expect(result?.targetAction).toBe('raise');
    expect(result?.confidence).toBe('strong');
  });

  it('returns null when player has folded', () => {
    const state = makeState({ phase: PokerPhase.SECOND_BET });
    state.players[0].folded = true;
    expect(getPokerHint(state)).toBeNull();
  });
});
