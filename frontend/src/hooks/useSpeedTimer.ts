import { useEffect, useRef, useState } from 'react';

/** localStorage key prefix for per-difficulty Speed best times (milliseconds). */
const BEST_TIME_KEY_PREFIX = 'speed_best_time_';

/** Reads the persisted best time (ms) for a CPU difficulty, or null when absent/invalid. */
function readBestTime(difficulty: number): number | null {
  try {
    const raw = localStorage.getItem(`${BEST_TIME_KEY_PREFIX}${difficulty}`);
    if (raw === null) return null;
    const n = Number.parseInt(raw, 10);
    return Number.isFinite(n) && n > 0 ? n : null;
  } catch {
    return null;
  }
}

/** Persists the best time (ms) for a CPU difficulty; failures are swallowed. */
function writeBestTime(difficulty: number, ms: number): void {
  try {
    localStorage.setItem(`${BEST_TIME_KEY_PREFIX}${difficulty}`, String(ms));
  } catch {
    // Ignore quota / disabled-storage errors — best time is a nice-to-have.
  }
}

/** Elapsed/best-time state returned by {@link useSpeedTimer}. */
export interface SpeedTimerState {
  /** Milliseconds elapsed for the current game (frozen once the game ends). */
  elapsedMs: number;
  /** Persisted best (fastest) time in ms for the current difficulty, or null. */
  bestMs: number | null;
  /** True when the just-finished win beat the previous best for this difficulty. */
  isNewBest: boolean;
}

/**
 * Tracks the elapsed time of a Speed game and persists the per-difficulty best
 * (fastest) time on a human win.
 *
 * Elapsed time is measured from a `Date.now()` start stamp so it stays accurate
 * even while the tab is backgrounded (the interval only refreshes the readout).
 * The timer restarts whenever `running` rises from false to true (a fresh game
 * or a reset), and freezes at the exact final time when `ended` becomes true.
 * The best time is recorded at most once per game via a ref guard.
 */
export function useSpeedTimer(running: boolean, ended: boolean, won: boolean, difficulty: number): SpeedTimerState {
  const [elapsedMs, setElapsedMs] = useState(0);
  const [bestMs, setBestMs] = useState<number | null>(() => readBestTime(difficulty));
  const [isNewBest, setIsNewBest] = useState(false);

  const startRef = useRef<number | null>(null);
  const prevRunningRef = useRef(false);
  const recordedRef = useRef(false);

  // Reload the persisted best whenever the difficulty changes.
  useEffect(() => {
    setBestMs(readBestTime(difficulty));
  }, [difficulty]);

  // Restart-on-rising-edge + live tick while running.
  useEffect(() => {
    if (running && !prevRunningRef.current) {
      startRef.current = Date.now();
      recordedRef.current = false;
      setElapsedMs(0);
      setIsNewBest(false);
    }
    prevRunningRef.current = running;

    if (!running) return;
    const tick = () => {
      if (startRef.current !== null) setElapsedMs(Date.now() - startRef.current);
    };
    tick();
    const id = setInterval(tick, 500);
    return () => clearInterval(id);
  }, [running]);

  // Freeze the final time and record the best (once) when the game ends.
  useEffect(() => {
    if (!ended || recordedRef.current || startRef.current === null) return;
    recordedRef.current = true;
    const finalMs = Date.now() - startRef.current;
    setElapsedMs(finalMs);
    if (won) {
      const prevBest = readBestTime(difficulty);
      if (prevBest === null || finalMs < prevBest) {
        writeBestTime(difficulty, finalMs);
        setBestMs(finalMs);
        setIsNewBest(true);
      }
    }
  }, [ended, won, difficulty]);

  return { elapsedMs, bestMs, isNewBest };
}
