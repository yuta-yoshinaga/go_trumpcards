import { describe, expect, it } from 'vitest';
import type { TehonbikiResponse } from '../../types/games/tehonbiki';
import { getTehonbikiHint } from './tehonbikiHint';

const state = (over: Partial<TehonbikiResponse> = {}): TehonbikiResponse => ({
  phase: 0,
  numbers: [1],
  betType: 'single',
  bet: 0,
  result: 0,
  payout: 0,
  chips: 1000,
  roundNumber: 1,
  remainingCards: 6,
  gameEndFlag: false,
  payoutNum: 9,
  payoutDen: 2,
  message: '',
  ...over,
});

describe('getTehonbikiHint', () => {
  it('suggests betting only during the open betting phase', () => {
    expect(getTehonbikiHint(state())).toEqual({
      targetAction: 'bet',
      reason: 'frontendHint.tehonbiki',
      confidence: 'moderate',
    });
    expect(getTehonbikiHint(state({ phase: 1 }))).toBeNull();
    expect(getTehonbikiHint(state({ gameEndFlag: true }))).toBeNull();
  });
});
