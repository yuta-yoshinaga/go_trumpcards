import { useLayoutEffect } from 'react';

/**
 * Locks body scroll while `active` is true. Saves and restores the previous
 * `document.body.style.overflow` on unmount or when `active` becomes false,
 * so nested modals stack cleanly without leaking `overflow: hidden` after
 * the last modal closes. Uses useLayoutEffect so the scrollbar toggle
 * happens synchronously before the browser paints, avoiding a flicker
 * where the modal appears one frame ahead of the scroll lock.
 */
export function useBodyScrollLock(active: boolean): void {
  useLayoutEffect(() => {
    if (!active) return;
    const prev = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => {
      document.body.style.overflow = prev;
    };
  }, [active]);
}
