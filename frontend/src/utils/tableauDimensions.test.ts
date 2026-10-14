import { describe, expect, it } from 'vitest';
import { tableauDimensions } from './tableauDimensions';

describe('tableauDimensions', () => {
  it('sizes cards and stack spacing from the full rendered column width', () => {
    expect(tableauDimensions(1008, 8, 8, 0.34)).toEqual({ cardWidth: 119, cardHeight: 179, cardOverlap: 40 });
  });
});
