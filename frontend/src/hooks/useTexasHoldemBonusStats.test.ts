import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import {
  outcomeFromTexasHoldemBonusResult,
  TEXASHOLDEMBONUS_HISTORY_KEY,
  TEXASHOLDEMBONUS_HISTORY_MAX,
  TexasHoldemBonusOutcome,
  tallyTexasHoldemBonusHistory,
  useTexasHoldemBonusStats,
} from './useTexasHoldemBonusStats';

beforeEach(() => localStorage.clear());

describe('useTexasHoldemBonusStats', () => {
  it('counts wins, losses, pushes and net', () => {
    expect(outcomeFromTexasHoldemBonusResult(1)).toBe(TexasHoldemBonusOutcome.WIN);
    expect(outcomeFromTexasHoldemBonusResult(-1)).toBe(TexasHoldemBonusOutcome.LOSS);
    expect(outcomeFromTexasHoldemBonusResult(0)).toBe(TexasHoldemBonusOutcome.PUSH);
    expect(
      tallyTexasHoldemBonusHistory([
        { outcome: 1, net: 200 },
        { outcome: 2, net: -100 },
        { outcome: 0, net: 0 },
      ]),
    ).toEqual({
      wins: 1,
      losses: 1,
      pushes: 1,
      hands: 3,
      net: 100,
    });
  });

  it('persists records and caps history at 200 rounds', () => {
    const { result } = renderHook(() => useTexasHoldemBonusStats());
    act(() => {
      for (let i = 0; i < TEXASHOLDEMBONUS_HISTORY_MAX + 1; i++) {
        result.current.recordRound({ outcome: TexasHoldemBonusOutcome.LOSS, net: -10 });
      }
    });
    expect(result.current.history).toHaveLength(TEXASHOLDEMBONUS_HISTORY_MAX);
    expect(JSON.parse(localStorage.getItem(TEXASHOLDEMBONUS_HISTORY_KEY) ?? '[]')).toHaveLength(200);
    expect(result.current.tally.losses).toBe(200);
  });
});
