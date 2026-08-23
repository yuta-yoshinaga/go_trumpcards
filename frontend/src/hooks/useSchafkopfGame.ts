import { useCallback } from 'react';
import { type SchafkopfConfigInput, type SchafkopfContract, schafkopfApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Schafkopf game configuration. */
export const DEFAULT_SCHAFKOPF_CONFIG: Required<SchafkopfConfigInput> = {
  cpuDifficulty: 1,
  baseChips: 1,
  startChips: 100,
  targetChips: 200,
};

/** CPU difficulty level options for Schafkopf. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target-chip options for Schafkopf. */
export const TARGET_CHIPS_OPTIONS = [150, 200, 300, 500] as const;

/**
 * Hook that manages Schafkopf game state and the multi-phase player actions
 * (declare/pass, call, play) plus trick/round advancement.
 *
 * Unlike the simple trick-taking games that reuse `useTrickGameBase`,
 * Schafkopf has a Pick → Call → Play flow, so the command set is built
 * directly on {@link useGameApi}.
 */
export function useSchafkopfGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<SchafkopfConfigInput>>(DEFAULT_SCHAFKOPF_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(schafkopfApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /**
   * Declares a contract in the Pick phase.
   *
   * `soloSuit` is only read for Solo, and the caller has to supply it — a
   * default here would silently turn a mis-clicked Solo into a spade Solo.
   */
  const handleDeclare = useCallback(
    (contract: SchafkopfContract, soloSuit?: number) => {
      void exec('pick', { pick: true, contract, soloSuit });
    },
    [exec],
  );

  /** Declares Rufspiel, the ace-calling contract. */
  const handlePick = useCallback(() => {
    handleDeclare(0);
  }, [handleDeclare]);

  /** Passes in the Pick phase, declaring nothing. */
  const handlePass = useCallback(() => {
    void exec('pick', { pick: false });
  }, [exec]);

  /** Calls the partner suit (1=♠ 2=♣ 3=♥) in the Call phase. */
  const handleCall = useCallback(
    (suit: number) => {
      void exec('call', { callSuit: suit });
    },
    [exec],
  );

  /** Plays the single currently-selected card in the Play phase. */
  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    void exec('play', { cardIndex: selectedCardIndices[0] });
  }, [exec, selectedCardIndices]);

  /** Advances to the next trick. */
  const handleNextTrick = useCallback(() => {
    void exec('next');
  }, [exec]);

  /** Advances to the next round. */
  const handleNextRound = useCallback(() => {
    void exec('nextround');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    schafkopfConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handlePick,
    handlePass,
    handleDeclare,
    handleCall,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  };
}
