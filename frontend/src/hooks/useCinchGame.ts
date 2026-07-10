import { useCallback } from 'react';
import { type CinchConfigInput, cinchApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Cinch (Double Pedro) game configuration. */
export const DEFAULT_CINCH_CONFIG: Required<CinchConfigInput> = {
  cpuDifficulty: 1,
  pointLimit: 21,
};

/** CPU difficulty level options for Cinch. */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available target match-point options for Cinch (first player to reach wins). */
export const POINT_LIMIT_OPTIONS = [11, 21, 31] as const;

/**
 * Hook that manages Cinch (Double Pedro / High Five) game state and its player
 * actions: bid (0=pass, 1-14), name trump (1-4), play a card, plus deal
 * advancement.
 *
 * Cinch is a 4-player 52-card All-Fours/Pitch-family bidding trick-taker. A Bid
 * phase decides the bidder, who then names trump and leads. There are 14 points
 * per deal; the bidding side must make its bid or is set back. First to the
 * target score (default 21) wins. The command set is built directly on
 * {@link useGameApi}.
 */
export function useCinchGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<CinchConfigInput>>(DEFAULT_CINCH_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(cinchApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Declares a bid in the Bid phase (0=pass, 1-14). */
  const handleBid = useCallback(
    (bid: number) => {
      void exec('bid', { bid });
    },
    [exec],
  );

  /** Names the trump suit (1=♠ 2=♣ 3=♥ 4=♦) after winning the bid. */
  const handleNameTrump = useCallback(
    (trumpSuit: number) => {
      void exec('trump', { trumpSuit });
    },
    [exec],
  );

  /** Plays the single currently-selected card in the Play phase. */
  const handlePlay = useCallback(() => {
    if (selectedCardIndices.length !== 1) return;
    void exec('play', { cardIndex: selectedCardIndices[0] });
  }, [exec, selectedCardIndices]);

  /** Advances to the next deal. */
  const handleNextDeal = useCallback(() => {
    void exec('nextround');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    cinchConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handleBid,
    handleNameTrump,
    handlePlay,
    handleNextDeal,
  };
}
