import { useCallback, useEffect, useRef, useState } from 'react';
import { doubtApi } from '../api/gameApi';
import type { DoubtConfig, DoubtCpuAction, DoubtPlayerData } from '../types/card';
import { valueName } from '../utils/cardUtils';
import { playerName } from '../utils/playerUtils';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Doubt game configuration. */
export const DEFAULT_DOUBT_CONFIG: DoubtConfig = {
  doubtWindowSec: 10,
  cpuMemoryLevel: 1,
  penaltyDrawLimit: 0,
  cpuHesitationEnabled: false,
  cpuMetaAI: false,
};

/** Available doubt window duration options in seconds. */
export const DOUBT_WINDOW_OPTIONS = [3, 5, 10, 15, 30, 60] as const;

/** CPU memory level options for Doubt. */
export const CPU_MEMORY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available penalty draw limit options for Doubt (0 = unlimited). */
export const PENALTY_DRAW_LIMIT_OPTIONS = [0, 3, 5, 10] as const;

/** Build a human-readable description of a Doubt play action. */
export function actionDesc(
  action: DoubtCpuAction,
  players: DoubtPlayerData[],
  t: (key: string, opts?: Record<string, unknown>) => string,
): string {
  const p = players[action.playerIdx];
  const name = p ? playerName(p.id, p.isHuman) : `Player ${action.playerIdx}`;
  const key = action.hasTell ? 'actionDescWithTell' : 'actionDesc';
  return t(key, { name, count: action.cardCount, value: valueName(action.claimedValue) });
}

/** Hook that manages Doubt game state, countdown timer, and player actions. */
export function useDoubtGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const [claimedValue, setClaimedValue] = useState(1);
  const [countdown, setCountdown] = useState<number | null>(null);
  const {
    config: doubtConfig,
    handleConfigChange,
    handleToggle: handleConfigToggle,
  } = useGameConfig<DoubtConfig>(DEFAULT_DOUBT_CONFIG);
  const countdownRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const autoSkipRef = useRef(false);
  const cpuDoubtersRef = useRef<number[]>([]);
  const playTurnStartRef = useRef<number>(0);

  const onSuccess = useCallback(() => {
    clearSelection();
    setClaimedValue(1);
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec, retry } = useGameApi(doubtApi.exec, { onSuccess });

  useEffect(() => {
    if (state) cpuDoubtersRef.current = state.cpuDoubters;
  }, [state]);

  // Clear the countdown interval on unmount. It was only ever cleared by
  // stopCountdown or by its own tick reaching zero, so unmounting mid-countdown
  // left it running: the next tick called setCountdown on a gone component, and
  // in a test environment that is `ReferenceError: window is not defined` thrown
  // from React's dispatchSetState — an unhandled error that fails the whole
  // vitest run even when every test passed. Surfaced on CI by #4429's added
  // tests shifting shard timing; latent before that.
  useEffect(
    () => () => {
      if (countdownRef.current !== null) {
        clearInterval(countdownRef.current);
        countdownRef.current = null;
      }
    },
    [],
  );

  const stopCountdown = useCallback(() => {
    autoSkipRef.current = false;
    if (countdownRef.current !== null) {
      clearInterval(countdownRef.current);
      countdownRef.current = null;
    }
    setCountdown(null);
  }, []);

  const startCountdown = useCallback(
    (sec: number) => {
      stopCountdown();
      setCountdown(sec);
      countdownRef.current = setInterval(() => {
        setCountdown((prev) => {
          const cur = prev as number;
          if (cur <= 1) {
            clearInterval(countdownRef.current as ReturnType<typeof setInterval>);
            countdownRef.current = null;
            autoSkipRef.current = true;
            return null;
          }
          return cur - 1;
        });
      }, 1000);
    },
    [stopCountdown],
  );

  const exec = useCallback(
    (...args: Parameters<typeof rawExec>) => {
      stopCountdown();
      return rawExec(...args);
    },
    [rawExec, stopCountdown],
  );

  useEffect(() => {
    exec('reset', undefined, undefined, undefined, DEFAULT_DOUBT_CONFIG);
  }, [exec]);

  useEffect(() => {
    if (countdown !== null) return;
    if (!autoSkipRef.current) return;
    autoSkipRef.current = false;
    exec('skip', undefined, undefined, cpuDoubtersRef.current);
  }, [countdown, exec]);

  useEffect(() => {
    if (state && state.phase === 0 && state.players[state.currentTurn]?.isHuman) {
      playTurnStartRef.current = Date.now();
    }
  }, [state]);

  useEffect(() => {
    if (!state) return;
    if (state.phase === 1 && state.lastAction !== null) {
      const lastActionPlayer = state.players[state.lastAction.playerIdx];
      if (lastActionPlayer && !lastActionPlayer.isHuman) {
        // CPU hesitation: delay before showing the result and starting countdown
        const lastCpuAction = state.cpuActions[state.cpuActions.length - 1];
        const hesMs = lastCpuAction?.hesitationMs ?? 0;
        if (hesMs > 0) {
          const timer = setTimeout(() => startCountdown(state.doubtWindowSec), hesMs);
          return () => clearTimeout(timer);
        }
        startCountdown(state.doubtWindowSec);
      }
    }
  }, [state, startCountdown]);

  const handlePlay = useCallback(() => {
    const elapsed = playTurnStartRef.current > 0 ? Date.now() - playTurnStartRef.current : 0;
    playTurnStartRef.current = 0;
    exec('play', selectedCardIndices, claimedValue, undefined, undefined, elapsed);
  }, [exec, selectedCardIndices, claimedValue]);

  const handleDoubt = useCallback(() => {
    stopCountdown();
    exec('doubt', undefined, undefined, [0, ...(state?.cpuDoubters ?? [])]);
  }, [exec, stopCountdown, state?.cpuDoubters]);

  const handleSkip = useCallback(() => {
    stopCountdown();
    exec('skip', undefined, undefined, state?.cpuDoubters);
  }, [exec, stopCountdown, state?.cpuDoubters]);

  const handleCpuDoubtConfirm = useCallback(() => {
    exec('doubt', undefined, undefined, state?.cpuDoubters);
  }, [exec, state?.cpuDoubters]);

  return {
    state,
    loading,
    error,
    exec,
    countdown,
    doubtConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    claimedValue,
    setClaimedValue,
    stopCountdown,
    handleConfigChange,
    handleConfigToggle,
    handlePlay,
    handleDoubt,
    handleSkip,
    handleCpuDoubtConfirm,
    retry,
  };
}
