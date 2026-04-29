import { type RefObject, useEffect, useRef } from 'react';
import { getFocusableElements } from '../utils/dom';

/**
 * Mobile nav focus trap with focus restore.
 *
 * - When `isOpen` flips to true, focus moves to the first focusable element inside `containerRef`.
 * - When `isOpen` flips to false, focus is restored to `restoreRef` (the toggle button).
 * - While `isActive` (mobile only) and `isOpen` are both true, Tab/Shift+Tab cycle inside the container.
 *
 * The trap is intentionally skipped on tablet+ because the nav renders inline
 * on those breakpoints and a trap would prevent normal Tab flow into page content.
 */
export function useNavFocusTrap(
  containerRef: RefObject<HTMLElement | null>,
  restoreRef: RefObject<HTMLElement | null>,
  isOpen: boolean,
  isActive: boolean,
): void {
  const wasOpen = useRef(false);

  useEffect(() => {
    const justOpened = isOpen && !wasOpen.current;
    const justClosed = !isOpen && wasOpen.current;
    wasOpen.current = isOpen;

    if (justClosed && restoreRef.current) {
      restoreRef.current.focus();
    }

    if (!isOpen || !containerRef.current) return;
    const container = containerRef.current;

    // Only move focus on the open transition; re-running the effect for a
    // viewport resize (isActive flip) must not steal focus from the user.
    if (justOpened) {
      const initial = getFocusableElements(container)[0];
      initial?.focus();
    }

    if (!isActive) return;

    const handleTab = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return;
      const focusable = getFocusableElements(container);
      if (focusable.length === 0) return;
      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      const active = document.activeElement as HTMLElement | null;
      if (e.shiftKey) {
        if (active === first || !container.contains(active)) {
          e.preventDefault();
          last.focus();
        }
      } else {
        if (active === last || !container.contains(active)) {
          e.preventDefault();
          first.focus();
        }
      }
    };

    document.addEventListener('keydown', handleTab);
    return () => document.removeEventListener('keydown', handleTab);
  }, [isOpen, isActive, containerRef, restoreRef]);
}
