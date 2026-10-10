import { useCallback } from 'react';
import { wizardApi } from '../api/gameApi';
import type { WizardConfig } from '../types/card';
import { useTrickGameBase } from './useTrickGameBase';

/** Default Wizard game configuration. */
export const DEFAULT_WIZARD_CONFIG: WizardConfig = {
  cpuDifficulty: 1,
};

/** CPU difficulty level options for Wizard. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Hook that manages Wizard game state, bidding, and player actions. */
export function useWizardGame() {
  const {
    exec,
    config: wizardConfig,
    ...base
  } = useTrickGameBase({
    apiFn: wizardApi.exec,
    defaultConfig: DEFAULT_WIZARD_CONFIG,
    getHint: (state) => state.hint ?? null,
  });

  const handleBid = useCallback(
    (bid: number) => {
      exec('bid', bid);
    },
    [exec],
  );

  return {
    ...base,
    exec,
    wizardConfig,
    handleBid,
  };
}
