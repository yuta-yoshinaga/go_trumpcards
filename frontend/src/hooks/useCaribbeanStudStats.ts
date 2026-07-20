import { useCallback, useMemo, useState } from 'react';

/** localStorage key for the Caribbean Stud session round history. */
export const CARIBBEANSTUD_HISTORY_KEY = 'trumpcards-caribbeanstud-history';

/** Maximum number of recent round records retained in localStorage. */
export const CARIBBEANSTUD_HISTORY_MAX = 200;

/**
 * Numeric outcome codes for a completed Caribbean Stud round. `WIN`/`LOSS` are the
 * signed hand results; `PUSH` is a tie (dealer qualified and hands compared equal).
 */
export const CaribbeanStudOutcome = {
  PUSH: 0,
  WIN: 1,
  LOSS: 2,
} as const;

/** A single outcome code (one of the `CaribbeanStudOutcome` values). */
export type CaribbeanStudOutcomeCode = (typeof CaribbeanStudOutcome)[keyof typeof CaribbeanStudOutcome];

/** One recorded round: its win/loss/push outcome plus the net chip delta. */
export interface CaribbeanStudRecord {
  outcome: CaribbeanStudOutcomeCode;
  net: number;
}

/** Session tallies aggregated over a history slice. */
export interface CaribbeanStudTally {
  wins: number;
  losses: number;
  pushes: number;
  /** Total rounds recorded (wins + losses + pushes). */
  hands: number;
  /** Cumulative net chip result across all recorded rounds. */
  net: number;
}

/** Maps a round's signed `result` to an outcome code (>0 win, <0 loss, 0 push). */
export function outcomeFromResult(result: number): CaribbeanStudOutcomeCode {
  if (result > 0) return CaribbeanStudOutcome.WIN;
  if (result < 0) return CaribbeanStudOutcome.LOSS;
  return CaribbeanStudOutcome.PUSH;
}

function isOutcomeCode(value: unknown): value is CaribbeanStudOutcomeCode {
  return (
    value === CaribbeanStudOutcome.WIN || value === CaribbeanStudOutcome.LOSS || value === CaribbeanStudOutcome.PUSH
  );
}

function isRecord(value: unknown): value is CaribbeanStudRecord {
  if (typeof value !== 'object' || value === null) return false;
  const r = value as Record<string, unknown>;
  return isOutcomeCode(r.outcome) && typeof r.net === 'number' && Number.isFinite(r.net);
}

/** Reads and validates the round history from localStorage; returns [] on any error. */
export function readCaribbeanStudHistory(): CaribbeanStudRecord[] {
  try {
    const raw = localStorage.getItem(CARIBBEANSTUD_HISTORY_KEY);
    if (!raw) return [];
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed.filter(isRecord);
  } catch {
    return [];
  }
}

/** Folds a history slice into win / loss / push counts and cumulative net. */
export function tallyCaribbeanStudHistory(history: readonly CaribbeanStudRecord[]): CaribbeanStudTally {
  const tally: CaribbeanStudTally = { wins: 0, losses: 0, pushes: 0, hands: 0, net: 0 };
  for (const rec of history) {
    if (rec.outcome === CaribbeanStudOutcome.WIN) tally.wins += 1;
    else if (rec.outcome === CaribbeanStudOutcome.LOSS) tally.losses += 1;
    else tally.pushes += 1;
    tally.hands += 1;
    tally.net += rec.net;
  }
  return tally;
}

/**
 * Hook that persists the Caribbean Stud session round history in localStorage.
 * `recordRound` appends one finished round (the caller records each round only
 * once), capping the stored history at `CARIBBEANSTUD_HISTORY_MAX`.
 * `clearHistory` empties it. `tally` is the derived win / loss / push counts and
 * cumulative net chips over the full history.
 */
export function useCaribbeanStudStats() {
  const [history, setHistory] = useState<CaribbeanStudRecord[]>(readCaribbeanStudHistory);

  const recordRound = useCallback((record: CaribbeanStudRecord) => {
    setHistory((prev) => {
      const next = [...prev, record].slice(-CARIBBEANSTUD_HISTORY_MAX);
      try {
        localStorage.setItem(CARIBBEANSTUD_HISTORY_KEY, JSON.stringify(next));
      } catch {
        /* storage unavailable / quota exceeded */
      }
      return next;
    });
  }, []);

  const clearHistory = useCallback(() => {
    setHistory([]);
    try {
      localStorage.removeItem(CARIBBEANSTUD_HISTORY_KEY);
    } catch {
      /* storage unavailable */
    }
  }, []);

  const tally = useMemo(() => tallyCaribbeanStudHistory(history), [history]);

  return { history, tally, recordRound, clearHistory };
}
