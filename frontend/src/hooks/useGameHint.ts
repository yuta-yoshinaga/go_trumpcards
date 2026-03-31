import { useMemo } from 'react';
import type {
  BaccaratResponse,
  BlackJackResponse,
  EuchreResponse,
  HeartsResponse,
  HoldemResponse,
  IndianPokerResponse,
  NapoleonResponse,
  OhHellResponse,
  OmahaResponse,
  PineappleResponse,
  PokerResponse,
  ShortDeckResponse,
  SpadesResponse,
  ThreeCardResponse,
  VideoPokerResponse,
} from '../types/card';
import type { HintResult } from '../types/hint';
import { getBaccaratHint } from '../utils/hints/baccaratHint';
import { getBlackjackHint } from '../utils/hints/blackjackHint';
import { getDeucesWildHint } from '../utils/hints/deuceswildHint';
import { getEuchreHint } from '../utils/hints/euchreHint';
import { getHeartsHint } from '../utils/hints/heartsHint';
import { getHoldemHint } from '../utils/hints/holdemHint';
import { getIndianPokerHint } from '../utils/hints/indianpokerHint';
import { getJokerPokerHint } from '../utils/hints/jokerpokerHint';
import { getNapoleonHint } from '../utils/hints/napoleonHint';
import { getOhHellHint } from '../utils/hints/ohhellHint';
import { getOmahaHint } from '../utils/hints/omahaHint';
import { getPineappleHint } from '../utils/hints/pineappleHint';
import { getPokerHint } from '../utils/hints/pokerHint';
import { getShortDeckHint } from '../utils/hints/shortdeckHint';
import { getSpadesHint } from '../utils/hints/spadesHint';
import { getThreeCardHint } from '../utils/hints/threecardHint';
import { getVideoPokerHint } from '../utils/hints/videopokerHint';
import { useLocalStorageToggle } from './useLocalStorageToggle';

/** Supported game names for the hint system. */
type HintGameName =
  | 'baccarat'
  | 'blackjack'
  | 'poker'
  | 'hearts'
  | 'spades'
  | 'holdem'
  | 'omaha'
  | 'shortdeck'
  | 'pineapple'
  | 'videopoker'
  | 'deuceswild'
  | 'jokerpoker'
  | 'indianpoker'
  | 'threecard'
  | 'euchre'
  | 'napoleon'
  | 'ohhell';

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
      case 'baccarat':
        return getBaccaratHint(state as BaccaratResponse);
      case 'blackjack':
        return getBlackjackHint(state as BlackJackResponse);
      case 'poker':
        return getPokerHint(state as PokerResponse);
      case 'hearts':
        return getHeartsHint(state as HeartsResponse);
      case 'spades':
        return getSpadesHint(state as SpadesResponse);
      case 'holdem':
        return getHoldemHint(state as HoldemResponse);
      case 'omaha':
        return getOmahaHint(state as OmahaResponse);
      case 'shortdeck':
        return getShortDeckHint(state as ShortDeckResponse);
      case 'pineapple':
        return getPineappleHint(state as PineappleResponse);
      case 'videopoker':
        return getVideoPokerHint(state as VideoPokerResponse);
      case 'deuceswild':
        return getDeucesWildHint(state as VideoPokerResponse);
      case 'jokerpoker':
        return getJokerPokerHint(state as VideoPokerResponse);
      case 'indianpoker':
        return getIndianPokerHint(state as IndianPokerResponse);
      case 'threecard':
        return getThreeCardHint(state as ThreeCardResponse);
      case 'euchre':
        return getEuchreHint(state as EuchreResponse);
      case 'napoleon':
        return getNapoleonHint(state as NapoleonResponse);
      case 'ohhell':
        return getOhHellHint(state as OhHellResponse);
      default:
        return null;
    }
  }, [gameName, hintEnabled, state]);

  return { hintEnabled, setHintEnabled, hint };
}
