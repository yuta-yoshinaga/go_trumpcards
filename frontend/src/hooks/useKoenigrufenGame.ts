import { useCallback } from 'react';
import { type KoenigrufenConfigInput, koenigrufenApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Number of cards the declarer buries in the talon exchange. */
export const KOENIGRUFEN_DISCARD_COUNT = 6;

/** Default Königrufen (ケーニッヒルーフェン) game configuration. */
export const DEFAULT_KOENIGRUFEN_CONFIG: Required<KoenigrufenConfigInput> = {
  cpuDifficulty: 1,
  targetDeals: 5,
};

/** CPU difficulty level options for Königrufen. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target deal-count options (match length; highest cumulative score wins). */
export const TARGET_DEALS_OPTIONS = [3, 5, 7] as const;

/**
 * Hook that manages Königrufen (ケーニッヒルーフェン) game state and its player
 * actions: declare the Rufer bid or pass, call a King's suit (naming the secret
 * partner), discard the 6-card talon, play a card, plus trick/round advancement.
 *
 * Königrufen is a 4-player tarock trick-taker on the 54-card tarock deck. The
 * human (seat 0) may or may not become the declarer depending on the auction.
 * The command set is built directly on {@link useGameApi}.
 */
export function useKoenigrufenGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<KoenigrufenConfigInput>>(DEFAULT_KOENIGRUFEN_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(koenigrufenApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Declares the Rufer contract in the Bid phase. */
  const handleBid = useCallback(() => {
    void exec('bid', { bid: 'rufer' });
  }, [exec]);

  /** Passes in the Bid phase. */
  const handlePass = useCallback(() => {
    void exec('pass');
  }, [exec]);

  /** Calls the King of the given suit (1-4) in the Call phase, naming the secret partner. */
  const handleCallKing = useCallback(
    (callSuit: number) => {
      void exec('callking', { callSuit });
    },
    [exec],
  );

  /** Discards the 6 currently-selected talon cards in the Talon phase. */
  const handleDiscard = useCallback(() => {
    if (selectedCardIndices.length !== KOENIGRUFEN_DISCARD_COUNT) return;
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
    koenigrufenConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handleBid,
    handlePass,
    handleCallKing,
    handleDiscard,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  };
}
