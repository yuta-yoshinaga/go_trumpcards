import { useCallback } from 'react';
import { type MinchiateConfigInput, minchiateApi } from '../api/gameApi';
import { MINCHIATE_SURPLUS } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Minchiate game configuration. */
export const DEFAULT_MINCHIATE_CONFIG: Required<MinchiateConfigInput> = {
  cpuDifficulty: 1,
  targetRounds: 4,
};

/** CPU difficulty level options for Minchiate. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/**
 * Round-count options for Minchiate.
 *
 * **Multiples of the player count only.** The deal rotates each round, so a
 * non-multiple ends the match with someone having taken the scarto more often
 * than the rest — the backend rejects it outright.
 */
export const TARGET_ROUNDS_OPTIONS = [4, 8, 12] as const;

/**
 * Hook that manages Minchiate game state: the dealer's scarto, the play
 * action, and trick/round advancement.
 *
 * Minchiate is a 4-player Florentine trick-taker in fixed 2v2 teams on a
 * 97-card tarot deck with 40 trumps. There is no bidding — the dealer buries
 * 13 surplus cards and play begins.
 */
export function useMinchiateGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<MinchiateConfigInput>>(DEFAULT_MINCHIATE_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(minchiateApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Buries the selected surplus cards. Requires exactly MINCHIATE_SURPLUS. */
  const handleScarto = useCallback(() => {
    if (selectedCardIndices.length !== MINCHIATE_SURPLUS) return;
    void exec('scarto', { cardIndices: selectedCardIndices });
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

  /**
   * Asks the server for a play recommendation.
   *
   * CUI と CLI からは呼べるのに、盤面には要求する手段が無かった。表示側の
   * `isRequestedHint` は `hintRequested` の messageCode を待つので、この
   * コマンドを送らないと永遠に出ない — 実質デッドコードだった (#4819)。
   */
  const handleHint = useCallback(() => {
    void exec('hint');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    minchiateConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handleScarto,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    handleHint,
  };
}
