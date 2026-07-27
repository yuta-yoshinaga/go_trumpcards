import { useCallback, useEffect, useState } from 'react';
import { euchreApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import type { EuchreConfig, EuchreHint } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';
import { useIsMounted } from './useIsMounted';

/** Default Euchre game configuration. */
export const DEFAULT_EUCHRE_CONFIG: EuchreConfig = {
  cpuDifficulty: 1,
  pointLimit: 10,
};

/** CPU difficulty level options for Euchre. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Euchre. */
export const POINT_LIMIT_OPTIONS = [5, 7, 10, 15, 21] as const;

/** Hook that manages Euchre game state, bidding, and player actions. */
export function useEuchreGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: euchreConfig, handleConfigChange } = useGameConfig<EuchreConfig>(DEFAULT_EUCHRE_CONFIG);
  const [hint, setHint] = useState<EuchreHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const [hintLoading, setHintLoading] = useState(false);

  const onSuccess = useCallback(() => {
    clearSelection();
    setHint(null);
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(euchreApi.exec, { onSuccess });

  const apiExec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    apiExec('reset', undefined, undefined, undefined, DEFAULT_EUCHRE_CONFIG);
  }, [apiExec]);

  const handleOrderUp = useCallback(
    (goAlone: boolean) => {
      apiExec('orderup', undefined, undefined, goAlone);
    },
    [apiExec],
  );

  const handlePass = useCallback(() => {
    apiExec('pass');
  }, [apiExec]);

  const handleCallTrump = useCallback(
    (suit: number, goAlone: boolean) => {
      apiExec('calltrump', undefined, suit, goAlone);
    },
    [apiExec],
  );

  const handleDiscard = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    apiExec('discard', selectedCardIndices[0]);
  }, [apiExec, selectedCardIndices]);

  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    apiExec('play', selectedCardIndices[0]);
  }, [apiExec, selectedCardIndices]);

  const handleNextTrick = useCallback(() => {
    apiExec('next');
  }, [apiExec]);

  const handleNextRound = useCallback(() => {
    apiExec('nextround');
  }, [apiExec]);

  const isMounted = useIsMounted();

  const handleHint = useCallback(async () => {
    setHintLoading(true);
    try {
      const res = await euchreApi.exec('hint');
      // Navigating away mid-request must not write to a gone component (#4447).
      if (!isMounted()) return;
      setHint(res.hint ?? null);
      setHintError(null);
    } catch {
      if (!isMounted()) return;
      setHintError(NETWORK_ERROR_MESSAGE());
    } finally {
      if (isMounted()) setHintLoading(false);
    }
  }, [isMounted]);

  return {
    state,
    loading,
    error,
    hint,
    hintError,
    hintLoading,
    apiExec,
    euchreConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleOrderUp,
    handlePass,
    handleCallTrump,
    handleDiscard,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    handleHint,
    retry,
  };
}
