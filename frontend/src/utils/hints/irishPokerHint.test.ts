import { describe, expect, it } from 'vitest';
import type { PineappleResponse } from '../../types/card';
import { HoldemPhase, PineapplePhase } from '../../types/phases';
import { getIrishPokerHint } from './irishPokerHint';

function makeState(overrides: Partial<PineappleResponse> = {}): PineappleResponse {
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
    rebuyEnabled: false,
    rebuyMaxCount: 0,
    rebuyChips: 0,
    rebuyPeriodHands: 0,
    rebuyCounts: [0],
    rebuyAvailable: false,
    addonEnabled: false,
    addonChips: 0,
    addonAfterHand: 0,
    addonUsed: [false],
    addonAvailable: false,
    muckAvailable: false,
    isDiscardPhase: false,
    discardDone: [false],
    initialDealCount: 4,
    liveBestHand: '',
    ...overrides,
  };
}

describe('getIrishPokerHint', () => {
  it('returns discardWeakest hint during the discard phase', () => {
    const state = makeState({
      phase: PineapplePhase.DISCARD,
      isDiscardPhase: true,
      discardDone: [false],
    });
    const hint = getIrishPokerHint(state);
    expect(hint?.targetAction).toBe('discard');
    expect(hint?.reason).toBe('hint.discardWeakest');
  });

  it('returns null when the human has already discarded', () => {
    const state = makeState({
      phase: PineapplePhase.DISCARD,
      isDiscardPhase: true,
      discardDone: [true],
    });
    expect(getIrishPokerHint(state)).toBeNull();
  });

  it('delegates to the holdem base hint outside discard phase', () => {
    const state = makeState({ phase: HoldemPhase.FLOP, isDiscardPhase: false });
    expect(() => getIrishPokerHint(state)).not.toThrow();
  });
});
