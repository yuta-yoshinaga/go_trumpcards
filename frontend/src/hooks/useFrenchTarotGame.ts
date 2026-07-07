import { useCallback } from 'react';
import { type FrenchTarotConfigInput, frenchtarotApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Number of cards the declarer buries in the écart (chien exchange). */
export const FRENCH_TAROT_ECART_COUNT = 6;

/** Default French Tarot (フレンチタロット) game configuration. */
export const DEFAULT_FRENCH_TAROT_CONFIG: Required<FrenchTarotConfigInput> = {
  cpuDifficulty: 1,
  targetDeals: 5,
};

/** CPU difficulty level options for French Tarot. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target deal-count options (match length; highest cumulative score wins). */
export const TARGET_DEALS_OPTIONS = [3, 5, 7] as const;

/** Contract bid strings accepted by the backend `bid` command (Pass uses the `pass` command). */
export type FrenchTarotContractName = 'petite' | 'garde' | 'gardesans' | 'gardecontre';

/**
 * Hook that manages French Tarot (フレンチタロット) game state and its player
 * actions: declare a bid (Petite / Garde / Garde Sans / Garde Contre) or pass,
 * discard the 6-card écart, play a card, plus trick/round advancement.
 *
 * French Tarot is a 4-player trick-taker on the 78-card tarot deck. The human
 * (seat 0) may or may not become the declarer depending on the auction. The
 * command set is built directly on {@link useGameApi}.
 */
export function useFrenchTarotGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<FrenchTarotConfigInput>>(DEFAULT_FRENCH_TAROT_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(frenchtarotApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Declares a contract in the Bid phase. */
  const handleBid = useCallback(
    (bid: FrenchTarotContractName) => {
      void exec('bid', { bid });
    },
    [exec],
  );

  /** Passes in the Bid phase. */
  const handlePass = useCallback(() => {
    void exec('pass');
  }, [exec]);

  /** Discards the 6 currently-selected écart cards in the Chien phase. */
  const handleDiscard = useCallback(() => {
    if (selectedCardIndices.length !== FRENCH_TAROT_ECART_COUNT) return;
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
    frenchtarotConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handleBid,
    handlePass,
    handleDiscard,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  };
}
