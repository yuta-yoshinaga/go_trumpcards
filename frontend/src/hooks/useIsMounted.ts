import { useCallback, useEffect, useRef } from 'react';

/**
 * Returns a stable getter for whether the calling component is still mounted.
 *
 * Use it to skip state writes that would otherwise happen after an `await`:
 *
 * ```ts
 * const res = await someApi();
 * if (!isMounted()) return;
 * setThing(res);
 * ```
 *
 * React treats `setState` after unmount as a silent no-op, so the motivation is not
 * a console warning — it is that the write can land after the surrounding
 * environment is gone, where React's internals reach for `window` and throw from
 * `dispatchSetState`. In CI that failed whole runs while reporting every test as
 * passed (#4444).
 *
 * The flag is re-armed in the effect body rather than relying on the `useRef(true)`
 * initialiser. `main.tsx` wraps the app in `StrictMode`, which in dev runs
 * mount → cleanup → remount specifically to surface this: a cleanup-only effect
 * latches the flag `false` on a component that is genuinely mounted, and every
 * later guarded write is silently skipped. That is not theoretical — it would have
 * left every game page stuck on its skeleton under `bun run dev` while production
 * builds, E2E and CI all stayed green (#4446).
 */
export function useIsMounted(): () => boolean {
  const mountedRef = useRef(true);
  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);
  return useCallback(() => mountedRef.current, []);
}
