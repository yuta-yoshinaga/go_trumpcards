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

  /** Load profile data from localStorage. Returns undefined if not found or malformed. */
  const loadProfile = useCallback((): unknown | undefined => {
    const data = localStorage.getItem(storageKey);
    if (!data) return undefined;
    try {
      return JSON.parse(data);
    } catch {
      return undefined;
    }
  }, [storageKey]);

  /** Remove profile data from localStorage. */
  const clearProfile = useCallback(() => {
    localStorage.removeItem(storageKey);
  }, [storageKey]);

  return { saveProfile, loadProfile, clearProfile };
}
