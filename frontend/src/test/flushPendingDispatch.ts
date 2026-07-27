import { vi } from 'vitest';

/**
 * Waits until a command triggered by `fireEvent` has actually reached the mocked
 * game API, so that a "nothing was dispatched" assertion means something.
 *
 * `exec` from {@link hooks/useGameApi.useGameApi | useGameApi} dispatches through
 * react-query's `mutateAsync`, which invokes the mutation function on a microtask
 * rather than synchronously. So this:
 *
 * ```ts
 * fireEvent.keyDown(document, { key: 'b' });
 * expect(mockApi).not.toHaveBeenCalled(); // passes even when 'b' DID fire
 * ```
 *
 * passes unconditionally — at that instant the call has not happened yet, whether
 * or not the key was bound. Awaiting this first makes the assertion real:
 *
 * ```ts
 * fireEvent.keyDown(document, { key: 'b' });
 * await flushPendingDispatch();
 * expect(mockApi).not.toHaveBeenCalled(); // now fails if 'b' fired
 * ```
 *
 * A single microtask is enough to observe the call, but this yields a macrotask so
 * that actions chaining more than one `await` (a rebet that advances the round and
 * then re-stakes it, say) are also fully settled.
 *
 * Under fake timers a plain `setTimeout` would never fire, so awaiting one would
 * hang until the test timed out rather than flush anything — that is not
 * hypothetical, it stalled all 24 of `PokerPage.test.tsx`'s fake-timer tests for
 * 10s each. So the timer is advanced explicitly when they are installed.
 *
 * Positive assertions do not need this — `waitFor` already retries.
 */
export async function flushPendingDispatch(): Promise<void> {
  if (vi.isFakeTimers()) {
    await vi.advanceTimersByTimeAsync(0);
    return;
  }
  await new Promise((resolve) => {
    setTimeout(resolve, 0);
  });
}
