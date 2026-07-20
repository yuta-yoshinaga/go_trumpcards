import { useCallback, useState } from 'react';
import type { GolfCard } from '../types/card';

/** localStorage key for the Golf 9-hole (9-deal cumulative) scorecard. */
export const GOLF_NINE_HOLE_KEY = 'trumpcards-golf-ninehole';

/** Number of deals (holes) that make up a full 9-hole round. */
export const GOLF_TOTAL_HOLES = 9;

/**
 * Persisted 9-hole state. `scores` holds the per-hole scores (cards left on the
 * tableau at deal end — lower is better) recorded so far, in play order.
 */
export interface GolfNineHoleState {
  /** Whether 9-hole mode is active. */
  enabled: boolean;
  /** Recorded per-hole scores (0..GOLF_TOTAL_HOLES entries). */
  scores: number[];
}

/** Returns a zeroed 9-hole state (mode off, no holes played). */
export function emptyGolfNineHoleState(): GolfNineHoleState {
  return { enabled: false, scores: [] };
}

function isValidState(value: unknown): value is GolfNineHoleState {
  if (typeof value !== 'object' || value === null) return false;
  const s = value as Record<string, unknown>;
  return (
    typeof s.enabled === 'boolean' &&
    Array.isArray(s.scores) &&
    s.scores.length <= GOLF_TOTAL_HOLES &&
    s.scores.every((n) => typeof n === 'number' && Number.isFinite(n))
  );
}

/** Reads and validates the 9-hole state from localStorage; returns a zeroed state on any error. */
export function readGolfNineHoleState(): GolfNineHoleState {
  try {
    const raw = localStorage.getItem(GOLF_NINE_HOLE_KEY);
    if (!raw) return emptyGolfNineHoleState();
    const parsed: unknown = JSON.parse(raw);
    if (!isValidState(parsed)) return emptyGolfNineHoleState();
    return parsed;
  } catch {
    return emptyGolfNineHoleState();
  }
}

/** 1-based number of the hole currently being played, capped at GOLF_TOTAL_HOLES. */
export function golfCurrentHole(state: GolfNineHoleState): number {
  return Math.min(state.scores.length + 1, GOLF_TOTAL_HOLES);
}

/** True once all GOLF_TOTAL_HOLES holes have a recorded score. */
export function golfNineHoleComplete(state: GolfNineHoleState): boolean {
  return state.scores.length >= GOLF_TOTAL_HOLES;
}

/** Cumulative total across all recorded holes (lower is better). */
export function golfNineHoleTotal(state: GolfNineHoleState): number {
  return state.scores.reduce((sum, n) => sum + n, 0);
}

/**
 * Pure reducer: appends a finished deal's score as the next hole. Once the round
 * is full (GOLF_TOTAL_HOLES holes) further calls are no-ops, so an accidental
 * double-record cannot inflate the card.
 */
export function recordGolfHole(state: GolfNineHoleState, score: number): GolfNineHoleState {
  if (state.scores.length >= GOLF_TOTAL_HOLES) return state;
  return { ...state, scores: [...state.scores, score] };
}

/**
 * Counts the cards still on the tableau — the Golf deal score. Golf clears the
 * board by removing cards to the waste, so cards left behind are the "strokes"
 * for that hole (0 on a clear).
 */
export function countGolfRemaining(layout: GolfCard[][]): number {
  return layout.reduce((sum, col) => sum + col.reduce((c, gc) => c + (gc.card && !gc.removed ? 1 : 0), 0), 0);
}

/**
 * Hook that persists the Golf 9-hole scorecard in localStorage. Enabling (or
 * disabling) the mode starts a fresh card; `recordHole` appends the just-finished
 * deal's remaining-card score (capped at 9 holes); `resetCard` clears the scores
 * while keeping the mode on. Modeled on `usePyramidStats`.
 */
export function useGolfNineHole() {
  const [nineHole, setNineHole] = useState<GolfNineHoleState>(readGolfNineHoleState);

  const persist = useCallback((next: GolfNineHoleState) => {
    setNineHole(next);
    try {
      localStorage.setItem(GOLF_NINE_HOLE_KEY, JSON.stringify(next));
    } catch {
      /* storage unavailable / quota exceeded */
    }
  }, []);

  const setEnabled = useCallback(
    (enabled: boolean) => {
      // Toggling the mode always starts a fresh scorecard.
      persist({ enabled, scores: [] });
    },
    [persist],
  );

  const recordHole = useCallback((score: number) => {
    setNineHole((prev) => {
      const next = recordGolfHole(prev, score);
      if (next === prev) return prev;
      try {
        localStorage.setItem(GOLF_NINE_HOLE_KEY, JSON.stringify(next));
      } catch {
        /* storage unavailable / quota exceeded */
      }
      return next;
    });
  }, []);

  const resetCard = useCallback(() => {
    setNineHole((prev) => {
      const next: GolfNineHoleState = { enabled: prev.enabled, scores: [] };
      try {
        localStorage.setItem(GOLF_NINE_HOLE_KEY, JSON.stringify(next));
      } catch {
        /* storage unavailable / quota exceeded */
      }
      return next;
    });
  }, []);

  return { nineHole, setEnabled, recordHole, resetCard };
}
