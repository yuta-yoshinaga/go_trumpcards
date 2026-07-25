import { type RefObject, useEffect, useRef } from 'react';
import { getFocusableElements } from '../utils/dom';

/**
 * Traps focus inside `containerRef` while `open`, closes on Escape, and restores
 * focus to the previously-focused element on close. The single source of truth
 * for modal-dialog keyboard behavior (issue #4312): on open it focuses the first
 * focusable descendant and cycles Tab / Shift+Tab within the container; Escape
 * calls `onClose`; unmount/close returns focus to whatever had it when the
 * dialog opened.
 *
 * Body scroll-locking is handled separately (by {@link useBodyScrollLock},
 * wired in the shared `Modal` component) so this hook stays focus-only and
 * reusable by dialogs that manage their own overlay.
 */
export function useFocusTrap(containerRef: RefObject<HTMLElement | null>, open: boolean, onClose: () => void): void {
  const triggerRef = useRef<Element | null>(null);
  // Keep the latest onClose without re-running the trap effect when a parent
  // passes a new inline callback each render.
  const onCloseRef = useRef(onClose);
  onCloseRef.current = onClose;

  useEffect(() => {
    if (!open || !containerRef.current) return;
    triggerRef.current = document.activeElement;
    const container = containerRef.current;

    // Focus the first focusable child, or the container itself (if it is
    // programmatically focusable, e.g. has tabIndex={-1}) when it has none — so
    // a content-less modal still traps focus rather than leaving it in the page.
    const focusable = getFocusableElements(container);
    (focusable[0] ?? container).focus();

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onCloseRef.current();
        return;
      }
      if (e.key !== 'Tab') return;
      const items = getFocusableElements(container);
      if (items.length === 0) return;
      const first = items[0];
      const last = items[items.length - 1];
      if (e.shiftKey) {
        if (document.activeElement === first || !container.contains(document.activeElement)) {
          e.preventDefault();
          last.focus();
        }
      } else {
        if (document.activeElement === last || !container.contains(document.activeElement)) {
          e.preventDefault();
          first.focus();
        }
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      if (triggerRef.current instanceof HTMLElement) {
        triggerRef.current.focus();
      }
    };
  }, [open, containerRef]);
}
