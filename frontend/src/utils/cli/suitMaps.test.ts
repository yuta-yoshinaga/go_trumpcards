import { describe, expect, it } from 'vitest';
import { BRIDGE_SUIT_MAP, STANDARD_SUIT_MAP } from './suitMaps';

describe('STANDARD_SUIT_MAP', () => {
  it('maps every alias of each suit to the same number', () => {
    expect(STANDARD_SUIT_MAP.spade).toBe(1);
    expect(STANDARD_SUIT_MAP.spades).toBe(1);
    expect(STANDARD_SUIT_MAP.s).toBe(1);
    expect(STANDARD_SUIT_MAP.heart).toBe(3);
    expect(STANDARD_SUIT_MAP.hearts).toBe(3);
    expect(STANDARD_SUIT_MAP.h).toBe(3);
    expect(STANDARD_SUIT_MAP.diamond).toBe(4);
    expect(STANDARD_SUIT_MAP.diamonds).toBe(4);
    expect(STANDARD_SUIT_MAP.d).toBe(4);
  });

  // Regression for #2152: `club` (singular) used to be accepted only by Mighty
  // and rejected by the other standard-suit games, an inconsistent UX bug. The
  // shared map now accepts club / clubs / clover uniformly.
  it('accepts club, clubs, and clover as the same suit', () => {
    expect(STANDARD_SUIT_MAP.club).toBe(2);
    expect(STANDARD_SUIT_MAP.clubs).toBe(2);
    expect(STANDARD_SUIT_MAP.clover).toBe(2);
    expect(STANDARD_SUIT_MAP.c).toBe(2);
  });

  it('returns undefined for unknown tokens', () => {
    expect(STANDARD_SUIT_MAP.bogus).toBeUndefined();
  });
});

describe('BRIDGE_SUIT_MAP', () => {
  it('uses bridge-specific ranking with no-trump highest', () => {
    expect(BRIDGE_SUIT_MAP.club).toBe(1);
    expect(BRIDGE_SUIT_MAP.diamond).toBe(2);
    expect(BRIDGE_SUIT_MAP.heart).toBe(3);
    expect(BRIDGE_SUIT_MAP.spade).toBe(4);
    expect(BRIDGE_SUIT_MAP.notrump).toBe(5);
    expect(BRIDGE_SUIT_MAP.nt).toBe(5);
  });

  // #2152 review: every suit must accept its plural form and single-letter
  // abbreviation, so the alias set matches STANDARD_SUIT_MAP everywhere.
  it('accepts plural and abbreviated aliases for each suit', () => {
    expect(BRIDGE_SUIT_MAP.clubs).toBe(1);
    expect(BRIDGE_SUIT_MAP.c).toBe(1);
    expect(BRIDGE_SUIT_MAP.diamonds).toBe(2);
    expect(BRIDGE_SUIT_MAP.d).toBe(2);
    expect(BRIDGE_SUIT_MAP.hearts).toBe(3);
    expect(BRIDGE_SUIT_MAP.h).toBe(3);
    expect(BRIDGE_SUIT_MAP.spades).toBe(4);
    expect(BRIDGE_SUIT_MAP.s).toBe(4);
  });

  // #2152 review: `clover` (the JP-localized club alias) must also resolve in
  // Bridge, matching STANDARD_SUIT_MAP — otherwise the PR would re-introduce
  // the cross-game inconsistency it set out to remove.
  it('accepts clover as a club alias', () => {
    expect(BRIDGE_SUIT_MAP.clover).toBe(1);
  });
});
