import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { useBodyScrollLock } from './useBodyScrollLock';

describe('useBodyScrollLock', () => {
  afterEach(() => {
    document.body.style.overflow = '';
  });

  it('does nothing when active is false', () => {
    document.body.style.overflow = 'auto';
    renderHook(() => useBodyScrollLock(false));
    expect(document.body.style.overflow).toBe('auto');
  });

  it('sets body overflow to hidden when active is true', () => {
    renderHook(() => useBodyScrollLock(true));
    expect(document.body.style.overflow).toBe('hidden');
  });

  it('restores the previous overflow value on unmount', () => {
    document.body.style.overflow = 'scroll';
    const { unmount } = renderHook(() => useBodyScrollLock(true));
    expect(document.body.style.overflow).toBe('hidden');
    unmount();
    expect(document.body.style.overflow).toBe('scroll');
  });

  it('restores when active flips from true to false', () => {
    document.body.style.overflow = 'auto';
    const { rerender } = renderHook(({ active }) => useBodyScrollLock(active), {
      initialProps: { active: true },
    });
    expect(document.body.style.overflow).toBe('hidden');
    rerender({ active: false });
    expect(document.body.style.overflow).toBe('auto');
  });

  it('locks when active flips from false to true', () => {
    document.body.style.overflow = 'auto';
    const { rerender } = renderHook(({ active }) => useBodyScrollLock(active), {
      initialProps: { active: false },
    });
    expect(document.body.style.overflow).toBe('auto');
    act(() => rerender({ active: true }));
    expect(document.body.style.overflow).toBe('hidden');
  });

  it('preserves empty string overflow on restore', () => {
    document.body.style.overflow = '';
    const { unmount } = renderHook(() => useBodyScrollLock(true));
    expect(document.body.style.overflow).toBe('hidden');
    unmount();
    expect(document.body.style.overflow).toBe('');
  });
});
