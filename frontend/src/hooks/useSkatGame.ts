import { useCallback, useEffect, useState } from 'react';
import { skatApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import type { SkatConfig, SkatHint } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';
import { useIsMounted } from './useIsMounted';

/** Default Skat game configuration. */
export const DEFAULT_SKAT_CONFIG: SkatConfig = {
  cpuDifficulty: 1,
  targetScore: 500,
};

/** CPU difficulty options for Skat. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target-score options. */
export const TARGET_SCORE_OPTIONS = [100, 250, 500, 1000] as const;

/** Hook that manages Skat game state and player actions. */
export function useSkatGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: skatConfig, handleConfigChange } = useGameConfig<SkatConfig>(DEFAULT_SKAT_CONFIG);
  const [hint, setHint] = useState<SkatHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const [hintLoading, setHintLoading] = useState(false);

  const onSuccess = useCallback(() => {
    clearSelection();
    setHint(null);
  }, [clearSelection]);

  const { state, loading, error, exec: dispatch, retry } = useGameApi(skatApi.exec, { onSuccess });

  /** Resets the game, applying the current config (CPU difficulty, target score). */
  const reset = useCallback(() => {
    dispatch('reset', { config: skatConfig });
  }, [dispatch, skatConfig]);

  // Fetch a fresh game on mount using the initial (default) config.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run reset once on mount with the initial config.
  useEffect(() => {
    reset();
  }, []);

  const handleBid = useCallback(
    (accept: boolean) => {
      dispatch('bid', { accept });
    },
    [dispatch],
  );

  const handlePickSkat = useCallback(
    (pickup: boolean) => {
      dispatch('pickskat', { pickup });
    },
    [dispatch],
  );

  const handleDiscard = useCallback(() => {
    if (selectedCardIndices.length !== 2) return;
    dispatch('discard', { discardA: selectedCardIndices[0], discardB: selectedCardIndices[1] });
  }, [dispatch, selectedCardIndices]);

  const handleDeclareGame = useCallback(
    (gameType: number, trumpSuit?: number) => {
      dispatch('game', { gameType, trumpSuit });
    },
    [dispatch],
  );

  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    dispatch('play', { cardIndex: selectedCardIndices[0] });
  }, [dispatch, selectedCardIndices]);

  const handleNextTrick = useCallback(() => {
    dispatch('next');
  }, [dispatch]);

  const handleNextRound = useCallback(() => {
    dispatch('nextround');
  }, [dispatch]);

  const isMounted = useIsMounted();

  const handleHint = useCallback(async () => {
    setHintLoading(true);
    try {
      const res = await skatApi.exec('hint');
      // Navigating away mid-request must not write to a gone component (#4447).
      if (!isMounted()) return;
      setHint(res.hint ?? null);
      setHintError(null);
    } catch {
      if (!isMounted()) return;
      setHintError(NETWORK_ERROR_MESSAGE());
    } finally {
      if (isMounted()) setHintLoading(false);
    }
  }, [isMounted]);

  return {
    state,
    loading,
    error,
    hint,
    hintError,
    hintLoading,
    dispatch,
    skatConfig,
    reset,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleBid,
    handlePickSkat,
    handleDiscard,
    handleDeclareGame,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    handleHint,
    retry,
  };
}
