import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import {
  type HeartsPlayContext,
  heartsIllegalReasonKey,
  heartsLegalPlayIndices,
  isHeartsLegalPlay,
} from './heartsLegal';

const c = (design: Card['design'], value: number): Card => ({ design, value });

const TWO_CLUBS = c('CLOVER', 2);
const QUEEN_SPADES = c('SPADE', 12);
const JACK_DIAMONDS = c('DIAMOND', 11);

/** Build a play context with sensible defaults. */
function ctx(overrides?: Partial<HeartsPlayContext>): HeartsPlayContext {
  return {
    currentTrick: [],
    heartsBroken: false,
    trickNumber: 2,
    omnibusJD: false,
    ...overrides,
  };
}

describe('isHeartsLegalPlay', () => {
  describe('first-trick lead', () => {
    const hand = [TWO_CLUBS, c('SPADE', 1), c('HEART', 5)];
    const firstLead = ctx({ trickNumber: 1, currentTrick: [] });

    it('allows only the 2♣ to lead the first trick', () => {
      expect(isHeartsLegalPlay(TWO_CLUBS, hand, firstLead)).toBe(true);
      expect(isHeartsLegalPlay(c('SPADE', 1), hand, firstLead)).toBe(false);
      expect(isHeartsLegalPlay(c('HEART', 5), hand, firstLead)).toBe(false);
    });

    it('does not force 2♣ when the hand lacks it (falls through to lead rules)', () => {
      const noClubs = [c('SPADE', 1), c('HEART', 5)];
      // No 2♣ held, so the 2♣ rule does not apply; the spade leads fine.
      expect(isHeartsLegalPlay(c('SPADE', 1), noClubs, firstLead)).toBe(true);
      // Hearts still cannot be led before broken while a non-heart remains.
      expect(isHeartsLegalPlay(c('HEART', 5), noClubs, firstLead)).toBe(false);
    });
  });

  describe('leading with hearts not broken', () => {
    it('forbids leading a heart while a non-heart remains', () => {
      const hand = [c('HEART', 5), c('SPADE', 9)];
      const lead = ctx({ currentTrick: [], heartsBroken: false });
      expect(isHeartsLegalPlay(c('HEART', 5), hand, lead)).toBe(false);
      expect(isHeartsLegalPlay(c('SPADE', 9), hand, lead)).toBe(true);
    });

    it('allows leading a heart when the hand is all hearts', () => {
      const hand = [c('HEART', 5), c('HEART', 9)];
      const lead = ctx({ currentTrick: [], heartsBroken: false });
      expect(isHeartsLegalPlay(c('HEART', 5), hand, lead)).toBe(true);
    });

    it('allows leading a heart once hearts are broken', () => {
      const hand = [c('HEART', 5), c('SPADE', 9)];
      const lead = ctx({ currentTrick: [], heartsBroken: true });
      expect(isHeartsLegalPlay(c('HEART', 5), hand, lead)).toBe(true);
    });
  });

  describe('following suit', () => {
    const trick = [{ playerIdx: 0, card: c('DIAMOND', 3) }];

    it('requires following the lead suit when able', () => {
      const hand = [c('DIAMOND', 8), c('SPADE', 9)];
      const following = ctx({ currentTrick: trick });
      expect(isHeartsLegalPlay(c('DIAMOND', 8), hand, following)).toBe(true);
      expect(isHeartsLegalPlay(c('SPADE', 9), hand, following)).toBe(false);
    });

    it('allows any card when void in the lead suit (later tricks)', () => {
      const hand = [c('SPADE', 9), c('HEART', 5)];
      const following = ctx({ currentTrick: trick, trickNumber: 3 });
      expect(isHeartsLegalPlay(c('SPADE', 9), hand, following)).toBe(true);
      expect(isHeartsLegalPlay(c('HEART', 5), hand, following)).toBe(true);
    });
  });

  describe('first-trick discard when void', () => {
    const trick = [{ playerIdx: 0, card: c('CLOVER', 2) }];

    it('forbids discarding point cards on the first trick while a safe card remains', () => {
      const hand = [c('HEART', 5), QUEEN_SPADES, c('SPADE', 9)];
      const following = ctx({ currentTrick: trick, trickNumber: 1 });
      expect(isHeartsLegalPlay(c('HEART', 5), hand, following)).toBe(false);
      expect(isHeartsLegalPlay(QUEEN_SPADES, hand, following)).toBe(false);
      expect(isHeartsLegalPlay(c('SPADE', 9), hand, following)).toBe(true);
    });

    it('allows point cards on the first trick when the hand is all point cards', () => {
      const hand = [c('HEART', 5), QUEEN_SPADES];
      const following = ctx({ currentTrick: trick, trickNumber: 1 });
      expect(isHeartsLegalPlay(c('HEART', 5), hand, following)).toBe(true);
      expect(isHeartsLegalPlay(QUEEN_SPADES, hand, following)).toBe(true);
    });

    it('treats J♦ as a point card only under the Omnibus option', () => {
      const hand = [JACK_DIAMONDS, c('SPADE', 9)];
      const base = { currentTrick: trick, trickNumber: 1 };
      expect(isHeartsLegalPlay(JACK_DIAMONDS, hand, ctx({ ...base, omnibusJD: true }))).toBe(false);
      expect(isHeartsLegalPlay(JACK_DIAMONDS, hand, ctx({ ...base, omnibusJD: false }))).toBe(true);
    });
  });
});

