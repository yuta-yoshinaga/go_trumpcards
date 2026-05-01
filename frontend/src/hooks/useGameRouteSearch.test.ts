import { renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { useGameRouteSearch } from './useGameRouteSearch';

describe('useGameRouteSearch', () => {
  it('returns null arrays when search term is empty', () => {
    const { result } = renderHook(() => useGameRouteSearch(''));
    expect(result.current.filteredRoutes).toBeNull();
    expect(result.current.filteredPaths).toBeNull();
  });

  it('matches against the Japanese label by default (ja test setup)', () => {
    // The test bootstrap initializes ja translations; "ブラックジャック" is the
    // Japanese label for blackjack. We use a substring so the test does not
    // hard-code the exact phrase and stays robust to label tweaks.
    const { result } = renderHook(() => useGameRouteSearch('ブラック'));
    const routes = result.current.filteredRoutes ?? [];
    expect(routes.length).toBeGreaterThan(0);
    expect(routes.some((r) => r.labelKey === 'nav.blackjack')).toBe(true);
  });

  it('matches against the English label even when the active language is ja', () => {
    // i18n test setup loads both ja and en bundles. Searching English text
    // while ja is the active language must still match — that is the whole
    // point of pulling both lng overrides into the searchable index.
    const { result } = renderHook(() => useGameRouteSearch('blackjack'));
    const routes = result.current.filteredRoutes ?? [];
    expect(routes.some((r) => r.labelKey === 'nav.blackjack')).toBe(true);
  });

  it('returns an empty result set for unmatched input', () => {
    const { result } = renderHook(() => useGameRouteSearch('zzz-no-such-game'));
    expect(result.current.filteredRoutes).toEqual([]);
    expect(result.current.filteredPaths?.size).toBe(0);
  });

  it('is case-insensitive', () => {
    const a = renderHook(() => useGameRouteSearch('BLACKJACK')).result.current.filteredRoutes;
    const b = renderHook(() => useGameRouteSearch('blackjack')).result.current.filteredRoutes;
    expect(a?.length).toBe(b?.length);
    expect(a?.length).toBeGreaterThan(0);
  });
});
