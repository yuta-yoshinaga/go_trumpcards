import { useEffect } from 'react';

/**
 * Warns the user before leaving a game page with an in-progress round.
 *
 * Covers browser-level navigation (tab close, reload, external link) via
 * `beforeunload`. SPA in-app navigation is intentionally not intercepted to
 * keep the hook independent of the router context; the guard is paired with
 * the existing `GameResetDialog` for explicit reset flows.
 */
export function useGameLeaveGuard(active: boolean, confirmMessage: string) {
  useEffect(() => {
    if (!active) return;
    const handler = (e: BeforeUnloadEvent) => {
      e.preventDefault();
      e.returnValue = confirmMessage;
      return confirmMessage;
    };
    window.addEventListener('beforeunload', handler);
    return () => window.removeEventListener('beforeunload', handler);
  }, [active, confirmMessage]);
}
