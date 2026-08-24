import { useCallback } from 'react';
import { type UnsunKarutaConfigInput, unsunKarutaApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Unsun Karuta (八人メリ) configuration. */
export const DEFAULT_UNSUN_KARUTA_CONFIG: Required<UnsunKarutaConfigInput> = {
  cpuDifficulty: 1,
  targetDeals: 4,
};

/** CPU difficulty options. */
export const UNSUN_CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/**
 * Match lengths.
 *
 * **Eight is a full circuit** — one deal per seat, which is where the
 * traditional game stops.
 */
export const UNSUN_TARGET_DEALS_OPTIONS = [1, 2, 4, 8] as const;

/**
 * Hook that manages Unsun Karuta state and its player actions: playing a card,
 * declaring meri/monchi with a lead, and trick/deal advancement.
 */
export function useUnsunKarutaGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<UnsunKarutaConfigInput>>(DEFAULT_UNSUN_KARUTA_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(unsunKarutaApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /**
   * Plays the selected card.
   *
   * **The declaration rides with the card.** Sending it on its own would leave
   * a board where someone has declared but not played.
   */
  const handlePlay = useCallback(
    (declare: boolean) => {
      if (selectedCardIndices.length !== 1) return;
      void exec('play', { cardIndex: selectedCardIndices[0], declare });
    },
    [exec, selectedCardIndices],
  );

  /** Advances to the next trick. */
  const handleNextTrick = useCallback(() => {
    void exec('next');
  }, [exec]);

  /** Advances to the next deal. */
  const handleNextRound = useCallback(() => {
    void exec('nextround');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    unsunConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  };
}
