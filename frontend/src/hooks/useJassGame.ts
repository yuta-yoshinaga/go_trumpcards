import { useCallback } from 'react';
import { jassApi } from '../api/gameApi';
import type { JassConfig } from '../types/card';
import { useTrickGameBase } from './useTrickGameBase';

/** Default Jass (Schieber) game configuration. */
export const DEFAULT_JASS_CONFIG: JassConfig = {
  cpuDifficulty: 1,
  targetScore: 1000,
  lastTrickBonus: 5,
  enableWeis: true,
};

/** CPU difficulty level options for Jass. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target-score options for Jass. */
export const TARGET_SCORE_OPTIONS = [500, 1000, 1500, 2500] as const;

/** Hook that manages Jass (Schieber) game state and player actions. */
export function useJassGame() {
  const base = useTrickGameBase({
    apiFn: jassApi.exec,
    defaultConfig: DEFAULT_JASS_CONFIG,
    getHint: (state) => state.hint ?? null,
  });

  const handleCallTrump = useCallback(
    (suit: number) => {
      void (base.exec as unknown as (command: string, s?: number) => Promise<void>)('calltrump', suit);
    },
    [base.exec],
  );

  const handleSchieben = useCallback(() => {
    void (base.exec as unknown as (command: string) => Promise<void>)('schieben');
  }, [base.exec]);

  return {
    state: base.state,
    loading: base.loading,
    error: base.error,
    hint: base.hint,
    hintError: base.hintError,
    hintLoading: base.hintLoading,
    exec: base.exec,
    jassConfig: base.config,
    selectedCardIndices: base.selectedCardIndices,
    toggleCard: base.toggleCard,
    handleConfigChange: base.handleConfigChange,
    handleToggle: base.handleToggle,
    handlePlay: base.handlePlay,
    handleNextTrick: base.handleNextTrick,
    handleNextRound: base.handleNextRound,
    handleHint: base.handleHint,
    handleCallTrump,
    handleSchieben,
    retry: base.retry,
  };
}
