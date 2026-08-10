import { describe, expect, it } from 'vitest';
import type { Card, FiveCardStudResponse } from '../../types/card';
import { FiveCardStudPhase } from '../../types/phases';
import { getFiveCardStudHint } from './fivecardstudHint';
import { getSokoHint } from './sokoHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hole?: Card[]; door?: Card[]; currentBet?: number; folded?: boolean };

function base({
  hole = [card('SPADE', 4)],
  door = [card('HEART', 7)],
  currentBet = 0,
  folded = false,
  ...overrides
}: Partial<FiveCardStudResponse> & Extra = {}) {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        holeCards: hole,
        doorCards: door,
        chips: 200,
        currentBet,
        folded,
        allIn: false,
        handRank: 0,
        handName: '',
        bestHand: [],
        playStyleName: '',
        totalHands: 0,
        vpip: 0,
      },
    ],
    communityCard: null,
    pot: 30,
    sidePots: [],
    dealerIdx: 1,
    currentTurn: 0,
    phase: FiveCardStudPhase.THIRD_STREET,
    gameEndFlag: false,
    lastBet: 0,
    minRaise: 10,
    bettingLimit: 0,
    maxBetAmount: 0,
    raiseCount: 0,
    roundResults: [],
    cpuActions: [],
    handCount: 1,
    ante: 5,
    bringIn: 5,
    smallBet: 10,
    bigBet: 20,
    message: '',
    ...overrides,
  } as FiveCardStudResponse;
}

describe('getSokoHint', () => {
  // Soko's betting decisions ARE Five Card Stud's — same streets, same bring-in,
  // same pot odds — and the extra hand ranks are resolved server-side before the
  // page sees them. So the contract this pins is the delegation itself: if
  // someone forks a private copy here, these tests catch the divergence.
  it('returns exactly what the Five Card Stud hint returns, on a pair', () => {
    const s = base({ hole: [card('SPADE', 9)], door: [card('HEART', 9)] });
    const hint = getSokoHint(s);
    expect(hint).not.toBeNull();
    expect(hint).toEqual(getFiveCardStudHint(s));
  });

  it('agrees with the Five Card Stud hint on a weak holding too', () => {
    const s = base({ hole: [card('SPADE', 3)], door: [card('HEART', 8)], lastBet: 40, currentBet: 0 });
    expect(getSokoHint(s)).toEqual(getFiveCardStudHint(s));
  });

  it('stays quiet once the game is over', () => {
    expect(getSokoHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet when another seat is on turn', () => {
    expect(getSokoHint(base({ currentTurn: 1 }))).toBeNull();
  });

  it('stays quiet after folding', () => {
    expect(getSokoHint(base({ folded: true }))).toBeNull();
  });

  it('stays quiet outside a betting street', () => {
    expect(getSokoHint(base({ phase: FiveCardStudPhase.SHOWDOWN }))).toBeNull();
  });

  // Negative control for the whole file: a factory that always returned null
  // would satisfy every "stays quiet" case above. At least one input must
  // produce a hint, which is also what check-hint-coverage.mjs demands of a
  // registered factory.
  it('is not a null stub', () => {
    const s = base({ hole: [card('SPADE', 9)], door: [card('HEART', 9)] });
    expect(getSokoHint(s)).not.toBeNull();
  });
});
