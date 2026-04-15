import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, LetItRideResponse, MaskedCard } from '../../../types/card';
import { LetItRidePhase } from '../../../types/phases';
import { formatLetitrideState } from './letitrideFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const maskedCard: MaskedCard = { design: '', value: 0 };

const betPhaseState: LetItRideResponse = {
  playerHand: [],
  communityCards: [],
  phase: LetItRidePhase.BET,
  chips: 1000,
  betAmount: 0,
  bet1Active: true,
  bet2Active: true,
  bet3Active: true,
  result: 0,
  handRank: 0,
  bet1Payout: 0,
  bet2Payout: 0,
  bet3Payout: 0,
  totalPayout: 0,
  message: '',
};

const firstDecisionState: LetItRideResponse = {
  ...betPhaseState,
  phase: LetItRidePhase.FIRST_DECISION,
  playerHand: [card('SPADE', 1), card('HEART', 10), card('DIAMOND', 11)],
  communityCards: [maskedCard, maskedCard],
  betAmount: 100,
  chips: 700,
};

const secondDecisionState: LetItRideResponse = {
  ...firstDecisionState,
  phase: LetItRidePhase.SECOND_DECISION,
  communityCards: [card('CLOVER', 12), maskedCard],
};

const endPhaseWin: LetItRideResponse = {
  ...betPhaseState,
  phase: LetItRidePhase.END,
  playerHand: [card('SPADE', 1), card('HEART', 10), card('DIAMOND', 11)],
  communityCards: [card('CLOVER', 12), card('SPADE', 13)],
  betAmount: 100,
  chips: 1300,
  result: 1,
  handRank: 9,
  bet1Active: true,
  bet2Active: true,
  bet3Active: true,
  bet1Payout: 100000,
  bet2Payout: 100000,
  bet3Payout: 100000,
  totalPayout: 300000,
  message: '勝利！',
};

describe('formatLetitrideState', () => {
  it('formats BET phase with chips and phase name, no cards', () => {
    const result = formatLetitrideState(betPhaseState);
    expect(result).toContain('chips: 1000');
    expect(result).toContain('phase: BET');
    expect(result).not.toContain('Your hand');
    expect(result).not.toContain('Community');
  });

  it('omits bet per spot line when betAmount is zero', () => {
    const result = formatLetitrideState(betPhaseState);
    expect(result).not.toContain('bet per spot');
  });

  it('formats FIRST DECISION phase with player hand and masked community cards', () => {
    const result = formatLetitrideState(firstDecisionState);
    expect(result).toContain('phase: FIRST DECISION');
    expect(result).toContain('Your hand:');
    expect(result).toContain('Community:');
    expect(result).toContain('??');
    expect(result).toContain('bet per spot: 100');
    expect(result).toContain('bet1: active');
    expect(result).toContain('bet2: active');
    expect(result).toContain('bet3: active');
  });

  it('formats SECOND DECISION phase with mixed community cards', () => {
    const result = formatLetitrideState(secondDecisionState);
    expect(result).toContain('phase: SECOND DECISION');
    // First community card is revealed (♣Q), second is masked
    expect(result).toContain('\u2663Q');
    expect(result).toContain('??');
  });

  it('formats END phase with payouts and fully revealed community cards', () => {
    const result = formatLetitrideState(endPhaseWin);
    expect(result).toContain('phase: END');
    expect(result).toContain('\u2663Q');
    expect(result).toContain('\u2660K');
    expect(result).not.toContain('??');
    expect(result).toContain('payout: bet1=100000 bet2=100000 bet3=100000');
    expect(result).toContain('total: 300000');
    expect(result).toContain('勝利！');
  });

  it('shows bet pulled status when bets are withdrawn', () => {
    const state: LetItRideResponse = {
      ...firstDecisionState,
      bet1Active: false,
      bet2Active: false,
    };
    const result = formatLetitrideState(state);
    expect(result).toContain('bet1: pulled');
    expect(result).toContain('bet2: pulled');
    expect(result).toContain('bet3: active');
  });

  it('formats unknown phase gracefully', () => {
    const state = { ...betPhaseState, phase: 99 };
    const result = formatLetitrideState(state);
    expect(result).toContain('phase: UNKNOWN');
  });

  it('omits message line when message is empty', () => {
    const result = formatLetitrideState(betPhaseState);
    // The empty message branch: no extra undefined/null line
    expect(result).not.toContain('undefined');
    expect(result).not.toContain('null');
  });

  it('includes message when present', () => {
    const state: LetItRideResponse = { ...betPhaseState, message: 'テストメッセージ' };
    const result = formatLetitrideState(state);
    expect(result).toContain('テストメッセージ');
  });
});
