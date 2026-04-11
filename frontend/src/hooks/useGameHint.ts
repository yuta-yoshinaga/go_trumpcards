import { useMemo } from 'react';
import type {
  BaccaratResponse,
  BlackJackResponse,
  CrazyEightsResponse,
  CribbageResponse,
  DaifugoResponse,
  DoubtResponse,
  EuchreResponse,
  FreeCellResponse,
  GinRummyResponse,
  GoFishResponse,
  HeartsResponse,
  HoldemResponse,
  IndianPokerResponse,
  KlondikeResponse,
  MemoryResponse,
  NapoleonResponse,
  OhHellResponse,
  OldMaidResponse,
  OmahaResponse,
  PineappleResponse,
  PokerResponse,
  PyramidResponse,
  SevensResponse,
  ShortDeckResponse,
  SpadesResponse,
  SpeedResponse,
  SpiderResponse,
  ThreeCardResponse,
  TriPeaksResponse,
  VideoPokerResponse,
} from '../types/card';
import type { HintResult } from '../types/hint';
import { getBaccaratHint } from '../utils/hints/baccaratHint';
import { getBlackjackHint } from '../utils/hints/blackjackHint';
import { getCrazyEightsHint } from '../utils/hints/crazyeightsHint';
import { getCribbageHint } from '../utils/hints/cribbageHint';
import { getDaifugoHint } from '../utils/hints/daifugoHint';
import { getDeucesWildHint } from '../utils/hints/deuceswildHint';
import { getDoubtHint } from '../utils/hints/doubtHint';
import { getEuchreHint } from '../utils/hints/euchreHint';
import { getFreeCellHint } from '../utils/hints/freecellHint';
import { getGinRummyHint } from '../utils/hints/ginrummyHint';
import { getGoFishHint } from '../utils/hints/gofishHint';
import { getHeartsHint } from '../utils/hints/heartsHint';
import { getHoldemHint } from '../utils/hints/holdemHint';
import { getIndianPokerHint } from '../utils/hints/indianpokerHint';
import { getJokerPokerHint } from '../utils/hints/jokerpokerHint';
import { getKlondikeHint } from '../utils/hints/klondikeHint';
import { getMemoryHint } from '../utils/hints/memoryHint';
import { getNapoleonHint } from '../utils/hints/napoleonHint';
import { getOhHellHint } from '../utils/hints/ohhellHint';
import { getOldMaidHint } from '../utils/hints/oldmaidHint';
import { getOmahaHint } from '../utils/hints/omahaHint';
import { getPineappleHint } from '../utils/hints/pineappleHint';
import { getPokerHint } from '../utils/hints/pokerHint';
import { getPyramidHint } from '../utils/hints/pyramidHint';
import { getSevensHint } from '../utils/hints/sevensHint';
import { getShortDeckHint } from '../utils/hints/shortdeckHint';
import { getSpadesHint } from '../utils/hints/spadesHint';
import { getSpeedHint } from '../utils/hints/speedHint';
import { getSpiderHint } from '../utils/hints/spiderHint';
import { getThreeCardHint } from '../utils/hints/threecardHint';
import { getTriPeaksHint } from '../utils/hints/tripeaksHint';
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
  | 'ohhell'
  | 'oldmaid'
  | 'doubt'
  | 'daifugo'
  | 'sevens'
  | 'crazyeights'
  | 'speed'
  | 'ginrummy'
  | 'cribbage'
  | 'klondike'
  | 'freecell'
  | 'spider'
  | 'pyramid'
  | 'tripeaks'
  | 'memory'
  | 'sevencardstud'
  | 'fortythieves'
  | 'paigow'
  | 'gofish';

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
      case 'oldmaid':
        return getOldMaidHint(state as OldMaidResponse);
      case 'doubt':
        return getDoubtHint(state as DoubtResponse);
      case 'daifugo':
        return getDaifugoHint(state as DaifugoResponse);
      case 'sevens':
        return getSevensHint(state as SevensResponse);
      case 'crazyeights':
        return getCrazyEightsHint(state as CrazyEightsResponse);
      case 'speed':
        return getSpeedHint(state as SpeedResponse);
      case 'klondike':
        return getKlondikeHint(state as KlondikeResponse);
      case 'freecell':
        return getFreeCellHint(state as FreeCellResponse);
      case 'spider':
        return getSpiderHint(state as SpiderResponse);
      case 'pyramid':
        return getPyramidHint(state as PyramidResponse);
      case 'tripeaks':
        return getTriPeaksHint(state as TriPeaksResponse);
      case 'memory':
        return getMemoryHint(state as MemoryResponse);
      case 'ginrummy':
        return getGinRummyHint(state as GinRummyResponse);
      case 'cribbage':
        return getCribbageHint(state as CribbageResponse);
      case 'gofish':
        return getGoFishHint(state as GoFishResponse);
      default:
        return null;
    }
  }, [gameName, hintEnabled, state]);

  return { hintEnabled, setHintEnabled, hint };
}
