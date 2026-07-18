import { describe, expect, it } from 'vitest';
import type { Card, MichiganBoodle } from '../types/card';
import { michiganBoodleGuides } from './michiganBoodleGuide';

const boodles: MichiganBoodle[] = [
  { card: { design: 'HEART', value: 1 }, chips: 2, claimedBy: -1 }, // A♥
  { card: { design: 'CLOVER', value: 13 }, chips: 2, claimedBy: 1 }, // K♣ (claimed)
  { card: { design: 'DIAMOND', value: 12 }, chips: 2, claimedBy: -1 }, // Q♦
  { card: { design: 'SPADE', value: 11 }, chips: 2, claimedBy: -1 }, // J♠
];

describe('michiganBoodleGuides', () => {
  it('marks a boodle collectible only when the hand holds its exact card', () => {
    const hand: Card[] = [
      { design: 'HEART', value: 1 }, // matches boodle 0
      { design: 'DIAMOND', value: 12 }, // matches boodle 2
      { design: 'SPADE', value: 2 }, // no boodle match (wrong value)
    ];
    const guides = michiganBoodleGuides(boodles, hand);
    expect(guides.map((g) => g.collectible)).toEqual([true, false, true, false]);
  });

  it('flags a boodle as claimed when claimedBy is a seat index', () => {
    const guides = michiganBoodleGuides(boodles, []);
    expect(guides.map((g) => g.claimed)).toEqual([false, true, false, false]);
  });

  it('reports no collectible boodles for an empty hand', () => {
    const guides = michiganBoodleGuides(boodles, []);
    expect(guides.every((g) => !g.collectible)).toBe(true);
  });

  it('requires both matching design and value (design-only near-miss is not collectible)', () => {
    const hand: Card[] = [{ design: 'HEART', value: 5 }]; // HEART but not A♥
    const guides = michiganBoodleGuides(boodles, hand);
    expect(guides[0]?.collectible).toBe(false);
  });
});
