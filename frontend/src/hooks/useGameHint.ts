import { useMemo } from 'react';
import type {
  BaccaratResponse,
  BlackJackResponse,
  CanastaResponse,
  CanfieldResponse,
  CaribbeanStudResponse,
  CrazyEightsResponse,
  CribbageResponse,
  DaifugoResponse,
  DoubtResponse,
  DurakResponse,
  EuchreResponse,
  FiftyOneResponse,
  FreeCellResponse,
  GinRummyResponse,
  GoFishResponse,
  HeartsResponse,
  HoldemResponse,
  IndianPokerResponse,
  KlondikeResponse,
  LetItRideResponse,
  MemoryResponse,
  NapoleonResponse,
  OhHellResponse,
  OldMaidResponse,
  OmahaResponse,
  PageOneResponse,
  PineappleResponse,
  PinochleResponse,
  PokerResponse,
  PokerSquaresResponse,
  PyramidResponse,
  RedDogResponse,
  ScorpionResponse,
  SevenCardStudResponse,
  SevensResponse,
  ShortDeckResponse,
  SpadesResponse,
  SpeedResponse,
  SpiderResponse,
  ThreeCardResponse,
  TriPeaksResponse,
  TwoTenJackResponse,
  VideoPokerResponse,
  WarResponse,
  WhistResponse,
  YukonResponse,
} from '../types/card';
import type { HintResult } from '../types/hint';
import { getBaccaratHint } from '../utils/hints/baccaratHint';
import { getBlackjackHint } from '../utils/hints/blackjackHint';
import { getCanastaHint } from '../utils/hints/canastaHint';
import { getCanfieldHint } from '../utils/hints/canfieldHint';
import { getCaribbeanStudHint } from '../utils/hints/caribbeanstudHint';
import { getCrazyEightsHint } from '../utils/hints/crazyeightsHint';
import { getCribbageHint } from '../utils/hints/cribbageHint';
import { getDaifugoHint } from '../utils/hints/daifugoHint';
import { getDeucesWildHint } from '../utils/hints/deuceswildHint';
import { getDoubtHint } from '../utils/hints/doubtHint';
import { getDurakHint } from '../utils/hints/durakHint';
import { getEuchreHint } from '../utils/hints/euchreHint';
import { getFiftyOneHint } from '../utils/hints/fiftyoneHint';
import { getFreeCellHint } from '../utils/hints/freecellHint';
import { getGinRummyHint } from '../utils/hints/ginrummyHint';
import { getGoFishHint } from '../utils/hints/gofishHint';
import { getHeartsHint } from '../utils/hints/heartsHint';
import { getHoldemHint } from '../utils/hints/holdemHint';
import { getIndianPokerHint } from '../utils/hints/indianpokerHint';
import { getJokerPokerHint } from '../utils/hints/jokerpokerHint';
import { getKlondikeHint } from '../utils/hints/klondikeHint';
import { getLetitrideHint } from '../utils/hints/letitrideHint';
import { getMemoryHint } from '../utils/hints/memoryHint';
import { getNapoleonHint } from '../utils/hints/napoleonHint';
import { getOhHellHint } from '../utils/hints/ohhellHint';
import { getOldMaidHint } from '../utils/hints/oldmaidHint';
import { getOmahaHint } from '../utils/hints/omahaHint';
import { getPageOneHint } from '../utils/hints/pageoneHint';
import { getPineappleHint } from '../utils/hints/pineappleHint';
import { getPinochleHint } from '../utils/hints/pinochleHint';
import { getPokerHint } from '../utils/hints/pokerHint';
import { getPokersquaresHint } from '../utils/hints/pokersquaresHint';
import { getPyramidHint } from '../utils/hints/pyramidHint';
import { getRazzHint } from '../utils/hints/razzHint';
import { getReddogHint } from '../utils/hints/reddogHint';
import { getScorpionHint } from '../utils/hints/scorpionHint';
import { getSevensHint } from '../utils/hints/sevensHint';
import { getShortDeckHint } from '../utils/hints/shortdeckHint';
import { getSpadesHint } from '../utils/hints/spadesHint';
import { getSpeedHint } from '../utils/hints/speedHint';
import { getSpiderHint } from '../utils/hints/spiderHint';
import { getThreeCardHint } from '../utils/hints/threecardHint';
import { getTriPeaksHint } from '../utils/hints/tripeaksHint';
import { getTwoTenJackHint } from '../utils/hints/twotenjackHint';
import { getVideoPokerHint } from '../utils/hints/videopokerHint';
import { getWarHint } from '../utils/hints/warHint';
import { getWhistHint } from '../utils/hints/whistHint';
import { getYukonHint } from '../utils/hints/yukonHint';
import { useLocalStorageToggle } from './useLocalStorageToggle';

