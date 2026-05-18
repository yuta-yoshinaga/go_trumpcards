/**
 * @vitest-environment jsdom
 */

import { renderHook } from '@testing-library/react';
import { HashRouter, useSearchParams } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { encodeMood, parseMood, parseSearchParams, type UserMoodInput } from './urlMoodCodec';

describe('encodeMood / parseMood', () => {
  it('round-trips a fully answered mood', () => {
    const input: UserMoodInput = {
      mood: [2, 3],
      skill: [0, 1],
      social: [1, 1],
      theme: [0, 0],
    };
    const encoded = encodeMood(input);
    expect(encoded).toBe('m=2,3&s=0,1&so=1,1&t=0,0');
    expect(parseMood(encoded)).toEqual(input);
  });

  it('represents skipped answers with `-`', () => {
    const input: UserMoodInput = {
      mood: [null, 3],
      skill: [null, null],
      social: [1, null],
      theme: [0, null],
    };
    const encoded = encodeMood(input);
    expect(encoded).toBe('m=-,3&s=-,-&so=1,-&t=0,-');
    expect(parseMood(encoded)).toEqual(input);
  });

  it('returns null for missing axis keys', () => {
    expect(parseMood('m=1,2&s=0,0&t=0,0')).toBeNull();
  });

  it('returns null for non-integer values', () => {
    expect(parseMood('m=1,abc&s=0,0&so=0,0&t=0,0')).toBeNull();
  });

  it('returns null when an axis has the wrong arity', () => {
    expect(parseMood('m=1&s=0,0&so=0,0&t=0,0')).toBeNull();
    expect(parseMood('m=1,2,3&s=0,0&so=0,0&t=0,0')).toBeNull();
  });

  it('rejects negative answer indices', () => {
    expect(parseMood('m=-1,2&s=0,0&so=0,0&t=0,0')).toBeNull();
  });

  it('rejects answer indices beyond the axis option count', () => {
    expect(parseMood('m=99,2&s=0,0&so=0,0&t=0,0')).toBeNull();
  });
});

describe('parseSearchParams (HashRouter integration)', () => {
  // Wrap a hook in HashRouter so useSearchParams reads the location.hash.
  function setupHash(hash: string) {
    window.location.hash = hash;
    return renderHook(() => useSearchParams()[0], {
      wrapper: ({ children }) => <HashRouter>{children}</HashRouter>,
    });
  }

  it('parses a HashRouter URL into UserMoodInput', () => {
    const { result } = setupHash('#/discover/result?m=2,3&s=0,1&so=1,1&t=0,0');
    const parsed = parseSearchParams(result.current);
    expect(parsed).toEqual({
      mood: [2, 3],
      skill: [0, 1],
      social: [1, 1],
      theme: [0, 0],
    });
  });

  it('returns null when HashRouter URL is missing required axes', () => {
    const { result } = setupHash('#/discover/result?m=2,3');
    expect(parseSearchParams(result.current)).toBeNull();
  });

  it('returns null when HashRouter URL has malformed values', () => {
    const { result } = setupHash('#/discover/result?m=99,abc&s=0,0&so=0,0&t=0,0');
    expect(parseSearchParams(result.current)).toBeNull();
  });
});
