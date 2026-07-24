import { useCallback, useState } from 'react';

/**
 * Builds a `read()` for a single flat stats record: parse + validate the value
 * persisted under `key`, falling back to `empty()` on a missing key, malformed
 * JSON, failed validation, or unavailable storage. Per-key-map stats (e.g.
 * Klondike variants, Spider difficulties) supply their own entry-sanitizing
 * reader instead. Issue #4303.
 */
export function createFlatStatsReader<Stat>(
  key: string,
  empty: () => Stat,
  validate: (value: unknown) => value is Stat,
): () => Stat {
  return () => {
    try {
      const raw = localStorage.getItem(key);
      if (!raw) return empty();
      const parsed: unknown = JSON.parse(raw);
      return validate(parsed) ? parsed : empty();
    } catch {
      return empty();
    }
  };
}

/** Configuration for {@link createLocalStorageStats}. */
export interface LocalStorageStatsConfig<Stat, Result, Update> {
  /** localStorage key the stats are persisted under. */
  key: string;
  /** Reads the current persisted stats (e.g. from {@link createFlatStatsReader}). */
  read: () => Stat;
  /**
   * Pure reducer folding a finished-game result into the previous stats,
   * returning the next stats and a game-specific "which personal bests were
   * beaten" update (a boolean for single-best games, an object for multi-best).
   */
  reduce: (prev: Stat, result: Result) => { stats: Stat; update: Update };
}

/**
 * Factory for a localStorage-backed best-record stats hook (issue #4303).
 * Centralizes the persist + useState plumbing copy-pasted across the solitaire /
 * best-record stats hooks; each game supplies its key, reader, and reduce.
 *
 * Returns `apply` (the provided `reduce`, re-exported so callers keep a pure
 * `applyXResult`) and `useStats()` — React state seeded from `read()`, plus
 * `recordResult`, which folds a result into the freshly-read stats, persists it
 * (swallowing quota / availability errors), and returns the reduce's `update`.
 */
export function createLocalStorageStats<Stat, Result, Update>(config: LocalStorageStatsConfig<Stat, Result, Update>) {
  const { key, read, reduce } = config;

  function useStats() {
    const [stats, setStats] = useState<Stat>(read);
    const recordResult = useCallback((result: Result): Update => {
      const { stats: next, update } = reduce(read(), result);
      setStats(next);
      try {
        localStorage.setItem(key, JSON.stringify(next));
      } catch {
        /* storage unavailable / quota exceeded */
      }
      return update;
    }, []);
    return { stats, recordResult };
  }

  return { apply: reduce, useStats };
}
