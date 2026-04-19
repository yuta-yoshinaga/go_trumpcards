import { useCallback, useEffect } from 'react';
import { badugiApi } from '../api/gameApi';
import { BadugiPhase } from '../types/phases';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';

/** State and actions for the Badugi page. Mirrors usePokerGame but without
 * draw-odds (Badugi's evaluator has no meaningful single-draw probability
 * table) and tracks the draw counter (0..3) via state.drawIndex. */
export function useBadugiGame() {
  const { selected, clear: clearSelection, toggle: toggleCard } = useCardSelection();

  const onSuccess = useCallback(() => clearSelection(), [clearSelection]);
  const gameApi = useGameApi(badugiApi.exec, { onSuccess });
  const { state, loading, error, exec: execAction, retry } = gameApi;

  useEffect(() => {
    execAction('reset');
  }, [execAction]);

  const phase = state?.phase ?? BadugiPhase.INIT;
  const humanPlayer = state?.players?.find((p) => p.isHuman);
  const canExchange = phase === BadugiPhase.DRAW && state?.currentTurn === humanPlayer?.id;

  return {
    state,
    loading,
    error,
    exec: execAction,
    retry,
    selected,
    toggleCard,
    clearSelection,
    canExchange,
  };
}
