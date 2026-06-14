import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, RedDogResponse } from '../../../types/card';
import { RedDogPhase } from '../../../types/phases';
import { formatReddogState } from './reddogFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });

const betPhaseState: RedDogResponse = {
  initialCards: [],
  phase: RedDogPhase.BET,
  ante: 0,
  raise: 0,
  spread: 0,
  result: 0,
  totalPayout: 0,
  chips: 1000,
  message: '',
};

const spreadDecisionState: RedDogResponse = {
  ...betPhaseState,
  phase: RedDogPhase.SPREAD_DECISION,
  initialCards: [card('SPADE', 3), card('HEART', 10)],
  ante: 100,
  spread: 6,
  chips: 900,
};

const endPhaseWin: RedDogResponse = {
  ...betPhaseState,
  phase: RedDogPhase.END,
  initialCards: [card('SPADE', 3), card('HEART', 10)],
  thirdCard: card('DIAMOND', 7),
  ante: 100,
  raise: 100,
  spread: 6,
  result: 1,
  totalPayout: 400,
  chips: 1200,
  message: '勝利！',
};

const endPhaseLose: RedDogResponse = {
  ...betPhaseState,
  phase: RedDogPhase.END,
  initialCards: [card('SPADE', 3), card('HEART', 10)],
  thirdCard: card('CLOVER', 2),
  ante: 100,
  spread: 6,
  result: -1,
  totalPayout: 0,
  chips: 900,
  message: '敗北',
};

describe('formatReddogState', () => {
  it('formats BET phase with chips and phase name, no cards', () => {
    const result = formatReddogState(betPhaseState);
    expect(result).toContain('chips: 1000');
    expect(result).toContain('phase: BET');
    expect(result).not.toContain('Initial');
    expect(result).not.toContain('Third');
  });

  it('omits ante line when ante is zero', () => {
    const result = formatReddogState(betPhaseState);
    expect(result).not.toContain('ante:');
  });

  it('formats SPREAD DECISION phase with initial cards and spread', () => {
    const result = formatReddogState(spreadDecisionState);
    expect(result).toContain('phase: SPREAD DECISION');
    expect(result).toContain('Initial:');
    expect(result).toContain('spread: 6');
    expect(result).toContain('ante: 100');
    expect(result).not.toContain('raise:');
  });

  it('formats END phase with result, payout, third card, and message', () => {
    const result = formatReddogState(endPhaseWin);
    expect(result).toContain('phase: END');
    expect(result).toContain('Third:');
    expect(result).toContain('result: WIN');
    expect(result).toContain('payout: 400');
    expect(result).toContain('spread: 6');
    expect(result).toContain('ante: 100');
    expect(result).toContain('raise: 100');
    expect(result).toContain('勝利！');
  });

  it('formats END phase loss', () => {
    const result = formatReddogState(endPhaseLose);
    expect(result).toContain('result: LOSE');
    expect(result).toContain('payout: 0');
    expect(result).toContain('敗北');
  });

  it('formats unknown phase gracefully', () => {
    const state = { ...betPhaseState, phase: 99 };
    const result = formatReddogState(state);
    expect(result).toContain('phase: UNKNOWN');
  });

  it('omits message line when message is empty', () => {
    const result = formatReddogState(betPhaseState);
    expect(result).not.toContain('undefined');
    expect(result).not.toContain('null');
  });

  it('does not show spread in non-spread phases', () => {
    const state: RedDogResponse = {
      ...betPhaseState,
      phase: RedDogPhase.INITIAL_DEALT,
      initialCards: [card('SPADE', 5), card('HEART', 5)],
      spread: 0,
    };
    const result = formatReddogState(state);
    expect(result).not.toContain('spread:');
  });

  it('shows push result', () => {
    const state: RedDogResponse = {
      ...betPhaseState,
      phase: RedDogPhase.END,
      result: 0,
      totalPayout: 100,
    };
    const result = formatReddogState(state);
    expect(result).toContain('result: PUSH');
  });
});
