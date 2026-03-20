import { useCallback, useEffect, useRef, useState } from 'react';
import { pokerApi } from '../api/gameApi';
import type { PokerOdds } from '../types/card';
import { PokerPhase } from '../types/phases';
import { toggleArrayItem } from '../utils/arrayUtils';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';

/** Hook that manages Poker game state, card exchange, and odds calculation. */
export function usePokerGame() {
  const { selected, setSelected, clear: clearSelection } = useCardSelection();
  const [odds, setOdds] = useState<PokerOdds[] | null>(null);
  const oddsTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const oddsGenRef = useRef(0);
  const mountedRef = useRef(true);

  const onSuccess = useCallback(() => {
    clearSelection();
    setOdds(null);
    oddsGenRef.current++;
  }, [clearSelection]);

  const { state, loading, error, exec } = useGameApi(pokerApi.exec, { onSuccess });

  useEffect(() => {
    exec('reset');
  }, [exec]);

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
      if (oddsTimerRef.current !== null) clearTimeout(oddsTimerRef.current);
    };
  }, []);

  const phase = state?.phase ?? PokerPhase.INIT;
  const humanPlayer = state?.players?.find((p) => p.isHuman);
  const canExchange = phase === PokerPhase.EXCHANGE && state?.currentTurn === humanPlayer?.id;

  const canExchangeRef = useRef(canExchange);
  canExchangeRef.current = canExchange;

  const toggleCard = useCallback(
    (idx: number) => {
      if (!canExchangeRef.current) return;
      setSelected((prev) => {
        const next = toggleArrayItem(prev, idx);
        if (oddsTimerRef.current !== null) clearTimeout(oddsTimerRef.current);
        if (next.length === 0) {
          setOdds(null);
        } else {
          oddsTimerRef.current = setTimeout(() => {
            if (!mountedRef.current) return;
            const gen = ++oddsGenRef.current;
            pokerApi
              .exec('odds', next)
              .then((res) => {
                if (gen === oddsGenRef.current) setOdds(res.odds ?? null);
              })
              .catch((err) => console.error('Failed to fetch poker odds:', err));
          }, 300);
        }
        return next;
      });
    },
    [setSelected],
  );

  return {
    state,
    loading,
    error,
    exec,
    selected,
    toggleCard,
    clearSelection,
    odds,
    canExchange,
  };
}
