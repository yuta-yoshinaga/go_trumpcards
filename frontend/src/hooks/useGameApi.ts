import { useMutation } from '@tanstack/react-query';
import { useCallback, useRef, useState } from 'react';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { useOptionalSound } from '../providers/SoundProvider';

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
function isRejectedAction(res: unknown): boolean {
  if (typeof res !== 'object' || res === null) return false;
  const r = res as { message?: unknown; messageCode?: unknown };
  return typeof r.message === 'string' && r.message.length > 0 && !r.messageCode;
}

/** Hook that wraps a game API function with loading, error, and state management. */
export function useGameApi<TState, TArgs extends unknown[]>(
  apiFn: (...args: TArgs) => Promise<TState>,
  options?: { onSuccess?: (res: TState) => void | Promise<void> },
): {
  state: TState | null;
  setState: React.Dispatch<React.SetStateAction<TState | null>>;
  loading: boolean;
  error: string | null;
  exec: (...args: TArgs) => Promise<void>;
  retry: () => Promise<void>;
} {
  const [state, setState] = useState<TState | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const apiFnRef = useRef(apiFn);
  apiFnRef.current = apiFn;
  const onSuccessRef = useRef(options?.onSuccess);
  onSuccessRef.current = options?.onSuccess;
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
      const res = await mutateAsyncRef.current(args);
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
      await onSuccessRef.current?.(res);
    } catch {
      setError(NETWORK_ERROR_MESSAGE());
      try {
        soundRef.current?.consumeExecClaim?.();
      } catch {
        // never let sound failures mask the API error
      }
    } finally {
      setLoading(false);
    }
  }, []);

  const retry = useCallback(async () => {
    if (lastArgsRef.current) {
      await execFn(...lastArgsRef.current);
    }
  }, [execFn]);

  return { state, setState, loading, error, exec: execFn, retry };
}