/** Hint function that takes game state and returns a hint result or null. */
type HintFn = (state: unknown) => HintResult | null;

/** Registry mapping game names to their hint functions. */
const hintFactories = {
  baccarat: (s) => getBaccaratHint(s as BaccaratResponse),
  blackjack: (s) => getBlackjackHint(s as BlackJackResponse),
  poker: (s) => getPokerHint(s as PokerResponse),
  hearts: (s) => getHeartsHint(s as HeartsResponse),
  spades: (s) => getSpadesHint(s as SpadesResponse),
  holdem: (s) => getHoldemHint(s as HoldemResponse),
  omaha: (s) => getOmahaHint(s as OmahaResponse),
  shortdeck: (s) => getShortDeckHint(s as ShortDeckResponse),
  pineapple: (s) => getPineappleHint(s as PineappleResponse),
  videopoker: (s) => getVideoPokerHint(s as VideoPokerResponse),
  deuceswild: (s) => getDeucesWildHint(s as VideoPokerResponse),
  jokerpoker: (s) => getJokerPokerHint(s as VideoPokerResponse),
  indianpoker: (s) => getIndianPokerHint(s as IndianPokerResponse),
  threecard: (s) => getThreeCardHint(s as ThreeCardResponse),
  euchre: (s) => getEuchreHint(s as EuchreResponse),
  fiftyone: (s) => getFiftyOneHint(s as FiftyOneResponse),
  napoleon: (s) => getNapoleonHint(s as NapoleonResponse),
  ohhell: (s) => getOhHellHint(s as OhHellResponse),
  oldmaid: (s) => getOldMaidHint(s as OldMaidResponse),
  doubt: (s) => getDoubtHint(s as DoubtResponse),
  daifugo: (s) => getDaifugoHint(s as DaifugoResponse),
  sevens: (s) => getSevensHint(s as SevensResponse),
  crazyeights: (s) => getCrazyEightsHint(s as CrazyEightsResponse),
  speed: (s) => getSpeedHint(s as SpeedResponse),
  klondike: (s) => getKlondikeHint(s as KlondikeResponse),
  freecell: (s) => getFreeCellHint(s as FreeCellResponse),
  spider: (s) => getSpiderHint(s as SpiderResponse),
  pyramid: (s) => getPyramidHint(s as PyramidResponse),
  tripeaks: (s) => getTriPeaksHint(s as TriPeaksResponse),
  memory: (s) => getMemoryHint(s as MemoryResponse),
  ginrummy: (s) => getGinRummyHint(s as GinRummyResponse),
  cribbage: (s) => getCribbageHint(s as CribbageResponse),
  gofish: (s) => getGoFishHint(s as GoFishResponse),
  caribbeanstud: (s) => getCaribbeanStudHint(s as CaribbeanStudResponse),
  durak: (s) => getDurakHint(s as DurakResponse),
  canasta: (s) => getCanastaHint(s as CanastaResponse),
  canfield: (s) => getCanfieldHint(s as CanfieldResponse),
  pinochle: (s) => getPinochleHint(s as PinochleResponse),
  twotenjack: (s) => getTwoTenJackHint(s as TwoTenJackResponse),
  sevencardstud: () => null,
  razz: (s) => getRazzHint(s as SevenCardStudResponse),
  fortythieves: () => null,
  paigow: () => null,
  pageone: (s) => getPageOneHint(s as PageOneResponse),
  pokersquares: (s) => getPokersquaresHint(s as PokerSquaresResponse),
  letitride: (s) => getLetitrideHint(s as LetItRideResponse),
  reddog: (s) => getReddogHint(s as RedDogResponse),
  war: (s) => getWarHint(s as WarResponse),
  whist: (s) => getWhistHint(s as WhistResponse),
  yukon: (s) => getYukonHint(s as YukonResponse),
  scorpion: (s) => getScorpionHint(s as ScorpionResponse),
} satisfies Record<string, HintFn>;

/** Supported game names for the hint system, derived from the registry. */
export type HintGameName = keyof typeof hintFactories;

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
    return hintFactories[gameName]?.(state) ?? null;
  }, [gameName, hintEnabled, state]);

  return { hintEnabled, setHintEnabled, hint };
}
