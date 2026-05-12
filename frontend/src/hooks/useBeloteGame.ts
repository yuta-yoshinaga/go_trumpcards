import { useCallback } from 'react';
import { beloteApi } from '../api/gameApi';
import type { BeloteConfig } from '../types/card';
import { useTrickGameBase } from './useTrickGameBase';

/** Default Belote game configuration. */
export const DEFAULT_BELOTE_CONFIG: BeloteConfig = {
  cpuDifficulty: 1,
  targetScore: 1000,
  dixDeDer: 10,
  enableBeloteRebelote: true,
};

/** CPU difficulty level options for Belote. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target-score options for Belote. */
export const TARGET_SCORE_OPTIONS = [500, 750, 1000, 1500] as const;

/** Hook that manages Belote game state and player actions. */
export function useBeloteGame() {
  const base = useTrickGameBase({
    apiFn: beloteApi.exec,
    defaultConfig: DEFAULT_BELOTE_CONFIG,
    getHint: (state) => state.hint ?? null,
  });

  const handleOrderUp = useCallback(() => {
    void (base.exec as unknown as (command: string) => Promise<void>)('orderup');
  }, [base.exec]);

  const handlePass = useCallback(() => {
    void (base.exec as unknown as (command: string) => Promise<void>)('pass');
  }, [base.exec]);

  const handleCallTrump = useCallback(
    (suit: number) => {
      void (base.exec as unknown as (command: string, cardIndex?: number, s?: number) => Promise<void>)(
        'calltrump',
        undefined,
        suit,
      );
    },
    [base.exec],
  );

  return {
    state: base.state,
    loading: base.loading,
    error: base.error,
    hint: base.hint,
    hintError: base.hintError,
    hintLoading: base.hintLoading,
    exec: base.exec,
    beloteConfig: base.config,
    selectedCardIndices: base.selectedCardIndices,
    toggleCard: base.toggleCard,
    handleConfigChange: base.handleConfigChange,
    handleToggle: base.handleToggle,
    handlePlay: base.handlePlay,
    handleNextTrick: base.handleNextTrick,
    handleNextRound: base.handleNextRound,
    handleHint: base.handleHint,
    handleOrderUp,
    handlePass,
    handleCallTrump,
    retry: base.retry,
  };
}
