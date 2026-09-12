import { describe, expect, it } from 'vitest';
import type { CardDesign } from '../types/card';
import golden from './__fixtures__/contractRummyLayoff.golden.json';
import { contractRummyCanAddToMeld } from './contractRummyLayoff';

const DESIGNS: CardDesign[] = ['SPADE', 'CLOVER', 'HEART', 'DIAMOND'];

describe('contractRummyCanAddToMeld golden vectors (shared with the Go domain)', () => {
  it.each(golden.cases)('$name', (c) => {
    const meld = c.meld.map((card) => ({ design: DESIGNS[card.suit], value: card.value }));
    const card = { design: DESIGNS[c.card.suit], value: c.card.value };
    expect(contractRummyCanAddToMeld(meld, card)).toBe(c.valid);
  });
});
