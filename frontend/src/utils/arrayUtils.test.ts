import { describe, expect, it } from 'bun:test';
import { toggleArrayItem } from './arrayUtils';

describe('toggleArrayItem', () => {
  it('adds item when not present', () => {
    expect(toggleArrayItem([1, 2], 3)).toEqual([1, 2, 3]);
  });

  it('removes item when already present', () => {
    expect(toggleArrayItem([1, 2, 3], 2)).toEqual([1, 3]);
  });

  it('works with empty array', () => {
    expect(toggleArrayItem([], 5)).toEqual([5]);
  });

  it('works with string arrays', () => {
    expect(toggleArrayItem(['a', 'b'], 'b')).toEqual(['a']);
    expect(toggleArrayItem(['a', 'b'], 'c')).toEqual(['a', 'b', 'c']);
  });
});
