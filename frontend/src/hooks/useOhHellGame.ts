import { useCallback, useEffect, useState } from 'react';
import { ohHellApi } from '../api/gameApi';
import type { OhHellConfig, OhHellHint } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';
import { useHintRequest } from './useHintRequest';

/** Default Oh Hell game configuration. */
export const DEFAULT_OH_HELL_CONFIG: OhHellConfig = {
  cpuDifficulty: 1,
  maxHandSize: 10,
  scoringVariant: 0,
  roundDirection: 1,
};

/** CPU difficulty level options for Oh Hell. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available max hand size options for Oh Hell. */
export const MAX_HAND_SIZE_OPTIONS = [3, 5, 7, 10, 13] as const;

/** Scoring variant options for Oh Hell. */
export const SCORING_VARIANT_OPTIONS = [
  { value: 0, labelKey: 'scoringStandard' },
  { value: 1, labelKey: 'scoringPenalty' },
] as const;

/** Round direction options for Oh Hell. */
export const ROUND_DIRECTION_OPTIONS = [
  { value: 0, labelKey: 'roundDownOnly' },
  { value: 1, labelKey: 'roundDownAndUp' },
] as const;

/** Hook that manages Oh Hell game state, bidding, and player actions. */
export function useOhHellGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: ohHellConfig, handleConfigChange } = useGameConfig<OhHellConfig>(DEFAULT_OH_HELL_CONFIG);
  const [hint, setHint] = useState<OhHellHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const [hintLoading, setHintLoading] = useState(false);

  const onSuccess = useCallback(() => {
    clearSelection();
    setHint(null);
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(ohHellApi.exec, { onSuccess });

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    exec('reset', undefined, undefined, DEFAULT_OH_HELL_CONFIG);
  }, [exec]);

  const handleBid = useCallback(
    (bid: number) => {
      exec('bid', bid);
    },
    [exec],
  );

  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    exec('play', undefined, selectedCardIndices[0]);
  }, [exec, selectedCardIndices]);

  const handleNextTrick = useCallback(() => {
    exec('next');
  }, [exec]);

  const handleNextRound = useCallback(() => {
    exec('nextround');
  }, [exec]);

  const handleHint = useHintRequest({
    fetchHint: () => ohHellApi.exec('hint'),
    selectHint: (res) => res.hint,
    setHint,
    setHintError,
    setHintLoading,
  });

  return {
    state,
    loading,
    error,
    hint,
    hintError,
    hintLoading,
    exec,
    ohHellConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleBid,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    handleHint,
    retry,
  };
}
