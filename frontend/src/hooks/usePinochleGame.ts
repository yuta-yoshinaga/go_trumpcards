import { useCallback, useEffect, useState } from 'react';
import { pinochleApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import type { PinochleConfig } from '../types/card';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/**
 * Server-computed hint returned by the Pinochle `hint` command. Exactly one of
 * the value fields is set depending on the current phase: `bidAmount` or `pass`
 * during bidding, `suit` when declaring trump, and `cardIndex` during play.
 * `reason` is an i18n key under the `hintReason` namespace.
 */
export interface PinochleHint {
  cardIndex?: number;
  bidAmount?: number;
  pass?: boolean;
  suit?: number;
  reason: string;
}

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

  const [hint, setHint] = useState<PinochleHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const [hintLoading, setHintLoading] = useState(false);

  // Any successful game action invalidates the previous hint.
  const onSuccess = useCallback(() => setHint(null), []);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(pinochleApi.exec, { onSuccess });

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

  // The `hint` command returns only the hint object (not full game state), so
  // it is fetched directly and stored separately rather than via gameExec,
  // which would otherwise clobber the rendered game state.
  const handleHint = useCallback(async () => {
    setHintLoading(true);
    try {
      const res = await pinochleApi.exec('hint');
      // Server marshals the bare hint fields at the top level (cardIndex,
      // bidAmount, pass, suit, reason); when no hint applies `reason` is absent.
      const raw = res as unknown as Partial<PinochleHint>;
      setHint(raw.reason ? (raw as PinochleHint) : null);
      setHintError(null);
    } catch {
      setHintError(NETWORK_ERROR_MESSAGE());
    } finally {
      setHintLoading(false);
    }
  }, []);

  return {
    state,
    loading,
    error,
    hint,
    hintError,
    hintLoading,
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
