import { useCallback, useMemo, useState } from 'react';

/** localStorage key for the Casino War win/loss round history. */
export const CASINOWAR_HISTORY_KEY = 'trumpcards-casinowar-history';

/** Maximum number of recent round outcomes retained in localStorage. */
export const CASINOWAR_HISTORY_MAX = 100;

/**
 * Numeric outcome codes for a completed Casino War round. `WIN`/`LOSS` drive the
 * trend bar's two sides; `TIE` (a push) is neutral and does not break a streak.
 */
export const CasinoWarOutcome = {
  TIE: 0,
  WIN: 1,
  LOSS: 2,
} as const;

/** A single recorded round outcome (one of the `CasinoWarOutcome` codes). */
export type CasinoWarOutcomeCode = (typeof CasinoWarOutcome)[keyof typeof CasinoWarOutcome];

interface CasinoWarRound {
  outcome: CasinoWarOutcomeCode;
  netChange: number;
}

/** Win / loss / tie counts aggregated over a history slice. */
export interface CasinoWarTally {
  wins: number;
  losses: number;
  ties: number;
}

/** Maps a round's signed `result` to an outcome code (>0 win, <0 loss, 0 push). */
export function outcomeFromResult(result: number): CasinoWarOutcomeCode {
  if (result > 0) return CasinoWarOutcome.WIN;
  if (result < 0) return CasinoWarOutcome.LOSS;
  return CasinoWarOutcome.TIE;
}

function isOutcomeCode(value: unknown): value is CasinoWarOutcomeCode {
  return value === CasinoWarOutcome.WIN || value === CasinoWarOutcome.LOSS || value === CasinoWarOutcome.TIE;
}

/** Reads and validates the round history from localStorage; returns [] on any error. */
export function readCasinoWarHistory(): CasinoWarOutcomeCode[] {
  return readCasinoWarRounds().map(({ outcome }) => outcome);
}

function readCasinoWarRounds(): CasinoWarRound[] {
  try {
    const raw = localStorage.getItem(CASINOWAR_HISTORY_KEY);
    if (!raw) return [];
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed.flatMap((entry) => {
      if (isOutcomeCode(entry)) return [{ outcome: entry, netChange: 0 }];
      if (
        typeof entry === 'object' &&
        entry !== null &&
        isOutcomeCode((entry as { outcome?: unknown }).outcome) &&
        typeof (entry as { netChange?: unknown }).netChange === 'number' &&
        Number.isFinite((entry as { netChange: number }).netChange)
      )
        return [
          {
            outcome: (entry as { outcome: CasinoWarOutcomeCode }).outcome,
            netChange: (entry as { netChange: number }).netChange,
          },
        ];
      return [];
    });
  } catch {
    return [];
  }
}

/** Folds a history slice into win / loss / tie counts. */
export function tallyCasinoWarHistory(history: readonly CasinoWarOutcomeCode[]): CasinoWarTally {
  const tally: CasinoWarTally = { wins: 0, losses: 0, ties: 0 };
  for (const code of history) {
    if (code === CasinoWarOutcome.WIN) tally.wins += 1;
    else if (code === CasinoWarOutcome.LOSS) tally.losses += 1;
    else tally.ties += 1;
  }
  return tally;
}

/**
 * Hook that persists the Casino War round-outcome history in localStorage.
 * `recordOutcome` appends one finished round (the caller is responsible for
 * recording each round only once), capping the stored history at
 * `CASINOWAR_HISTORY_MAX`. `clearHistory` empties it. `tally` is the derived
 * win / loss / tie counts over the full history.
 */
export function useCasinoWarStats() {
  const [rounds, setRounds] = useState<CasinoWarRound[]>(readCasinoWarRounds);
  const history = useMemo(() => rounds.map(({ outcome }) => outcome), [rounds]);

  const recordOutcome = useCallback((outcome: CasinoWarOutcomeCode, netChange = 0) => {
    setRounds((prev) => {
      const next = [...prev, { outcome, netChange }].slice(-CASINOWAR_HISTORY_MAX);
      try {
        localStorage.setItem(CASINOWAR_HISTORY_KEY, JSON.stringify(next));
      } catch {
        /* storage unavailable / quota exceeded */
      }
      return next;
    });
  }, []);

  const clearHistory = useCallback(() => {
    setRounds([]);
    try {
      localStorage.removeItem(CASINOWAR_HISTORY_KEY);
    } catch {
      /* storage unavailable */
    }
  }, []);

  const tally = useMemo(() => tallyCasinoWarHistory(history), [history]);
  const cumulativeNetChange = useMemo(() => rounds.reduce((sum, round) => sum + round.netChange, 0), [rounds]);

  return { history, tally, cumulativeNetChange, recordOutcome, clearHistory };
}
