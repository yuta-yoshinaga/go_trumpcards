import { useCallback, useRef } from 'react';
import { useReducedMotion } from './useReducedMotion';

interface CardGestureOptions {
  /** Called when card is tapped. */
  onTap?: () => void;
  /** Called when card is swiped up (play/discard). */
  onSwipeUp?: () => void;
  /** Gesture disabled flag. */
  disabled?: boolean;
}

interface GestureBindHandlers {
  onClick: () => void;
  onPointerUp: (e: React.PointerEvent) => void;
  onPointerDown: (e: React.PointerEvent) => void;
}

const SWIPE_THRESHOLD = 30;

/** Hook providing tap and swipe-up gesture handlers for card interactions. */
export function useCardGesture({ onTap, onSwipeUp, disabled = false }: CardGestureOptions): GestureBindHandlers {
  const reduced = useReducedMotion();
  const startYRef = useRef(0);

  const onClick = useCallback(() => {
    if (disabled) return;
    onTap?.();
  }, [onTap, disabled]);

  const onPointerDown = useCallback(
    (e: React.PointerEvent) => {
      if (disabled || reduced) return;
      startYRef.current = e.clientY;
    },
    [disabled, reduced],
  );

  const onPointerUp = useCallback(
    (e: React.PointerEvent) => {
      if (disabled || reduced) return;
      const deltaY = startYRef.current - e.clientY;
      if (deltaY > SWIPE_THRESHOLD && onSwipeUp) {
        onSwipeUp();
      }
    },
    [onSwipeUp, disabled, reduced],
  );

  return { onClick, onPointerDown, onPointerUp };
}
