import { describe, expect, it } from 'vitest';
import { rummy500PickupCount } from './rummy500PickupCount';

describe('rummy500PickupCount', () => {
  it('takes only the top card when drawing from the last index', () => {
    // A 4-card pile, drawing from the top (idx 3) takes just that one card.
    expect(rummy500PickupCount(4, 3)).toBe(1);
  });

  it('takes the chosen card plus every card above it for a mid-pile index', () => {
    // Drawing from idx 1 of a 4-card pile takes idx 1, 2 and 3 → 3 cards.
    expect(rummy500PickupCount(4, 1)).toBe(3);
  });

  it('takes the whole pile when drawing from the bottom (idx 0)', () => {
    expect(rummy500PickupCount(4, 0)).toBe(4);
  });

  it('returns 0 for out-of-range indices', () => {
    expect(rummy500PickupCount(4, -1)).toBe(0);
    expect(rummy500PickupCount(4, 4)).toBe(0);
    expect(rummy500PickupCount(0, 0)).toBe(0);
  });
});
