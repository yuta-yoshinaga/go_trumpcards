import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import type {
  AccordionResponse,
  AcesUpResponse,
  AgnesResponse,
  AlaskaResponse,
  AllFoursResponse,
  AluetteResponse,
  AmericanToadResponse,
  AnacondaResponse,
  AndarBaharResponse,
  AuldLangSyneResponse,
  BaccaratResponse,
  BadugiResponse,
  BakersDozenResponse,
  BalootResponse,
  BanLuckResponse,
  BarbuResponse,
  BaseballPokerResponse,
  BasraResponse,
  BauernschnapsenResponse,
  BeggarMyNeighbourResponse,
  BeleagueredCastleResponse,
  BeloteResponse,
  BeziqueResponse,
  BhabhiResponse,
  BidEuchreResponse,
  BidWhistResponse,
  BigBenResponse,
  BigTwoResponse,
  BisleyResponse,
  BlackHoleResponse,
  BlackJackResponse,
  BlackJackSwitchResponse,
  BostonResponse,
  BotifarraResponse,
  BouillotteResponse,
  BourreResponse,
  BraidResponse,
  BridgeResponse,
  BriscolaResponse,
  BristolResponse,
  BrusquembilleResponse,
  BuraResponse,
  BurracoResponse,
  CalabresellaResponse,
  CalculationResponse,
  CallBreakResponse,
  CanastaResponse,
  CanfieldResponse,
  CaribbeanDrawResponse,
  CaribbeanStudResponse,
  CariocaResponse,
  CasinoHoldemResponse,
  CasinoWarResponse,
  CassinoResponse,
  CatchTenResponse,
  CegoResponse,
  ChemindeFerResponse,
  ChinchonResponse,
  ChinesePokerResponse,
  ChineseTenResponse,
  CinchResponse,
  CincinnatiResponse,
  ClockSolitaireResponse,
  ColoradoResponse,
  ColourWhistResponse,
  CongressResponse,
  ConquianResponse,
  ContractRummyResponse,
  CourtPieceResponse,
  CrazyEightsResponse,
  CrazyFourPokerResponse,
  CrazyQuiltResponse,
  CrescentResponse,
  CribbageResponse,
  CribbageSquaresResponse,
  CruelResponse,
  CuarentaResponse,
  CuckooResponse,
  CucumberResponse,
  CurdsAndWheyResponse,
  DaifugoResponse,
  DesmocheResponse,
  DeuceToSevenResponse,
  DiplomatResponse,
  DoppelkopfResponse,
  DoubleAttackResponse,
  DoubleKlondikeResponse,
  DoubtResponse,
  DoudizhuResponse,
  DragonTigerResponse,
  DramahaResponse,
  DuchessResponse,
  DurakResponse,
  EasthavenResponse,
  EcarteResponse,
  EgyptianRatscrewResponse,
  EightOffResponse,
  EscobaResponse,
  EstimationResponse,
  EuchreResponse,
  FaroResponse,
  FiftyOneResponse,
  FiveCardStudResponse,
  FiveHundredResponse,
  FlowerGardenResponse,
  FollowTheQueenResponse,
  FortressResponse,
  FortyAndEightResponse,
  FortyFivesResponse,
  FortyThievesResponse,
  FourCardPokerResponse,
  FourSeasonsResponse,
  FourteenOutResponse,
  FreeBetResponse,
  FreeCellResponse,
  FrenchTarotResponse,
  GaigelResponse,
  GanjifaResponse,
  GapsResponse,
  GermanWhistResponse,
  GinRummyResponse,
  GoFishResponse,
  GolfResponse,
  GongZhuResponse,
  GoofspielResponse,
  GoStopResponse,
  GrandfathersClockResponse,
  GuandanResponse,
  GutsResponse,
  HachiHachiResponse,
  HandAndFootResponse,
  HasenpfefferResponse,
  HeartsResponse,
  HighCardFlushResponse,
  HokmResponse,
  HoldemResponse,
  HoneymoonBridgeResponse,
  HorseResponse,
  IndianPokerResponse,
  IndianRummyResponse,
  IronCrossResponse,
  IsraeliWhistResponse,
  JassResponse,
  JulepeResponse,
  KaiserResponse,
  KalookiResponse,
  KarnoffelResponse,
  KempsResponse,
  KilleResponse,
  KingAlbertResponse,
  KingoResponse,
  KingResponse,
  KlaberjassResponse,
  KlaverjasResponse,
  KlondikeResponse,
  KnockoutWhistResponse,
  KoenigrufenResponse,
  KoiKoiResponse,
  LaBelleLucieResponse,
  LaughAndLieDownResponse,
  LetItRideResponse,
  LingerLongerResponse,
  LiteratureResponse,
  LobaResponse,
  LooResponse,
  MacauResponse,
  MachiavelliResponse,
  MadrassoResponse,
  ManilleResponse,
  MaoResponse,
  MariasResponse,
  MemoryResponse,
  MendikotResponse,
  MichiganResponse,
  MightyResponse,
  MinchiateResponse,
  MinibridgeResponse,
  MississippiStudResponse,
  MissMilliganResponse,
  MonteBankResponse,
  MonteCarloResponse,
  MrsMopResponse,
  MushiResponse,
  MusResponse,
  NainJauneResponse,
  NapoleonResponse,
  NapoleonsSquareResponse,
  NapResponse,
  NarcoticResponse,
  NertzResponse,
  NinetyNineResponse,
  NiuNiuResponse,
  OasisPokerResponse,
  OhHellResponse,
  OichoKabuResponse,
  OldMaidResponse,
  OmahaResponse,
  OmbreResponse,
  OpenFaceChineseResponse,
  OsmosisResponse,
  PageOneResponse,
  PaiGowResponse,
  PanResponse,
  PasurResponse,
  PenguinResponse,
  PerseveranceResponse,
  PigResponse,
  PigsTailResponse,
  PineappleResponse,
  PinochleResponse,
  PiquetResponse,
  PishtiResponse,
  PitchResponse,
  PochResponse,
  PokerResponse,
  PokerSquaresResponse,
  PolignacResponse,
  PontoonResponse,
  PopeJoanResponse,
  PreferenceResponse,
  PresidentResponse,
  PrimeroResponse,
  PrsiResponse,
  PutResponse,
  PyramidResponse,
  QuadrilleResponse,
  RamschResponse,
  RamsResponse,
  RankAndFileResponse,
  RedDogResponse,
  ReversisResponse,
  RikkenResponse,
  RistikontraResponse,
  RollingStoneResponse,
  RookResponse,
  RoyalCotillionResponse,
  Rummy500Response,
  RussianBankResponse,
  RussianPokerResponse,
  RussianSolitaireResponse,
  SakuraResponse,
  SalicLawResponse,
  SambaResponse,
  ScartoResponse,
  SchafkopfResponse,
  SchnapsenResponse,
  ScopaResponse,
  ScoponeResponse,
  ScorpionResponse,
  SeahavenTowersResponse,
  SedmaResponse,
  SergeantMajorResponse,
  SetteEMezzoResponse,
  SevenBridgeResponse,
  SevenCardStudResponse,
  SevensResponse,
  SevenTwentySevenResponse,
  ShamrocksResponse,
  SheepsheadResponse,
  ShelemResponse,
  ShengJiResponse,
  ShitheadResponse,
  ShortDeckResponse,
  SimpleSimonResponse,
  SirTommyResponse,
  SixBidSoloResponse,
  SixCardGolfResponse,
  SjavsResponse,
  SkatResponse,
  SkitgubbeResponse,
  SlapjackResponse,
  SlobberhannesResponse,
  SlyFoxResponse,
  SnapResponse,
  SoloWhistResponse,
  SomersetResponse,
  SpadesResponse,
  SpeculationResponse,
  SpeedResponse,
  SpideretteResponse,
  SpiderResponse,
  SpiteAndMaliceResponse,
  SpoilFiveResponse,
  SpoonsResponse,
  StalactitesResponse,
  StealingBundlesResponse,
  StHelenaResponse,
  StreetsAndAlleysResponse,
  SuecaResponse,
  SultanResponse,
  TablanetResponse,
  TarabishResponse,
  TarneebResponse,
  TarocchiniResponse,
  TeenDoPaanchResponse,
  TeenPattiResponse,
  TerraceResponse,
  TexasHoldemBonusResponse,
  ThirtyOneResponse,
  ThreeCardBragResponse,
  ThreeCardResponse,
  ThreeCardRummyResponse,
  ThreeThirteenResponse,
  TichuResponse,
  TienLenResponse,
  ToepenResponse,
  TonkResponse,
  TrappolaResponse,
  TrashResponse,
  TrenteEtQuaranteResponse,
  TressetteResponse,
  TrexResponse,
  TriPeaksResponse,
  TrogguResponse,
  TrucoResponse,
  TuSacResponse,
  TuteResponse,
  TwentyNineResponse,
  TwoTenJackResponse,
  TysiacResponse,
  UltimateTexasHoldemResponse,
  UltiResponse,
  VideoPokerResponse,
  VintResponse,
  ViraResponse,
  WarResponse,
  WaspResponse,
  WattenResponse,
  WhistResponse,
  WhiteheadResponse,
  WindmillResponse,
  WizardResponse,
  YanivResponse,
  YukonResponse,
  ZhengResponse,
  ZwanzigerrufenResponse,
  ZwickerResponse,
} from '../types/card';
import type { HintResult } from '../types/hint';
import { getAccordionHint } from '../utils/hints/accordionHint';
import { getAcesUpHint } from '../utils/hints/acesupHint';
import { getAgnesHint } from '../utils/hints/agnesHint';
import { getAlaskaHint } from '../utils/hints/alaskaHint';
import { getAllFoursHint } from '../utils/hints/allfoursHint';
import { getAluetteHint } from '../utils/hints/aluetteHint';
import { getAmericanToadHint } from '../utils/hints/americantoadHint';
import { getAnacondaHint } from '../utils/hints/anacondaHint';
import { getAndarbaharHint } from '../utils/hints/andarbaharHint';
import { getAuldLangSyneHint } from '../utils/hints/auldlangsyneHint';
import { getBaccaratHint } from '../utils/hints/baccaratHint';
import { getBadugiHint } from '../utils/hints/badugiHint';
import { getBakersdozenHint } from '../utils/hints/bakersdozenHint';
import { getBalootHint } from '../utils/hints/balootHint';
import { getBanluckHint } from '../utils/hints/banluckHint';
import { getBarbuHint } from '../utils/hints/barbuHint';
import { getBaseballpokerHint } from '../utils/hints/baseballpokerHint';
import { getBasraHint } from '../utils/hints/basraHint';
import { getBauernschnapsenHint } from '../utils/hints/bauernschnapsenHint';
import { getBeggarMyNeighbourHint } from '../utils/hints/beggarmyneighbourHint';
import { getBeleagueredcastleHint } from '../utils/hints/beleagueredcastleHint';
import { getBeloteHint } from '../utils/hints/beloteHint';
import { getBeziqueHint } from '../utils/hints/beziqueHint';
import { getBhabhiHint } from '../utils/hints/bhabhiHint';
import { getBidEuchreHint } from '../utils/hints/bideuchreHint';
import { getBidWhistHint } from '../utils/hints/bidwhistHint';
import { getBigBenHint } from '../utils/hints/bigbenHint';
import { getBigTwoHint } from '../utils/hints/bigtwoHint';
import { getBisleyHint } from '../utils/hints/bisleyHint';
import { getBlackHoleHint } from '../utils/hints/blackholeHint';
import { getBlackjackHint } from '../utils/hints/blackjackHint';
import { getBlackjackswitchHint } from '../utils/hints/blackjackswitchHint';
import { getBostonHint } from '../utils/hints/bostonHint';
import { getBotifarraHint } from '../utils/hints/botifarraHint';
import { getBouillotteHint } from '../utils/hints/bouillotteHint';
import { getBourreHint } from '../utils/hints/bourreHint';
import { getBraidHint } from '../utils/hints/braidHint';
import { getBridgeHint } from '../utils/hints/bridgeHint';
import { getBriscolaHint } from '../utils/hints/briscolaHint';
import { getBristolHint } from '../utils/hints/bristolHint';
import { getBrusquembilleHint } from '../utils/hints/brusquembilleHint';
import { getBuraHint } from '../utils/hints/buraHint';
import { getBurracoHint } from '../utils/hints/burracoHint';
import { getCalabresellaHint } from '../utils/hints/calabresellaHint';
import { getCalculationHint } from '../utils/hints/calculationHint';
import { getCallBreakHint } from '../utils/hints/callbreakHint';
import { getCanastaHint } from '../utils/hints/canastaHint';
import { getCanfieldHint } from '../utils/hints/canfieldHint';
import { getCaribbeanDrawHint } from '../utils/hints/caribbeandrawHint';
import { getCaribbeanStudHint } from '../utils/hints/caribbeanstudHint';
import { getCariocaHint } from '../utils/hints/cariocaHint';
import { getCasinoHoldemHint } from '../utils/hints/casinoholdemHint';
import { getCasinowarHint } from '../utils/hints/casinowarHint';
import { getCassinoHint } from '../utils/hints/cassinoHint';
import { getCatchTenHint } from '../utils/hints/catchtenHint';
import { getCegoHint } from '../utils/hints/cegoHint';
import { getChemindeferHint } from '../utils/hints/chemindeferHint';
import { getChinchonHint } from '../utils/hints/chinchonHint';
import { getChinesePokerHint } from '../utils/hints/chinesepokerHint';
import { getChineseTenHint } from '../utils/hints/chinesetenHint';
import { getCinchHint } from '../utils/hints/cinchHint';
import { getCincinnatiHint } from '../utils/hints/cincinnatiHint';
import { getClocksolitaireHint } from '../utils/hints/clocksolitaireHint';
import { getColoradoHint } from '../utils/hints/coloradoHint';
import { getColourwhistHint } from '../utils/hints/colourwhistHint';
import { getCongressHint } from '../utils/hints/congressHint';
import { getConquianHint } from '../utils/hints/conquianHint';
import { getContractRummyHint } from '../utils/hints/contractrummyHint';
import { getCourtPieceHint } from '../utils/hints/courtPieceHint';
import { getCrazyEightsHint } from '../utils/hints/crazyeightsHint';
import { getCrazyfourpokerHint } from '../utils/hints/crazyfourpokerHint';
import { getCrazyPineappleHint } from '../utils/hints/crazyPineappleHint';
import { getCrazyQuiltHint } from '../utils/hints/crazyquiltHint';
import { getCrescentHint } from '../utils/hints/crescentHint';
import { getCribbageHint } from '../utils/hints/cribbageHint';
import { getCribbageSquaresHint } from '../utils/hints/cribbagesquaresHint';
import { getCruelHint } from '../utils/hints/cruelHint';
import { getCuarentaHint } from '../utils/hints/cuarentaHint';
import { getCuckooHint } from '../utils/hints/cuckooHint';
import { getCucumberHint } from '../utils/hints/cucumberHint';
import { getCurdsAndWheyHint } from '../utils/hints/curdsandwheyHint';
import { getDaifugoHint } from '../utils/hints/daifugoHint';
import { getDesmocheHint } from '../utils/hints/desmocheHint';
import { getDeucesWildHint } from '../utils/hints/deuceswildHint';
import { getDeuceToSevenHint } from '../utils/hints/deuceToSevenHint';
import { getDiplomatHint } from '../utils/hints/diplomatHint';
import { getDoppelkopfHint } from '../utils/hints/doppelkopfHint';
import { getDoubleattackHint } from '../utils/hints/doubleattackHint';
import { getDoubleKlondikeHint } from '../utils/hints/doubleklondikeHint';
import { getDoubtHint } from '../utils/hints/doubtHint';
import { getDoudizhuHint } from '../utils/hints/doudizhuHint';
import { getDragontigerHint } from '../utils/hints/dragontigerHint';
import { getDramahaHint } from '../utils/hints/dramahaHint';
import { getDuchessHint } from '../utils/hints/duchessHint';
import { getDurakHint } from '../utils/hints/durakHint';
import { getEasthavenHint } from '../utils/hints/easthavenHint';
import { getEcarteHint } from '../utils/hints/ecarteHint';
import { getEgyptianRatscrewHint } from '../utils/hints/egyptianratscrewHint';
import { getEightOffHint } from '../utils/hints/eightoffHint';
import { getEscobaHint } from '../utils/hints/escobaHint';
import { getEstimationHint } from '../utils/hints/estimationHint';
import { getEuchreHint } from '../utils/hints/euchreHint';
import { getFaroHint } from '../utils/hints/faroHint';
import { getFiftyOneHint } from '../utils/hints/fiftyoneHint';
import { getFiveCardStudHint } from '../utils/hints/fivecardstudHint';
import { getFiveHundredHint } from '../utils/hints/fivehundredHint';
import { getFlowergardenHint } from '../utils/hints/flowergardenHint';
import { getFollowTheQueenHint } from '../utils/hints/followthequeenHint';
import { getFortressHint } from '../utils/hints/fortressHint';
import { getFortyAndEightHint } from '../utils/hints/fortyandeightHint';
import { getFortyFivesHint } from '../utils/hints/fortyFivesHint';
import { getFortyThievesHint } from '../utils/hints/fortythievesHint';
import { getFourCardPokerHint } from '../utils/hints/fourcardpokerHint';
import { getFourSeasonsHint } from '../utils/hints/fourseasonsHint';
import { getFourteenOutHint } from '../utils/hints/fourteenoutHint';
import { getFreebetHint } from '../utils/hints/freebetHint';
import { getFreeCellHint } from '../utils/hints/freecellHint';
import { getFrenchTarotHint } from '../utils/hints/frenchtarotHint';
import { getGaigelHint } from '../utils/hints/gaigelHint';
import { getGanjifaHint } from '../utils/hints/ganjifaHint';
import { getGapsHint } from '../utils/hints/gapsHint';
import { getGermanWhistHint } from '../utils/hints/germanwhistHint';
import { getGinRummyHint } from '../utils/hints/ginrummyHint';
import { getGoFishHint } from '../utils/hints/gofishHint';
import { getGolfHint } from '../utils/hints/golfHint';
import { getGongZhuHint } from '../utils/hints/gongzhuHint';
import { getGoofspielHint } from '../utils/hints/goofspielHint';
import { getGoStopHint } from '../utils/hints/gostopHint';
import { getGrandfathersClockHint } from '../utils/hints/grandfathersclockHint';
import { getGuandanHint } from '../utils/hints/guandanHint';
import { getGutsHint } from '../utils/hints/gutsHint';
import { getHachiHachiHint } from '../utils/hints/hachihachiHint';
import { getHandAndFootHint } from '../utils/hints/handandfootHint';
import { getHasenpfefferHint } from '../utils/hints/hasenpfefferHint';
import { getHeartsHint } from '../utils/hints/heartsHint';
import { getHighCardFlushHint } from '../utils/hints/highcardflushHint';
import { getHokmHint } from '../utils/hints/hokmHint';
import { getHoldemHint } from '../utils/hints/holdemHint';
import { getHoneymoonBridgeHint } from '../utils/hints/honeymoonbridgeHint';
import { getHorseHint } from '../utils/hints/horseHint';
import { getIndianPokerHint } from '../utils/hints/indianpokerHint';
import { getIndianRummyHint } from '../utils/hints/indianRummyHint';
import { getIrishPokerHint } from '../utils/hints/irishPokerHint';
import { getIroncrossHint } from '../utils/hints/ironcrossHint';
import { getIsraeliWhistHint } from '../utils/hints/israeliwhistHint';
import { getJassHint } from '../utils/hints/jassHint';
import { getJokerPokerHint } from '../utils/hints/jokerpokerHint';
import { getJulepeHint } from '../utils/hints/julepeHint';
import { getKaiserHint } from '../utils/hints/kaiserHint';
import { getKalookiHint } from '../utils/hints/kalookiHint';
import { getKarnoffelHint } from '../utils/hints/karnoffelHint';
import { getKempsHint } from '../utils/hints/kempsHint';
import { getKilleHint } from '../utils/hints/killeHint';
import { getKingalbertHint } from '../utils/hints/kingalbertHint';
import { getKingHint } from '../utils/hints/kingHint';
import { getKingoHint } from '../utils/hints/kingoHint';
import { getKlaberjassHint } from '../utils/hints/klaberjassHint';
import { getKlaverjasHint } from '../utils/hints/klaverjasHint';
import { getKlondikeHint } from '../utils/hints/klondikeHint';
import { getKnockoutWhistHint } from '../utils/hints/knockoutWhistHint';
import { getKoenigrufenHint } from '../utils/hints/koenigrufenHint';
import { getKoiKoiHint } from '../utils/hints/koikoiHint';
import { getLaBelleLucieHint } from '../utils/hints/labellelucieHint';
import { getLaughAndLieDownHint } from '../utils/hints/laughandliedownHint';
import { getLetitrideHint } from '../utils/hints/letitrideHint';
import { getLingerLongerHint } from '../utils/hints/lingerlongerHint';
import { getLiteratureHint } from '../utils/hints/literatureHint';
import { getLobaHint } from '../utils/hints/lobaHint';
import { getLooHint } from '../utils/hints/looHint';
import { getMacauHint } from '../utils/hints/macauHint';
import { getMachiavelliHint } from '../utils/hints/machiavelliHint';
import { getMadrassoHint } from '../utils/hints/madrassoHint';
import { getManilleHint } from '../utils/hints/manilleHint';
import { getMaoHint } from '../utils/hints/maoHint';
import { getMariasHint } from '../utils/hints/mariasHint';
import { getMemoryHint } from '../utils/hints/memoryHint';
import { getMendikotHint } from '../utils/hints/mendikotHint';
import { getMichiganHint } from '../utils/hints/michiganHint';
import { getMightyHint } from '../utils/hints/mightyHint';
import { getMinchiateHint } from '../utils/hints/minchiateHint';
import { getMinibridgeHint } from '../utils/hints/minibridgeHint';
import { getMississippiStudHint } from '../utils/hints/mississippiStudHint';
import { getMissMilliganHint } from '../utils/hints/missmilliganHint';
import { getMontebankHint } from '../utils/hints/montebankHint';
import { getMonteCarloHint } from '../utils/hints/montecarloHint';
import { getMrsMopHint } from '../utils/hints/mrsMopHint';
import { getMusHint } from '../utils/hints/musHint';
import { getMushiHint } from '../utils/hints/mushiHint';
import { getNainJauneHint } from '../utils/hints/nainjauneHint';
import { getNapHint } from '../utils/hints/napHint';
import { getNapoleonHint } from '../utils/hints/napoleonHint';
import { getNapoleonsSquareHint } from '../utils/hints/napoleonssquareHint';
import { getNarcoticHint } from '../utils/hints/narcoticHint';
import { getNertzHint } from '../utils/hints/nertzHint';
import { getNinetyNineHint } from '../utils/hints/ninetynineHint';
import { getNiuNiuHint } from '../utils/hints/niuniuHint';
import { getOasisPokerHint } from '../utils/hints/oasispokerHint';
import { getOhHellHint } from '../utils/hints/ohhellHint';
import { getOichokabuHint } from '../utils/hints/oichokabuHint';
import { getOldMaidHint } from '../utils/hints/oldmaidHint';
import { getOmahaHiLoHint } from '../utils/hints/omahaHiLoHint';
import { getOmahaHint } from '../utils/hints/omahaHint';
import { getOmbreHint } from '../utils/hints/ombreHint';
import { getOpenFaceChineseHint } from '../utils/hints/openfacechineseHint';
import { getOsmosisHint } from '../utils/hints/osmosisHint';
import { getPageOneHint } from '../utils/hints/pageoneHint';
import { getPaiGowHint } from '../utils/hints/paigowHint';
import { getPanHint } from '../utils/hints/panHint';
import { getPasurHint } from '../utils/hints/pasurHint';
import { getPenguinHint } from '../utils/hints/penguinHint';
import { getPerseveranceHint } from '../utils/hints/perseveranceHint';
import { getPigHint } from '../utils/hints/pigHint';
import { getPigstailHint } from '../utils/hints/pigstailHint';
import { getPineappleHint } from '../utils/hints/pineappleHint';
import { getPinochleHint } from '../utils/hints/pinochleHint';
import { getPiquetHint } from '../utils/hints/piquetHint';
import { getPishtiHint } from '../utils/hints/pishtiHint';
import { getPitchHint } from '../utils/hints/pitchHint';
import { getPochHint } from '../utils/hints/pochHint';
import { getPokerHint } from '../utils/hints/pokerHint';
import { getPokersquaresHint } from '../utils/hints/pokersquaresHint';
import { getPolignacHint } from '../utils/hints/polignacHint';
import { getPontoonHint } from '../utils/hints/pontoonHint';
import { getPopeJoanHint } from '../utils/hints/popejoanHint';
import { getPreferenceHint } from '../utils/hints/preferenceHint';
import { getPresidentHint } from '../utils/hints/presidentHint';
import { getPrimeroHint } from '../utils/hints/primeroHint';
import { getPrsiHint } from '../utils/hints/prsiHint';
import { getPutHint } from '../utils/hints/putHint';
import { getPyramidHint } from '../utils/hints/pyramidHint';
import { getQuadrilleHint } from '../utils/hints/quadrilleHint';
import { getRamschHint } from '../utils/hints/ramschHint';
import { getRamsHint } from '../utils/hints/ramsHint';
import { getRankAndFileHint } from '../utils/hints/rankandfileHint';
import { getRazzHint } from '../utils/hints/razzHint';
import { getReddogHint } from '../utils/hints/reddogHint';
import { getReversisHint } from '../utils/hints/reversisHint';
import { getRikkenHint } from '../utils/hints/rikkenHint';
import { getRistikontraHint } from '../utils/hints/ristikontraHint';
import { getRollingStoneHint } from '../utils/hints/rollingstoneHint';
import { getRookHint } from '../utils/hints/rookHint';
import { getRoyalCotillionHint } from '../utils/hints/royalcotillionHint';
import { getRummy500Hint } from '../utils/hints/rummy500Hint';
import { getRussianBankHint } from '../utils/hints/russianbankHint';
import { getRussianPokerHint } from '../utils/hints/russianpokerHint';
import { getRussianSolitaireHint } from '../utils/hints/russianSolitaireHint';
import { getSakuraHint } from '../utils/hints/sakuraHint';
import { getSalicLawHint } from '../utils/hints/saliclawHint';
import { getSambaHint } from '../utils/hints/sambaHint';
import { getScartoHint } from '../utils/hints/scartoHint';
import { getSchafkopfHint } from '../utils/hints/schafkopfHint';
import { getSchnapsenHint } from '../utils/hints/schnapsenHint';
import { getScopaHint } from '../utils/hints/scopaHint';
import { getScoponeHint } from '../utils/hints/scoponeHint';
import { getScorpionHint } from '../utils/hints/scorpionHint';
import { getSeahavenTowersHint } from '../utils/hints/seahavenTowersHint';
import { getSedmaHint } from '../utils/hints/sedmaHint';
import { getSergeantMajorHint } from '../utils/hints/sergeantmajorHint';
import { getSetteEMezzoHint } from '../utils/hints/settemezzoHint';
import { getSevenbridgeHint } from '../utils/hints/sevenbridgeHint';
import { getSevenCardStudHint } from '../utils/hints/sevencardstudHint';
import { getSevensHint } from '../utils/hints/sevensHint';
import { getSevenTwentySevenHint } from '../utils/hints/seventwentysevenHint';
import { getShamrocksHint } from '../utils/hints/shamrocksHint';
import { getSheepsheadHint } from '../utils/hints/sheepsheadHint';
import { getShelemHint } from '../utils/hints/shelemHint';
import { getShengJiHint } from '../utils/hints/shengjiHint';
import { getShitheadHint } from '../utils/hints/shitheadHint';
import { getShortDeckHint } from '../utils/hints/shortdeckHint';
import { getSimpleSimonHint } from '../utils/hints/simplesimonHint';
import { getSirTommyHint } from '../utils/hints/sirtommyHint';
import { getSixBidSoloHint } from '../utils/hints/sixbidsoloHint';
import { getSixcardgolfHint } from '../utils/hints/sixcardgolfHint';
import { getSjavsHint } from '../utils/hints/sjavsHint';
import { getSkatHint } from '../utils/hints/skatHint';
import { getSkitgubbeHint } from '../utils/hints/skitgubbeHint';
import { getSlapjackHint } from '../utils/hints/slapjackHint';
import { getSlobberhannesHint } from '../utils/hints/slobberhannesHint';
import { getSlyFoxHint } from '../utils/hints/slyfoxHint';
import { getSnapHint } from '../utils/hints/snapHint';
import { getSokoHint } from '../utils/hints/sokoHint';
import { getSoloWhistHint } from '../utils/hints/soloWhistHint';
import { getSomersetHint } from '../utils/hints/somersetHint';
import { getSpadesHint } from '../utils/hints/spadesHint';
import { getSpeculationHint } from '../utils/hints/speculationHint';
import { getSpeedHint } from '../utils/hints/speedHint';
import { getSpideretteHint } from '../utils/hints/spideretteHint';
import { getSpiderHint } from '../utils/hints/spiderHint';
import { getSpiteAndMaliceHint } from '../utils/hints/spiteAndMaliceHint';
import { getSpoilFiveHint } from '../utils/hints/spoilFiveHint';
import { getSpoonsHint } from '../utils/hints/spoonsHint';
import { getStalactitesHint } from '../utils/hints/stalactitesHint';
import { getStealingBundlesHint } from '../utils/hints/stealingbundlesHint';
import { getStHelenaHint } from '../utils/hints/sthelenaHint';
import { getStreetsandalleysHint } from '../utils/hints/streetsandalleysHint';
import { getSuecaHint } from '../utils/hints/suecaHint';
import { getSultanHint } from '../utils/hints/sultanHint';
import { getTablanetHint } from '../utils/hints/tablanetHint';
import { getTarabishHint } from '../utils/hints/tarabishHint';
import { getTarneebHint } from '../utils/hints/tarneebHint';
import { getTarocchiniHint } from '../utils/hints/tarocchiniHint';
import { getTeenDoPaanchHint } from '../utils/hints/teendopaanchHint';
import { getTeenPattiHint } from '../utils/hints/teenPattiHint';
import { getTerraceHint } from '../utils/hints/terraceHint';
import { getTexasHoldemBonusHint } from '../utils/hints/texasHoldemBonusHint';
import { getThirtyOneHint } from '../utils/hints/thirtyoneHint';
import { getThreeCardBragHint } from '../utils/hints/threeCardBragHint';
import { getThreeCardHint } from '../utils/hints/threecardHint';
import { getThreeCardRummyHint } from '../utils/hints/threecardrummyHint';
import { getThreeThirteenHint } from '../utils/hints/threethirteenHint';
import { getTichuHint } from '../utils/hints/tichuHint';
import { getTienLenHint } from '../utils/hints/tienlenHint';
import { getToepenHint } from '../utils/hints/toepenHint';
import { getTonkHint } from '../utils/hints/tonkHint';
import { getTrappolaHint } from '../utils/hints/trappolaHint';
import { getTrashHint } from '../utils/hints/trashHint';
import { getTrenteEtQuaranteHint } from '../utils/hints/trenteetquaranteHint';
import { getTressetteHint } from '../utils/hints/tressetteHint';
import { getTrexHint } from '../utils/hints/trexHint';
import { getTriPeaksHint } from '../utils/hints/tripeaksHint';
import { getTrogguHint } from '../utils/hints/trogguHint';
import { getTrucoHint } from '../utils/hints/trucoHint';
import { getTusacHint } from '../utils/hints/tusacHint';
import { getTuteHint } from '../utils/hints/tuteHint';
import { getTwentyNineHint } from '../utils/hints/twentyNineHint';
import { getTwoTenJackHint } from '../utils/hints/twotenjackHint';
import { getTysiacHint } from '../utils/hints/tysiacHint';
import { getUltiHint } from '../utils/hints/ultiHint';
import { getUltimateTexasHoldemHint } from '../utils/hints/ultimateTexasHoldemHint';
import { getVideoPokerHint } from '../utils/hints/videopokerHint';
import { getVintHint } from '../utils/hints/vintHint';
import { getViraHint } from '../utils/hints/viraHint';
import { getWarHint } from '../utils/hints/warHint';
import { getWaspHint } from '../utils/hints/waspHint';
import { getWattenHint } from '../utils/hints/wattenHint';
import { getWhistHint } from '../utils/hints/whistHint';
import { getWhiteheadHint } from '../utils/hints/whiteheadHint';
import { getWindmillHint } from '../utils/hints/windmillHint';
import { getWizardHint } from '../utils/hints/wizardHint';
import { getYanivHint } from '../utils/hints/yanivHint';
import { getYukonHint } from '../utils/hints/yukonHint';
import { getZhengHint } from '../utils/hints/zhengHint';
import { getZwanzigerrufenHint } from '../utils/hints/zwanzigerrufenHint';
import { getZwickerHint } from '../utils/hints/zwickerHint';
import { useLocalStorageToggle } from './useLocalStorageToggle';

