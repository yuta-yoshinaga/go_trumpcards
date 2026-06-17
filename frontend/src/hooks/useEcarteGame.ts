import { useCallback } from 'react';
import { type EcarteConfigInput, ecarteApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Écarté game configuration. */
export const DEFAULT_ECARTE_CONFIG: Required<EcarteConfigInput> = {
  cpuDifficulty: 1,
  targetScore: 5,
};

/** CPU difficulty level options for Écarté. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target-score options for Écarté (first player to reach wins). */
export const TARGET_SCORE_OPTIONS = [3, 5, 7] as const;

/**
 * Hook that manages Écarté game state: the Exchange-phase negotiation
 * (propose/stand, accept/refuse, discard) sub-steps, the Play action, and deal
 * advancement.
 *
 * Écarté is a 2-player French 32-card trick game. Before play, the elder
 * (non-dealer) chooses Propose or Stand; if proposed, the dealer Accepts or
 * Refuses; on accept, each player discards any number of cards and draws
 * replacements (the multi-selected hand indices), repeating until the stock
 * empties. Play is 5 strict must-follow tricks (rank K>Q>J>A>10>9>8>7). Scores
 * accumulate across deals to a target (default 5); the higher total wins.
 */
export function useEcarteGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<EcarteConfigInput>>(DEFAULT_ECARTE_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(ecarteApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Plays the single currently-selected card in the Play phase. */
  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    void exec('play', { cardIndex: selectedCardIndices[0] });
  }, [exec, selectedCardIndices]);

  /** Elder proposes an exchange (ElderDecide sub-step). */
  const handlePropose = useCallback(() => {
    void exec('propose');
  }, [exec]);

  /** Elder stands and proceeds to play (ElderDecide sub-step). */
  const handleStand = useCallback(() => {
    void exec('stand');
  }, [exec]);

  /** Dealer responds to a proposal: accept (true) or refuse (false). */
  const handleRespond = useCallback(
    (accept: boolean) => {
      void exec('respond', { accept });
    },
    [exec],
  );

  /** Discards the currently-selected cards and draws replacements. */
  const handleDiscard = useCallback(() => {
    void exec('discard', { discardIndices: selectedCardIndices });
  }, [exec, selectedCardIndices]);

  /** Advances to the next deal after a round (deal) ends. */
  const handleNextRound = useCallback(() => {
    void exec('next');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    ecarteConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handlePlay,
    handlePropose,
    handleStand,
    handleRespond,
    handleDiscard,
    handleNextRound,
  };
}
