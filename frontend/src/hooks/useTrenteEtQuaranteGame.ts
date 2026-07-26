import { useCallback, useState } from 'react';
import { trenteetquaranteApi } from '../api/gameApi';
import { TrenteEtQuaranteBetType, TrenteEtQuarantePhase } from '../types/phases';
import { useGameApi } from './useGameApi';

/** A previously placed bet, remembered so it can be replayed in the next round. */
export interface LastBet {
  bet: number;
  stake: number;
}

/**
 * Hook that manages Trente et Quarante (Rouge et Noir) banking-game state,
 * bet selection, and action dispatch.
 *
 * Trente et Quarante has no player card decisions: the player picks one of the
 * four even-money bets (Noir, Rouge, Couleur, Inverse) and a stake, then the
 * dealer deals both rows and resolves the round in a single step. The command
 * set is built directly on {@link useGameApi}. `nextround` starts the next round
 * (chips persist server-side); the `bet` command takes `(bet, stake)`.
 */
export function useTrenteEtQuaranteGame() {
  const [betType, setBetType] = useState<number>(TrenteEtQuaranteBetType.NOIR);
  const [betAmount, setBetAmount] = useState(100);
  const [lastBet, setLastBet] = useState<LastBet | null>(null);

  const { state, loading, error, exec: execApi, retry } = useGameApi(trenteetquaranteApi.exec);

  const isBetPhase = state?.phase === TrenteEtQuarantePhase.BET;
  const isResultPhase = state?.phase === TrenteEtQuarantePhase.RESULT;

  /** Places the currently selected bet, which deals both rows and resolves. */
  const handleBet = useCallback(() => {
    setLastBet({ bet: betType, stake: betAmount });
    execApi('bet', betType, betAmount);
  }, [execApi, betType, betAmount]);

  /** Starts the next round (chips carry over on the server). */
  const handleNextRound = useCallback(() => execApi('nextround'), [execApi]);

  /** Whether the last bet can be replayed given the current chip stack. */
  const canRebet = lastBet !== null && lastBet.stake > 0 && state !== null && lastBet.stake <= state.chips;

  /** Replays the previous bet: advances to a fresh round, then re-bets. */
  const handleRebet = useCallback(async () => {
    if (lastBet === null) return;
    await execApi('nextround');
    await execApi('bet', lastBet.bet, lastBet.stake);
  }, [execApi, lastBet]);

  return {
    state,
    loading,
    error,
    retry,
    execApi,
    betType,
    setBetType,
    betAmount,
    setBetAmount,
    lastBet,
    isBetPhase,
    isResultPhase,
    canRebet,
    handleBet,
    handleNextRound,
    handleRebet,
  };
}
