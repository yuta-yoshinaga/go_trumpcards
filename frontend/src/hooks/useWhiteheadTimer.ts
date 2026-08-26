import { useCallback, useEffect, useRef, useState } from 'react';

/** Hook that provides an elapsed-seconds timer and Vegas time bonus calculation. */
export function useWhiteheadTimer(isPlaying: boolean) {
  const [elapsedSeconds, setElapsedSeconds] = useState(0);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const clearTimer = useCallback(() => {
    if (intervalRef.current !== null) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
  }, []);

  const resetTimer = useCallback(() => {
    clearTimer();
    setElapsedSeconds(0);
  }, [clearTimer]);

  useEffect(() => {
    if (isPlaying) {
      clearTimer();
      intervalRef.current = setInterval(() => {
        setElapsedSeconds((prev) => prev + 1);
      }, 1000);
    } else {
      clearTimer();
    }
    return clearTimer;
  }, [isPlaying, clearTimer]);

  const timeBonus = useCallback((seconds: number) => {
    if (seconds <= 0) return 0;
    return Math.floor(700000 / seconds);
  }, []);

  return { elapsedSeconds, resetTimer, timeBonus };
}
