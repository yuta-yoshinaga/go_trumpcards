import { describe, expect, it } from 'vitest';
import { razzHint } from './razzHint';

describe('razzHint', () => {
  it('returns null', () => {
    expect(razzHint({})).toBeNull();
  });
});
