import { useCallback } from 'react';
import { type KingConfigInput, kingApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default King game configuration. */
export const DEFAULT_KING_CONFIG: Required<KingConfigInput> = {
  cpuDifficulty: 1,
};

/** CPU difficulty level options for King. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/**
 * Hook that manages King game state and its player actions: select the deal's
 * contract, play a card, plus advance to the next deal and request a hint.
 *
 * King is a 4-player 52-card compendium trick-avoidance game. There is no
 * bidding and no talon exchange; each deal the dealer simply chooses one of the
 * seven unused contracts, then all seats play thirteen must-follow tricks. The
 * command set is built directly on {@link useGameApi}.
 */
export function useKingGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<KingConfigInput>>(DEFAULT_KING_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(kingApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /**
   * Selects the deal's contract (0..6). For contract 6 ("King (Trump)") a trump
   * suit (1..4) must be supplied; for all other contracts pass -1.
   */
  const selectContract = useCallback(
    (contract: number, trumpSuit = -1) => {
      void exec('contract', { contract, trumpSuit });
    },
    [exec],
  );

  /** Plays the single currently-selected card in the Play phase. */
  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    void exec('play', { handIndex: selectedCardIndices[0] });
  }, [exec, selectedCardIndices]);

  /** Advances to the next deal. */
  const handleNextDeal = useCallback(() => {
    void exec('next');
  }, [exec]);

  /** Requests a backend hint. */
  const hint = useCallback(() => {
    void exec('hint');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    kingConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    selectContract,
    handlePlay,
    handleNextDeal,
    hint,
  };
}
