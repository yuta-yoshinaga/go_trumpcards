import { useCallback, useState } from 'react';
import { type CirullaConfigInput, cirullaApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Cirulla configuration. */
export const DEFAULT_CIRULLA_CONFIG: Required<CirullaConfigInput> = {
  cpuDifficulty: 1,
  targetScore: 51,
};

/** CPU difficulty options. */
export const CIRULLA_CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Target-score options. */
export const CIRULLA_TARGET_OPTIONS = [11, 21, 31, 51] as const;

/**
 * Hook that manages Cirulla state and its player actions: choosing a hand
 * card, choosing which table group to take, laying off, and moving on.
 *
 * **The capture groups come from the server.** Three capture rules interact
 * (same value, sum to the card, sum to fifteen) plus the ace sweep, so the page
 * picks from `captureOptions` rather than re-deriving them.
 */
export function useCirullaGame() {
  const [selectedHandIdx, setSelectedHandIdx] = useState<number | null>(null);
  const { config, handleConfigChange } = useGameConfig<Required<CirullaConfigInput>>(DEFAULT_CIRULLA_CONFIG);

  const onSuccess = useCallback(() => {
    setSelectedHandIdx(null);
  }, []);

  const { state, loading, error, exec, retry } = useGameApi(cirullaApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Selects (or deselects) a hand card. */
  const selectHand = useCallback((idx: number) => {
    setSelectedHandIdx((prev) => (prev === idx ? null : idx));
  }, []);

  /** Plays the selected card, taking the given table group (empty = lay off). */
  const play = useCallback(
    (captureIndices: number[]) => {
      if (selectedHandIdx === null) return;
      void exec('play', {
        handIndex: selectedHandIdx,
        captureIndices: captureIndices.length > 0 ? captureIndices : undefined,
      });
    },
    [exec, selectedHandIdx],
  );

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
    cirullaConfig: config,
    handleConfigChange,
    selectedHandIdx,
    selectHand,
    play,
    handleNextRound,
    reset,
  };
}
