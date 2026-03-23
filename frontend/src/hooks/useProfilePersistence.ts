import { useCallback } from 'react';

/** LocalStorage key prefix for meta-AI profiles. */
const KEY_PREFIX = 'metaai_';

/**
 * Hook that manages meta-AI profile persistence in localStorage.
 * Provides save/load/clear operations for a given game key.
 */
export function useProfilePersistence(gameKey: string) {
  const storageKey = `${KEY_PREFIX}${gameKey}`;

  /** Save profile data to localStorage. No-op if profile is falsy. */
  const saveProfile = useCallback(
    (profile: unknown) => {
      if (profile) {
        localStorage.setItem(storageKey, JSON.stringify(profile));
      }
    },
    [storageKey],
  );

  /** Load profile data from localStorage. Returns undefined if not found. */
  const loadProfile = useCallback((): unknown | undefined => {
    const data = localStorage.getItem(storageKey);
    return data ? JSON.parse(data) : undefined;
  }, [storageKey]);

  /** Remove profile data from localStorage. */
  const clearProfile = useCallback(() => {
    localStorage.removeItem(storageKey);
  }, [storageKey]);

  return { saveProfile, loadProfile, clearProfile };
}
