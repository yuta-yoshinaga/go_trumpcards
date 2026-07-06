import { useCallback, useState } from 'react';
import { type KoiKoiConfigInput, koikoiApi } from '../api/gameApi';
import type { KoiKoiResponse } from '../types/card';
import { useGameApi } from './useGameApi';

/** Default Koi-Koi (こいこい) game configuration. */
export const DEFAULT_KOIKOI_CONFIG: Required<KoiKoiConfigInput> = {
  cpuDifficulty: 1,
  targetScore: 50,
};

/** CPU difficulty level options for Koi-Koi. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Target-score options that end the match. */
export const TARGET_SCORE_OPTIONS = [30, 50, 100] as const;

/**
 * Hook that manages Koi-Koi (こいこい) game state, selections, and action dispatch.
 *
 * Koi-Koi is a 2-player hanafuda capture game. On the human's Play turn the
 * player selects a hand card; if that card matches two field cards the player
 * must additionally pick which field card to capture (`fieldIndex`). When a yaku
 * completes, the KoiKoiDecision phase offers koi-koi (continue) or shobu (stop).
 * Built directly on {@link useGameApi}.
 */
export function useKoiKoiGame() {
  const [handIndex, setHandIndex] = useState<number | null>(null);
  const [fieldIndex, setFieldIndex] = useState<number | null>(null);
  const [configInput, setConfigInput] = useState<Required<KoiKoiConfigInput>>(DEFAULT_KOIKOI_CONFIG);

  /** Clears the current hand-card and field-card selection. */
  const clearSelection = useCallback(() => {
    setHandIndex(null);
    setFieldIndex(null);
  }, []);

  const onSuccess = useCallback(
    async (_res: KoiKoiResponse) => {
      clearSelection();
    },
    [clearSelection],
  );

  const { loading, error, state, exec: callApi, retry } = useGameApi(koikoiApi.exec, { onSuccess });

  /** Updates a single config field. */
  const handleConfigChange = useCallback((key: keyof KoiKoiConfigInput, value: number) => {
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

  /** Calls koi-koi (continue the round for more yaku). */
  const callKoiKoi = useCallback(() => {
    callApi('koikoi');
  }, [callApi]);

  /** Calls shobu (stop and score the completed yaku). */
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
    callKoiKoi,
    callStop,
    handleNextRound,
    handleResetWithConfig,
  };
}
