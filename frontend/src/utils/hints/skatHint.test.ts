import { describe, expect, it } from 'vitest';
import type { SkatResponse } from '../../types/card';
import { getSkatHint } from './skatHint';

describe('getSkatHint', () => {
  it('returns null because Skat hints are produced server-side', () => {
    const state = { phase: 0 } as unknown as SkatResponse;
    expect(getSkatHint(state)).toBeNull();
  });
});
