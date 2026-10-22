import { describe, expect, it } from 'vitest';
import { duchessSourceKey } from './duchessSourceKey';

describe('duchessSourceKey', () => {
  it('builds the key for each supported source zone', () => {
    expect(duchessSourceKey({ zone: 'tableau', col: 2, cardIndex: 4 })).toBe('tableau-2-4');
    expect(duchessSourceKey({ zone: 'reserve', col: 3 })).toBe('reserve-3');
    expect(duchessSourceKey({ zone: 'waste' })).toBe('waste');
  });

  it('does not build keys for missing sources or required indices', () => {
    expect(duchessSourceKey(null)).toBeUndefined();
    expect(duchessSourceKey({ zone: 'tableau', col: 2 })).toBeUndefined();
    expect(duchessSourceKey({ zone: 'tableau', cardIndex: 4 })).toBeUndefined();
    expect(duchessSourceKey({ zone: 'reserve' })).toBeUndefined();
    expect(duchessSourceKey({ zone: 'foundation', col: 0 })).toBeUndefined();
  });
});
