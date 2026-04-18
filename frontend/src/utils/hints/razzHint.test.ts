import { describe, expect, it } from 'vitest';
import type { SevenCardStudResponse } from '../../types/card';
import { SevenCardStudPhase } from '../../types/phases';
import { getRazzHint } from './razzHint';

function makeState(overrides: Partial<SevenCardStudResponse> = {}): SevenCardStudResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        holeCards: [],
        doorCards: [],
        chips: 1000,
        currentBet: 0,
        folded: false,
        allIn: false,
        handRank: 0,
        handName: '',
        bestHand: [],
        playStyleName: '',
        totalHands: 0,
        vpip: 0,
        pfr: 0,
        threeBet: 0,
        af: '',
      },
    ],
    communityCard: null,
    pot: 0,
    sidePots: [],
    dealerIdx: 0,
    currentTurn: 0,
    phase: SevenCardStudPhase.THIRD_STREET,
    gameEndFlag: false,
    lastBet: 0,
    minRaise: 0,
    bettingLimit: 0,
    raiseCount: 0,
    maxBetAmount: 0,
    roundResults: [],
    cpuActions: [],
    handCount: 1,
    ante: 0,
    bringIn: 0,
    smallBet: 0,
    bigBet: 0,
    tournamentMode: false,
    anteLevelHands: 0,
    anteMultiplier: 1,
    tableSize: 4,
    bringInPlayerIdx: 0,
    rebuyAvailable: false,
    addonAvailable: false,
    rebuyCounts: [],
    addonUsed: [],
    rebuyEnabled: false,
    addonEnabled: false,
    rebuyMaxCount: 0,
    rebuyChips: 0,
    addonChips: 0,
    rebuyPeriodHands: 0,
    addonAfterHand: 0,
    rebuyPhaseType: 0,
    muckAvailable: false,
    message: '',
    ...overrides,
  };
}

describe('getRazzHint', () => {
  it('returns null outside betting phases', () => {
    const state = makeState({ phase: SevenCardStudPhase.SHOWDOWN });
    expect(getRazzHint(state)).toBeNull();
  });

  it('returns null when it is not the human turn', () => {
    const state = makeState({ currentTurn: 1 });
    expect(getRazzHint(state)).toBeNull();
  });

  it('returns null when human has folded', () => {
    const state = makeState();
    state.players[0].folded = true;
    expect(getRazzHint(state)).toBeNull();
  });

  it('recommends raise on three low cards', () => {
    const state = makeState();
    state.players[0].holeCards = [
      { design: 'SPADE', value: 1 },
      { design: 'HEART', value: 3 },
    ];
    state.players[0].doorCards = [{ design: 'DIAMOND', value: 5 }];
    const hint = getRazzHint(state);
    expect(hint?.targetAction).toBe('raise');
    expect(hint?.confidence).toBe('strong');
  });

  it('recommends call with two low cards and one high', () => {
    const state = makeState();
    state.players[0].holeCards = [
      { design: 'SPADE', value: 2 },
      { design: 'HEART', value: 11 },
    ];
    state.players[0].doorCards = [{ design: 'DIAMOND', value: 6 }];
    const hint = getRazzHint(state);
    expect(hint?.targetAction).toBe('call');
    expect(hint?.confidence).toBe('moderate');
  });

  it('recommends fold when the hand is paired', () => {
    const state = makeState();
    state.players[0].holeCards = [
      { design: 'SPADE', value: 5 },
      { design: 'HEART', value: 5 },
    ];
    state.players[0].doorCards = [{ design: 'DIAMOND', value: 3 }];
    const hint = getRazzHint(state);
    expect(hint?.targetAction).toBe('fold');
  });

  it('recommends fold when most cards are high', () => {
    const state = makeState();
    state.players[0].holeCards = [
      { design: 'SPADE', value: 11 },
      { design: 'HEART', value: 12 },
    ];
    state.players[0].doorCards = [{ design: 'DIAMOND', value: 13 }];
    const hint = getRazzHint(state);
    expect(hint?.targetAction).toBe('fold');
  });

  it('returns null when the human has no cards yet', () => {
    const state = makeState();
    expect(getRazzHint(state)).toBeNull();
  });

  it('recommends fold on 7th street when only 3 of 7 cards are low', () => {
    const state = makeState({ phase: SevenCardStudPhase.SEVENTH_STREET });
    state.players[0].holeCards = [
      { design: 'SPADE', value: 2 },
      { design: 'HEART', value: 4 },
      { design: 'CLUB', value: 6 },
      { design: 'DIAMOND', value: 10 },
    ];
    state.players[0].doorCards = [
      { design: 'SPADE', value: 11 },
      { design: 'HEART', value: 12 },
      { design: 'CLUB', value: 13 },
    ];
    const hint = getRazzHint(state);
    expect(hint?.targetAction).toBe('fold');
  });

  it('recommends raise on 7th street when all 7 cards are low', () => {
    const state = makeState({ phase: SevenCardStudPhase.SEVENTH_STREET });
    state.players[0].holeCards = [
      { design: 'SPADE', value: 1 },
      { design: 'HEART', value: 2 },
      { design: 'CLUB', value: 3 },
      { design: 'DIAMOND', value: 4 },
    ];
    state.players[0].doorCards = [
      { design: 'SPADE', value: 5 },
      { design: 'HEART', value: 6 },
      { design: 'CLUB', value: 7 },
    ];
    const hint = getRazzHint(state);
    expect(hint?.targetAction).toBe('raise');
    expect(hint?.confidence).toBe('strong');
  });

  it('recommends call on 5th street with 4 of 5 cards low', () => {
    const state = makeState({ phase: SevenCardStudPhase.FIFTH_STREET });
    state.players[0].holeCards = [
      { design: 'SPADE', value: 2 },
      { design: 'HEART', value: 4 },
    ];
    state.players[0].doorCards = [
      { design: 'CLUB', value: 6 },
      { design: 'DIAMOND', value: 7 },
      { design: 'SPADE', value: 12 },
    ];
    const hint = getRazzHint(state);
    expect(hint?.targetAction).toBe('call');
    expect(hint?.confidence).toBe('moderate');
  });
});
