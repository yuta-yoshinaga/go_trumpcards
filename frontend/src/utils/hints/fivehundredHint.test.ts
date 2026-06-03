import { describe, expect, it } from 'vitest';
import type { FiveHundredResponse } from '../../types/card';
import { getFiveHundredHint } from './fivehundredHint';

describe('getFiveHundredHint', () => {
  it('returns null (client-side hints are not modelled for 500)', () => {
    const state = { phase: 2, players: [], currentTrick: [] } as unknown as FiveHundredResponse;
    expect(getFiveHundredHint(state)).toBeNull();
  });
});
