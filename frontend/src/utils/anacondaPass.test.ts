import { describe, expect, it } from 'vitest';
import golden from './__fixtures__/anacondaPassRecipient.golden.json';
import { anacondaPassRecipient } from './anacondaPass';

describe('anacondaPassRecipient golden vectors (shared with the CUI presenter)', () => {
  it('has vectors to check', () => {
    expect(golden.cases.length).toBeGreaterThan(0);
  });

  it.each(golden.cases)('$name', (c) => {
    const seats = c.seats.map((s) => ({ isHuman: s.human, out: s.out }));

    expect(anacondaPassRecipient(seats)).toBe(c.recipient);
  });
});
