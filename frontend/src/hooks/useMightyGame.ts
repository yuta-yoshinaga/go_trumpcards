import { useCallback, useEffect, useState } from 'react';
import { mightyApi } from '../api/gameApi';
import type { MightyConfig, MightyHint } from '../types/card';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';
import { useHintRequest } from './useHintRequest';

/** Default Mighty game configuration. */
export const DEFAULT_MIGHTY_CONFIG: MightyConfig = {
  cpuDifficulty: 1,
  minBid: 13,
  noTrumpExtra: 2,
  pointLimit: 100,
};

/** CPU difficulty level options for Mighty. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Mighty. */
export const POINT_LIMIT_OPTIONS = [50, 100, 200, 300, 500] as const;

/** Available minimum bid options for Mighty. */
export const MIN_BID_OPTIONS = [13, 14, 15, 16] as const;

/** Hook that manages Mighty game state, bidding, declarations, and player actions. */
export function useMightyGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config: mightyConfig, handleConfigChange } = useGameConfig<MightyConfig>(DEFAULT_MIGHTY_CONFIG);
  const [hint, setHint] = useState<MightyHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const [hintLoading, setHintLoading] = useState(false);

  const onSuccess = useCallback(() => {
    clearSelection();
    setHint(null);
  }, [clearSelection]);
  const { state, loading, error, exec: rawCall, retry } = useGameApi(mightyApi.exec, { onSuccess });

  const apiCall = useCallback((...args: Parameters<typeof rawCall>) => rawCall(...args), [rawCall]);

  useEffect(() => {
    apiCall(
      'reset',
      undefined,
      undefined,
      undefined,
      undefined,
      undefined,
      undefined,
      undefined,
      undefined,
      DEFAULT_MIGHTY_CONFIG,
    );
  }, [apiCall]);

  const handleBid = useCallback(
    (bid: number, noTrump: boolean) => {
      apiCall('bid', bid, noTrump);
    },
    [apiCall],
  );

  const handlePass = useCallback(() => {
    apiCall('bid', 0, false);
  }, [apiCall]);

  const handleTrumpAndFriend = useCallback(
    (trumpSuit: number, partnerSuit: number, partnerValue: number) => {
      apiCall('trump', undefined, undefined, undefined, trumpSuit, partnerSuit, partnerValue);
    },
    [apiCall],
  );

  const handleExchange = useCallback(
    (discardIndices: number[]) => {
      if (discardIndices.length !== 3) return;
      apiCall('exchange', undefined, undefined, undefined, undefined, undefined, undefined, discardIndices);
    },
    [apiCall],
  );

  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    apiCall('play', undefined, undefined, selectedCardIndices[0]);
  }, [apiCall, selectedCardIndices]);

  const handleJokerLead = useCallback(
    (jokerLeadSuit: number) => {
      if (selectedCardIndices.length !== 1) return;
      apiCall(
        'jokerlead',
        undefined,
        undefined,
        selectedCardIndices[0],
        undefined,
        undefined,
        undefined,
        undefined,
        jokerLeadSuit,
      );
    },
    [apiCall, selectedCardIndices],
  );

  const handleNextTrick = useCallback(() => {
    apiCall('next');
  }, [apiCall]);

  const handleNextRound = useCallback(() => {
    apiCall('nextround');
  }, [apiCall]);

  const handleHint = useHintRequest({
    fetchHint: () => mightyApi.exec('hint'),
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
    apiCall,
    mightyConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleBid,
    handlePass,
    handleTrumpAndFriend,
    handleExchange,
    handlePlay,
    handleJokerLead,
    handleNextTrick,
    handleNextRound,
    handleHint,
    retry,
  };
}
