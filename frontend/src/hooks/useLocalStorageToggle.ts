import { useCallback, useState } from 'react';

/** Reads a boolean value from localStorage, returning defaultValue if absent or invalid. */
function readBoolean(key: string, defaultValue: boolean): boolean {
  const stored = localStorage.getItem(key);
  if (stored === 'true') return true;
  if (stored === 'false') return false;
  return defaultValue;
}

/** A hook that persists a boolean toggle in localStorage. */
export function useLocalStorageToggle(key: string, defaultValue: boolean): [boolean, (value: boolean) => void] {
  const [value, setValue] = useState(() => readBoolean(key, defaultValue));

  const setAndPersist = useCallback(
    (newValue: boolean) => {
      setValue(newValue);
      localStorage.setItem(key, String(newValue));
    },
    [key],
  );

  return [value, setAndPersist];
}

/** Reads a finite number from localStorage, returning defaultValue if absent or invalid. */
function readNumber(key: string, defaultValue: number): number {
  const stored = localStorage.getItem(key);
  if (stored === null) return defaultValue;
  const parsed = Number(stored);
  return Number.isFinite(parsed) ? parsed : defaultValue;
}

/** A hook that persists a numeric value in localStorage. */
export function useLocalStorageNumber(key: string, defaultValue: number): [number, (value: number) => void] {
  const [value, setValue] = useState(() => readNumber(key, defaultValue));

  const setAndPersist = useCallback(
    (newValue: number) => {
      setValue(newValue);
      localStorage.setItem(key, String(newValue));
    },
    [key],
  );

  return [value, setAndPersist];
}
