import { describe, expect, it } from 'vitest';
import type { BidWhistResponse } from '../../types/card';
import { getBidWhistHint } from './bidwhistHint';

describe('getBidWhistHint', () => {
  it('returns null (client-side hints are not modelled for Bid Whist)', () => {
    const state = { phase: 3, players: [], currentTrick: [] } as unknown as BidWhistResponse;
    expect(getBidWhistHint(state)).toBeNull();
  });
});
