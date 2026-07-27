import { useCallback, useRef } from 'react';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { useIsMounted } from './useIsMounted';

/** Options for {@link useHintRequest}. */
export interface UseHintRequestOptions<TRes, THint> {
  /** Fetches the hint, e.g. `() => golfApi.exec('hint')`. */
  fetchHint: () => Promise<TRes>;
  /**
   * Pulls the hint out of the response, e.g. `(res) => res.hint`. A missing hint may
   * be returned as `undefined`; it is normalised to `null` before being stored, so
   * callers do not each have to remember the `?? null`.
   */
  selectHint: (res: TRes) => THint | null | undefined;
  setHint: (hint: THint | null) => void;
  setHintError: (message: string | null) => void;
  /** Only some games show a spinner while the hint is in flight. */
  setHintLoading?: (loading: boolean) => void;
}

/**
 * Builds the "ask the server for a hint" handler that 20 game hooks had written out
 * by hand, identically apart from which API they call and whether they track a
 * loading flag.
 *
 * Beyond the duplication, each copy had to remember that the response arrives after
 * an `await` and so may outlive the page: writing state to a gone component is a
 * silent no-op in React until the surrounding environment is torn down, at which
 * point React reaches for `window` and throws from `dispatchSetState` — failing CI
 * runs that report every test as passed. Centralising it means that reasoning lives
 * (and is tested) in one place instead of twenty. See issues #4444 and #4447.
 *
 * The returned callback is stable: options are read through a ref, so callers can
 * pass a fresh object literal every render without invalidating dependents.
 */
export function useHintRequest<TRes, THint>(options: UseHintRequestOptions<TRes, THint>): () => Promise<void> {
  const optionsRef = useRef(options);
  optionsRef.current = options;
  const isMounted = useIsMounted();

  return useCallback(async () => {
    const { fetchHint, selectHint, setHint, setHintError, setHintLoading } = optionsRef.current;
    setHintLoading?.(true);
    try {
      const res = await fetchHint();
      if (!isMounted()) return;
      setHint(selectHint(res) ?? null);
      setHintError(null);
    } catch {
      if (!isMounted()) return;
      setHintError(NETWORK_ERROR_MESSAGE());
    } finally {
      if (isMounted()) setHintLoading?.(false);
    }
  }, [isMounted]);
}
