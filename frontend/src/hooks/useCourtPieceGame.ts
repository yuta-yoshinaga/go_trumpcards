import { useCallback } from 'react';
import { type CourtPieceConfigInput, courtPieceApi } from '../api/gameApi';
import { useCardSelection } from './useCardSelection';
import { useGameApi } from './useGameApi';
import { useGameConfig } from './useGameConfig';

/** Default Court Piece (Rang) game configuration. */
export const DEFAULT_COURT_PIECE_CONFIG: Required<CourtPieceConfigInput> = {
  cpuDifficulty: 1,
  pointLimit: 7,
};

/** CPU difficulty level options for Court Piece (Rang). */
export const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

/** Available point-limit options for Court Piece (Rang) (first team to reach wins). */
export const POINT_LIMIT_OPTIONS = [5, 7, 9] as const;

/** Trump suit options for Court Piece (Rang) declaration (1=♠ 2=♣ 3=♥ 4=♦). */
export const TRUMP_SUIT_OPTIONS: { value: number; key: string }[] = [
  { value: 1, key: 'suit.spade' },
  { value: 2, key: 'suit.club' },
  { value: 3, key: 'suit.heart' },
  { value: 4, key: 'suit.diamond' },
];

/**
 * Hook that manages Court Piece (Rang) game state, the trump-declaration
 * action, the play action, and trick/round advancement.
 *
 * Court Piece is a South-Asian 4-player, 2-team (seats 0&2 vs 1&3) trick-taker
 * with no numeric bidding. The caller (Hakim) peeks at the first 5 cards and
 * declares a trump suit; the teams then play 13 tricks. A team taking 7+ of the
 * 13 tricks wins the round (Sar = +1 point); sweeping all tricks or winning
 * consecutive rounds adds a Court bonus (+2). The first team to reach the point
 * limit (default 7) wins.
 */
export function useCourtPieceGame() {
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const { config, handleConfigChange } = useGameConfig<Required<CourtPieceConfigInput>>(DEFAULT_COURT_PIECE_CONFIG);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);

  const { state, loading, error, exec, retry } = useGameApi(courtPieceApi.exec, { onSuccess });

  /** Resets the game, applying the current config. */
  const reset = useCallback(() => {
    void exec('reset', { config });
  }, [exec, config]);

  /** Declares the trump suit (1=♠ 2=♣ 3=♥ 4=♦) during the TrumpDeclaration phase. */
  const handleDeclareTrump = useCallback(
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

  /** Advances to the next trick. */
  const handleNextTrick = useCallback(() => {
    void exec('next');
  }, [exec]);

  /** Advances to the next round. */
  const handleNextRound = useCallback(() => {
    void exec('nextround');
  }, [exec]);

  return {
    state,
    loading,
    error,
    exec,
    retry,
    courtPieceConfig: config,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handleDeclareTrump,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  };
}
