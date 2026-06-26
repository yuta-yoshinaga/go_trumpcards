import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, DragonTigerResponse } from '../../../types/card';
import { DragonTigerBetType, DragonTigerPhase } from '../../../types/phases';
import { formatDragonTigerState } from './dragontigerFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });

const baseState: DragonTigerResponse = {
  phase: DragonTigerPhase.BET,
  chips: 1000,
  betAmount: 0,
  betType: DragonTigerBetType.DRAGON,
  result: 0,
  payout: 0,
  history: [],
  message: '',
};

describe('formatDragonTigerState', () => {
  it('renders header, chips, and phase', () => {
    const out = formatDragonTigerState(baseState);
    expect(out).toContain('Dragon Tiger');
    expect(out).toContain('chips: 1000');
    expect(out).toContain('phase: BET');
  });

  it('shows the active bet target and amount', () => {
    const out = formatDragonTigerState({ ...baseState, betAmount: 50, betType: DragonTigerBetType.TIGER });
    expect(out).toContain('bet: 50 on Tiger');
  });

  it('shows both cards and the result/payout at end', () => {
    const out = formatDragonTigerState({
      ...baseState,
      phase: DragonTigerPhase.END,
      dragonCard: card('SPADE', 13),
      tigerCard: card('HEART', 7),
      result: 1,
      payout: 200,
    });
    expect(out).toContain('Dragon: ♠K');
    expect(out).toContain('Tiger: ♥7');
    expect(out).toContain('result: Dragon  payout: 200');
  });

  it('maps the wire result value (1/-1/0) to Dragon/Tiger/Tie', () => {
    const end = { ...baseState, phase: DragonTigerPhase.END };
    expect(formatDragonTigerState({ ...end, result: 1 })).toContain('result: Dragon');
    expect(formatDragonTigerState({ ...end, result: -1 })).toContain('result: Tiger');
    expect(formatDragonTigerState({ ...end, result: 0 })).toContain('result: Tie');
  });

  it('renders the big-road history with target names', () => {
    const out = formatDragonTigerState({ ...baseState, history: [0, 1, 2] });
    expect(out).toContain('history: Dragon Tiger Tie');
  });

  it('renders UNKNOWN for an unexpected phase', () => {
    expect(formatDragonTigerState({ ...baseState, phase: 99 })).toContain('phase: UNKNOWN');
  });
});
