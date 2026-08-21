import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { stalactitesFoundationTarget } from './stalactitesFoundationTarget';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const emptyFoundation = (): Card[][] => [[], [], [], []];

// Stalactites ignores suit, starts every pile at the deal's base rank (not Ace)
// and wraps King -> Ace. The FreeCell version this was cloned from got all three
// wrong; each is pinned below with a negative control.
describe('stalactitesFoundationTarget', () => {
  it('opens the first empty pile with the base rank, whatever the suit', () => {
    for (const design of ['SPADE', 'CLOVER', 'HEART', 'DIAMOND'] as CardDesign[]) {
      expect(stalactitesFoundationTarget(card(design, 7), emptyFoundation(), 7)).toEqual({
        zone: 'foundation',
        col: 0,
      });
    }
  });

  it('rejects any rank other than the base rank on an empty pile', () => {
    // An Ace is not special here: FreeCell would have accepted it.
    expect(stalactitesFoundationTarget(card('SPADE', 1), emptyFoundation(), 7)).toBeNull();
    expect(stalactitesFoundationTarget(card('SPADE', 8), emptyFoundation(), 7)).toBeNull();
  });

  it('continues a pile with the next rank up, ignoring suit', () => {
    const foundation: Card[][] = [[card('SPADE', 7)], [], [], []];
    // A heart continues a spade pile, because suit is not consulted at all.
    expect(stalactitesFoundationTarget(card('HEART', 8), foundation, 7)).toEqual({
      zone: 'foundation',
      col: 0,
    });
  });

  it('wraps past the King to the Ace', () => {
    const foundation: Card[][] = [[card('SPADE', 13)], [], [], []];
    expect(stalactitesFoundationTarget(card('CLOVER', 1), foundation, 12)).toEqual({
      zone: 'foundation',
      col: 0,
    });
    expect(stalactitesFoundationTarget(card('CLOVER', 2), foundation, 12)).toBeNull();
  });

  it('prefers continuing a pile over opening an empty one', () => {
    // Both are legal for an 8: pile 0 continues, pile 1 would open at base 8.
    const foundation: Card[][] = [[card('SPADE', 7)], [], [], []];
    expect(stalactitesFoundationTarget(card('HEART', 8), foundation, 8)).toEqual({
      zone: 'foundation',
      col: 0,
    });
  });

  it('rejects a rank that neither continues nor opens a pile', () => {
    const foundation: Card[][] = [[card('SPADE', 7)], [], [], []];
    expect(stalactitesFoundationTarget(card('HEART', 9), foundation, 7)).toBeNull();
    expect(stalactitesFoundationTarget(card('HEART', 6), foundation, 7)).toBeNull();
  });
});
