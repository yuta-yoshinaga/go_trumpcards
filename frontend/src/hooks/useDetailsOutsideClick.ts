import { type RefObject, useEffect } from 'react';

/**
 * Close any open `<details>` element under `containerRef` when the user
 * mousedowns outside that details. Only runs while `enabled` is true.
 *
 * The mobile/tablet nav keeps categories always-open, so the caller passes
 * `enabled = false` on those breakpoints to avoid the layout shift between
 * mousedown and mouseup that would otherwise land clicks on the wrong link.
 */
export function useDetailsOutsideClick(containerRef: RefObject<HTMLElement | null>, enabled: boolean): void {
  useEffect(() => {
    if (!enabled) return;
    const handleOutsideClick = (e: MouseEvent) => {
      if (!containerRef.current) return;
      const openDetails = containerRef.current.querySelectorAll('details[open]');
      for (const details of openDetails) {
        if (!details.contains(e.target as Node)) {
          details.removeAttribute('open');
        }
      }
    };
    document.addEventListener('mousedown', handleOutsideClick);
    return () => document.removeEventListener('mousedown', handleOutsideClick);
  }, [enabled, containerRef]);
}
