import { useCallback } from 'react';
import { type CostlyColoursConfigInput, costlycoloursApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Costly Colours configuration. */
export const DEFAULT_COSTLYCOLOURS_CONFIG: Required<CostlyColoursConfigInput> = {
  cpuDifficulty: 1,
  targetScore: 61,
};

/** CPU difficulty options. */
export const COSTLYCOLOURS_CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/**
 * Target-score options.
 *
 * **Sources split on the target**: Cotton (1674) says 61, Parlett 121. Both are
 * real, so both are offered and 61 — the original — is the default.
 */
export const COSTLYCOLOURS_TARGET_OPTIONS = [31, 61, 121] as const;

/**
 * Hook that manages Costly Colours state and its player actions: accepting or
 * refusing the exchange, playing a card, and moving on.
 *
 * **Which cards are playable comes from the server**, since a card that would
 * take the count past 31 cannot be played.
 */
export function useCostlyColoursGame() {
  const { config, handleConfigChange } =
    useGameConfig<Required<CostlyColoursConfigInput>>(DEFAULT_COSTLYCOLOURS_CONFIG);
  const { state, loading, error, exec, retry } = useGameApi(costlycoloursApi.exec);

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /**
   * Answers the exchange offer.
   *
   * **The decision is always explicit.** Refusing pegs a point for the
   * opponent, so it is never defaulted on either side of the wire.
   */
  const mog = useCallback(
    (accept: boolean) => {
      void exec('mog', { accept });
    },
    [exec],
  );

  /** Plays the hand card at `idx`. */
  const play = useCallback(
    (idx: number) => {
      void exec('play', { handIndex: idx });
    },
    [exec],
  );

  /** Advances to the next deal. */
  const handleNextDeal = useCallback(() => {
    void exec('nextdeal');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    costlyColoursConfig: config,
    handleConfigChange,
    mog,
    play,
    handleNextDeal,
    reset,
  };
}
