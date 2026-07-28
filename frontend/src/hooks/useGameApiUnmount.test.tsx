import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import { StrictMode } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { useGameApi } from './useGameApi';

/**
 * Regression test for #4444.
 *
 * `exec` awaits the API call and then sets state. If the component unmounts while
 * the request is in flight, those setters used to run anyway. React treats
 * `setState` after unmount as a silent no-op, so nothing complained — until the
 * test environment was torn down first, at which point React's internals reach for
 * `window` and throw `ReferenceError: window is not defined` from `dispatchSetState`.
 * That surfaced as a whole-suite failure reporting **every test as passed**, which
 * is close to the most misleading way a build can break.
 *
 * This test reproduces it deterministically by deleting `globalThis.window` after
 * unmount and before the in-flight promise resolves. It lives in its own file
 * because of that: vitest isolates by file, so tearing `window` down here cannot
 * affect anything else.
 */

/** Renders a component using the hook and hands back its `exec` plus the unmount fn. */
function mountHarness(apiFn: () => Promise<unknown>) {
  const captured: { exec?: (...args: unknown[]) => Promise<void> } = {};
  function Harness() {
    const { exec } = useGameApi(apiFn as (...args: unknown[]) => Promise<unknown>);
    captured.exec = exec;
    return <div />;
  }
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  const { unmount } = render(
    <QueryClientProvider client={client}>
      <Harness />
    </QueryClientProvider>,
  );
  return { exec: captured.exec as (...args: unknown[]) => Promise<void>, unmount };
}

describe('useGameApi unmount safety (#4444)', () => {
  it('does not touch React state once unmounted, even after the environment is gone', async () => {
    let resolveApi: ((value: unknown) => void) | undefined;
    const apiFn = vi.fn(
      () =>
        new Promise<unknown>((resolve) => {
          resolveApi = resolve;
        }),
    );

    const { exec, unmount } = mountHarness(apiFn);
    // Start a request and leave it in flight. react-query's `mutateAsync` invokes the
    // mutation function on a microtask rather than synchronously, so the call has not
    // happened yet on the line after `exec` — see flushPendingDispatch and #4439.
    const pending = exec('reset');
    await new Promise((resolve) => {
      setTimeout(resolve, 0);
    });
    expect(apiFn).toHaveBeenCalledTimes(1);

    unmount();

    const realWindow = globalThis.window;
    try {
      // Simulate the environment teardown that turns a harmless no-op setState into
      // a thrown ReferenceError.
      // @ts-expect-error deliberately removing the global for this one assertion
      delete globalThis.window;
      resolveApi?.({ message: '', messageCode: 'x' });
      // Without the mounted guard this rejects with
      // "ReferenceError: window is not defined" from React's dispatchSetState.
      await expect(pending).resolves.toBeUndefined();
    } finally {
      globalThis.window = realWindow;
    }
  });
  // The mounted flag must be re-armed by the effect body, not just initialised at
  // `useRef(true)`. StrictMode deliberately runs mount -> cleanup -> remount in dev to
  // surface exactly this, and `main.tsx` wraps the whole app in it. With a
  // cleanup-only effect the flag latches `false` after that cycle and never recovers,
  // so every later `exec` silently skips its `setState` and every game page sits on
  // its skeleton forever under `bun run dev` — while production builds, E2E and CI all
  // stay green, because they do not double-invoke effects.
  it('still updates state after a StrictMode mount/cleanup/remount cycle', async () => {
    const apiFn = vi.fn(() => Promise.resolve({ message: '', messageCode: 'ok' }));
    const captured: { exec?: (...args: unknown[]) => Promise<void> } = {};
    function Harness() {
      const { exec, state } = useGameApi(apiFn as (...args: unknown[]) => Promise<unknown>);
      captured.exec = exec;
      return <div data-testid="loaded">{state === null ? 'empty' : 'loaded'}</div>;
    }
    const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
    render(
      <StrictMode>
        <QueryClientProvider client={client}>
          <Harness />
        </QueryClientProvider>
      </StrictMode>,
    );
    expect(screen.getByTestId('loaded')).toHaveTextContent('empty');

    await captured.exec?.('reset');
    await waitFor(() => expect(screen.getByTestId('loaded')).toHaveTextContent('loaded'));
  });
});
