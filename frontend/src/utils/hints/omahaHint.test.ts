import { describe, expect, it } from 'vitest';
import type { OmahaResponse } from '../../types/card';
import { HoldemPhase } from '../../types/phases';
import { getOmahaHint } from './omahaHint';

function makeState(overrides: Partial<OmahaResponse> = {}): OmahaResponse {
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

describe('getOmahaHint', () => {
  it('returns null in SHOWDOWN phase', () => {
    expect(getOmahaHint(makeState({ phase: HoldemPhase.SHOWDOWN }))).toBeNull();
  });

  it('returns null when no human player', () => {
    const state = makeState({ players: [{ ...makeState().players[0], isHuman: false }] });
    expect(getOmahaHint(state)).toBeNull();
  });

  it('suggests raise with strong equity', () => {
    const state = makeState({ equity: { winProbability: 0.8, handOdds: [] }, potOdds: 0.2 });
    expect(getOmahaHint(state)?.targetAction).toBe('raise');
  });

  it('suggests fold for weak hand rank', () => {
    expect(getOmahaHint(makeState())?.targetAction).toBe('fold');
  });

  it('suggests call for decent hand rank', () => {
    const state = makeState();
    state.players[0].handRank = 2;
    expect(getOmahaHint(state)?.targetAction).toBe('call');
  });
});
