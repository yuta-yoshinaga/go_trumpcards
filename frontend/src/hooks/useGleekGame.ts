import { useCallback } from 'react';
import { type GleekConfigInput, gleekApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Gleek game configuration. */
export const DEFAULT_GLEEK_CONFIG: Required<GleekConfigInput> = {
  cpuDifficulty: 1,
  targetRounds: 5,
};

/** CPU difficulty level options for Gleek. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target deal-count options for Gleek (match length; highest cumulative score wins). */
export const TARGET_ROUNDS_OPTIONS = [3, 5, 7] as const;

/**
 * Hook that manages Gleek game state and its player actions: bidding for the
 * stock, throwing the buyer's discards, playing a card, and trick/deal
 * advancement.
 *
 * Gleek is a 3-player trick-taker on a 44-card pack with four scoring stages in
 * one deal. Only the auction, the discards and the card play take input — the
 * ruff and the melds are readings of the settled hands, scored by the server.
 * The command set is built directly on {@link useGameApi}.
 */
export function useGleekGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<GleekConfigInput>>(DEFAULT_GLEEK_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(gleekApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /**
   * Bids in the auction. `0` drops out; any other amount must be the server's
   * `nextBidAmount`, since raises go up in fixed steps.
   */
  const handleBid = useCallback(
    (bid: number) => {
      void exec('bid', { bid });
    },
    [exec],
  );

  /**
   * Throws the currently-selected cards as the buyer's discards.
   *
   * **Without this the board freezes right after the auction**: play is
   * rejected until the buyer is back down to a full hand.
   */
  const handleDiscard = useCallback(() => {
    if (selectedCardIndices.length === 0) return;
    void exec('discard', { discardIndices: [...selectedCardIndices] });
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

  /** Advances to the next deal. */
  const handleNextRound = useCallback(() => {
    void exec('nextround');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    gleekConfig: config,
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
