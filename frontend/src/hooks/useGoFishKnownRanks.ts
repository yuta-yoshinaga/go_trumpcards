import { useEffect, useMemo, useRef, useState } from 'react';
import type { GoFishResponse } from '../types/card';

/** Map keyed by player id with the set of ranks that player is known to (have) hold. */
export type KnownRanksMap = Record<number, number[]>;

/**
 * Track the ranks each player is publicly known to hold based on their past asks.
 *
 * Why this exists: Go Fish forces players to mentally remember every "Do you have
 * a 7?" question to play strategically. Hard-difficulty CPUs read the same
 * action log instantly; humans should not be at an information disadvantage.
 * This hook accumulates `lastAsk` rolls across rerenders and removes ranks once
 * a player has booked them (they no longer hold any of that rank).
 */
export function useGoFishKnownRanks(state: GoFishResponse | null): KnownRanksMap {
  // The server keeps this itself and the CUI has always read it from there. When
  // it is present, use it: the accumulation below lives in component state, so a
  // reload or a remount used to erase the memory mid-game (#6312).
  const fromServer = useMemo<KnownRanksMap | null>(() => {
    if (!state?.players?.some((p) => p.knownRanks !== undefined)) return null;
    const out: KnownRanksMap = {};
    for (const p of state.players) out[p.id] = p.knownRanks ?? [];
    return out;
  }, [state?.players]);

  const [knownByPlayer, setKnownByPlayer] = useState<Record<number, Set<number>>>({});
  const lastAskKey = useRef<string | null>(null);
  const lastTurn = useRef<number>(0);

  useEffect(() => {
    if (!state) return;
    if (state.turnNumber === 1 && lastTurn.current !== 1) {
      setKnownByPlayer({});
      lastAskKey.current = null;
    }
    lastTurn.current = state.turnNumber;
    if (state.lastAsk) {
      const key = `${state.turnNumber}-${state.lastAsk.playerIdx}-${state.lastAsk.targetIdx}-${state.lastAsk.rank}`;
      if (key !== lastAskKey.current) {
        lastAskKey.current = key;
        const askerId = state.lastAsk.playerIdx;
        const rank = state.lastAsk.rank;
        setKnownByPlayer((prev) => {
          const askerSet = new Set(prev[askerId] ?? []);
          askerSet.add(rank);
          return { ...prev, [askerId]: askerSet };
        });
      }
    }
  }, [state]);

  const rebuilt = useMemo(() => {
    if (!state) return {};
    const out: KnownRanksMap = {};
    for (const p of state.players) {
      const known = knownByPlayer[p.id];
      if (!known || known.size === 0) {
        out[p.id] = [];
        continue;
      }
      const bookedRanks = new Set(p.books.map((b) => b.rank));
      out[p.id] = Array.from(known)
        .filter((r) => !bookedRanks.has(r))
        .sort((a, b) => a - b);
    }
    return out;
  }, [knownByPlayer, state]);

  return fromServer ?? rebuilt;
}
