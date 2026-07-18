import { describe, expect, it } from 'vitest';
import type { Card, CourtPieceTrickCard } from '../types/card';
import { courtPieceLegalPlayIndices, isCourtPieceLegalPlay } from './courtPieceLegal';

const hand: Card[] = [
  { design: 'HEART', value: 12 },
  { design: 'HEART', value: 13 },
  { design: 'SPADE', value: 1 },
  { design: 'CLOVER', value: 5 },
];

/** Builds a lead trick card of the given suit design. */
function lead(design: Card['design']): CourtPieceTrickCard {
  return { playerIdx: 1, card: { design, value: 7 } };
}

describe('isCourtPieceLegalPlay', () => {
  it('allows any card when leading (empty trick)', () => {
    for (const card of hand) {
      expect(isCourtPieceLegalPlay(card, hand, [])).toBe(true);
    }
  });

  it('requires following the lead suit when the hand holds it', () => {
    const trick = [lead('HEART')];
    expect(isCourtPieceLegalPlay(hand[0], hand, trick)).toBe(true); // ♥Q follows
    expect(isCourtPieceLegalPlay(hand[1], hand, trick)).toBe(true); // ♥K follows
    expect(isCourtPieceLegalPlay(hand[2], hand, trick)).toBe(false); // ♠A illegal
    expect(isCourtPieceLegalPlay(hand[3], hand, trick)).toBe(false); // ♣5 illegal
  });

  it('allows any card when void in the lead suit', () => {
    const trick = [lead('DIAMOND')]; // hand has no diamonds
    for (const card of hand) {
      expect(isCourtPieceLegalPlay(card, hand, trick)).toBe(true);
    }
  });
});

describe('courtPieceLegalPlayIndices', () => {
  it('returns every index when leading', () => {
    expect(courtPieceLegalPlayIndices(hand, [])).toEqual([0, 1, 2, 3]);
  });

  it('returns only the lead-suit indices when the hand can follow', () => {
    expect(courtPieceLegalPlayIndices(hand, [lead('HEART')])).toEqual([0, 1]);
  });

  it('returns every index when void in the lead suit', () => {
    expect(courtPieceLegalPlayIndices(hand, [lead('DIAMOND')])).toEqual([0, 1, 2, 3]);
  });
});
