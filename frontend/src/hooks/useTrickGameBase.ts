import { useCallback, useEffect, useRef, useState } from 'react';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';
import { useIsMounted } from './useIsMounted';

/**
 * Options passed to {@link useTrickGameBase} to configure the base trick-taking game hook.
 * @template TState - The game state type returned by the API.
 * @template TArgs - The tuple of argument types accepted by the API exec function.
 * @template TConfig - The game configuration type.
 * @template THint - The hint type embedded in the game state.
 */
export interface TrickGameBaseOptions<TState, TArgs extends unknown[], TConfig extends object, THint> {
  /** The API exec function used to dispatch commands. */
  apiFn: (...args: TArgs) => Promise<TState>;
  /** The default game configuration applied on reset. */
  defaultConfig: TConfig;
  /** Extracts the hint from the game state, or returns null/undefined if absent. */
  getHint: (state: TState) => THint | null | undefined;
}

/**
 * Return value of {@link useTrickGameBase}.
 * @template TState - The game state type returned by the API.
 * @template TArgs - The tuple of argument types accepted by the API exec function.
 * @template TConfig - The game configuration type.
 * @template THint - The hint type.
 */
export interface TrickGameBaseResult<TState, TArgs extends unknown[], TConfig extends object, THint> {
  /** Current game state, or null before the first successful API call. */
  state: TState | null;
  /** True while an API call is in flight. */
  loading: boolean;
  /** Error message from the last failed API call, or null. */
  error: string | null;
  /** The latest hint returned by the hint command, or null. */
  hint: THint | null;
  /** Error message from the last failed hint request, or null. */
  hintError: string | null;
  /** True while a hint request is in flight. */
  hintLoading: boolean;
  /** Raw exec function to dispatch arbitrary API commands. */
  exec: (...args: TArgs) => Promise<void>;
  /** Current game configuration. */
  config: TConfig;
  /** Indices of currently selected cards. */
  selectedCardIndices: number[];
  /** Toggles selection of the card at the given index. */
  toggleCard: (idx: number) => void;
  /** Clears all selected cards. */
  clearSelection: () => void;
  /** Updates a numeric config field from a string value. */
  handleConfigChange: ReturnType<typeof useGameConfig<TConfig>>['handleConfigChange'];
  /** Toggles a boolean config field. */
  handleToggle: ReturnType<typeof useGameConfig<TConfig>>['handleToggle'];
  /** Plays the single currently-selected card. No-op if zero or multiple cards are selected. */
  handlePlay: () => void;
  /** Advances to the next trick. */
  handleNextTrick: () => void;
  /** Advances to the next round. */
  handleNextRound: () => void;
  /** Fetches a hint from the server. */
  handleHint: () => Promise<void>;
  /** Retries the last failed API call. */
  retry: () => Promise<void>;
}

/**
 * Factory hook that encapsulates the common logic shared by all trick-taking game hooks
 * (Hearts, Spades, etc.). Handles card selection, config management, hint state, and
 * the standard set of game commands (play, next trick, next round, hint).
 *
 * Game-specific actions (e.g. `handlePass` for Hearts, `handleBid` for Spades) should
 * be built on top of the returned `exec` function in the game-specific hook.
 *
 * @template TState - The game state type returned by the API.
 * @template TArgs - The tuple of argument types accepted by the API exec function.
 * @template TConfig - The game configuration type.
 * @template THint - The hint type embedded in the game state.
 * @param options - Configuration for the base hook.
 * @returns Common trick-taking game state and action handlers.
 */
export function useTrickGameBase<TState, TArgs extends unknown[], TConfig extends object, THint>(
  options: TrickGameBaseOptions<TState, TArgs, TConfig, THint>,
): TrickGameBaseResult<TState, TArgs, TConfig, THint> {
  const { apiFn, defaultConfig, getHint } = options;

  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange, handleToggle } = useGameConfig<TConfig>(defaultConfig);
  const [hint, setHint] = useState<THint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const [hintLoading, setHintLoading] = useState(false);

  const onSuccess = useCallback(() => {
    clearSelection();
    setHint(null);
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(apiFn, { onSuccess });

  const defaultConfigRef = useRef(defaultConfig);

  // Trick-taking APIs all share the shape (command, arg1?, arg2?, config?).
  // The generic TArgs tuple prevents a direct call, so each dispatch casts
  // through this widened command signature. Casts are compile-time only.
  type ExecCmd = (command: string, ...rest: unknown[]) => Promise<void>;
  type ApiCmd = (command: string, ...rest: unknown[]) => Promise<TState>;

  useEffect(() => {
    (exec as unknown as ExecCmd)('reset', undefined, undefined, defaultConfigRef.current);
  }, [exec]);

  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    (exec as unknown as ExecCmd)('play', undefined, selectedCardIndices[0]);
  }, [exec, selectedCardIndices]);

  const handleNextTrick = useCallback(() => {
    (exec as unknown as ExecCmd)('next');
  }, [exec]);

  const handleNextRound = useCallback(() => {
    (exec as unknown as ExecCmd)('nextround');
  }, [exec]);

  const isMounted = useIsMounted();

  const handleHint = useCallback(async () => {
    setHintLoading(true);
    try {
      const res = await (apiFn as unknown as ApiCmd)('hint');
      // The player may have navigated away while the hint was in flight; writing
      // state to a gone component can throw once the environment is torn down. #4447
      if (!isMounted()) return;
      setHint(getHint(res) ?? null);
      setHintError(null);
    } catch {
      if (!isMounted()) return;
      setHintError(NETWORK_ERROR_MESSAGE());
    } finally {
      if (isMounted()) setHintLoading(false);
    }
  }, [apiFn, getHint, isMounted]);

  return {
    state,
    loading,
    error,
    hint,
    hintError,
    hintLoading,
    exec,
    config,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleToggle,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    handleHint,
    retry,
  };
}
