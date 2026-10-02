import { afterEach, describe, expect, it } from 'vitest';
import { BURRACO_SORT_STORAGE_KEY, loadBurracoSortMode, saveBurracoSortMode } from './burracoSort';

describe('burraco sort mode persistence', () => {
  afterEach(() => localStorage.clear());

  it('defaults to original when nothing is stored', () => {
    expect(loadBurracoSortMode()).toBe('original');
  });

  it('round-trips a saved mode through localStorage', () => {
    saveBurracoSortMode('suit');
    expect(localStorage.getItem(BURRACO_SORT_STORAGE_KEY)).toBe('suit');
    expect(loadBurracoSortMode()).toBe('suit');
  });

  it('ignores an invalid stored value', () => {
    localStorage.setItem(BURRACO_SORT_STORAGE_KEY, 'bogus');
    expect(loadBurracoSortMode()).toBe('original');
  });
});
