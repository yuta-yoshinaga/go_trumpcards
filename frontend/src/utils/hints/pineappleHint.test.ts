import { describe, expect, it } from 'vitest';
import type { PineappleResponse } from '../../types/card';
import { HoldemPhase, PineapplePhase } from '../../types/phases';
import { getPineappleHint } from './pineappleHint';

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
    isDiscardPhase: false,
    discardDone: [false],
    initialDealCount: 3,
    liveBestHand: '',
    ...overrides,
  };
}

describe('getPineappleHint', () => {
  // Discard phase
  it('suggests discard in discard phase', () => {
    const state = makeState({ isDiscardPhase: true, phase: PineapplePhase.DISCARD });
    const result = getPineappleHint(state);
    expect(result?.targetAction).toBe('discard');
    expect(result?.reason).toBe('hint.discardWeakest');
    expect(result?.confidence).toBe('moderate');
  });

  it('returns null in discard phase when human already discarded', () => {
    const state = makeState({ isDiscardPhase: true, discardDone: [true] });
    expect(getPineappleHint(state)).toBeNull();
  });

  it('returns null in discard phase when no human player', () => {
    const state = makeState({ isDiscardPhase: true, players: [{ ...makeState().players[0], isHuman: false }] });
    expect(getPineappleHint(state)).toBeNull();
  });

  // Delegates to holdemBaseHint for non-discard phases
  it('suggests raise with positive EV in betting phase', () => {
    const state = makeState({ equity: { winProbability: 0.7, handOdds: [] }, potOdds: 0.3 });
    const result = getPineappleHint(state);
    expect(result?.targetAction).toBe('raise');
    expect(result?.reason).toBe('hint.positiveEV');
  });

  it('returns null in SHOWDOWN phase', () => {
    expect(getPineappleHint(makeState({ phase: HoldemPhase.SHOWDOWN }))).toBeNull();
  });

  it('suggests fold for weak hand rank', () => {
    expect(getPineappleHint(makeState())?.targetAction).toBe('fold');
  });
});
