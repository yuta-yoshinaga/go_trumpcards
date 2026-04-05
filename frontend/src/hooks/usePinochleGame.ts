import { useCallback, useEffect } from 'react';
import { pinochleApi } from '../api/gameApi';
import type { PinochleConfig } from '../types/card';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Pinochle game configuration. */
export const DEFAULT_PINOCHLE_CONFIG: PinochleConfig = {
  cpuDifficulty: 1,
  pointLimit: 1500,
};

/** CPU difficulty level options for Pinochle. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Pinochle. */
export const POINT_LIMIT_OPTIONS = [500, 1000, 1500, 2000, 3000] as const;

/** Hook that manages Pinochle game state and player actions. */
export function usePinochleGame() {
  const { config: pinochleConfig, handleConfigChange } = useGameConfig<PinochleConfig>(DEFAULT_PINOCHLE_CONFIG);

  const { state, loading, error, exec: rawExec, retry } = useGameApi(pinochleApi.exec);

  const gameExec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    gameExec('reset', undefined, DEFAULT_PINOCHLE_CONFIG);
  }, [gameExec]);

  const handleReset = useCallback(() => {
    gameExec('reset', undefined, pinochleConfig);
  }, [gameExec, pinochleConfig]);

  const handleBid = useCallback(
    (amount: number) => {
      gameExec('bid', undefined, undefined, amount);
    },
    [gameExec],
  );

  const handlePass = useCallback(() => {
    gameExec('pass');
  }, [gameExec]);

  const handleCallTrump = useCallback(
    (suit: number) => {
      gameExec('trump', undefined, undefined, undefined, suit);
    },
    [gameExec],
  );

  const handleConfirmMelds = useCallback(() => {
    gameExec('meld');
  }, [gameExec]);

  const handlePlay = useCallback(
    (cardIndex: number) => {
      gameExec('play', cardIndex);
    },
    [gameExec],
  );

  const handleNextTrick = useCallback(() => {
    gameExec('next');
  }, [gameExec]);

  const handleNextRound = useCallback(() => {
    gameExec('nextround');
  }, [gameExec]);

  const handleHint = useCallback(() => {
    gameExec('hint');
  }, [gameExec]);

  return {
    state,
    loading,
    error,
    exec: rawExec,
    pinochleConfig,
    handleConfigChange,
    handleReset,
    handleBid,
    handlePass,
    handleCallTrump,
    handleConfirmMelds,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    handleHint,
    retry,
  };
}
