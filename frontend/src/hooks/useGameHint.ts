import { useMemo } from 'react';
import type { BlackJackResponse, HeartsResponse, PokerResponse, SpadesResponse } from '../types/card';
import type { HintResult } from '../types/hint';
import { getBlackjackHint } from '../utils/hints/blackjackHint';
import { getHeartsHint } from '../utils/hints/heartsHint';
import { getPokerHint } from '../utils/hints/pokerHint';
import { getSpadesHint } from '../utils/hints/spadesHint';
import { useLocalStorageToggle } from './useLocalStorageToggle';

/** Supported game names for the hint system. */
type HintGameName = 'blackjack' | 'poker' | 'hearts' | 'spades';

/** Return type of the useGameHint hook. */
export interface UseGameHintReturn {
  /** Whether hints are enabled by the user. */
  hintEnabled: boolean;
  /** Toggle hint display on/off. */
  setHintEnabled: (value: boolean) => void;
  /** The current hint result, or null if no hint available. */
  hint: HintResult | null;
}

/** Provides hint state and computation for a specific game. */
export function useGameHint(gameName: HintGameName, state: unknown): UseGameHintReturn {
  const [hintEnabled, setHintEnabled] = useLocalStorageToggle(`hint_enabled_${gameName}`, false);

  const hint = useMemo(() => {
    if (!hintEnabled || !state) return null;
    switch (gameName) {
      case 'blackjack':
        return getBlackjackHint(state as BlackJackResponse);
      case 'poker':
        return getPokerHint(state as PokerResponse);
      case 'hearts':
        return getHeartsHint(state as HeartsResponse);
      case 'spades':
        return getSpadesHint(state as SpadesResponse);
      default:
        return null;
    }
  }, [gameName, hintEnabled, state]);

  return { hintEnabled, setHintEnabled, hint };
}
