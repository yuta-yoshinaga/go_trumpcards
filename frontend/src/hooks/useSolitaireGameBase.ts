import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { useAutoCompleteState } from './useAutoCompleteState';
import { useGameApi } from './useGameApi';
import { useIsMounted } from './useIsMounted';

/** Options for {@link useSolitaireGameBase}. */
export interface SolitaireGameBaseOptions<THint, THintRes> {
  /**
   * Game-specific cleanup callback fired before any of the shared actions
   * (reset, giveup, undo, autocomplete) so hooks can clear their own
   * selection state in lockstep.
   */
  onClearSelection?: () => void;
  /**
   * Hint API call. Returning `null`/`undefined` clears the displayed hint.
   * Errors are caught and surfaced as `hintError`.
   */
  hintApi?: () => Promise<THintRes>;
  /** Picks the hint payload out of the hint API response (default: `res.hint`). */
  selectHint?: (res: THintRes) => THint | null | undefined;
}

/** Public surface of {@link useSolitaireGameBase}. */
export interface SolitaireGameBase<TState, TArgs extends unknown[], THint> {
  state: TState | null;
  loading: boolean;
  error: string | null;
  retry: () => Promise<void>;
  apiCall: (...args: TArgs) => Promise<void>;
  hint: THint | null;
  setHint: React.Dispatch<React.SetStateAction<THint | null>>;
  hintError: string | null;
  isAutoCompleting: boolean;
  startAutoComplete: () => void;
  /** Calls `onClearSelection`, clears hint, then forwards args to `apiCall`. */
  runAction: (...args: TArgs) => void;
  /** Convenience: `runAction('reset' as TArgs[0])`. */
  handleReset: () => void;
  /** Convenience: `runAction('giveup' as TArgs[0])`. */
  handleGiveUp: () => void;
  /** Convenience: `runAction('undo' as TArgs[0])`. */
  handleUndo: () => void;
  /** Wraps `hintApi()` in try/catch; clears hint on missing payload. */
  handleHint: () => Promise<void>;
  /** Calls cleanup, starts the auto-complete state machine, then forwards `'autocomplete'` to apiCall. */
  handleAutoComplete: () => void;
}

/**
 * Shared scaffolding for solitaire-style game hooks. Wraps {@link useGameApi}
 * with the hint state, auto-complete state, the on-mount initial reset, and
 * the four "clear-selection + clear-hint + apiCall(cmd)" actions every solitaire
 * game replicated by hand. Game-specific selection state and selection
 * handlers stay in the per-game hook.
 *
 * The returned object is memoized and the action callbacks are kept stable
 * across renders even when the caller passes an inline options literal — see
 * the optionsRef below — so per-game hooks can safely list the returned
 * `base` object (or its individual properties) in their own useCallback /
 * useEffect deps without re-creating their handlers on every render.
 *
 * @typeParam TState  Response shape returned by the game's API.
 * @typeParam TArgs   Argument tuple of the game's API exec function.
 * @typeParam THint   Hint payload type for that game.
 */
export function useSolitaireGameBase<TState, TArgs extends unknown[], THint, THintRes = TState>(
  apiFn: (...args: TArgs) => Promise<TState>,
  options: SolitaireGameBaseOptions<THint, THintRes> = {},
): SolitaireGameBase<TState, TArgs, THint> {
  const { state, loading, error, exec: apiCall, retry } = useGameApi(apiFn);
  const [hint, setHint] = useState<THint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const { isAutoCompleting, startAutoComplete } = useAutoCompleteState();

  // Keep the latest options in a ref so the action callbacks below stay
  // stable even when the caller passes an inline `{ hintApi: () => ... }`
  // literal (a fresh reference every render). PR #1573 review.
  const optionsRef = useRef(options);
  optionsRef.current = options;

  useEffect(() => {
    void apiCall(...(['reset'] as unknown as TArgs));
  }, [apiCall]);

  const runAction = useCallback(
    (...args: TArgs) => {
      optionsRef.current.onClearSelection?.();
      setHint(null);
      void apiCall(...args);
    },
    [apiCall],
  );

  const isMounted = useIsMounted();

  const handleReset = useCallback(() => runAction(...(['reset'] as unknown as TArgs)), [runAction]);
  const handleGiveUp = useCallback(() => runAction(...(['giveup'] as unknown as TArgs)), [runAction]);
  const handleUndo = useCallback(() => runAction(...(['undo'] as unknown as TArgs)), [runAction]);

  const handleHint = useCallback(async () => {
    const { hintApi, selectHint } = optionsRef.current;
    if (!hintApi) return;
    try {
      const res = await hintApi();
      // The player may have navigated away while the hint was in flight; writing
      // state to a gone component can throw once the environment is torn down. #4447
      if (!isMounted()) return;
      const value = selectHint ? selectHint(res) : (res as unknown as { hint?: THint | null }).hint;
      setHint(value ?? null);
      setHintError(null);
    } catch {
      if (!isMounted()) return;
      setHintError(NETWORK_ERROR_MESSAGE());
    }
  }, [isMounted]);

  const handleAutoComplete = useCallback(() => {
    optionsRef.current.onClearSelection?.();
    setHint(null);
    startAutoComplete();
    void apiCall(...(['autocomplete'] as unknown as TArgs));
  }, [apiCall, startAutoComplete]);

  // Memoize the returned object so per-game hooks can safely depend on
  // `base` without their own callbacks tearing down on every render.
  return useMemo(
    () => ({
      state,
      loading,
      error,
      retry,
      apiCall,
      hint,
      setHint,
      hintError,
      isAutoCompleting,
      startAutoComplete,
      runAction,
      handleReset,
      handleGiveUp,
      handleUndo,
      handleHint,
      handleAutoComplete,
    }),
    [
      state,
      loading,
      error,
      retry,
      apiCall,
      hint,
      hintError,
      isAutoCompleting,
      startAutoComplete,
      runAction,
      handleReset,
      handleGiveUp,
      handleUndo,
      handleHint,
      handleAutoComplete,
    ],
  );
}
