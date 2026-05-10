import { describe, expect, it } from 'vitest';
import type { BlackJackSwitchResponse } from '../../types/card';
import { getBlackjackswitchHint } from './blackjackswitchHint';

const baseState: BlackJackSwitchResponse = {
  hands: [],
  dealerCards: [],
  dealerScore: 0,
  phase: 1,
  currentHandIdx: 0,
  chips: 1000,
  switched: false,
  dealerPushed22: false,
  overallResult: 0,
  totalPayout: 0,
  message: '',
};

describe('getBlackjackswitchHint', () => {
  it('returns null in BET phase (no decision yet)', () => {
    expect(getBlackjackswitchHint({ ...baseState, phase: 1 })).toBeNull();
  });

  it('returns null in SWITCH phase (TODO: replace with strategy table)', () => {
    expect(getBlackjackswitchHint({ ...baseState, phase: 2 })).toBeNull();
  });

  it('returns null in ACTION phase (TODO: replace with strategy table)', () => {
    expect(getBlackjackswitchHint({ ...baseState, phase: 3 })).toBeNull();
  });

  it('returns null in END phase (no decision after game ends)', () => {
    expect(getBlackjackswitchHint({ ...baseState, phase: 4 })).toBeNull();
  });
});
