import { useCallback, useLayoutEffect, useRef, useState } from 'react';

/** Hook that provides an elapsed-seconds timer and Vegas time bonus calculation. */
export function useKlondikeTimer(isPlaying: boolean) {
  const [elapsedSeconds, setElapsedSeconds] = useState(0);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const startedAtRef = useRef<number | null>(null);
  const accumulatedSecondsRef = useRef(0);

  const clearTimer = useCallback(() => {
    if (intervalRef.current !== null) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
  }, []);

  const resetTimer = useCallback(() => {
    accumulatedSecondsRef.current = 0;
    if (startedAtRef.current !== null) startedAtRef.current = Date.now();
    setElapsedSeconds(0);
  }, []);

  useLayoutEffect(() => {
    if (isPlaying) {
      clearTimer();
      startedAtRef.current = Date.now();
      intervalRef.current = setInterval(() => {
        const startedAt = startedAtRef.current;
        if (startedAt !== null) {
          setElapsedSeconds(accumulatedSecondsRef.current + Math.floor((Date.now() - startedAt) / 1000));
        }
      }, 1000);
    } else {
      clearTimer();
      const startedAt = startedAtRef.current;
      if (startedAt !== null) {
        accumulatedSecondsRef.current += Math.floor((Date.now() - startedAt) / 1000);
        startedAtRef.current = null;
        setElapsedSeconds(accumulatedSecondsRef.current);
      }
    }
    return clearTimer;
  }, [isPlaying, clearTimer]);

  const timeBonus = useCallback((seconds: number) => {
    if (seconds <= 0) return 0;
    return Math.floor(700000 / seconds);
  }, []);

  return { elapsedSeconds, resetTimer, timeBonus };
}
