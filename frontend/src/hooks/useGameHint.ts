import { useMemo } from 'react';
import type {
  AccordionResponse,
  BaccaratResponse,
  BadugiResponse,
  BakersDozenResponse,
  BarbuResponse,
  BeleagueredCastleResponse,
  BeloteResponse,
  BidWhistResponse,
  BigTwoResponse,
  BlackJackResponse,
  BlackJackSwitchResponse,
  BristolResponse,
  BurracoResponse,
  CalculationResponse,
  CallBreakResponse,
  CanastaResponse,
  CanfieldResponse,
  CaribbeanStudResponse,
  CasinoHoldemResponse,
  CasinoWarResponse,
  CassinoResponse,
  ClockSolitaireResponse,
  CrazyEightsResponse,
  CribbageResponse,
  CruelResponse,
  DaifugoResponse,
  DeuceToSevenResponse,
  DoubtResponse,
  DoudizhuResponse,
  DragonTigerResponse,
  DurakResponse,
  EasthavenResponse,
  EgyptianRatscrewResponse,
  EightOffResponse,
  EuchreResponse,
  FiftyOneResponse,
  FiveHundredResponse,
  FourCardPokerResponse,
  FreeCellResponse,
  GapsResponse,
  GinRummyResponse,
  GoFishResponse,
  GongZhuResponse,
  HeartsResponse,
  HighCardFlushResponse,
  HoldemResponse,
  IndianPokerResponse,
  KlondikeResponse,
  LetItRideResponse,
  MacauResponse,
  MemoryResponse,
  MightyResponse,
  MississippiStudResponse,
  MonteCarloResponse,
  NapoleonResponse,
  NertzResponse,
  OhHellResponse,
  OldMaidResponse,
  OmahaResponse,
  OsmosisResponse,
  PageOneResponse,
  PenguinResponse,
  PigsTailResponse,
  PineappleResponse,
  PinochleResponse,
  PitchResponse,
  PokerResponse,
  PokerSquaresResponse,
  PresidentResponse,
  PyramidResponse,
  RedDogResponse,
  Rummy500Response,
  RussianPokerResponse,
  RussianSolitaireResponse,
  SchnapsenResponse,
  ScopaResponse,
  ScorpionResponse,
  SeahavenTowersResponse,
  SevenBridgeResponse,
  SevenCardStudResponse,
  SevensResponse,
  ShitheadResponse,
  ShortDeckResponse,
  SixCardGolfResponse,
  SkatResponse,
  SlapjackResponse,
  SpadesResponse,
  SpeedResponse,
  SpideretteResponse,
  SpiderResponse,
  SpiteAndMaliceResponse,
  TarneebResponse,
  TexasHoldemBonusResponse,
  ThirtyOneResponse,
  ThreeCardResponse,
  TichuResponse,
  TienLenResponse,
  TrashResponse,
  TressetteResponse,
  TriPeaksResponse,
  TwoTenJackResponse,
  UltimateTexasHoldemResponse,
  VideoPokerResponse,
  WarResponse,
  WaspResponse,
  WhistResponse,
  YanivResponse,
  YukonResponse,
} from '../types/card';
import type { HintResult } from '../types/hint';
import { getAccordionHint } from '../utils/hints/accordionHint';
import { getBaccaratHint } from '../utils/hints/baccaratHint';
import { getBadugiHint } from '../utils/hints/badugiHint';
import { getBakersdozenHint } from '../utils/hints/bakersdozenHint';
import { getBarbuHint } from '../utils/hints/barbuHint';
import { getBeleagueredcastleHint } from '../utils/hints/beleagueredcastleHint';
import { getBeloteHint } from '../utils/hints/beloteHint';
import { getBidWhistHint } from '../utils/hints/bidwhistHint';
import { getBigTwoHint } from '../utils/hints/bigtwoHint';
import { getBlackjackHint } from '../utils/hints/blackjackHint';
import { getBlackjackswitchHint } from '../utils/hints/blackjackswitchHint';
import { getBristolHint } from '../utils/hints/bristolHint';
import { getBurracoHint } from '../utils/hints/burracoHint';
import { getCalculationHint } from '../utils/hints/calculationHint';
import { getCallBreakHint } from '../utils/hints/callbreakHint';
import { getCanastaHint } from '../utils/hints/canastaHint';
import { getCanfieldHint } from '../utils/hints/canfieldHint';
import { getCaribbeanStudHint } from '../utils/hints/caribbeanstudHint';
import { getCasinoHoldemHint } from '../utils/hints/casinoholdemHint';
import { getCasinowarHint } from '../utils/hints/casinowarHint';
import { getCassinoHint } from '../utils/hints/cassinoHint';
import { getClocksolitaireHint } from '../utils/hints/clocksolitaireHint';
import { getCrazyEightsHint } from '../utils/hints/crazyeightsHint';
import { getCrazyPineappleHint } from '../utils/hints/crazyPineappleHint';
import { getCribbageHint } from '../utils/hints/cribbageHint';
import { getCruelHint } from '../utils/hints/cruelHint';
import { getDaifugoHint } from '../utils/hints/daifugoHint';
import { getDeucesWildHint } from '../utils/hints/deuceswildHint';
import { getDeuceToSevenHint } from '../utils/hints/deuceToSevenHint';
import { getDoubtHint } from '../utils/hints/doubtHint';
import { getDoudizhuHint } from '../utils/hints/doudizhuHint';
import { getDragontigerHint } from '../utils/hints/dragontigerHint';
import { getDurakHint } from '../utils/hints/durakHint';
import { getEasthavenHint } from '../utils/hints/easthavenHint';
import { getEgyptianRatscrewHint } from '../utils/hints/egyptianratscrewHint';
import { getEightOffHint } from '../utils/hints/eightoffHint';
import { getEuchreHint } from '../utils/hints/euchreHint';
import { getFiftyOneHint } from '../utils/hints/fiftyoneHint';
import { getFiveHundredHint } from '../utils/hints/fivehundredHint';
import { getFourCardPokerHint } from '../utils/hints/fourcardpokerHint';
import { getFreeCellHint } from '../utils/hints/freecellHint';
import { getGapsHint } from '../utils/hints/gapsHint';
import { getGinRummyHint } from '../utils/hints/ginrummyHint';
import { getGoFishHint } from '../utils/hints/gofishHint';
import { getGongZhuHint } from '../utils/hints/gongzhuHint';
import { getHeartsHint } from '../utils/hints/heartsHint';
import { getHighCardFlushHint } from '../utils/hints/highcardflushHint';
import { getHoldemHint } from '../utils/hints/holdemHint';
import { getIndianPokerHint } from '../utils/hints/indianpokerHint';
import { getIrishPokerHint } from '../utils/hints/irishPokerHint';
import { getJokerPokerHint } from '../utils/hints/jokerpokerHint';
import { getKlondikeHint } from '../utils/hints/klondikeHint';
import { getLetitrideHint } from '../utils/hints/letitrideHint';
import { getMacauHint } from '../utils/hints/macauHint';
import { getMemoryHint } from '../utils/hints/memoryHint';
import { getMightyHint } from '../utils/hints/mightyHint';
import { getMississippiStudHint } from '../utils/hints/mississippiStudHint';
import { getMonteCarloHint } from '../utils/hints/montecarloHint';
import { getNapoleonHint } from '../utils/hints/napoleonHint';
import { getNertzHint } from '../utils/hints/nertzHint';
import { getOhHellHint } from '../utils/hints/ohhellHint';
import { getOldMaidHint } from '../utils/hints/oldmaidHint';
import { getOmahaHiLoHint } from '../utils/hints/omahaHiLoHint';
import { getOmahaHint } from '../utils/hints/omahaHint';
import { getOsmosisHint } from '../utils/hints/osmosisHint';
import { getPageOneHint } from '../utils/hints/pageoneHint';
import { getPenguinHint } from '../utils/hints/penguinHint';
import { getPigstailHint } from '../utils/hints/pigstailHint';
import { getPineappleHint } from '../utils/hints/pineappleHint';
import { getPinochleHint } from '../utils/hints/pinochleHint';
import { getPitchHint } from '../utils/hints/pitchHint';
import { getPokerHint } from '../utils/hints/pokerHint';
import { getPokersquaresHint } from '../utils/hints/pokersquaresHint';
import { getPresidentHint } from '../utils/hints/presidentHint';
import { getPyramidHint } from '../utils/hints/pyramidHint';
import { getRazzHint } from '../utils/hints/razzHint';
import { getReddogHint } from '../utils/hints/reddogHint';
import { getRummy500Hint } from '../utils/hints/rummy500Hint';
import { getRussianPokerHint } from '../utils/hints/russianpokerHint';
import { getRussianSolitaireHint } from '../utils/hints/russianSolitaireHint';
import { getSchnapsenHint } from '../utils/hints/schnapsenHint';
import { getScopaHint } from '../utils/hints/scopaHint';
import { getScorpionHint } from '../utils/hints/scorpionHint';
import { getSeahavenTowersHint } from '../utils/hints/seahavenTowersHint';
import { getSevenbridgeHint } from '../utils/hints/sevenbridgeHint';
import { getSevensHint } from '../utils/hints/sevensHint';
import { getShitheadHint } from '../utils/hints/shitheadHint';
import { getShortDeckHint } from '../utils/hints/shortdeckHint';
import { getSixcardgolfHint } from '../utils/hints/sixcardgolfHint';
import { getSkatHint } from '../utils/hints/skatHint';
import { getSlapjackHint } from '../utils/hints/slapjackHint';
import { getSpadesHint } from '../utils/hints/spadesHint';
import { getSpeedHint } from '../utils/hints/speedHint';
import { getSpideretteHint } from '../utils/hints/spideretteHint';
import { getSpiderHint } from '../utils/hints/spiderHint';
import { getSpiteAndMaliceHint } from '../utils/hints/spiteAndMaliceHint';
import { getTarneebHint } from '../utils/hints/tarneebHint';
import { getTexasHoldemBonusHint } from '../utils/hints/texasHoldemBonusHint';
import { getThirtyOneHint } from '../utils/hints/thirtyoneHint';
import { getThreeCardHint } from '../utils/hints/threecardHint';
import { getTichuHint } from '../utils/hints/tichuHint';
import { getTienLenHint } from '../utils/hints/tienlenHint';
import { getTrashHint } from '../utils/hints/trashHint';
import { getTressetteHint } from '../utils/hints/tressetteHint';
import { getTriPeaksHint } from '../utils/hints/tripeaksHint';
import { getTwoTenJackHint } from '../utils/hints/twotenjackHint';
import { getUltimateTexasHoldemHint } from '../utils/hints/ultimateTexasHoldemHint';
import { getVideoPokerHint } from '../utils/hints/videopokerHint';
import { getWarHint } from '../utils/hints/warHint';
import { getWaspHint } from '../utils/hints/waspHint';
import { getWhistHint } from '../utils/hints/whistHint';
import { getYanivHint } from '../utils/hints/yanivHint';
import { getYukonHint } from '../utils/hints/yukonHint';
import { useLocalStorageToggle } from './useLocalStorageToggle';

