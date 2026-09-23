import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { penguinFoundationTarget } from './penguinFoundationTarget';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const emptyFoundation = (): Card[][] => [[], [], [], []];

describe('penguinFoundationTarget', () => {
  it('sends only the base-rank card to an empty foundation pile', () => {
    expect(penguinFoundationTarget(card('SPADE', 4), emptyFoundation(), 4)).toEqual({ zone: 'foundation', col: 0 });
  });

  it('rejects an Ace on an empty pile when baseRank is not Ace', () => {
    expect(penguinFoundationTarget(card('SPADE', 1), emptyFoundation(), 4)).toBeNull();
  });

  it('wraps from King to Ace on a matching-suit pile', () => {
    const foundation = [[card('SPADE', 13)], [], [], []];
    expect(penguinFoundationTarget(card('SPADE', 1), foundation, 4)).toEqual({ zone: 'foundation', col: 0 });
  });

  it('rejects a card with a different suit from the pile top', () => {
    const foundation = [[], [], [card('SPADE', 4)], []];
    expect(penguinFoundationTarget(card('HEART', 5), foundation, 4)).toBeNull();
  });

  it('rejects a card with no legal foundation target', () => {
    expect(penguinFoundationTarget(card('CLOVER', 7), emptyFoundation(), 4)).toBeNull();
  });

  it('rejects a joker because it has no foundation index', () => {
    expect(penguinFoundationTarget(card('JOKER', 1), emptyFoundation(), 4)).toBeNull();
  });
});
