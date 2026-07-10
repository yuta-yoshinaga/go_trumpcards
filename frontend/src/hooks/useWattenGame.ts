import { useCallback } from 'react';
import { type WattenConfigInput, wattenApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Watten (ヴァッテン) game configuration. */
export const DEFAULT_WATTEN_CONFIG: Required<WattenConfigInput> = {
  cpuDifficulty: 1,
  targetScore: 15,
  maxRaises: 4,
};

/** CPU difficulty level options for Watten. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target-score options for Watten (first team to reach it wins the match). */
export const TARGET_SCORE_OPTIONS = [11, 15, 21] as const;

/**
 * Hook that manages Watten (ヴァッテン) game state and its player actions.
 *
 * Watten is a Bavarian/Austrian 4-player, 2-team trick-taker with a raise/bluff
 * stake mechanic. The command set is built directly on {@link useGameApi}: the
 * dealer declares a Schlag rank + critical suit (`declare`), plays a card
 * (`play`), may raise the stake (`raise`), holds/folds a pending raise
 * (`respond`), and advances to the next deal (`nextround`).
 */
export function useWattenGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<WattenConfigInput>>(DEFAULT_WATTEN_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(wattenApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', undefined, undefined, undefined, undefined, config);
  }, [exec, config]);

  /** Declares the Schlag rank + critical (trump) suit (dealer only, Declare phase). */
  const handleDeclare = useCallback(
    (rank: number, suit: number) => {
      void exec('declare', rank, suit);
    },
    [exec],
  );

  /** Plays the single currently-selected card in the Play phase. */
  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    void exec('play', undefined, undefined, selectedCardIndices[0]);
  }, [exec, selectedCardIndices]);

  /** Raises the stake (human lead only, when the response allows it). */
  const handleRaise = useCallback(() => {
    void exec('raise');
  }, [exec]);

  /** Responds to a pending raise: hold (accept) or fold (concede). */
  const handleRespond = useCallback(
    (hold: boolean) => {
      void exec('respond', undefined, undefined, undefined, hold);
    },
    [exec],
  );

  /** Advances to the next deal (after a deal completes). */
  const handleNextRound = useCallback(() => {
    void exec('nextround');
  }, [exec]);

  /** Fetches a server hint for the current turn. */
  const handleHint = useCallback(() => {
    void exec('hint');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    wattenConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handleDeclare,
    handlePlay,
    handleRaise,
    handleRespond,
    handleNextRound,
    handleHint,
  };
}
