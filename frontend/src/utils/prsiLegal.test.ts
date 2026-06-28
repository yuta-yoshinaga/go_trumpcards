import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { isPrsiLegalPlay } from './prsiLegal';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('isPrsiLegalPlay', () => {
  const top = c('HEART', 9);

  it('allows a matching suit against the discard top', () => {
    expect(isPrsiLegalPlay(c('HEART', 10), top, 0)).toBe(true);
  });

  it('allows a matching rank against the discard top', () => {
    expect(isPrsiLegalPlay(c('SPADE', 9), top, 0)).toBe(true);
  });

  it('rejects a card matching neither suit nor rank', () => {
    expect(isPrsiLegalPlay(c('SPADE', 10), top, 0)).toBe(false);
  });

  it('allows any card on the opening play (no discard top)', () => {
    expect(isPrsiLegalPlay(c('SPADE', 10), null, 0)).toBe(true);
  });

  describe('under an active 7-stack penalty', () => {
    it('allows only a 7', () => {
      expect(isPrsiLegalPlay(c('SPADE', 7), top, 2)).toBe(true);
      expect(isPrsiLegalPlay(c('DIAMOND', 7), top, 4)).toBe(true);
    });

    it('rejects a non-7 even when suit or rank matches', () => {
      expect(isPrsiLegalPlay(c('HEART', 10), top, 2)).toBe(false); // matching suit but not a 7
      expect(isPrsiLegalPlay(c('SPADE', 9), top, 2)).toBe(false); // matching rank but not a 7
    });

    it('rejects a non-7 even with no discard top', () => {
      expect(isPrsiLegalPlay(c('SPADE', 10), null, 2)).toBe(false);
    });
  });
});
