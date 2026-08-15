import { renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { SITE_NAME } from '../constants/site';
import { useDocumentTitle } from './useDocumentTitle';

afterEach(() => {
  document.title = '';
});

describe('useDocumentTitle', () => {
  it('sets the title while mounted', () => {
    renderHook(() => useDocumentTitle('おすすめを探す'));
    expect(document.title).toBe(`おすすめを探す - ${SITE_NAME}`);
  });

  // The whole point of the cleanup: leaving a page must not leave its name in
  // the tab. Before this hook existed, only game pages managed the title, so
  // navigating to /discover kept whatever the previous page had set (or fell
  // back to the bare site name with no page identity at all). See issue #5360.
  it('restores the bare site name on unmount', () => {
    const { unmount } = renderHook(() => useDocumentTitle('商標とクレジット'));
    expect(document.title).toBe(`商標とクレジット - ${SITE_NAME}`);
    unmount();
    expect(document.title).toBe(SITE_NAME);
  });

  it('follows a changing title without needing a remount', () => {
    const { rerender } = renderHook(({ title }) => useDocumentTitle(title), {
      initialProps: { title: 'おすすめを探す' },
    });
    rerender({ title: '結果' });
    expect(document.title).toBe(`結果 - ${SITE_NAME}`);
  });

  // A page whose translation has not loaded yet would otherwise render
  // " - Trump Cards" with a dangling separator.
  it('falls back to the bare site name when the title is empty', () => {
    renderHook(() => useDocumentTitle(''));
    expect(document.title).toBe(SITE_NAME);
  });
});
