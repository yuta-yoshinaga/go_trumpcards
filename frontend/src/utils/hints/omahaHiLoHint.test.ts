import { describe, expect, it } from 'vitest';
import type { Card, OmahaResponse } from '../../types/card';
import { HoldemPhase } from '../../types/phases';
import { getOmahaHiLoHint } from './omahaHiLoHint';

/** Build a minimal OmahaResponse with sensible defaults that tests can override. */
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
    isHiLo: true,
    ...overrides,
  };
}

const c = (value: number): Card => ({ design: 'SPADE', value });
const cd = (value: number): Card => ({ design: 'DIAMOND', value });
const ch = (value: number): Card => ({ design: 'HEART', value });
const cc = (value: number): Card => ({ design: 'CLOVER', value });

describe('getOmahaHiLoHint', () => {
  it('returns null in SHOWDOWN phase', () => {
    expect(getOmahaHiLoHint(makeState({ phase: HoldemPhase.SHOWDOWN }))).toBeNull();
  });

  it('returns null when no human player', () => {
    const state = makeState({ players: [{ ...makeState().players[0], isHuman: false }] });
    expect(getOmahaHiLoHint(state)).toBeNull();
  });

  it('suggests raise with strong equity (no low data)', () => {
    const state = makeState({ equity: { winProbability: 0.8, handOdds: [] }, potOdds: 0.2 });
    expect(getOmahaHiLoHint(state)?.targetAction).toBe('raise');
  });

  it('suggests fold for weak hand rank without low draw', () => {
    expect(getOmahaHiLoHint(makeState())?.targetAction).toBe('fold');
  });

  it('suggests call for decent hand rank without low draw', () => {
    const state = makeState();
    state.players[0].handRank = 2;
    expect(getOmahaHiLoHint(state)?.targetAction).toBe('call');
  });

  it('falls back to base hint for non-Hi-Lo state', () => {
    const state = makeState({ isHiLo: false });
    state.players[0].handRank = 3;
    expect(getOmahaHiLoHint(state)?.reason).toBe('hint.strongHandRank');
  });

  it('suggests scoop when strong equity and low draw is viable', () => {
    const state = makeState({
      equity: { winProbability: 0.8, handOdds: [] },
      potOdds: 0.2,
      // Hole has A-2 (two unique low ranks); board has 3-4 (low cards already).
      communityCards: [c(3), cd(4), ch(13)],
    });
    state.players[0].cards = [c(1), cd(2), ch(11), cc(12)];
    const hint = getOmahaHiLoHint(state);
    expect(hint?.targetAction).toBe('raise');
    expect(hint?.reason).toBe('hint.scoopChance');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests scoop when strong hand rank and low draw is viable', () => {
    const state = makeState({
      // Board has 2 low ranks already; one more card to come can complete low.
      communityCards: [c(3), cd(4), ch(13), cc(11)],
    });
    state.players[0].cards = [c(1), cd(2), ch(11), cc(12)];
    state.players[0].handRank = 3; // Strong hi
    const hint = getOmahaHiLoHint(state);
    expect(hint?.targetAction).toBe('raise');
    expect(hint?.reason).toBe('hint.scoopChance');
  });

  it('suggests low draw call when high is weak but low draw is strong', () => {
    const state = makeState({
      // Three low ranks already on the board; low qualifies for sure.
      communityCards: [c(3), cd(4), ch(7), cc(13)],
    });
    state.players[0].cards = [c(1), cd(2), ch(11), cc(12)];
    state.players[0].handRank = 0; // Weak hi
    const hint = getOmahaHiLoHint(state);
    expect(hint?.targetAction).toBe('call');
    expect(hint?.reason).toBe('hint.lowDraw');
    expect(hint?.confidence).toBe('moderate');
  });

  it('falls back to base hint when hole has only 1 unique low rank', () => {
    const state = makeState({
      communityCards: [c(3), cd(4), ch(7)],
    });
    // Only one unique low rank in hole (Ace) — cannot form a low.
    state.players[0].cards = [c(1), cd(11), ch(12), cc(13)];
    state.players[0].handRank = 3; // Strong
    const hint = getOmahaHiLoHint(state);
    expect(hint?.reason).toBe('hint.strongHandRank');
  });

  it('falls back to base hint when no low possible (river with insufficient board low)', () => {
    // River = 5 community cards, only 1 low — low cannot qualify.
    const state = makeState({
      phase: HoldemPhase.RIVER,
      communityCards: [c(3), cd(11), ch(12), cc(13), c(10)],
    });
    state.players[0].cards = [c(1), cd(2), ch(11), cc(12)];
    state.players[0].handRank = 3;
    const hint = getOmahaHiLoHint(state);
    expect(hint?.reason).toBe('hint.strongHandRank');
  });

  it('counts paired low cards as a single rank', () => {
    // Hole: A-A — only one unique low rank (Ace), so cannot form a low.
    const state = makeState({
      communityCards: [c(3), cd(4), ch(7)],
    });
    state.players[0].cards = [c(1), cd(1), ch(11), cc(12)];
    state.players[0].handRank = 3;
    const hint = getOmahaHiLoHint(state);
    expect(hint?.reason).toBe('hint.strongHandRank');
  });
});
