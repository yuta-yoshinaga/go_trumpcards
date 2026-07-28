import { useCallback, useState } from 'react';
import { type DoppelkopfConfigInput, doppelkopfApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';
import { useIsMounted } from './useIsMounted';

/** Default Doppelkopf game configuration. */
export const DEFAULT_DOPPELKOPF_CONFIG: Required<DoppelkopfConfigInput> = {
  cpuDifficulty: 1,
  baseChips: 2,
  startChips: 20,
  targetChips: 40,
};

/** CPU difficulty level options for Doppelkopf. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target-chip options for Doppelkopf. */
export const TARGET_CHIPS_OPTIONS = [40, 60, 100, 200] as const;

/**
 * Hook that manages Doppelkopf game state and the player actions
 * (play, announce Re/Kontra) plus trick/round advancement.
 *
 * Doppelkopf is a plain trick-taking flow (no pick/bury/call), so the command
 * set is built directly on {@link useGameApi}. The only action beyond playing a
 * card is `announce`, available during the first trick.
 */
export function useDoppelkopfGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<DoppelkopfConfigInput>>(DEFAULT_DOPPELKOPF_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(doppelkopfApi.exec, { onSuccess });

  const [hintLoading, setHintLoading] = useState(false);

  const isMounted = useIsMounted();

  /** Requests a play hint from the server; guards against double-clicks. */
  const handleHint = useCallback(async () => {
    setHintLoading(true);
    try {
      await exec('hint');
    } finally {
      // The hint request outliving the page must not write to it (#4447).
      if (isMounted()) setHintLoading(false);
    }
  }, [exec, isMounted]);

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Plays the single currently-selected card in the Play phase. */
  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    void exec('play', { cardIndex: selectedCardIndices[0] });
  }, [exec, selectedCardIndices]);

  /** Announces Re or Kontra (first trick only). */
  const handleAnnounce = useCallback(() => {
    void exec('announce');
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
    doppelkopfConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handlePlay,
    handleAnnounce,
    handleNextTrick,
    handleNextRound,
    handleHint,
    hintLoading,
  };
}
