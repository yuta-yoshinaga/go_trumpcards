import { useCallback } from 'react';
import { type TysiacConfigInput, tysiacApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Tysiąc (Thousand) game configuration. */
export const DEFAULT_TYSIAC_CONFIG: Required<TysiacConfigInput> = {
  cpuDifficulty: 1,
  targetPoints: 1000,
};

/** CPU difficulty level options for Tysiąc. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target match-point options for Tysiąc (first player to reach wins). */
export const TARGET_POINTS_OPTIONS = [500, 1000, 1500] as const;

/**
 * Hook that manages Tysiąc (Thousand) game state and its player actions: bid
 * raise/pass, talon discard, play a card, plus trick/round advancement.
 *
 * Tysiąc is a Polish 3-player 24-card trump trick-taker. A Bid phase decides
 * the Declarer, who then exchanges via the Talon (discarding one card to each
 * opponent) and plays the contract. Trump is set dynamically by declaring a
 * marriage (K+Q) while leading. The command set is built directly on
 * {@link useGameApi}.
 */
export function useTysiacGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<TysiacConfigInput>>(DEFAULT_TYSIAC_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(tysiacApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Raises (+10) or passes in the Bid phase. */
  const handleBid = useCallback(
    (raise: boolean) => {
      void exec('bid', { raise });
    },
    [exec],
  );

  /** Discards the currently-selected card to an opponent during the Talon exchange. */
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
    tysiacConfig: config,
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
