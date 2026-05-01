import { useEffect, useRef } from 'react';

/**
 * Calls the supplied api function with `"reset"` exactly once on mount.
 *
 * Every game page used to repeat the same `useEffect(() => apiCall("reset"), [apiCall])`
 * block to fetch a fresh game state from the server when the page renders.
 * This hook captures that pattern in one place so a future change (e.g.,
 * "restore from session if available, otherwise reset") only has to be made
 * here. The api callback is captured in a ref so identity changes between
 * renders never re-trigger the reset — the effect intentionally fires once
 * per mount.
 */
export function useMountReset(apiExec: (...args: ['reset']) => unknown): void {
  const apiExecRef = useRef(apiExec);
  apiExecRef.current = apiExec;

  useEffect(() => {
    void apiExecRef.current('reset');
  }, []);
}
