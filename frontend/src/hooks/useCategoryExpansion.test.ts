import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { CATEGORY_EXPANSION_KEY, useCategoryExpansion } from './useCategoryExpansion';

describe('useCategoryExpansion', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  afterEach(() => {
    localStorage.clear();
  });

  it('defaults every category to expanded on first visit (no stored state)', () => {
    const { result } = renderHook(() => useCategoryExpansion());
    expect(result.current.isExpanded('nav.category.table')).toBe(true);
    expect(result.current.isExpanded('nav.category.solitaire')).toBe(true);
    expect(result.current.isExpanded('something.never.seen.before')).toBe(true);
  });

  it('persists a collapse to localStorage', () => {
    const { result } = renderHook(() => useCategoryExpansion());
    act(() => {
      result.current.setExpanded('nav.category.table', false);
    });
    expect(result.current.isExpanded('nav.category.table')).toBe(false);
    const stored = JSON.parse(localStorage.getItem(CATEGORY_EXPANSION_KEY) ?? '{}');
    expect(stored).toEqual({ 'nav.category.table': false });
  });

  it('restores a saved collapse on a fresh hook instance', () => {
    localStorage.setItem(CATEGORY_EXPANSION_KEY, JSON.stringify({ 'nav.category.poker': false }));
    const { result } = renderHook(() => useCategoryExpansion());
    expect(result.current.isExpanded('nav.category.poker')).toBe(false);
    expect(result.current.isExpanded('nav.category.table')).toBe(true); // missing keys still default to open
  });

  it('ignores corrupted localStorage entries', () => {
    localStorage.setItem(CATEGORY_EXPANSION_KEY, 'not json');
    const { result } = renderHook(() => useCategoryExpansion());
    expect(result.current.isExpanded('nav.category.table')).toBe(true);
  });

  it('skips writes when the value did not change (no-op toggle)', () => {
    const { result } = renderHook(() => useCategoryExpansion());
    expect(localStorage.getItem(CATEGORY_EXPANSION_KEY)).toBeNull();
    act(() => {
      // Default is true, setting it to true again should be a no-op.
      result.current.setExpanded('nav.category.table', true);
    });
    expect(localStorage.getItem(CATEGORY_EXPANSION_KEY)).toBeNull();
  });
});
