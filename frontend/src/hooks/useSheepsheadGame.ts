import { useCallback } from 'react';
import { type SheepsheadConfigInput, sheepsheadApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Sheepshead game configuration. */
export const DEFAULT_SHEEPSHEAD_CONFIG: Required<SheepsheadConfigInput> = {
  cpuDifficulty: 1,
  baseChips: 1,
  startChips: 100,
  targetChips: 200,
};

/** CPU difficulty level options for Sheepshead. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target-chip options for Sheepshead. */
export const TARGET_CHIPS_OPTIONS = [150, 200, 300, 500] as const;

/** Number of cards the picker must bury in the Bury phase. */
export const SHEEPSHEAD_BURY_COUNT = 2;

/**
 * Hook that manages Sheepshead game state and the multi-phase player actions
 * (pick/pass, bury, call, play) plus trick/round advancement.
 *
 * Unlike the simple trick-taking games that reuse `useTrickGameBase`,
 * Sheepshead has a Pick → Bury → Call → Play flow, so the command set is built
 * directly on {@link useGameApi}.
 */
export function useSheepsheadGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<SheepsheadConfigInput>>(DEFAULT_SHEEPSHEAD_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(sheepsheadApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Picks (takes) the blind in the Pick phase. */
  const handlePick = useCallback(() => {
    void exec('pick', { pick: true });
  }, [exec]);

  /** Passes on the blind in the Pick phase. */
  const handlePass = useCallback(() => {
    void exec('pick', { pick: false });
  }, [exec]);

  /** Buries the two currently-selected cards (picker, Bury phase). */
  const handleBury = useCallback(() => {
    if (selectedCardIndices.length !== SHEEPSHEAD_BURY_COUNT) return;
    void exec('bury', { buryIndices: selectedCardIndices });
  }, [exec, selectedCardIndices]);

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
    sheepsheadConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handlePick,
    handlePass,
    handleBury,
    handleCall,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  };
}
