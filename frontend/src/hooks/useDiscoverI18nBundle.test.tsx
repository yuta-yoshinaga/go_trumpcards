/**
 * @vitest-environment jsdom
 */
import { act, renderHook, waitFor } from '@testing-library/react';
import i18n from 'i18next';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { __resetDiscoverI18nBundleCacheForTests, useDiscoverI18nBundle } from './useDiscoverI18nBundle';

const NS = 'discover';

describe('useDiscoverI18nBundle', () => {
  // The test setup eagerly loads discover into both ja and en (skipLazy: false).
  // We snapshot and restore those bundles per test so we can drive the hook
  // through the "not loaded yet" path explicitly.
  let snapshotJa: ReturnType<typeof i18n.getResourceBundle> | undefined;
  let snapshotEn: ReturnType<typeof i18n.getResourceBundle> | undefined;

  beforeEach(() => {
    snapshotJa = i18n.getResourceBundle('ja', NS);
    snapshotEn = i18n.getResourceBundle('en', NS);
    __resetDiscoverI18nBundleCacheForTests();
  });

  afterEach(() => {
    if (snapshotJa) i18n.addResourceBundle('ja', NS, snapshotJa, true, true);
    if (snapshotEn) i18n.addResourceBundle('en', NS, snapshotEn, true, true);
    __resetDiscoverI18nBundleCacheForTests();
    i18n.changeLanguage('ja');
  });

  it('returns true immediately when the bundle is already loaded', () => {
    const { result } = renderHook(() => useDiscoverI18nBundle());
    expect(result.current).toBe(true);
  });

  it('loads the bundle via dynamic import and flips ready to true', async () => {
    i18n.removeResourceBundle('ja', NS);
    i18n.removeResourceBundle('en', NS);
    const { result } = renderHook(() => useDiscoverI18nBundle());
    // First synchronous read is false (resource missing).
    expect(result.current).toBe(false);
    // After dynamic import resolves, hook re-renders with ready=true.
    await waitFor(() => expect(result.current).toBe(true));
    expect(i18n.hasResourceBundle('ja', NS)).toBe(true);
  });

  it('normalizes BCP-47 region tags (en-US) down to the base lang (en)', async () => {
    i18n.removeResourceBundle('ja', NS);
    i18n.removeResourceBundle('en', NS);
    await act(async () => {
      await i18n.changeLanguage('en-US');
    });
    const { result } = renderHook(() => useDiscoverI18nBundle());
    await waitFor(() => expect(result.current).toBe(true));
    // Must register against the base lang `en`, not `en-US`.
    expect(i18n.hasResourceBundle('en', NS)).toBe(true);
    expect(i18n.hasResourceBundle('en-US', NS)).toBe(false);
  });

  it('falls back to ja when an unknown locale resolves to a missing bundle', async () => {
    i18n.removeResourceBundle('ja', NS);
    i18n.removeResourceBundle('en', NS);
    await act(async () => {
      await i18n.changeLanguage('zz');
    });
    const { result } = renderHook(() => useDiscoverI18nBundle());
    await waitFor(() => expect(result.current).toBe(true));
    expect(i18n.hasResourceBundle('ja', NS)).toBe(true);
  });

  it('dedupes concurrent loads — only one in-flight import per language', async () => {
    i18n.removeResourceBundle('ja', NS);
    i18n.removeResourceBundle('en', NS);
    const addSpy = vi.spyOn(i18n, 'addResourceBundle');
    const a = renderHook(() => useDiscoverI18nBundle());
    const b = renderHook(() => useDiscoverI18nBundle());
    await waitFor(() => expect(a.result.current).toBe(true));
    await waitFor(() => expect(b.result.current).toBe(true));
    // The bundle should have been registered exactly once despite two
    // simultaneous renders.
    expect(addSpy).toHaveBeenCalledTimes(1);
    addSpy.mockRestore();
  });
});
