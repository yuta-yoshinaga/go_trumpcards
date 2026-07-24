import { useCallback } from 'react';
import { type UltiConfigInput, ultiApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Ulti (Ultimo) game configuration. */
export const DEFAULT_ULTI_CONFIG: Required<UltiConfigInput> = {
  cpuDifficulty: 1,
  targetRounds: 5,
};

/** CPU difficulty level options for Ulti. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target deal-count options for Ulti (match length; highest coin balance wins). */
export const TARGET_ROUNDS_OPTIONS = [3, 5, 7] as const;

/** Contract identifiers accepted by the backend bid command. */
export type UltiContractName = 'party' | 'betli' | 'durchmarsch' | 'ulti';

/**
 * Hook that manages Ulti (Ultimo) game state and its player actions: declare a
 * contract (Party with a trump suit / Betli / Durchmarsch), discard 2 talon
 * cards, play a card, plus trick/round advancement.
 *
 * Ulti is a 3-player Hungarian contract trick-taker on a 32-card deck. The human
 * (seat 0) is always the declarer versus a 2-CPU defending coalition. The
 * command set is built directly on {@link useGameApi}.
 */
export function useUltiGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<UltiConfigInput>>(DEFAULT_ULTI_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(ultiApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /**
   * Declares a contract in the Bid phase. A Party or Ulti contract carries the
   * chosen trump suit (1=♠ 2=♣ 3=♥ 4=♦); Betli/Durchmarsch ignore the trump suit.
   */
  const handleBid = useCallback(
    (contract: UltiContractName, trumpSuit?: number) => {
      void exec('bid', { contract, trumpSuit });
    },
    [exec],
  );

  /** Discards the 2 currently-selected talon cards in the Discard phase. */
  const handleDiscard = useCallback(() => {
    if (selectedCardIndices.length !== 2) return;
    void exec('discard', { cardIndices: [...selectedCardIndices] });
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
    ultiConfig: config,
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
