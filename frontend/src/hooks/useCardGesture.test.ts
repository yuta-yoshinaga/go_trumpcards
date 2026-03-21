import { afterEach, beforeEach, describe, expect, it, vi } from 'bun:test';
import { renderHook } from '@testing-library/react';
import { useCardGesture } from './useCardGesture';
import * as useReducedMotionModule from './useReducedMotion';

describe('useCardGesture', () => {
  let spy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    spy = vi.spyOn(useReducedMotionModule, 'useReducedMotion').mockReturnValue(false);
  });

  afterEach(() => {
    spy.mockRestore();
  });

  it('onClick calls onTap', () => {
    const onTap = vi.fn();
    const { result } = renderHook(() => useCardGesture({ onTap }));
    result.current.onClick();
    expect(onTap).toHaveBeenCalledTimes(1);
  });

  it('onClick does nothing when disabled', () => {
    const onTap = vi.fn();
    const { result } = renderHook(() => useCardGesture({ onTap, disabled: true }));
    result.current.onClick();
    expect(onTap).not.toHaveBeenCalled();
  });

  it('onClick does nothing when onTap is not provided', () => {
    const { result } = renderHook(() => useCardGesture({}));
    result.current.onClick(); // must not throw
  });

  it('swipe up calls onSwipeUp when threshold exceeded', () => {
    const onSwipeUp = vi.fn();
    const { result } = renderHook(() => useCardGesture({ onSwipeUp }));
    result.current.onPointerDown({ clientY: 200 } as React.PointerEvent);
    result.current.onPointerUp({ clientY: 150 } as React.PointerEvent);
    expect(onSwipeUp).toHaveBeenCalledTimes(1);
  });

  it('does not call onSwipeUp when threshold not exceeded', () => {
    const onSwipeUp = vi.fn();
    const { result } = renderHook(() => useCardGesture({ onSwipeUp }));
    result.current.onPointerDown({ clientY: 200 } as React.PointerEvent);
    result.current.onPointerUp({ clientY: 190 } as React.PointerEvent);
    expect(onSwipeUp).not.toHaveBeenCalled();
  });

  it('does not call swipe handlers when disabled', () => {
    const onSwipeUp = vi.fn();
    const { result } = renderHook(() => useCardGesture({ onSwipeUp, disabled: true }));
    result.current.onPointerDown({ clientY: 200 } as React.PointerEvent);
    result.current.onPointerUp({ clientY: 150 } as React.PointerEvent);
    expect(onSwipeUp).not.toHaveBeenCalled();
  });

  it('does not call swipe handlers when reduced motion is preferred', () => {
    spy.mockReturnValue(true);
    const onSwipeUp = vi.fn();
    const { result } = renderHook(() => useCardGesture({ onSwipeUp }));
    result.current.onPointerDown({ clientY: 200 } as React.PointerEvent);
    result.current.onPointerUp({ clientY: 150 } as React.PointerEvent);
    expect(onSwipeUp).not.toHaveBeenCalled();
  });

  it('does not call onSwipeUp when onSwipeUp is not provided', () => {
    const { result } = renderHook(() => useCardGesture({}));
    result.current.onPointerDown({ clientY: 200 } as React.PointerEvent);
    result.current.onPointerUp({ clientY: 150 } as React.PointerEvent);
    // No error thrown
  });

  it('does not call onTap after a swipe (prevents double-trigger)', () => {
    const onTap = vi.fn();
    const onSwipeUp = vi.fn();
    const { result } = renderHook(() => useCardGesture({ onTap, onSwipeUp }));
    result.current.onPointerDown({ clientY: 200 } as React.PointerEvent);
    result.current.onPointerUp({ clientY: 150 } as React.PointerEvent);
    expect(onSwipeUp).toHaveBeenCalledTimes(1);
    result.current.onClick();
    expect(onTap).not.toHaveBeenCalled();
  });

  it('resets swipe state on pointerCancel', () => {
    const onTap = vi.fn();
    const onSwipeUp = vi.fn();
    const { result } = renderHook(() => useCardGesture({ onTap, onSwipeUp }));
    result.current.onPointerDown({ clientY: 200 } as React.PointerEvent);
    result.current.onPointerCancel();
    result.current.onClick();
    expect(onTap).toHaveBeenCalledTimes(1);
  });

  it('allows tap after a new pointerDown resets swipe state', () => {
    const onTap = vi.fn();
    const onSwipeUp = vi.fn();
    const { result } = renderHook(() => useCardGesture({ onTap, onSwipeUp }));
    // First: swipe
    result.current.onPointerDown({ clientY: 200 } as React.PointerEvent);
    result.current.onPointerUp({ clientY: 150 } as React.PointerEvent);
    expect(onSwipeUp).toHaveBeenCalledTimes(1);
    // Second: new tap (pointerDown resets swipedRef)
    result.current.onPointerDown({ clientY: 100 } as React.PointerEvent);
    result.current.onPointerUp({ clientY: 100 } as React.PointerEvent);
    result.current.onClick();
    expect(onTap).toHaveBeenCalledTimes(1);
  });
});
