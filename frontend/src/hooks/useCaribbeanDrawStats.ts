import { useCallback, useMemo, useState } from 'react';

/** localStorage key for the Caribbean Draw session round history. */
export const CARIBBEANDRAW_HISTORY_KEY = 'trumpcards-caribbeandraw-history';

/** Maximum number of recent round records retained in localStorage. */
export const CARIBBEANDRAW_HISTORY_MAX = 200;

/**
 * Numeric outcome codes for a completed Caribbean Draw round. `WIN`/`LOSS` are the
 * signed hand results; `PUSH` is a tie (dealer qualified and hands compared equal).
 */
export const CaribbeanDrawOutcome = {
  PUSH: 0,
  WIN: 1,
  LOSS: 2,
} as const;

/** A single outcome code (one of the `CaribbeanDrawOutcome` values). */
export type CaribbeanDrawOutcomeCode = (typeof CaribbeanDrawOutcome)[keyof typeof CaribbeanDrawOutcome];

/** One recorded round: its win/loss/push outcome plus the net chip delta. */
export interface CaribbeanDrawRecord {
  outcome: CaribbeanDrawOutcomeCode;
  net: number;
}

/** Session tallies aggregated over a history slice. */
export interface CaribbeanDrawTally {
  wins: number;
  losses: number;
  pushes: number;
  /** Total rounds recorded (wins + losses + pushes). */
  hands: number;
  /** Cumulative net chip result across all recorded rounds. */
  net: number;
}

/** Maps a round's signed `result` to an outcome code (>0 win, <0 loss, 0 push). */
export function outcomeFromResult(result: number): CaribbeanDrawOutcomeCode {
  if (result > 0) return CaribbeanDrawOutcome.WIN;
  if (result < 0) return CaribbeanDrawOutcome.LOSS;
  return CaribbeanDrawOutcome.PUSH;
}

function isOutcomeCode(value: unknown): value is CaribbeanDrawOutcomeCode {
  return (
    value === CaribbeanDrawOutcome.WIN || value === CaribbeanDrawOutcome.LOSS || value === CaribbeanDrawOutcome.PUSH
  );
}

function isRecord(value: unknown): value is CaribbeanDrawRecord {
  if (typeof value !== 'object' || value === null) return false;
  const r = value as Record<string, unknown>;
  return isOutcomeCode(r.outcome) && typeof r.net === 'number' && Number.isFinite(r.net);
}

/** Reads and validates the round history from localStorage; returns [] on any error. */
export function readCaribbeanDrawHistory(): CaribbeanDrawRecord[] {
  try {
    const raw = localStorage.getItem(CARIBBEANDRAW_HISTORY_KEY);
    if (!raw) return [];
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed.filter(isRecord);
  } catch {
    return [];
  }
}

/** Folds a history slice into win / loss / push counts and cumulative net. */
export function tallyCaribbeanDrawHistory(history: readonly CaribbeanDrawRecord[]): CaribbeanDrawTally {
  const tally: CaribbeanDrawTally = { wins: 0, losses: 0, pushes: 0, hands: 0, net: 0 };
  for (const rec of history) {
    if (rec.outcome === CaribbeanDrawOutcome.WIN) tally.wins += 1;
    else if (rec.outcome === CaribbeanDrawOutcome.LOSS) tally.losses += 1;
    else tally.pushes += 1;
    tally.hands += 1;
    tally.net += rec.net;
  }
  return tally;
}

/**
 * Hook that persists the Caribbean Draw session round history in localStorage.
 * `recordRound` appends one finished round (the caller records each round only
 * once), capping the stored history at `CARIBBEANDRAW_HISTORY_MAX`.
 * `clearHistory` empties it. `tally` is the derived win / loss / push counts and
 * cumulative net chips over the full history.
 */
export function useCaribbeanDrawStats() {
  const [history, setHistory] = useState<CaribbeanDrawRecord[]>(readCaribbeanDrawHistory);

  const recordRound = useCallback((record: CaribbeanDrawRecord) => {
    setHistory((prev) => {
      const next = [...prev, record].slice(-CARIBBEANDRAW_HISTORY_MAX);
      try {
        localStorage.setItem(CARIBBEANDRAW_HISTORY_KEY, JSON.stringify(next));
      } catch {
        /* storage unavailable / quota exceeded */
      }
      return next;
    });
  }, []);

  const clearHistory = useCallback(() => {
    setHistory([]);
    try {
      localStorage.removeItem(CARIBBEANDRAW_HISTORY_KEY);
    } catch {
      /* storage unavailable */
    }
  }, []);

  const tally = useMemo(() => tallyCaribbeanDrawHistory(history), [history]);

  return { history, tally, recordRound, clearHistory };
}
