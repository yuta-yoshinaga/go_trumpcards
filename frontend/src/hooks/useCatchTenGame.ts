import { catchtenApi } from '../api/gameApi';
import type { CatchTenConfig } from '../types/card';
import { useTrickGameBase } from './useTrickGameBase';

/** Default Catch the Ten game configuration. */
export const DEFAULT_CATCHTEN_CONFIG: CatchTenConfig = {
  cpuDifficulty: 1,
  pointLimit: 41,
};

/** CPU difficulty level options for Catch the Ten. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Catch the Ten. */
export const POINT_LIMIT_OPTIONS = [21, 31, 41, 51] as const;

/** Hook that manages Catch the Ten game state and player actions. */
export function useCatchTenGame() {
  const base = useTrickGameBase({
    apiFn: catchtenApi.exec,
    defaultConfig: DEFAULT_CATCHTEN_CONFIG,
    getHint: (state) => state.hint ?? null,
  });

  return {
    state: base.state,
    loading: base.loading,
    error: base.error,
    hint: base.hint,
    hintError: base.hintError,
    hintLoading: base.hintLoading,
    exec: base.exec,
    catchtenConfig: base.config,
    selectedCardIndices: base.selectedCardIndices,
    toggleCard: base.toggleCard,
    clearSelection: base.clearSelection,
    handleConfigChange: base.handleConfigChange,
    handlePlay: base.handlePlay,
    handleNextTrick: base.handleNextTrick,
    handleNextRound: base.handleNextRound,
    handleHint: base.handleHint,
    retry: base.retry,
  };
}
