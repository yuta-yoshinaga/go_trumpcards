import { describe, expect, it } from 'vitest';
import type { CardDesign, SpideretteTableauCard } from '../types/card';
import golden from './__fixtures__/spideretteMoves.golden.json';
import { spideretteCanSelectSource } from './spideretteMoves';

const DESIGNS: CardDesign[] = ['JOKER', 'SPADE', 'CLOVER', 'HEART', 'DIAMOND'];

describe('spideretteCanSelectSource golden vectors (shared with the Go domain)', () => {
  it('has vectors to check', () => {
    expect(golden.cases.length).toBeGreaterThan(0);
  });

  it.each(golden.cases)('$name', (c) => {
    const column: SpideretteTableauCard[] = c.cards.map((card) => ({
      card: { design: DESIGNS[card.suit], value: card.value },
      faceUp: card.faceUp,
    }));
    expect(spideretteCanSelectSource(column, c.sourceIndex)).toBe(c.valid);
  });
});
