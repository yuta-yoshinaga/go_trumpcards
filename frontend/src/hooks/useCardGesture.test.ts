import { renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { useCardGesture } from './useCardGesture';

vi.mock('./useReducedMotion', () => ({
  useReducedMotion: vi.fn(() => false),
}));

import { useReducedMotion } from './useReducedMotion';

describe('useCardGesture', () => {
  it('onClick calls onTap', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    const onTap = vi.fn();
    const { result } = renderHook(() => useCardGesture({ onTap }));
    result.current.onClick();
    expect(onTap).toHaveBeenCalledTimes(1);
  });

  it('onClick does nothing when disabled', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    const onTap = vi.fn();
    const { result } = renderHook(() => useCardGesture({ onTap, disabled: true }));
    result.current.onClick();
    expect(onTap).not.toHaveBeenCalled();
  });

  it('swipe up calls onSwipeUp when threshold exceeded', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    const onSwipeUp = vi.fn();
    const { result } = renderHook(() => useCardGesture({ onSwipeUp }));
    result.current.onPointerDown({ clientY: 200 } as React.PointerEvent);
    result.current.onPointerUp({ clientY: 150 } as React.PointerEvent);
    expect(onSwipeUp).toHaveBeenCalledTimes(1);
  });

  it('does not call onSwipeUp when threshold not exceeded', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    const onSwipeUp = vi.fn();
    const { result } = renderHook(() => useCardGesture({ onSwipeUp }));
    result.current.onPointerDown({ clientY: 200 } as React.PointerEvent);
    result.current.onPointerUp({ clientY: 190 } as React.PointerEvent);
    expect(onSwipeUp).not.toHaveBeenCalled();
  });

  it('does not call swipe handlers when disabled', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    const onSwipeUp = vi.fn();
    const { result } = renderHook(() => useCardGesture({ onSwipeUp, disabled: true }));
    result.current.onPointerDown({ clientY: 200 } as React.PointerEvent);
    result.current.onPointerUp({ clientY: 150 } as React.PointerEvent);
    expect(onSwipeUp).not.toHaveBeenCalled();
  });

  it('does not call swipe handlers when reduced motion is preferred', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    const onSwipeUp = vi.fn();
    const { result } = renderHook(() => useCardGesture({ onSwipeUp }));
    result.current.onPointerDown({ clientY: 200 } as React.PointerEvent);
    result.current.onPointerUp({ clientY: 150 } as React.PointerEvent);
    expect(onSwipeUp).not.toHaveBeenCalled();
  });

  it('does not call onSwipeUp when onSwipeUp is not provided', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    const { result } = renderHook(() => useCardGesture({}));
    result.current.onPointerDown({ clientY: 200 } as React.PointerEvent);
    result.current.onPointerUp({ clientY: 150 } as React.PointerEvent);
    // No error thrown
  });
});
