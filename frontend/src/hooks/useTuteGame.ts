import { useCallback } from 'react';
import { type TuteConfigInput, tuteApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Tute game configuration. */
export const DEFAULT_TUTE_CONFIG: Required<TuteConfigInput> = {
  cpuDifficulty: 1,
  targetPoints: 121,
};

/** CPU difficulty level options for Tute. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target-point options for Tute (first team to reach wins). */
export const TARGET_POINTS_OPTIONS = [101, 121, 151, 201] as const;

/**
 * Hook that manages Tute game state and the player actions (play a card,
 * declare a King+Queen marriage, declare Tute) plus trick/round advancement.
 *
 * Tute is a Spanish 4-player (2 vs 2) trump trick-taker. The command set is
 * built directly on {@link useGameApi}. Beyond playing a card, the human may
 * declare a marriage for a suit they hold both K+Q in, or declare Tute (four
 * Kings or four Queens) for an instant win.
 */
export function useTuteGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<TuteConfigInput>>(DEFAULT_TUTE_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(tuteApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Plays the single currently-selected card in the Play phase. */
  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    void exec('play', { cardIndex: selectedCardIndices[0] });
  }, [exec, selectedCardIndices]);

  /** Declares a King+Queen marriage for the given suit (1=♠ 2=♣ 3=♥ 4=♦). */
  const handleDeclareMarriage = useCallback(
    (suit: number) => {
      void exec('marriage', { suit });
    },
    [exec],
  );

  /** Declares Tute (four Kings or four Queens) for an instant win. */
  const handleDeclareTute = useCallback(() => {
    void exec('tute');
  }, [exec]);

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
    tuteConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handlePlay,
    handleDeclareMarriage,
    handleDeclareTute,
    handleNextTrick,
    handleNextRound,
  };
}
