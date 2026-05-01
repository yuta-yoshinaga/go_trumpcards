import { renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { useGameRoundGuard } from './useGameRoundGuard';

describe('useGameRoundGuard', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('registers a beforeunload listener with the i18n confirm message when active', () => {
    let registered: ((e: BeforeUnloadEvent) => unknown) | null = null;
    vi.spyOn(window, 'addEventListener').mockImplementation((ev, handler) => {
      if (ev === 'beforeunload') registered = handler as (e: BeforeUnloadEvent) => unknown;
    });
    renderHook(() => useGameRoundGuard(true));
    expect(registered).not.toBeNull();
    const event = { preventDefault: vi.fn(), returnValue: '' } as unknown as BeforeUnloadEvent;
    // biome-ignore lint/style/noNonNullAssertion: null check above confirms registered is set
    registered!(event);
    expect(event.preventDefault).toHaveBeenCalled();
    // Tests run with ja locale loaded.
    expect(event.returnValue).toBe('ページを離れると現在のラウンドが破棄されます。続けますか？');
  });

  it('does not register listener when inactive', () => {
    const addSpy = vi.spyOn(window, 'addEventListener');
    renderHook(() => useGameRoundGuard(false));
    const beforeUnloadCalls = addSpy.mock.calls.filter(([ev]) => ev === 'beforeunload');
    expect(beforeUnloadCalls).toHaveLength(0);
  });
});
