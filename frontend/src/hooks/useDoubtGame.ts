import { useCallback, useEffect, useRef, useState } from 'react';
import { doubtApi } from '../api/gameApi';
import type { DoubtConfig, DoubtCpuAction, DoubtPlayerData } from '../types/card';
import { valueName } from '../utils/cardUtils';
import { playerName } from '../utils/playerUtils';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';

export const DEFAULT_DOUBT_CONFIG: DoubtConfig = { doubtWindowSec: 10, cpuMemoryLevel: 1, penaltyDrawLimit: 0 };

export const DOUBT_WINDOW_OPTIONS = [3, 5, 10] as const;

export const CPU_MEMORY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

export const PENALTY_DRAW_LIMIT_OPTIONS = [0, 3, 5, 10] as const;

export function actionDesc(
  action: DoubtCpuAction,
  players: DoubtPlayerData[],
  t: (key: string, opts?: Record<string, unknown>) => string,
): string {
  const p = players[action.playerIdx];
  const name = p ? playerName(p.id, p.isHuman) : `Player ${action.playerIdx}`;
  return t('actionDesc', { name, count: action.cardCount, value: valueName(action.claimedValue) });
}

export function useDoubtGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const [claimedValue, setClaimedValue] = useState(1);
  const [countdown, setCountdown] = useState<number | null>(null);
  const [doubtConfig, setDoubtConfig] = useState<DoubtConfig>(DEFAULT_DOUBT_CONFIG);
  const countdownRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const autoSkipRef = useRef(false);
  const cpuDoubtersRef = useRef<number[]>([]);

  const onSuccess = useCallback(() => {
    clearSelection();
    setClaimedValue(1);
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec } = useGameApi(doubtApi.exec, { onSuccess });

  useEffect(() => {
    if (state) cpuDoubtersRef.current = state.cpuDoubters;
  }, [state]);

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

  const handleConfigChange = useCallback((key: keyof DoubtConfig, value: string) => {
    const parsed = Number(value);
    if (!Number.isNaN(parsed)) {
      setDoubtConfig((prev) => ({ ...prev, [key]: parsed }));
    }
  }, []);

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
    if (!state) return;
    if (state.phase === 1 && state.lastAction !== null) {
      const lastActionPlayer = state.players[state.lastAction.playerIdx];
      if (lastActionPlayer && !lastActionPlayer.isHuman) {
        startCountdown(state.doubtWindowSec);
      }
    }
  }, [state, startCountdown]);

  const handlePlay = useCallback(() => {
    exec('play', selectedCardIndices, claimedValue);
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
    claimedValue,
    setClaimedValue,
    stopCountdown,
    handleConfigChange,
    handlePlay,
    handleDoubt,
    handleSkip,
    handleCpuDoubtConfirm,
  };
}
