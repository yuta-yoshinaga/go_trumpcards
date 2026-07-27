import { useCallback, useEffect, useState } from 'react';
import { wizardApi } from '../api/gameApi';
import type { WizardConfig, WizardHint } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';
import { useHintRequest } from './useHintRequest';

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
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: wizardConfig, handleConfigChange } = useGameConfig<WizardConfig>(DEFAULT_WIZARD_CONFIG);
  const [hint, setHint] = useState<WizardHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const [hintLoading, setHintLoading] = useState(false);

  const onSuccess = useCallback(() => {
    clearSelection();
    setHint(null);
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(wizardApi.exec, { onSuccess });

  const exec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    exec('reset', undefined, undefined, DEFAULT_WIZARD_CONFIG);
  }, [exec]);

  const handleBid = useCallback(
    (bid: number) => {
      exec('bid', bid);
    },
    [exec],
  );

  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    exec('play', undefined, selectedCardIndices[0]);
  }, [exec, selectedCardIndices]);

  const handleNextTrick = useCallback(() => {
    exec('next');
  }, [exec]);

  const handleNextRound = useCallback(() => {
    exec('nextround');
  }, [exec]);

  const handleHint = useHintRequest({
    fetchHint: () => wizardApi.exec('hint'),
    selectHint: (res) => res.hint,
    setHint,
    setHintError,
    setHintLoading,
  });

  return {
    state,
    loading,
    error,
    hint,
    hintError,
    hintLoading,
    exec,
    wizardConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleBid,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    handleHint,
    retry,
  };
}
