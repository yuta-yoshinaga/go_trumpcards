import { describe, expect, it } from 'vitest';
import type { Card, DramahaResponse } from '../../../types/card';
import { DramahaPhase } from '../../../types/phases';
import { formatDramahaState } from './dramahaFormatter';

const c = (design: Card['design'], value: number): Card => ({ design, value });

/** Five hole cards that are a pair of fours and nothing more. */
const PAIR_HOLE = [c('SPADE', 4), c('SPADE', 8), c('DIAMOND', 4), c('HEART', 9), c('CLOVER', 2)];
/** A board of four spades — with the two hole spades it makes an Omaha flush. */
const SPADE_BOARD = [c('SPADE', 13), c('SPADE', 11), c('SPADE', 6), c('SPADE', 3), c('HEART', 10)];

function makeState(overrides: Partial<DramahaResponse> = {}): DramahaResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cards: PAIR_HOLE,
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
        af: '0',
      },
    ],
    communityCards: [],
    pot: 100,
    sidePots: [],
    dealerIdx: 0,
    currentTurn: 0,
    phase: DramahaPhase.FLOP,
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

describe('formatDramahaState', () => {
  it('titles the output Dramaha, not the game it was cloned from', () => {
    const out = formatDramahaState(makeState());
    expect(out).toContain('Dramaha');
    expect(out).not.toContain("Texas Hold'em");
  });

  it('names the draw phase, which the Hold-em phase table has no entry for', () => {
    expect(formatDramahaState(makeState({ phase: DramahaPhase.DRAW }))).toContain('DRAW');
  });

  it('prints how to draw while the draw round is open, and not otherwise', () => {
    expect(formatDramahaState(makeState({ phase: DramahaPhase.DRAW }))).toContain('Draw round');
    expect(formatDramahaState(makeState({ phase: DramahaPhase.TURN }))).not.toContain('Draw round');
  });

  it('prints both halves of the split, which read differently on the same cards', () => {
    const out = formatDramahaState(makeState({ communityCards: SPADE_BOARD, phase: DramahaPhase.RIVER }));
    expect(out).toContain('Omaha hand: flush');
    expect(out).toContain('Draw hand: one pair');
  });

  it('prints the draw hand before there is a board to make an Omaha hand from', () => {
    const out = formatDramahaState(makeState({ phase: DramahaPhase.PRE_FLOP }));
    expect(out).toContain('Draw hand: one pair');
    expect(out).not.toContain('Omaha hand:');
  });

  it('says the pot always splits', () => {
    expect(formatDramahaState(makeState())).toContain('Pot always splits');
  });

  it('shows no hands for a folded seat', () => {
    const state = makeState({ communityCards: SPADE_BOARD });
    state.players[0].folded = true;
    const out = formatDramahaState(state);
    expect(out).not.toContain('Omaha hand:');
    expect(out).not.toContain('Draw hand:');
  });

  it('surfaces the muck prompt', () => {
    expect(formatDramahaState(makeState({ muckAvailable: true }))).toContain('Muck or Show?');
  });
});
