import { describe, expect, it } from 'vitest';
import type { TexasHoldemBonusResponse } from '../../types/card';
import { TexasHoldemBonusPhase } from '../../types/phases';
import { getTexasHoldemBonusHint } from './texasHoldemBonusHint';

const baseState: TexasHoldemBonusResponse = {
  playerHand: [],
  dealerHand: [],
  community: [],
  phase: TexasHoldemBonusPhase.BET,
  chips: 1000,
  anteBet: 0,
  bonusBet: 0,
  flopBet: 0,
  turnBet: 0,
  riverBet: 0,
  totalPlayBet: 0,
  result: 0,
  antePayout: 0,
  playPayout: 0,
  bonusPayout: 0,
  totalPayout: 0,
  playerHandRank: 0,
  dealerHandRank: 0,
  message: '',
};

describe('getTexasHoldemBonusHint', () => {
  it('returns null in BET phase', () => {
    expect(getTexasHoldemBonusHint(baseState)).toBeNull();
  });

  it('returns null when playerHand is empty', () => {
    expect(getTexasHoldemBonusHint({ ...baseState, phase: TexasHoldemBonusPhase.PRE_FLOP })).toBeNull();
  });

  it('returns play for pocket pair pre-flop', () => {
    const state = {
      ...baseState,
      phase: TexasHoldemBonusPhase.PRE_FLOP,
      playerHand: [
        { design: 'SPADE' as const, value: 8 },
        { design: 'CLOVER' as const, value: 8 },
      ],
    };
    expect(getTexasHoldemBonusHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.pocketPair',
      confidence: 'strong',
    });
  });

  it('returns play when one card is an Ace pre-flop', () => {
    const state = {
      ...baseState,
      phase: TexasHoldemBonusPhase.PRE_FLOP,
      playerHand: [
        { design: 'SPADE' as const, value: 1 },
        { design: 'CLOVER' as const, value: 5 },
      ],
    };
    const hint = getTexasHoldemBonusHint(state);
    expect(hint?.targetAction).toBe('play');
    expect(hint?.reason).toBe('hint.acePlay');
  });

  it('returns play for suited broadway pre-flop', () => {
    const state = {
      ...baseState,
      phase: TexasHoldemBonusPhase.PRE_FLOP,
      playerHand: [
        { design: 'SPADE' as const, value: 13 },
        { design: 'SPADE' as const, value: 12 },
      ],
    };
    // K-Q has Ace? No, neither is Ace. Both broadway, suited.
    const hint = getTexasHoldemBonusHint(state);
    expect(hint?.reason).toBe('hint.suitedBroadway');
    expect(hint?.confidence).toBe('moderate');
  });

  it('returns play for unsuited broadway pre-flop', () => {
    const state = {
      ...baseState,
      phase: TexasHoldemBonusPhase.PRE_FLOP,
      playerHand: [
        { design: 'SPADE' as const, value: 13 },
        { design: 'CLOVER' as const, value: 12 },
      ],
    };
    const hint = getTexasHoldemBonusHint(state);
    expect(hint?.reason).toBe('hint.broadwayCards');
  });

  it('returns fold for weak unsuited non-broadway pre-flop', () => {
    const state = {
      ...baseState,
      phase: TexasHoldemBonusPhase.PRE_FLOP,
      playerHand: [
        { design: 'SPADE' as const, value: 7 },
        { design: 'CLOVER' as const, value: 2 },
      ],
    };
    const hint = getTexasHoldemBonusHint(state);
    expect(hint?.targetAction).toBe('fold');
    expect(hint?.reason).toBe('hint.weakHand');
  });

  it('returns raise on flop with pair or better', () => {
    const state = {
      ...baseState,
      phase: TexasHoldemBonusPhase.FLOP,
      playerHand: [
        { design: 'SPADE' as const, value: 1 },
        { design: 'CLOVER' as const, value: 1 },
      ],
      playerHandRank: 1,
    };
    const hint = getTexasHoldemBonusHint(state);
    expect(hint?.targetAction).toBe('raise');
  });

  it('returns check on turn with no pair', () => {
    const state = {
      ...baseState,
      phase: TexasHoldemBonusPhase.TURN,
      playerHand: [
        { design: 'SPADE' as const, value: 7 },
        { design: 'CLOVER' as const, value: 2 },
      ],
      playerHandRank: 0,
    };
    const hint = getTexasHoldemBonusHint(state);
    expect(hint?.targetAction).toBe('check');
  });

  it('returns null in END phase', () => {
    const state = {
      ...baseState,
      phase: TexasHoldemBonusPhase.END,
      playerHand: [
        { design: 'SPADE' as const, value: 1 },
        { design: 'CLOVER' as const, value: 1 },
      ],
    };
    expect(getTexasHoldemBonusHint(state)).toBeNull();
  });
});
