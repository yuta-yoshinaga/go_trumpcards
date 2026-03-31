import { describe, expect, it } from 'vitest';
import type { HoldemResponse } from '../../types/card';
import { HoldemPhase } from '../../types/phases';
import { getHoldemBaseHint } from './holdemBaseHint';

function makeState(overrides: Partial<HoldemResponse> = {}): HoldemResponse {
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
        bestHand: [],
        playStyleName: '',
        totalHands: 0,
        vpip: 0,
        pfr: 0,
        threeBet: 0,
        af: '0',
      },
    ],
    communityCards: [],
    pot: 100,
    sidePots: [],
    dealerIdx: 0,
    currentTurn: 0,
    phase: HoldemPhase.FLOP,
    gameEndFlag: false,
    lastBet: 0,
    minRaise: 0,
    bettingLimit: 0,
    raiseCount: 0,
    maxBetAmount: 0,
    roundResults: [],
    cpuActions: [],
    message: '',
    handCount: 1,
    smallBlind: 5,
    bigBlind: 10,
    tournamentMode: false,
    blindLevelHands: 0,
    blindMultiplier: 0,
    tableSize: 6,
    rebuyPhaseType: 0,
    rebuyChips: 0,
    rebuyMaxCount: 0,
    rebuyCounts: [],
    addonChips: 0,
    rebuyAvailable: false,
    addonAvailable: false,
    rebuyEnabled: false,
    addonEnabled: false,
    rebuyPeriodHands: 0,
    addonAfterHand: 0,
    addonUsed: [],
    muckAvailable: false,
    ...overrides,
  };
}

describe('getHoldemBaseHint', () => {
  it('returns null in INIT phase', () => {
    expect(getHoldemBaseHint(makeState({ phase: HoldemPhase.INIT }))).toBeNull();
  });

  it('returns null in SHOWDOWN phase', () => {
    expect(getHoldemBaseHint(makeState({ phase: HoldemPhase.SHOWDOWN }))).toBeNull();
  });

  it('returns null in END phase', () => {
    expect(getHoldemBaseHint(makeState({ phase: HoldemPhase.END }))).toBeNull();
  });

  it('returns null in REBUY phase', () => {
    expect(getHoldemBaseHint(makeState({ phase: HoldemPhase.REBUY }))).toBeNull();
  });

  it('returns null when no human player', () => {
    const state = makeState({ players: [{ ...makeState().players[0], isHuman: false }] });
    expect(getHoldemBaseHint(state)).toBeNull();
  });

  it('returns null when human is folded', () => {
    const state = makeState();
    state.players[0].folded = true;
    expect(getHoldemBaseHint(state)).toBeNull();
  });

  it('returns null when human is all-in', () => {
    const state = makeState();
    state.players[0].allIn = true;
    expect(getHoldemBaseHint(state)).toBeNull();
  });

  // Equity-based hints
  it('suggests raise when win probability is well above pot odds', () => {
    const state = makeState({ equity: { winProbability: 0.7, handOdds: [] }, potOdds: 0.3 });
    const result = getHoldemBaseHint(state);
    expect(result?.targetAction).toBe('raise');
    expect(result?.reason).toBe('hint.positiveEV');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests call when win probability roughly equals pot odds', () => {
    const state = makeState({ equity: { winProbability: 0.35, handOdds: [] }, potOdds: 0.3 });
    const result = getHoldemBaseHint(state);
    expect(result?.targetAction).toBe('call');
    expect(result?.reason).toBe('hint.marginalEV');
    expect(result?.confidence).toBe('moderate');
  });

  it('suggests fold when win probability is below pot odds', () => {
    const state = makeState({ equity: { winProbability: 0.2, handOdds: [] }, potOdds: 0.4 });
    const result = getHoldemBaseHint(state);
    expect(result?.targetAction).toBe('fold');
    expect(result?.reason).toBe('hint.negativeEV');
    expect(result?.confidence).toBe('moderate');
  });

  // Hand rank fallback hints (no equity data)
  it('suggests raise for strong hand rank (three of a kind+)', () => {
    const state = makeState();
    state.players[0].handRank = 3;
    const result = getHoldemBaseHint(state);
    expect(result?.targetAction).toBe('raise');
    expect(result?.reason).toBe('hint.strongHandRank');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests call for decent hand rank (one pair+)', () => {
    const state = makeState();
    state.players[0].handRank = 1;
    const result = getHoldemBaseHint(state);
    expect(result?.targetAction).toBe('call');
    expect(result?.reason).toBe('hint.decentHandRank');
  });

  it('suggests fold for weak hand rank (high card)', () => {
    const state = makeState();
    state.players[0].handRank = 0;
    const result = getHoldemBaseHint(state);
    expect(result?.targetAction).toBe('fold');
    expect(result?.reason).toBe('hint.weakHandRank');
  });

  it('uses equity hint over hand rank when equity is available', () => {
    const state = makeState({ equity: { winProbability: 0.8, handOdds: [] }, potOdds: 0.2 });
    state.players[0].handRank = 0; // weak hand but strong equity
    const result = getHoldemBaseHint(state);
    expect(result?.targetAction).toBe('raise');
    expect(result?.reason).toBe('hint.positiveEV');
  });
});
