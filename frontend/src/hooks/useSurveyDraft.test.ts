/**
 * @vitest-environment jsdom
 */
import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { DISCOVER_DRAFT_KEY, DRAFT_SCHEMA_VERSION, readDraft, useSurveyDraft } from './useSurveyDraft';

describe('useSurveyDraft', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('starts empty when localStorage is empty', () => {
    const { result } = renderHook(() => useSurveyDraft());
    expect(result.current.axes.mood).toEqual([null, null]);
    expect(result.current.axes.skill).toEqual([null, null]);
    expect(result.current.axes.social).toEqual([null, null]);
    expect(result.current.axes.theme).toEqual([null, null]);
  });

  it('restores a stored draft on mount', () => {
    // Each value is now a per-sub-question option index, not a master-list index:
    // mood Q1 ∈ [0,1], Q2 ∈ [0,1] ; skill Q1 ∈ [0..2], Q2 ∈ [0,1] ; etc.
    localStorage.setItem(
      DISCOVER_DRAFT_KEY,
      JSON.stringify({
        v: DRAFT_SCHEMA_VERSION,
        axes: { mood: [1, null], skill: [null, null], social: [1, 0], theme: [0, 0] },
      }),
    );
    const { result } = renderHook(() => useSurveyDraft());
    expect(result.current.axes.mood).toEqual([1, null]);
    expect(result.current.axes.social).toEqual([1, 0]);
  });

  it('drops a draft with a mismatched schema version', () => {
    localStorage.setItem(
      DISCOVER_DRAFT_KEY,
      JSON.stringify({ v: 99, axes: { mood: [2, 3], skill: [0, 0], social: [1, 1], theme: [0, 0] } }),
    );
    expect(readDraft()).toBeNull();
    // And the corrupt entry is wiped so the next read also returns null.
    expect(localStorage.getItem(DISCOVER_DRAFT_KEY)).toBeNull();
  });

  it('drops a draft when JSON is malformed', () => {
    localStorage.setItem(DISCOVER_DRAFT_KEY, 'not-json');
    expect(readDraft()).toBeNull();
  });

  it('clamps out-of-range answer indices to null on read', () => {
    // mood Q1 only has 2 options now, so 99 is dropped; Q2 also has 2 options, so 1 is valid.
    localStorage.setItem(
      DISCOVER_DRAFT_KEY,
      JSON.stringify({
        v: DRAFT_SCHEMA_VERSION,
        axes: { mood: [99, 1], skill: [-1, 0], social: [0, 0], theme: [0, 0] },
      }),
    );
    const restored = readDraft();
    expect(restored?.axes.mood).toEqual([null, 1]);
    expect(restored?.axes.skill).toEqual([null, 0]);
  });

  it('setAnswer persists the change to localStorage', () => {
    const { result } = renderHook(() => useSurveyDraft());
    act(() => result.current.setAnswer('mood', 0, 1));
    expect(result.current.axes.mood).toEqual([1, null]);
    const stored = JSON.parse(localStorage.getItem(DISCOVER_DRAFT_KEY) ?? '{}');
    expect(stored.axes.mood).toEqual([1, null]);
    expect(stored.v).toBe(DRAFT_SCHEMA_VERSION);
  });

  it('reset wipes both in-memory state and localStorage', () => {
    const { result } = renderHook(() => useSurveyDraft());
    act(() => {
      result.current.setAnswer('mood', 0, 1);
      result.current.setAnswer('skill', 1, 1);
    });
    act(() => result.current.reset());
    expect(result.current.axes.mood).toEqual([null, null]);
    expect(result.current.axes.skill).toEqual([null, null]);
    expect(localStorage.getItem(DISCOVER_DRAFT_KEY)).toBeNull();
  });

  it('survives a localStorage.setItem that throws', () => {
    const setItem = vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('QuotaExceededError');
    });
    const { result } = renderHook(() => useSurveyDraft());
    // Should not throw even though setItem fails on every update.
    expect(() => act(() => result.current.setAnswer('mood', 0, 1))).not.toThrow();
    expect(result.current.axes.mood).toEqual([1, null]);
    setItem.mockRestore();
  });
});
