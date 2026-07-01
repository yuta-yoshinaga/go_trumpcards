import { useCallback } from 'react';
import { type CalabresellaConfigInput, calabresellaApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Calabresella (Terziglio) game configuration. */
export const DEFAULT_CALABRESELLA_CONFIG: Required<CalabresellaConfigInput> = {
  cpuDifficulty: 1,
  targetPoints: 21,
};

/** CPU difficulty level options for Calabresella. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target match-point options for Calabresella (first player to reach wins). */
export const TARGET_POINTS_OPTIONS = [11, 21, 31] as const;

/**
 * Hook that manages Calabresella (Terziglio) game state and its player actions:
 * bid (pass/chiamo/solo), monte discard, play a card, plus trick/round
 * advancement.
 *
 * Calabresella is a Calabrian/Italian 3-player 40-card (Tressette-family)
 * trick-taker. A Bid phase decides the Soloist, who then takes the 4-card monte
 * and discards four cards down to twelve before playing alone against the
 * coalition of the other two players. There is no trump. The command set is
 * built directly on {@link useGameApi}.
 */
export function useCalabresellaGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<CalabresellaConfigInput>>(DEFAULT_CALABRESELLA_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(calabresellaApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Declares a bid in the Bid phase (0=pass, 1=chiamo, 2=solo). */
  const handleBid = useCallback(
    (bid: number) => {
      void exec('bid', { bid });
    },
    [exec],
  );

  /** Discards the currently-selected card during the monte exchange (four times as Soloist). */
  const handleDiscard = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    void exec('discard', { cardIndex: selectedCardIndices[0] });
  }, [exec, selectedCardIndices]);

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
    calabresellaConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handleBid,
    handleDiscard,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  };
}
