import { describe, expect, it } from 'vitest';
import type { ShortDeckResponse } from '../../types/card';
import { HoldemPhase } from '../../types/phases';
import { getShortDeckHint } from './shortdeckHint';

function makeState(overrides: Partial<ShortDeckResponse> = {}): ShortDeckResponse {
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
    phase: HoldemPhase.PRE_FLOP,
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

describe('getShortDeckHint', () => {
  it('returns null in END phase', () => {
    expect(getShortDeckHint(makeState({ phase: HoldemPhase.END }))).toBeNull();
  });

  it('returns null when human is all-in', () => {
    const state = makeState();
    state.players[0].allIn = true;
    expect(getShortDeckHint(state)).toBeNull();
  });

  it('suggests fold with negative EV', () => {
    const state = makeState({ equity: { winProbability: 0.1, handOdds: [] }, potOdds: 0.5 });
    expect(getShortDeckHint(state)?.targetAction).toBe('fold');
  });

  it('suggests raise for strong hand rank', () => {
    const state = makeState();
    state.players[0].handRank = 4;
    expect(getShortDeckHint(state)?.targetAction).toBe('raise');
  });

  it('suggests call for marginal EV', () => {
    const state = makeState({ equity: { winProbability: 0.35, handOdds: [] }, potOdds: 0.3 });
    expect(getShortDeckHint(state)?.targetAction).toBe('call');
  });
});
