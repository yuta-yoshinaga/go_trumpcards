import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { DOPPELKOPF_TRUMP_ORDER, isDoppelkopfTrump } from './doppelkopfTrump';

const card = (design: Card['design'], value: number): Card => ({ design, value });

describe('isDoppelkopfTrump', () => {
  it('treats every Diamond as a trump', () => {
    for (const v of [1, 9, 10, 11, 12, 13]) {
      expect(isDoppelkopfTrump(card('DIAMOND', v))).toBe(true);
    }
  });

  it('treats every Queen and Jack as a trump regardless of suit', () => {
    for (const d of ['SPADE', 'CLOVER', 'HEART', 'DIAMOND'] as const) {
      expect(isDoppelkopfTrump(card(d, 12))).toBe(true); // Queen
      expect(isDoppelkopfTrump(card(d, 11))).toBe(true); // Jack
    }
  });

  it('treats the ♥10 (Dulle) as a trump', () => {
    expect(isDoppelkopfTrump(card('HEART', 10))).toBe(true);
  });

  it('treats fail cards as non-trump', () => {
    expect(isDoppelkopfTrump(card('SPADE', 1))).toBe(false); // ♠A
    expect(isDoppelkopfTrump(card('CLOVER', 10))).toBe(false); // ♣10
    expect(isDoppelkopfTrump(card('HEART', 1))).toBe(false); // ♥A
    expect(isDoppelkopfTrump(card('HEART', 13))).toBe(false); // ♥K
    expect(isDoppelkopfTrump(card('SPADE', 13))).toBe(false); // ♠K
  });
});

describe('DOPPELKOPF_TRUMP_ORDER', () => {
  it('lists the 13 distinct trump ranks strongest first', () => {
    expect(DOPPELKOPF_TRUMP_ORDER[0]).toBe('♥10');
    expect(DOPPELKOPF_TRUMP_ORDER.at(-1)).toBe('♦9');
    expect(DOPPELKOPF_TRUMP_ORDER).toHaveLength(13);
    expect(new Set(DOPPELKOPF_TRUMP_ORDER).size).toBe(13);
  });
});
