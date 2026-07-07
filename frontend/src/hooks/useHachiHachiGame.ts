import { useCallback, useState } from 'react';
import { type HachiHachiConfigInput, hachihachiApi } from '../api/gameApi';
import type { HachiHachiResponse } from '../types/card';
import { useGameApi } from './useGameApi';

/** Default Hachi-Hachi (八八) game configuration. */
export const DEFAULT_HACHIHACHI_CONFIG: Required<HachiHachiConfigInput> = {
  cpuDifficulty: 1,
  targetRounds: 3,
};

/** CPU difficulty level options for Hachi-Hachi. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Target-round options that end the match. */
export const TARGET_ROUNDS_OPTIONS = [3, 6, 12] as const;

/**
 * Hook that manages Hachi-Hachi (八八) game state, selections, and action dispatch.
 *
 * Hachi-Hachi is a 3-player hanafuda capture game. On the human's Play turn the
 * player selects a hand card; if that card matches two field cards the player
 * must additionally pick which field card to capture (`fieldIndex`). Unlike
 * Koi-Koi there is no koi-koi/stop decision — the round simply ends when every
 * hand is exhausted and captured piles are scored against the 88 baseline.
 * Built directly on {@link useGameApi}.
 */
export function useHachiHachiGame() {
  const [handIndex, setHandIndex] = useState<number | null>(null);
  const [fieldIndex, setFieldIndex] = useState<number | null>(null);
  const [configInput, setConfigInput] = useState<Required<HachiHachiConfigInput>>(DEFAULT_HACHIHACHI_CONFIG);

  /** Clears the current hand-card and field-card selection. */
  const clearSelection = useCallback(() => {
    setHandIndex(null);
    setFieldIndex(null);
  }, []);

  const onSuccess = useCallback(
    async (_res: HachiHachiResponse) => {
      clearSelection();
    },
    [clearSelection],
  );

  const { loading, error, state, exec: callApi, retry } = useGameApi(hachihachiApi.exec, { onSuccess });

  /** Updates a single config field. */
  const handleConfigChange = useCallback((key: keyof HachiHachiConfigInput, value: number) => {
    setConfigInput((prev) => ({ ...prev, [key]: value }));
  }, []);

  /** Selects a hand card, clearing any prior field selection. */
  const selectHand = useCallback((idx: number) => {
    setHandIndex((prev) => (prev === idx ? null : idx));
    setFieldIndex(null);
  }, []);

  /** Plays a hand card, optionally targeting a specific field card for a 2-way match. */
  const playCard = useCallback(
    (cardIndex: number, targetField?: number) => {
      callApi('play', targetField === undefined ? { cardIndex } : { cardIndex, fieldIndex: targetField });
    },
    [callApi],
  );

  /** Advances to the next round. */
  const handleNextRound = useCallback(() => {
    callApi('nextround');
  }, [callApi]);

  /** Resets the game, applying the current config. */
  const handleResetWithConfig = useCallback(() => {
    callApi('reset', { config: configInput });
  }, [callApi, configInput]);

  return {
    state,
    loading,
    error,
    retry,
    callApi,
    handIndex,
    setHandIndex,
    fieldIndex,
    setFieldIndex,
    selectHand,
    clearSelection,
    configInput,
    handleConfigChange,
    playCard,
    handleNextRound,
    handleResetWithConfig,
  };
}