describe('heartsLegalPlayIndices', () => {
  it('returns only the 2♣ index for the first-trick lead', () => {
    const hand = [c('SPADE', 1), TWO_CLUBS, c('HEART', 5)];
    expect(heartsLegalPlayIndices(hand, ctx({ trickNumber: 1, currentTrick: [] }))).toEqual([1]);
  });

  it('returns all non-heart indices when leading before hearts break', () => {
    const hand = [c('HEART', 5), c('SPADE', 9), c('CLOVER', 7)];
    expect(heartsLegalPlayIndices(hand, ctx({ currentTrick: [], heartsBroken: false }))).toEqual([1, 2]);
  });

  it('returns every index when unrestricted (void, later trick)', () => {
    const hand = [c('SPADE', 9), c('HEART', 5)];
    const trick = [{ playerIdx: 0, card: c('DIAMOND', 3) }];
    expect(heartsLegalPlayIndices(hand, ctx({ currentTrick: trick, trickNumber: 3 }))).toEqual([0, 1]);
  });
});

describe('heartsIllegalReasonKey', () => {
  it('returns the 2♣ reason on the first-trick lead', () => {
    const hand = [TWO_CLUBS, c('SPADE', 1)];
    expect(heartsIllegalReasonKey(hand, ctx({ trickNumber: 1, currentTrick: [] }))).toBe(
      'illegalReason.mustLeadTwoClubs',
    );
  });

  it('returns the hearts-not-broken reason when leading a heart is blocked', () => {
    const hand = [c('HEART', 5), c('SPADE', 9)];
    expect(heartsIllegalReasonKey(hand, ctx({ currentTrick: [], heartsBroken: false }))).toBe(
      'illegalReason.heartsNotBroken',
    );
  });

  it('returns null when leading freely (hearts broken)', () => {
    const hand = [c('HEART', 5), c('SPADE', 9)];
    expect(heartsIllegalReasonKey(hand, ctx({ currentTrick: [], heartsBroken: true }))).toBeNull();
  });

  it('returns null when leading and the hand is all hearts', () => {
    const hand = [c('HEART', 5), c('HEART', 9)];
    expect(heartsIllegalReasonKey(hand, ctx({ currentTrick: [], heartsBroken: false }))).toBeNull();
  });

  it('returns the follow-suit reason when the lead suit is held', () => {
    const hand = [c('DIAMOND', 8), c('SPADE', 9)];
    const trick = [{ playerIdx: 0, card: c('DIAMOND', 3) }];
    expect(heartsIllegalReasonKey(hand, ctx({ currentTrick: trick }))).toBe('illegalReason.mustFollowSuit');
  });

  it('returns the first-trick point-card reason when void with a safe card', () => {
    const hand = [c('HEART', 5), c('SPADE', 9)];
    const trick = [{ playerIdx: 0, card: c('CLOVER', 2) }];
    expect(heartsIllegalReasonKey(hand, ctx({ currentTrick: trick, trickNumber: 1 }))).toBe(
      'illegalReason.noPointsFirstTrick',
    );
  });

  it('returns null when void with no restriction (later trick)', () => {
    const hand = [c('HEART', 5), c('SPADE', 9)];
    const trick = [{ playerIdx: 0, card: c('CLOVER', 2) }];
    expect(heartsIllegalReasonKey(hand, ctx({ currentTrick: trick, trickNumber: 3 }))).toBeNull();
  });

  it('returns null on the first trick when void but holding only point cards', () => {
    const hand = [c('HEART', 5), QUEEN_SPADES];
    const trick = [{ playerIdx: 0, card: c('CLOVER', 2) }];
    expect(heartsIllegalReasonKey(hand, ctx({ currentTrick: trick, trickNumber: 1 }))).toBeNull();
  });
});
