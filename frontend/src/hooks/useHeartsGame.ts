import { useCallback } from 'react';
import { heartsApi } from '../api/gameApi';
import { useTrickGameBase } from './useTrickGameBase';

/** Default Hearts game configuration. */
export const DEFAULT_HEARTS_CONFIG: HeartsConfig = {
  cpuDifficulty: 1,
  pointLimit: 100,
  omnibusJD: false,
};

/** CPU difficulty level options for Hearts. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Hearts. */
export const POINT_LIMIT_OPTIONS = [50, 100, 150, 200] as const;

/** Hook that manages Hearts game state and player actions. */
export function useHeartsGame() {
  const base = useTrickGameBase({
    apiFn: heartsApi.exec,
    defaultConfig: DEFAULT_HEARTS_CONFIG,
    getHint: (state) => state.hint ?? null,
  });

  const { exec, selectedCardIndices } = base;

  const handlePass = useCallback(() => {
    exec('pass', selectedCardIndices);
  }, [exec, selectedCardIndices]);

  return {
    state: base.state,
    loading: base.loading,
    error: base.error,
    hint: base.hint,
    hintError: base.hintError,
    hintLoading: base.hintLoading,
    exec: base.exec,
    heartsConfig: base.config,
    selectedCardIndices: base.selectedCardIndices,
    toggleCard: base.toggleCard,
    clearSelection: base.clearSelection,
    handleConfigChange: base.handleConfigChange,
    handleToggle: base.handleToggle,
    handlePass,
    handlePlay: base.handlePlay,
    handleNextTrick: base.handleNextTrick,
    handleNextRound: base.handleNextRound,
    handleHint: base.handleHint,
  };
}
