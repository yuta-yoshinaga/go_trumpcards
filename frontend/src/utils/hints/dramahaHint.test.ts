import { describe, expect, it } from 'vitest';
import type { Card, DramahaResponse } from '../../types/card';
import { DramahaPhase } from '../../types/phases';
import { dramahaDiscardSuggestion, getDramahaHint } from './dramahaHint';

const c = (design: Card['design'], value: number): Card => ({ design, value });

/** Five hole cards that rank as trips — strong on the draw side alone. */
const TRIP_KINGS = [c('SPADE', 13), c('HEART', 13), c('DIAMOND', 13), c('CLOVER', 4), c('SPADE', 9)];
/** A full house: every card is part of the hand, so nothing is worth exchanging. */
const FULL_HOUSE = [c('SPADE', 13), c('HEART', 13), c('DIAMOND', 13), c('CLOVER', 4), c('SPADE', 4)];
/** Five unpaired, unconnected, unsuited cards — nothing on either side. */
const RAGS = [c('SPADE', 2), c('HEART', 5), c('DIAMOND', 9), c('CLOVER', 11), c('SPADE', 13)];

function makeState(overrides: Partial<DramahaResponse> = {}): DramahaResponse {
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

/** makeState with the human holding `cards`. */
function withHole(cards: Card[], overrides: Partial<DramahaResponse> = {}): DramahaResponse {
  const state = makeState(overrides);
  state.players[0].cards = cards;
  return state;
}

describe('getDramahaHint — betting rounds', () => {
  it('returns null in SHOWDOWN phase', () => {
    expect(getDramahaHint(withHole(TRIP_KINGS, { phase: DramahaPhase.SHOWDOWN }))).toBeNull();
  });

  it('returns null in END and REBUY phases', () => {
    expect(getDramahaHint(withHole(TRIP_KINGS, { phase: DramahaPhase.END }))).toBeNull();
    expect(getDramahaHint(withHole(TRIP_KINGS, { phase: DramahaPhase.REBUY }))).toBeNull();
  });

  it('returns null when no human player', () => {
    const state = makeState({ players: [{ ...makeState().players[0], isHuman: false }] });
    expect(getDramahaHint(state)).toBeNull();
  });

  it('returns null once the human has folded', () => {
    const state = withHole(TRIP_KINGS);
    state.players[0].folded = true;
    expect(getDramahaHint(state)).toBeNull();
  });

  it('suggests raise with strong equity', () => {
    const state = withHole(RAGS, { equity: { winProbability: 0.8, handOdds: [] }, potOdds: 0.2 });
    expect(getDramahaHint(state)?.targetAction).toBe('raise');
  });

  it('suggests fold when neither half is worth chips', () => {
    expect(getDramahaHint(withHole(RAGS))?.targetAction).toBe('fold');
  });

  it('suggests call for a decent Omaha hand rank', () => {
    const state = withHole(RAGS);
    state.players[0].handRank = 2;
    expect(getDramahaHint(state)?.targetAction).toBe('call');
  });

  it('calls rather than folds when only the draw half is strong', () => {
    // The Omaha side would fold on this rank; the five hole cards are trips, so
    // half the pot is already made and folding gives it away.
    const state = withHole(TRIP_KINGS, { communityCards: [c('HEART', 2), c('CLOVER', 7), c('DIAMOND', 3)] });
    const hint = getDramahaHint(state);
    expect(hint?.targetAction).toBe('call');
    expect(hint?.reason).toBe('hint.drawHalfOnly');
  });

  it('raises for the scoop when both halves are strong', () => {
    const state = withHole(TRIP_KINGS, {
      equity: { winProbability: 0.8, handOdds: [] },
      potOdds: 0.2,
    });
    const hint = getDramahaHint(state);
    expect(hint?.targetAction).toBe('raise');
    expect(hint?.reason).toBe('hint.scoopChance');
  });
});

describe('getDramahaHint — the draw round', () => {
  it('says stand pat when every card is already working', () => {
    const hint = getDramahaHint(withHole(FULL_HOUSE, { phase: DramahaPhase.DRAW }));
    expect(hint?.targetAction).toBe('standpat');
    expect(hint?.reason).toBe('hint.standPat');
  });

  it('still exchanges the idle cards of a made three of a kind', () => {
    // Trips are strong enough to bet, but the two cards outside them make
    // nothing and the draw half is the whole five.
    const hint = getDramahaHint(withHole(TRIP_KINGS, { phase: DramahaPhase.DRAW }));
    expect(hint?.targetAction).toBe('draw');
    expect(hint?.targetIndices).toEqual([3, 4]);
  });

  it('names the exact positions to exchange', () => {
    // A pair of fours at positions 0 and 2; the other three make nothing.
    const hole = [c('CLOVER', 4), c('HEART', 9), c('DIAMOND', 4), c('SPADE', 2), c('CLOVER', 7)];
    const hint = getDramahaHint(withHole(hole, { phase: DramahaPhase.DRAW }));
    expect(hint?.targetAction).toBe('draw');
    expect(hint?.targetIndices).toEqual([1, 3, 4]);
    expect(hint?.reasonParams).toEqual({ count: 3 });
  });

  it('keeps the two highest and draws three when nothing pairs', () => {
    // RAGS: 2,5,9,J,K → keep K(4) and J(3), exchange 2,5,9 at 0,1,2.
    const hint = getDramahaHint(withHole(RAGS, { phase: DramahaPhase.DRAW }));
    expect(hint?.targetIndices).toEqual([0, 1, 2]);
  });

  it('still advises an all-in seat, which has no bet left to make but does still draw', () => {
    const state = withHole(RAGS, { phase: DramahaPhase.DRAW });
    state.players[0].allIn = true;
    expect(getDramahaHint(state)?.targetAction).toBe('draw');
  });

  it('advises the draw round even though no betting is legal there', () => {
    // The Hold'em base hint reasons about pot odds and would answer for a
    // betting street; the draw round has to be answered on its own terms.
    const hint = getDramahaHint(
      withHole(RAGS, { phase: DramahaPhase.DRAW, equity: { winProbability: 0.9, handOdds: [] }, potOdds: 0.1 }),
    );
    expect(hint?.targetAction).toBe('draw');
  });
});

describe('dramahaDiscardSuggestion', () => {
  it('plays an ace high, so a wheel-looking hand keeps the ace', () => {
    // A,3,4,7,9 — nothing pairs. Ace (14) and 9 are the two highest.
    const hole = [c('SPADE', 1), c('HEART', 3), c('DIAMOND', 4), c('CLOVER', 7), c('SPADE', 9)];
    expect(dramahaDiscardSuggestion(hole)).toEqual([1, 2, 3]);
  });

  it('keeps every card of a two-pair holding', () => {
    const hole = [c('SPADE', 6), c('HEART', 6), c('DIAMOND', 10), c('CLOVER', 10), c('SPADE', 3)];
    expect(dramahaDiscardSuggestion(hole)).toEqual([4]);
  });

  it('returns nothing for a hand that is not five cards', () => {
    expect(dramahaDiscardSuggestion([c('SPADE', 6), c('HEART', 6)])).toEqual([]);
  });
});
