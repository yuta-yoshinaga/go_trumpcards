/** Thin wrapper around navigator.vibrate() with graceful degradation. */
export function useHaptics() {
  const canVibrate = typeof navigator !== 'undefined' && typeof navigator.vibrate === 'function';

  const tapVibrate = () => {
    if (canVibrate) navigator.vibrate(10);
  };

  const selectVibrate = () => {
    if (canVibrate) navigator.vibrate(20);
  };

  const winVibrate = () => {
    if (canVibrate) navigator.vibrate([50, 30, 50]);
  };

  return { tapVibrate, selectVibrate, winVibrate, canVibrate };
}
