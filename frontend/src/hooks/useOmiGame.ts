import { useCallback } from 'react';
import { omiApi } from '../api/gameApi';
import type { OmiConfig } from '../types/card';
import { useTrickGameBase } from './useTrickGameBase';

/** Default Omi game configuration. */
export const DEFAULT_OMI_CONFIG: OmiConfig = {
  cpuDifficulty: 1,
  pointLimit: 10,
};

/** CPU difficulty level options for Omi. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Omi. */
export const POINT_LIMIT_OPTIONS = [5, 7, 10, 15, 21] as const;

/** Hook that manages Omi game state, trump calling, and trick play. */
export function useOmiGame() {
  const {
    exec: apiExec,
    config: omiConfig,
    selectedCardIndices,
    ...base
  } = useTrickGameBase({
    apiFn: omiApi.exec,
    defaultConfig: DEFAULT_OMI_CONFIG,
    getHint: (state) => state.hint ?? null,
  });

  const handleCallTrump = useCallback(
    (suit: number) => {
      apiExec('calltrump', undefined, suit);
    },
    [apiExec],
  );

  // Not the base hook's handlePlay: Omi's API takes the card index in slot 2,
  // where useTrickGameBase sends it in slot 3.
  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    apiExec('play', selectedCardIndices[0]);
  }, [apiExec, selectedCardIndices]);

  return {
    ...base,
    apiExec,
    omiConfig,
    selectedCardIndices,
    handlePlay,
    handleCallTrump,
  };
}
