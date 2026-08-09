import { flushPendingDispatch } from './flushPendingDispatch';

/**
 * Deterministic stand-in for `waitFor` in tests that fake `setInterval`.
 *
 * `waitFor` cannot retry in those tests. @testing-library decides whether fake
 * timers are installed by looking for Jest (`typeof jest !== 'undefined'` in
 * `jestFakeTimersAreEnabled`), which is never defined under Vitest — so it
 * always takes its real-timer path:
 *
 * ```js
 * intervalId = setInterval(checkRealTimersCallback, interval); // faked -> never fires
 * setTimeout(onTimeout, timeout);                              // real   -> still expires
 * ```
 *
 * The poll is frozen while the deadline keeps running on the wall clock, so a
 * DOM condition survives only on the MutationObserver `waitFor` also installs,
 * and a **non-DOM** condition (`expect(mockApi).toHaveBeenCalled()`) has no
 * retry mechanism at all — it passes only if it is already true at the first
 * check, and otherwise burns the full timeout and fails. Measured directly: a
 * DOM condition satisfied 120ms after the wait began passed, while the same
 * wait on a spy failed after 5005ms.
 *
 * This retries by advancing the clock instead of consuming real time, so the
 * result does not depend on how loaded the machine is.
 *
 * @param assert - Assertion to retry; the last failure is rethrown if it never passes.
 * @param turns - How many settle turns to attempt before giving up.
 */
export async function settleUntil(assert: () => void, turns = 20): Promise<void> {
  let lastError: unknown;
  for (let i = 0; i < turns; i++) {
    try {
      assert();
      return;
    } catch (error) {
      lastError = error;
    }
    await flushPendingDispatch();
  }
  throw lastError;
}
