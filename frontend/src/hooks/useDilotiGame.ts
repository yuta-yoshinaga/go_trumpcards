import { useCallback, useState } from 'react';
import { type DilotiConfigInput, dilotiApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Diloti configuration. */
export const DEFAULT_DILOTI_CONFIG: Required<DilotiConfigInput> = {
  cpuDifficulty: 1,
  targetScore: 61,
};

/** CPU difficulty options. */
export const DILOTI_CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Target-score options. */
export const DILOTI_TARGET_OPTIONS = [21, 41, 61, 101] as const;

/**
 * Hook that manages Diloti state and its player actions: choosing a hand card,
 * then taking a group, declaring a value, or laying the card off.
 *
 * **The legal moves come from the server.** Captures sum to the played card's
 * own rank, face cards take exactly one match, and declarations can be raised
 * or grouped — so the page picks from `takeOptions` / `declareOptions` /
 * `canTrail` rather than re-deriving rules that interact.
 */
export function useDilotiGame() {
  const [selectedHandIdx, setSelectedHandIdx] = useState<number | null>(null);
  const { config, handleConfigChange } = useGameConfig<Required<DilotiConfigInput>>(DEFAULT_DILOTI_CONFIG);

  const onSuccess = useCallback(() => {
    setSelectedHandIdx(null);
  }, []);

  const { state, loading, error, exec, retry } = useGameApi(dilotiApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Selects (or deselects) a hand card. */
  const selectHand = useCallback((idx: number) => {
    setSelectedHandIdx((prev) => (prev === idx ? null : idx));
  }, []);

  /** Takes the given table cards and declarations with the selected card. */
  const take = useCallback(
    (tableIdxs: number[], declIdxs: number[]) => {
      if (selectedHandIdx === null) return;
      void exec('play', {
        handIndex: selectedHandIdx,
        action: 'capture',
        tableIndices: tableIdxs.length > 0 ? tableIdxs : undefined,
        declIndices: declIdxs.length > 0 ? declIdxs : undefined,
      });
    },
    [exec, selectedHandIdx],
  );

  /** Declares `value`, folding in the given loose table cards. */
  const declare = useCallback(
    (value: number, tableIdxs: number[]) => {
      if (selectedHandIdx === null) return;
      void exec('play', {
        handIndex: selectedHandIdx,
        action: 'declare',
        tableIndices: tableIdxs.length > 0 ? tableIdxs : undefined,
        declValue: value,
      });
    },
    [exec, selectedHandIdx],
  );

  /** Lays the selected card on the table. */
  const trail = useCallback(() => {
    if (selectedHandIdx === null) return;
    void exec('play', { handIndex: selectedHandIdx, action: 'trail' });
  }, [exec, selectedHandIdx]);

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
    dilotiConfig: config,
    handleConfigChange,
    selectedHandIdx,
    selectHand,
    take,
    declare,
    trail,
    handleNextRound,
    reset,
  };
}
