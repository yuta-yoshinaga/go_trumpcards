import { useCallback } from 'react';
import { type CegoConfigInput, cegoApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Number of cards the declarer keeps in the Cego exchange (the rest are laid down). */
export const CEGO_KEEP_COUNT = 1;

/** Default Cego (チェゴ) game configuration. */
export const DEFAULT_CEGO_CONFIG: Required<CegoConfigInput> = {
  cpuDifficulty: 1,
  targetDeals: 5,
};

/** CPU difficulty level options for Cego. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target deal-count options (match length; highest cumulative score wins). */
export const TARGET_DEALS_OPTIONS = [3, 5, 7] as const;

/**
 * Hook that manages Cego (チェゴ) game state and its player actions: declare the
 * play bid or pass, choose a contract (Cego / Handspiel), make the Cego exchange
 * (keep exactly 1 card), play a card, plus trick/round advancement.
 *
 * Cego is a 4-player Baden tarock trick-taker on the 54-card tarock deck. The
 * human (seat 0) may or may not become the declarer depending on the auction.
 * The command set is built directly on {@link useGameApi}.
 */
export function useCegoGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<CegoConfigInput>>(DEFAULT_CEGO_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(cegoApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Declares the play bid in the Bid phase. */
  const handleBid = useCallback(() => {
    void exec('bid', { bid: 'play' });
  }, [exec]);

  /** Passes in the Bid phase. */
  const handlePass = useCallback(() => {
    void exec('pass');
  }, [exec]);

  /** Chooses the contract ('cego' or 'handspiel') in the Contract phase. */
  const handleContract = useCallback(
    (contract: 'cego' | 'handspiel') => {
      void exec('contract', { contract });
    },
    [exec],
  );

  /** Makes the Cego exchange, keeping the single currently-selected card. */
  const handleExchange = useCallback(() => {
    if (selectedCardIndices.length !== CEGO_KEEP_COUNT) return;
    void exec('discard', { cardIndices: [selectedCardIndices[0]] });
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
    cegoConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handleBid,
    handlePass,
    handleContract,
    handleExchange,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  };
}
