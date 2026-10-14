import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { useElementWidth } from './useElementWidth';

class MockResizeObserver {
  static instances: MockResizeObserver[] = [];
  callback: ResizeObserverCallback;
  observe = vi.fn();
  disconnect = vi.fn();

  constructor(callback: ResizeObserverCallback) {
    this.callback = callback;
    MockResizeObserver.instances.push(this);
  }

  notify(width: number) {
    this.callback([{ contentRect: { width } } as ResizeObserverEntry], this as unknown as ResizeObserver);
  }
}

describe('useElementWidth', () => {
  const originalResizeObserver = window.ResizeObserver;

  afterEach(() => {
    Object.defineProperty(window, 'ResizeObserver', { configurable: true, value: originalResizeObserver });
    MockResizeObserver.instances = [];
  });

  it('returns the floored observed width', () => {
    Object.defineProperty(window, 'ResizeObserver', { configurable: true, value: MockResizeObserver });
    const ref = { current: document.createElement('div') };
    const { result } = renderHook(() => useElementWidth(ref, true));

    act(() => MockResizeObserver.instances[0]?.notify(1008.9));

    expect(result.current).toBe(1008);
  });

  it('does not observe when disabled', () => {
    Object.defineProperty(window, 'ResizeObserver', { configurable: true, value: MockResizeObserver });
    const ref = { current: document.createElement('div') };
    renderHook(() => useElementWidth(ref, false));

    expect(MockResizeObserver.instances).toHaveLength(0);
  });

  it('disconnects on unmount', () => {
    Object.defineProperty(window, 'ResizeObserver', { configurable: true, value: MockResizeObserver });
    const ref = { current: document.createElement('div') };
    const { unmount } = renderHook(() => useElementWidth(ref, true));

    unmount();

    expect(MockResizeObserver.instances[0]?.disconnect).toHaveBeenCalledOnce();
  });
});