/** Hint function that takes game state and returns a hint result or null. */
type HintFn = (state: unknown) => HintResult | null;

/** Registry mapping game names to their hint functions. */
export const hintFactories = {
  baccarat: (s) => getBaccaratHint(s as BaccaratResponse),
  bideuchre: (s) => getBidEuchreHint(s as BidEuchreResponse),
  blackjack: (s) => getBlackjackHint(s as BlackJackResponse),
  spanish21: (s) => getBlackjackHint(s as BlackJackResponse),
  pontoon: (s) => getPontoonHint(s as PontoonResponse),
  poker: (s) => getPokerHint(s as PokerResponse),
  handandfoot: (s) => getHandAndFootHint(s as HandAndFootResponse),
  hearts: (s) => getHeartsHint(s as HeartsResponse),
  gongzhu: (s) => getGongZhuHint(s as GongZhuResponse),
  spoons: (s) => getSpoonsHint(s as SpoonsResponse),
  sixbidsolo: (s) => getSixBidSoloHint(s as SixBidSoloResponse),
  spades: (s) => getSpadesHint(s as SpadesResponse),
  callbreak: (s) => getCallBreakHint(s as CallBreakResponse),
  tarneeb: (s) => getTarneebHint(s as TarneebResponse),
  pishti: (s) => getPishtiHint(s as PishtiResponse),
  ristikontra: (s) => getRistikontraHint(s as RistikontraResponse),
  piquet: (s) => getPiquetHint(s as PiquetResponse),
  pitch: (s) => getPitchHint(s as PitchResponse),
  holdem: (s) => getHoldemHint(s as HoldemResponse),
  omaha: (s) => getOmahaHint(s as OmahaResponse),
  omahahilo: (s) => getOmahaHiLoHint(s as OmahaResponse),
  dramaha: (s) => getDramahaHint(s as DramahaResponse),
  bigo: (s) => getOmahaHint(s as OmahaResponse),
  courchevel: (s) => getOmahaHint(s as OmahaResponse),
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
  threecardrummy: (s) => getThreeCardRummyHint(s as ThreeCardRummyResponse),
  tichu: (s) => getTichuHint(s as TichuResponse),
  highcardflush: (s) => getHighCardFlushHint(s as HighCardFlushResponse),
  escoba: (s) => getEscobaHint(s as EscobaResponse),
  euchre: (s) => getEuchreHint(s as EuchreResponse),
  belote: (s) => getBeloteHint(s as BeloteResponse),
  jass: (s) => getJassHint(s as JassResponse),
  bauernschnapsen: (s) => getBauernschnapsenHint(s as BauernschnapsenResponse),
  gaigel: (s) => getGaigelHint(s as GaigelResponse),
  bigtwo: (s) => getBigTwoHint(s as BigTwoResponse),
  threethirteen: (s) => getThreeThirteenHint(s as ThreeThirteenResponse),
  tienlen: (s) => getTienLenHint(s as TienLenResponse),
  zheng: (s) => getZhengHint(s as ZhengResponse),
  fivecardstud: (s) => getFiveCardStudHint(s as FiveCardStudResponse),
  soko: (s) => getSokoHint(s as FiveCardStudResponse),
  fourseasons: (s) => getFourSeasonsHint(s as FourSeasonsResponse),
  colorado: (s) => getColoradoHint(s as ColoradoResponse),
  slyfox: (s) => getSlyFoxHint(s as SlyFoxResponse),
  cribbagesquares: (s) => getCribbageSquaresHint(s as CribbageSquaresResponse),
  diplomat: (s) => getDiplomatHint(s as DiplomatResponse),
  royalcotillion: (s) => getRoyalCotillionHint(s as RoyalCotillionResponse),
  crazyquilt: (s) => getCrazyQuiltHint(s as CrazyQuiltResponse),
  fivehundred: (s) => getFiveHundredHint(s as FiveHundredResponse),
  rook: (s) => getRookHint(s as RookResponse),
  schnapsen: (s) => getSchnapsenHint(s as SchnapsenResponse),
  germanwhist: (s) => getGermanWhistHint(s as GermanWhistResponse),
  slobberhannes: (s) => getSlobberhannesHint(s as SlobberhannesResponse),
  polignac: (s) => getPolignacHint(s as PolignacResponse),
  reversis: (s) => getReversisHint(s as ReversisResponse),
  rams: (s) => getRamsHint(s as RamsResponse),
  tarabish: (s) => getTarabishHint(s as TarabishResponse),
  baloot: (s) => getBalootHint(s as BalootResponse),
  estimation: (s) => getEstimationHint(s as EstimationResponse),
  israeliwhist: (s) => getIsraeliWhistHint(s as IsraeliWhistResponse),
  hokm: (s) => getHokmHint(s as HokmResponse),
  shelem: (s) => getShelemHint(s as ShelemResponse),
  mendikot: (s) => getMendikotHint(s as MendikotResponse),
  bhabhi: (s) => getBhabhiHint(s as BhabhiResponse),
  teendopaanch: (s) => getTeenDoPaanchHint(s as TeenDoPaanchResponse),
  hasenpfeffer: (s) => getHasenpfefferHint(s as HasenpfefferResponse),
  sergeantmajor: (s) => getSergeantMajorHint(s as SergeantMajorResponse),
  honeymoonbridge: (s) => getHoneymoonBridgeHint(s as HoneymoonBridgeResponse),
  minibridge: (s) => getMinibridgeHint(s as MinibridgeResponse),
  pasur: (s) => getPasurHint(s as PasurResponse),
  snap: (s) => getSnapHint(s as SnapResponse),
  rollingstone: (s) => getRollingStoneHint(s as RollingStoneResponse),
  lingerlonger: (s) => getLingerLongerHint(s as LingerLongerResponse),
  pig: (s) => getPigHint(s as PigResponse),
  stealingbundles: (s) => getStealingBundlesHint(s as StealingBundlesResponse),
  cucumber: (s) => getCucumberHint(s as CucumberResponse),
  goofspiel: (s) => getGoofspielHint(s as GoofspielResponse),
  faro: (s) => getFaroHint(s as FaroResponse),
  fiftyone: (s) => getFiftyOneHint(s as FiftyOneResponse),
  napoleon: (s) => getNapoleonHint(s as NapoleonResponse),
  mighty: (s) => getMightyHint(s as MightyResponse),
  ohhell: (s) => getOhHellHint(s as OhHellResponse),
  wizard: (s) => getWizardHint(s as WizardResponse),
  niuniu: (s) => getNiuNiuHint(s as NiuNiuResponse),
  ninetynine: (s) => getNinetyNineHint(s as NinetyNineResponse),
  oldmaid: (s) => getOldMaidHint(s as OldMaidResponse),
  doubt: (s) => getDoubtHint(s as DoubtResponse),
  daifugo: (s) => getDaifugoHint(s as DaifugoResponse),
  settemezzo: (s) => getSetteEMezzoHint(s as SetteEMezzoResponse),
  sevens: (s) => getSevensHint(s as SevensResponse),
  conquian: (s) => getConquianHint(s as ConquianResponse),
  chinchon: (s) => getChinchonHint(s as ChinchonResponse),
  crazyeights: (s) => getCrazyEightsHint(s as CrazyEightsResponse),
  prsi: (s) => getPrsiHint(s as PrsiResponse),
  speed: (s) => getSpeedHint(s as SpeedResponse),
  klondike: (s) => getKlondikeHint(s as KlondikeResponse),
  whitehead: (s) => getWhiteheadHint(s as WhiteheadResponse),
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
  indianrummy: (s) => getIndianRummyHint(s as IndianRummyResponse),
  machiavelli: (s) => getMachiavelliHint(s as MachiavelliResponse),
  cuarenta: (s) => getCuarentaHint(s as CuarentaResponse),
  cribbage: (s) => getCribbageHint(s as CribbageResponse),
  cuckoo: (s) => getCuckooHint(s as CuckooResponse),
  gofish: (s) => getGoFishHint(s as GoFishResponse),
  guandan: (s) => getGuandanHint(s as GuandanResponse),
  golf: (s) => getGolfHint(s as GolfResponse),
  caribbeandraw: (s) => getCaribbeanDrawHint(s as CaribbeanDrawResponse),
  caribbeanstud: (s) => getCaribbeanStudHint(s as CaribbeanStudResponse),
  oasispoker: (s) => getOasisPokerHint(s as OasisPokerResponse),
  casinoholdem: (s) => getCasinoHoldemHint(s as CasinoHoldemResponse),
  texasholdembonus: (s) => getTexasHoldemBonusHint(s as TexasHoldemBonusResponse),
  ultimatetexasholdem: (s) => getUltimateTexasHoldemHint(s as UltimateTexasHoldemResponse),
  mississippistud: (s) => getMississippiStudHint(s as MississippiStudResponse),
  durak: (s) => getDurakHint(s as DurakResponse),
  canasta: (s) => getCanastaHint(s as CanastaResponse),
  samba: (s) => getSambaHint(s as SambaResponse),
  boston: (s) => getBostonHint(s as BostonResponse),
  bourre: (s) => getBourreHint(s as BourreResponse),
  bridge: (s) => getBridgeHint(s as BridgeResponse),
  bristol: (s) => getBristolHint(s as BristolResponse),
  burraco: (s) => getBurracoHint(s as BurracoResponse),
  canfield: (s) => getCanfieldHint(s as CanfieldResponse),
  agnes: (s) => getAgnesHint(s as AgnesResponse),
  openfacechinese: (s) => getOpenFaceChineseHint(s as OpenFaceChineseResponse),
  osmosis: (s) => getOsmosisHint(s as OsmosisResponse),
  pinochle: (s) => getPinochleHint(s as PinochleResponse),
  twotenjack: (s) => getTwoTenJackHint(s as TwoTenJackResponse),
  followthequeen: (s) => getFollowTheQueenHint(s as FollowTheQueenResponse),
  sevencardstud: (s) => getSevenCardStudHint(s as SevenCardStudResponse),
  // Hi-Lo shares the stud page and, with it, stud's absence of a frontend hint.
  sevencardstudhilo: (s) => getSevenCardStudHint(s as SevenCardStudResponse),
  razz: (s) => getRazzHint(s as SevenCardStudResponse),
  badugi: (s) => getBadugiHint(s as BadugiResponse),
  deucetoseven: (s) => getDeuceToSevenHint(s as DeuceToSevenResponse),
  fortythieves: (s) => getFortyThievesHint(s as FortyThievesResponse),
  fortyandeight: (s) => getFortyAndEightHint(s as FortyAndEightResponse),
  sultan: (s) => getSultanHint(s as SultanResponse),
  bakersdozen: (s) => getBakersdozenHint(s as BakersDozenResponse),
  perseverance: (s) => getPerseveranceHint(s as PerseveranceResponse),
  fourteenout: (s) => getFourteenOutHint(s as FourteenOutResponse),
  narcotic: (s) => getNarcoticHint(s as NarcoticResponse),
  mrsmop: (s) => getMrsMopHint(s as MrsMopResponse),
  rankandfile: (s) => getRankAndFileHint(s as RankAndFileResponse),
  beleagueredcastle: (s) => getBeleagueredcastleHint(s as BeleagueredCastleResponse),
  fortress: (s) => getFortressHint(s as FortressResponse),
  somerset: (s) => getSomersetHint(s as SomersetResponse),
  stalactites: (s) => getStalactitesHint(s as StalactitesResponse),
  streetsandalleys: (s) => getStreetsandalleysHint(s as StreetsAndAlleysResponse),
  kille: (s) => getKilleHint(s as KilleResponse),
  kingalbert: (s) => getKingalbertHint(s as KingAlbertResponse),
  karnoffel: (s) => getKarnoffelHint(s as KarnoffelResponse),
  king: (s) => getKingHint(s as KingResponse),
  flowergarden: (s) => getFlowergardenHint(s as FlowerGardenResponse),
  tonk: (s) => getTonkHint(s as TonkResponse),
  thirtyone: (s) => getThirtyOneHint(s as ThirtyOneResponse),
  yaniv: (s) => getYanivHint(s as YanivResponse),
  trappola: (s) => getTrappolaHint(s as TrappolaResponse),
  julepe: (s) => getJulepeHint(s as JulepeResponse),
  schafkopf: (s) => getSchafkopfHint(s as SchafkopfResponse),
  madrasso: (s) => getMadrassoHint(s as MadrassoResponse),
  tressette: (s) => getTressetteHint(s as TressetteResponse),
  paigow: (s) => getPaiGowHint(s as PaiGowResponse),
  chinesepoker: (s) => getChinesePokerHint(s as ChinesePokerResponse),
  pageone: (s) => getPageOneHint(s as PageOneResponse),
  pigtail: (s) => getPigstailHint(s as PigsTailResponse),
  pokersquares: (s) => getPokersquaresHint(s as PokerSquaresResponse),
  montecarlo: (s) => getMonteCarloHint(s as MonteCarloResponse),
  letitride: (s) => getLetitrideHint(s as LetItRideResponse),
  reddog: (s) => getReddogHint(s as RedDogResponse),
  casinowar: (s) => getCasinowarHint(s as CasinoWarResponse),
  oichokabu: (s) => getOichokabuHint(s as OichoKabuResponse),
  kingo: (s) => getKingoHint(s as KingoResponse),
  tusac: (s) => getTusacHint(s as TuSacResponse),
  sakura: (s) => getSakuraHint(s as SakuraResponse),
  zwanzigerrufen: (s) => getZwanzigerrufenHint(s as ZwanzigerrufenResponse),
  troggu: (s) => getTrogguHint(s as TrogguResponse),
  horse: (s) => getHorseHint(s as HorseResponse),
  andarbahar: (s) => getAndarbaharHint(s as AndarBaharResponse),
  botifarra: (s) => getBotifarraHint(s as BotifarraResponse),
  rikken: (s) => getRikkenHint(s as RikkenResponse),
  chemindefer: (s) => getChemindeferHint(s as ChemindeFerResponse),
  crazyfourpoker: (s) => getCrazyfourpokerHint(s as CrazyFourPokerResponse),
  doubleattack: (s) => getDoubleattackHint(s as DoubleAttackResponse),
  freebet: (s) => getFreebetHint(s as FreeBetResponse),
  banluck: (s) => getBanluckHint(s as BanLuckResponse),
  montebank: (s) => getMontebankHint(s as MonteBankResponse),
  speculation: (s) => getSpeculationHint(s as SpeculationResponse),
  cincinnati: (s) => getCincinnatiHint(s as CincinnatiResponse),
  ironcross: (s) => getIroncrossHint(s as IronCrossResponse),
  baseballpoker: (s) => getBaseballpokerHint(s as BaseballPokerResponse),
  colourwhist: (s) => getColourwhistHint(s as ColourWhistResponse),
  dragontiger: (s) => getDragontigerHint(s as DragonTigerResponse),
  blackjackswitch: (s) => getBlackjackswitchHint(s as BlackJackSwitchResponse),
  war: (s) => getWarHint(s as WarResponse),
  vint: (s) => getVintHint(s as VintResponse),
  watten: (s) => getWattenHint(s as WattenResponse),
  whist: (s) => getWhistHint(s as WhistResponse),
  catchten: (s) => getCatchTenHint(s as CatchTenResponse),
  yukon: (s) => getYukonHint(s as YukonResponse),
  alaska: (s) => getAlaskaHint(s as AlaskaResponse),
  russiansolitaire: (s) => getRussianSolitaireHint(s as RussianSolitaireResponse),
  cruel: (s) => getCruelHint(s as CruelResponse),
  scopone: (s) => getScoponeHint(s as ScoponeResponse),
  scorpion: (s) => getScorpionHint(s as ScorpionResponse),
  wasp: (s) => getWaspHint(s as WaspResponse),
  easthaven: (s) => getEasthavenHint(s as EasthavenResponse),
  accordion: (s) => getAccordionHint(s as AccordionResponse),
  russianbank: (s) => getRussianBankHint(s as RussianBankResponse),
  briscola: (s) => getBriscolaHint(s as BriscolaResponse),
  brusquembille: (s) => getBrusquembilleHint(s as BrusquembilleResponse),
  acesup: (s) => getAcesUpHint(s as AcesUpResponse),
  blackhole: (s) => getBlackHoleHint(s as BlackHoleResponse),
  curdsandwhey: (s) => getCurdsAndWheyHint(s as CurdsAndWheyResponse),
  simplesimon: (s) => getSimpleSimonHint(s as SimpleSimonResponse),
  literature: (s) => getLiteratureHint(s as LiteratureResponse),
  labellelucie: (s) => getLaBelleLucieHint(s as LaBelleLucieResponse),
  shamrocks: (s) => getShamrocksHint(s as ShamrocksResponse),
  doubleklondike: (s) => getDoubleKlondikeHint(s as DoubleKlondikeResponse),
  calculation: (s) => getCalculationHint(s as CalculationResponse),
  sirtommy: (s) => getSirTommyHint(s as SirTommyResponse),
  auldlangsyne: (s) => getAuldLangSyneHint(s as AuldLangSyneResponse),
  bisley: (s) => getBisleyHint(s as BisleyResponse),
  napoleonssquare: (s) => getNapoleonsSquareHint(s as NapoleonsSquareResponse),
  grandfathersclock: (s) => getGrandfathersClockHint(s as GrandfathersClockResponse),
  bigben: (s) => getBigBenHint(s as BigBenResponse),
  duchess: (s) => getDuchessHint(s as DuchessResponse),
  windmill: (s) => getWindmillHint(s as WindmillResponse),
  americantoad: (s) => getAmericanToadHint(s as AmericanToadResponse),
  congress: (s) => getCongressHint(s as CongressResponse),
  saliclaw: (s) => getSalicLawHint(s as SalicLawResponse),
  terrace: (s) => getTerraceHint(s as TerraceResponse),
  braid: (s) => getBraidHint(s as BraidResponse),
  missmilligan: (s) => getMissMilliganHint(s as MissMilliganResponse),
  sevenbridge: (s) => getSevenbridgeHint(s as SevenBridgeResponse),
  sheepshead: (s) => getSheepsheadHint(s as SheepsheadResponse),
  doppelkopf: (s) => getDoppelkopfHint(s as DoppelkopfResponse),
  trash: (s) => getTrashHint(s as TrashResponse),
  president: (s) => getPresidentHint(s as PresidentResponse),
  cassino: (s) => getCassinoHint(s as CassinoResponse),
  scopa: (s) => getScopaHint(s as ScopaResponse),
  barbu: (s) => getBarbuHint(s as BarbuResponse),
  macau: (s) => getMacauHint(s as MacauResponse),
  clocksolitaire: (s) => getClocksolitaireHint(s as ClockSolitaireResponse),
  spiteandmalice: (s) => getSpiteAndMaliceHint(s as SpiteAndMaliceResponse),
  ramsch: (s) => getRamschHint(s as RamschResponse),
  skat: (s) => getSkatHint(s as SkatResponse),
  shithead: (s) => getShitheadHint(s as ShitheadResponse),
  nertz: (s) => getNertzHint(s as NertzResponse),
  slapjack: (s) => getSlapjackHint(s as SlapjackResponse),
  egyptianratscrew: (s) => getEgyptianRatscrewHint(s as EgyptianRatscrewResponse),
  contractrummy: (s) => getContractRummyHint(s as ContractRummyResponse),
  carioca: (s) => getCariocaHint(s as CariocaResponse),
  crescent: (s) => getCrescentHint(s as CrescentResponse),
  sthelena: (s) => getStHelenaHint(s as StHelenaResponse),
  spiderette: (s) => getSpideretteHint(s as SpideretteResponse),
  gaps: (s) => getGapsHint(s as GapsResponse),
  fourcardpoker: (s) => getFourCardPokerHint(s as FourCardPokerResponse),
  rummy500: (s) => getRummy500Hint(s as Rummy500Response),
  russianpoker: (s) => getRussianPokerHint(s as RussianPokerResponse),
  sixcardgolf: (s) => getSixcardgolfHint(s as SixCardGolfResponse),
  doudizhu: (s) => getDoudizhuHint(s as DoudizhuResponse),
  bidwhist: (s) => getBidWhistHint(s as BidWhistResponse),
  mus: (s) => getMusHint(s as MusResponse),
  tute: (s) => getTuteHint(s as TuteResponse),
  sueca: (s) => getSuecaHint(s as SuecaResponse),
  kaiser: (s) => getKaiserHint(s as KaiserResponse),
  kalooki: (s) => getKalookiHint(s as KalookiResponse),
  kemps: (s) => getKempsHint(s as KempsResponse),
  klaberjass: (s) => getKlaberjassHint(s as KlaberjassResponse),
  klaverjas: (s) => getKlaverjasHint(s as KlaverjasResponse),
  manille: (s) => getManilleHint(s as ManilleResponse),
  mao: (s) => getMaoHint(s as MaoResponse),
  marias: (s) => getMariasHint(s as MariasResponse),
  tysiac: (s) => getTysiacHint(s as TysiacResponse),
  calabresella: (s) => getCalabresellaHint(s as CalabresellaResponse),
  ombre: (s) => getOmbreHint(s as OmbreResponse),
  quadrille: (s) => getQuadrilleHint(s as QuadrilleResponse),
  ulti: (s) => getUltiHint(s as UltiResponse),
  scarto: (s) => getScartoHint(s as ScartoResponse),
  shengji: (s) => getShengJiHint(s as ShengJiResponse),
  cego: (s) => getCegoHint(s as CegoResponse),
  frenchtarot: (s) => getFrenchTarotHint(s as FrenchTarotResponse),
  koenigrufen: (s) => getKoenigrufenHint(s as KoenigrufenResponse),
  cinch: (s) => getCinchHint(s as CinchResponse),
  loo: (s) => getLooHint(s as LooResponse),
  basra: (s) => getBasraHint(s as BasraResponse),
  hachihachi: (s) => getHachiHachiHint(s as HachiHachiResponse),
  koikoi: (s) => getKoiKoiHint(s as KoiKoiResponse),
  gostop: (s) => getGoStopHint(s as GoStopResponse),
  tablanet: (s) => getTablanetHint(s as TablanetResponse),
  trenteetquarante: (s) => getTrenteEtQuaranteHint(s as TrenteEtQuaranteResponse),
  sedma: (s) => getSedmaHint(s as SedmaResponse),
  knockoutwhist: (s) => getKnockoutWhistHint(s as KnockoutWhistResponse),
  spoilfive: (s) => getSpoilFiveHint(s as SpoilFiveResponse),
  solowhist: (s) => getSoloWhistHint(s as SoloWhistResponse),
  fortyfives: (s) => getFortyFivesHint(s as FortyFivesResponse),
  nap: (s) => getNapHint(s as NapResponse),
  ganjifa: (s) => getGanjifaHint(s as GanjifaResponse),
  aluette: (s) => getAluetteHint(s as AluetteResponse),
  minchiate: (s) => getMinchiateHint(s as MinchiateResponse),
  tarocchini: (s) => getTarocchiniHint(s as TarocchiniResponse),
  preference: (s) => getPreferenceHint(s as PreferenceResponse),
  vira: (s) => getViraHint(s as ViraResponse),
  twentynine: (s) => getTwentyNineHint(s as TwentyNineResponse),
  courtpiece: (s) => getCourtPieceHint(s as CourtPieceResponse),
  bezique: (s) => getBeziqueHint(s as BeziqueResponse),
  ecarte: (s) => getEcarteHint(s as EcarteResponse),
  threecardbrag: (s) => getThreeCardBragHint(s as ThreeCardBragResponse),
  teenpatti: (s) => getTeenPattiHint(s as TeenPattiResponse),
  beggarmyneighbour: (s) => getBeggarMyNeighbourHint(s as BeggarMyNeighbourResponse),
  allfours: (s) => getAllFoursHint(s as AllFoursResponse),
  guts: (s) => getGutsHint(s as GutsResponse),
  seventwentyseven: (s) => getSevenTwentySevenHint(s as SevenTwentySevenResponse),
  anaconda: (s) => getAnacondaHint(s as AnacondaResponse),
  bura: (s) => getBuraHint(s as BuraResponse),
  mushi: (s) => getMushiHint(s as MushiResponse),
  toepen: (s) => getToepenHint(s as ToepenResponse),
  chineseten: (s) => getChineseTenHint(s as ChineseTenResponse),
  skitgubbe: (s) => getSkitgubbeHint(s as SkitgubbeResponse),
  laughandliedown: (s) => getLaughAndLieDownHint(s as LaughAndLieDownResponse),
  sjavs: (s) => getSjavsHint(s as SjavsResponse),
  trex: (s) => getTrexHint(s as TrexResponse),
  loba: (s) => getLobaHint(s as LobaResponse),
  desmoche: (s) => getDesmocheHint(s as DesmocheResponse),
  zwicker: (s) => getZwickerHint(s as ZwickerResponse),
  poch: (s) => getPochHint(s as PochResponse),
  popejoan: (s) => getPopeJoanHint(s as PopeJoanResponse),
  nainjaune: (s) => getNainJauneHint(s as NainJauneResponse),
  bouillotte: (s) => getBouillotteHint(s as BouillotteResponse),
  primero: (s) => getPrimeroHint(s as PrimeroResponse),
  michigan: (s) => getMichiganHint(s as MichiganResponse),
  pan: (s) => getPanHint(s as PanResponse),
  put: (s) => getPutHint(s as PutResponse),
  truco: (s) => getTrucoHint(s as TrucoResponse),
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

  // **言語も依存に入れる。**Nertz のようにゾーン名を訳して `reasonParams` に
  // 焼き込むファクトリがあり、言語を切り替えたときにテンプレートだけ新しい言語で
  // 埋め込みが古いまま、という混在文が出る (#4885 のレビュー指摘)。
  const { i18n } = useTranslation();
  const language = i18n.language;
  const hint = useMemo(() => {
    if (!hintEnabled || !state) return null;
    // `language` を読むのは、訳語を `reasonParams` に焼き込むファクトリ (Nertz) を
    // 言語切り替えで作り直すため。値そのものは使わない。
    void language;
    return hintFactories[gameName]?.(state) ?? null;
  }, [gameName, hintEnabled, state, language]);

  return { hintEnabled, setHintEnabled, hint };
}
