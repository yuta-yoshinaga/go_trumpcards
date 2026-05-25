/**
 * @vitest-environment jsdom
 */

import { renderHook } from '@testing-library/react';
import { HashRouter, useSearchParams } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { encodeMood, parseMood, parseSearchParams, type UserMoodInput } from './urlMoodCodec';

describe('encodeMood / parseMood', () => {
  it('round-trips a fully answered mood (per-sub-question option indices)', () => {
    // mood Q1 has 2 options (0..1), Q2 has 2 options (0..1).
    // skill Q1 has 3 options (0..2), Q2 has 2 options (0..1).
    // social Q1 has 3 options (0..2), Q2 has 2 options (0..1).
    // theme Q1 has 4 options (0..3), Q2 has 2 options (0..1).
    const input: UserMoodInput = {
      mood: [1, 1],
      skill: [2, 1],
      social: [1, 0],
      theme: [3, 0],
    };
    const encoded = encodeMood(input);
    expect(encoded).toBe('m=1,1&s=2,1&so=1,0&t=3,0');
    expect(parseMood(encoded)).toEqual(input);
  });

  it('represents skipped answers with `-`', () => {
    const input: UserMoodInput = {
      mood: [null, 1],
      skill: [null, null],
      social: [1, null],
      theme: [0, null],
    };
    const encoded = encodeMood(input);
    expect(encoded).toBe('m=-,1&s=-,-&so=1,-&t=0,-');
    expect(parseMood(encoded)).toEqual(input);
  });

  it('returns null for missing axis keys', () => {
    expect(parseMood('m=1,1&s=0,0&t=0,0')).toBeNull();
  });

  it('returns null for non-integer values', () => {
    expect(parseMood('m=1,abc&s=0,0&so=0,0&t=0,0')).toBeNull();
  });

  it('returns null when an axis has the wrong arity', () => {
    expect(parseMood('m=1&s=0,0&so=0,0&t=0,0')).toBeNull();
    expect(parseMood('m=1,0,1&s=0,0&so=0,0&t=0,0')).toBeNull();
  });

  it('rejects negative answer indices', () => {
    expect(parseMood('m=-1,1&s=0,0&so=0,0&t=0,0')).toBeNull();
  });

  it('rejects answer indices beyond the per-question option count', () => {
    // mood Q1 only has 2 options, so 2 is out of range.
    expect(parseMood('m=2,0&s=0,0&so=0,0&t=0,0')).toBeNull();
    // skill Q2 only has 2 options, so 2 is out of range for Q2 even though Q1 accepts 2.
    expect(parseMood('m=0,0&s=2,2&so=0,0&t=0,0')).toBeNull();
  });
});

describe('parseSearchParams (HashRouter integration)', () => {
  function setupHash(hash: string) {
    window.location.hash = hash;
    return renderHook(() => useSearchParams()[0], {
      wrapper: ({ children }) => <HashRouter>{children}</HashRouter>,
    });
  }

  it('parses a HashRouter URL into UserMoodInput', () => {
    const { result } = setupHash('#/discover/result?m=1,1&s=2,1&so=1,0&t=3,0');
    const parsed = parseSearchParams(result.current);
    expect(parsed).toEqual({
      mood: [1, 1],
      skill: [2, 1],
      social: [1, 0],
      theme: [3, 0],
    });
  });

  it('returns null when HashRouter URL is missing required axes', () => {
    const { result } = setupHash('#/discover/result?m=1,1');
    expect(parseSearchParams(result.current)).toBeNull();
  });

  it('returns null when HashRouter URL has malformed values', () => {
    const { result } = setupHash('#/discover/result?m=99,abc&s=0,0&so=0,0&t=0,0');
    expect(parseSearchParams(result.current)).toBeNull();
  });
});
