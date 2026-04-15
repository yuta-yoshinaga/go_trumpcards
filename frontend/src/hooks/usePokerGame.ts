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
  const [oddsError, setOddsError] = useState<string | null>(null);
  const oddsTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const oddsGenRef = useRef(0);
  const mountedRef = useRef(true);

  const onSuccess = useCallback(() => {
    clearSelection();
    setOdds(null);
    setOddsError(null);
    oddsGenRef.current++;
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(pokerApi.exec, { onSuccess });

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

  const fetchOdds = useCallback((cardIndices: number[]) => {
    if (oddsTimerRef.current !== null) clearTimeout(oddsTimerRef.current);
    if (cardIndices.length === 0) {
      setOdds(null);
      setOddsError(null);
      return;
    }
    oddsTimerRef.current = setTimeout(() => {
      if (!mountedRef.current) return;
      const gen = ++oddsGenRef.current;
      pokerApi
        .exec('odds', cardIndices)
        .then((res) => {
          if (gen !== oddsGenRef.current) return;
          setOdds(res.odds ?? null);
          setOddsError(null);
        })
        .catch((err) => {
          if (gen !== oddsGenRef.current) return;
          setOddsError('oddsFetchFailed');
          console.error('Poker odds fetch error:', err);
        });
    }, 300);
  }, []);

  const toggleCard = useCallback(
    (idx: number) => {
      if (!canExchangeRef.current) return;
      setSelected((prev) => {
        const next = toggleArrayItem(prev, idx);
        fetchOdds(next);
        return next;
      });
    },
    [setSelected, fetchOdds],
  );

  const retryOdds = useCallback(() => {
    fetchOdds(selected);
  }, [fetchOdds, selected]);

  return {
    state,
    loading,
    error,
    exec,
    selected,
    toggleCard,
    clearSelection,
    odds,
    oddsError,
    retryOdds,
    canExchange,
    retry,
  };
}
