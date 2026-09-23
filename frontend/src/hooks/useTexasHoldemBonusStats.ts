import { useCallback, useMemo, useState } from 'react';

/** localStorage key for Texas Hold'em Bonus session round history. */
export const TEXASHOLDEMBONUS_HISTORY_KEY = 'trumpcards-texasholdembonus-history';

/** Maximum number of recent Texas Hold'em Bonus rounds retained. */
export const TEXASHOLDEMBONUS_HISTORY_MAX = 200;

/** Numeric outcome codes for a completed Texas Hold'em Bonus round. */
export const TexasHoldemBonusOutcome = { PUSH: 0, WIN: 1, LOSS: 2 } as const;

/** A Texas Hold'em Bonus session round outcome and net chip change. */
export interface TexasHoldemBonusRecord {
  outcome: (typeof TexasHoldemBonusOutcome)[keyof typeof TexasHoldemBonusOutcome];
  net: number;
}

/** Aggregated Texas Hold'em Bonus session totals. */
export interface TexasHoldemBonusTally {
  wins: number;
  losses: number;
  pushes: number;
  hands: number;
  net: number;
}

/** Maps a signed game result to a session outcome. */
export function outcomeFromTexasHoldemBonusResult(result: number): TexasHoldemBonusRecord['outcome'] {
  if (result > 0) return TexasHoldemBonusOutcome.WIN;
  if (result < 0) return TexasHoldemBonusOutcome.LOSS;
  return TexasHoldemBonusOutcome.PUSH;
}

function isRecord(value: unknown): value is TexasHoldemBonusRecord {
  if (typeof value !== 'object' || value === null) return false;
  const record = value as Record<string, unknown>;
  return (
    (record.outcome === 0 || record.outcome === 1 || record.outcome === 2) &&
    typeof record.net === 'number' &&
    Number.isFinite(record.net)
  );
}

/** Reads valid Texas Hold'em Bonus session records from localStorage. */
export function readTexasHoldemBonusHistory(): TexasHoldemBonusRecord[] {
  try {
    const raw = localStorage.getItem(TEXASHOLDEMBONUS_HISTORY_KEY);
    if (!raw) return [];
    const parsed: unknown = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed.filter(isRecord) : [];
  } catch {
    return [];
  }
}

/** Aggregates Texas Hold'em Bonus wins, losses, pushes, hands, and net chips. */
export function tallyTexasHoldemBonusHistory(history: readonly TexasHoldemBonusRecord[]): TexasHoldemBonusTally {
  const tally: TexasHoldemBonusTally = { wins: 0, losses: 0, pushes: 0, hands: 0, net: 0 };
  for (const record of history) {
    if (record.outcome === TexasHoldemBonusOutcome.WIN) tally.wins += 1;
    else if (record.outcome === TexasHoldemBonusOutcome.LOSS) tally.losses += 1;
    else tally.pushes += 1;
    tally.hands += 1;
    tally.net += record.net;
  }
  return tally;
}

/** Provides persistent session statistics for Texas Hold'em Bonus rounds. */
export function useTexasHoldemBonusStats() {
  const [history, setHistory] = useState<TexasHoldemBonusRecord[]>(readTexasHoldemBonusHistory);
  const recordRound = useCallback((record: TexasHoldemBonusRecord) => {
    setHistory((previous) => {
      const next = [...previous, record].slice(-TEXASHOLDEMBONUS_HISTORY_MAX);
      try {
        localStorage.setItem(TEXASHOLDEMBONUS_HISTORY_KEY, JSON.stringify(next));
      } catch {
        /* storage unavailable / quota exceeded */
      }
      return next;
    });
  }, []);
  const clearHistory = useCallback(() => {
    setHistory([]);
    try {
      localStorage.removeItem(TEXASHOLDEMBONUS_HISTORY_KEY);
    } catch {
      /* storage unavailable */
    }
  }, []);
  const tally = useMemo(() => tallyTexasHoldemBonusHistory(history), [history]);
  return { history, tally, recordRound, clearHistory };
}
