import { describe, expect, it } from 'vitest';
import type { HoldemResponse } from '../../types/card';
import { HoldemPhase } from '../../types/phases';
import { getHoldemHint } from './holdemHint';

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

describe('getHoldemHint', () => {
  it('returns null in INIT phase', () => {
    expect(getHoldemHint(makeState({ phase: HoldemPhase.INIT }))).toBeNull();
  });

  it('returns null when human is folded', () => {
    const state = makeState();
    state.players[0].folded = true;
    expect(getHoldemHint(state)).toBeNull();
  });

  it('suggests raise with positive EV', () => {
    const state = makeState({ equity: { winProbability: 0.7, handOdds: [] }, potOdds: 0.3 });
    const result = getHoldemHint(state);
    expect(result?.targetAction).toBe('raise');
    expect(result?.reason).toBe('hint.positiveEV');
  });

  it('suggests fold for weak hand rank without equity', () => {
    const result = getHoldemHint(makeState());
    expect(result?.targetAction).toBe('fold');
    expect(result?.reason).toBe('hint.weakHandRank');
  });

  it('suggests call for decent hand rank', () => {
    const state = makeState();
    state.players[0].handRank = 1;
    expect(getHoldemHint(state)?.targetAction).toBe('call');
  });
});
