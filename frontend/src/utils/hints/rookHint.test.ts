import { describe, expect, it } from 'vitest';
import type { RookResponse } from '../../types/card';
import { getRookHint } from './rookHint';

describe('getRookHint', () => {
  it('returns null (client-side hints are not modelled for Rook)', () => {
    const state = { phase: 2, players: [], currentTrick: [] } as unknown as RookResponse;
    expect(getRookHint(state)).toBeNull();
  });
});
