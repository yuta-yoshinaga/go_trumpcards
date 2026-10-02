import { afterEach, describe, expect, it } from 'vitest';
import { BIRIBA_SORT_STORAGE_KEY, loadBiribaSortMode, saveBiribaSortMode } from './biribaSort';

describe('biriba sort mode persistence', () => {
  afterEach(() => localStorage.clear());

  it('defaults to original when nothing is stored', () => {
    expect(loadBiribaSortMode()).toBe('original');
  });

  it('round-trips a saved mode through localStorage', () => {
    saveBiribaSortMode('suit');
    expect(localStorage.getItem(BIRIBA_SORT_STORAGE_KEY)).toBe('suit');
    expect(loadBiribaSortMode()).toBe('suit');
  });

  it('ignores an invalid stored value', () => {
    localStorage.setItem(BIRIBA_SORT_STORAGE_KEY, 'bogus');
    expect(loadBiribaSortMode()).toBe('original');
  });
});
