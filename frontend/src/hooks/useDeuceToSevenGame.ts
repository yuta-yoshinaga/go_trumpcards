import { useCallback, useEffect } from 'react';
import { deuceToSevenApi } from '../api/gameApi';
import { DeuceToSevenPhase } from '../types/phases';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';

/** State and actions for the 2-7 Triple Draw page. Mirrors useBadugiGame: it
 * tracks the draw counter (0..3) via state.drawIndex and exposes card-selection
 * helpers for the draw phase. */
export function useDeuceToSevenGame() {
  const { selected, clear: clearSelection, toggle: toggleCard } = useCardSelection();

  const onSuccess = useCallback(() => clearSelection(), [clearSelection]);
  const gameApi = useGameApi(deuceToSevenApi.exec, { onSuccess });
  const { state, loading, error, exec: execAction, retry } = gameApi;

  useEffect(() => {
    execAction('reset');
  }, [execAction]);

  const phase = state?.phase ?? DeuceToSevenPhase.INIT;
  const humanPlayer = state?.players?.find((p) => p.isHuman);
  const canExchange = phase === DeuceToSevenPhase.DRAW && state?.currentTurn === humanPlayer?.id;

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
