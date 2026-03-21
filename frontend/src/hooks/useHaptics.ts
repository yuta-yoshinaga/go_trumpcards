import { useCallback, useMemo } from 'react';

/** Thin wrapper around navigator.vibrate() with graceful degradation. */
export function useHaptics() {
  const canVibrate = useMemo(() => typeof navigator !== 'undefined' && typeof navigator.vibrate === 'function', []);

  const tapVibrate = useCallback(() => {
    if (canVibrate) navigator.vibrate(10);
  }, [canVibrate]);

  const selectVibrate = useCallback(() => {
    if (canVibrate) navigator.vibrate(20);
  }, [canVibrate]);

  const winVibrate = useCallback(() => {
    if (canVibrate) navigator.vibrate([50, 30, 50]);
  }, [canVibrate]);

  return { tapVibrate, selectVibrate, winVibrate, canVibrate };
}
