import { describe, expect, it } from 'vitest';
import type { Card, CoincheTrickCard } from '../types/card';
import { coincheLegalPlayIndices, isCoincheLegalPlay } from './coincheLegal';

// Trump is HEART (suit number 3) throughout unless noted.
const TRUMP = 3;

/** Builds a trick card played by the given seat. */
function played(playerIdx: number, design: Card['design'], value: number): CoincheTrickCard {
  return { playerIdx, card: { design, value } };
}

describe('isCoincheLegalPlay', () => {
  it('allows any card when leading (empty trick)', () => {
    const hand: Card[] = [
      { design: 'SPADE', value: 7 },
      { design: 'HEART', value: 11 },
    ];
    for (const card of hand) {
      expect(isCoincheLegalPlay(card, hand, [], TRUMP, 0)).toBe(true);
    }
  });

  it('requires following a non-trump lead suit when the hand holds it', () => {
    const hand: Card[] = [
      { design: 'SPADE', value: 7 },
      { design: 'SPADE', value: 8 },
      { design: 'CLOVER', value: 9 },
      { design: 'HEART', value: 11 },
    ];
    const trick = [played(1, 'SPADE', 12)];
    expect(isCoincheLegalPlay(hand[0], hand, trick, TRUMP, 0)).toBe(true); // ♠7 follows
    expect(isCoincheLegalPlay(hand[1], hand, trick, TRUMP, 0)).toBe(true); // ♠8 follows
    expect(isCoincheLegalPlay(hand[2], hand, trick, TRUMP, 0)).toBe(false); // ♣9 illegal
    expect(isCoincheLegalPlay(hand[3], hand, trick, TRUMP, 0)).toBe(false); // ♥J illegal
  });

  it('forces a trump when void of the lead suit and holding trump (obligation à couper)', () => {
    const hand: Card[] = [
      { design: 'CLOVER', value: 9 },
      { design: 'HEART', value: 11 },
      { design: 'HEART', value: 7 },
    ];
    const trick = [played(1, 'SPADE', 12)]; // opponent leads, no trump yet
    expect(isCoincheLegalPlay(hand[0], hand, trick, TRUMP, 0)).toBe(false); // ♣9 illegal (must trump)
    expect(isCoincheLegalPlay(hand[1], hand, trick, TRUMP, 0)).toBe(true); // ♥J legal
    expect(isCoincheLegalPlay(hand[2], hand, trick, TRUMP, 0)).toBe(true); // ♥7 legal (no trump in trick yet)
  });

  it('forces over-trumping when able (obligation à monter)', () => {
    const hand: Card[] = [
      { design: 'HEART', value: 11 }, // trump rank 8 — beats 9
      { design: 'HEART', value: 7 }, // trump rank 1 — cannot beat
      { design: 'CLOVER', value: 5 },
    ];
    // Opponent 1 leads spade; opponent 3 trumps with ♥9 (trump rank 7). Seat 0 is void of spade.
    const trick = [played(1, 'SPADE', 12), played(3, 'HEART', 9)];
    expect(isCoincheLegalPlay(hand[0], hand, trick, TRUMP, 0)).toBe(true); // ♥J over-trumps
    expect(isCoincheLegalPlay(hand[1], hand, trick, TRUMP, 0)).toBe(false); // ♥7 fails to over-trump
    expect(isCoincheLegalPlay(hand[2], hand, trick, TRUMP, 0)).toBe(false); // ♣5 illegal
  });

  it('waives the trump obligation when the partner is currently winning', () => {
    const hand: Card[] = [
      { design: 'CLOVER', value: 5 },
      { design: 'HEART', value: 11 },
    ];
    // Seat 0's partner is seat 2, who leads/holds the winning ♠A over seat 1's ♠7.
    const trick = [played(1, 'SPADE', 7), played(2, 'SPADE', 1)];
    for (const card of hand) {
      expect(isCoincheLegalPlay(card, hand, trick, TRUMP, 0)).toBe(true);
    }
  });

  it('allows any card when void of both the lead suit and trump', () => {
    const hand: Card[] = [
      { design: 'CLOVER', value: 5 },
      { design: 'DIAMOND', value: 9 },
    ];
    const trick = [played(1, 'SPADE', 12)];
    for (const card of hand) {
      expect(isCoincheLegalPlay(card, hand, trick, TRUMP, 0)).toBe(true);
    }
  });

  it('requires following a trump lead and over-trumping when able', () => {
    const hand: Card[] = [
      { design: 'HEART', value: 11 }, // trump rank 8
      { design: 'HEART', value: 8 }, // trump rank 2
      { design: 'CLOVER', value: 5 },
    ];
    const trick = [played(1, 'HEART', 7)]; // trump lead, rank 1
    expect(isCoincheLegalPlay(hand[0], hand, trick, TRUMP, 0)).toBe(true); // ♥J beats rank 1
    expect(isCoincheLegalPlay(hand[1], hand, trick, TRUMP, 0)).toBe(true); // ♥8 beats rank 1
    expect(isCoincheLegalPlay(hand[2], hand, trick, TRUMP, 0)).toBe(false); // ♣5 must follow trump
  });

  it('allows any card on a trump lead when void of trump', () => {
    const hand: Card[] = [
      { design: 'SPADE', value: 7 },
      { design: 'CLOVER', value: 9 },
    ];
    const trick = [played(1, 'HEART', 7)];
    for (const card of hand) {
      expect(isCoincheLegalPlay(card, hand, trick, TRUMP, 0)).toBe(true);
    }
  });
});

describe('coincheLegalPlayIndices', () => {
  it('returns every index when leading', () => {
    const hand: Card[] = [
      { design: 'SPADE', value: 7 },
      { design: 'HEART', value: 11 },
      { design: 'CLOVER', value: 9 },
    ];
    expect(coincheLegalPlayIndices(hand, [], TRUMP, 0)).toEqual([0, 1, 2]);
  });

  it('returns only the follow-suit indices', () => {
    const hand: Card[] = [
      { design: 'SPADE', value: 7 },
      { design: 'CLOVER', value: 9 },
      { design: 'SPADE', value: 8 },
    ];
    const trick = [played(1, 'SPADE', 12)];
    expect(coincheLegalPlayIndices(hand, trick, TRUMP, 0)).toEqual([0, 2]);
  });

  it('returns only the over-trumping index under the monter obligation', () => {
    const hand: Card[] = [
      { design: 'HEART', value: 11 },
      { design: 'HEART', value: 7 },
      { design: 'CLOVER', value: 5 },
    ];
    const trick = [played(1, 'SPADE', 12), played(3, 'HEART', 9)];
    expect(coincheLegalPlayIndices(hand, trick, TRUMP, 0)).toEqual([0]);
  });
});