/** Hint function that takes game state and returns a hint result or null. */
type HintFn = (state: unknown) => HintResult | null;

/** Registry mapping game names to their hint functions. */
const hintFactories = {
  baccarat: (s) => getBaccaratHint(s as BaccaratResponse),
  blackjack: (s) => getBlackjackHint(s as BlackJackResponse),
  spanish21: (s) => getBlackjackHint(s as BlackJackResponse),
  poker: (s) => getPokerHint(s as PokerResponse),
  hearts: (s) => getHeartsHint(s as HeartsResponse),
  gongzhu: (s) => getGongZhuHint(s as GongZhuResponse),
  spades: (s) => getSpadesHint(s as SpadesResponse),
  callbreak: (s) => getCallBreakHint(s as CallBreakResponse),
  tarneeb: (s) => getTarneebHint(s as TarneebResponse),
  pitch: (s) => getPitchHint(s as PitchResponse),
  holdem: (s) => getHoldemHint(s as HoldemResponse),
  omaha: (s) => getOmahaHint(s as OmahaResponse),
  omahahilo: (s) => getOmahaHiLoHint(s as OmahaResponse),
  bigo: (s) => getOmahaHint(s as OmahaResponse),
  bigohilo: (s) => getOmahaHiLoHint(s as OmahaResponse),
  shortdeck: (s) => getShortDeckHint(s as ShortDeckResponse),
  pineapple: (s) => getPineappleHint(s as PineappleResponse),
  crazypineapple: (s) => getCrazyPineappleHint(s as PineappleResponse),
  irishpoker: (s) => getIrishPokerHint(s as PineappleResponse),
  videopoker: (s) => getVideoPokerHint(s as VideoPokerResponse),
  deuceswild: (s) => getDeucesWildHint(s as VideoPokerResponse),
  jokerpoker: (s) => getJokerPokerHint(s as VideoPokerResponse),
  indianpoker: (s) => getIndianPokerHint(s as IndianPokerResponse),
  threecard: (s) => getThreeCardHint(s as ThreeCardResponse),
  tichu: (s) => getTichuHint(s as TichuResponse),
  highcardflush: (s) => getHighCardFlushHint(s as HighCardFlushResponse),
  euchre: (s) => getEuchreHint(s as EuchreResponse),
  belote: (s) => getBeloteHint(s as BeloteResponse),
  bigtwo: (s) => getBigTwoHint(s as BigTwoResponse),
  tienlen: (s) => getTienLenHint(s as TienLenResponse),
  fivehundred: (s) => getFiveHundredHint(s as FiveHundredResponse),
  schnapsen: (s) => getSchnapsenHint(s as SchnapsenResponse),
  fiftyone: (s) => getFiftyOneHint(s as FiftyOneResponse),
  napoleon: (s) => getNapoleonHint(s as NapoleonResponse),
  mighty: (s) => getMightyHint(s as MightyResponse),
  ohhell: (s) => getOhHellHint(s as OhHellResponse),
  oldmaid: (s) => getOldMaidHint(s as OldMaidResponse),
  doubt: (s) => getDoubtHint(s as DoubtResponse),
  daifugo: (s) => getDaifugoHint(s as DaifugoResponse),
  sevens: (s) => getSevensHint(s as SevensResponse),
  crazyeights: (s) => getCrazyEightsHint(s as CrazyEightsResponse),
  speed: (s) => getSpeedHint(s as SpeedResponse),
  klondike: (s) => getKlondikeHint(s as KlondikeResponse),
  freecell: (s) => getFreeCellHint(s as FreeCellResponse),
  // Baker's Game shares the FreeCell response shape and stacking heuristics.
  bakersgame: (s) => getFreeCellHint(s as FreeCellResponse),
  eightoff: (s) => getEightOffHint(s as EightOffResponse),
  penguin: (s) => getPenguinHint(s as PenguinResponse),
  seahaventowers: (s) => getSeahavenTowersHint(s as SeahavenTowersResponse),
  spider: (s) => getSpiderHint(s as SpiderResponse),
  pyramid: (s) => getPyramidHint(s as PyramidResponse),
  tripeaks: (s) => getTriPeaksHint(s as TriPeaksResponse),
  memory: (s) => getMemoryHint(s as MemoryResponse),
  ginrummy: (s) => getGinRummyHint(s as GinRummyResponse),
  cribbage: (s) => getCribbageHint(s as CribbageResponse),
  gofish: (s) => getGoFishHint(s as GoFishResponse),
  caribbeanstud: (s) => getCaribbeanStudHint(s as CaribbeanStudResponse),
  casinoholdem: (s) => getCasinoHoldemHint(s as CasinoHoldemResponse),
  texasholdembonus: (s) => getTexasHoldemBonusHint(s as TexasHoldemBonusResponse),
  ultimatetexasholdem: (s) => getUltimateTexasHoldemHint(s as UltimateTexasHoldemResponse),
  mississippistud: (s) => getMississippiStudHint(s as MississippiStudResponse),
  durak: (s) => getDurakHint(s as DurakResponse),
  canasta: (s) => getCanastaHint(s as CanastaResponse),
  bristol: (s) => getBristolHint(s as BristolResponse),
  burraco: (s) => getBurracoHint(s as BurracoResponse),
  canfield: (s) => getCanfieldHint(s as CanfieldResponse),
  osmosis: (s) => getOsmosisHint(s as OsmosisResponse),
  pinochle: (s) => getPinochleHint(s as PinochleResponse),
  twotenjack: (s) => getTwoTenJackHint(s as TwoTenJackResponse),
  sevencardstud: () => null,
  razz: (s) => getRazzHint(s as SevenCardStudResponse),
  badugi: (s) => getBadugiHint(s as BadugiResponse),
  deucetoseven: (s) => getDeuceToSevenHint(s as DeuceToSevenResponse),
  fortythieves: () => null,
  bakersdozen: (s) => getBakersdozenHint(s as BakersDozenResponse),
  beleagueredcastle: (s) => getBeleagueredcastleHint(s as BeleagueredCastleResponse),
  tonk: () => null,
  thirtyone: (s) => getThirtyOneHint(s as ThirtyOneResponse),
  yaniv: (s) => getYanivHint(s as YanivResponse),
  tressette: (s) => getTressetteHint(s as TressetteResponse),
  paigow: () => null,
  chinesepoker: () => null,
  pageone: (s) => getPageOneHint(s as PageOneResponse),
  pigtail: (s) => getPigstailHint(s as PigsTailResponse),
  pokersquares: (s) => getPokersquaresHint(s as PokerSquaresResponse),
  montecarlo: (s) => getMonteCarloHint(s as MonteCarloResponse),
  letitride: (s) => getLetitrideHint(s as LetItRideResponse),
  reddog: (s) => getReddogHint(s as RedDogResponse),
  casinowar: (s) => getCasinowarHint(s as CasinoWarResponse),
  dragontiger: (s) => getDragontigerHint(s as DragonTigerResponse),
  blackjackswitch: (s) => getBlackjackswitchHint(s as BlackJackSwitchResponse),
  war: (s) => getWarHint(s as WarResponse),
  whist: (s) => getWhistHint(s as WhistResponse),
  yukon: (s) => getYukonHint(s as YukonResponse),
  russiansolitaire: (s) => getRussianSolitaireHint(s as RussianSolitaireResponse),
  cruel: (s) => getCruelHint(s as CruelResponse),
  scorpion: (s) => getScorpionHint(s as ScorpionResponse),
  wasp: (s) => getWaspHint(s as WaspResponse),
  easthaven: (s) => getEasthavenHint(s as EasthavenResponse),
  accordion: (s) => getAccordionHint(s as AccordionResponse),
  calculation: (s) => getCalculationHint(s as CalculationResponse),
  sevenbridge: (s) => getSevenbridgeHint(s as SevenBridgeResponse),
  trash: (s) => getTrashHint(s as TrashResponse),
  president: (s) => getPresidentHint(s as PresidentResponse),
  cassino: (s) => getCassinoHint(s as CassinoResponse),
  scopa: (s) => getScopaHint(s as ScopaResponse),
  barbu: (s) => getBarbuHint(s as BarbuResponse),
  macau: (s) => getMacauHint(s as MacauResponse),
  clocksolitaire: (s) => getClocksolitaireHint(s as ClockSolitaireResponse),
  spiteandmalice: (s) => getSpiteAndMaliceHint(s as SpiteAndMaliceResponse),
  skat: (s) => getSkatHint(s as SkatResponse),
  shithead: (s) => getShitheadHint(s as ShitheadResponse),
  nertz: (s) => getNertzHint(s as NertzResponse),
  slapjack: (s) => getSlapjackHint(s as SlapjackResponse),
  egyptianratscrew: (s) => getEgyptianRatscrewHint(s as EgyptianRatscrewResponse),
  contractrummy: () => null,
  crescent: () => null,
  spiderette: (s) => getSpideretteHint(s as SpideretteResponse),
  gaps: (s) => getGapsHint(s as GapsResponse),
  fourcardpoker: (s) => getFourCardPokerHint(s as FourCardPokerResponse),
  rummy500: (s) => getRummy500Hint(s as Rummy500Response),
  russianpoker: (s) => getRussianPokerHint(s as RussianPokerResponse),
  sixcardgolf: (s) => getSixcardgolfHint(s as SixCardGolfResponse),
  doudizhu: (s) => getDoudizhuHint(s as DoudizhuResponse),
  bidwhist: (s) => getBidWhistHint(s as BidWhistResponse),
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
