import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useTransientToast } from './useTransientToast';

describe('useTransientToast', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it('does not show on the initial render when skipInitial (default)', () => {
    const { result } = renderHook(({ trigger }) => useTransientToast(trigger, 1000), {
      initialProps: { trigger: 'a' },
    });
    expect(result.current.visible).toBe(false);
  });

  it('shows on a subsequent trigger change and auto-dismisses', () => {
    const { result, rerender } = renderHook(({ trigger }) => useTransientToast(trigger, 1000), {
      initialProps: { trigger: 'a' },
    });
    act(() => rerender({ trigger: 'b' }));
    expect(result.current.visible).toBe(true);
    act(() => vi.advanceTimersByTime(1000));
    expect(result.current.visible).toBe(false);
  });

  it('shows on the initial render when skipInitial is false', () => {
    const { result } = renderHook(() => useTransientToast(2, 1000, { skipInitial: false, active: true }));
    expect(result.current.visible).toBe(true);
  });

  it('does not show when active is false, even on a trigger change', () => {
    const { result, rerender } = renderHook(({ trigger, active }) => useTransientToast(trigger, 1000, { active }), {
      initialProps: { trigger: 0, active: false },
    });
    act(() => rerender({ trigger: 2, active: false }));
    expect(result.current.visible).toBe(false);
  });

  it('dismiss() hides an already-visible toast', () => {
    const { result, rerender } = renderHook(({ trigger }) => useTransientToast(trigger, 1000), {
      initialProps: { trigger: 'a' },
    });
    act(() => rerender({ trigger: 'b' }));
    expect(result.current.visible).toBe(true);
    act(() => result.current.dismiss());
    expect(result.current.visible).toBe(false);
  });

  it('closes on Escape unless a modal dialog is open', () => {
    const { result, rerender } = renderHook(({ trigger }) => useTransientToast(trigger, 1000), {
      initialProps: { trigger: 'a' },
    });
    // Modal open: Escape is ignored.
    const dialog = document.createElement('div');
    dialog.setAttribute('role', 'dialog');
    dialog.setAttribute('aria-modal', 'true');
    document.body.appendChild(dialog);
    act(() => rerender({ trigger: 'b' }));
    act(() => window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' })));
    expect(result.current.visible).toBe(true);
    document.body.removeChild(dialog);
    // No modal: Escape closes it.
    act(() => window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' })));
    expect(result.current.visible).toBe(false);
  });
});
