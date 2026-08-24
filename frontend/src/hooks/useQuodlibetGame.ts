import { useCallback } from 'react';
import { type QuodlibetConfigInput, quodlibetApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Quodlibet configuration. */
export const DEFAULT_QUODLIBET_CONFIG: Required<QuodlibetConfigInput> = {
  cpuDifficulty: 1,
  autoSelectContract: false,
};

/** CPU difficulty options. */
export const QUODLIBET_CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/**
 * Hook that manages Quodlibet state and its player actions: choosing the deal's
 * contract, playing a card, passing in a shedding contract, and moving on to
 * the next deal.
 */
export function useQuodlibetGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange, handleToggle } =
    useGameConfig<Required<QuodlibetConfigInput>>(DEFAULT_QUODLIBET_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(quodlibetApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Chooses this deal's contract. */
  const handleSelectContract = useCallback(
    (contract: number) => {
      void exec('contract', { contract });
    },
    [exec],
  );

  /** Plays the selected card. */
  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    void exec('play', { cardIndex: selectedCardIndices[0] });
  }, [exec, selectedCardIndices]);

  /**
   * Passes.
   *
   * **Only legal when a shedding contract leaves nothing playable** — the page
   * gates the button on the response's `canPass`, and the domain rejects it
   * again, so a stale board cannot slip a free pass through.
   */
  const handlePass = useCallback(() => {
    void exec('pass');
  }, [exec]);

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
    quodlibetConfig: config,
    handleConfigChange,
    handleToggle,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handleSelectContract,
    handlePlay,
    handlePass,
    handleNextDeal,
  };
}
