import { useCallback, useEffect, useState } from 'react';

/** CPU replay animation speed preference (#1649). */
export type ReplaySpeed = 'normal' | 'fast' | 'instant';

/** localStorage key used to persist {@link ReplaySpeed}. */
export const REPLAY_SPEED_STORAGE_KEY = 'cpuReplaySpeed';

/** Default speed when the user has not chosen one. */
export const DEFAULT_REPLAY_SPEED: ReplaySpeed = 'normal';

/** All accepted {@link ReplaySpeed} values, used to validate untrusted input. */
export const VALID_SPEEDS: ReadonlySet<ReplaySpeed> = new Set<ReplaySpeed>(['normal', 'fast', 'instant']);

/** Type guard — returns true when `value` is one of the supported {@link ReplaySpeed} strings. */
export function isReplaySpeed(value: unknown): value is ReplaySpeed {
  return typeof value === 'string' && VALID_SPEEDS.has(value as ReplaySpeed);
}

/** Reads a {@link ReplaySpeed} from localStorage, falling back to the default. */
function readSpeed(): ReplaySpeed {
  try {
    const raw = localStorage.getItem(REPLAY_SPEED_STORAGE_KEY);
    if (isReplaySpeed(raw)) return raw;
  } catch {
    // localStorage unavailable (private mode, SSR) — fall through to default
  }
  return DEFAULT_REPLAY_SPEED;
}

/** Returns the multiplier applied to replay delays for the given speed. */
export function multiplierForSpeed(speed: ReplaySpeed): number {
  switch (speed) {
    case 'fast':
      return 0.3;
    case 'instant':
      return 0;
    default:
      return 1;
  }
}

/**
 * Reads the current replay-speed multiplier from localStorage.
 * Designed to be called per-delay so changes mid-animation take effect on the
 * next step. Returns 1 (normal) when nothing is stored or the value is invalid.
 */
export function getReplaySpeedMultiplier(): number {
  return multiplierForSpeed(readSpeed());
}

/**
 * Persistent React state for the CPU replay speed.
 * Stored under {@link REPLAY_SPEED_STORAGE_KEY} so the choice carries across
 * games and reloads.
 */
export function useReplaySpeed(): [ReplaySpeed, (next: ReplaySpeed) => void] {
  const [speed, setSpeed] = useState<ReplaySpeed>(() => readSpeed());

  // Cross-tab sync: when another tab updates the preference, mirror the change
  // here so the visible select reflects the latest value (animation timing is
  // already fresh because runReplay reads localStorage per step).
  useEffect(() => {
    const handleStorage = (event: StorageEvent) => {
      if (event.key !== REPLAY_SPEED_STORAGE_KEY) return;
      if (event.newValue === null) {
        setSpeed(DEFAULT_REPLAY_SPEED);
        return;
      }
      if (isReplaySpeed(event.newValue)) setSpeed(event.newValue);
    };
    window.addEventListener('storage', handleStorage);
    return () => window.removeEventListener('storage', handleStorage);
  }, []);

  const setAndPersist = useCallback((next: ReplaySpeed) => {
    setSpeed(next);
    try {
      localStorage.setItem(REPLAY_SPEED_STORAGE_KEY, next);
    } catch {
      // ignore quota / private-mode errors
    }
  }, []);

  return [speed, setAndPersist];
}
