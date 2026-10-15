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
  const originalResizeObserver = globalThis.ResizeObserver;

  afterEach(() => {
    Object.defineProperty(globalThis, 'ResizeObserver', { configurable: true, value: originalResizeObserver });
    MockResizeObserver.instances = [];
  });

  it('starts observing when an element is attached later and returns its floored width', () => {
    Object.defineProperty(globalThis, 'ResizeObserver', { configurable: true, value: MockResizeObserver });
    const { result } = renderHook(() => useElementWidth<HTMLDivElement>());

    expect(MockResizeObserver.instances).toHaveLength(0);

    act(() => result.current[0](document.createElement('div')));

    expect(MockResizeObserver.instances).toHaveLength(1);
    expect(MockResizeObserver.instances[0]?.observe).toHaveBeenCalledOnce();
    act(() => MockResizeObserver.instances[0]?.notify(1008.9));
    expect(result.current[1]).toBe(1008);
  });

  it('disconnects when the element is detached or replaced', () => {
    Object.defineProperty(globalThis, 'ResizeObserver', { configurable: true, value: MockResizeObserver });
    const { result, unmount } = renderHook(() => useElementWidth<HTMLDivElement>());

    act(() => result.current[0](document.createElement('div')));
    const firstObserver = MockResizeObserver.instances[0];
    act(() => result.current[0](null));

    expect(firstObserver?.disconnect).toHaveBeenCalledOnce();
    act(() => result.current[0](document.createElement('div')));
    const secondObserver = MockResizeObserver.instances[1];
    act(() => result.current[0](null));
    expect(secondObserver?.disconnect).toHaveBeenCalledOnce();
    unmount();
  });

  it('returns zero when ResizeObserver is unavailable', () => {
    Object.defineProperty(globalThis, 'ResizeObserver', { configurable: true, value: undefined });
    const { result } = renderHook(() => useElementWidth<HTMLDivElement>());

    act(() => result.current[0](document.createElement('div')));

    expect(result.current[1]).toBe(0);
  });
});
