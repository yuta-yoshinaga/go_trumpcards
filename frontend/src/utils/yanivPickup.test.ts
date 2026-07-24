import { describe, expect, it } from 'vitest';
import { isPickupable, pickupableIndices } from './yanivPickup';

describe('pickupableIndices', () => {
  it('returns nothing for an empty discard', () => {
    expect(pickupableIndices(0)).toEqual([]);
    expect(pickupableIndices(-1)).toEqual([]);
  });

  it('marks the single card of a single/one-card discard', () => {
    expect(pickupableIndices(1)).toEqual([0]);
  });

  it('marks both ends of a two-card discard', () => {
    expect(pickupableIndices(2)).toEqual([0, 1]);
  });

  it('marks only the two ends of a discarded run/set (middle excluded)', () => {
    expect(pickupableIndices(3)).toEqual([0, 2]);
    expect(pickupableIndices(5)).toEqual([0, 4]);
  });
});

describe('isPickupable', () => {
  it('treats the two ends of a run as pickup-able and the middle as not', () => {
    expect(isPickupable(0, 4)).toBe(true);
    expect(isPickupable(3, 4)).toBe(true);
    expect(isPickupable(1, 4)).toBe(false);
    expect(isPickupable(2, 4)).toBe(false);
  });

  it('treats the lone card of a single discard as pickup-able', () => {
    expect(isPickupable(0, 1)).toBe(true);
  });
});
