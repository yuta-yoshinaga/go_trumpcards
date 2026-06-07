import { describe, expect, it } from 'vitest';
import type { SchnapsenResponse } from '../../types/card';
import { getSchnapsenHint } from './schnapsenHint';

describe('getSchnapsenHint', () => {
  it('returns null (authoritative hint is computed server-side)', () => {
    expect(getSchnapsenHint({} as SchnapsenResponse)).toBeNull();
  });
});
