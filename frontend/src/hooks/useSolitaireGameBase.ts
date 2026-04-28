import { useCallback, useEffect, useState } from 'react';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { useAutoCompleteState } from './useAutoCompleteState';
import { useGameApi } from './useGameApi';

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

  const onClear = options.onClearSelection;
  const hintApi = options.hintApi;
  const selectHint = options.selectHint;

  useEffect(() => {
    void apiCall(...(['reset'] as unknown as TArgs));
  }, [apiCall]);

  const runAction = useCallback(
    (...args: TArgs) => {
      onClear?.();
      setHint(null);
      void apiCall(...args);
    },
    [apiCall, onClear],
  );

  const handleReset = useCallback(() => runAction(...(['reset'] as unknown as TArgs)), [runAction]);
  const handleGiveUp = useCallback(() => runAction(...(['giveup'] as unknown as TArgs)), [runAction]);
  const handleUndo = useCallback(() => runAction(...(['undo'] as unknown as TArgs)), [runAction]);

  const handleHint = useCallback(async () => {
    if (!hintApi) return;
    try {
      const res = await hintApi();
      const value = selectHint ? selectHint(res) : (res as unknown as { hint?: THint | null }).hint;
      setHint(value ?? null);
      setHintError(null);
    } catch {
      setHintError(NETWORK_ERROR_MESSAGE());
    }
  }, [hintApi, selectHint]);

  const handleAutoComplete = useCallback(() => {
    onClear?.();
    setHint(null);
    startAutoComplete();
    void apiCall(...(['autocomplete'] as unknown as TArgs));
  }, [apiCall, onClear, startAutoComplete]);

  return {
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
  };
}
