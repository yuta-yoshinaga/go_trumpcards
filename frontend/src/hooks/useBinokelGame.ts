import { useCallback, useEffect, useState } from 'react';
import { binokelApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import type { BinokelConfig } from '../types/card';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';
import { useIsMounted } from './useIsMounted';

/**
 * Server-computed hint returned by the Binokel `hint` command. Exactly one of
 * the value fields is set depending on the current phase: `bidAmount` or `pass`
 * during bidding, `suit` when declaring trump, and `cardIndex` during play.
 * `reason` is an i18n key under the `hintReason` namespace.
 */
export interface BinokelHint {
  cardIndex?: number;
  bidAmount?: number;
  pass?: boolean;
  suit?: number;
  reason: string;
}

/** Default Binokel game configuration. */
export const DEFAULT_BINOKEL_CONFIG: BinokelConfig = {
  cpuDifficulty: 1,
  pointLimit: 1500,
};

/** CPU difficulty level options for Binokel. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point limit options for Binokel. */
export const POINT_LIMIT_OPTIONS = [500, 1000, 1500, 2000, 3000] as const;

/** Hook that manages Binokel game state and player actions. */
export function useBinokelGame() {
  const { config: binokelConfig, handleConfigChange } = useGameConfig<BinokelConfig>(DEFAULT_BINOKEL_CONFIG);

  const [hint, setHint] = useState<BinokelHint | null>(null);
  const [hintError, setHintError] = useState<string | null>(null);
  const [hintLoading, setHintLoading] = useState(false);

  // Any successful game action invalidates the previous hint.
  const onSuccess = useCallback(() => setHint(null), []);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(binokelApi.exec, { onSuccess });

  const gameExec = useCallback((...args: Parameters<typeof rawExec>) => rawExec(...args), [rawExec]);

  useEffect(() => {
    gameExec('reset', undefined, DEFAULT_BINOKEL_CONFIG);
  }, [gameExec]);

  const handleReset = useCallback(() => {
    gameExec('reset', undefined, binokelConfig);
  }, [gameExec, binokelConfig]);

  const handleBid = useCallback(
    (amount: number) => {
      gameExec('bid', undefined, undefined, amount);
    },
    [gameExec],
  );

  const handlePass = useCallback(() => {
    gameExec('pass');
  }, [gameExec]);

  const handleDiscard = useCallback(
    (discardIndices: number[]) => {
      gameExec('discard', undefined, undefined, undefined, undefined, discardIndices);
    },
    [gameExec],
  );

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
  const isMounted = useIsMounted();

  const handleHint = useCallback(async () => {
    setHintLoading(true);
    try {
      const res = await binokelApi.exec('hint');
      // Navigating away mid-request must not write to a gone component (#4447).
      if (!isMounted()) return;
      // **ヒントは state レスポンスの `hint` に入る。**以前は裸のヒント構造体が
      // 返っていたので最上位から読んでいたが、他ゲームと同じ形に揃えた (#4483)。
      const hint = res.hint as BinokelHint | undefined;
      setHint(hint?.reason ? hint : null);
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
    exec: rawExec,
    binokelConfig,
    handleConfigChange,
    handleReset,
    handleBid,
    handlePass,
    handleDiscard,
    handleCallTrump,
    handleConfirmMelds,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    handleHint,
    retry,
  };
}
