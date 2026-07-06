import { useCallback, useState } from 'react';
import { type GoStopConfigInput, gostopApi } from '../api/gameApi';
import type { GoStopResponse } from '../types/card';
import { useGameApi } from './useGameApi';

/** Default Go-Stop (ゴーストップ) game configuration. */
export const DEFAULT_GOSTOP_CONFIG: Required<GoStopConfigInput> = {
  cpuDifficulty: 1,
  targetScore: 7,
};

/** CPU difficulty level options for Go-Stop. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Target-score options that end the match. */
export const TARGET_SCORE_OPTIONS = [3, 5, 7, 10] as const;

/**
 * Hook that manages Go-Stop (ゴーストップ) game state, selections, and action
 * dispatch.
 *
 * Go-Stop is the Korean sibling of Koi-Koi: a 2-player hanafuda capture game.
 * On the human's Play turn the player selects a hand card; if that card matches
 * two field cards the player must additionally pick which field card to capture
 * (`fieldIndex`). When the target score is reached the GoDecision phase offers
 * go (continue) or stop (bank). Built directly on {@link useGameApi}.
 */
export function useGoStopGame() {
  const [handIndex, setHandIndex] = useState<number | null>(null);
  const [fieldIndex, setFieldIndex] = useState<number | null>(null);
  const [configInput, setConfigInput] = useState<Required<GoStopConfigInput>>(DEFAULT_GOSTOP_CONFIG);

  /** Clears the current hand-card and field-card selection. */
  const clearSelection = useCallback(() => {
    setHandIndex(null);
    setFieldIndex(null);
  }, []);

  const onSuccess = useCallback(
    async (_res: GoStopResponse) => {
      clearSelection();
    },
    [clearSelection],
  );

  const { loading, error, state, exec: callApi, retry } = useGameApi(gostopApi.exec, { onSuccess });

  /** Updates a single config field. */
  const handleConfigChange = useCallback((key: keyof GoStopConfigInput, value: number) => {
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

  /** Calls go (continue the round for more points). */
  const callGo = useCallback(() => {
    callApi('go');
  }, [callApi]);

  /** Calls stop (bank the current score and end the round). */
  const callStop = useCallback(() => {
    callApi('stop');
  }, [callApi]);

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
    callGo,
    callStop,
    handleNextRound,
    handleResetWithConfig,
  };
}
