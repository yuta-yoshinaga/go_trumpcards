import { describe, expect, it } from 'vitest';
import type { TehonbikiResponse } from '../../../types/games/tehonbiki';
import { formatTehonbikiState } from './tehonbikiFormatter';

const base: TehonbikiResponse = {
  phase: 0,
  numbers: [1, 4],
  betType: 'double',
  bet: 50,
  result: 0,
  payout: 0,
  chips: 1000,
  roundNumber: 2,
  remainingCards: 6,
  gameEndFlag: false,
  payoutNum: 9,
  payoutDen: 5,
  message: '',
};

describe('formatTehonbikiState', () => {
  it('formats the wager without inventing card or deck fields', () => {
    const out = formatTehonbikiState(base);
    expect(out).toContain('Tehonbiki');
    expect(out).toContain('Round: 2');
    expect(out).toContain('Wager: double 1,4');
    expect(out).not.toContain('Gate:');
  });

  it('formats the revealed parent card and payout', () => {
    const out = formatTehonbikiState({ ...base, phase: 1, parentCard: 4, result: 1, payout: 90 });
    expect(out).toContain('Parent card: 4');
    expect(out).toContain('Result: 1 (payout 90)');
  });
});
