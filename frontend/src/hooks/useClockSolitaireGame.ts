import { useCallback, useEffect } from 'react';
import { clocksolitaireApi } from '../api/gameApi';
import { useGameApi } from './useGameApi';

/** Hook that manages Clock Solitaire game state with step and autoplay actions. */
export function useClockSolitaireGame() {
  const { state, loading, error, exec, retry } = useGameApi(clocksolitaireApi.exec);

  useEffect(() => {
    exec('reset');
  }, [exec]);

  const handleReset = useCallback(() => {
    exec('reset');
  }, [exec]);

  const handleStep = useCallback(() => {
    exec('step');
  }, [exec]);

  const handleAutoPlay = useCallback(() => {
    exec('autoplay');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    handleReset,
    handleStep,
    handleAutoPlay,
  };
}
