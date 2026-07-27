import { useCallback, useEffect, useState } from 'react';
import { napoleonApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import type { NapoleonConfig, NapoleonHint } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';
import { useIsMounted } from './useIsMounted';

/** Default Napoleon game configuration. */
export const DEFAULT_NAPOLEON_CONFIG: NapoleonConfig = {
  cpuDifficulty: 1,
  minBid: 12,
  pointLimit: 100,
};

/** CPU difficulty level options for Napoleon. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Napoleon. */
export const POINT_LIMIT_OPTIONS = [50, 100, 200, 300, 500] as const;

/** Available minimum bid options for Napoleon. */
export const MIN_BID_OPTIONS = [12, 13, 14, 15] as const;

/** Hook that manages Napoleon game state, bidding, declarations, and player actions. */
export function useNapoleonGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: napoleonConfig, handleConfigChange } = useGameConfig<NapoleonConfig>(DEFAULT_NAPOLEON_CONFIG);
  const [hint, setHint] = useState<NapoleonHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const [hintLoading, setHintLoading] = useState(false);

  const onSuccess = useCallback(() => {
    clearSelection();
    setHint(null);
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(napoleonApi.exec, { onSuccess });

  const apiExec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    apiExec('reset', undefined, undefined, undefined, undefined, undefined, undefined, DEFAULT_NAPOLEON_CONFIG);
  }, [apiExec]);

  const handleBid = useCallback(
    (bid: number) => {
      apiExec('bid', bid);
    },
    [apiExec],
  );

  const handlePass = useCallback(() => {
    apiExec('bid', 0);
  }, [apiExec]);

  const handleTrumpDeclaration = useCallback(
    (trumpSuit: number, adjutantSuit: number, adjutantValue: number) => {
      apiExec('trump', undefined, trumpSuit, adjutantSuit, adjutantValue);
    },
    [apiExec],
  );

  const handleExchange = useCallback(
    (discardIndex: number) => {
      apiExec('exchange', undefined, undefined, undefined, undefined, discardIndex);
    },
    [apiExec],
  );

  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    apiExec('play', undefined, undefined, undefined, undefined, undefined, selectedCardIndices[0]);
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
      const res = await napoleonApi.exec('hint');
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
    napoleonConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleBid,
    handlePass,
    handleTrumpDeclaration,
    handleExchange,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    handleHint,
    retry,
  };
}
