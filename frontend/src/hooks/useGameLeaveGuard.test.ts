import { renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { useGameLeaveGuard } from './useGameLeaveGuard';

describe('useGameLeaveGuard', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('registers a beforeunload listener when active', () => {
    const addSpy = vi.spyOn(window, 'addEventListener');
    renderHook(() => useGameLeaveGuard(true, 'Round in progress'));
    expect(addSpy).toHaveBeenCalledWith('beforeunload', expect.any(Function));
  });

  it('does not register listener when inactive', () => {
    const addSpy = vi.spyOn(window, 'addEventListener');
    renderHook(() => useGameLeaveGuard(false, 'msg'));
    const beforeUnloadCalls = addSpy.mock.calls.filter(([ev]) => ev === 'beforeunload');
    expect(beforeUnloadCalls).toHaveLength(0);
  });

  it('removes listener on unmount', () => {
    const removeSpy = vi.spyOn(window, 'removeEventListener');
    const { unmount } = renderHook(() => useGameLeaveGuard(true, 'msg'));
    unmount();
    expect(removeSpy).toHaveBeenCalledWith('beforeunload', expect.any(Function));
  });

  it('handler sets returnValue on the event', () => {
    let registered: ((e: BeforeUnloadEvent) => unknown) | null = null;
    vi.spyOn(window, 'addEventListener').mockImplementation((ev, handler) => {
      if (ev === 'beforeunload') registered = handler as (e: BeforeUnloadEvent) => unknown;
    });
    renderHook(() => useGameLeaveGuard(true, 'Round in progress'));
    expect(registered).not.toBeNull();
    const event = { preventDefault: vi.fn(), returnValue: '' } as unknown as BeforeUnloadEvent;
    // biome-ignore lint/style/noNonNullAssertion: null check above confirms registered is set
    registered!(event);
    expect(event.preventDefault).toHaveBeenCalled();
    expect(event.returnValue).toBe('Round in progress');
  });
});
