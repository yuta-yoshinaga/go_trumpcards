import { useMutation } from '@tanstack/react-query';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useOptionalSound } from '../providers/SoundProvider';
import { describeApiFailure } from '../utils/describeApiFailure';

/** Retry callback metadata consumed by ErrorAlert. */
export type GameApiRetry = (() => Promise<void>) & { retryable?: boolean };

/**
 * Central sound tap (see the sound-centralization design):
 *
 *   exec(command, ...) resolves
 *        │
 *        ├─ failure ──► consume any pending claim, play nothing
 *        │              (errorBuzz is ErrorAlert's job)
 *        └─ success ──► claim pending? ──yes──► skip (chipClick already played)
 *                            │no
 *                            ├─ rejected action? ─yes─► silent (no card moved)
 *                            │      │no
 *                            ├─ command === 'reset' ──► shuffle (deal/redeal)
 *                            └─ otherwise ────────────► cardPlace
 *
 * The sound context is read through a ref so `exec` keeps its stable `[]`
 * identity: `playSound` changes identity on every mute toggle, and 60+
 * pages key mount/reset effects on `[exec]` — folding the context into the
 * callback deps would re-deal games when the user mutes.
 */

/**
 * True when a 200 response actually reports a rejected action (illegal move,
 * out-of-turn play) rather than a state change.
 *
 * Every web presenter surfaces a rule rejection as `message = err.Error()`
 * with an EMPTY `messageCode`, while every success-path message is paired with
 * a `messageCode` (verified across all 210 `*WebPresenter.go`). So the pair is
 * a safe discriminator: no legitimate state change is ever misread as a
 * rejection. The ~11 presenters that do set a code on rejection simply fall
 * through and still sound, i.e. failures degrade to the old behavior.
 */
export function isRejectedAction(res: unknown): boolean {
  if (typeof res !== 'object' || res === null) return false;
  const r = res as { message?: unknown; messageCode?: unknown };
  return typeof r.message === 'string' && r.message.length > 0 && !r.messageCode;
}

/** Hook that wraps a game API function with loading, error, and state management. */
export function useGameApi<TState, TArgs extends unknown[]>(
  apiFn: (...args: TArgs) => Promise<TState>,
  options?: {
    onSuccess?: (res: TState, args: TArgs) => void | Promise<void>;
    onError?: (error: unknown, args: TArgs) => void;
  },
): {
  state: TState | null;
  setState: React.Dispatch<React.SetStateAction<TState | null>>;
  loading: boolean;
  error: string | null;
  exec: (...args: TArgs) => Promise<void>;
  retry: GameApiRetry;
} {
  const [state, setState] = useState<TState | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [retryable, setRetryable] = useState(true);
  // `exec` awaits the request before setting state, so a component that unmounts
  // mid-flight would otherwise still be written to on resolve. React treats that as
  // a silent no-op, which is exactly why it went unnoticed: it only breaks once the
  // surrounding environment is gone, where React's internals reach for `window` and
  // throw from `dispatchSetState`. In CI that failed whole runs while reporting every
  // test as passed. See issue #4444.
  //
  // The effect body re-arms the flag rather than relying on the `useRef(true)`
  // initialiser: StrictMode runs mount -> cleanup -> remount in dev, so a
  // cleanup-only effect would latch this `false` on a component that is genuinely
  // mounted, and every later `exec` would skip its `setState` — leaving every game
  // page on its skeleton under `bun run dev` while CI stayed green. Same pattern as
  // usePokerGame.
  const mountedRef = useRef(true);
  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);
  const apiFnRef = useRef(apiFn);
  apiFnRef.current = apiFn;
  const onSuccessRef = useRef(options?.onSuccess);
  onSuccessRef.current = options?.onSuccess;
  const onErrorRef = useRef(options?.onError);
  onErrorRef.current = options?.onError;
  const lastArgsRef = useRef<TArgs | null>(null);
  const sound = useOptionalSound();
  const soundRef = useRef(sound);
  soundRef.current = sound;

  const mutation = useMutation({
    mutationFn: (args: TArgs) => apiFnRef.current(...args),
  });
  const mutateAsyncRef = useRef(mutation.mutateAsync);
  mutateAsyncRef.current = mutation.mutateAsync;

  const execFn = useCallback(async (...args: TArgs) => {
    lastArgsRef.current = args;
    setLoading(true);
    try {
      setError(null);
      setRetryable(true);
      const res = await mutateAsyncRef.current(args);
      // Everything past this point touches the component or its side effects, so a
      // gone component gets none of it — no state write, no sound, no onSuccess.
      if (!mountedRef.current) return;
      setState(res);
      // Sound must never break exec: optional-call the context methods
      // (test files mock the provider module with partial shapes) and
      // swallow anything the audio layer throws.
      try {
        const snd = soundRef.current;
        if (snd && !snd.consumeExecClaim?.() && !isRejectedAction(res)) {
          snd.playSound?.(args[0] === 'reset' ? 'shuffle' : 'cardPlace');
        }
      } catch {
        // never let sound failures reach game state
      }
      await onSuccessRef.current?.(res, args);
    } catch (e) {
      if (!mountedRef.current) return;
      const failure = describeApiFailure(e);
      setError(failure.message);
      setRetryable(failure.retryable);
      onErrorRef.current?.(e, args);
      try {
        soundRef.current?.consumeExecClaim?.();
      } catch {
        // never let sound failures mask the API error
      }
    } finally {
      if (mountedRef.current) setLoading(false);
    }
  }, []);

  const retryRequest = useCallback(async () => {
    if (lastArgsRef.current) {
      await execFn(...lastArgsRef.current);
    }
  }, [execFn]);

  const retry = useMemo<GameApiRetry>(
    () => Object.assign(() => retryRequest(), { retryable }),
    [retryRequest, retryable],
  );
  return { state, setState, loading, error, exec: execFn, retry };
}
