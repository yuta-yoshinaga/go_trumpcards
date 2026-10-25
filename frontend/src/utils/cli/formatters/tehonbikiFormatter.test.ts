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
  minBet: 10,
  maxBet: 500,
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
  it('formats only an explicitly requested server hint', () => {
    expect(formatTehonbikiState({ ...base, hint: { reason: 'Choose numbers 1, 2, and 3.' } })).not.toContain('HINT:');
    expect(
      formatTehonbikiState({
        ...base,
        hint: { reason: 'Choose numbers 1, 2, and 3.' },
        messageCode: 'tehonbiki.hintRequested',
      }),
    ).toContain('HINT: Choose numbers 1, 2, and 3.');
    expect(formatTehonbikiState({ ...base, messageCode: 'tehonbiki.noHint' })).not.toContain('HINT:');
  });

  it('prints server message text but never a raw message code', () => {
    expect(formatTehonbikiState({ ...base, message: 'Bet must be at least 10.' })).toContain(
      'Bet must be at least 10.',
    );
    expect(formatTehonbikiState({ ...base, message: '', messageCode: 'tehonbiki.noHint' })).not.toContain(
      'tehonbiki.noHint',
    );
  });

  it('formats the revealed parent card and payout', () => {
    const out = formatTehonbikiState({ ...base, phase: 1, parentCard: 4, result: 1, payout: 90 });
    expect(out).toContain('Parent card: 4');
    expect(out).toContain('Result: 1 (payout 90)');
  });
});
