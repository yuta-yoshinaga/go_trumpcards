import { useCallback } from 'react';
import { type MusConfigInput, musApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Mus game configuration. */
export const DEFAULT_MUS_CONFIG: Required<MusConfigInput> = {
  cpuDifficulty: 1,
  targetAmarrakos: 40,
};

/** CPU difficulty level options for Mus. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target-amarrakos options for Mus. */
export const TARGET_AMARRAKOS_OPTIONS = [30, 40, 50] as const;

/**
 * Hook that manages Mus game state and the vying (betting) player actions.
 *
 * Mus is not trick-taking, so unlike most card games it builds its command set
 * directly on {@link useGameApi}. The flow is: Mus (call mus/cut) → Discard
 * (exchange cards) → four betting rounds (Grande/Chica/Pares/Juego) where the
 * player paso/envido/ordago/quiero/noquiero → Showdown → next round.
 */
export function useMusGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<MusConfigInput>>(DEFAULT_MUS_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(musApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /**
   * Declares Mus (`true` = call Mus and exchange) or Corte/cut (`false` =
   * end the Mus phase and start betting) in the Mus phase.
   */
  const handleMus = useCallback(
    (mus: boolean) => {
      void exec('mus', { mus });
    },
    [exec],
  );

  /** Exchanges the currently-selected cards in the Discard phase (empty keeps all). */
  const handleDiscard = useCallback(() => {
    void exec('discard', { discardIndices: selectedCardIndices });
  }, [exec, selectedCardIndices]);

  /**
   * Submits a bet action in a betting round.
   *
   * @param action - 0=paso 1=envido 2=ordago 3=quiero 4=noquiero
   * @param amount - the Envido stake (ignored by other actions)
   */
  const handleBet = useCallback(
    (action: number, amount = 0) => {
      void exec('bet', { betAction: action, betAmount: amount });
    },
    [exec],
  );

  /** Advances to the next round. */
  const handleNextRound = useCallback(() => {
    void exec('next');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    musConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handleMus,
    handleDiscard,
    handleBet,
    handleNextRound,
  };
}
