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

  it.each([
    ['negative index', [], -1, false],
    ['index past the end', [], 0, false],
    ['empty column', [], 0, false],
    ['last card in a sequence', [{ card: { design: 'SPADE' as CardDesign, value: 13 }, faceUp: true }], 0, true],
    ['face-up card without a card value', [{ card: null, faceUp: true }], 0, false],
  ])('%s handles early-return and boundary inputs', (_name, column, cardIndex, expected) => {
    expect(spideretteCanSelectSource(column, cardIndex)).toBe(expected);
  });
});
