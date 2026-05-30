import { describe, expect, it } from 'vitest';
import type { ChinesePokerResponse } from '../../types/card';
import { chinesepokerHint } from './chinesepokerHint';

describe('chinesepokerHint', () => {
  it('returns null (stub)', () => {
    expect(chinesepokerHint({} as ChinesePokerResponse)).toBeNull();
  });
});
