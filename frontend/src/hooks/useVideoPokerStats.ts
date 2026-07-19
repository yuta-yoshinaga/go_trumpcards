import { useCallback, useEffect, useRef, useState } from 'react';
import type { VideoPokerResponse } from '../types/card';
import { videoPokerHandNameToRowKey } from '../utils/videoPokerPayout';
import {
  emptyVideoPokerStats,
  readVideoPokerStats,
  recordVideoPokerResult,
  type VideoPokerStats,
  writeVideoPokerStats,
} from '../utils/videoPokerStats';

/** What {@link useVideoPokerStats} returns. */
export interface UseVideoPokerStats {
  /** Current session statistics for the variant. */
  stats: VideoPokerStats;
  /** Clears the session statistics (both in state and localStorage). */
  clear: () => void;
}

/**
 * Tracks and persists per-variant Video Poker session statistics.
 *
 * Records exactly one hand per RESULT-phase response: a ref keyed on the state
 * object reference guards against double-counting when the effect re-runs
 * without a fresh result. When `enabled` is false the hook is inert (no reads,
 * writes, or recording) so sibling variants that share the component are
 * unaffected.
 */
export function useVideoPokerStats(
  gameName: string,
  state: VideoPokerResponse | null,
  isResultPhase: boolean,
  enabled: boolean,
): UseVideoPokerStats {
  const [stats, setStats] = useState<VideoPokerStats>(() =>
    enabled ? readVideoPokerStats(gameName) : emptyVideoPokerStats(),
  );
  const recordedRef = useRef<VideoPokerResponse | null>(null);

  useEffect(() => {
    if (!enabled || !isResultPhase || !state) return;
    // Record once per result: the state object is a fresh reference per API
    // response, so a repeated effect run for the same hand is skipped here.
    if (recordedRef.current === state) return;
    recordedRef.current = state;
    setStats((prev) => {
      const next = recordVideoPokerResult(prev, {
        bet: state.betAmount,
        payout: state.payout,
        rowKey: videoPokerHandNameToRowKey(state.handName),
      });
      writeVideoPokerStats(gameName, next);
      return next;
    });
  }, [enabled, isResultPhase, state, gameName]);

  const clear = useCallback(() => {
    const empty = emptyVideoPokerStats();
    setStats(empty);
    if (enabled) writeVideoPokerStats(gameName, empty);
  }, [enabled, gameName]);

  return { stats, clear };
}
