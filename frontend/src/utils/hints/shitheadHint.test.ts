import { describe, expect, it } from 'vitest';
import type { ShitheadResponse } from '../../types/card';
import { getShitheadHint } from './shitheadHint';

describe('getShitheadHint', () => {
  it('returns null because Shithead hints are not yet implemented client-side', () => {
    const state = { currentTurn: 0 } as unknown as ShitheadResponse;
    expect(getShitheadHint(state)).toBeNull();
  });
});
