/** Card suit design identifier. */
export type CardDesign = 'SPADE' | 'CLOVER' | 'HEART' | 'DIAMOND' | 'JOKER';

/**
 * Server response i18n fields common to every game's `*Response` type.
 * Mirrors the Go backend's `WebOutputBase` (internal/adapter/controller/
 * web_output_base.go): every game's WebOutput embeds it, so every game's
 * frontend Response extends this. See issue #2098.
 */
export interface BaseGameResponse {
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
}

/**
 * A playing card with suit design and numeric value.
 *
 * Standard 52-card French-deck cards carry only `design` + `value` and render
 * via a static PNG (`/images/{prefix}{NN}.png`). Cards from non-52 decks
 * (tarot, hanafuda, kabu, Wizard, …) have no PNG art, so the backend
 * additionally sends a self-describing face descriptor (`glyph`/`label`/
 * `color`/`deck`) and the frontend draws them procedurally via `CardFace`.
 * When `deck` is set, `CardImage` switches to the procedural path. See
 * ADR-0033.
 */
export interface Card {
  design: CardDesign;
  value: number;
  /** Center face symbol for procedurally-drawn cards (e.g. "✦"). */
  glyph?: string;
  /** Corner rank/name label for procedurally-drawn cards (e.g. "Wizard"). */
  label?: string;
  /** Color tint token (e.g. "red", "black", "purple", "green"). */
  color?: string;
  /** Deck family id (e.g. "wizard"); when set, the card renders procedurally. */
  deck?: string;
}

/** A face-down card sentinel returned by the backend when the card must remain hidden
 * (e.g., dealer's hole cards in Caribbean Stud during the action phase). */
export interface MaskedCard {
  design: '';
  value: 0;
}

/** Type guard distinguishing a face-down `MaskedCard` from a revealed `Card`. */
export function isMaskedCard(card: Card | MaskedCard): card is MaskedCard {
  return card.design === '';
}

/** Bracket data for betting profile export. */
export interface BettingProfileBracketData {
  aggressive: number;
  total: number;
}

/** Exported betting human profile data (Poker/Holdem/Omaha). */
export interface BettingHumanProfileData {
  aggressiveByBracket: [BettingProfileBracketData, BettingProfileBracketData, BettingProfileBracketData];
  foldToBetCount: number;
  foldToBetTotal: number;
  gamesPlayed: number;
  hesitationCount: number;
  hesitationMean: number;
  hesitationM2: number;
}

/** Bracket data for doubt profile export. */
export interface DoubtProfileBracketData {
  bluffs: number;
  total: number;
}

/** Exported doubt human profile data. */
export interface DoubtHumanProfileData {
  bluffsByBracket: [DoubtProfileBracketData, DoubtProfileBracketData, DoubtProfileBracketData];
  doubtCorrect: number;
  doubtTotal: number;
  gamesPlayed: number;
  hesitationCount: number;
  hesitationMean: number;
  hesitationM2: number;
}

/** Exported old maid human profile data. */
export interface OldMaidHumanProfileData {
  positionBuckets: [number, number, number];
  totalPicks: number;
  shuffleCount: number;
  drawCount: number;
  gamesPlayed: number;
}

/** Bracket data for Indian Poker profile export. */
export interface IndianPokerProfileBracketData {
  aggressive: number;
  total: number;
}

/** Exported Indian Poker human profile data. */
export interface IndianPokerHumanProfileData {
  aggressiveByBracket: [IndianPokerProfileBracketData, IndianPokerProfileBracketData, IndianPokerProfileBracketData];
  foldToBetCount: number;
  foldToBetTotal: number;
  gamesPlayed: number;
  hesitationCount: number;
  hesitationMean: number;
  hesitationM2: number;
}

/** Meta-AI statistics for betting games (Poker/Holdem/Omaha). */
export interface BettingMetaAI {
  enabled: boolean;
  gamesPlayed: number;
  bluffRate: number;
  foldRate: number;
  hesitationMean: number;
}

/** Meta-AI statistics for Indian Poker CPU adaptation. */
export interface IndianPokerMetaAI {
  enabled: boolean;
  gamesPlayed: number;
  bluffRate: number;
  foldRate: number;
  hesitationMean: number;
}

/** A single entry in the game action log. */
export interface ActionLogEntry {
  turnNumber: number;
  playerIdx: number;
  actionType: string;
  detail: string;
  cards?: Card[];
}

/** Response containing action log entries. */
export interface ActionLogResponse {
  entries: ActionLogEntry[];
}

/** A single BlackJack hand with score, cards, and status flags. */
export interface BlackJackHand {
  score: number;
  cards: Card[];
  bet: number;
  stood: boolean;
  doubled: boolean;
  busted: boolean;
  isBlackJack: boolean;
  canSplit: boolean;
  surrendered: boolean;
  canSurrender: boolean;
}

/** BlackJack player (dealer or human) with chips and cards. */
export interface BlackJackPlayer {
  score?: number;
  cards?: Card[];
  chips: number;
}

/** BlackJack game phase (1=Bet, 2=Deal, 3=Insurance, 4=Action, 5=End, 6=EarlySurrender). */
export type BlackJackPhase = 1 | 2 | 3 | 4 | 5 | 6;

/** CPU player seat in BlackJack with chips and hands. */
export interface BlackJackCpuSeat {
  chips: number;
  hands: BlackJackHand[];
  insuranceBet: number;
}

/** Result of a BlackJack side bet (Perfect Pairs, 21+3). */
export interface BlackJackSideBetResult {
  betType: number;
  resultType: number;
  resultName: string;
  betAmount: number;
  payout: number;
}

/** Full BlackJack game state returned from the API. */
export interface BlackJackResponse extends BaseGameResponse {
  dealer: BlackJackPlayer;
  player: BlackJackPlayer;
  hands?: BlackJackHand[];
  currentHandIdx: number;
  phase: BlackJackPhase;
  insuranceBet: number;
  insuranceAvailable: boolean;
  hintEnabled: boolean;
  suggestedAction: number;
  deckCount: number;
  dealerHitsSoft17: boolean;
  countingEnabled: boolean;
  cpuPlayerCount: number;
  runningCount: number;
  trueCount: number;
  cpuPlayers?: BlackJackCpuSeat[];
  perfectPairsBet: number;
  twentyOnePlus3Bet: number;
  sideBetResults?: BlackJackSideBetResult[];
  /** i18n keys of variant bonuses achieved this round (Spanish 21, e.g. `spanish21.bonus.777.spade`). */
  bonuses?: string[];
  doubleAfterSplit: boolean;
  countingSystem: number;
  deckPenetration: number;
  multiHandCount: number;
  surrenderRule: number;
}

/** Poker player data including hand, chips, and status. */
export interface PokerPlayerData {
  id: number;
  isHuman: boolean;
  cards: Card[];
  chips: number;
  currentBet: number;
  folded: boolean;
  allIn: boolean;
  handRank: number;
  handName: string;
  exchangeCount: number;
  playStyleName: string;
}

/** CPU betting action in Poker. */
export interface PokerCpuAction {
  playerIdx: number;
  action: number;
  amount: number;
}

/** CPU card exchange result in Poker. */
export interface PokerCpuExchange {
  playerIdx: number;
  exchangeCount: number;
}

/** Poker round result for a single player. */
export interface PokerResult {
  playerIdx: number;
  handRank: number;
  handName: string;
  kickers: string;
  wonAmount: number;
}

/** Side pot in Poker with eligible players. */
export interface PokerSidePot {
  amount: number;
  eligiblePlayers: number[];
}

/** Probability of achieving a specific poker hand rank. */
export interface PokerOdds {
  handRank: number;
  handName: string;
  probability: number;
  count: number;
  total: number;
}

/** Poker game phase (0=Init, 1=Deal, 2=Exchange, 3=SecondBet, 4=End). */
export type PokerPhase = 0 | 1 | 2 | 3 | 4;

/** Full Poker game state returned from the API. */
export interface PokerResponse extends BaseGameResponse {
  players: PokerPlayerData[];
  pot: number;
  sidePots: PokerSidePot[];
  dealerIdx: number;
  currentTurn: number;
  phase: PokerPhase;
  gameEndFlag: boolean;
  lastBet: number;
  minRaise: number;
  ante: number;
  jokerCount: number;
  bettingLimit: number;
  raiseCount: number;
  maxBetAmount: number;
  roundResults: PokerResult[];
  cpuActions: PokerCpuAction[];
  cpuExchanges: PokerCpuExchange[];
  odds?: PokerOdds[];
  isLowball: boolean;
  metaAI?: BettingMetaAI;
  profile?: BettingHumanProfileData;
}

/** Badugi seat snapshot returned by the /badugi/exec API. */
export interface BadugiPlayerData {
  id: number;
  isHuman: boolean;
  cards: Card[];
  chips: number;
  currentBet: number;
  folded: boolean;
  allIn: boolean;
  /** BadugiHand.Size (1..4) after showdown, 0 otherwise. */
  handSize: number;
  handName: string;
  /** Cards exchanged in the most recent draw. */
  drawCount: number;
  /** Cumulative draws across all three draw rounds. */
  totalDraws: number;
  playStyleName: string;
  /** Best-subset selection revealed at showdown. */
  bestCards?: Card[];
}

/** CPU betting action in Badugi. */
export interface BadugiCpuAction {
  playerIdx: number;
  action: number;
  amount: number;
  drawIndex: number;
  roundLabel: string;
}

/** CPU draw result in Badugi. */
export interface BadugiCpuExchange {
  playerIdx: number;
  drawIndex: number;
  exchangeCount: number;
}

/** Badugi showdown result for a single player. */
export interface BadugiResult {
  playerIdx: number;
  handSize: number;
  handName: string;
  wonAmount: number;
}

/** Badugi side pot with eligible player seats. */
export interface BadugiSidePot {
  amount: number;
  eligiblePlayers: number[];
}

/** Meta-AI statistics for Badugi CPU adaptation. */
export interface BadugiMetaAI {
  enabled: boolean;
  gamesPlayed: number;
  bluffRate: number;
  foldRate: number;
  hesitationMean: number;
}

/** Badugi phase discriminator: 0 Init, 1 Deal, 2 Bet, 3 Draw, 4 Showdown, 5 End. */
export type BadugiPhaseId = 0 | 1 | 2 | 3 | 4 | 5;

/** Full Badugi game state returned by the API. */
export interface BadugiResponse extends BaseGameResponse {
  players: BadugiPlayerData[];
  pot: number;
  sidePots: BadugiSidePot[];
  dealerIdx: number;
  currentTurn: number;
  phase: BadugiPhaseId;
  drawIndex: number;
  gameEndFlag: boolean;
  lastBet: number;
  minRaise: number;
  ante: number;
  bettingLimit: number;
  raiseCount: number;
  maxBetAmount: number;
  roundResults: BadugiResult[];
  cpuActions: BadugiCpuAction[];
  cpuExchanges: BadugiCpuExchange[];
  metaAI?: BadugiMetaAI;
  profile?: BettingHumanProfileData;
}

/** 2-7 Triple Draw seat snapshot returned by the /deucetoseven/exec API. */
export interface DeuceToSevenPlayerData {
  id: number;
  isHuman: boolean;
  cards: Card[];
  chips: number;
  currentBet: number;
  folded: boolean;
  allIn: boolean;
  /** Poker category (0 High Card … 9 Royal Flush) after showdown, 0 otherwise. */
  handRank: number;
  handName: string;
  /** Cards exchanged in the most recent draw. */
  drawCount: number;
  /** Cumulative draws across all three draw rounds. */
  totalDraws: number;
  playStyleName: string;
}

/** CPU betting action in 2-7 Triple Draw. */
export interface DeuceToSevenCpuAction {
  playerIdx: number;
  action: number;
  amount: number;
  drawIndex: number;
  roundLabel: string;
}

/** CPU draw result in 2-7 Triple Draw. */
export interface DeuceToSevenCpuExchange {
  playerIdx: number;
  drawIndex: number;
  exchangeCount: number;
}

/** 2-7 Triple Draw showdown result for a single player. */
export interface DeuceToSevenResult {
  playerIdx: number;
  handRank: number;
  handName: string;
  wonAmount: number;
}

/** 2-7 Triple Draw side pot with eligible player seats. */
export interface DeuceToSevenSidePot {
  amount: number;
  eligiblePlayers: number[];
}

/** Meta-AI statistics for 2-7 Triple Draw CPU adaptation. */
export interface DeuceToSevenMetaAI {
  enabled: boolean;
  gamesPlayed: number;
  bluffRate: number;
  foldRate: number;
  hesitationMean: number;
}

/** 2-7 Triple Draw phase discriminator: 0 Init, 1 Deal, 2 Bet, 3 Draw, 4 Showdown, 5 End. */
export type DeuceToSevenPhaseId = 0 | 1 | 2 | 3 | 4 | 5;

/** Full 2-7 Triple Draw game state returned by the API. */
export interface DeuceToSevenResponse extends BaseGameResponse {
  players: DeuceToSevenPlayerData[];
  pot: number;
  sidePots: DeuceToSevenSidePot[];
  dealerIdx: number;
  currentTurn: number;
  phase: DeuceToSevenPhaseId;
  drawIndex: number;
  gameEndFlag: boolean;
  lastBet: number;
  minRaise: number;
  ante: number;
  bettingLimit: number;
  raiseCount: number;
  maxBetAmount: number;
  roundResults: DeuceToSevenResult[];
  cpuActions: DeuceToSevenCpuAction[];
  cpuExchanges: DeuceToSevenCpuExchange[];
  metaAI?: DeuceToSevenMetaAI;
  profile?: BettingHumanProfileData;
}

/** Old Maid player data with hand and finish status. */
export interface OldMaidPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  cardCount: number;
  cards: Card[];
}

/** CPU draw/discard action in Old Maid. */
export interface CpuAction {
  drawPlayerIdx: number;
  drawFromIdx: number;
  drawnCard: Card | null;
  discardedPairs: number;
  discardedCards?: Card[];
  hesitationMs?: number;
}

/** History entry for a card draw in Old Maid. */
export interface DrawHistoryEntry {
  drawPlayerIdx: number;
  drawFromIdx: number;
  discardedPairs: number;
  drawerFinished: boolean;
  targetFinished: boolean;
}

/** Meta-AI statistics for Old Maid CPU adaptation. */
export interface OldMaidMetaAI {
  enabled: boolean;
  gamesPlayed: number;
  edgePickRate: number;
}

/** Full Old Maid game state returned from the API. */
export interface OldMaidResponse extends BaseGameResponse {
  players: OldMaidPlayerData[];
  currentTurn: number;
  nextDrawTargetIdx: number;
  gameEndFlag: boolean;
  hasDrawn: boolean;
  lastDrawPlayerIdx: number;
  lastDrawFromIdx: number;
  lastDrawCard: Card | null;
  lastDiscardedPairs: number;
  lastDiscardedCards?: Card[];
  cpuActions: CpuAction[];
  humanAction?: CpuAction | null;
  drawHistory: DrawHistoryEntry[];
  cpuHighlightedCardIdx: number;
  removedCard: Card | null;
  mode: number;
  metaAI?: OldMaidMetaAI;
  profile?: OldMaidHumanProfileData;
}

/** Daifugo player data with rank and card count. */
export interface DaifugoPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  rank: number;
  cardCount: number;
  cards: Card[];
  illegalFinishPenalty?: boolean;
}

/** A play or pass action in Daifugo. */
export interface DaifugoAction {
  playerIdx: number;
  playedCards: Card[] | null; // null = pass
}

/** Daifugo game rule configuration. */
export interface DaifugoConfig {
  jokerCount: number;
  eightCutEnabled: boolean;
  suitLockMode: number;
  elevenBackEnabled: boolean;
  sequenceEnabled: boolean;
  cardExchangeEnabled: boolean;
  blindExchangeEnabled: boolean;
  fiveSkipEnabled: boolean;
  fiveSkipCount: number;
  sevenPassEnabled: boolean;
  tenDiscardEnabled: boolean;
  spadeThreeEnabled: boolean;
  capitalFallEnabled: boolean;
  nineReverseEnabled: boolean;
  coupDetatEnabled: boolean;
  numberLockEnabled: boolean;
  sandstormEnabled: boolean;
  emperorEnabled: boolean;
  sequenceRevolutionEnabled: boolean;
  sequenceLockEnabled: boolean;
  illegalFinishEnabled: boolean;
  queenBomberEnabled: boolean;
  cpuDifficulty: number;
}

/** Input type alias for Daifugo configuration. */
export type DaifugoConfigInput = DaifugoConfig;

/** Card exchange action between ranked players in Daifugo. */
export interface DaifugoExchangeAction {
  fromPlayerIdx: number;
  toPlayerIdx: number;
  cards: Card[];
}

/** Full Daifugo game state returned from the API. */
export interface DaifugoResponse extends BaseGameResponse {
  players: DaifugoPlayerData[];
  currentTurn: number;
  tableCards: Card[];
  lastPlayPlayerIdx: number;
  gameEndFlag: boolean;
  revolutionActive: boolean;
  elevenBackActive: boolean;
  suitLocked: boolean;
  lockedSuit: string;
  tableIsSequence: boolean;
  config: DaifugoConfig;
  exchangeActions: DaifugoExchangeAction[];
  cpuActions: DaifugoAction[];
  humanAction: DaifugoAction | null;
  pendingAction: 'none' | 'sevenPass' | 'tenDiscard' | 'queenBomber';
  pendingActionTarget: number;
  reverseDirection: boolean;
  numberLocked: boolean;
  sequenceLocked: boolean;
  sortMode: number;
}

/** Big Two player data. */
export interface BigTwoPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  rank: number;
  cardCount: number;
  cards: Card[];
}

/** A play or pass action in Big Two. */
export interface BigTwoAction {
  playerIdx: number;
  playedCards: Card[] | null;
}

/** Big Two game rule configuration. */
export interface BigTwoConfig {
  cpuDifficulty: number;
}

/** Input type alias for Big Two configuration. */
export type BigTwoConfigInput = BigTwoConfig;

/** Full Big Two game state returned from the API. */
export interface BigTwoResponse extends BaseGameResponse {
  players: BigTwoPlayerData[];
  currentTurn: number;
  tableCards: Card[];
  tablePlayType: number;
  lastPlayPlayerIdx: number;
  gameEndFlag: boolean;
  cpuActions: BigTwoAction[];
  humanAction: BigTwoAction | null;
  config: BigTwoConfig;
}

/** Tien Len player data. */
export interface TienLenPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  rank: number;
  cardCount: number;
  cards: Card[];
}

/** A play or pass action in Tien Len. */
export interface TienLenAction {
  playerIdx: number;
  playedCards: Card[] | null;
}

/** Tien Len game rule configuration. */
export interface TienLenConfig {
  cpuDifficulty: number;
}

/** Input type alias for Tien Len configuration. */
export type TienLenConfigInput = TienLenConfig;

/** Full Tien Len game state returned from the API. */
export interface TienLenResponse extends BaseGameResponse {
  players: TienLenPlayerData[];
  currentTurn: number;
  tableCards: Card[];
  tablePlayType: number;
  lastPlayPlayerIdx: number;
  gameEndFlag: boolean;
  cpuActions: TienLenAction[];
  humanAction: TienLenAction | null;
  config: TienLenConfig;
}

/** Zheng Shangyou player data. */
export interface ZhengPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  rank: number;
  cardCount: number;
  cards: Card[];
}

/** A play or pass action in Zheng Shangyou (null playedCards = pass). */
export interface ZhengAction {
  playerIdx: number;
  playedCards: Card[] | null;
}

/** Zheng Shangyou game rule configuration. */
export interface ZhengConfig {
  cpuDifficulty: number;
}

/** Input type alias for Zheng Shangyou configuration. */
export type ZhengConfigInput = ZhengConfig;

/** Full Zheng Shangyou game state returned from the API. */
export interface ZhengResponse extends BaseGameResponse {
  players: ZhengPlayerData[];
  currentTurn: number;
  tableCards: Card[];
  tablePlayType: number;
  lastPlayPlayerIdx: number;
  gameEndFlag: boolean;
  cpuActions: ZhengAction[];
  humanAction: ZhengAction | null;
  config: ZhengConfig;
}

/** Sevens player data with pass count and card info. */
export interface SevensPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  rank: number;
  cardCount: number;
  passesUsed: number;
  maxPasses: number;
  cards: Card[];
  lastPlayedJoker: boolean;
}

/** A play or pass action in Sevens. */
export interface SevensAction {
  playerIdx: number;
  playedCard: Card | null; // null = pass
  targetSuit: number;
  targetValue: number;
  forcedPass: boolean;
}

/** Sevens game rule configuration. */
export interface SevensConfig {
  tunnelEnabled: boolean;
  tunnelSkipWidth: number;
  jokerCount: number;
  cpuStrategy: number;
  maxPasses: number;
  noJokerFinish: boolean;
  jokerReclaimEnabled: boolean;
  endStopEnabled: boolean;
  jokerConsecutiveBanned: boolean;
}

/** Full Sevens game state returned from the API. */
export interface SevensResponse extends BaseGameResponse {
  players: SevensPlayerData[];
  currentTurn: number;
  tableMinVals: number[]; // index 0 unused; 1=SPADE, 2=CLOVER, 3=HEART, 4=DIAMOND
  tableMaxVals: number[];
  tablePlaced: number[]; // bitmask per suit; bit i = value i placed
  config: SevensConfig;
  gameEndFlag: boolean;
  cpuActions: SevensAction[];
  humanAction: SevensAction | null;
}

/** Doubt player data with card count and finish status. */
export interface DoubtPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  cardCount: number;
  cards: Card[];
}

/** CPU play action in Doubt with bluff information. */
export interface DoubtCpuAction {
  playerIdx: number;
  claimedValue: number;
  cardCount: number;
  isBluff: boolean;
  hasTell?: boolean;
  hesitationMs?: number;
}

/** Result of a doubt challenge in Doubt. */
export interface DoubtDoubtResult {
  doubterIdx: number;
  cardPlayerIdx: number;
  wasLying: boolean;
  loserIdx: number;
  cardCount: number;
  discardedCount: number;
  revealedCards: Card[];
}

/** Doubt game configuration options. */
export interface DoubtConfig {
  doubtWindowSec: number;
  cpuMemoryLevel: number; // 0=Easy, 1=Normal, 2=Hard
  penaltyDrawLimit: number; // 0=unlimited, >0=max cards loser picks up
  cpuHesitationEnabled: boolean;
  cpuMetaAI: boolean;
}

/** Full Doubt game state returned from the API. */
export interface DoubtResponse extends BaseGameResponse {
  players: DoubtPlayerData[];
  currentTurn: number;
  phase: 0 | 1 | 2; // 0=Play, 1=Doubt, 2=End
  tableCardCount: number;
  lastAction: DoubtCpuAction | null;
  cpuDoubters: number[];
  cpuActions: DoubtCpuAction[];
  humanAction: DoubtCpuAction | null;
  lastDoubtResult: DoubtDoubtResult | null;
  gameEndFlag: boolean;
  winnerIdx: number;
  doubtWindowSec: number;
  penaltyDrawLimit: number;
  metaAI?: DoubtMetaAI;
  profile?: DoubtHumanProfileData;
}

/** Meta-AI statistics for Doubt CPU adaptation. */
export interface DoubtMetaAI {
  enabled: boolean;
  gamesPlayed: number;
  bluffRate: number;
  doubtAccuracy: number;
  hesitationMean: number;
}

/** Texas Hold'em player data with stats and hand info. */
export interface HoldemPlayerData {
  id: number;
  isHuman: boolean;
  cards: Card[];
  chips: number;
  currentBet: number;
  folded: boolean;
  allIn: boolean;
  handRank: number;
  handName: string;
  bestHand: Card[];
  /** Best 5-card low hand (Omaha Hi-Lo only; populated at showdown when qualified). */
  lowBestHand?: Card[];
  /** True if the player has a qualifying low hand (Omaha Hi-Lo only). */
  lowQualifies?: boolean;
  playStyleName: string;
  totalHands: number;
  vpip: number;
  pfr: number;
  threeBet: number;
  af: string;
}

/** CPU betting action in Texas Hold'em. */
export interface HoldemCpuAction {
  playerIdx: number;
  action: number;
  amount: number;
}

/** Hold'em round result for a single player.
 *
 * Hi-Lo (Omaha 8 or Better) split-pot games populate the optional Low* and
 * HiWonAmount/LowWonAmount fields; for non-Hi-Lo games they are absent. */
export interface HoldemResult {
  playerIdx: number;
  handRank: number;
  handName: string;
  kickers: string;
  bestHand: Card[];
  wonAmount: number;
  mucked: boolean;
  lowBestHand?: Card[];
  lowKickers?: string;
  lowQualifies?: boolean;
  hiWonAmount?: number;
  lowWonAmount?: number;
}

/** Side pot in Hold'em with eligible players. */
export interface HoldemSidePot {
  amount: number;
  eligiblePlayers: number[];
}

/** Full Texas Hold'em game state returned from the API. */
export interface HoldemResponse extends BaseGameResponse {
  players: HoldemPlayerData[];
  communityCards: Card[];
  pot: number;
  sidePots: HoldemSidePot[];
  dealerIdx: number;
  currentTurn: number;
  phase: number;
  gameEndFlag: boolean;
  lastBet: number;
  minRaise: number;
  bettingLimit: number;
  raiseCount: number;
  maxBetAmount: number;
  roundResults: HoldemResult[];
  cpuActions: HoldemCpuAction[];
  handCount: number;
  smallBlind: number;
  bigBlind: number;
  tournamentMode: boolean;
  blindLevelHands: number;
  blindMultiplier: number;
  tableSize: number;
  rebuyPhaseType: number;
  rebuyChips: number;
  rebuyMaxCount: number;
  rebuyCounts: number[];
  addonChips: number;
  rebuyAvailable: boolean;
  addonAvailable: boolean;
  rebuyEnabled: boolean;
  addonEnabled: boolean;
  rebuyPeriodHands: number;
  addonAfterHand: number;
  addonUsed: boolean[];
  muckAvailable: boolean;
  /** True when the variant is Omaha Hi-Lo (8 or Better) — split-pot logic active. */
  isHiLo?: boolean;
  equity?: HoldemEquity;
  potOdds?: number;
  metaAI?: BettingMetaAI;
  profile?: BettingHumanProfileData;
}

/** Equity calculation result for Hold'em hand. */
export interface HoldemEquity {
  winProbability: number;
  handOdds: HoldemHandOdds[];
}

/** Probability of achieving a specific hand rank in Hold'em. */
export interface HoldemHandOdds {
  handRank: number;
  handName: string;
  probability: number;
}

// --- Pineapple Poker ---

/** Pineapple Poker response extending Hold'em with discard phase fields. */
export interface PineappleResponse extends HoldemResponse {
  isDiscardPhase: boolean;
  discardDone: boolean[];
  initialDealCount: number;
}

/** Irish Poker shares the same response shape as Pineapple. */
export type IrishPokerResponse = PineappleResponse;

// --- Omaha Hold'em ---
// Omaha shares identical response/player structures with Holdem
/** Omaha player data (same structure as Hold'em). */
export type OmahaPlayerData = HoldemPlayerData;
/** Omaha CPU action (same structure as Hold'em). */
export type OmahaCpuAction = HoldemCpuAction;
/** Omaha round result (same structure as Hold'em). */
export type OmahaResult = HoldemResult;
/** Omaha side pot (same structure as Hold'em). */
export type OmahaSidePot = HoldemSidePot;
/** Omaha equity (same structure as Hold'em). */
export type OmahaEquity = HoldemEquity;
/** Omaha hand odds (same structure as Hold'em). */
export type OmahaHandOdds = HoldemHandOdds;
/** Omaha response (same structure as Hold'em). */
export type OmahaResponse = HoldemResponse;

// --- Short Deck Hold'em ---

/** Short Deck Hold'em player data (same structure as Hold'em). */
export type ShortDeckPlayerData = HoldemPlayerData;
/** Short Deck Hold'em side pot (same structure as Hold'em). */
export type ShortDeckSidePot = HoldemSidePot;
/** Short Deck Hold'em equity (same structure as Hold'em). */
export type ShortDeckEquity = HoldemEquity;
/** Short Deck Hold'em hand odds (same structure as Hold'em). */
export type ShortDeckHandOdds = HoldemHandOdds;
/** Short Deck Hold'em response (same structure as Hold'em). */
export type ShortDeckResponse = HoldemResponse;

// --- Hearts ---

/** Hearts player data with scores and trick count. */
export interface HeartsPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
  /** Captured penalty cards so far: every heart plus the Q♠ (J♦ excluded). */
  penaltyCards: Card[];
}

/** A card played in a Hearts trick. */
export interface HeartsTrickCard {
  playerIdx: number;
  card: Card;
}

/** Hearts game configuration. */
export interface HeartsConfig {
  cpuDifficulty: number;
  pointLimit: number;
  omnibusJD: boolean;
}

/** A suggested hint for Hearts. */
export interface HeartsHint {
  cardIndices: number[];
  reason: string;
}

/** Full Hearts game state returned from the API. */
export interface HeartsResponse extends BaseGameResponse {
  players: HeartsPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  currentTrick: HeartsTrickCard[];
  heartsBroken: boolean;
  passDirection: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  leadPlayerIdx: number;
  config: HeartsConfig;
  hint?: HeartsHint;
}

// --- Gong Zhu (拱猪) ---

/** Gong Zhu player data with scores and trick count. */
export interface GongZhuPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  /** Point cards this player has captured so far (hearts, ♠Q pig, ♦J sheep, ♣10 doubler). Public info revealed as tricks are taken. */
  capturedPointCards: Card[];
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
}

/** A card played in a Gong Zhu trick. */
export interface GongZhuTrickCard {
  playerIdx: number;
  card: Card;
}

/** Which point cards have been exposed (stakes doubled). */
export interface GongZhuExposure {
  pig: boolean;
  sheep: boolean;
  ace: boolean;
  doubler: boolean;
}

/** Gong Zhu game configuration. */
export interface GongZhuConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A suggested hint for Gong Zhu. */
export interface GongZhuHint {
  cardIndices: number[];
  reason: string;
}

/** Full Gong Zhu game state returned from the API. */
export interface GongZhuResponse extends BaseGameResponse {
  players: GongZhuPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  currentTrick: GongZhuTrickCard[];
  heartsBroken: boolean;
  exposed: GongZhuExposure;
  exposableIndices: number[];
  gameEndFlag: boolean;
  winnerIdx: number;
  leadPlayerIdx: number;
  config: GongZhuConfig;
  hint?: GongZhuHint;
}

// --- Tressette ---

/** A Tressette player's public/own state. */
export interface TressettePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  teamId: number;
}

/** A card played in a Tressette trick. */
export interface TressetteTrickCard {
  playerIdx: number;
  card: Card;
}

/** Tressette game configuration. */
export interface TressetteConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Tressette. */
export interface TressetteHint {
  cardIndices: number[];
  reason: string;
}

/** Full Tressette game state returned from the API. */
export interface TressetteResponse extends BaseGameResponse {
  players: TressettePlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  currentTrick: TressetteTrickCard[];
  lastTrick: TressetteTrickCard[];
  lastTrickWinner: number;
  leadPlayerIdx: number;
  teamScores: number[];
  teamRoundThirds: number[];
  playableIndices: number[];
  gameEndFlag: boolean;
  winnerTeam: number;
  config: TressetteConfig;
  hint?: TressetteHint;
}

// --- Sheepshead ---

/** Sheepshead game phase (0=Pick 1=Bury 2=Call 3=Play 4=TrickEnd 5=RoundEnd 6=GameEnd). */
export type SheepsheadPhase = 0 | 1 | 2 | 3 | 4 | 5 | 6;

/** A Sheepshead player's public/own state. Cards are non-empty only for the human. */
export interface SheepsheadPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  chips: number;
}

/** A card played into the current Sheepshead trick. */
export interface SheepsheadTrickCard {
  playerIdx: number;
  card: Card;
}

/** Sheepshead game configuration. */
export interface SheepsheadConfig {
  cpuDifficulty: number;
  baseChips: number;
  startChips: number;
  targetChips: number;
}

/** A suggested hint for Sheepshead, computed by the backend. */
export interface SheepsheadHint {
  cardIndices: number[];
  /** Suggested called suit (0=none, 1=♠, 2=♣, 3=♥). Relevant in the Call phase. */
  suit: number;
  /** Whether the hint recommends picking the blind (Pick phase). */
  pick: boolean;
  reason: string;
}

/** Full Sheepshead game state returned from the API. */
export interface SheepsheadResponse extends BaseGameResponse {
  players: SheepsheadPlayer[];
  phase: SheepsheadPhase;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  currentTrick: SheepsheadTrickCard[];
  /** Number of cards in the blind (only the count is exposed during the Pick phase). */
  blindCount: number;
  /** The two buried cards; empty until RoundEnd/GameEnd. */
  buried: Card[];
  /** Index of the picker, or -1 until decided. */
  pickerIdx: number;
  /** Index of the picker's partner, or -1 until revealed/round end. */
  partnerIdx: number;
  /** Called partner suit (0=none, 1=♠, 2=♣, 3=♥). */
  calledSuit: number;
  /** Whether the partner has been revealed. */
  partnerRevealed: boolean;
  /** Number of players who have passed in the Pick phase. */
  passCount: number;
  /** Suits the picker may call this turn (non-empty only in the Call phase). */
  callableSuits: number[];
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  /** Card points captured by the picker's team this round. */
  roundPickerPoints: number;
  /** Score multiplier applied to this round's result. */
  roundMultiplier: number;
  /** Whether the picker's team won the round. */
  roundPickerWon: boolean;
  gameEndFlag: boolean;
  /** Winning player index, or -1 until the game ends. */
  winnerIdx: number;
  hint?: SheepsheadHint | null;
  config: SheepsheadConfig;
}

// --- Mus ---

/**
 * Mus game phase
 * (0=Mus 1=Discard 2=Grande 3=Chica 4=Pares 5=Juego 6=Showdown 7=RoundEnd 8=GameEnd).
 */
export type MusPhaseValue = 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8;

/**
 * A Mus player's public/own state. `cards` is populated for the human at all
 * times and for opponents only once the phase reaches Showdown (>=6).
 */
export interface MusPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  /** The score (amarrakos) of the team this player belongs to. */
  teamScore: number;
}

/**
 * Result of one of the four betting rounds (Grande / Chica / Pares / Juego).
 * `kind` identifies the round, `stake` the amarrakos awarded, `team` the winner.
 */
export interface MusRoundResult {
  kind: number;
  stake: number;
  team: number;
}

/** Mus game configuration. */
export interface MusConfig {
  cpuDifficulty: number;
  targetAmarrakos: number;
}

/** A suggested hint for Mus, computed by the backend. */
export interface MusHint {
  /** Whether the hint recommends calling Mus / exchanging (Mus phase). */
  mus: boolean;
  /** Suggested bet action (0=paso 1=envido 2=ordago 3=quiero 4=noquiero). */
  action: number;
  /** Suggested Envido amount. */
  amount: number;
  /** Suggested card indices to discard (Discard phase). */
  indices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Full Mus game state returned from the API. */
export interface MusResponse extends BaseGameResponse {
  players: MusPlayer[];
  phase: MusPhaseValue;
  roundNumber: number;
  /** Index of the mano (lead) player. */
  manoIdx: number;
  /** Team that currently holds the active bet, or -1 when none. */
  betTeam: number;
  /** Pending stake amount (-1=ordago/all-in, 0=none). */
  pendingStake: number;
  /** Team that placed the most recent bet, or -1. */
  lastBettorTeam: number;
  /** Index of the player to act in the Mus phase. */
  musTurn: number;
  /** Index of the player to act in the Discard phase. */
  discardTurn: number;
  /** Number of Mus/exchange cycles completed this round. */
  musCycle: number;
  /** Team amarrakos (scores) — [team0, team1]. */
  amarrakos: number[];
  /** Per-round results indexed by Grande/Chica/Pares/Juego. */
  results: MusRoundResult[];
  gameEndFlag: boolean;
  /** Winning team index, or -1 until the game ends. */
  winnerTeam: number;
  /** Team the human player belongs to, or -1 when none. */
  humanTeam: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  /** Whether the Paso bet action is legal for the human right now. */
  canPaso: boolean;
  /** Whether the Envido bet action is legal for the human right now. */
  canEnvido: boolean;
  /** Whether the Ordago (all-in) bet action is legal for the human right now. */
  canOrdago: boolean;
  /** Whether the Quiero (accept) bet action is legal for the human right now. */
  canQuiero: boolean;
  /** Whether the No Quiero (decline) bet action is legal for the human right now. */
  canNoQuiero: boolean;
  hint?: MusHint | null;
  config: MusConfig;
}

// --- Doppelkopf ---

/** Doppelkopf game phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd). */
export type DoppelkopfPhaseValue = 0 | 1 | 2 | 3;

/** A Doppelkopf player's public/own state. Cards are non-empty only for the human. */
export interface DoppelkopfPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  chips: number;
  /** Whether this player is on the Re team. False until teams are revealed. */
  isRe: boolean;
}

/** A card played into the current Doppelkopf trick. */
export interface DoppelkopfTrickCard {
  playerIdx: number;
  card: Card;
}

/** Doppelkopf game configuration. */
export interface DoppelkopfConfig {
  cpuDifficulty: number;
  baseChips: number;
  startChips: number;
  targetChips: number;
}

/** A suggested hint for Doppelkopf, computed by the backend. */
export interface DoppelkopfHint {
  cardIndices: number[];
  reason: string;
}

/** Full Doppelkopf game state returned from the API. */
export interface DoppelkopfResponse extends BaseGameResponse {
  players: DoppelkopfPlayer[];
  phase: DoppelkopfPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  currentTrick: DoppelkopfTrickCard[];
  /** Each player's Re-team membership; all false until teams are revealed (4 elements). */
  reTeam: boolean[];
  /** Whether one player holds both ♣Q (a solo Re). */
  soloRe: boolean;
  /** Whether the Re/Kontra teams have been revealed. */
  teamsRevealed: boolean;
  /** Whether Re has been announced this round. */
  reAnnounced: boolean;
  /** Whether Kontra has been announced this round. */
  kontraAnnounced: boolean;
  /** Whether the human may announce Re/Kontra right now (first trick only). */
  canAnnounce: boolean;
  /** Whether the human is on the Re team. Always known, even before reveal. */
  youAreRe: boolean;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  /** Card points captured by the Re team this round. */
  roundRePoints: number;
  /** Whether the Re team won the round. */
  roundReWon: boolean;
  /** Game points awarded for this round. */
  roundGamePoints: number;
  gameEndFlag: boolean;
  /** Winning player index, or -1 until the game ends. */
  winnerIdx: number;
  hint?: DoppelkopfHint | null;
  config: DoppelkopfConfig;
}

// --- Tute ---

/** Tute game phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd). */
export type TutePhaseValue = 0 | 1 | 2 | 3;

/** A Tute player's public/own state. Cards are non-empty only for the human. */
export interface TutePlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative score of the team this player belongs to. */
  teamScore: number;
}

/** A card played into the current Tute trick. */
export interface TuteTrickCard {
  playerIdx: number;
  card: Card;
}

/** Tute game configuration. */
export interface TuteConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Tute, computed by the backend. */
export interface TuteHint {
  cardIndices: number[];
  /** Suggested marriage-declaration suit (0=none, 1=♠ 2=♣ 3=♥ 4=♦). */
  marriage: number;
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Full Tute game state returned from the API. */
export interface TuteResponse extends BaseGameResponse {
  players: TutePlayer[];
  phase: TutePhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Trump suit (1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  currentTrick: TuteTrickCard[];
  /** Declared-marriage suits; valid indices 1-4 (index 0 unused, 5 elements). */
  declaredSuits: boolean[];
  /** Team scores — [team0, team1]. */
  teamScores: number[];
  /** Card points captured per team this round — [team0, team1]. */
  roundTeamPoints: number[];
  /** Whether the human may declare a marriage (K+Q) right now. */
  canDeclareMarriage: boolean;
  /** Whether the human may declare Tute (four Kings or four Queens) for an instant win. */
  canDeclareTute: boolean;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning team index, or -1 until the game ends. */
  winnerTeam: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: TuteHint | null;
  config: TuteConfig;
}

// --- Sueca ---

/** Sueca game phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd). */
export type SuecaPhaseValue = 0 | 1 | 2 | 3;

/** A Sueca player's public/own state. Cards are non-empty only for the human. */
export interface SuecaPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative game points of the team this player belongs to. */
  teamGamePoints: number;
}

/** A card played into the current Sueca trick. */
export interface SuecaTrickCard {
  playerIdx: number;
  card: Card;
}

/** Sueca game configuration. */
export interface SuecaConfig {
  cpuDifficulty: number;
  targetGamePoints: number;
}

/** A suggested hint for Sueca, computed by the backend. */
export interface SuecaHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Full Sueca game state returned from the API. */
export interface SuecaResponse extends BaseGameResponse {
  players: SuecaPlayer[];
  phase: SuecaPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Trump suit (1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  currentTrick: SuecaTrickCard[];
  /** Cumulative game points per team — [team0, team1]. */
  teamGamePoints: number[];
  /** Card points captured per team this round — [team0, team1]. */
  roundCardPoints: number[];
  /** Winning team of the most recent round, or -1 when undecided/draw. */
  roundWinnerTeam: number;
  /** Game points awarded for the most recent round. */
  roundGamePoints: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning team index, or -1 until the game ends. */
  winnerTeam: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: SuecaHint | null;
  config: SuecaConfig;
}

// --- Klaverjas ---

/** Klaverjas game phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd). */
export type KlaverjasPhaseValue = 0 | 1 | 2 | 3;

/** A Klaverjas player's public/own state. Cards are non-empty only for the human. */
export interface KlaverjasPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of the team this player belongs to. */
  teamScore: number;
}

/** A card played into the current Klaverjas trick. */
export interface KlaverjasTrickCard {
  playerIdx: number;
  card: Card;
}

/** Klaverjas game configuration. */
export interface KlaverjasConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Klaverjas, computed by the backend. */
export interface KlaverjasHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Full Klaverjas game state returned from the API. */
export interface KlaverjasResponse extends BaseGameResponse {
  players: KlaverjasPlayer[];
  phase: KlaverjasPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Trump suit (1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  currentTrick: KlaverjasTrickCard[];
  /** Cumulative match scores per team — [team0, team1]. */
  teamScores: number[];
  /** Card points captured per team this round — [team0, team1]. */
  roundCardPoints: number[];
  /** Roem (run/marriage) bonus points per team this round — [team0, team1]. */
  roundRoem: number[];
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning team index, or -1 until the game ends. */
  winnerTeam: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: KlaverjasHint | null;
  config: KlaverjasConfig;
}

// --- Manille ---

/** Manille game phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd). */
export type ManillePhaseValue = 0 | 1 | 2 | 3;

/** A Manille player's public/own state. Cards are non-empty only for the human. */
export interface ManillePlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of the team this player belongs to. */
  teamScore: number;
}

/** A card played into the current Manille trick. */
export interface ManilleTrickCard {
  playerIdx: number;
  card: Card;
}

/** Manille game configuration. */
export interface ManilleConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Manille, computed by the backend. */
export interface ManilleHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Full Manille game state returned from the API. */
export interface ManilleResponse extends BaseGameResponse {
  players: ManillePlayer[];
  phase: ManillePhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Trump suit (1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  currentTrick: ManilleTrickCard[];
  /** Cumulative match scores per team — [team0, team1]. */
  teamScores: number[];
  /** Card points captured per team this round — [team0, team1]. */
  roundCardPoints: number[];
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning team index, or -1 until the game ends. */
  winnerTeam: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: ManilleHint | null;
  config: ManilleConfig;
}

// --- Sedma ---

/** Sedma game phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd). */
export type SedmaPhaseValue = 0 | 1 | 2 | 3;

/** A Sedma player's public/own state. Cards are non-empty only for the human. */
export interface SedmaPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of the team this player belongs to. */
  teamScore: number;
}

/** A card played into the current Sedma trick. */
export interface SedmaTrickCard {
  playerIdx: number;
  card: Card;
}

/** Sedma game configuration. */
export interface SedmaConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Sedma, computed by the backend. */
export interface SedmaHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Sedma game state returned from the API.
 *
 * Sedma is a Czech/Slovak no-trump capture trick-taker, so — unlike the
 * Manille shape it mirrors — there is intentionally no `trumpSuit` field.
 */
export interface SedmaResponse extends BaseGameResponse {
  players: SedmaPlayer[];
  phase: SedmaPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  currentTrick: SedmaTrickCard[];
  /** Cumulative match scores per team — [team0, team1]. */
  teamScores: number[];
  /** Card points captured per team this round — [team0, team1]. */
  roundCardPoints: number[];
  /** Indices in the human's hand that are legal to play (every card is playable on the human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning team index, or -1 until the game ends. */
  winnerTeam: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: SedmaHint | null;
  config: SedmaConfig;
}

// --- Knockout Whist ---

/** Knockout Whist game phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd 4=TrumpSelect). */
export type KnockoutWhistPhaseValue = 0 | 1 | 2 | 3 | 4;

/** A Knockout Whist player's public/own state. Cards are non-empty only for the human. */
export interface KnockoutWhistPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  /** Total tricks taken across the match (cumulative). */
  trickCount: number;
  /** Whether this player has been knocked out of the match. */
  eliminated: boolean;
  /** Remaining Dogbone survival tokens (each player starts with 1). */
  dogbones: number;
  /** Tricks taken in the current round (resets each round). */
  roundTricks: number;
}

/** A card played into the current Knockout Whist trick. */
export interface KnockoutWhistTrickCard {
  playerIdx: number;
  card: Card;
}

/** Knockout Whist game configuration (CPU difficulty only — no target points). */
export interface KnockoutWhistConfig {
  cpuDifficulty: number;
}

/** A suggested hint for Knockout Whist, computed by the backend. */
export interface KnockoutWhistHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Knockout Whist game state returned from the API.
 *
 * Knockout Whist is a British play-only survival trick-taker: each round the
 * hand shrinks by one card, the previous round's winner's longest suit becomes
 * trump (auto), and a player who wins zero tricks must spend a Dogbone to
 * survive — or is eliminated. Last player standing wins.
 */
export interface KnockoutWhistResponse extends BaseGameResponse {
  players: KnockoutWhistPlayer[];
  phase: KnockoutWhistPhaseValue;
  roundNumber: number;
  /** Number of cards dealt this round (8 - roundNumber, down to 1). */
  handSize: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Trump suit (1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  /** Seat index of the round's winner, or -1. */
  roundWinnerIdx: number;
  currentTrick: KnockoutWhistTrickCard[];
  /** Number of players still in the match (not eliminated). */
  activeCount: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: KnockoutWhistHint | null;
  config: KnockoutWhistConfig;
}

// --- Spoil Five ---

/** Spoil Five game phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd). */
export type SpoilFivePhaseValue = 0 | 1 | 2 | 3;

/** A Spoil Five player's public/own state. Cards are non-empty only for the human. */
export interface SpoilFivePlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  /** Total tricks taken across the match (cumulative). */
  trickCount: number;
  /** Match points scored so far (first to targetPoints wins). */
  score: number;
  /** Tricks taken in the current round (resets each round; first to 3 takes the pot). */
  roundTricks: number;
}

/** A card played into the current Spoil Five trick. */
export interface SpoilFiveTrickCard {
  playerIdx: number;
  card: Card;
}

/** Spoil Five game configuration. */
export interface SpoilFiveConfig {
  cpuDifficulty: number;
  /** Match points needed to win (default 30). */
  targetPoints: number;
}

/** A suggested hint for Spoil Five, computed by the backend. */
export interface SpoilFiveHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Spoil Five game state returned from the API.
 *
 * Spoil Five (Maw) is an Irish play-only trick-taker for 5 players on a 52-card
 * deck (5 cards each). Trump is the turned-up card. Fixed top trumps — the trump
 * 5 (highest), trump J, and ♥A (always a trump) — may be held back rather than
 * following suit (Reneging). The first player to win 3 of the 5 tricks takes the
 * pot immediately; if nobody reaches 3 it is a Spoil (流局) and the pot carries
 * to the next round. First player to targetPoints wins the match.
 */
export interface SpoilFiveResponse extends BaseGameResponse {
  players: SpoilFivePlayer[];
  phase: SpoilFivePhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Trump suit (1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  /** Accumulated pot, awarded to the first player to win 3 tricks. */
  pot: number;
  /** Seat index of the round's winner, or -1 on a Spoil (流局). */
  roundWinnerIdx: number;
  currentTrick: SpoilFiveTrickCard[];
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: SpoilFiveHint | null;
  config: SpoilFiveConfig;
}

// --- Mariáš ---

/** Mariáš game phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd). */
export type MariasPhaseValue = 0 | 1 | 2 | 3;

/** A Mariáš player's public/own state. Cards are non-empty only for the human. */
export interface MariasPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match (game-point) score of this individual player. */
  score: number;
  /** Whether this player is the round's Soloist (plays alone vs the 2 Defenders). */
  isSoloist: boolean;
}

/** A card played into the current Mariáš trick. */
export interface MariasTrickCard {
  playerIdx: number;
  card: Card;
}

/** Mariáš game configuration. */
export interface MariasConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Mariáš, computed by the backend. */
export interface MariasHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Full Mariáš game state returned from the API. */
export interface MariasResponse extends BaseGameResponse {
  players: MariasPlayer[];
  phase: MariasPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the round's Soloist. */
  soloistIdx: number;
  /** Trump suit (1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  currentTrick: MariasTrickCard[];
  /** Cumulative match (game-point) scores per player — [p0, p1, p2]. */
  playerScores: number[];
  /** Card points captured per player this round — [p0, p1, p2]. */
  roundCardPoints: number[];
  /** Marriage (K+Q same suit) points scored per player this round — [p0, p1, p2]. */
  roundMarriage: number[];
  /** Seat index of the last (10th) trick winner, or -1. */
  lastTrickWinner: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: MariasHint | null;
  config: MariasConfig;
}

// --- King ---

/**
 * King game phase, mirrored from the Go domain string constants
 * (sync: internal/domain/King.go).
 *   - `selectContract` — the dealer chooses the deal's contract.
 *   - `play` — the 13-trick must-follow play phase.
 *   - `dealEnd` — one deal finished; scores settled, waiting for the next deal.
 *   - `gameEnd` — all seven contracts played; the match is over.
 */
export type KingPhaseValue = 'selectContract' | 'play' | 'dealEnd' | 'gameEnd';

/** A King player's public/own state. Cards are non-empty only for the human. */
export interface KingPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  /** Tricks captured this deal. */
  trickCount: number;
  /** Cumulative match score (King is trick-avoidance, so higher = fewer penalties). */
  totalScore: number;
}

/** A card played into the current King trick. */
export interface KingTrickCard {
  playerIdx: number;
  card: Card;
}

/** King game configuration. */
export interface KingConfig {
  cpuDifficulty: number;
}

/** Per-deal scoring detail surfaced at the end of each deal. */
export interface KingDealDetail {
  /** Contract index played this deal (0=No Tricks … 6=King Trump). */
  contract: number;
  /** Trump suit for the King (Trump) contract (1=♠ 2=♣ 3=♥ 4=♦), else -1. */
  trumpSuit: number;
  /** Seat index of the dealer who chose the contract. */
  dealerIdx: number;
  /** Points gained per player this deal, keyed by seat index. */
  gained: Record<number, number>;
}

/** A suggested hint for King, computed by the backend. */
export interface KingHint {
  cardIndices: number[];
  /** i18n reason suffix identifier (e.g. `avoid_low`, `win_high`). */
  reason: string;
}

/**
 * Full King game state returned from the API.
 *
 * King is a 4-player 52-card compendium trick-avoidance game. Each match is
 * exactly seven deals; the dealer of each deal chooses one of seven unused
 * contracts (0=No Tricks … 6=King Trump), then all four seats play thirteen
 * must-follow tricks. The highest total score (i.e. the fewest penalty points)
 * wins the match.
 */
export interface KingResponse extends BaseGameResponse {
  players: KingPlayer[];
  phase: KingPhaseValue;
  /** Current deal index (0-based) within the seven-deal match. */
  dealNumber: number;
  /** Total deals per match (always 7). */
  totalDeals: number;
  dealerIdx: number;
  /** Seat index whose turn it currently is. */
  currentTurn: number;
  /** Contract chosen this deal (0..6), or -1 before selection. */
  currentContract: number;
  /** Trump suit for the King (Trump) contract (1=♠ 2=♣ 3=♥ 4=♦), else -1. */
  trumpSuit: number;
  trickNumber: number;
  currentTrick: KingTrickCard[];
  lastTrick: KingTrickCard[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Which of the seven contracts have already been played this match. */
  usedContracts: boolean[];
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  config: KingConfig;
  /** Seat indices of the match winner(s); empty until the game ends. */
  roundWinners: number[];
  /** Scoring detail for the most recently completed deal, or null. */
  lastDealDetail?: KingDealDetail | null;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: KingHint | null;
}

// --- Tysiąc (Thousand) ---

/** Tysiąc game phase (0=Bid 1=Talon 2=Play 3=TrickEnd 4=RoundEnd 5=GameEnd). */
export type TysiacPhaseValue = 0 | 1 | 2 | 3 | 4 | 5;

/** A Tysiąc player's public/own state. Cards are non-empty only for the human. */
export interface TysiacPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of this individual player. */
  score: number;
  /** Whether this player is the round's Declarer (won the bid, plays the contract). */
  isDeclarer: boolean;
}

/** A card played into the current Tysiąc trick. */
export interface TysiacTrickCard {
  playerIdx: number;
  card: Card;
}

/** Tysiąc game configuration. */
export interface TysiacConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Tysiąc, computed by the backend. */
export interface TysiacHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Tysiąc (Thousand) game state returned from the API.
 *
 * Tysiąc is a Polish 3-player 24-card trick-taker with a Bid phase, a Talon
 * exchange phase, and marriage (K+Q) declarations during play. The Declarer
 * wins the bid and tries to meet the contract; trump is set dynamically by
 * declaring a marriage (so `trumpSuit` starts at 0 = unset).
 */
export interface TysiacResponse extends BaseGameResponse {
  players: TysiacPlayer[];
  phase: TysiacPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the forehand (first to bid / lead). */
  forehandIdx: number;
  /** Seat index of the round's Declarer (bid winner). */
  declarerIdx: number;
  /** The Declarer's contract (target card points for the round). */
  contract: number;
  /** The current highest bid in the Bid phase. */
  currentBid: number;
  /** Trump suit (0=unset, 1=♠ 2=♣ 3=♥ 4=♦). 0 until a marriage is declared. */
  trumpSuit: number;
  currentTrick: TysiacTrickCard[];
  /** Cumulative match scores per player — [p0, p1, p2]. */
  playerScores: number[];
  /** Card points captured per player this round — [p0, p1, p2]. */
  roundCardPoints: number[];
  /** Marriage (K+Q same suit) points scored per player this round — [p0, p1, p2]. */
  roundMarriage: number[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: TysiacHint | null;
  config: TysiacConfig;
}

// --- Calabresella (Terziglio) ---

/** Calabresella game phase (0=Bid 1=Discard 2=Play 3=TrickEnd 4=RoundEnd 5=GameEnd). */
export type CalabresellaPhaseValue = 0 | 1 | 2 | 3 | 4 | 5;

/** A Calabresella player's public/own state. Cards are non-empty only for the human. */
export interface CalabresellaPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of this individual player. */
  score: number;
  /** Whether this player is the round's Soloist (won the bid, plays alone). */
  isSoloist: boolean;
  /** Thirds of a point captured by this player in the current round. */
  roundThirds: number;
}

/** A card played into the current Calabresella trick. */
export interface CalabresellaTrickCard {
  playerIdx: number;
  card: Card;
}

/** Calabresella game configuration. */
export interface CalabresellaConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Calabresella, computed by the backend. */
export interface CalabresellaHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Calabresella (Terziglio) game state returned from the API.
 *
 * Calabresella is a Calabrian/Italian 3-player 40-card (Tressette-family)
 * trick-taker with a Bid phase, a monte exchange (discard four) phase, and no
 * trump. One Soloist (bid winner) plays alone against the coalition of the
 * other two and must capture more than half of the 33 thirds to win the round.
 */
export interface CalabresellaResponse extends BaseGameResponse {
  players: CalabresellaPlayer[];
  phase: CalabresellaPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  /** Seat index of the player whose turn it is to bid. */
  currentBidderIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the forehand (first to bid / lead). */
  forehandIdx: number;
  /** Seat index of the round's Soloist (bid winner), or -1 until decided. */
  soloistIdx: number;
  /** The winning bid (0=none, 1=chiamo, 2=solo). */
  winningBid: number;
  currentTrick: CalabresellaTrickCard[];
  /** Cumulative match scores per player — [p0, p1, p2]. */
  playerScores: number[];
  /** Thirds of a point captured per player this round — [p0, p1, p2]. */
  roundThirds: number[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: CalabresellaHint | null;
  config: CalabresellaConfig;
}

// --- Ombre (Hombre) ---

/** Ombre game phase (0=Bid 1=Play 2=TrickEnd 3=RoundEnd 4=GameEnd). */
export type OmbrePhaseValue = 0 | 1 | 2 | 3 | 4;

/** An Ombre player's public/own state. Cards are non-empty only for the human during play. */
export interface OmbrePlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of this individual player. */
  score: number;
  /** Whether this player is the round's Ombre (won the bid, plays alone). */
  isOmbre: boolean;
}

/** A card played into the current Ombre trick. */
export interface OmbreTrickCard {
  playerIdx: number;
  card: Card;
}

/** Ombre game configuration. */
export interface OmbreConfig {
  cpuDifficulty: number;
  /** Number of deals that make up the match; the highest cumulative score wins. */
  targetRounds: number;
}

/** A suggested hint for Ombre, computed by the backend. */
export interface OmbreHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Ombre (Hombre) game state returned from the API.
 *
 * Ombre is a 3-player soloist-vs-coalition trick-taker on a 40-card Spanish
 * deck (no 8/9/10). A Bid phase (pass / entrar / solo) plus a chosen trump suit
 * decides the Ombre, who then plays alone against the coalition of the other
 * two. The trump group ranks Spadille (♠A) > Manille (7 of trump) > Basto (♣A)
 * > Punto (Ace of trump, red only) > K > Q > J > 6..2 of trump.
 */
export interface OmbreResponse extends BaseGameResponse {
  players: OmbrePlayer[];
  phase: OmbrePhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  /** Seat index of the player whose turn it is to bid. */
  currentBidderIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the forehand (first to bid / lead). */
  forehandIdx: number;
  /** Seat index of the round's Ombre (bid winner), or -1 until decided. */
  ombreIdx: number;
  /** The winning bid (0=pass/none, 1=entrar, 2=solo). */
  winningBid: number;
  /** The trump suit (1=♠ 2=♣ 3=♥ 4=♦), or -1 until chosen. */
  trumpSuit: number;
  currentTrick: OmbreTrickCard[];
  /** Cumulative match scores per player — [p0, p1, p2]. */
  playerScores: number[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Deal outcome (0=None, 1=Sacar, 2=Puesta, 3=Codille). */
  outcome: number;
  /** Match result from the human's perspective (-1 lose, 0 none, 1 win). */
  result: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to play a card. */
  isHumanTurn: boolean;
  /** Whether it is currently the human's turn to bid. */
  isHumanBidTurn: boolean;
  hint?: OmbreHint | null;
  config: OmbreConfig;
}

// --- Ulti (Ultimo) ---

/** Ulti game phase (0=Bid 1=Discard 2=Play 3=TrickEnd 4=RoundEnd 5=GameEnd). */
export type UltiPhaseValue = 0 | 1 | 2 | 3 | 4 | 5;

/** An Ulti player's public/own state. Cards are non-empty only for the human declarer. */
export interface UltiPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Card-points captured in tricks so far this deal (A/10 = 10 each, +10 last trick). */
  cardPoints: number;
  /** Cumulative coin balance across the match. */
  coins: number;
  /** Whether this player is the declarer (always the human, seat 0). */
  isDeclarer: boolean;
}

/** A card played into the current Ulti trick. */
export interface UltiTrickCard {
  playerIdx: number;
  card: Card;
}

/** Ulti game configuration. */
export interface UltiConfig {
  cpuDifficulty: number;
  /** Number of deals that make up the match; the highest coin balance wins. */
  targetRounds: number;
}

/** A suggested hint for Ulti, computed by the backend. */
export interface UltiHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Ulti (Ultimo) game state returned from the API.
 *
 * Ulti is a 3-player Hungarian contract trick-taker on a 32-card deck
 * (A,10,K,Q,J,9,8,7; trick rank A>10>K>Q>J>9>8>7). The human (seat 0) is
 * always the declarer versus a 2-CPU defending coalition. After the Bid phase
 * (Party / Betli / Durchmarsch, with a trump suit for Party) the declarer takes
 * the 2-card talon and discards 2, then all three play out ten tricks.
 */
export interface UltiResponse extends BaseGameResponse {
  players: UltiPlayer[];
  phase: UltiPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the declarer (always the human, seat 0). */
  declarerIdx: number;
  /** The declared contract (0=None, 1=Party, 2=Betli, 3=Durchmarsch). */
  contract: number;
  /** The trump suit (1=♠ 2=♣ 3=♥ 4=♦), or -1 when none / not a Party contract. */
  trumpSuit: number;
  /** Number of face-down talon cards remaining. */
  talonCount: number;
  /** Whether the declarer has picked up the talon. */
  talonTaken: boolean;
  /** Number of cards discarded so far in the Discard phase. */
  discardCount: number;
  currentTrick: UltiTrickCard[];
  /** Cumulative coin balance per player — [p0, p1, p2]. */
  playerCoins: number[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Deal outcome (0=None, 1=Win/contract made, 2=Loss/contract failed). */
  outcome: number;
  /** Match result from the human's perspective (-1 lose, 0 none, 1 win). */
  result: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to act (play or discard). */
  isHumanTurn: boolean;
  /** Whether it is currently the human's turn to declare a contract. */
  isHumanBidTurn: boolean;
  hint?: UltiHint | null;
  config: UltiConfig;
}

// --- French Tarot ---

/** French Tarot game phase (0=Bid 1=Chien/écart 2=Play 3=TrickEnd 4=RoundEnd 5=GameEnd). */
export type FrenchTarotPhaseValue = 0 | 1 | 2 | 3 | 4 | 5;

/** A French Tarot player's public/own state. Cards are non-empty only for the human. */
export interface FrenchTarotPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Card-points captured in tricks so far this deal (French Tarot half-point card values). */
  cardPoints: number;
  /** Cumulative match score of this individual player. */
  score: number;
  /** Whether this player is the declarer (contract holder) this deal. */
  isDeclarer: boolean;
}

/** A card played into the current French Tarot trick. */
export interface FrenchTarotTrickCard {
  playerIdx: number;
  card: Card;
}

/** French Tarot game configuration. */
export interface FrenchTarotConfig {
  cpuDifficulty: number;
  /** Number of deals that make up the match; the highest cumulative score wins. */
  targetDeals: number;
}

/** A suggested hint for French Tarot, computed by the backend. */
export interface FrenchTarotHint {
  /** Suggested bid value during the Bid phase, or null/undefined outside it. */
  bid?: number | null;
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full French Tarot (フレンチタロット) game state returned from the API.
 *
 * French Tarot is a 4-player trick-taking game on the 78-card tarot deck (four
 * 14-card suits, 21 numbered trumps, and the Excuse). The human is seat 0. After
 * the auction (Pass / Petite / Garde / Garde Sans / Garde Contre) the highest
 * bidder becomes the declarer, may exchange the 6-card chien (écart) on a
 * Petite/Garde, then all four play out the tricks. The declarer must reach a
 * bouts-based card-point target to win the deal.
 */
export interface FrenchTarotResponse extends BaseGameResponse {
  players: FrenchTarotPlayer[];
  phase: FrenchTarotPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the player currently to bid (Bid phase). */
  bidPlayerIdx: number;
  /** The highest bid so far (0=none/pass, 1=Petite, 2=Garde, 3=Garde Sans, 4=Garde Contre). */
  highestBid: number;
  /** Seat index of the current highest bidder, or -1. */
  highestBidder: number;
  /** Seat index of the declarer, or -1 until decided. */
  declarerIdx: number;
  /** The declared contract (0=None, 1=Petite, 2=Garde, 3=Garde Sans, 4=Garde Contre). */
  contract: number;
  /** Number of cards in the chien (talon). */
  chienCount: number;
  /** The chien cards — non-empty only when revealed to a human declarer during écart. */
  chien: Card[];
  /** Whether the chien has been revealed. */
  chienRevealed: boolean;
  /** Seat index that receives the chien's stashed card points (declarer or -1). */
  stashOwner: number;
  currentTrick: FrenchTarotTrickCard[];
  /** Cumulative match score per player — [p0, p1, p2, p3]. */
  playerScores: number[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Deal outcome (0=None, 1=Win/contract made, 2=Loss/contract failed). */
  outcome: number;
  /** Match result from the human's perspective (-1 lose, 0 none, 1 win). */
  result: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to act (play). */
  isHumanTurn: boolean;
  /** Whether it is currently the human's turn to bid. */
  isHumanBidTurn: boolean;
  /** Whether it is currently the human's turn to discard the écart (6 cards). */
  isHumanDiscard: boolean;
  hint?: FrenchTarotHint | null;
  config: FrenchTarotConfig;
}

// --- Scarto (Piedmontese Tarot) ---

/** Scarto game phase (0=Scarto/discard 1=Play 2=TrickEnd 3=RoundEnd 4=GameEnd). */
export type ScartoPhaseValue = 0 | 1 | 2 | 3 | 4;

/** A Scarto player's public/own state. Cards are non-empty only for the human. */
export interface ScartoPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Card-points captured in tricks so far this deal (Italian tarocchi card values). */
  cardPoints: number;
  /** Cumulative match score of this individual player. */
  score: number;
  /** Whether this player is the dealer this deal (the dealer performs the scarto). */
  isDealer: boolean;
}

/** A card played into the current Scarto trick. */
export interface ScartoTrickCard {
  playerIdx: number;
  card: Card;
}

/** Scarto game configuration. */
export interface ScartoConfig {
  cpuDifficulty: number;
  /** Number of deals that make up the match; the highest cumulative score wins. */
  targetDeals: number;
}

/** A suggested hint for Scarto, computed by the backend. */
export interface ScartoHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Scarto (スカルト) game state returned from the API.
 *
 * Scarto is a simple 3-player Italian tarocchi trick-taker on the 78-card tarot
 * deck (four 14-card suits, 21 numbered trumps, and the Excuse). The human is
 * seat 0. There is no bidding, no chien, and no partnership: the dealer buries
 * three low pip cards (the scarto), then the three players play trump-priority
 * tricks. Each deal is scored as a zero-sum settlement against the average of
 * the three players' captured card-points; the highest cumulative score after
 * the set number of deals wins.
 */
export interface ScartoResponse extends BaseGameResponse {
  players: ScartoPlayer[];
  phase: ScartoPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  /** Seat index of the dealer, who performs the scarto (discard). */
  dealerIdx: number;
  /** Number of cards the dealer has already buried this deal (0 until the scarto is done, then 3). */
  scartoCount: number;
  currentTrick: ScartoTrickCard[];
  /** Cumulative match score per player — [p0, p1, p2]. */
  playerScores: number[];
  /** Signed settlement of the most recent deal per player — [p0, p1, p2]. */
  dealScores: number[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Deal outcome from the human's perspective (0=None/average, 1=above average, 2=below average). */
  outcome: number;
  /** Match result from the human's perspective (-1 lose, 0 none, 1 win). */
  result: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 for a draw / undecided. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to act (play). */
  isHumanTurn: boolean;
  /** Whether it is currently the human's turn to perform the scarto (they are the dealer). */
  isHumanScarto: boolean;
  hint?: ScartoHint | null;
  config: ScartoConfig;
}

// --- Königrufen (Tarock) ---

/** Königrufen game phase (0=Bid 1=Call-a-king 2=Talon/discard 3=Play 4=TrickEnd 5=RoundEnd 6=GameEnd). */
export type KoenigrufenPhaseValue = 0 | 1 | 2 | 3 | 4 | 5 | 6;

/** A Königrufen player's public/own state. Cards are non-empty only for the human. */
export interface KoenigrufenPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Card-points captured in tricks so far this deal. */
  cardPoints: number;
  /** Cumulative match score of this individual player. */
  score: number;
  /** Whether this player is the declarer (contract holder) this deal. */
  isDeclarer: boolean;
  /** Whether this player is the declarer's secret partner. Only ever true once partnerRevealed is true. */
  isPartner: boolean;
}

/** A card played into the current Königrufen trick. */
export interface KoenigrufenTrickCard {
  playerIdx: number;
  card: Card;
}

/** Königrufen game configuration. */
export interface KoenigrufenConfig {
  cpuDifficulty: number;
  /** Number of deals that make up the match; the highest cumulative score wins. */
  targetDeals: number;
}

/** A suggested hint for Königrufen, computed by the backend. */
export interface KoenigrufenHint {
  /** Suggested bid value during the Bid phase, or null/undefined outside it. */
  bid?: number | null;
  /** Suggested King suit to call during the Call phase (1-4), or null/undefined outside it. */
  callSuit?: number | null;
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Königrufen (ケーニッヒルーフェン) game state returned from the API.
 *
 * Königrufen is a 4-player tarock trick-taker on the 54-card tarock deck. After
 * the auction the declarer calls a King (King-calling / Rufer); whoever holds
 * that King becomes the declarer's secret partner. The declarer then exchanges
 * the talon (buries 6 cards) and the four play out the tricks. The partner's
 * identity stays hidden (`partnerIdx` is -1) until `partnerRevealed` is true.
 */
export interface KoenigrufenResponse extends BaseGameResponse {
  players: KoenigrufenPlayer[];
  phase: KoenigrufenPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the player currently to bid (Bid phase). */
  bidPlayerIdx: number;
  /** The highest bid so far (0=none/pass, 1=Rufer). */
  highestBid: number;
  /** Seat index of the current highest bidder, or -1. */
  highestBidder: number;
  /** Seat index of the declarer, or -1 until decided. */
  declarerIdx: number;
  /** The declared contract (0=None, 1=Rufer). */
  contract: number;
  /** The called King's suit (1=Spade 2=Clover 3=Heart 4=Diamond), or -1 until called. */
  calledKing: number;
  /** Seat index of the declarer's secret partner — always -1 until partnerRevealed is true. */
  partnerIdx: number;
  /** Whether the secret partner has been revealed (partner shown only when true). */
  partnerRevealed: boolean;
  /** Number of cards in the talon (buried stash). */
  talonCount: number;
  /** The talon cards — non-empty only when revealed to a human declarer during the discard. */
  talon: Card[];
  /** Seat index that receives the talon's stashed card points (declarer or -1). */
  stashOwner: number;
  currentTrick: KoenigrufenTrickCard[];
  /** Cumulative match score per player — [p0, p1, p2, p3]. */
  playerScores: number[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Deal outcome (0=None, 1=Win/contract made, 2=Loss/contract failed). */
  outcome: number;
  /** Match result from the human's perspective (-1 lose, 0 none, 1 win). */
  result: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends (also -1 on a draw). */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to act (play). */
  isHumanTurn: boolean;
  /** Whether it is currently the human's turn to bid. */
  isHumanBidTurn: boolean;
  /** Whether it is the human declarer's turn to call a King (Call phase). */
  isHumanCall: boolean;
  /** Whether it is the human declarer's turn to discard the talon (6 cards). */
  isHumanDiscard: boolean;
  hint?: KoenigrufenHint | null;
  config: KoenigrufenConfig;
}

// --- Cego (Baden Tarock) ---

/** Cego game phase (0=Bid 1=Contract 2=Exchange 3=Play 4=TrickEnd 5=RoundEnd 6=GameEnd). */
export type CegoPhaseValue = 0 | 1 | 2 | 3 | 4 | 5 | 6;

/** A Cego player's public/own state. Cards are non-empty only for the human. */
export interface CegoPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Card-points captured in tricks so far this deal. */
  cardPoints: number;
  /** Cumulative match score of this individual player. */
  score: number;
  /** Whether this player is the declarer (contract holder) this deal. */
  isDeclarer: boolean;
}

/** A card played into the current Cego trick. */
export interface CegoTrickCard {
  playerIdx: number;
  card: Card;
}

/** Cego game configuration. */
export interface CegoConfig {
  cpuDifficulty: number;
  /** Number of deals that make up the match; the highest cumulative score wins. */
  targetDeals: number;
}

/** A suggested hint for Cego, computed by the backend. */
export interface CegoHint {
  /** Suggested bid value during the Bid phase, or null/undefined outside it. */
  bid?: number | null;
  /** Suggested contract during the Contract phase (1=Cego 2=Handspiel), or null/undefined outside it. */
  contract?: number | null;
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Cego (チェゴ) game state returned from the API.
 *
 * Cego is a 4-player Baden tarock trick-taker on the 54-card tarock deck. One
 * declarer plays 1-vs-3. After the auction the declarer chooses a contract —
 * Cego (lay down all but one dealt card and pick up the 10-card blind) or
 * Handspiel (keep the dealt hand) — then the four play out 11 tricks. The
 * blind's contents are never revealed (only `blindCount`).
 */
export interface CegoResponse extends BaseGameResponse {
  players: CegoPlayer[];
  phase: CegoPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the player currently to bid (Bid phase). */
  bidPlayerIdx: number;
  /** The highest bid so far (0=none/pass, 1=play). */
  highestBid: number;
  /** Seat index of the current highest bidder, or -1. */
  highestBidder: number;
  /** Seat index of the declarer, or -1 until decided. */
  declarerIdx: number;
  /** The winning bid (0=None, 1=Play). */
  contract: number;
  /** The chosen contract type (0=None, 1=Cego, 2=Handspiel). */
  contractType: number;
  /** Number of cards in the blind (Cego stash) — the contents stay hidden. */
  blindCount: number;
  /** The blind cards — always empty; the blind is hidden (use blindCount). */
  blind: Card[];
  /** Seat index that receives the blind's stashed card points (declarer side or opponents). */
  stashOwner: number;
  currentTrick: CegoTrickCard[];
  /** Cumulative match score per player — [p0, p1, p2, p3]. */
  playerScores: number[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Deal outcome (0=None, 1=Win/contract made, 2=Loss/contract failed). */
  outcome: number;
  /** Match result from the human's perspective (-1 lose, 0 none, 1 win). */
  result: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends (also -1 on a draw). */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to act (play). */
  isHumanTurn: boolean;
  /** Whether it is currently the human's turn to bid. */
  isHumanBidTurn: boolean;
  /** Whether it is the human declarer's turn to choose a contract (Contract phase). */
  isHumanContract: boolean;
  /** Whether it is the human declarer's turn to make the Cego exchange (keep exactly 1 card). */
  isHumanExchange: boolean;
  hint?: CegoHint | null;
  config: CegoConfig;
}

// --- Cinch ---

/** Cinch game phase (0=Bid 1=NameTrump 2=Play 3=TrickEnd 4=RoundEnd 5=GameEnd). */
export type CinchPhaseValue = 0 | 1 | 2 | 3 | 4 | 5;

/** A Cinch player's public/own state. Cards are non-empty only for the human. */
export interface CinchPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** This player's bid this deal (0=pass, 1-14; -1 if not yet bid). */
  bid: number;
  /** Cumulative match score of this individual player. */
  totalScore: number;
}

/** A card played into the current Cinch trick. */
export interface CinchTrickCard {
  playerIdx: number;
  card: Card;
}

/** Per-deal scoring breakdown for Cinch (surfaced at round/game end). */
export interface CinchDealDetail {
  /** Trump suit for the scored deal (1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  /** Seat index of the bid winner (bidder) for the deal. */
  bidderIdx: number;
  /** The winning bid amount. */
  bid: number;
  /** Whether the bidding side was set back (failed to make its bid). */
  setBack: boolean;
  /** Points captured per player this deal, keyed by seat index. */
  points: Record<number, number>;
  /** Match points gained (or lost) per player this deal, keyed by seat index. */
  gained: Record<number, number>;
}

/** Cinch game configuration. */
export interface CinchConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A suggested hint for Cinch, computed by the backend. */
export interface CinchHint {
  cardIndices: number[];
  /** Suggested bid amount (present for a bid-phase hint). */
  bid?: number | null;
  /** Suggested trump suit (present for a name-trump-phase hint). */
  trumpSuit?: number | null;
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Cinch (Double Pedro / High Five) game state returned from the API.
 *
 * Cinch is a 4-player (1 human + 3 CPU, individual scoring) All-Fours/Pitch-family
 * bidding trick-taker on a standard 52-card deck. Nine cards are dealt to each
 * player; players bid 0 (pass) or 1-14, and the high bidder names trump and leads.
 * There are 14 points per deal — High/King/Ten/Jack of trump = 1 each, the Right
 * Pedro (5 of trump) = 5, and the Left Pedro (5 of the same color as trump, which
 * ranks just below the trump 5) = 5. The bidding side must capture at least its
 * bid or it is set back; the first player to reach the target score (default 21)
 * wins.
 */
export interface CinchResponse extends BaseGameResponse {
  players: CinchPlayer[];
  phase: CinchPhaseValue;
  roundNumber: number;
  trickNumber: number;
  totalTricks: number;
  dealerIdx: number;
  /** Seat index of the player whose turn it is to act. */
  currentTurn: number;
  /** Seat index of the player whose turn it is to bid. */
  bidPlayerIdx: number;
  /** The current highest bid this deal (0 if none yet). */
  currentBid: number;
  /** Seat index of the bid winner, or -1 until decided. */
  bidWinnerIdx: number;
  /** Trump suit (0=unset, 1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  currentTrick: CinchTrickCard[];
  lastTrick: CinchTrickCard[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerIdx: number;
  /** Seat indices of players who have reached / won at game end. */
  roundWinners: number[];
  lastDealDetail?: CinchDealDetail | null;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: CinchHint | null;
  config: CinchConfig;
}

// --- Loo (Lanterloo) ---

/** Loo game phase (0=Decide 1=Play 2=TrickEnd 3=RoundEnd). */
export type LooPhaseValue = 0 | 1 | 2 | 3;

/** A Loo player's public/own state. Cards are non-empty only for the human. */
export interface LooPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Whether this player is participating in the current deal (play vs pass). */
  playing: boolean;
  /** Cumulative chip balance of this individual player (can be negative). */
  chips: number;
}

/** A card played into the current Loo trick. */
export interface LooTrickCard {
  playerIdx: number;
  card: Card;
}

/** Per-deal settlement breakdown for Loo (surfaced at round end). */
export interface LooDealDetail {
  /** Pot size at the start of the deal (used to size the per-trick payout). */
  potStart: number;
  /** Trump suit for the scored deal (1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  /** Whether each seat participated (played) this deal, keyed by seat index. */
  playing: boolean[];
  /** Tricks captured per participating seat this deal, keyed by seat index. */
  tricks: Record<number, number>;
  /** Chips gained (or lost) per player this deal, keyed by seat index. */
  gained: Record<number, number>;
  /** Seat indices of players who were "looed" (played but took no tricks). */
  looed: number[];
  /** Chips carried over into the next deal's pot. */
  potCarry: number;
}

/** Loo game configuration. */
export interface LooConfig {
  cpuDifficulty: number;
  ante: number;
}

/** A suggested hint for Loo, computed by the backend. */
export interface LooHint {
  cardIndices: number[];
  /** Suggested participation decision (present for a decide-phase hint). */
  decision?: boolean | null;
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Loo (Lanterloo) game state returned from the API.
 *
 * Loo is a 4-player (1 human + 3 CPU, individual chips) pot-based gambling
 * trick-taker on a standard 52-card deck. Each player is dealt five cards; the
 * turn-up card sets trump. Players decide to play or pass, then participants
 * fight five must-follow / must-head tricks, each trick winning one-fifth of the
 * pot. A participant who wins no trick is "looed" and pays a penalty into the
 * next deal's pot. There is no game-over target — it is a repeating deal loop, so
 * `gameEndFlag` is always false.
 */
export interface LooResponse extends BaseGameResponse {
  players: LooPlayer[];
  phase: LooPhaseValue;
  roundNumber: number;
  trickNumber: number;
  totalTricks: number;
  dealerIdx: number;
  /** Seat index of the player whose turn it is to act. */
  currentTurn: number;
  /** Seat index of the player whose turn it is to decide (play/pass). */
  decidePlayerIdx: number;
  /** Trump suit (0=unset, 1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  /** The turn-up card whose suit becomes trump. */
  turnUp?: Card | null;
  /** Current pot size (chips available for distribution this deal). */
  pot: number;
  /** Pot size at the start of the deal (used to size the per-trick payout). */
  potStart: number;
  currentTrick: LooTrickCard[];
  lastTrick: LooTrickCard[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  /** Always false — Loo has no game-over; it is a repeating deal loop. */
  gameEndFlag: boolean;
  lastDealDetail?: LooDealDetail | null;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: LooHint | null;
  config: LooConfig;
}

// --- Basra (Bastra) ---

/** Basra game phase (0=Play 1=GameEnd). */
export type BasraPhaseValue = 0 | 1;

/** A Basra player's public/own state. Cards are non-empty only for the human. */
export interface BasraPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  /** Number of cards captured so far this game. */
  capturedCount: number;
  /** Number of Basra sweeps (clearing the table with a single non-Jack card). */
  basraCount: number;
  /** Final score (populated at game end). */
  score: number;
}

/** Per-game scoring breakdown for Basra (surfaced at game end). */
export interface BasraScoreDetail {
  /** Captured card counts per seat, keyed by seat index. */
  cards: Record<number, number>;
  /** Ace counts per seat, keyed by seat index. */
  aces: Record<number, number>;
  /** Basra sweep counts per seat, keyed by seat index. */
  basras: Record<number, number>;
  /** Seat index holding the 7♦, or -1. */
  hasSevenDiamonds: number;
  /** Seat index holding the 10♦, or -1. */
  hasTenDiamonds: number;
  /** Seat index with the (unique) most captured cards, or -1 on a tie. */
  mostCards: number;
  /** Points gained per seat this game, keyed by seat index. */
  gained: Record<number, number>;
}

/** Basra game configuration. */
export interface BasraConfig {
  cpuDifficulty: number;
}

/** A suggested hint for Basra, computed by the backend. */
export interface BasraHint {
  cardIndices: number[];
  /** Suggested table-card indices to capture (present for a capture hint). */
  tableIndices?: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Basra (Bastra) game state returned from the API.
 *
 * Basra is a 4-player (1 human + 3 CPU, individual scoring) fishing/capture game
 * on a standard 52-card deck. Each player is dealt four cards with four face-up on
 * the table. A number card captures same-rank cards and any table subset summing
 * to its value; a Jack sweeps the whole table (except other Jacks). Clearing the
 * table with a single non-Jack card scores a "Basra" bonus. When the stock is
 * exhausted the game ends (`gameEndFlag` true) and scores are tallied.
 */
export interface BasraResponse extends BaseGameResponse {
  players: BasraPlayer[];
  phase: BasraPhaseValue;
  /** Number of packs dealt so far (deal counter). */
  roundNumber: number;
  /** Seat index of the player whose turn it is to act. */
  currentTurn: number;
  /** Cards currently face-up on the table. */
  tableCards: Card[];
  /** Seat index of the last player who captured, or -1. */
  lastCaptureIdx: number;
  /** Number of cards left in the stock. */
  remainingDeck: number;
  /** Indices in the human's hand that are legal to play (non-empty on human turn). */
  playableIndices: number[];
  /** Map of hand index -> table indices that hand card can capture (human turn). */
  captureOptions: Record<number, number[]>;
  /** Seat indices of the winner(s) at game end. */
  winners: number[];
  /** Whether the game has ended (stock exhausted and scored). */
  gameEndFlag: boolean;
  lastDealDetail?: BasraScoreDetail | null;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: BasraHint | null;
  config: BasraConfig;
}

// --- Koi-Koi (こいこい) ---

/** Koi-Koi phase value (0=Play 1=KoiKoiDecision 2=RoundEnd 3=GameEnd). */
export type KoiKoiPhaseValue = 0 | 1 | 2 | 3;

/** A single completed yaku (combination) with its point value. */
export interface KoiKoiYaku {
  /** Yaku identifier key (e.g. "goko", "tane", "kasu"); localized on the frontend. */
  key: string;
  /** Point value awarded for this yaku. */
  points: number;
}

/** A Koi-Koi player's public/own state. Hand `cards` are non-empty only for the human. */
export interface KoiKoiPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  /** Hand cards (populated only for the human). */
  cards: Card[];
  /** Cards captured into this player's pile so far this round. */
  captured: Card[];
  capturedCount: number;
  /** Cumulative match score. */
  score: number;
  /** Whether this player has called koi-koi (continue) this round. */
  calledKoiKoi: boolean;
  /** Yaku the player currently holds (from captured cards). */
  yaku: KoiKoiYaku[];
  /** Total points of the player's current yaku. */
  yakuPoints: number;
}

/** Result detail for one completed Koi-Koi round. */
export interface KoiKoiRoundResult {
  /** Winning seat index (-1 on a draw / exhausted round). */
  winner: number;
  /** Yaku the winner scored. */
  yaku: KoiKoiYaku[];
  /** Sum of yaku points before the koi-koi multiplier. */
  basePoints: number;
  /** Multiplier applied (2 when koi-koi was called). */
  multiplier: number;
  /** Final points awarded (basePoints × multiplier). */
  total: number;
  /** Number of koi-koi calls that occurred in the round. */
  koikoiCount: number;
}

/** Koi-Koi game configuration. */
export interface KoiKoiConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** A suggested hint for Koi-Koi, computed by the backend. */
export interface KoiKoiHint {
  cardIndex: number;
  fieldIndex: number;
  /** 1 = suggest calling koi-koi, 0 = suggest stopping (shobu). */
  koikoi: number;
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Koi-Koi (こいこい) game state returned from the API.
 *
 * Koi-Koi is a 2-player hanafuda capture game. On the human's turn the player
 * plays a hand card to capture a matching field card (same month), then draws
 * from the stock which may also capture. When a new yaku (combination) is
 * completed the player decides between koi-koi (continue for more, doubling the
 * stakes) and shobu (stop and score). Cards carry a hanafuda face descriptor
 * (`glyph`/`label`/`color`/`deck`) and render procedurally via `CardImage`.
 */
export interface KoiKoiResponse extends BaseGameResponse {
  players: KoiKoiPlayer[];
  phase: KoiKoiPhaseValue;
  /** Round (deal) counter. */
  roundNumber: number;
  /** Seat index of the player whose turn it is to act. */
  currentTurn: number;
  /** Cards currently face-up on the field. */
  fieldCards: Card[];
  /** Number of cards left in the stock. */
  remainingDeck: number;
  /** Number of koi-koi calls so far this round (drives the multiplier). */
  koikoiCount: number;
  /** Indices in the human's hand that are legal to play (human Play turn). */
  playableIndices: number[];
  /** Map of hand index -> field indices that hand card can capture (present when 2-way choice). */
  captureOptions: Record<number, number[]>;
  /** Yaku the acting player just completed, pending a koi-koi/shobu decision. */
  pendingYaku: KoiKoiYaku[];
  /** Total points of the pending yaku. */
  pendingPoints: number;
  /** Winning seat index of the just-finished round, or -1. */
  roundWinner: number;
  /** Winning seat index of the whole match, or -1. */
  winner: number;
  /** Whether the match has ended. */
  gameEndFlag: boolean;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  /** Result of the most recently completed round (RoundEnd phase), or null. */
  lastRoundResult?: KoiKoiRoundResult | null;
  hint?: KoiKoiHint | null;
  config: KoiKoiConfig;
}

// --- Hachi-Hachi (八八 / はちはち) ---

/** Hachi-Hachi phase value (0=Play 1=RoundEnd 2=GameEnd). */
export type HachiHachiPhaseValue = 0 | 1 | 2;

/** A single completed yaku (combination) with its bonus point value. */
export interface HachiHachiYaku {
  /** Yaku identifier key (e.g. "goko", "inoshikacho"); localized on the frontend. */
  key: string;
  /** Bonus point value awarded for this yaku. */
  points: number;
}

/** A Hachi-Hachi player's public/own state. Hand `cards` are non-empty only for the human. */
export interface HachiHachiPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  /** Hand cards (populated only for the human). */
  cards: Card[];
  /** Cards captured into this player's pile so far this round. */
  captured: Card[];
  capturedCount: number;
  /** Cumulative signed match score (each round settled against the 88 baseline). */
  score: number;
  /** This player's signed delta from the most recently completed round. */
  roundDelta: number;
  /** Raw card-point total of the captured pile (Bright 20 / Animal 10 / Ribbon 5 / Chaff 1). */
  rawScore: number;
  /** Yaku the player currently holds (from captured cards). */
  yaku: HachiHachiYaku[];
}

/** One player's round-settlement breakdown, present in {@link HachiHachiRoundResult}. */
export interface HachiHachiPlayerScore {
  playerIdx: number;
  /** Raw card-point total for the round. */
  rawScore: number;
  /** Yaku bonuses the player scored this round. */
  yaku: HachiHachiYaku[];
  /** Sum of yaku bonus points. */
  bonus: number;
  /** Signed delta from the 88 baseline (rawScore + bonus − 88). */
  delta: number;
}

/** Result detail for one completed Hachi-Hachi round (RoundEnd phase). */
export interface HachiHachiRoundResult {
  /** Per-player settlement breakdowns for the round. */
  scores: HachiHachiPlayerScore[];
  /** Seat index with the highest delta this round. */
  best: number;
}

/** Hachi-Hachi game configuration. */
export interface HachiHachiConfig {
  cpuDifficulty: number;
  /** Number of rounds (deals) played before the match is settled. */
  targetRounds: number;
}

/** A suggested hint for Hachi-Hachi, computed by the backend. */
export interface HachiHachiHint {
  cardIndex: number;
  fieldIndex: number;
  /** i18n reason suffix identifier (e.g. "capture", "discard_low"). */
  reason: string;
}

/**
 * Full Hachi-Hachi (八八) game state returned from the API.
 *
 * Hachi-Hachi is the classic 3-player Japanese hanafuda game on the 48-card
 * flower deck. Players capture field cards of the same month with their hand
 * and stock draws; when every hand is exhausted the round's captured piles are
 * scored by card points (Bright 20 / Animal 10 / Ribbon 5 / Chaff 1) plus yaku
 * bonuses, and each player settles against the 88 baseline. Unlike Koi-Koi
 * there is no koi-koi/stop decision — phases are simply Play, RoundEnd, and
 * GameEnd. Cards carry a hanafuda face descriptor (`glyph`/`label`/`color`/
 * `deck`) and render procedurally via `CardImage`.
 */
export interface HachiHachiResponse extends BaseGameResponse {
  players: HachiHachiPlayer[];
  phase: HachiHachiPhaseValue;
  /** Round (deal) counter. */
  roundNumber: number;
  /** Seat index of the player whose turn it is to act. */
  currentTurn: number;
  /** Cards currently face-up on the field. */
  fieldCards: Card[];
  /** Number of cards left in the stock. */
  remainingDeck: number;
  /** Indices in the human's hand that are legal to play (human Play turn). */
  playableIndices: number[];
  /** Map of hand index -> field indices that hand card can capture (present when 2-way choice). */
  captureOptions: Record<number, number[]>;
  /** Winning seat index of the whole match (highest cumulative score), or -1. */
  winner: number;
  /** Match result enum value from the backend. */
  result: number;
  /** Whether the match has ended. */
  gameEndFlag: boolean;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  /** Settlement of the most recently completed round (RoundEnd phase), or null. */
  lastRoundResult?: HachiHachiRoundResult | null;
  hint?: HachiHachiHint | null;
  config: HachiHachiConfig;
}

// --- Go-Stop (Godori / ゴーストップ) ---

/** Go-Stop phase value (0=Play 1=GoDecision 2=RoundEnd 3=GameEnd). */
export type GoStopPhaseValue = 0 | 1 | 2 | 3;

/**
 * Korean Go-Stop scoring breakdown for a player's captured pile. Points are
 * split into the five categories (gwang/godori/tti/yeol/pi), then the `base`
 * total is multiplied by the go multiplier to produce `goScore`.
 */
export interface GoStopBreakdown {
  /** Bright (光 / 광) points. */
  gwang: number;
  /** Five-bird (五鳥 / 고도리) points. */
  godori: number;
  /** Ribbon (띠) points. */
  tti: number;
  /** Animal (열끗) points. */
  yeol: number;
  /** Junk (피) points. */
  pi: number;
  /** Sum of all category points before the go multiplier. */
  base: number;
  /** Number of "go" calls made this round. */
  goCount: number;
  /** Multiplier applied for the go calls. */
  goMult: number;
  /** Points after applying the go multiplier. */
  goScore: number;
  /** Number of bright cards captured. */
  brightCount: number;
  /** Number of ribbon cards captured. */
  ribbonCount: number;
  /** Number of animal cards captured. */
  animalCount: number;
  /** Number of junk cards captured. */
  piCount: number;
}

/** A Go-Stop player's public/own state. Hand `cards` are non-empty only for the human. */
export interface GoStopPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  /** Hand cards (populated only for the human). */
  cards: Card[];
  /** Cards captured into this player's pile so far this round. */
  captured: Card[];
  capturedCount: number;
  /** Cumulative match score. */
  score: number;
  /** Number of "go" calls this player has made this round. */
  goCount: number;
  /** Current scoring breakdown from captured cards, or null. */
  breakdown: GoStopBreakdown | null;
  /** Current total points (base × go multiplier). */
  points: number;
}

/** Result detail for one completed Go-Stop round. */
export interface GoStopRoundResult {
  /** Winning seat index (-1 on a draw / exhausted round). */
  winner: number;
  /** The winner's scoring breakdown. */
  breakdown: GoStopBreakdown | null;
  /** Base points before go/bak multipliers. */
  basePoints: number;
  /** Points after the go multiplier. */
  goScore: number;
  /** Combined bak (penalty-doubling) multiplier applied to the loser's payment. */
  bakMult: number;
  /** Final points transferred to the winner. */
  total: number;
  /** Whether gwang-bak (bright penalty) applied. */
  gwangBak: boolean;
  /** Whether pi-bak (junk penalty) applied. */
  piBak: boolean;
  /** Whether go-bak (go-call penalty) applied. */
  goBak: boolean;
  /** Number of go calls in the round. */
  goCount: number;
}

/** Go-Stop game configuration. */
export interface GoStopConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** A suggested hint for Go-Stop, computed by the backend. */
export interface GoStopHint {
  cardIndex: number;
  fieldIndex: number;
  /** 1 = suggest calling go, 0 = suggest stopping (-1 during Play). */
  go: number;
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Go-Stop (Godori / ゴーストップ) game state returned from the API.
 *
 * Go-Stop is the Korean sibling of Koi-Koi, played with the same 48-card
 * hanafuda deck. On the human's turn the player plays a hand card to capture a
 * matching field card (same month), then draws from the stock which may also
 * capture. When the target score is reached the GoDecision phase offers go
 * (continue for more) or stop (bank the points). Cards carry a hanafuda face
 * descriptor (`glyph`/`label`/`color`/`deck`) and render procedurally via
 * `CardImage`.
 */
export interface GoStopResponse extends BaseGameResponse {
  players: GoStopPlayer[];
  phase: GoStopPhaseValue;
  /** Round (deal) counter. */
  roundNumber: number;
  /** Seat index of the player whose turn it is to act. */
  currentTurn: number;
  /** Cards currently face-up on the field. */
  fieldCards: Card[];
  /** Number of cards left in the stock. */
  remainingDeck: number;
  /** Indices in the human's hand that are legal to play (human Play turn). */
  playableIndices: number[];
  /** Map of hand index -> field indices that hand card can capture (present when 2-way choice). */
  captureOptions: Record<number, number[]>;
  /** Breakdown of the score that triggered the go/stop decision, or null. */
  pendingBreakdown: GoStopBreakdown | null;
  /** Total points pending a go/stop decision. */
  pendingPoints: number;
  /** Winning seat index of the just-finished round, or -1. */
  roundWinner: number;
  /** Winning seat index of the whole match, or -1. */
  winner: number;
  /** Raw result enum value. */
  result: number;
  /** Whether the match has ended. */
  gameEndFlag: boolean;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  /** Result of the most recently completed round (RoundEnd phase), or null. */
  lastRoundResult?: GoStopRoundResult | null;
  hint?: GoStopHint | null;
  config: GoStopConfig;
}

// --- Tablanet (Tablić) ---

/** Tablanet game phase (0=Play 1=GameEnd). */
export type TablanetPhaseValue = 0 | 1;

/** A Tablanet player's public/own state. Cards are non-empty only for the human. */
export interface TablanetPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  /** Number of cards captured so far this game. */
  capturedCount: number;
  /** Number of Tabla sweeps (clearing the table with a single non-Jack card). */
  tablaCount: number;
  /** Final score (populated at game end). */
  score: number;
}

/** Per-game scoring breakdown for Tablanet (surfaced at game end). */
export interface TablanetScoreDetail {
  /** Captured card counts per seat, keyed by seat index. */
  cards: Record<number, number>;
  /** Ace counts per seat, keyed by seat index. */
  aces: Record<number, number>;
  /** Jack counts per seat, keyed by seat index. */
  jacks: Record<number, number>;
  /** Tabla sweep counts per seat, keyed by seat index. */
  tablas: Record<number, number>;
  /** Seat index holding the 10♦, or -1. */
  hasTenDiamonds: number;
  /** Seat index holding the 2♣, or -1. */
  hasTwoClubs: number;
  /** Seat index with the (unique) most captured cards, or -1 on a tie. */
  mostCards: number;
  /** Points gained per seat this game, keyed by seat index. */
  gained: Record<number, number>;
}

/** Tablanet game configuration. */
export interface TablanetConfig {
  cpuDifficulty: number;
}

/** A suggested hint for Tablanet, computed by the backend. */
export interface TablanetHint {
  cardIndices: number[];
  /** Suggested table-card indices to capture (present for a capture hint). */
  tableIndices?: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Tablanet (Tablić) game state returned from the API.
 *
 * Tablanet is a 4-player (1 human + 3 CPU, individual scoring) fishing/capture
 * game on a standard 52-card deck. Each player is dealt four cards with four
 * face-up on the table. A number card captures same-rank cards and any table
 * subset summing to its value; a Jack sweeps the whole table (except other Jacks).
 * Clearing the table with a single non-Jack card scores a "Tabla" bonus. When the
 * stock is exhausted the game ends (`gameEndFlag` true) and scores are tallied.
 */
export interface TablanetResponse extends BaseGameResponse {
  players: TablanetPlayer[];
  phase: TablanetPhaseValue;
  /** Number of packs dealt so far (deal counter). */
  roundNumber: number;
  /** Seat index of the player whose turn it is to act. */
  currentTurn: number;
  /** Cards currently face-up on the table. */
  tableCards: Card[];
  /** Seat index of the last player who captured, or -1. */
  lastCaptureIdx: number;
  /** Number of cards left in the stock. */
  remainingDeck: number;
  /** Indices in the human's hand that are legal to play (non-empty on human turn). */
  playableIndices: number[];
  /** Map of hand index -> table indices that hand card can capture (human turn). */
  captureOptions: Record<number, number[]>;
  /** Seat indices of the winner(s) at game end. */
  winners: number[];
  /** Whether the game has ended (stock exhausted and scored). */
  gameEndFlag: boolean;
  lastDealDetail?: TablanetScoreDetail | null;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: TablanetHint | null;
  config: TablanetConfig;
}

// --- Solo Whist ---

/** Solo Whist game phase (0=Bid 1=Play 2=TrickEnd 3=RoundEnd 4=GameEnd). */
export type SoloWhistPhaseValue = 0 | 1 | 2 | 3 | 4;

/** A Solo Whist player's public/own state. Cards are non-empty only for the human. */
export interface SoloWhistPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of this individual player. */
  score: number;
  /** Whether this player is the round's declarer (plays alone vs the 3 defenders). */
  isDeclarer: boolean;
}

/** A card played into the current Solo Whist trick. */
export interface SoloWhistTrickCard {
  playerIdx: number;
  card: Card;
}

/** Solo Whist game configuration. */
export interface SoloWhistConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Solo Whist, computed by the backend. */
export interface SoloWhistHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

// --- Auction Forty-Fives ---

/** Auction Forty-Fives game phase (0=Bid 1=Play 2=TrickEnd 3=RoundEnd 4=GameEnd). */
export type FortyFivesPhaseValue = 0 | 1 | 2 | 3 | 4;

/** An Auction Forty-Fives player's public/own state. Cards are non-empty only for the human. */
export interface FortyFivesPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of this player's TEAM (seats 0&2 = team 0, 1&3 = team 1). */
  teamScore: number;
  /** Whether this player is the round's declarer (the highest bidder). */
  isDeclarer: boolean;
}

/** A card played into the current Auction Forty-Fives trick. */
export interface FortyFivesTrickCard {
  playerIdx: number;
  card: Card;
}

/** Auction Forty-Fives game configuration. */
export interface FortyFivesConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Auction Forty-Fives, computed by the backend. */
export interface FortyFivesHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Server response for the Auction Forty-Fives game (4 players, 2 teams). */
export interface FortyFivesResponse extends BaseGameResponse {
  players: FortyFivesPlayer[];
  phase: FortyFivesPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the round's declarer (highest bidder), or -1 before bidding resolves. */
  declarerIdx: number;
  /** Winning contract value (0=Pass 15 20 25). */
  contract: number;
  /** Trump suit (0=none during bid, else 1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  /** Each player's bid this round — [p0, p1, p2, p3]. */
  bids: number[];
  currentTrick: FortyFivesTrickCard[];
  /** Cumulative match scores per team — [teamA, teamB]. */
  teamScores: number[];
  /** Points scored by each team this round — [teamA, teamB]. */
  roundTeamPoints: number[];
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning team index (0 or 1), or -1 until the game ends. */
  winnerTeam: number;
  /** Whether it is currently the human's turn to play a card. */
  isHumanTurn: boolean;
  /** Whether it is currently the human's turn to bid. */
  isHumanBidTurn: boolean;
  hint?: FortyFivesHint | null;
  config: FortyFivesConfig;
}

// --- Twenty-Nine (29) ---

/** Twenty-Nine (29) game phase (0=Bid 1=Play 2=TrickEnd 3=RoundEnd 4=GameEnd). */
export type TwentyNinePhaseValue = 0 | 1 | 2 | 3 | 4;

/** A Twenty-Nine (29) player's public/own state. Cards are non-empty only for the human. */
export interface TwentyNinePlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative game-point score of this player's TEAM (seats 0&2 = team 0, 1&3 = team 1). */
  teamScore: number;
  /** Whether this player is the round's declarer (the winning bidder). */
  isDeclarer: boolean;
}

/** A card played into the current Twenty-Nine (29) trick. */
export interface TwentyNineTrickCard {
  playerIdx: number;
  card: Card;
}

/** Twenty-Nine (29) game configuration. */
export interface TwentyNineConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Twenty-Nine (29), computed by the backend. */
export interface TwentyNineHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Server response for the Twenty-Nine (29) game (4 players, 2 teams, hidden trump). */
export interface TwentyNineResponse extends BaseGameResponse {
  players: TwentyNinePlayer[];
  phase: TwentyNinePhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the round's declarer (winning bidder), or -1 before bidding resolves. */
  declarerIdx: number;
  /** Winning contract value (0=Pass 16 20 24 28). */
  contract: number;
  /** Trump suit (0=none/hidden during bid, else 1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  /** Whether the hidden trump suit has been revealed yet. */
  trumpRevealed: boolean;
  /** Each player's bid this round — [p0, p1, p2, p3]. */
  bids: number[];
  currentTrick: TwentyNineTrickCard[];
  /** Cumulative game-point scores per team — [teamA, teamB]. */
  teamScores: number[];
  /** Card points captured by each team this round — [teamA, teamB]. */
  roundTeamPoints: number[];
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning team index (0 or 1), or -1 until the game ends. */
  winnerTeam: number;
  /** Whether it is currently the human's turn to play a card. */
  isHumanTurn: boolean;
  /** Whether it is currently the human's turn to bid. */
  isHumanBidTurn: boolean;
  hint?: TwentyNineHint | null;
  config: TwentyNineConfig;
}

// --- Court Piece / Rang ---

/** Court Piece (Rang) game phase (0=TrumpDeclaration 1=Play 2=TrickEnd 3=RoundEnd 4=GameEnd). */
export type CourtPiecePhaseValue = 0 | 1 | 2 | 3 | 4;

/** A Court Piece (Rang) player's public/own state. Cards are non-empty only for the human. */
export interface CourtPiecePlayer {
  id: number;
  isHuman: boolean;
  /** Team index (seats 0&2 = team 0, 1&3 = team 1). */
  team: number;
  cardCount: number;
  cards: Card[];
  /** Round points (tricks won this round). */
  roundScore: number;
  /** Cumulative game-point (Sar) score of this player's team. */
  cumulativeScore: number;
  trickCount: number;
}

/** A card played into the current Court Piece (Rang) trick. */
export interface CourtPieceTrickCard {
  playerIdx: number;
  card: Card;
}

/** Court Piece (Rang) game configuration. */
export interface CourtPieceConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A suggested hint for Court Piece (Rang), computed by the backend. */
export interface CourtPieceHint {
  /** Card index to play (Play phase). */
  cardIndex?: number;
  /** Trump suit to declare (TrumpDeclaration phase, 1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit?: number;
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Server response for the Court Piece (Rang) game (4 players, 2 teams, called trump). */
export interface CourtPieceResponse extends BaseGameResponse {
  players: CourtPiecePlayer[];
  phase: CourtPiecePhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  /** Seat index of the caller (Hakim) who declares the trump suit. */
  callerIdx: number;
  /** Trump suit (0=undeclared during TrumpDeclaration, else 1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  currentTrick: CourtPieceTrickCard[];
  /** Cumulative game-point (Sar) scores per team — [teamA, teamB]. */
  teamScores: number[];
  /** Consecutive round wins by the {@link lastWinnerTeam} (drives the Court bonus). */
  consecutiveWins: number;
  /** Team that won the previous round (or -1 before any round resolves). */
  lastWinnerTeam: number;
  /** Whether the previous round was a Court (sweep / consecutive bonus). */
  lastRoundCourt: boolean;
  gameEndFlag: boolean;
  /** Winning team index (0 or 1), or -1 until the game ends. */
  winnerTeam: number;
  /** Seat index of the player who led the current trick. */
  leadPlayerIdx: number;
  hint?: CourtPieceHint | null;
  config: CourtPieceConfig;
}

// --- Bezique ---

/** Bezique game phase (0=Play 1=Meld 2=RoundEnd 3=GameEnd). */
export type BeziquePhaseValue = 0 | 1 | 2 | 3;

/** A Bezique player's public/own state. Cards are non-empty only for the human. */
export interface BeziquePlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  /** Points scored from melds and brisques in the current deal. */
  roundScore: number;
  /** Cumulative match score accumulated across deals. */
  cumulativeScore: number;
  trickCount: number;
}

/** A card played into the current Bezique trick. */
export interface BeziqueTrickCard {
  playerIdx: number;
  card: Card;
}

/**
 * A meld the trick winner may declare during the Meld phase. `type` is the
 * meld kind (0=marriage 1=Bezique 2=four aces 3=four kings 4=four queens
 * 5=four jacks); `suit` is the marriage suit (1=♠ 2=♣ 3=♥ 4=♦, or -1 for
 * Bezique and four-of-a-kind); `points` is the score it would award.
 */
export interface BeziqueMeld {
  type: number;
  suit: number;
  points: number;
}

/** Bezique game configuration. */
export interface BeziqueConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** A suggested hint for Bezique, computed by the backend (may carry a card index OR a meld index, where -1 = skip). */
export interface BeziqueHint {
  /** Card index to play (Play phase). */
  cardIndex?: number;
  /** Meld index to declare (Meld phase); -1 means skip the meld. */
  meldIndex?: number;
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Server response for the Bezique game (2 players, melds, two-phase trick play). */
export interface BeziqueResponse extends BaseGameResponse {
  players: BeziquePlayer[];
  /** Points scored in the current deal, indexed by seat. */
  dealPoints: number[];
  /** Of the deal points, the portion from melds (trick portion = dealPoints - dealMeldPoints). */
  dealMeldPoints: number[];
  /** Cumulative match score, indexed by seat. */
  matchScore: number[];
  phase: BeziquePhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Trump suit (0=♠ 1=♣ 2=♥ 3=♦ — the deck's suit ordinal). */
  trumpSuit: number;
  /** The face-up card that fixed the trump (present until the stock empties). */
  trumpCard?: Card;
  currentTrick: BeziqueTrickCard[];
  /** Cards remaining in the stock (phase 2 begins when this reaches 0). */
  stockRemaining: number;
  /** Whether the deal has entered the strict must-follow endgame (phase 2). */
  isEndgame: boolean;
  /** Melds the human may declare this Meld phase (empty otherwise). */
  availableMelds: BeziqueMeld[];
  gameEndFlag: boolean;
  /** Winning seat index (0 or 1), or -1 until the game ends. */
  winnerIdx: number;
  hint?: BeziqueHint | null;
  config: BeziqueConfig;
}

// --- Écarté ---

/** Écarté game phase (0=Exchange 1=Play 2=RoundEnd 3=GameEnd). */
export type EcartePhaseValue = 0 | 1 | 2 | 3;

/**
 * Écarté negotiation sub-step within the Exchange phase
 * (0=ElderDecide 1=DealerRespond 2=ElderDiscard 3=DealerDiscard).
 */
export type EcarteNegStepValue = 0 | 1 | 2 | 3;

/** An Écarté player's public/own state. Cards are non-empty only for the human. */
export interface EcartePlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  /** Points scored in the current deal. */
  roundScore: number;
  /** Cumulative match score accumulated across deals. */
  cumulativeScore: number;
  trickCount: number;
}

/** A card played into the current Écarté trick. */
export interface EcarteTrickCard {
  playerIdx: number;
  card: Card;
}

/** Écarté game configuration. */
export interface EcarteConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/**
 * A suggested hint for Écarté, computed by the backend. During the Play phase
 * it carries a `cardIndex`; during the Exchange phase it carries an `action`
 * string (e.g. `propose`, `stand`, `accept`, `refuse`, `discard`).
 */
export interface EcarteHint {
  /** Card index to play (Play phase). */
  cardIndex?: number;
  /** Exchange-phase action identifier (Exchange phase). */
  action?: string;
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Écarté game state returned from the API.
 *
 * Écarté is a 2-player French 32-card trick game. Before play, an Exchange
 * phase lets the elder (non-dealer) Propose or Stand; if proposed, the dealer
 * Accepts or Refuses; on accept, each player discards any number of cards and
 * draws replacements, then the elder decides again (repeating until the stock
 * empties). Play is 5 strict must-follow tricks (rank K>Q>J>A>10>9>8>7).
 * Winning 3+ tricks scores 1 point, all 5 (Vole) scores 2; holding the King of
 * trump scores +1, a turned King gives the dealer +1, and a dealer who refuses
 * then loses gives the elder +1. Scores accumulate to a target (default 5).
 */
export interface EcarteResponse extends BaseGameResponse {
  players: EcartePlayer[];
  /** Points scored in the current deal, indexed by seat. */
  dealPoints: number[];
  /** Cumulative match score, indexed by seat. */
  matchScore: number[];
  phase: EcartePhaseValue;
  /** Exchange-phase negotiation sub-step (only meaningful in phase 0). */
  negStep: EcarteNegStepValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the elder (non-dealer) player. */
  elderIdx: number;
  leadPlayerIdx: number;
  /** Trump suit (1=♠ 2=♣ 3=♥ 4=♦; 0=undeclared). */
  trumpSuit: number;
  /** The face-up card that fixed the trump (present until the stock empties). */
  trumpCard?: Card;
  currentTrick: EcarteTrickCard[];
  /** Cards remaining in the stock. */
  stockRemaining: number;
  /** Whether the dealer refused the most recent exchange proposal. */
  refusalByDealer: boolean;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  validPlays: number[];
  gameEndFlag: boolean;
  /** Winning seat index (0 or 1), or -1 until the game ends. */
  winnerIdx: number;
  hint?: EcarteHint | null;
  config: EcarteConfig;
}

// --- Three Card Brag ---

/** Three Card Brag game phase (0=Betting 1=Showdown 2=RoundEnd 3=GameEnd). */
export type ThreeCardBragPhaseValue = 0 | 1 | 2 | 3;

/**
 * A Three Card Brag player's public/own state. `cards` is populated for the
 * human (once seen) and for everyone at showdown; `handName` is set only when
 * a hand is revealed.
 */
export interface ThreeCardBragPlayer {
  id: number;
  isHuman: boolean;
  /** Remaining chips. */
  chips: number;
  /** Whether the player has looked at their hand (Seen) vs still Blind. */
  seen: boolean;
  /** Whether the player has folded out of the current deal. */
  folded: boolean;
  /** Whether the player has been eliminated (busted) from the match. */
  out: boolean;
  /** Chips this player has wagered into the pot this deal. */
  roundBet: number;
  cardCount: number;
  cards: Card[];
  /** The hand ranking name, set once the hand is revealed. */
  handName?: string;
}

/** Three Card Brag game configuration. */
export interface ThreeCardBragConfig {
  cpuDifficulty: number;
  /** Chips put in the pot by each player at the start of a deal. */
  ante: number;
  /** Chips each player begins the match with. */
  startingChips: number;
}

/**
 * A suggested hint for Three Card Brag, computed by the backend. `action` is
 * the suggested betting action (e.g. `see`, `bet`, `raise`, `fold`, `show`).
 */
export interface ThreeCardBragHint {
  /** Suggested betting action identifier. */
  action: string;
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Three Card Brag game state returned from the API.
 *
 * Three Card Brag is a 4-player British vying game (an ancestor of poker)
 * played with a 52-card deck, 3 cards each, and chips wagered into a pot. Each
 * player is Blind or Seen; on their turn they can See (reveal, Blind→Seen), Bet
 * (call the stake — Blind pays the stake, Seen pays double), Raise, or Fold.
 * When two players remain a Seen player may Show to force a showdown. The last
 * player standing in a deal wins the pot, chip-busted players are eliminated,
 * and the last player with chips wins the match. Hand ranking is
 * Prial > Running Flush > Run > Flush > Pair > High Card.
 */
export interface ThreeCardBragResponse extends BaseGameResponse {
  players: ThreeCardBragPlayer[];
  /** Chips currently in the pot. */
  pot: number;
  /** Current stake a Blind player must match to bet. */
  stake: number;
  phase: ThreeCardBragPhaseValue;
  roundNumber: number;
  dealerIdx: number;
  currentPlayerIdx: number;
  /** Winning seat index of the current deal, or -1 until it ends. */
  roundWinnerIdx: number;
  /** Winning seat index of the match, or -1 until the game ends. */
  matchWinnerIdx: number;
  /** Whether the deal has reached a showdown (hands revealed). */
  isShowdown: boolean;
  /** Whether a Seen player may Show (force a showdown) right now. */
  canShow: boolean;
  gameEndFlag: boolean;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: ThreeCardBragHint | null;
  config: ThreeCardBragConfig;
}

// --- Teen Patti ---

/** Teen Patti game phase (0=Betting 1=SideShow 2=Showdown 3=RoundEnd 4=GameEnd). */
export type TeenPattiPhaseValue = 0 | 1 | 2 | 3 | 4;

/**
 * A Teen Patti player's public/own state. `cards` is populated for the human
 * (once seen) and for everyone at showdown; `handName` is set only when a hand
 * is revealed.
 */
export interface TeenPattiPlayer {
  id: number;
  isHuman: boolean;
  /** Remaining chips. */
  chips: number;
  /** Whether the player has looked at their hand (Seen) vs still Blind. */
  seen: boolean;
  /** Whether the player has folded out of the current deal. */
  folded: boolean;
  /** Whether the player has been eliminated (busted) from the match. */
  out: boolean;
  /** Chips this player has wagered into the pot this deal. */
  roundBet: number;
  cardCount: number;
  cards: Card[];
  /** The hand ranking name, set once the hand is revealed. */
  handName?: string;
}

/** Teen Patti game configuration. */
export interface TeenPattiConfig {
  cpuDifficulty: number;
  /** Chips put in the pot by each player at the start of a deal. */
  ante: number;
  /** Chips each player begins the match with. */
  startingChips: number;
}

/**
 * A suggested hint for Teen Patti, computed by the backend. `action` is the
 * suggested betting action (e.g. `see`, `bet`, `raise`, `fold`, `show`,
 * `sideshow`).
 */
export interface TeenPattiHint {
  /** Suggested betting action identifier. */
  action: string;
  /** i18n reason suffix identifier. */
  reason: string;
}

/** One participant's revealed hand in a resolved Side Show. */
export interface TeenPattiSideShowHand {
  /** Seat index of this participant. */
  playerIdx: number;
  /** Hand ranking name key (see `hand.*` i18n). */
  handName: string;
  /** The participant's three cards, revealed for the comparison. */
  cards: Card[];
}

/**
 * The comparison result of the most recent accepted Side Show, present only
 * when the human was a participant (CPU-vs-CPU comparisons stay hidden).
 */
export interface TeenPattiSideShowResult {
  /** Seat index that requested the Side Show. */
  requesterIdx: number;
  /** Seat index that accepted the Side Show. */
  targetIdx: number;
  /** Seat index that won the comparison. */
  winnerIdx: number;
  /** Seat index that lost and folded. */
  loserIdx: number;
  /** The requester's revealed hand. */
  requester: TeenPattiSideShowHand;
  /** The target's revealed hand. */
  target: TeenPattiSideShowHand;
}

/**
 * Full Teen Patti game state returned from the API.
 *
 * Teen Patti is the Indian variant of Three Card Brag — a 4-player vying game
 * played with a 52-card deck, 3 cards each, and chips wagered into a pot. Each
 * player is Blind or Seen; on their turn they can See (reveal, Blind→Seen), Bet
 * (call the stake — Blind pays the stake, Seen pays double), Raise, or Fold.
 * When two players remain a Seen player may Show to force a showdown. Teen
 * Patti additionally lets a Seen player request a **Side Show** with the
 * previous Seen player (a private hand comparison; the loser folds), which the
 * target then accepts or declines. The last player standing in a deal wins the
 * pot, chip-busted players are eliminated, and the last player with chips wins
 * the match. Hand ranking is Trail (trio) > Pure Sequence (straight flush) >
 * Sequence (straight) > Color (flush) > Pair > High Card.
 */
export interface TeenPattiResponse extends BaseGameResponse {
  players: TeenPattiPlayer[];
  /** Chips currently in the pot. */
  pot: number;
  /** Current stake a Blind player must match to bet. */
  stake: number;
  phase: TeenPattiPhaseValue;
  roundNumber: number;
  dealerIdx: number;
  currentPlayerIdx: number;
  /** Winning seat index of the current deal, or -1 until it ends. */
  roundWinnerIdx: number;
  /** Winning seat index of the match, or -1 until the game ends. */
  matchWinnerIdx: number;
  /** Whether the deal has reached a showdown (hands revealed). */
  isShowdown: boolean;
  /** Whether a Seen player may Show (force a showdown) right now. */
  canShow: boolean;
  /** Whether the current player may request a Side Show right now. */
  canRequestSideShow: boolean;
  /** Seat index that requested a Side Show, or -1 when none pending. */
  sideShowRequester: number;
  /** Seat index asked to accept/decline a Side Show, or -1 when none pending. */
  sideShowTarget: number;
  gameEndFlag: boolean;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: TeenPattiHint | null;
  /** Result of the last human-involved Side Show comparison, if any. */
  lastSideShow?: TeenPattiSideShowResult | null;
  config: TeenPattiConfig;
}

// --- Préférence ---

/** Préférence game phase (0=Bid 1=Play 2=TrickEnd 3=RoundEnd 4=GameEnd). */
export type PreferencePhaseValue = 0 | 1 | 2 | 3 | 4;

/** A Préférence player's public/own state. Cards are non-empty only for the human. */
export interface PreferencePlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of this individual player. */
  score: number;
  /** Whether this player is the round's declarer (plays alone vs the 2 defenders). */
  isDeclarer: boolean;
}

/** A card played into the current Préférence trick. */
export interface PreferenceTrickCard {
  playerIdx: number;
  card: Card;
}

/** Préférence game configuration. */
export interface PreferenceConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Préférence, computed by the backend. */
export interface PreferenceHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Full Préférence game state returned from the API. */
export interface PreferenceResponse extends BaseGameResponse {
  players: PreferencePlayer[];
  phase: PreferencePhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the round's declarer, or -1 before bidding resolves. */
  declarerIdx: number;
  /** Winning contract (0=Pass 1=Six 2=Misère 3=Seven 4=Eight). */
  contract: number;
  /** Trump suit (0=none during bid / Misère, else 1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  /** Each player's bid this round (0-4) — [p0, p1, p2]. */
  bids: number[];
  currentTrick: PreferenceTrickCard[];
  /** Cumulative match scores per player — [p0, p1, p2]. */
  playerScores: number[];
  /** Tricks captured per player this round — [p0, p1, p2]. */
  roundTricks: number[];
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to play a card. */
  isHumanTurn: boolean;
  /** Whether it is currently the human's turn to bid. */
  isHumanBidTurn: boolean;
  hint?: PreferenceHint | null;
  config: PreferenceConfig;
}

// --- Nap (Napoleon) ---

/** Nap game phase (0=Bid 1=Play 2=TrickEnd 3=RoundEnd 4=GameEnd). */
export type NapPhaseValue = 0 | 1 | 2 | 3 | 4;

/** A Nap player's public/own state. Cards are non-empty only for the human. */
export interface NapPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative chip score of this individual player. */
  score: number;
  /** Whether this player is the round's declarer (plays to make the bid). */
  isDeclarer: boolean;
}

/** A card played into the current Nap trick. */
export interface NapTrickCard {
  playerIdx: number;
  card: Card;
}

/** Nap game configuration. */
export interface NapConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Nap, computed by the backend. */
export interface NapHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Full Nap game state returned from the API. */
export interface NapResponse extends BaseGameResponse {
  players: NapPlayer[];
  phase: NapPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the round's declarer, or -1 before bidding resolves. */
  declarerIdx: number;
  /** Winning contract (0=Pass 2=Two 3=Three 4=Four 5=Nap; the value is the bid trick count). */
  contract: number;
  /** Trump suit (0 during bid, else 1=♠ 2=♣ 3=♥ 4=♦ in play). */
  trumpSuit: number;
  /** Each player's bid this round (0/2/3/4/5) — [p0, p1, p2, p3]. */
  bids: number[];
  currentTrick: NapTrickCard[];
  /** Cumulative chip scores per player — [p0, p1, p2, p3]. */
  playerScores: number[];
  /** Tricks captured per player this round — [p0, p1, p2, p3]. */
  roundTricks: number[];
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to play a card. */
  isHumanTurn: boolean;
  /** Whether it is currently the human's turn to bid. */
  isHumanBidTurn: boolean;
  hint?: NapHint | null;
  config: NapConfig;
}

/** Full Solo Whist game state returned from the API. */
export interface SoloWhistResponse extends BaseGameResponse {
  players: SoloWhistPlayer[];
  phase: SoloWhistPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the round's declarer, or -1 before bidding resolves. */
  declarerIdx: number;
  /** Winning contract (0=Pass 1=Solo 2=Misère 3=Abundance). */
  contract: number;
  /** Trump suit (0=none for Misère, else 1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  /** Each player's bid this round (0-3) — [p0, p1, p2, p3]. */
  bids: number[];
  currentTrick: SoloWhistTrickCard[];
  /** Cumulative match scores per player — [p0, p1, p2, p3]. */
  playerScores: number[];
  /** Tricks captured per player this round — [p0, p1, p2, p3]. */
  roundTricks: number[];
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to play a card. */
  isHumanTurn: boolean;
  /** Whether it is currently the human's turn to bid. */
  isHumanBidTurn: boolean;
  hint?: SoloWhistHint | null;
  config: SoloWhistConfig;
}

// --- Spades ---

/** Spades player data with bid, scores, and bags. */
export interface SpadesPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  bid: number;
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
  bags: number;
}

/** A card played in a Spades trick. */
export interface SpadesTrickCard {
  playerIdx: number;
  card: Card;
}

/** Spades game configuration. */
export interface SpadesConfig {
  cpuDifficulty: number;
  pointLimit: number;
  nilBonus: number;
  bagPenaltyThreshold: number;
}

/** A suggested hint for Spades. */
export interface SpadesHint {
  cardIndex?: number;
  bid?: number;
  reason: string;
}

/** Full Spades game state returned from the API. */
export interface SpadesResponse extends BaseGameResponse {
  players: SpadesPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  currentTrick: SpadesTrickCard[];
  spadesBroken: boolean;
  gameEndFlag: boolean;
  winnerIdx: number;
  leadPlayerIdx: number;
  config: SpadesConfig;
  hint?: SpadesHint;
}

// --- Call Break ---

/** Call Break player data with bid and integer×10 scores. */
export interface CallBreakPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  bid: number;
  /** Round score in internal int×10 form (e.g. 41 == 4.1 points). */
  roundScore: number;
  /** Cumulative score in internal int×10 form. */
  cumulativeScore: number;
  trickCount: number;
}

/** A card played in a Call Break trick. */
export interface CallBreakTrickCard {
  playerIdx: number;
  card: Card;
}

/** Call Break game configuration. */
export interface CallBreakConfig {
  cpuDifficulty: number;
  maxRounds: number;
}

/** A suggested hint for Call Break. */
export interface CallBreakHint {
  cardIndex?: number;
  bid?: number;
  reason: string;
}

/** Full Call Break game state returned from the API. */
export interface CallBreakResponse extends BaseGameResponse {
  players: CallBreakPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  currentTrick: CallBreakTrickCard[];
  spadesBroken: boolean;
  gameEndFlag: boolean;
  winnerIdx: number;
  leadPlayerIdx: number;
  config: CallBreakConfig;
  hint?: CallBreakHint;
  /**
   * Indices in the human player's hand that are legal to play this turn.
   * Empty array outside the play phase / when it is not the human's turn.
   */
  validPlayIndices: number[];
}

// --- Pitch (Setback / Auction Pitch) ---

/** Pitch player data (4-player single-handed scoring). */
export interface PitchPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  bid: number;
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
}

/** A card played in a Pitch trick. */
export interface PitchTrickCard {
  playerIdx: number;
  card: Card;
}

/** Pitch game configuration. */
export interface PitchConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A suggested hint for Pitch. */
export interface PitchHint {
  cardIndex?: number;
  bid?: number;
  reason: string;
}

/** Full Pitch game state returned from the API. */
export interface PitchResponse extends BaseGameResponse {
  players: PitchPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  dealerIdx: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  currentBid: number;
  bidWinnerIdx: number;
  trumpSuit: number;
  currentTrick: PitchTrickCard[];
  /** Cards of the just-completed trick (empty on the round's first trick). */
  lastTrick: PitchTrickCard[];
  /** Winner index of the just-completed trick, or -1 when none. */
  lastTrickWinner: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  leadPlayerIdx: number;
  validPlayIndices: number[];
  config: PitchConfig;
  hint?: PitchHint;
}

// --- Two Ten Jack (ツーテンジャック) ---

/** Two Ten Jack player data (4-player team game: seats 0,2 vs 1,3). */
export interface TwoTenJackPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
  capturedPoints: number;
}

/** A card played in a Two Ten Jack trick. */
export interface TwoTenJackTrickCard {
  playerIdx: number;
  card: Card;
}

/** Two Ten Jack game configuration. */
export interface TwoTenJackConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A suggested hint for Two Ten Jack. */
export interface TwoTenJackHint {
  cardIndex?: number;
  trumpSuit?: number;
  reason: string;
}

/** Full Two Ten Jack game state returned from the API. */
export interface TwoTenJackResponse extends BaseGameResponse {
  players: TwoTenJackPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  declarerIdx: number;
  trumpSuit: number;
  currentTrick: TwoTenJackTrickCard[];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  config: TwoTenJackConfig;
  hint?: TwoTenJackHint;
}

// --- Crazy Eights (クレイジーエイト) ---

/** Crazy Eights player data with scores. */
export interface CrazyEightsPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
}

/** Crazy Eights game configuration. */
export interface CrazyEightsConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** Full Crazy Eights game state returned from the API. */
export interface CrazyEightsResponse extends BaseGameResponse {
  players: CrazyEightsPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  chosenSuit: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  config: CrazyEightsConfig;
}

// --- Prší (プルシー) ---

/** Prší player data. */
export interface PrsiPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
}

/** Prší game configuration (no point limit — first to empty hand wins). */
export interface PrsiConfig {
  cpuDifficulty: number;
}

/** Full Prší game state returned from the API. */
export interface PrsiResponse extends BaseGameResponse {
  players: PrsiPlayerData[];
  phase: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  penaltyDrawCount: number;
  pendingSkips: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  config: PrsiConfig;
}

// --- Macau (マカオ) ---

/** Macau player data with scores and declaration state. */
export interface MacauPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  hasDeclared: boolean;
}

/** Macau game configuration. */
export interface MacauConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** Full Macau game state returned from the API. */
export interface MacauResponse extends BaseGameResponse {
  players: MacauPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  chosenSuit: number;
  penaltyDrawCount: number;
  direction: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  config: MacauConfig;
}

// --- Mao (マオ) ---

/** Mao player data with scores and declaration state. */
export interface MaoPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  hasDeclared: boolean;
}

/** Mao game configuration. */
export interface MaoConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/**
 * Full Mao game state returned from the API.
 *
 * The hidden rule is never sent to the client. Only indirect signals are
 * exposed: {@link MaoResponse.awaitingWord} (a word may be required),
 * {@link MaoResponse.rulePenalty} (the last action broke the hidden rule),
 * {@link MaoResponse.correctCount} (successful compliances so far), and
 * {@link MaoResponse.hintUnlocked}/{@link MaoResponse.ruleHint} (a vague hint,
 * populated only after 3 correct compliances).
 */
export interface MaoResponse extends BaseGameResponse {
  players: MaoPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  chosenSuit: number;
  penaltyDrawCount: number;
  direction: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  awaitingWord: boolean;
  correctCount: number;
  hintUnlocked: boolean;
  ruleHint: string;
  rulePenalty: boolean;
  config: MaoConfig;
}

// --- Page One (ページワン) ---

/** Page One player data with scores. */
export interface PageOnePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  hasDeclared: boolean;
}

/** Page One game configuration. */
export interface PageOneConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** Full Page One game state returned from the API. */
export interface PageOneResponse extends BaseGameResponse {
  players: PageOnePlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  config: PageOneConfig;
}

// --- Gin Rummy (ジンラミー) ---

/** Gin Rummy player data with scores. */
export interface GinRummyPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
}

/** A meld (set or run) in Gin Rummy. */
export interface GinRummyMeld {
  cards: Card[];
}

/** Gin Rummy game configuration. */
export interface GinRummyConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

// --- Indian Rummy (インドラミー) ---

/** Indian Rummy player data with scores, deadwood, and pure-sequence flag. */
export interface IndianRummyPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  deadwood: number;
  hasPureSequence: boolean;
}

/** Indian Rummy game configuration. */
export interface IndianRummyConfig {
  playerCount: number;
  cpuDifficulty: number;
  targetRounds: number;
}

/** Full Indian Rummy game state returned from the API. */
export interface IndianRummyResponse extends BaseGameResponse {
  players: IndianRummyPlayer[];
  phase: number;
  roundNumber: number;
  targetRounds: number;
  currentPlayerIdx: number;
  dealerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  wildJoker: Card | null;
  wildRank: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  declarerIdx: number;
  declarationValid: boolean;
  config: IndianRummyConfig;
}

// --- Machiavelli (マキャヴェッリ) ---

/** Machiavelli player data with scores and deadwood. */
export interface MachiavelliPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  deadwood: number;
}

/** A meld on the shared Machiavelli table (kind: 0=set, 1=run). */
export interface MachiavelliMeld {
  cards: Card[];
  kind: number;
}

/** Machiavelli game configuration. */
export interface MachiavelliConfig {
  playerCount: number;
  cpuDifficulty: number;
  targetRounds: number;
}

/** Full Machiavelli game state returned from the API. */
export interface MachiavelliResponse extends BaseGameResponse {
  players: MachiavelliPlayer[];
  table: MachiavelliMeld[];
  phase: number;
  roundNumber: number;
  targetRounds: number;
  currentPlayerIdx: number;
  dealerIdx: number;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  roundWinnerIdx: number;
  config: MachiavelliConfig;
}

// --- Panguingue / Pan (パングインゲ) ---

/** A meld (set or rope/run) laid on the table by a Panguingue player. */
export interface PanMeld {
  cards: Card[];
}

/** Panguingue player data with laid melds, chips, and scores. */
export interface PanPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  laidMelds: PanMeld[];
  meldedCount: number;
  chips: number;
  handPoints: number;
  roundScore: number;
  cumulativeScore: number;
}

/** Panguingue game configuration. */
export interface PanConfig {
  playerCount: number;
  cpuDifficulty: number;
  targetRounds: number;
}

/** Full Panguingue game state returned from the API. */
export interface PanResponse extends BaseGameResponse {
  players: PanPlayer[];
  phase: number;
  roundNumber: number;
  targetRounds: number;
  currentPlayerIdx: number;
  dealerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  deckSize: number;
  winMeldCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  panDeclarerIdx: number;
  config: PanConfig;
}

// --- Tonk (トンク) ---

/** Tonk player data with scores. */
export interface TonkPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
}

/** A meld (set or run) in Tonk. */
export interface TonkMeld {
  cards: Card[];
}

/** Tonk game configuration. */
export interface TonkConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** Full Tonk game state returned from the API. */
export interface TonkResponse extends BaseGameResponse {
  players: TonkPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  knockerIdx: number;
  knockerMelds: TonkMeld[];
  knockerDeadwood: Card[];
  opponentMelds: TonkMeld[];
  opponentDeadwood: Card[];
  isTonk: boolean;
  isUndercut: boolean;
  config: TonkConfig;
}

// --- Thirty-One (サーティワン / Scat) ---

/** Thirty-One player data with lives and best-suit score. */
export interface ThirtyOnePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  lives: number;
  score: number;
  isEliminated: boolean;
}

/** Thirty-One game configuration. */
export interface ThirtyOneConfig {
  cpuDifficulty: number;
  initialLives: number;
}

/** Full Thirty-One game state returned from the API. */
export interface ThirtyOneResponse extends BaseGameResponse {
  players: ThirtyOnePlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  knockerIdx: number;
  thirtyOneIdx: number;
  roundWinnerIdx: number;
  roundLosers: number[];
  config: ThirtyOneConfig;
}

// --- Yaniv (ヤニブ) ---

/** Yaniv player data with cumulative penalty score and revealed hand total. */
export interface YanivPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  score: number;
  handTotal: number;
  isEliminated: boolean;
}

/** Yaniv game configuration. */
export interface YanivConfig {
  cpuDifficulty: number;
  scoreLimit: number;
}

/** Full Yaniv game state returned from the API. */
export interface YanivResponse extends BaseGameResponse {
  players: YanivPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  pickupCards: Card[];
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  callerIdx: number;
  asafWinnerIdx: number;
  isAsaf: boolean;
  roundScores: number[];
  config: YanivConfig;
}

// --- Seven Bridge (セブンブリッジ) ---

/** A meld (set or run) shared across players in Seven Bridge. */
export interface SevenBridgeMeld {
  cards: Card[];
}

/** Seven Bridge player data with hand, melds and scores. */
export interface SevenBridgePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  melds: SevenBridgeMeld[];
  roundScore: number;
  cumulativeScore: number;
}

/** Seven Bridge game configuration. */
export interface SevenBridgeConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** Full Seven Bridge game state returned from the API. */
export interface SevenBridgeResponse extends BaseGameResponse {
  players: SevenBridgePlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  roundWinnerIdx: number;
  config: SevenBridgeConfig;
}

// --- Rummy 500 ---

/** Rummy 500 player data with hand, laid melds and scores. */
export interface Rummy500PlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  laidMelds: Rummy500Meld[];
}

/** A meld (set or run) laid by a player in Rummy 500. */
export interface Rummy500Meld {
  cards: Card[];
}

/** Rummy 500 game configuration. */
export interface Rummy500Config {
  cpuDifficulty: number;
  pointLimit: number;
}

/** Full Rummy 500 game state returned from the API. */
export interface Rummy500Response extends BaseGameResponse {
  players: Rummy500PlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardPile: Card[];
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  roundEnderIdx: number;
  config: Rummy500Config;
}

/** Full Gin Rummy game state returned from the API. */
export interface GinRummyResponse extends BaseGameResponse {
  players: GinRummyPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  knockerIdx: number;
  knockerMelds: GinRummyMeld[];
  knockerDeadwood: Card[];
  isGin: boolean;
  config: GinRummyConfig;
}

// --- Conquian (コンキャン) ---

/** A table meld (set or run) in Conquian. */
export interface ConquianMeld {
  cards: Card[];
}

/** Conquian player data with face-up table melds and rounds won. */
export interface ConquianPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  melds: ConquianMeld[];
  wins: number;
}

/** Conquian game configuration. */
export interface ConquianConfig {
  cpuDifficulty: number;
  targetWins: number;
}

/** Full Conquian game state returned from the API. */
export interface ConquianResponse extends BaseGameResponse {
  players: ConquianPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  roundWinnerIdx: number;
  tookDiscard: boolean;
  config: ConquianConfig;
}

// --- Chinchón (チンチョン) ---

/** Chinchón player data with round and cumulative scores and elimination flag. */
export interface ChinchonPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  eliminated: boolean;
}

/** A meld (set or run) laid down by the knocker in Chinchón. */
export interface ChinchonMeld {
  cards: Card[];
}

/** Chinchón game configuration. */
export interface ChinchonConfig {
  cpuDifficulty: number;
  playerCount: number;
  knockThreshold: number;
  eliminationLimit: number;
}

/** Full Chinchón game state returned from the API. */
export interface ChinchonResponse extends BaseGameResponse {
  players: ChinchonPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  knockerIdx: number;
  knockerMelds: ChinchonMeld[];
  config: ChinchonConfig;
}

/** Three Thirteen player data with deadwood, round, and cumulative scores. */
export interface ThreeThirteenPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  deadwood: number;
  roundScore: number;
  cumulativeScore: number;
}

/** Three Thirteen game configuration. */
export interface ThreeThirteenConfig {
  cpuDifficulty: number;
  playerCount: number;
}

/** Full Three Thirteen game state returned from the API. */
export interface ThreeThirteenResponse extends BaseGameResponse {
  players: ThreeThirteenPlayerData[];
  phase: number;
  round: number;
  wildRank: number;
  dealCount: number;
  currentPlayerIdx: number;
  knockerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  config: ThreeThirteenConfig;
}

// --- Memory (神経衰弱) ---

/** Memory player data with pair count and captured-pair representative cards. */
export interface MemoryPlayerData {
  id: number;
  isHuman: boolean;
  pairCount: number;
  /** One representative card per captured pair, in ascending rank order. */
  pairs: Card[];
}

/** A card on the Memory game board. */
export interface MemoryBoardCard {
  card: Card | null;
  faceUp: boolean;
  taken: boolean;
}

/** Memory game configuration. */
export interface MemoryConfig {
  cpuDifficulty: number;
}

/** Full Memory game state returned from the API. */
export interface MemoryResponse extends BaseGameResponse {
  players: MemoryPlayerData[];
  board: MemoryBoardCard[];
  phase: number;
  currentPlayerIdx: number;
  firstFlipPos: number;
  secondFlipPos: number;
  lastMatchResult: boolean;
  gameEndFlag: boolean;
  winnerIdx: number;
  turnNumber: number;
  config: MemoryConfig;
}

// --- Klondike (ソリティア) ---

/** A card in a Klondike tableau column with face-up status. */
export interface KlondikeTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Klondike. */
export interface KlondikeHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Klondike game state returned from the API. */
export interface KlondikeResponse extends BaseGameResponse {
  tableau: KlondikeTableauCard[][];
  stockCount: number;
  waste: Card[];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  drawCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  score: number;
  scoringMode: number;
  hint?: KlondikeHint;
}

// --- Canfield (キャンフィールド) ---

/** A single card on a Canfield tableau column. */
export interface CanfieldTableauCard {
  card: Card;
}

/** A suggested move hint in Canfield. */
export interface CanfieldHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Canfield game state returned from the API. */
export interface CanfieldResponse extends BaseGameResponse {
  tableau: CanfieldTableauCard[][];
  reserve: Card[];
  stockCount: number;
  waste: Card[];
  foundation: Card[][];
  baseRank: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  hint?: CanfieldHint;
}

// --- Agnes Sorel (アグネス・ソレル) ---

/** A single card on an Agnes Sorel tableau column (carries face-up state). */
export interface AgnesTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Agnes Sorel. */
export interface AgnesHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Agnes Sorel game state returned from the API. */
export interface AgnesResponse extends BaseGameResponse {
  tableau: AgnesTableauCard[][];
  stockCount: number;
  foundation: Card[][];
  baseRank: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  hint?: AgnesHint;
}

// --- Osmosis (オズモシス / 浸透) ---

/** A suggested move hint in Osmosis. */
export interface OsmosisHint {
  fromZone: string;
  fromCol: number;
  toCol: number;
}

/** Full Osmosis game state returned from the API. */
export interface OsmosisResponse extends BaseGameResponse {
  reserve: Card[][];
  stockCount: number;
  waste: Card[];
  foundation: Card[][];
  baseRank: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  hint?: OsmosisHint;
}

// --- Bristol (ブリストル) ---

/** A suggested move hint in Bristol. */
export interface BristolHint {
  fromZone: string;
  fromCol: number;
  toZone: string;
  toCol: number;
}

/** Full Bristol game state returned from the API. */
export interface BristolResponse extends BaseGameResponse {
  tableau: Card[][];
  fan: Card[][];
  stockCount: number;
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  hint?: BristolHint;
}

/** Source or target zone for a Bristol card move. */
export interface BristolMoveZone {
  zone: string;
  col?: number;
}

// --- La Belle Lucie (ラ・ベル・ルーシー) ---

/** A suggested move hint in La Belle Lucie. */
export interface LaBelleLucieHint {
  fromFan: number;
  toFan: number;
  toFoundation: boolean;
}

/** Full La Belle Lucie game state returned from the API. */
export interface LaBelleLucieResponse extends BaseGameResponse {
  /** Fans of cards (top card is last); the count varies after a redeal. */
  fans: Card[][];
  /** The 4 foundations (A→K by suit). */
  foundation: Card[][];
  /** Remaining gather-and-reshuffle redeals (0–3). */
  redealsLeft: number;
  /** Current phase (0=Playing, 1=GameClear, 2=GameOver). */
  phase: number;
  moveCount: number;
  canUndo: boolean;
  hint?: LaBelleLucieHint;
}

// --- Simple Simon (シンプル・サイモン) ---

/** A suggested move hint in Simple Simon. */
export interface SimpleSimonHint {
  fromCol: number;
  cardIndex: number;
  toCol: number;
}

/** Full Simple Simon game state returned from the API. */
export interface SimpleSimonResponse extends BaseGameResponse {
  /** The 10 tableau columns (top card is last). */
  columns: Card[][];
  /** Number of complete K-A suits removed (0-4). */
  completedSuits: number;
  /** Current phase (0=Playing, 1=GameClear, 2=GameOver). */
  phase: number;
  moveCount: number;
  canUndo: boolean;
  hint?: SimpleSimonHint;
}

// --- Double Klondike (ダブル・クロンダイク) ---

/** A tableau card in Double Klondike; face-down cards hide their value. */
export interface DoubleKlondikeTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Double Klondike. */
export interface DoubleKlondikeHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Double Klondike game state returned from the API. */
export interface DoubleKlondikeResponse extends BaseGameResponse {
  /** The 9 tableau columns (top card is last). */
  tableau: DoubleKlondikeTableauCard[][];
  /** Cards left in the stock. */
  stockCount: number;
  /** The waste pile (top card is last). */
  waste: Card[];
  /** The 8 foundations (A-K by suit, two per suit). */
  foundation: Card[][];
  /** Current phase (0=Playing, 1=GameClear, 2=GameOver). */
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  hint?: DoubleKlondikeHint;
}

// --- Black Hole (ブラックホール) ---

/** A suggested move hint in Black Hole. */
export interface BlackHoleHint {
  fan: number;
}

/** Full Black Hole game state returned from the API. */
export interface BlackHoleResponse extends BaseGameResponse {
  /** The 17 fans (top card is last). */
  fans: Card[][];
  /** The central black hole pile (top card is last). */
  blackHole: Card[];
  /** Current phase (0=Playing, 1=GameClear, 2=GameOver). */
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  hint?: BlackHoleHint;
}

// --- FreeCell (フリーセル) ---

/** A suggested move hint in FreeCell. */
export interface FreeCellHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full FreeCell game state returned from the API. */
export interface FreeCellResponse extends BaseGameResponse {
  tableau: (Card | null)[][];
  freeCells: (Card | null)[];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: FreeCellHint;
}

// --- Eight Off (エイトオフ) ---

/** A suggested move hint in Eight Off. */
export interface EightOffHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Eight Off game state returned from the API. */
export interface EightOffResponse extends BaseGameResponse {
  tableau: (Card | null)[][];
  freeCells: (Card | null)[];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: EightOffHint;
}

// --- Penguin (ペンギン) ---

/** A suggested move hint in Penguin. */
export interface PenguinHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Penguin game state returned from the API. */
export interface PenguinResponse extends BaseGameResponse {
  tableau: (Card | null)[][];
  freeCells: (Card | null)[];
  foundation: Card[][];
  baseRank: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: PenguinHint;
}

// --- Seahaven Towers (シーヘイブンタワーズ) ---

/** A suggested move hint in Seahaven Towers. */
export interface SeahavenTowersHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Seahaven Towers game state returned from the API. */
export interface SeahavenTowersResponse extends BaseGameResponse {
  tableau: (Card | null)[][];
  reservedCells: (Card | null)[];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: SeahavenTowersHint;
}

/** Result of a Baccarat side bet (player pair, banker pair). */
export interface BaccaratSideBetResult {
  betType: number;
  resultType: number;
  resultName: string;
  betAmount: number;
  payout: number;
}

/** Full Baccarat game state returned from the API. */
export interface BaccaratResponse extends BaseGameResponse {
  playerHand: Card[];
  bankerHand: Card[];
  playerHandValue: number;
  bankerHandValue: number;
  phase: number;
  chips: number;
  betAmount: number;
  betType: number;
  result: number;
  payout: number;
  history: number[];
  playerPairBet: number;
  bankerPairBet: number;
  sideBetResults: BaccaratSideBetResult[];
}

// --- Napoleon (ナポレオン) ---

/** Napoleon player data with bid, roles, scores, and picture card count. */
export interface NapoleonPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  bid: number;
  isNapoleon: boolean;
  isAdjutant: boolean;
  adjutantRevealed: boolean;
  pictureCards: number;
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
}

/** A card played in a Napoleon trick. */
export interface NapoleonTrickCard {
  playerIdx: number;
  card: Card;
}

/** Napoleon game configuration. */
export interface NapoleonConfig {
  cpuDifficulty: number;
  minBid: number;
  pointLimit: number;
}

/** A suggested hint for Napoleon. */
export interface NapoleonHint {
  cardIndex?: number;
  bid?: number;
  trumpSuit?: number;
  adjutantSuit?: number;
  adjutantValue?: number;
  discardIndex?: number;
  reason: string;
}

/** Full Napoleon game state returned from the API. */
export interface NapoleonResponse extends BaseGameResponse {
  players: NapoleonPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  currentTrick: NapoleonTrickCard[];
  trumpSuit: number;
  adjutantCard: Card | null;
  napoleonIdx: number;
  adjutantIdx: number;
  adjutantRevealed: boolean;
  highestBid: number;
  highestBidder: number;
  kitty: Card[];
  gameEndFlag: boolean;
  winnerTeam: number;
  config: NapoleonConfig;
  hint?: NapoleonHint;
}

// --- Mighty (マイティ) ---

/** Mighty player data with bid, roles, scores, and point-card count. */
export interface MightyPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  bid: number;
  bidNoTrump: boolean;
  isDeclarer: boolean;
  isPartner: boolean;
  partnerRevealed: boolean;
  pointCards: number;
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
}

/** A card played in a Mighty trick. */
export interface MightyTrickCard {
  playerIdx: number;
  card: Card;
  isJokerLead?: boolean;
  leadDemandSuit?: number;
}

/** Mighty game configuration. */
export interface MightyConfig {
  cpuDifficulty: number;
  minBid: number;
  noTrumpExtra: number;
  pointLimit: number;
}

/** A suggested hint for Mighty. */
export interface MightyHint {
  cardIndex?: number;
  bid?: number;
  bidNoTrump?: boolean;
  trumpSuit?: number;
  partnerSuit?: number;
  partnerValue?: number;
  discardIndices?: number[];
  jokerLeadSuit?: number;
  reason: string;
}

/** Full Mighty game state returned from the API. */
export interface MightyResponse extends BaseGameResponse {
  players: MightyPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  currentTrick: MightyTrickCard[];
  trumpSuit: number;
  partnerCard?: Card | null;
  declarerIdx: number;
  partnerIdx: number;
  partnerRevealed: boolean;
  highestBid: number;
  highestBidder: number;
  winningBidNoTrump: boolean;
  kitty?: Card[];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  config: MightyConfig;
  hint?: MightyHint;
}

// --- 500 (Five Hundred) ---

/** A bid (contract) in 500. */
export interface FiveHundredBidData {
  kind: number;
  tricks: number;
  suit: number;
  value: number;
}

/** A 500 player's per-round state. */
export interface FiveHundredPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
  bid?: FiveHundredBidData | null;
  passed: boolean;
  isDeclarer: boolean;
}

/** A card played in a 500 trick. */
export interface FiveHundredTrickCard {
  playerIdx: number;
  card: Card;
}

/** 500 game configuration. */
export interface FiveHundredConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** A suggested hint for 500. */
export interface FiveHundredHint {
  bidKind?: number;
  bidTricks?: number;
  bidSuit?: number;
  pass?: boolean;
  discardIndices?: number[];
  cardIndex?: number;
  jokerSuit?: number;
  reason: string;
}

/** Full 500 game state returned from the API. */
export interface FiveHundredResponse extends BaseGameResponse {
  players: FiveHundredPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  leadPlayerIdx: number;
  trumpSuit: number;
  contractKind: number;
  contractTricks: number;
  contractValue: number;
  declarerIdx: number;
  highestBid?: FiveHundredBidData | null;
  highestBidder: number;
  jokerLeadSuit: number;
  kittyCount: number;
  currentTrick: FiveHundredTrickCard[];
  teamScores: [number, number];
  gameEndFlag: boolean;
  winnerTeam: number;
  config: FiveHundredConfig;
  hint?: FiveHundredHint;
}

// --- Rook (ルーク) ---

/** A Rook player's per-round state. */
export interface RookPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
  points: number;
  bid: number;
  passed: boolean;
  isDeclarer: boolean;
}

/** A card played in a Rook trick. */
export interface RookTrickCard {
  playerIdx: number;
  card: Card;
}

/** Rook game configuration. */
export interface RookConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** A suggested hint for Rook. */
export interface RookHint {
  bid?: number;
  pass?: boolean;
  discardIndices?: number[];
  trumpColor?: number;
  cardIndex?: number;
  reason: string;
}

/** Full Rook game state returned from the API. */
export interface RookResponse extends BaseGameResponse {
  players: RookPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  leadPlayerIdx: number;
  /** Declared trump color (1=red, 2=gold, 3=green, 4=black; -1 until declared). */
  trumpColor: number;
  contractBid: number;
  declarerIdx: number;
  highestBid: number;
  highestBidder: number;
  nestCount: number;
  /** Nest (widow) cards; only populated for the human declarer during nest exchange. */
  nest: Card[];
  currentTrick: RookTrickCard[];
  teamScores: [number, number];
  teamPoints: [number, number];
  gameEndFlag: boolean;
  winnerTeam: number;
  config: RookConfig;
  hint?: RookHint;
}

// --- Bid Whist (ビッド・ホイスト) ---

/** A Bid Whist bid: target tricks over the book plus a direction. */
export interface BidWhistBidData {
  tricks: number;
  direction: number;
}

/** A Bid Whist player's per-round state. */
export interface BidWhistPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
  bid?: BidWhistBidData | null;
  passed: boolean;
  isDeclarer: boolean;
}

/** A card played in a Bid Whist trick. */
export interface BidWhistTrickCard {
  playerIdx: number;
  card: Card;
}

/** Bid Whist game configuration. */
export interface BidWhistConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** A suggested hint for Bid Whist. */
export interface BidWhistHint {
  bidTricks?: number;
  bidDirection?: number;
  pass?: boolean;
  trumpSuit?: number;
  discardIndices?: number[];
  cardIndex?: number;
  reason: string;
}

/** Full Bid Whist game state returned from the API. */
export interface BidWhistResponse extends BaseGameResponse {
  players: BidWhistPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  leadPlayerIdx: number;
  trumpSuit: number;
  contractTricks: number;
  contractDirection: number;
  declarerIdx: number;
  highestBid?: BidWhistBidData | null;
  highestBidder: number;
  kittyCount: number;
  /**
   * Indices, within the human declarer's hand, of the six cards that came from
   * the kitty. Populated only while the human is exchanging during
   * KITTY_EXCHANGE; empty in every other phase and for a CPU declarer.
   */
  kittyIndices: number[];
  currentTrick: BidWhistTrickCard[];
  teamScores: [number, number];
  gameEndFlag: boolean;
  winnerTeam: number;
  config: BidWhistConfig;
  hint?: BidWhistHint;
}

// --- Skat (スカート) ---

/** A Skat player's per-round state. */
export interface SkatPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  bid: number;
  isDeclarer: boolean;
  cardPoints: number;
  roundsWon: number;
  roundsLost: number;
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
}

/** A card played in a Skat trick. */
export interface SkatTrickCard {
  playerIdx: number;
  card: Card;
}

/** Skat game configuration. */
export interface SkatConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** A suggested hint for Skat. */
export interface SkatHint {
  cardIndex?: number;
  bid?: number;
  gameType?: number;
  trumpSuit?: number;
  pickSkat?: boolean;
  discardIndex?: number;
  reason: string;
}

/** Full Skat game state returned from the API. */
export interface SkatResponse extends BaseGameResponse {
  players: SkatPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  currentTrick: SkatTrickCard[];
  forehandIdx: number;
  middlehandIdx: number;
  rearhandIdx: number;
  dealerIdx: number;
  declarerIdx: number;
  currentBid: number;
  activeBidActorIdx: number;
  gameType: number;
  trumpSuit: number;
  skat?: Card[];
  originalSkat?: Card[];
  pickedSkat: boolean;
  declarerCardPoints: number;
  defendersCardPoints: number;
  winnerSide: number;
  gameValue: number;
  gameEndFlag: boolean;
  leadPlayerIdx: number;
  config: SkatConfig;
  hint?: SkatHint;
}

// --- Shithead / Karma (シットヘッド / カーマ) ---

/** A Shithead player's per-game state. */
export interface ShitheadPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  rank: number;
  handCount: number;
  handCards: Card[];
  faceUpCards: Card[];
  faceDownCount: number;
}

/** A single Shithead action (play or pickup). */
export interface ShitheadAction {
  playerIdx: number;
  source: string;
  playedCards: Card[];
  pickup: boolean;
  burned: boolean;
  skipped: boolean;
}

/** Shithead local rule configuration. */
export interface ShitheadConfig {
  magicTwo: boolean;
  magicSeven: boolean;
  magicEight: boolean;
  magicTen: boolean;
  fourOfAKindBurn: boolean;
  cpuDifficulty: number;
}

/** Full Shithead game state returned from the API. */
export interface ShitheadResponse extends BaseGameResponse {
  players: ShitheadPlayerData[];
  currentTurn: number;
  currentSource: string;
  discardPile: Card[];
  stockSize: number;
  skipNext: boolean;
  sevenActive: boolean;
  gameEndFlag: boolean;
  config: ShitheadConfig;
  cpuActions: ShitheadAction[];
  humanAction?: ShitheadAction;
}

// --- Spider Solitaire (スパイダーソリティア) ---

/** A suggested move hint in Spider Solitaire. */
export interface SpiderHint {
  fromCol: number;
  cardIndex: number;
  toCol: number;
}

/** Tableau card with face-up state in Spider. */
export interface SpiderTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** Full Spider Solitaire game state returned from the API. */
export interface SpiderResponse extends BaseGameResponse {
  tableau: SpiderTableauCard[][];
  stockCount: number;
  completedSuits: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  score: number;
  difficulty: number;
  hint?: SpiderHint;
}

// --- Spiderette (スパイダレット) ---

/** Hint returned by the Spiderette /hint endpoint. */
export interface SpideretteHint {
  fromCol: number;
  cardIndex: number;
  toCol: number;
}

/** Tableau card with face-up state in Spiderette. */
export interface SpideretteTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** Full Spiderette Solitaire game state returned from the API. */
export interface SpideretteResponse extends BaseGameResponse {
  tableau: SpideretteTableauCard[][];
  stockCount: number;
  completedSuits: number;
  score: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: SpideretteHint;
}

// --- Indian Poker (インディアンポーカー) ---

/** Indian Poker player data with card, chips, and betting status. */
export interface IndianPokerPlayerOutput {
  id: number;
  isHuman: boolean;
  card: Card | null;
  chips: number;
  currentBet: number;
  folded: boolean;
  allIn: boolean;
  cardRank: number;
  playStyleName: string;
}

/** Indian Poker round result for a single player. */
export interface IndianPokerResultOutput {
  playerIdx: number;
  card: Card | null;
  cardRank: number;
  wonAmount: number;
}

/** CPU betting action in Indian Poker. */
export interface IndianPokerCpuActionOutput {
  playerIdx: number;
  action: number;
  amount: number;
}

/** Side pot in Indian Poker with eligible players. */
export interface IndianPokerSidePot {
  amount: number;
  eligiblePlayers: number[];
}

/** Full Indian Poker game state returned from the API. */
export interface IndianPokerResponse extends BaseGameResponse {
  players: IndianPokerPlayerOutput[];
  pot: number;
  sidePots: IndianPokerSidePot[];
  dealerIdx: number;
  currentTurn: number;
  phase: number;
  gameEndFlag: boolean;
  lastBet: number;
  minRaise: number;
  bettingLimit: number;
  raiseCount: number;
  maxBetAmount: number;
  roundResults: IndianPokerResultOutput[];
  cpuActions: IndianPokerCpuActionOutput[];
  handCount: number;
  ante: number;
  actionLog?: ActionLogEntry[];
  metaAI?: IndianPokerMetaAI;
  profile?: IndianPokerHumanProfileData;
}

// --- Euchre (ユーカー) ---

/** Euchre player data with team, trick count, and hand. */
export interface EuchrePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
}

/** A card played in a Euchre trick. */
export interface EuchreTrickCard {
  playerIdx: number;
  card: Card;
}

/** Euchre game configuration. */
export interface EuchreConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A suggested hint for Euchre. */
export interface EuchreHint {
  cardIndex?: number;
  orderUp?: boolean;
  suit?: number;
  goAlone?: boolean;
  reason: string;
}

/** Full Euchre game state returned from the API. */
export interface EuchreResponse extends BaseGameResponse {
  players: EuchrePlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  trumpSuit: number;
  faceUpCard: Card | null;
  makerTeam: number;
  goingAlone: boolean;
  goingAlonePlayerIdx: number;
  currentTrick: EuchreTrickCard[];
  teamScores: number[];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  config: EuchreConfig;
  hint?: EuchreHint;
}

// --- Belote (ベロート) ---

/** Belote player data with team, trick count, and hand. */
export interface BelotePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
}

/** A card played in a Belote trick. */
export interface BeloteTrickCard {
  playerIdx: number;
  card: Card;
}

/** Belote game configuration. */
export interface BeloteConfig {
  cpuDifficulty: number;
  targetScore: number;
  dixDeDer: number;
  enableBeloteRebelote: boolean;
}

/** A suggested hint for Belote. */
export interface BeloteHint {
  cardIndex?: number;
  orderUp?: boolean;
  suit?: number;
  reason: string;
}

/** Full Belote game state returned from the API. */
export interface BeloteResponse extends BaseGameResponse {
  players: BelotePlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  trumpSuit: number;
  faceUpCard: Card | null;
  makerTeam: number;
  makerPlayerIdx: number;
  currentTrick: BeloteTrickCard[];
  teamScores: number[];
  roundPoints: number[];
  roundBeloteBonus: number[];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  config: BeloteConfig;
  hint?: BeloteHint;
}

// --- Jass (Schieber) (ヤス) ---

/** Jass player data with team, trick count, and hand. */
export interface JassPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
}

/** A card played in a Jass trick. */
export interface JassTrickCard {
  playerIdx: number;
  card: Card;
}

/** Jass game configuration. */
export interface JassConfig {
  cpuDifficulty: number;
  targetScore: number;
  lastTrickBonus: number;
  enableWeis: boolean;
}

/** A suggested hint for Jass. */
export interface JassHint {
  cardIndex?: number;
  schieben?: boolean;
  suit?: number;
  reason: string;
}

/** Full Jass (Schieber) game state returned from the API. */
export interface JassResponse extends BaseGameResponse {
  players: JassPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  forehandIdx: number;
  trumpSuit: number;
  schieben: boolean;
  makerTeam: number;
  makerPlayerIdx: number;
  currentTrick: JassTrickCard[];
  lastTrick: JassTrickCard[];
  lastTrickWinner: number;
  teamScores: number[];
  roundPoints: number[];
  roundWeisPoints: number[];
  roundStockPoints: number[];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  config: JassConfig;
  hint?: JassHint;
}

// --- Watten (ヴァッテン) ---

/** A Watten player's public/own state. Cards are non-empty only for the human during play. */
export interface WattenPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
}

/** A card played into the current Watten trick. */
export interface WattenTrickCard {
  playerIdx: number;
  card: Card;
}

/** Watten game configuration. */
export interface WattenConfig {
  cpuDifficulty: number;
  targetScore: number;
  maxRaises: number;
}

/** A suggested hint for Watten, computed by the backend. */
export interface WattenHint {
  /** The suggested action: declare, raise, play, hold, or fold. */
  action: string;
  cardIndex?: number;
  rank?: number;
  suit?: number;
  reason: string;
}

/**
 * Full Watten (ヴァッテン) game state returned from the API.
 *
 * Watten is a Bavarian/Austrian 4-player, 2-team trick-taker with a raise/bluff
 * stake mechanic. Seats 0 & 2 form team 0, seats 1 & 3 form team 1; the human is
 * seat 0. The dealer declares a Schlag rank and a critical (trump) suit, teams
 * play five tricks, and either team may raise the stake for the other to hold or
 * fold. Winning at least three tricks scores the stake; first team to the target
 * score wins.
 */
export interface WattenResponse extends BaseGameResponse {
  players: WattenPlayer[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  dealerIdx: number;
  leadPlayerIdx: number;
  /** The declared Schlag rank (1=A, 7..13), or 0 when unset. */
  schlagRank: number;
  /** The declared critical (trump) suit (1=♠ 2=♣ 3=♥ 4=♦), or 0 when unset. */
  criticalSuit: number;
  /** The current accepted stake (starts at 2). */
  stake: number;
  /** The proposed stake after a raise, awaiting a hold/fold response (0 when none). */
  pendingStake: number;
  raiseCount: number;
  /** The team that proposed the pending raise, or -1 when none. */
  raiserTeam: number;
  /** Seat index of the player who must hold/fold a pending raise, or -1 when none. */
  responderIdx: number;
  /** Whether the human (as lead) may raise the stake right now. */
  canRaise: boolean;
  currentTrick: WattenTrickCard[];
  teamScores: number[];
  teamTricks: number[];
  /** The team that won the most recent completed deal, or -1 until decided. */
  dealWinnerTeam: number;
  gameEndFlag: boolean;
  winnerTeam: number;
  /** Match result from the human's (team 0) perspective: -1 lose, 0 none, 1 win. */
  result: number;
  hint?: WattenHint | null;
  config: WattenConfig;
}

// --- Gaigel (ガイゲル) ---

/** Gaigel player data with team, trick count, and hand. */
export interface GaigelPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
}

/** A card played in a Gaigel trick. */
export interface GaigelTrickCard {
  playerIdx: number;
  card: Card;
}

/** Gaigel game configuration. */
export interface GaigelConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** A suggested hint for Gaigel. */
export interface GaigelHint {
  cardIndex?: number;
  reason: string;
  isMarriage: boolean;
}

/** Full Gaigel game state returned from the API. */
export interface GaigelResponse extends BaseGameResponse {
  players: GaigelPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  dealerIdx: number;
  trumpSuit: number;
  trumpCard?: Card;
  stockRemaining: number;
  currentTrick: GaigelTrickCard[];
  teamScores: number[];
  roundPoints: number[];
  roundMarriage: number[];
  marriageIndices: number[];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  config: GaigelConfig;
  hint?: GaigelHint;
}

// --- Contract Bridge (コントラクトブリッジ) ---

/** Bridge player data with team, trick count, and hand. */
export interface BridgePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
}

/** A card played in a Bridge trick. */
export interface BridgeTrickCard {
  playerIdx: number;
  card: Card;
}

/** A bid entry in the Bridge bid history. */
export interface BridgeBidEntry {
  playerIdx: number;
  bidType: number;
  level: number;
  suit: number;
}

/** Bridge game configuration. */
export interface BridgeConfig {
  cpuDifficulty: number;
}

/** A suggested hint for Bridge. */
export interface BridgeHint {
  cardIndex?: number;
  bidType?: number;
  bidLevel?: number;
  bidSuit?: number;
  reason: string;
}

/** Full Bridge game state returned from the API. */
export interface BridgeResponse extends BaseGameResponse {
  players: BridgePlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  trumpSuit: number;
  contractLevel: number;
  contractSuit: number;
  doubled: number;
  declarerIdx: number;
  dummyIdx: number;
  bidHistory: BridgeBidEntry[];
  vulnerability: boolean[];
  currentTrick: BridgeTrickCard[];
  teamScores: number[];
  gamesWon: number[];
  belowLine: number[];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  openingLeadDone: boolean;
  dummyHand: Card[] | null;
  config: BridgeConfig;
  hint?: BridgeHint;
}

// --- Pyramid Solitaire (ピラミッド) ---

/** A card in the pyramid with removal and exposure status. */
export interface PyramidCard {
  card: Card | null;
  removed: boolean;
  exposed: boolean;
}

/** A suggested pair/king removal hint in Pyramid. */
export interface PyramidHint {
  type: string;
  row1: number;
  col1: number;
  row2: number;
  col2: number;
}

/** Full Pyramid game state returned from the API. */
export interface PyramidResponse extends BaseGameResponse {
  pyramid: PyramidCard[][];
  stockCount: number;
  waste: Card[];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: PyramidHint;
}

// --- TriPeaks (トリピークス) ---

/** A card in the TriPeaks tableau with removal and exposure status. */
export interface TriPeaksCard {
  card: Card | null;
  removed: boolean;
  exposed: boolean;
}

/** A suggested hint in TriPeaks. */
export interface TriPeaksHint {
  type: string;
  row: number;
  col: number;
}

/** Full TriPeaks game state returned from the API. */
export interface TriPeaksResponse extends BaseGameResponse {
  layout: TriPeaksCard[][];
  stockCount: number;
  waste: Card[];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: TriPeaksHint;
}

/** Full Video Poker game state returned from the API. */
export interface VideoPokerResponse extends BaseGameResponse {
  hand: Card[];
  phase: number;
  chips: number;
  betAmount: number;
  result: number;
  payout: number;
  handRank: number;
  handName: string;
  heldIndices: boolean[];
  variantName: string;
}

// --- Cribbage (クリベッジ) ---

/** Cribbage player data with scores. */
export interface CribbagePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
}

/** Cribbage score detail breakdown. */
export interface CribbageScoreDetail {
  fifteens: number;
  pairs: number;
  runs: number;
  flush: number;
  nobs: number;
  total: number;
}

/** Cribbage game configuration. */
export interface CribbageConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** Full Cribbage game state returned from the API. */
export interface CribbageResponse extends BaseGameResponse {
  players: CribbagePlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  dealerIdx: number;
  crib: Card[];
  starter: Card | null;
  pegCount: number;
  pegPlayedCards: Card[];
  showPhaseStep: number;
  handScoreDetails: (CribbageScoreDetail | null)[];
  gameEndFlag: boolean;
  winnerIdx: number;
  config: CribbageConfig;
}

// --- Oh Hell (オー・ヘル) ---

/** Oh Hell player data with scores. */
export interface OhHellPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  bid: number;
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
}

/** A card played in an Oh Hell trick. */
export interface OhHellTrickCard {
  playerIdx: number;
  card: Card;
}

/** Oh Hell game configuration. */
export interface OhHellConfig {
  cpuDifficulty: number;
  maxHandSize: number;
  scoringVariant: number;
  roundDirection: number;
}

/** A suggested hint for Oh Hell. */
export interface OhHellHint {
  cardIndex?: number;
  bid?: number;
  reason: string;
}

/** Full Oh Hell game state returned from the API. */
export interface OhHellResponse extends BaseGameResponse {
  players: OhHellPlayerData[];
  phase: number;
  roundNumber: number;
  totalRounds: number;
  handSize: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  currentTrick: OhHellTrickCard[];
  trumpCard: Card | null;
  trumpSuit: number;
  restrictedBid: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  leadPlayerIdx: number;
  hint?: OhHellHint;
  config: OhHellConfig;
}

// --- Wizard (ウィザード) ---

/** Wizard player data with scores. */
export interface WizardPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  bid: number;
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
}

/** A card played in a Wizard trick. */
export interface WizardTrickCard {
  playerIdx: number;
  card: Card;
}

/** Wizard game configuration. */
export interface WizardConfig {
  cpuDifficulty: number;
}

/** A suggested hint for Wizard. */
export interface WizardHint {
  cardIndex?: number;
  bid?: number;
  reason: string;
}

/** Full Wizard game state returned from the API. */
export interface WizardResponse extends BaseGameResponse {
  players: WizardPlayerData[];
  phase: number;
  roundNumber: number;
  totalRounds: number;
  handSize: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  currentTrick: WizardTrickCard[];
  trumpCard: Card | null;
  trumpSuit: number;
  restrictedBid: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  leadPlayerIdx: number;
  hint?: WizardHint;
  config: WizardConfig;
}

// --- Ninety-Nine (ナインティナイン) ---

/** Ninety-Nine player data with scores. */
export interface NinetyNinePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  bid: number;
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
  buriedCount: number;
}

/** A card played in a Ninety-Nine trick. */
export interface NinetyNineTrickCard {
  playerIdx: number;
  card: Card;
}

/** Ninety-Nine game configuration. */
export interface NinetyNineConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** A suggested hint for Ninety-Nine. */
export interface NinetyNineHint {
  cardIndex?: number;
  buryIndices?: number[];
  reason: string;
}

/** Full Ninety-Nine game state returned from the API. */
export interface NinetyNineResponse extends BaseGameResponse {
  players: NinetyNinePlayerData[];
  phase: number;
  dealNumber: number;
  targetScore: number;
  handSize: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  trumpSuit: number;
  currentTrick: NinetyNineTrickCard[];
  gameEndFlag: boolean;
  winnerIdx: number;
  leadPlayerIdx: number;
  hint?: NinetyNineHint;
  config: NinetyNineConfig;
}

// --- Three Card Poker (スリーカードポーカー) ---

/** Three Card Poker API response. */
export interface ThreeCardResponse extends BaseGameResponse {
  playerHand: Card[];
  dealerHand: Card[];
  phase: number;
  chips: number;
  anteBet: number;
  pairPlusBet: number;
  playBet: number;
  result: number;
  antePayout: number;
  playPayout: number;
  anteBonusPayout: number;
  pairPlusPayout: number;
  totalPayout: number;
  dealerQualified: boolean;
  playerHandRank: number;
  dealerHandRank: number;
}

// --- Caribbean Stud Poker (カリビアンスタッドポーカー) ---

/** Caribbean Stud Poker API response. */
export interface CaribbeanStudResponse extends BaseGameResponse {
  playerHand: Card[];
  /** Dealer hand: during the action phase only the first card is revealed and
   * the remaining slots are `MaskedCard`. After the end phase all 5 are real `Card`s. */
  dealerHand: (Card | MaskedCard)[];
  phase: number;
  chips: number;
  anteBet: number;
  jackpotBet: number;
  playBet: number;
  result: number;
  antePayout: number;
  playPayout: number;
  jackpotPayout: number;
  totalPayout: number;
  dealerQualified: boolean;
  playerHandRank: number;
  dealerHandRank: number;
}

// --- Casino Hold'em (カジノホールデム) ---

/** Casino Hold'em API response. */
export interface CasinoHoldemResponse extends BaseGameResponse {
  /** Player's two hole cards. */
  playerHand: Card[];
  /** Dealer's hole cards: masked as `MaskedCard` until the showdown (only after a call). */
  dealerHand: (Card | MaskedCard)[];
  /** Community cards (flop / turn / river). Length grows from 3 (flop) → 5 (showdown). */
  community: Card[];
  phase: number;
  chips: number;
  anteBet: number;
  /** AA Bonus side bet wager. */
  bonusBet: number;
  /** Call bet placed at flop (2× ante). 0 if folded. */
  callBet: number;
  result: number;
  /** Whether the dealer qualified (Pair of Fours or better). */
  dealerQualify: boolean;
  antePayout: number;
  callPayout: number;
  bonusPayout: number;
  totalPayout: number;
  playerHandRank: number;
  dealerHandRank: number;
}

// --- Texas Hold'em Bonus Poker (テキサスホールデムボーナスポーカー) ---

/** Texas Hold'em Bonus Poker API response. */
export interface TexasHoldemBonusResponse extends BaseGameResponse {
  /** Player's two hole cards. */
  playerHand: Card[];
  /** Dealer's hole cards: masked as `MaskedCard` until the showdown. */
  dealerHand: (Card | MaskedCard)[];
  /** Community cards (flop / turn / river). Length grows from 0 → 5 over phases. */
  community: Card[];
  phase: number;
  chips: number;
  anteBet: number;
  bonusBet: number;
  flopBet: number;
  turnBet: number;
  riverBet: number;
  totalPlayBet: number;
  result: number;
  antePayout: number;
  playPayout: number;
  bonusPayout: number;
  totalPayout: number;
  playerHandRank: number;
  dealerHandRank: number;
}

// --- Ultimate Texas Hold'em (アルティメット・テキサスホールデム) ---

/** Ultimate Texas Hold'em API response. */
export interface UltimateTexasHoldemResponse extends BaseGameResponse {
  /** Player's two hole cards. */
  playerHand: Card[];
  /** Dealer's hole cards: masked as `MaskedCard` until the showdown. */
  dealerHand: (Card | MaskedCard)[];
  /** Community cards (flop / turn / river). Length grows from 0 → 5 over phases. */
  community: Card[];
  phase: number;
  chips: number;
  anteBet: number;
  blindBet: number;
  tripsBet: number;
  playBet: number;
  folded: boolean;
  result: number;
  dealerQualified: boolean;
  antePayout: number;
  blindPayout: number;
  playPayout: number;
  tripsPayout: number;
  totalPayout: number;
  playerHandRank: number;
  dealerHandRank: number;
}

// --- Mississippi Stud (ミシシッピ・スタッド) ---

/** Mississippi Stud API response. */
export interface MississippiStudResponse extends BaseGameResponse {
  /** Player's two hole cards (revealed once the round starts). */
  playerHand: Card[];
  /** Community cards: masked as `MaskedCard` until the matching street is revealed. */
  communityCards: (Card | MaskedCard)[];
  /** Per-card reveal state for `communityCards` (length 3). */
  communityRevealed: boolean[];
  phase: number;
  chips: number;
  anteAmount: number;
  /** 3rd / 4th / 5th street bet multipliers (0=未ベット, 1/2/3=倍率). Length 3. */
  streetMultipliers: number[];
  folded: boolean;
  totalBet: number;
  result: number;
  handRank: number;
  /** Applied payout multiplier (-1=push, 0=loss, positive=win). */
  payoutMultiplier: number;
  antePayout: number;
  streetPayouts: number[];
  totalPayout: number;
}

// --- Pai Gow Poker (パイゴウポーカー) ---

/** Pai Gow Poker API response. */
export interface PaiGowResponse extends BaseGameResponse {
  playerCards: Card[];
  dealerCards: Card[];
  playerHighHand: Card[];
  playerLowHand: Card[];
  dealerHighHand: Card[];
  dealerLowHand: Card[];
  phase: number;
  chips: number;
  bet: number;
  result: number;
  highHandResult: number;
  lowHandResult: number;
  payout: number;
  commission: number;
  playerHighRank: number;
  playerLowRank: number;
  dealerHighRank: number;
  dealerLowRank: number;
}

/** Chinese Poker game state response. */
export interface ChinesePokerResponse extends BaseGameResponse {
  playerCards: Card[];
  dealerCards: Card[];
  playerFront: Card[];
  playerMiddle: Card[];
  playerBack: Card[];
  dealerFront: Card[];
  dealerMiddle: Card[];
  dealerBack: Card[];
  phase: number;
  chips: number;
  bet: number;
  result: number;
  frontResult: number;
  middleResult: number;
  backResult: number;
  payout: number;
  playerFrontRank: number;
  playerMiddleRank: number;
  playerBackRank: number;
  dealerFrontRank: number;
  dealerMiddleRank: number;
  dealerBackRank: number;
  playerRoyalty: number;
  dealerRoyalty: number;
  scoop: boolean;
}

/** Speed player data with hand and draw pile info. */
export interface SpeedPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  drawPileSize: number;
}

/** Speed CPU action record. */
export interface SpeedCpuAction {
  cardIndex: number;
  pileIndex: number;
}

/** Speed hint information. */
export interface SpeedHint {
  cardIndex: number;
  pileIndex: number;
  found: boolean;
}

/** Speed game configuration. */
export interface SpeedConfig {
  cpuDifficulty: number;
  autoFlip: boolean;
}

/** War player data with face-down pile and discard pile sizes. */
export interface WarPlayerData {
  id: number;
  isHuman: boolean;
  drawPileSize: number;
  discardPileSize: number;
  totalCards: number;
}

/** War game configuration. */
export interface WarConfig {
  maxRounds: number;
}

/** Full War game state returned from the API. */
export interface WarResponse extends BaseGameResponse {
  players: WarPlayerData[];
  phase: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  playerRevealed: Card | null;
  cpuRevealed: Card | null;
  warPotSize: number;
  lastWinnerIdx: number;
  lastBurialCount: number;
  roundsPlayed: number;
  config: WarConfig;
}

/** Full Speed game state returned from the API. */
export interface SpeedResponse extends BaseGameResponse {
  players: SpeedPlayerData[];
  centerPiles: Card[];
  phase: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  cpuActions?: SpeedCpuAction[];
  hint?: SpeedHint;
  config: SpeedConfig;
}

/** Go Fish player data with hand, book count, and completed books. */
export interface GoFishPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  bookCount: number;
  books: GoFishBook[];
}

/** A completed 4-of-a-kind book in Go Fish. */
export interface GoFishBook {
  rank: number;
  cards: Card[];
}

/** CPU action record in Go Fish. */
export interface GoFishCpuAction {
  askPlayerIdx: number;
  askTargetIdx: number;
  askRank: number;
  success: boolean;
  cardsReceived: number;
  drawnCard: Card | null;
  bookFormed: boolean;
  bookRank: number;
}

/** Information about the last ask action in Go Fish. */
export interface GoFishLastAsk {
  playerIdx: number;
  targetIdx: number;
  rank: number;
  success: boolean;
  cardsReceived: Card[];
  drawnCard: Card | null;
  bookFormed: boolean;
  bookRank: number;
}

/** Go Fish game configuration. */
export interface GoFishConfig {
  cpuDifficulty: number;
}

/** Full Go Fish game state returned from the API. */
export interface GoFishResponse extends BaseGameResponse {
  players: GoFishPlayerData[];
  phase: number;
  currentTurn: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  turnNumber: number;
  deckRemaining: number;
  lastAsk: GoFishLastAsk | null;
  cpuActions: GoFishCpuAction[];
  humanAction: GoFishCpuAction | null;
  config: GoFishConfig;
}

// --- Canasta (カナスタ) ---

/** Canasta game configuration. */
export interface CanastaConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A single meld on the table in Canasta. */
export interface CanastaMeldData {
  cards: Card[];
  isNatural: boolean;
  isCanasta: boolean;
  rank: number;
}

/** Canasta player data with melds and red 3s. */
export interface CanastaPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  melds: CanastaMeldData[];
  red3Count: number;
  red3s: Card[];
  roundScore: number;
  cumulativeScore: number;
  hasCanasta: boolean;
  hasInitMeld: boolean;
}

/** Full Canasta game state returned from the API. */
export interface CanastaResponse extends BaseGameResponse {
  players: CanastaPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  discardPileCount: number;
  isFrozen: boolean;
  gameEndFlag: boolean;
  winnerIdx: number;
  config: CanastaConfig;
}

// --- Samba (サンバ) ---

/** Samba game configuration. */
export interface SambaConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/**
 * A single meld on the table in Samba. `kind` distinguishes same-rank sets
 * (0) from suited sequences (1); `isCanasta`/`isSamba` flag the completed
 * seven-card set (canasta) and seven-card sequence (samba) respectively.
 */
export interface SambaMeldData {
  cards: Card[];
  kind: number; // 0 = set, 1 = sequence
  isNatural: boolean;
  isCanasta: boolean;
  isSamba: boolean;
  rank: number;
}

/** Samba player data with melds, red 3s, and partnership (team) affiliation. */
export interface SambaPlayerData {
  id: number;
  team: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  melds: SambaMeldData[];
  red3Count: number;
  red3s: Card[];
  roundScore: number;
  cumulativeScore: number;
  hasCanasta: boolean;
  hasSamba: boolean;
  hasInitMeld: boolean;
}

/** Full Samba game state returned from the API. */
export interface SambaResponse extends BaseGameResponse {
  players: SambaPlayerData[];
  teamScores: number[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  discardPileCount: number;
  isFrozen: boolean;
  gameEndFlag: boolean;
  winnerIdx: number;
  config: SambaConfig;
}

// --- Hand and Foot (ハンド・アンド・フット) ---

/** Hand and Foot game configuration. */
export interface HandAndFootConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A single team meld on the table in Hand and Foot. */
export interface HandAndFootMeldData {
  cards: Card[];
  isNatural: boolean;
  isCanasta: boolean;
  rank: number;
}

/** Per-team meld and red-3 data in Hand and Foot. */
export interface HandAndFootTeamData {
  team: number;
  melds: HandAndFootMeldData[];
  red3Count: number;
  red3s: Card[];
}

/** Hand and Foot player data. Melds and red 3s are held per team, not per player. */
export interface HandAndFootPlayerData {
  id: number;
  team: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  footCount: number;
  inFoot: boolean;
  roundScore: number;
  cumulativeScore: number;
}

/** Full Hand and Foot game state returned from the API. */
export interface HandAndFootResponse extends BaseGameResponse {
  players: HandAndFootPlayerData[];
  teams: HandAndFootTeamData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  discardPileCount: number;
  isFrozen: boolean;
  gameEndFlag: boolean;
  winnerTeam: number;
  config: HandAndFootConfig;
}

// --- Burraco (ブラーコ) ---

/** Burraco game configuration. */
export interface BurracoConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A single meld on the table in Burraco. */
export interface BurracoMeldData {
  cards: Card[];
  isNatural: boolean;
  isBurraco: boolean;
  rank: number;
}

/** Burraco player data with melds, red 3s, and pozzetto status. */
export interface BurracoPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  melds: BurracoMeldData[];
  red3Count: number;
  red3s: Card[];
  roundScore: number;
  cumulativeScore: number;
  hasBurraco: boolean;
  hasInitMeld: boolean;
  tookPozzetto: boolean;
}

/** Full Burraco game state returned from the API. */
export interface BurracoResponse extends BaseGameResponse {
  players: BurracoPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  /** The full discard pile, oldest (bottom) first. In Burraco the whole pile is
   * taken at once, so its contents are public information for all players. */
  discardPile: Card[];
  drawPileCount: number;
  discardPileCount: number;
  pozzettoCount: number;
  isFrozen: boolean;
  gameEndFlag: boolean;
  winnerIdx: number;
  config: BurracoConfig;
}

/** Pinochle game configuration. */
export interface PinochleConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** Pinochle meld data. */
export interface PinochleMeldData {
  type: number;
  points: number;
  cards: Card[];
}

/** Pinochle trick card data. */
export interface PinochleTrickCard {
  playerIdx: number;
  card: Card;
}

/** Pinochle player data. */
export interface PinochlePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
  bid: number;
  hasPassed: boolean;
  meldScore: number;
  trickPoints: number;
}

/** Full Pinochle game state returned from the API. */
export interface PinochleResponse extends BaseGameResponse {
  players: PinochlePlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  trumpSuit: number;
  highestBid: number;
  highestBidder: number;
  currentTrick: PinochleTrickCard[];
  teamScores: [number, number];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  playerMelds: PinochleMeldData[][];
  validPlayIndices?: number[];
  hint?: {
    cardIndex?: number;
    bidAmount?: number;
    pass?: boolean;
    suit?: number;
    reason: string;
  };
  config: PinochleConfig;
}

// --- Piquet (ピケ) ---

/** Piquet game configuration. */
export interface PiquetConfig {
  cpuDifficulty: number;
  dealsPerPartie: number;
}

/** Piquet player data. */
export interface PiquetPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  declScore: number;
  trickScore: number;
  bonusScore: number;
  roundScore: number;
  matchScore: number;
}

/** Piquet trick card data. */
export interface PiquetTrickCard {
  playerIdx: number;
  card: Card;
}

/** Piquet claim (Point/Sequence/Set declaration evidence). */
export interface PiquetClaim {
  length: number;
  topRank: number;
  pipTotal: number;
  suit: number;
  cards: Card[];
}

/** Piquet declaration result. */
export interface PiquetDeclaration {
  kind: number;
  elderClaim?: PiquetClaim;
  youngerClaim?: PiquetClaim;
  winner: number;
  score: number;
  scoredBy: number;
  sets?: PiquetClaim[];
}

/** Full Piquet game state returned from the API. */
export interface PiquetResponse extends BaseGameResponse {
  players: PiquetPlayerData[];
  phase: number;
  dealNumber: number;
  dealsPerPartie: number;
  elderIdx: number;
  youngerIdx: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  trickNumber: number;
  tricksWon: [number, number];
  exchangeTurn: number;
  elderExchangedCnt: number;
  youngerExchangedCnt: number;
  elderTalon: Card[];
  youngerTalon: Card[];
  elderRevealedTalon: Card[];
  youngerRevealedTalon: Card[];
  carteBlanche: [boolean, boolean];
  declStage: number;
  declResults: PiquetDeclaration[];
  currentTrick: PiquetTrickCard[];
  legalPlayIndices?: number[];
  gameEndFlag: boolean;
  winnerIdx: number;
  hint?: {
    cardIndex?: number;
    discardIndices?: number[];
    reason: string;
  };
  config: PiquetConfig;
}

// --- Golf Solitaire (ゴルフ) ---

/** A card in the Golf tableau with removal and exposure status. */
export interface GolfCard {
  card: Card | null;
  removed: boolean;
  exposed: boolean;
}

/** A suggested hint in Golf Solitaire. */
export interface GolfHint {
  type: string;
  col: number;
}

/** Full Golf Solitaire game state returned from the API. */
export interface GolfResponse extends BaseGameResponse {
  layout: GolfCard[][];
  stockCount: number;
  waste: Card[];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: GolfHint;
}

// --- Aces Up (四つ葉のクローバー) ---

/** A card in an Aces Up column with action availability flags. */
export interface AcesUpCard {
  card: Card;
  top: boolean;
  removable: boolean;
  movable: boolean;
}

/** A suggested hint in Aces Up. */
export interface AcesUpHint {
  type: 'remove' | 'move' | 'draw';
  col: number;
}

/** Full Aces Up game state returned from the API. */
export interface AcesUpResponse extends BaseGameResponse {
  columns: AcesUpCard[][];
  stockCount: number;
  discardCount: number;
  /** The most recently removed card (top of the discard pile); absent when nothing has been discarded. */
  discardTop?: Card | null;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: AcesUpHint;
}

// --- Pig's Tail ---

/** Pig's Tail player output from the server. */
export interface PigsTailPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
}

/** Pig's Tail CPU action record. */
export interface PigsTailCpuAction {
  drawPlayerIdx: number;
  drawnCard: Card | null;
  penaltyFlag: boolean;
  penaltyCount: number;
  hesitationMs?: number;
}

/** Pig's Tail game state response. */
export interface PigsTailResponse extends BaseGameResponse {
  players: PigsTailPlayer[];
  circleCount: number;
  centerTop: Card | null;
  centerCount: number;
  currentTurn: number;
  gameEndFlag: boolean;
  loserIdx: number;
  lastDrawCard: Card | null;
  lastPenalty: boolean;
  cpuActions: PigsTailCpuAction[];
  humanAction: PigsTailCpuAction | null;
}

// --- Seven Card Stud ---

/** Player data in Seven Card Stud. */
export interface SevenCardStudPlayerData {
  id: number;
  isHuman: boolean;
  holeCards: Card[];
  doorCards: Card[];
  chips: number;
  currentBet: number;
  folded: boolean;
  allIn: boolean;
  handRank: number;
  handName: string;
  bestHand: Card[];
  playStyleName: string;
  totalHands: number;
  vpip: number;
  pfr: number;
  threeBet: number;
  af: string;
}

/** CPU betting action in Seven Card Stud. */
export interface SevenCardStudCpuAction {
  playerIdx: number;
  action: number;
  amount: number;
}

/** Seven Card Stud round result for a single player. */
export interface SevenCardStudResult {
  playerIdx: number;
  handRank: number;
  handName: string;
  kickers: string;
  bestHand: Card[];
  wonAmount: number;
  mucked: boolean;
}

/** Side pot in Seven Card Stud with eligible players. */
export interface SevenCardStudSidePot {
  amount: number;
  eligiblePlayers: number[];
}

/** Full Seven Card Stud game state returned from the API. */
export interface SevenCardStudResponse extends BaseGameResponse {
  players: SevenCardStudPlayerData[];
  communityCard: Card | null;
  pot: number;
  sidePots: SevenCardStudSidePot[];
  dealerIdx: number;
  currentTurn: number;
  phase: number;
  gameEndFlag: boolean;
  lastBet: number;
  minRaise: number;
  bettingLimit: number;
  raiseCount: number;
  maxBetAmount: number;
  roundResults: SevenCardStudResult[];
  cpuActions: SevenCardStudCpuAction[];
  handCount: number;
  ante: number;
  bringIn: number;
  smallBet: number;
  bigBet: number;
  tournamentMode: boolean;
  anteLevelHands: number;
  anteMultiplier: number;
  tableSize: number;
  bringInPlayerIdx: number;
  rebuyAvailable: boolean;
  addonAvailable: boolean;
  rebuyCounts: number[];
  addonUsed: boolean[];
  rebuyEnabled: boolean;
  addonEnabled: boolean;
  rebuyMaxCount: number;
  rebuyChips: number;
  addonChips: number;
  rebuyPeriodHands: number;
  addonAfterHand: number;
  rebuyPhaseType: number;
  muckAvailable: boolean;
  metaAI?: BettingMetaAI;
  profile?: BettingHumanProfileData;
}

// --- Five Card Stud ---

/** Player data in Five Card Stud. */
export interface FiveCardStudPlayerData {
  id: number;
  isHuman: boolean;
  holeCards: Card[];
  doorCards: Card[];
  chips: number;
  currentBet: number;
  folded: boolean;
  allIn: boolean;
  handRank: number;
  handName: string;
  bestHand: Card[];
  playStyleName: string;
  totalHands: number;
  vpip: number;
  pfr: number;
  threeBet: number;
  af: string;
}

/** CPU betting action in Five Card Stud. */
export interface FiveCardStudCpuAction {
  playerIdx: number;
  action: number;
  amount: number;
}

/** Five Card Stud round result for a single player. */
export interface FiveCardStudResult {
  playerIdx: number;
  handRank: number;
  handName: string;
  kickers: string;
  bestHand: Card[];
  wonAmount: number;
  mucked: boolean;
}

/** Side pot in Five Card Stud with eligible players. */
export interface FiveCardStudSidePot {
  amount: number;
  eligiblePlayers: number[];
}

/** Full Five Card Stud game state returned from the API. */
export interface FiveCardStudResponse extends BaseGameResponse {
  players: FiveCardStudPlayerData[];
  communityCard: Card | null;
  pot: number;
  sidePots: FiveCardStudSidePot[];
  dealerIdx: number;
  currentTurn: number;
  phase: number;
  gameEndFlag: boolean;
  lastBet: number;
  minRaise: number;
  bettingLimit: number;
  raiseCount: number;
  maxBetAmount: number;
  roundResults: FiveCardStudResult[];
  cpuActions: FiveCardStudCpuAction[];
  handCount: number;
  ante: number;
  bringIn: number;
  smallBet: number;
  bigBet: number;
  tournamentMode: boolean;
  anteLevelHands: number;
  anteMultiplier: number;
  tableSize: number;
  bringInPlayerIdx: number;
  rebuyAvailable: boolean;
  addonAvailable: boolean;
  rebuyCounts: number[];
  addonUsed: boolean[];
  rebuyEnabled: boolean;
  addonEnabled: boolean;
  rebuyMaxCount: number;
  rebuyChips: number;
  addonChips: number;
  rebuyPeriodHands: number;
  addonAfterHand: number;
  rebuyPhaseType: number;
  muckAvailable: boolean;
  metaAI?: BettingMetaAI;
  profile?: BettingHumanProfileData;
}

// --- Clock Solitaire (クロックソリティア) ---

/** A card in a Clock Solitaire pile with face-up status. */
export interface ClockSolitaireCard {
  card: Card | null;
  faceUp: boolean;
}

/** Full Clock Solitaire game state returned from the API. */
export interface ClockSolitaireResponse extends BaseGameResponse {
  piles: ClockSolitaireCard[][];
  faceUpCount: number[];
  phase: number;
  stepCount: number;
  currentCard?: Card;
}

/** Durak player data. */
export interface DurakPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  cardCount: number;
  cards: Card[];
}

/** Durak table pair (attack + optional defense card). */
export interface DurakTablePair {
  attack: Card;
  defense: Card | null;
}

/** Durak CPU/human action record. */
export interface DurakAction {
  playerIdx: number;
  actionType: number; // 0=attack, 1=defend, 2=pass, 3=take
  card: Card | null;
  attackIdx: number;
}

/** Durak game rule configuration. */
export interface DurakConfig {
  playerCount: number;
  cpuDifficulty: number;
  transferEnabled: boolean;
}

/** Input type alias for Durak configuration. */
export type DurakConfigInput = DurakConfig;

/** Full Durak game state returned from the API. */
export interface DurakResponse extends BaseGameResponse {
  players: DurakPlayerData[];
  currentTurn: number;
  phase: number;
  attackerIdx: number;
  defenderIdx: number;
  tablePairs: DurakTablePair[];
  trumpSuit: string;
  trumpCard: Card | null;
  stockCount: number;
  loserIdx: number;
  gameEndFlag: boolean;
  config: DurakConfig;
  cpuActions: DurakAction[];
  humanAction: DurakAction | null;
  boutNumber: number;
  sortMode: number;
}

// --- Forty Thieves (フォーティシーブス) ---

/** A single tableau card in Forty Thieves with face-up/face-down state. */
export interface FortyThievesTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Forty Thieves. */
export interface FortyThievesHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Forty Thieves game state returned from the API. */
export interface FortyThievesResponse extends BaseGameResponse {
  tableau: FortyThievesTableauCard[][];
  stockCount: number;
  waste: Card[];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: FortyThievesHint;
}

/** Source or target zone for a Forty Thieves card move. */
export interface FortyThievesMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

// --- Forty and Eight (フォーティ・アンド・エイト) ---

/** A single tableau card in Forty and Eight with face-up/face-down state. */
export interface FortyAndEightTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Forty and Eight. */
export interface FortyAndEightHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Forty and Eight game state returned from the API. */
export interface FortyAndEightResponse extends BaseGameResponse {
  tableau: FortyAndEightTableauCard[][];
  stockCount: number;
  waste: Card[];
  foundation: Card[][];
  redealUsed: boolean;
  canRedeal: boolean;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: FortyAndEightHint;
}

/** Source or target zone for a Forty and Eight card move. */
export interface FortyAndEightMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

// --- Sultan of Turkey (スルタン) ---

/** A suggested move hint in Sultan of Turkey. */
export interface SultanHint {
  fromZone: string;
  fromIdx: number;
  toFoundation: number;
}

/** Full Sultan of Turkey game state returned from the API. */
export interface SultanResponse extends BaseGameResponse {
  foundation: Card[][];
  divan: (Card | null)[];
  stockCount: number;
  waste: Card[];
  redealCount: number;
  canRedeal: boolean;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: SultanHint;
}

/** Source zone for a Sultan of Turkey card move (divan slot or waste). */
export interface SultanMoveZone {
  zone: string;
  divanIdx?: number;
}

// --- Crescent (クレセント・ソリティア) ---

/** A single tableau card in Crescent (always face-up). */
export interface CrescentTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Crescent. */
export interface CrescentHint {
  fromCol: number;
  toZone: string;
  toCol: number;
  redeal: boolean;
}

/** Full Crescent game state returned from the API. */
export interface CrescentResponse extends BaseGameResponse {
  tableau: CrescentTableauCard[][];
  foundation: Card[][];
  redealsRemaining: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: CrescentHint;
}

/** Source or target zone for a Crescent card move. */
export interface CrescentMoveZone {
  zone: 'tableau' | 'foundation';
  col?: number;
}

// --- Baker's Dozen (ベーカーズ・ダズン) ---

/** A single tableau card in Baker's Dozen. */
export interface BakersDozenTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Baker's Dozen. */
export interface BakersDozenHint {
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Baker's Dozen game state returned from the API. */
export interface BakersDozenResponse extends BaseGameResponse {
  tableau: BakersDozenTableauCard[][];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: BakersDozenHint;
}

/** Source or target zone for a Baker's Dozen card move. */
export interface BakersDozenMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** A suggested move hint in Calculation. */
export interface CalculationHint {
  fromZone: string;
  wasteIdx: number;
  foundationIdx: number;
}

/** Full Calculation game state returned from the API. */
export interface CalculationResponse extends BaseGameResponse {
  foundations: Card[][];
  wastes: Card[][];
  stockCount: number;
  stockTop?: Card;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: CalculationHint;
}

/** Source or target zone for a Calculation card move. */
export interface CalculationMoveZone {
  zone: 'stock' | 'waste' | 'foundation';
  idx?: number;
}

/** Single Fifty-one player state from the API. */
export interface FiftyOnePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  score: number;
}

/** Fifty-one game configuration. */
export interface FiftyOneConfig {
  cpuDifficulty: number;
}

/** Full Fifty-one game state returned from the API. */
export interface FiftyOneResponse extends BaseGameResponse {
  players: FiftyOnePlayerData[];
  tableCards: Card[];
  phase: number;
  currentTurn: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  turnNumber: number;
  stopCallerIdx: number;
  lastAction: string;
  lastHandIdx: number;
  lastTableIdx: number;
  config: FiftyOneConfig;
}

// --- Yukon (ユーコン) ---

/** A suggested move hint in Yukon. */
export interface YukonHint {
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** API response shape for a Yukon game. */
export interface YukonResponse extends BaseGameResponse {
  tableau: KlondikeTableauCard[][];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: YukonHint;
}

// --- Russian Solitaire (ロシアンソリティア) ---

/** A suggested move hint in Russian Solitaire. */
export interface RussianSolitaireHint {
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** API response shape for a Russian Solitaire game. */
export interface RussianSolitaireResponse extends BaseGameResponse {
  tableau: KlondikeTableauCard[][];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: RussianSolitaireHint;
}

// --- Cruel (クルーエル) ---

/** A suggested move hint in Cruel. */
export interface CruelHint {
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** API response shape for a Cruel game. */
export interface CruelResponse extends BaseGameResponse {
  tableau: KlondikeTableauCard[][];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: CruelHint;
}

// --- Scorpion (スコーピオン) ---

/** A suggested move hint in Scorpion. */
export interface ScorpionHint {
  fromCol: number;
  cardIndex: number;
  toCol: number;
}

/** API response shape for a Scorpion game. */
export interface ScorpionResponse extends BaseGameResponse {
  tableau: KlondikeTableauCard[][];
  stockCount: number;
  completedSuits: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: ScorpionHint;
}

// --- Wasp (ワスプ) ---

/** A suggested move hint in Wasp. */
export interface WaspHint {
  fromCol: number;
  cardIndex: number;
  toCol: number;
}

/** API response shape for a Wasp game. */
export interface WaspResponse extends BaseGameResponse {
  tableau: KlondikeTableauCard[][];
  stockCount: number;
  completedSuits: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: WaspHint;
}

// --- Easthaven (イーストヘイブン) ---

/** A suggested move hint in Easthaven. */
export interface EasthavenHint {
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** API response shape for an Easthaven game. */
export interface EasthavenResponse extends BaseGameResponse {
  tableau: KlondikeTableauCard[][];
  foundation: Card[][];
  stockCount: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: EasthavenHint;
}

// --- Accordion (アコーディオン) ---

/** A single pile in Accordion. Only the top card is revealed; size tracks stacked depth. */
export interface AccordionPile {
  cards: Card[];
  size: number;
}

/** A suggested move hint in Accordion. */
export interface AccordionHint {
  fromIdx: number;
  toIdx: number;
}

/** API response shape for an Accordion game. */
export interface AccordionResponse extends BaseGameResponse {
  piles: AccordionPile[];
  pileCount: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: AccordionHint;
}

// --- Trash (トラッシュ) ---

/** A single slot (position 1..10) for one Trash player. */
export interface TrashSlot {
  /** Face-down slots omit this field. Only face-up cards expose their identity. */
  card?: Card;
  faceUp: boolean;
}

/** One player's full Trash state: 10 slots plus a flag for the CPU. */
export interface TrashPlayerState {
  slots: TrashSlot[];
  isCpu: boolean;
}

/** API response shape for a Trash game. */
export interface TrashResponse extends BaseGameResponse {
  phase: number;
  current: number;
  players: [TrashPlayerState, TrashPlayerState];
  stockSize: number;
  discardSize: number;
  discardTop?: Card;
  pending?: Card;
  moveCount: number;
  winner: number;
}

// --- Whist (ホイスト) ---

/** Whist player data with team, scores, and trick count. */
export interface WhistPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
  team: number;
}

/** A card played in a Whist trick. */
export interface WhistTrickCard {
  playerIdx: number;
  card: Card;
}

/** Whist game configuration. */
export interface WhistConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A suggested hint for Whist. */
export interface WhistHint {
  cardIndex?: number;
  reason: string;
}

/** Full Whist game state returned from the API. */
export interface WhistResponse extends BaseGameResponse {
  players: WhistPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  currentTrick: WhistTrickCard[];
  trumpSuit: number;
  dealerIdx: number;
  teamScores: [number, number];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  config: WhistConfig;
  hint?: WhistHint;
}

// --- Catch the Ten (スコッチ・ホイスト) ---

/** Catch the Ten player data with team, scores, and trick count. */
export interface CatchTenPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
  team: number;
}

/** A card played in a Catch the Ten trick. */
export interface CatchTenTrickCard {
  playerIdx: number;
  card: Card;
}

/** Catch the Ten game configuration. */
export interface CatchTenConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A suggested hint for Catch the Ten. */
export interface CatchTenHint {
  cardIndex?: number;
  reason: string;
}

/** Full Catch the Ten game state returned from the API. */
export interface CatchTenResponse extends BaseGameResponse {
  players: CatchTenPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  currentTrick: CatchTenTrickCard[];
  trumpSuit: number;
  dealerIdx: number;
  teamScores: [number, number];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  config: CatchTenConfig;
  hint?: CatchTenHint;
}

// --- Briscola (ブリスコラ) ---

/** Briscola player data with points and trick count. */
export interface BriscolaPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  points: number;
  trickCount: number;
}

/** A card played in a Briscola trick. */
export interface BriscolaTrickCard {
  playerIdx: number;
  card: Card;
}

/** Briscola game configuration. */
export interface BriscolaConfig {
  cpuDifficulty: number;
}

/** A suggested hint for Briscola. */
export interface BriscolaHint {
  cardIndex?: number;
  reason: string;
}

/** Full Briscola game state returned from the API. */
export interface BriscolaResponse extends BaseGameResponse {
  players: BriscolaPlayerData[];
  phase: number;
  trickNumber: number;
  currentPlayerIdx: number;
  currentTrick: BriscolaTrickCard[];
  trumpSuit: number;
  /** Face-up trump card (omitted once the stock is exhausted). */
  trumpCard?: Card;
  dealerIdx: number;
  leadPlayerIdx: number;
  /**
   * Cards remaining in the stock; this does NOT include the face-up trump
   * card (which is tracked separately via `trumpCard` until drawn last).
   */
  stockRemaining: number;
  gameEndFlag: boolean;
  /** -1 = tie or unfinished. */
  winnerIdx: number;
  config: BriscolaConfig;
  hint?: BriscolaHint;
}

// --- Schnapsen / Sixty-Six (シュナプセン / 66) ---

/** Schnapsen player data with points and trick count. */
export interface SchnapsenPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  points: number;
  trickCount: number;
}

/** A card played in a Schnapsen trick. */
export interface SchnapsenTrickCard {
  playerIdx: number;
  card: Card;
}

/** Schnapsen game configuration. */
export interface SchnapsenConfig {
  cpuDifficulty: number;
}

/** A suggested hint for Schnapsen. */
export interface SchnapsenHint {
  cardIndex?: number;
  reason: string;
  isMarriage: boolean;
}

/** Full Schnapsen game state returned from the API. */
export interface SchnapsenResponse extends BaseGameResponse {
  players: SchnapsenPlayerData[];
  phase: number;
  trickNumber: number;
  currentPlayerIdx: number;
  currentTrick: SchnapsenTrickCard[];
  trumpSuit: number;
  /** Face-up trump upcard (omitted once the stock is exhausted). */
  trumpCard?: Card;
  dealerIdx: number;
  leadPlayerIdx: number;
  /** Cards remaining in the stock (excludes the face-up trump upcard). */
  stockRemaining: number;
  /** True once the stock is exhausted and must-follow rules apply (phase 2). */
  isEndgame: boolean;
  /** Indices in the human hand that are legal to play right now. */
  validPlays: number[];
  /** Indices in the human hand that can start a marriage declaration. */
  marriagePlays: number[];
  gameEndFlag: boolean;
  /** -1 = tie or unfinished. */
  winnerIdx: number;
  config: SchnapsenConfig;
  hint?: SchnapsenHint;
}

// --- Truco (トゥルコ) ---

/** Truco player data with per-hand baza (trick) count. */
export interface TrucoPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
}

/** A card played in a Truco baza (trick). */
export interface TrucoTrickCard {
  playerIdx: number;
  card: Card;
}

/** Truco game configuration. */
export interface TrucoConfig {
  cpuDifficulty: number;
  matchTarget: number;
}

/** A suggested hint for Truco (action is play / call / accept / decline). */
export interface TrucoHint {
  action: string;
  cardIndex?: number;
  reason: string;
}

/** Full Truco game state returned from the API. */
export interface TrucoResponse extends BaseGameResponse {
  players: TrucoPlayerData[];
  phase: number;
  handNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  /** Player who must respond to a pending Truco call; -1 when not awaiting a response. */
  responderIdx: number;
  currentTrick: TrucoTrickCard[];
  /** Outcome of each completed baza this hand: 0/1 = winner, -1 = parda (tie). */
  trickResults: number[];
  leadPlayerIdx: number;
  manoIdx: number;
  dealerIdx: number;
  /** Current agreed stake for the hand (1..4). */
  handStake: number;
  /** Accepted betting level (0=none, 1=Truco, 2=Retruco, 3=Vale Cuatro). */
  acceptedLevel: number;
  /** Proposed level while awaiting a response (0 otherwise). */
  pendingLevel: number;
  /** Index of the player whose Truco call is pending (-1 otherwise). */
  trucoCallerIdx: number;
  /** Whether the human may declare / raise Truco right now. */
  canDeclareTruco: boolean;
  /** Points needed to win the match. */
  matchTarget: number;
  /** Cumulative match points [p0, p1]. */
  matchPoints: number[];
  /** Winner of the most recent hand (-1 = unresolved). */
  handWinnerIdx: number;
  gameEndFlag: boolean;
  /** -1 = unfinished. */
  winnerIdx: number;
  config: TrucoConfig;
  hint?: TrucoHint;
}

// --- Poker Squares (ポーカー・スクエア) ---

/** Single cell of the 5x5 Poker Squares board. */
export interface PokerSquaresBoardCell {
  /** Placed card, or `null` when the cell is empty. */
  card: Card | null;
}

/** Poker Squares API response. */
export interface PokerSquaresResponse extends BaseGameResponse {
  /** 5x5 board. Empty cells have `card === null`. */
  board: PokerSquaresBoardCell[][];
  /** Next card to place, or `null` once all 25 cards have been placed. */
  currentCard: Card | null;
  /** Number of cards placed so far (0..25). */
  placedCount: number;
  /** 0 = playing, 1 = complete. */
  phase: number;
  /** Whether the last action can be undone. */
  canUndo: boolean;
  /** Score per row (length 5). */
  rowScores: number[];
  /** Score per column (length 5). */
  colScores: number[];
  /** Sum of all row and column scores. */
  totalScore: number;
}

// --- Monte Carlo Solitaire (モンテカルロ・ソリティア) ---

/** Single cell of the 5x5 Monte Carlo board. */
export interface MonteCarloBoardCell {
  /** Card on this cell, or `null` when empty (gap awaiting compression). */
  card: Card | null;
}

/** Suggested hint for Monte Carlo Solitaire. */
export interface MonteCarloHint {
  /** "remove" suggests a pair to take off; "deal" suggests pressing the Deal button. */
  action: 'remove' | 'deal';
  /** First cell of the pair (for `action === 'remove'`). */
  fromR?: number;
  fromC?: number;
  /** Second cell of the pair (for `action === 'remove'`). */
  toR?: number;
  toC?: number;
}

/** Monte Carlo Solitaire API response. */
export interface MonteCarloResponse extends BaseGameResponse {
  /** 5x5 board. Empty cells (post-removal, pre-deal) have `card === null`. */
  board: MonteCarloBoardCell[][];
  /** 0 = playing, 1 = game clear, 2 = game over. */
  phase: number;
  /** Cards remaining in the stock (52 - drawn so far). */
  stockCount: number;
  /** Cards removed from the board so far (must hit 52 to win). */
  removedCount: number;
  /** Number of times the player has pressed Deal. */
  dealCount: number;
  /** Whether the last action can be undone. */
  canUndo: boolean;
  /** True when no remove pairs exist and the stock cannot help. */
  isStalemate: boolean;
  /** Server-generated hint, present only on `/montecarlo/exec` with `command: "hint"`. */
  hint?: MonteCarloHint;
}

// --- Let It Ride (レット・イット・ライド) ---

/** Let It Ride API response. */
export interface LetItRideResponse extends BaseGameResponse {
  playerHand: Card[];
  /** Community cards: masked as `MaskedCard` until revealed by phase progression. */
  communityCards: (Card | MaskedCard)[];
  phase: number;
  chips: number;
  betAmount: number;
  bet1Active: boolean;
  bet2Active: boolean;
  bet3Active: boolean;
  result: number;
  handRank: number;
  bet1Payout: number;
  bet2Payout: number;
  bet3Payout: number;
  totalPayout: number;
}

// --- Red Dog (レッドドッグ) ---

/** Red Dog API response. */
export interface RedDogResponse extends BaseGameResponse {
  /** Initial 2 cards. */
  initialCards: Card[];
  /** Third card revealed at end (or after raise/stay). */
  thirdCard?: Card;
  phase: number;
  chips: number;
  ante: number;
  raise: number;
  /** Spread = |rank2 - rank1| - 1, 0 when consecutive or pair. */
  spread: number;
  result: number;
  totalPayout: number;
}

// --- Casino War (カジノウォー) ---

/** Casino War API response. */
export interface CasinoWarResponse extends BaseGameResponse {
  /** Player's initial card. */
  playerCard?: Card;
  /** Dealer's initial card. */
  dealerCard?: Card;
  /** Player's war card (only set after going to war). */
  playerWarCard?: Card;
  /** Dealer's war card (only set after going to war). */
  dealerWarCard?: Card;
  /** Burn cards face-down between initial and war (length 0 or 3). */
  burnCards: Card[];
  phase: number;
  chips: number;
  ante: number;
  /** Additional bet placed when going to war (equal to ante). */
  warBet: number;
  result: number;
  totalPayout: number;
}

// --- Oicho-Kabu (おいちょかぶ) ---

/**
 * Oicho-Kabu API response.
 *
 * A kabufuda (40-card, values 1–10) baccarat-style banking game: one human
 * "child" vs a CPU "banker". The banker's hand stays hidden (empty array,
 * rank 0) until the round ends. Rank is the sum of card values mod 10 (9 best,
 * 0 worst); ties push and a win pays 1:1.
 */
export interface OichoKabuResponse extends BaseGameResponse {
  /** The child's (player's) cards. */
  playerHand: Card[];
  /** The banker's cards — empty until the round ends. */
  bankerHand: Card[];
  /** The child's rank (sum mod 10), 0–9. */
  playerRank: number;
  /** The banker's rank (sum mod 10) — 0 until the round ends. */
  bankerRank: number;
  phase: number;
  chips: number;
  /** Chips wagered this round. */
  bet: number;
  result: number;
  totalPayout: number;
}

/**
 * Trente et Quarante (Rouge et Noir) game state response.
 *
 * A pure banking game with no player card decisions: the player picks one of
 * four even-money bets (Noir, Rouge, Couleur, Inverse) plus a stake, then the
 * dealer immediately deals two rows — Noir (black) first, then Rouge (red) —
 * each summed until the total reaches 31 or more. The row with the LOWER total
 * wins. Betting resolves the round in one step, so there are only two phases:
 * Bet (0) and Result (1).
 */
export interface TrenteEtQuaranteResponse extends BaseGameResponse {
  /** Game phase: 0=Bet, 1=Result. */
  phase: number;
  /** Number of rounds resolved so far. */
  roundNumber: number;
  /** Player's remaining chip stack. */
  chips: number;
  /** Selected bet: 0=Noir, 1=Rouge, 2=Couleur, 3=Inverse. */
  currentBet: number;
  /** Amount wagered on the current round. */
  stake: number;
  /** Cards dealt to the Noir (black) row, summed until the total reaches 31+. */
  noirRow: Card[];
  /** Cards dealt to the Rouge (red) row, summed until the total reaches 31+. */
  rougeRow: Card[];
  /** Pip total of the Noir row (31–40 once the row is complete). */
  noirTotal: number;
  /** Pip total of the Rouge row (31–40 once the row is complete). */
  rougeTotal: number;
  /** Winning row (the LOWER total): 0=Noir, 1=Rouge; -1 for none (push/refait). */
  winningRow: number;
  /** Whether the first card dealt (Noir row's first card) is red — drives Couleur/Inverse display. */
  firstCardRed: boolean;
  /** True when both rows tie at 31 (a "Refait" — half the stake goes to the house). */
  refait: boolean;
  /** Round result from the player's perspective: 1=win, 0=push, -1=lose. */
  result: number;
  /** Gross chips returned to the stack this round (0=lose, stake/2=refait, stake=push, stake*2=win). */
  payout: number;
  /** Number of cards remaining in the shoe. */
  remainingDeck: number;
  /** True once the round has resolved. */
  gameEndFlag: boolean;
  /** Educational hint offered during the bet phase. */
  hint?: {
    /** Suggested bet type: 0=Noir, 1=Rouge, 2=Couleur, 3=Inverse. */
    bet: number;
    /** i18n reason key for the suggestion. */
    reason: string;
  };
  /** Local-rule configuration. */
  config: {
    /** Bet type pre-selected at the start of each round. */
    defaultBet: number;
  };
}

/** Dragon Tiger game state response. Bet types: 0=Dragon, 1=Tiger, 2=Tie. */
export interface DragonTigerResponse extends BaseGameResponse {
  /** Card dealt to the Dragon slot. */
  dragonCard?: Card;
  /** Card dealt to the Tiger slot. */
  tigerCard?: Card;
  phase: number;
  chips: number;
  betAmount: number;
  /** 0=Dragon, 1=Tiger, 2=Tie */
  betType: number;
  /** Domain GameResult on the wire: 1=Dragon wins, -1=Tiger wins, 0=Tie */
  result: number;
  payout: number;
  /** Big Road history. 0=Dragon, 1=Tiger, 2=Tie. */
  history: number[];
}

/** A single hand within a Blackjack Switch round. */
export interface BlackJackSwitchHand {
  /** Cards in this hand. Null entries represent face-down cards (e.g. dealer hole). */
  cards: (Card | null)[];
  score: number;
  bet: number;
  stood: boolean;
  doubled: boolean;
  busted: boolean;
  /** True when the hand is a 2-card 21 (natural). Pays 1:1 in Blackjack Switch. */
  isBJ: boolean;
  /** Domain GameResult: 1=Win, 0=Draw, -1=Lose. */
  result: number;
  /** Per-hand payout (bet returned + winnings). */
  payout: number;
}

/** Blackjack Switch game state response. */
export interface BlackJackSwitchResponse extends BaseGameResponse {
  /** Two player hands; the player may switch the second card between them. */
  hands: BlackJackSwitchHand[];
  /** Dealer's cards. The hole card is null until the round ends. */
  dealerCards: (Card | null)[];
  /** Visible dealer score (up-card only mid-round; full score at end). */
  dealerScore: number;
  /** 1=BET, 2=SWITCH, 3=ACTION, 4=END. */
  phase: number;
  /** Index of the player hand currently acting (during ACTION phase). */
  currentHandIdx: number;
  chips: number;
  /** True when the player exercised the switch this round. */
  switched: boolean;
  /** True when the dealer ended on 22 (push rule, not natural BJ). */
  dealerPushed22: boolean;
  /** Aggregate of hand results: 1=net win, 0=draw, -1=net loss. */
  overallResult: number;
  /** Sum of per-hand payouts. */
  totalPayout: number;
}

/** President player data. */
export interface PresidentPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  rank: number;
  cardCount: number;
  cards: Card[];
}

/** A play or pass action in President. */
export interface PresidentAction {
  playerIdx: number;
  playedCards: Card[] | null; // null = pass
}

/** Card exchange action in President. */
export interface PresidentExchangeAction {
  fromPlayerIdx: number;
  toPlayerIdx: number;
  cards: Card[];
}

/** President game rule configuration. */
export interface PresidentConfig {
  revolutionEnabled: boolean;
  cardExchangeEnabled: boolean;
  passFieldFlushEnabled: boolean;
  cpuDifficulty: number;
}

/** Full President game state returned from the API. */
export interface PresidentResponse extends BaseGameResponse {
  players: PresidentPlayerData[];
  currentTurn: number;
  tableCards: Card[];
  lastPlayPlayerIdx: number;
  gameEndFlag: boolean;
  revolutionActive: boolean;
  config: PresidentConfig;
  exchangeActions: PresidentExchangeAction[];
  cpuActions: PresidentAction[];
  humanAction: PresidentAction | null;
}

/** Cassino player data. */
export interface CassinoPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  capturedCount: number;
  sweepCount: number;
  totalScore: number;
}

/** A build on the Cassino table. */
export interface CassinoBuild {
  ownerIdx: number;
  value: number;
  groups: Card[][];
  isMulti: boolean;
}

/** A take, build, or trail action in Cassino. */
export interface CassinoAction {
  playerIdx: number;
  type: 'take' | 'build' | 'trail';
  playedCard: Card | null;
  capturedCards: Card[];
  buildValue: number;
  isSweep: boolean;
}

/** Cassino score detail (per round). */
export interface CassinoScoreDetail {
  cards: Record<number, number>;
  spades: Record<number, number>;
  aces: Record<number, number>;
  hasBigCasino: number;
  hasLittleCasino: number;
  sweeps: Record<number, number>;
  gained: Record<number, number>;
}

/** Cassino game rule configuration. */
export interface CassinoConfig {
  targetScore: number;
  multiBuildEnabled: boolean;
  sweepBonusEnabled: boolean;
  cpuDifficulty: number;
}

/** Full Cassino game state returned from the API. */
export interface CassinoResponse extends BaseGameResponse {
  players: CassinoPlayerData[];
  currentTurn: number;
  tableCards: Card[];
  builds: CassinoBuild[];
  lastCaptureIdx: number;
  gameEndFlag: boolean;
  phase: string;
  config: CassinoConfig;
  cpuActions: CassinoAction[];
  humanAction: CassinoAction | null;
  remainingDeck: number;
  packsDealt: number;
  roundWinners: number[];
  lastRoundDetail: CassinoScoreDetail | null;
}

/** Scopa player data. */
export interface ScopaPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  capturedCount: number;
  scopaCount: number;
  totalScore: number;
}

/** A play/lay action in Scopa. */
export interface ScopaAction {
  playerIdx: number;
  playedCard: Card | null;
  capturedCards: Card[];
  isScopa: boolean;
}

/** Scopa score detail (per round). */
export interface ScopaScoreDetail {
  cards: Record<number, number>;
  diamonds: Record<number, number>;
  sevens: Record<number, number>;
  hasSetteBello: number;
  scopas: Record<number, number>;
  gained: Record<number, number>;
}

/** Scopa game rule configuration. */
export interface ScopaConfig {
  targetScore: number;
  cpuDifficulty: number;
}

/** Full Scopa game state returned from the API. */
export interface ScopaResponse extends BaseGameResponse {
  players: ScopaPlayerData[];
  currentTurn: number;
  tableCards: Card[];
  lastCaptureIdx: number;
  gameEndFlag: boolean;
  phase: string;
  config: ScopaConfig;
  cpuActions: ScopaAction[];
  humanAction: ScopaAction | null;
  remainingDeck: number;
  packsDealt: number;
  roundWinners: number[];
  lastRoundDetail: ScopaScoreDetail | null;
}

// --- Scopone (スコポーネ) ---

/** Scopone player data (4 players in 2 teams). */
export interface ScoponePlayerData {
  id: number;
  isHuman: boolean;
  team: number;
  handCount: number;
  cards: Card[];
  capturedCount: number;
  scopaCount: number;
}

/** Scopone per-round score breakdown (per team, 2-element tuples). */
export interface ScoponeScoreDetail {
  cards: [number, number];
  diamonds: [number, number];
  sevens: [number, number];
  scopas: [number, number];
  gained: [number, number];
  settebello: number;
}

/** Scopone game rule configuration. */
export interface ScoponeConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** Full Scopone game state returned from the API. */
export interface ScoponeResponse extends BaseGameResponse {
  players: ScoponePlayerData[];
  tableCards: Card[];
  phase: string;
  roundNumber: number;
  currentTurn: number;
  dealerIdx: number;
  teamScores: number[];
  lastCaptureIdx: number;
  winnerTeam: number;
  gameEndFlag: boolean;
  isHumanTurn: boolean;
  /** Per human hand-card index, the list of valid table-index capture sets. */
  handCaptures: number[][][];
  lastRoundDetail?: ScoponeScoreDetail | null;
  config: ScoponeConfig;
}

// --- Escoba (エスコバ) ---

/** Escoba player data (4 players, free-for-all / no teams). */
export interface EscobaPlayerData {
  id: number;
  isHuman: boolean;
  handCount: number;
  cards: Card[];
  capturedCount: number;
  /** The captured pile's actual cards. Populated only for the human player; CPUs stay count-only (empty array). */
  capturedCards: Card[];
  escobaCount: number;
  score: number;
}

/**
 * Escoba per-round score breakdown (per-player arrays, one entry per player).
 * `aceEspada` / `seteEspada` are the player indices who took the Ace♠ and 7♠.
 */
export interface EscobaScoreDetail {
  cards: number[];
  espadas: number[];
  sevens: number[];
  oros: number[];
  escobas: number[];
  gained: number[];
  aceEspada: number;
  seteEspada: number;
}

/** Escoba game rule configuration. */
export interface EscobaConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** Full Escoba game state returned from the API. */
export interface EscobaResponse extends BaseGameResponse {
  players: EscobaPlayerData[];
  tableCards: Card[];
  phase: string;
  roundNumber: number;
  currentTurn: number;
  dealerIdx: number;
  stockRemaining: number;
  lastCaptureIdx: number;
  winnerIdx: number;
  gameEndFlag: boolean;
  isHumanTurn: boolean;
  /** Per human hand-card index, the list of valid table-index capture sets summing to 15. */
  handCaptures: number[][][];
  lastRoundDetail?: EscobaScoreDetail | null;
  config: EscobaConfig;
}

// --- Barbu (バルブ) ---

/** Per-game configuration for Barbu. */
export interface BarbuConfig {
  cpuDifficulty: number;
}

/** A single Barbu player's view. Cards are populated only for the human. */
export interface BarbuPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  dominoRank: number;
  totalScore: number;
}

/** A single card played into the current/last trick. */
export interface BarbuTrickCard {
  playerIdx: number;
  card: Card;
}

/** One deal's scoring breakdown. */
export interface BarbuDealDetail {
  contract: number;
  trumpSuit: number;
  dealerIdx: number;
  gained: Record<number, number>;
}

/** API response shape for a Barbu game. */
export interface BarbuResponse extends BaseGameResponse {
  players: BarbuPlayerData[];
  phase: string;
  dealNumber: number;
  totalDeals: number;
  dealerIdx: number;
  currentTurn: number;
  currentContract: number;
  trumpSuit: number;
  trickNumber: number;
  currentTrick: BarbuTrickCard[];
  lastTrick: BarbuTrickCard[];
  lastTrickWinner: number;
  tablePlaced: number[];
  dominoPlayable: number[];
  usedContracts: boolean[];
  gameEndFlag: boolean;
  config: BarbuConfig;
  roundWinners: number[];
  lastDealDetail: BarbuDealDetail | null;
  dealHistory: BarbuDealDetail[];
}

// --- Spite and Malice (スパイト・アンド・マリス) ---

/** A single Spite & Malice player's view. Hand may contain `null` when
 * the opponent's cards are hidden from view. */
export interface SpiteAndMalicePlayerState {
  hand: (Card | null)[];
  goalTop?: Card;
  goalSize: number;
  sides: [Card[], Card[], Card[], Card[]];
  isCpu: boolean;
}

/** Hint information for the next recommended Spite & Malice move. */
export interface SpiteAndMaliceHint {
  source: 'goal' | 'hand' | 'side';
  index: number;
  foundationIdx: number;
  discard: boolean;
}

/** API response shape for a Spite & Malice game. */
export interface SpiteAndMaliceResponse extends BaseGameResponse {
  phase: number;
  current: number;
  players: [SpiteAndMalicePlayerState, SpiteAndMalicePlayerState];
  foundations: Card[][];
  foundationTops: number[];
  stockSize: number;
  completedSize: number;
  moveCount: number;
  winner: number;
  goalSize: number;
  cpuDifficulty: number;
  /** True when the human can auto-complete at least one foundation move on their turn. */
  canAutoComplete: boolean;
  hint?: SpiteAndMaliceHint;
}

/** Source or target zone for a Spite & Malice move. */
export interface SpiteAndMaliceMoveZone {
  zone: 'hand' | 'goal' | 'side' | 'foundation';
  idx?: number;
}

// --- Nertz / Pounce (ナーツ / パウンス) ---

/** Tableau card with face-up state in a Nertz player area. */
export interface NertzTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** Nertz player view (per-player tableau, nertz pile, waste, and stock). */
export interface NertzPlayerData {
  name: string;
  isHuman: boolean;
  deckIdx: number;
  score: number;
  nertzSize: number;
  nertzTop?: Card;
  tableau: NertzTableauCard[][];
  wasteTop?: Card;
  wasteSize: number;
  stockSize: number;
}

/** Nertz shared foundation pile. */
export interface NertzFoundationData {
  top?: Card;
  suit: number;
  size: number;
}

/** Nertz suggested move hint. */
export interface NertzHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Nertz local rule configuration. */
export interface NertzConfig {
  playerCount: number;
  drawCount: number;
  targetScore: number;
  cpuDifficulty: number;
  cpuTickMoves: number;
}

/** Source or target zone for a Nertz move. */
export interface NertzMoveZone {
  zone: 'nertz' | 'waste' | 'tableau' | 'foundation';
  col?: number;
  idx?: number;
  cardIndex?: number;
}

/** Full Nertz game state returned from the API. */
export interface NertzResponse extends BaseGameResponse {
  phase: number;
  roundNumber: number;
  winnerIdx: number;
  matchWinner: number;
  moveCount: number;
  canUndo: boolean;
  playerCount: number;
  drawCount: number;
  targetScore: number;
  cpuDifficulty: number;
  /** CPU per-tick budget (resolved from cpuDifficulty when 0). */
  cpuTickMoves: number;
  players: NertzPlayerData[];
  foundations: NertzFoundationData[];
  hint?: NertzHint;
}

/** Player snapshot for Slapjack. */
export interface SlapjackPlayerData {
  name: string;
  isHuman: boolean;
  stockSize: number;
}

/** Full Slapjack game state returned from the API. */
export interface SlapjackResponse extends BaseGameResponse {
  phase: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  currentTurnIdx: number;
  isHumanTurn: boolean;
  isTopJack: boolean;
  centerPileSize: number;
  topCard?: Card | null;
  players: SlapjackPlayerData[];
  cpuDifficulty: number;
  pendingKind: number;
  pendingDeadlineMs: number;
  lastEventKind: number;
  lastEventPlayerIdx: number;
}

/** Player snapshot for Egyptian Ratscrew. */
export interface EgyptianRatscrewPlayerData {
  name: string;
  isHuman: boolean;
  stockSize: number;
}

/** Full Egyptian Ratscrew game state returned from the API. */
export interface EgyptianRatscrewResponse extends BaseGameResponse {
  phase: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  currentTurnIdx: number;
  isHumanTurn: boolean;
  isTopFaceCard: boolean;
  isSlappable: boolean;
  centerPileSize: number;
  topCard?: Card | null;
  players: EgyptianRatscrewPlayerData[];
  cpuDifficulty: number;
  chanceRemaining: number;
  chanceFromIdx: number;
  pendingKind: number;
  pendingDeadlineMs: number;
  lastEventKind: number;
  lastEventPlayerIdx: number;
  lastSlapReason: number;
}

// --- Contract Rummy (コントラクトラミー) ---

/** Contract Rummy meld: a set or run of cards laid down on the table. */
export interface ContractRummyMeld {
  cards: Card[];
}

/** Contract Rummy contract slot: a single requirement of the round's contract. */
export interface ContractRummyContractSlot {
  /** 0 = set (same rank), 1 = run (same suit consecutive). */
  kind: number;
  /** Number of cards required to fill this slot. */
  size: number;
}

/** Contract Rummy player state. */
export interface ContractRummyPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  melds: ContractRummyMeld[];
  /** Whether the player has met this round's contract. */
  contractMet: boolean;
  roundScore: number;
  cumulativeScore: number;
}

/** Contract Rummy game configuration. */
export interface ContractRummyConfig {
  cpuDifficulty: number;
  failContractPenalty: number;
}

/** Contract Rummy API response. */
export interface ContractRummyResponse extends BaseGameResponse {
  players: ContractRummyPlayer[];
  /** 0 = draw, 1 = play, 2 = round end, 3 = game end. */
  phase: number;
  roundNumber: number;
  totalRounds: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  roundWinnerIdx: number;
  /** The current round's contract (sequence of slots to satisfy). */
  contractSlots: ContractRummyContractSlot[];
  config: ContractRummyConfig;
}

// --- Carioca (カリオカ) ---

/** Carioca meld: a set (trío) or run (escala) of cards laid down on the table. */
export interface CariocaMeld {
  cards: Card[];
}

/** Carioca contract slot: a single requirement of the round's contract. */
export interface CariocaContractSlot {
  /** 0 = set (same rank), 1 = run (same suit consecutive). */
  kind: number;
  /** Number of cards required to fill this slot. */
  size: number;
}

/** Carioca player state. */
export interface CariocaPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  melds: CariocaMeld[];
  /** Whether the player has met this round's contract. */
  contractMet: boolean;
  roundScore: number;
  cumulativeScore: number;
}

/** Carioca game configuration. */
export interface CariocaConfig {
  playerCount: number;
  cpuDifficulty: number;
  failContractPenalty: number;
}

/** Carioca API response. */
export interface CariocaResponse extends BaseGameResponse {
  players: CariocaPlayer[];
  /** 0 = draw, 1 = play, 2 = round end, 3 = game end. */
  phase: number;
  roundNumber: number;
  totalRounds: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  roundWinnerIdx: number;
  /** The current round's contract (sequence of slots to satisfy). */
  contractSlots: CariocaContractSlot[];
  config: CariocaConfig;
}

// --- Kalooki (カルーキ) ---

/** Kalooki meld: a set or run of cards laid face-up on the table. */
export interface KalookiMeld {
  cards: Card[];
}

/** Kalooki player state with face-up table melds, opening flag, and scores. */
export interface KalookiPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  melds: KalookiMeld[];
  /** Whether the player has made their opening meld(s) meeting the threshold. */
  hasOpened: boolean;
  roundScore: number;
  cumulativeScore: number;
}

/** Kalooki game configuration. */
export interface KalookiConfig {
  cpuDifficulty: number;
  playerCount: number;
  openingThreshold: number;
}

/** Kalooki API response. */
export interface KalookiResponse extends BaseGameResponse {
  players: KalookiPlayer[];
  /** 0 = draw, 1 = meld, 2 = round end, 3 = game end. */
  phase: number;
  /** Minimum points required for a player's opening meld. */
  openingThreshold: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  roundWinnerIdx: number;
  config: KalookiConfig;
}

// --- Oasis Poker (オアシスポーカー) ---

/** Oasis Poker API response. */
export interface OasisPokerResponse extends BaseGameResponse {
  playerHand: Card[];
  /** Dealer hand: during bet/exchange/action phases only the first card is revealed and
   * the remaining slots are `MaskedCard`. After the end phase all 5 are real `Card`s. */
  dealerHand: (Card | MaskedCard)[];
  phase: number;
  chips: number;
  anteBet: number;
  jackpotBet: number;
  /** Number of cards exchanged this round (0..5). */
  exchangeCount: number;
  /** Fee collected for exchanging cards (ante × exchangeCount). */
  exchangeFee: number;
  playBet: number;
  result: number;
  antePayout: number;
  playPayout: number;
  jackpotPayout: number;
  totalPayout: number;
  dealerQualified: boolean;
  playerHandRank: number;
  dealerHandRank: number;
}

// --- Russian Poker (ロシアンポーカー) ---

/** Russian Poker game state from the /russianpoker/exec endpoint. */
export interface RussianPokerResponse extends BaseGameResponse {
  playerHand: Card[];
  dealerHand: (Card | MaskedCard)[];
  phase: number;
  chips: number;
  anteBet: number;
  exchangeCount: number;
  exchangeFee: number;
  bought6th: boolean;
  buy6thFee: number;
  forceExchanged: boolean;
  forceExchangeFee: number;
  playBet: number;
  result: number;
  antePayout: number;
  playPayout: number;
  totalPayout: number;
  dealerQualified: boolean;
  playerHandRank: number;
  dealerHandRank: number;
}

// --- Beleaguered Castle (包囲された城) ---

/** A single tableau card in Beleaguered Castle. */
export interface BeleagueredCastleTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Beleaguered Castle. */
export interface BeleagueredCastleHint {
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Beleaguered Castle game state returned from the API. */
export interface BeleagueredCastleResponse extends BaseGameResponse {
  tableau: BeleagueredCastleTableauCard[][];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: BeleagueredCastleHint;
}

/** Source or target zone for a Beleaguered Castle card move. */
export interface BeleagueredCastleMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

// --- Streets and Alleys ---

/** A single tableau card in Streets and Alleys. */
export interface StreetsAndAlleysTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Streets and Alleys. */
export interface StreetsAndAlleysHint {
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Streets and Alleys game state returned from the API. */
export interface StreetsAndAlleysResponse extends BaseGameResponse {
  tableau: StreetsAndAlleysTableauCard[][];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: StreetsAndAlleysHint;
}

/** Source or target zone for a Streets and Alleys card move. */
export interface StreetsAndAlleysMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

// --- King Albert ---

/** A single tableau card in King Albert. */
export interface KingAlbertTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in King Albert. */
export interface KingAlbertHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full King Albert game state returned from the API. */
export interface KingAlbertResponse extends BaseGameResponse {
  tableau: KingAlbertTableauCard[][];
  reserve: (Card | null)[];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: KingAlbertHint;
}

/** Source or target zone for a King Albert card move. */
export interface KingAlbertMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

// --- Flower Garden ---

/** A single tableau card in Flower Garden. */
export interface FlowerGardenTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Flower Garden. */
export interface FlowerGardenHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Flower Garden game state returned from the API. */
export interface FlowerGardenResponse extends BaseGameResponse {
  tableau: FlowerGardenTableauCard[][];
  reserve: (Card | null)[];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: FlowerGardenHint;
}

/** Source or target zone for a Flower Garden card move. */
export interface FlowerGardenMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

// --- Tarneeb ---

/** Tarneeb player data with team and current bid. */
export interface TarneebPlayerData {
  id: number;
  isHuman: boolean;
  team: number;
  cardCount: number;
  cards: Card[];
  bid: number;
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
}

/** A card played in a Tarneeb trick. */
export interface TarneebTrickCard {
  playerIdx: number;
  card: Card;
}

/** Tarneeb game configuration. */
export interface TarneebConfig {
  cpuDifficulty: number;
  pointLimit: number;
  minBid: number;
}

/** A suggested hint for Tarneeb. */
export interface TarneebHint {
  cardIndex?: number;
  bid?: number;
  trumpSuit?: number;
  reason: string;
}

/** Full Tarneeb game state returned from the API. */
export interface TarneebResponse extends BaseGameResponse {
  players: TarneebPlayerData[];
  teamScores: number[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  bidWinnerIdx: number;
  highestBid: number;
  trumpSuit: number;
  redealCount: number;
  dealerIdx: number;
  currentTrick: TarneebTrickCard[];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  config: TarneebConfig;
  hint?: TarneebHint;
}

// --- High Card Flush (ハイカードフラッシュ) ---

/** High Card Flush API response. */
export interface HighCardFlushResponse extends BaseGameResponse {
  playerHand: Card[];
  dealerHand: Card[];
  phase: number;
  chips: number;
  anteBet: number;
  flushBonusBet: number;
  straightFlushBet: number;
  raiseBet: number;
  result: number;
  antePayout: number;
  raisePayout: number;
  flushBonusPayout: number;
  straightFlushPayout: number;
  totalPayout: number;
  dealerQualified: boolean;
  playerFlushLen: number;
  dealerFlushLen: number;
  playerStraightFlushLen: number;
  maxRaiseMultiplier: number;
}

// --- Gaps / Montana (ギャップス) ---

/** A suggested next-move hint in Gaps. */
export interface GapsHint {
  fromRow: number;
  fromCol: number;
  toRow: number;
  toCol: number;
}

/** Full Gaps game state returned from the API. */
export interface GapsResponse extends BaseGameResponse {
  /** 4-row x 13-col grid. `null` cells are gaps. */
  grid: (Card | null)[][];
  redealsUsed: number;
  redealsRemaining: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: GapsHint;
}

// --- Four Card Poker (フォーカードポーカー) ---

/** Six Card Golf grid slot. */
export interface SixCardGolfSlot {
  card: Card | null;
  faceUp: boolean;
}

/** Six Card Golf player data. */
export interface SixCardGolfPlayerData {
  id: number;
  isHuman: boolean;
  grid: SixCardGolfSlot[];
  roundScore: number;
  cumulativeScore: number;
  allFaceUp: boolean;
}

/** Six Card Golf game config. */
export interface SixCardGolfConfig {
  playerCount: number;
  cpuDifficulty: number;
  rounds: number;
}

/** Six Card Golf API response. */
export interface SixCardGolfResponse extends BaseGameResponse {
  players: SixCardGolfPlayerData[];
  phase: number;
  roundNumber: number;
  totalRounds: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  drawnCard: Card | null;
  drawnFromDiscard: boolean;
  canFlip: boolean;
  finalTurnTrigger: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  config: SixCardGolfConfig;
}

/** Four Card Poker API response. */
export interface FourCardPokerResponse extends BaseGameResponse {
  /** Player's 5-card hand. */
  playerHand: Card[];
  /** Dealer hand: during the action phase only the upcard is revealed
   * (length 1); after the end phase all 6 cards are revealed. */
  dealerHand: Card[];
  /** Player's best 4-card subset (populated at end phase). */
  playerBest: Card[];
  /** Dealer's best 4-card subset (populated at end phase). */
  dealerBest: Card[];
  phase: number;
  chips: number;
  anteBet: number;
  acesUpBet: number;
  playBet: number;
  playMultiplier: number;
  result: number;
  antePayout: number;
  playPayout: number;
  anteBonusPayout: number;
  acesUpPayout: number;
  totalPayout: number;
  playerHandRank: number;
  dealerHandRank: number;
}

/** Dou Dizhu player action record. */
export interface DoudizhuAction {
  playerIdx: number;
  playedCards: Card[] | null;
  bidValue: number;
}

/** Dou Dizhu player data. */
export interface DoudizhuPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  isLandlord: boolean;
  cardCount: number;
  cards: Card[];
}

/** Dou Dizhu config. */
export interface DoudizhuConfig {
  cpuDifficulty: number;
}

/** Dou Dizhu API response. */
export interface DoudizhuResponse extends BaseGameResponse {
  players: DoudizhuPlayerData[];
  phase: string;
  currentTurn: number;
  tableCards: Card[];
  tableCombo: string;
  kittyCards: Card[];
  landlordIdx: number;
  baseBid: number;
  highestBid: number;
  bombCount: number;
  scores: number[];
  gameEndFlag: boolean;
  config: DoudizhuConfig;
  cpuActions: DoudizhuAction[];
  humanAction: DoudizhuAction | null;
}

/** Tichu player action record. */
export interface TichuAction {
  playerIdx: number;
  playedCards: Card[] | null;
  declType: number;
  isPass: boolean;
}

/** Tichu player data. */
export interface TichuPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  team: number;
  rank: number;
  declType: number;
  cardCount: number;
  cards: Card[];
}

/** Tichu config. */
export interface TichuConfig {
  cpuDifficulty: number;
}

/** Tichu API response. */
export interface TichuResponse extends BaseGameResponse {
  players: TichuPlayerData[];
  phase: string;
  currentTurn: number;
  tableCards: Card[];
  tableCombo: string;
  lastPlayIdx: number;
  startLeader: number;
  finishOrder: number[];
  scores: number[];
  isOneTwo: boolean;
  bombCount: number;
  gameEndFlag: boolean;
  config: TichuConfig;
  cpuActions: TichuAction[];
  humanAction: TichuAction | null;
}

/** Bourré player data. */
export interface BourrePlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  folded: boolean;
  decided: boolean;
  drawn: boolean;
  bourreed: boolean;
  chips: number;
  tricks: number;
  cardCount: number;
  cards: Card[];
}

/** A single card played into a Bourré trick. */
export interface BourreTrickCardData {
  playerIdx: number;
  card: Card | null;
}

/** Bourré hand result for one player. */
export interface BourreResultData {
  playerIdx: number;
  tricks: number;
  wonAmount: number;
  bourreed: boolean;
  folded: boolean;
}

/** Bourré config. */
export interface BourreConfig {
  cpuDifficulty: number;
}

/** Bourré API response. */
export interface BourreResponse extends BaseGameResponse {
  players: BourrePlayerData[];
  phase: string;
  currentPlayerIdx: number;
  dealerIdx: number;
  pot: number;
  carryPot: number;
  trumpSuit: string;
  trumpCard: Card | null;
  trickNumber: number;
  currentTrick: BourreTrickCardData[];
  lastTrick: BourreTrickCardData[];
  lastTrickWinner: number;
  leadPlayerIdx: number;
  handNumber: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  validPlays: number[];
  results: BourreResultData[];
  config: BourreConfig;
}

// --- Spoons ---

/** Spoons game phase (0=Pass 1=Grab 2=RoundEnd 3=GameEnd). */
export type SpoonsPhaseValue = 0 | 1 | 2 | 3;

/**
 * A Spoons player's public/own state. `hand` is non-empty only for the human
 * (seat 0); CPU hands are returned as an empty array.
 */
export interface SpoonsPlayer {
  /** Display name ("あなた" / "CPU"). */
  name: string;
  isHuman: boolean;
  /** Number of cards currently held. */
  handSize: number;
  /** The player's cards — populated only for the human. */
  hand: Card[];
  /** Number of S-P-O-O-N-S letters collected (0–6). */
  letters: number;
  /** Whether the player has been eliminated (6 letters). */
  eliminated: boolean;
  /** Whether the player currently holds a grabbed spoon. */
  hasSpoon: boolean;
}

/**
 * Full Spoons game state returned from the API.
 *
 * Spoons is a 4-player pass-and-grab speed game played with a 52-card deck (4
 * cards each). Players continuously pass a card to the next player; when someone
 * collects four of a kind they grab a spoon and everyone races for the
 * remaining spoons (one fewer than the number of players). The player left
 * without a spoon gains a letter — S, P, O, O, N, S. After six letters that
 * player is eliminated; the last player standing wins.
 */
export interface SpoonsResponse extends BaseGameResponse {
  phase: SpoonsPhaseValue;
  gameEndFlag: boolean;
  /** Winning seat index, or -1 until the game ends. */
  winnerIdx: number;
  /** Seat index whose turn it currently is. */
  currentPlayerIdx: number;
  /** Seat index of the "feeder" who draws from the draw pile this round. */
  feederIdx: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  /** Spoons still on the table to be grabbed. */
  spoonsRemaining: number;
  /** Whether the grab window is open (race to grab a spoon). */
  grabWindowOpen: boolean;
  /** Seat index of the first player to grab this round, or -1 until one grabs. */
  firstGrabberIdx: number;
  /** Seat index of the player who missed out this round, or -1 until decided. */
  roundLoserIdx: number;
  /** Current round number (1-based). */
  roundNumber: number;
  /** Cards remaining in the feeder's draw pile. */
  drawPileSize: number;
  players: SpoonsPlayer[];
  cpuDifficulty: number;
}

/** Cuckoo phase value: 0=Turn, 1=Refuse, 2=RoundEnd, 3=GameEnd. */
export type CuckooPhaseValue = 0 | 1 | 2 | 3;

/**
 * A single Cuckoo player as returned from the API.
 *
 * Four seats each hold a single card. The human (seat 0) always sees their own
 * card; opponents' cards are `null` until they are revealed at round end or by
 * a King reveal.
 */
export interface CuckooPlayer {
  /** Seat index (0 = human). */
  id: number;
  isHuman: boolean;
  /** The player's single card, or `null` while hidden / eliminated. */
  card: Card | null;
  /** Remaining lives (♥). 0 means eliminated. */
  lives: number;
  /** Whether this player has been knocked out of the game. */
  isEliminated: boolean;
  /** Whether this player has revealed a King to block a swap. */
  kingRevealed: boolean;
  /** Whether it is currently this player's turn. */
  isCurrentTurn: boolean;
}

/**
 * Full Cuckoo (a.k.a. Chase the Ace / Ranter-Go-Round) game state.
 *
 * A simple 4-player life-survival game. Each player holds one card and three
 * lives. On your turn you keep your card or swap it with your neighbour (the
 * dealer swaps with the stock); a King holder may refuse an incoming swap by
 * revealing the King. After everyone acts, the holder(s) of the lowest card
 * lose a life; at zero lives a player is eliminated. The last player standing
 * wins.
 */
export interface CuckooResponse extends BaseGameResponse {
  players: CuckooPlayer[];
  phase: CuckooPhaseValue;
  /** Current round number (1-based). */
  roundNumber: number;
  /** Seat index whose turn it currently is. */
  currentPlayerIdx: number;
  /** Seat index of the dealer this round. */
  dealerIdx: number;
  /** Cards remaining in the stock. */
  stockCount: number;
  gameEndFlag: boolean;
  /** Winning seat index, or -1 until the game ends. */
  winnerIdx: number;
  /** Seat attempting a swap, or -1. */
  pendingSwapFrom: number;
  /** Target seat of a pending swap (the King holder who may refuse), or -1. */
  pendingSwapTo: number;
  /** The lowest card value held this round, or -1 until decided. */
  roundLowest: number;
  /** Seat indices that held the lowest card and lost a life this round. */
  roundLosers: number[];
  config: CuckooConfig;
}

/** Cuckoo configuration as returned from the API. */
export interface CuckooConfig {
  cpuDifficulty: number;
  initialLives: number;
}

/**
 * Pişti game phase, mirroring the backend `PishtiPhase` string values
 * (internal/domain/Pishti.go). The phase is a string, not a numeric enum.
 */
export type PishtiPhase = 'play' | 'roundEnd' | 'gameEnd';

/** A single Pişti player as returned from the API. */
export interface PishtiPlayer {
  /** Seat index (0 = human). */
  id: number;
  isHuman: boolean;
  /** Number of cards currently in hand. */
  cardCount: number;
  /** The player's hand cards (populated only for the human). */
  cards: Card[];
  /** Total number of cards captured so far. */
  capturedCount: number;
  /** Accumulated Pişti bonus points. */
  pistiBonus: number;
  /** Final score (populated once the game ends). */
  finalScore: number;
}

/** Pişti configuration as returned from / sent to the API. */
export interface PishtiConfig {
  /** Number of players (2-4). */
  playerCnt: number;
  /** CPU difficulty (0=Easy, 1=Normal, 2=Hard). */
  cpuDifficulty: number;
}

/** Server response for the Pişti game (POST /pishti/exec). */
export interface PishtiResponse extends BaseGameResponse {
  players: PishtiPlayer[];
  /** Seat index whose turn it currently is. */
  currentTurn: number;
  /** All cards currently on the central pile, bottom to top. */
  pile: Card[];
  /** The top card of the pile, or null when the pile is empty. */
  pileTop: Card | null;
  /** Number of cards on the pile. */
  pileCount: number;
  /** Seat index of the most recent capturer, or -1. */
  lastCaptureIdx: number;
  gameEndFlag: boolean;
  /** Current game phase (a string, not a numeric enum). */
  phase: PishtiPhase | string;
  /** Cards remaining in the stock. */
  remainingDeck: number;
  /** Winning seat indices (may tie), empty until the game ends. */
  winners: number[];
  /** Final scores indexed by seat, empty until the game ends. */
  finalScores: number[];
  config: PishtiConfig;
}

/** A single Cuarenta player as returned from the API. */
export interface CuarentaPlayer {
  /** Seat index (0 = human; seats {0,2}=Team A, {1,3}=Team B). */
  id: number;
  /** Team index (0 = Team A, 1 = Team B). */
  team: number;
  isHuman: boolean;
  /** Number of cards currently in hand. */
  cardCount: number;
  /** The player's hand cards (populated only for the human). */
  cards: Card[];
  /** Total number of cards captured by this player so far this round. */
  capturedCount: number;
}

/** A single Cuarenta play action (human or CPU), describing what was captured. */
export interface CuarentaAction {
  /** Seat index of the acting player. */
  playerIdx: number;
  /** The card that was played, or null. */
  playedCard: Card | null;
  /** Cards captured by this play (empty when the card was laid on the table). */
  capturedCards: Card[];
  /** True when this play scored a caída (+2). */
  isCaida: boolean;
  /** True when this play cleared the table (limpia, +1). */
  isLimpia: boolean;
  /** Extra ronda points scored by this play (0 when none). */
  rondaBonus: number;
}

/** Round-end scoring breakdown keyed by team index. */
export interface CuarentaScoreDetail {
  /** Cards captured this round, keyed by team index. */
  capturedCount: Record<string, number>;
  /** Caída points, keyed by team index. */
  caida: Record<string, number>;
  /** Ronda points, keyed by team index. */
  ronda: Record<string, number>;
  /** Limpia points, keyed by team index. */
  limpia: Record<string, number>;
  /** Team index awarded the most-cards bonus, or -1 when none. */
  mostCards: number;
  /** Total points gained this round, keyed by team index. */
  gained: Record<string, number>;
}

/** Cuarenta configuration as returned from / sent to the API. */
export interface CuarentaConfig {
  /** Target score to win the game (default 40). */
  targetScore: number;
  /** CPU difficulty (0=Easy, 1=Normal, 2=Hard). */
  cpuDifficulty: number;
}

/** Server response for the Cuarenta game (POST /cuarenta/exec). */
export interface CuarentaResponse extends BaseGameResponse {
  players: CuarentaPlayer[];
  /** Seat index whose turn it currently is. */
  currentTurn: number;
  /** All cards currently on the central table. */
  tableCards: Card[];
  /** Seat index of the most recent capturer, or -1. */
  lastCaptureIdx: number;
  gameEndFlag: boolean;
  /** Current game phase (0=Play, 1=RoundEnd, 2=GameEnd). */
  phase: number;
  /** Cumulative score per team, indexed by team. */
  teamScores: number[];
  /** Cards remaining in the stock. */
  remainingDeck: number;
  /** Winning team indices (may tie), empty until the game ends. */
  roundWinners: number[];
  /** CPU actions that occurred since the last human play. */
  cpuActions: CuarentaAction[];
  /** The human's most recent action, or null. */
  humanAction: CuarentaAction | null;
  /** The most recent round-end scoring breakdown, or null. */
  lastRoundDetail: CuarentaScoreDetail | null;
  config: CuarentaConfig;
}

/** A single chip bet placed on one rank of the Faro layout. */
export interface FaroBet {
  /** Rank the chip is placed on (1=A .. 13=K). */
  rank: number;
  /** Wagered chip amount. */
  amount: number;
  /** True when the bet is a "copper" — wagering the rank to lose rather than win. */
  copper: boolean;
}

/** Server response for the Faro game (POST /faro/exec). */
export interface FaroResponse extends BaseGameResponse {
  /** Current phase (1=Betting, 2=Turn, 3=Call, 4=RoundEnd, 5=GameEnd). */
  phase: number;
  /** Player's remaining bankroll in chips. */
  chips: number;
  /** Chips currently placed on the layout, one entry per wagered rank. */
  bets: FaroBet[];
  /** The burned "soda" card (first card of the deal), or null before the deal. */
  soda: Card | null;
  /** The most recent turn's losing card (1st turned — bank collects), or null. */
  losingCard: Card | null;
  /** The most recent turn's winning card (2nd turned — pays the player), or null. */
  winningCard: Card | null;
  /** True when the last turn was a split (both cards the same rank — bank takes half). */
  split: boolean;
  /** Number of turns dealt so far this round. */
  turnsPlayed: number;
  /** Total number of turns in a full round. */
  turnsTotal: number;
  /** Number of cards still left in the dealing box. */
  remaining: number;
  /** The final three cards available to call (populated during the Call phase). */
  callCards: Card[];
  /** The player's predicted order for the called cards (rank values), empty when none. */
  callOrder: number[];
  /** True when the most recent call was correct (paid 4:1). */
  callWon: boolean;
  /** Net chip change for the just-finished round. */
  totalPayout: number;
  gameEndFlag: boolean;
}

/** A single player as returned from the Open Face Chinese Poker (OFC) API. */
export interface OpenFaceChinesePlayer {
  /** Seat index (0 = human). */
  id: number;
  /** True for the human player. */
  isHuman: boolean;
  /** Top row (up to 3 cards). */
  front: Card[];
  /** Middle row (up to 5 cards). */
  middle: Card[];
  /** Bottom row (up to 5 cards). */
  back: Card[];
  /** The pending card(s) awaiting placement (human only; empty for CPU). */
  pending: Card[];
  /** Net points scored in the just-finished round. */
  roundScore: number;
  /** Royalty bonus points earned this round. */
  royalty: number;
  /** True when the hand is fouled (rows not in non-decreasing strength order). */
  fouled: boolean;
  /** True when the player qualified for Fantasyland. */
  fantasyland: boolean;
  /** Cumulative score across all rounds. */
  totalScore: number;
}

/** Open Face Chinese Poker (OFC) config echoed back by the server. */
export interface OpenFaceChineseConfig {
  cpuDifficulty: number;
  playerCount: number;
  targetRounds: number;
}

/** A placement hint returned by the Open Face Chinese Poker (OFC) /openfacechinese/exec endpoint. */
export interface OpenFaceChineseHint {
  /** Suggested row (0=front, 1=middle, 2=back). */
  row: number;
  /** Human-readable rationale for the suggestion. */
  reason: string;
}

/** Server response for the Open Face Chinese Poker (OFC) game (POST /openfacechinese/exec). */
export interface OpenFaceChineseResponse extends BaseGameResponse {
  /** Current phase (0=Placing, 1=RoundEnd, 2=GameEnd). */
  phase: number;
  /** 1-based round number. */
  roundNumber: number;
  /** Seat index whose turn it currently is. */
  currentPlayerIdx: number;
  /** Seat index of the current dealer. */
  dealerIdx: number;
  /** The card the human must place this turn (present only on the human's turn). */
  currentCard?: Card;
  /** True when the game has ended. */
  gameEndFlag: boolean;
  /** Winning seat index, or -1 for a draw. */
  winnerIdx: number;
  /** True when it is the human player's turn to place a card. */
  isHumanTurn: boolean;
  /** Optional placement hint (present only on a hint request). */
  hint?: OpenFaceChineseHint;
  /** One entry per player. */
  players: OpenFaceChinesePlayer[];
  /** Echoed game configuration. */
  config: OpenFaceChineseConfig;
}

/** A single Russian Bank (Crapette) player as returned from the API. */
export interface RussianBankPlayer {
  /** Seat index (0 = human). */
  id: number;
  /** True for the human player. */
  isHuman: boolean;
  /** Number of cards left in the reserve (the pile to empty to win). */
  reserveCount: number;
  /** Top reserve card (face up), if any. */
  reserveTop?: Card;
  /** Number of face-down cards left in hand. */
  handCount: number;
  /** Number of cards in the waste pile. */
  wasteCount: number;
  /** Top waste card, if any. */
  wasteTop?: Card;
  /** Number of times this player caught the opponent with "stop". */
  stopPoints: number;
}

/** Russian Bank (Crapette) config echoed back by the server. */
export interface RussianBankConfig {
  cpuDifficulty: number;
}

/** A move hint returned by the Russian Bank /russianbank/exec endpoint. */
export interface RussianBankHint {
  /** Source zone (0=reserve, 1=waste, 2=tableau). */
  zone: number;
  /** True when the source is the opponent's pile. */
  fromOpponent: boolean;
  /** Source tableau column (when zone=tableau). */
  col: number;
  /** True when the destination is a foundation. */
  toFoundation: boolean;
  /** Destination tableau column (when toFoundation is false). */
  toCol: number;
}

/** Server response for the Russian Bank (Crapette) game (POST /russianbank/exec). */
export interface RussianBankResponse extends BaseGameResponse {
  /** Current phase (0=Idle, 1=Playing, 2=GameEnd). */
  phase: number;
  /** Seat index whose turn it currently is. */
  currentPlayerIdx: number;
  /** True when the game has ended. */
  gameEndFlag: boolean;
  /** Winning seat index, or -1 for a draw. */
  winnerIdx: number;
  /** True when it is the human player's turn. */
  isHumanTurn: boolean;
  /** True when the human may call "stop" on the CPU. */
  canCallStop: boolean;
  /** True when the human's last move can be undone. */
  canUndo: boolean;
  /** Total moves played so far. */
  moveCount: number;
  /** The 4 shared tableau columns (top card is last). */
  tableau: Card[][];
  /** The 8 shared foundations (A-up by suit; top card is last). */
  foundations: Card[][];
  /** Optional move hint (present only on a hint request). */
  hint?: RussianBankHint;
  /** One entry per player (index 0 = human). */
  players: RussianBankPlayer[];
  /** Echoed game configuration. */
  config: RussianBankConfig;
}

/** Kemps phase value: 0=Exchange, 1=Declare, 2=RoundEnd, 3=GameEnd. */
export type KempsPhaseValue = 0 | 1 | 2 | 3;

/**
 * A single Kemps player as returned from the API.
 *
 * Four seats split into two teams (even seats = Team A, odd seats = Team B).
 * Only the human (seat 0) has a populated `hand`; CPU hands are an empty array.
 */
export interface KempsPlayer {
  /** Display name ("あなた" / "CPU"). */
  name: string;
  isHuman: boolean;
  /** Team number (0 = Team A, 1 = Team B). */
  team: number;
  /** Number of cards currently held (always 4 during play). */
  handSize: number;
  /** The player's cards — populated only for the human. */
  hand: Card[];
  /** Whether this player currently holds four of a kind (human only). */
  hasFourOfAKind: boolean;
}

/**
 * Full Kemps game state returned from the API.
 *
 * Kemps is a 4-player, 2-team matching game. Each turn a player swaps one hand
 * card for a card on the shared 4-card field, trying to collect four of a kind.
 * When the human's partner secretly signals, the team races to declare "Kemps!"
 * for a point; the opposing team can call "Counter-Kemps!" against a seat to
 * steal it (−1 if wrong). First team to the target score (default 5) wins.
 */
export interface KempsResponse extends BaseGameResponse {
  phase: KempsPhaseValue;
  gameEndFlag: boolean;
  /** Winning team (0 or 1), or -1 until the game ends. */
  winnerTeam: number;
  /** Seat index whose turn it currently is. */
  currentPlayerIdx: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  /** Team scores indexed by team number (Team A = index 0). */
  teamScores: number[];
  /** The shared field of cards available to swap with. */
  field: Card[];
  /** The human's secret signal type (0=Sound, 1=Blink). */
  signalType: number;
  /** Whether the human's partner is currently signaling (human-only cue). */
  partnerSignaling: boolean;
  /** Whether an opponent may be signaling (vague human-only cue). */
  opponentSignaling: boolean;
  /** Seat index that completed four of a kind, or -1. */
  fourHolderIdx: number;
  /** Round result code (0=none, 1=Kemps, 2=Counter, 3=CounterFail, 4=Miss). */
  roundResult: number;
  /** Team that won the most recent round, or -1. */
  roundWinnerTeam: number;
  /** Current round number (1-based). */
  roundNumber: number;
  players: KempsPlayer[];
  cpuDifficulty: number;
  /** Team score required to win (default 5). */
  targetScore: number;
}

/** Player data for a Beggar-My-Neighbour participant. */
export interface BeggarMyNeighbourPlayerData {
  /** Player index (0 = human, 1 = CPU). */
  id: number;
  /** Whether this player is the human. */
  isHuman: boolean;
  /** Number of cards in the face-down draw pile. */
  drawPileSize: number;
  /** Number of cards in the discard pile (refills draw pile when empty). */
  discardPileSize: number;
  /** Total cards held (drawPileSize + discardPileSize). */
  totalCards: number;
}

/** Beggar-My-Neighbour game configuration. */
export interface BeggarMyNeighbourConfig {
  /** Maximum rounds before the game is decided by card count. */
  maxRounds: number;
}

/** Full Beggar-My-Neighbour game state returned from the API. */
export interface BeggarMyNeighbourResponse extends BaseGameResponse {
  players: BeggarMyNeighbourPlayerData[];
  /** Current game phase (0=Play, 1=PayPenalty, 2=Collect, 3=GameEnd). */
  phase: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  /** Index of the player whose turn it is (0=human, 1=CPU). */
  currentPlayerIdx: number;
  /** Index of the player who played the last penalty card (-1 if none). */
  penaltyOwnerIdx: number;
  /** Number of penalty cards still to be paid. */
  penaltyRemaining: number;
  /** Number of cards in the central pile. */
  centralPileSize: number;
  /** The last card played onto the central pile, or null. */
  lastCardPlayed: Card | null;
  /** Number of collection rounds completed. */
  roundsPlayed: number;
  config: BeggarMyNeighbourConfig;
}

// --- All Fours (Seven Up / Old Sledge) ---

/** All Fours player data (2-player: 0 = human elder hand, 1 = CPU dealer). */
export interface AllFoursPlayerData {
  /** Player index (0 = non-dealer/human, 1 = dealer/CPU). */
  id: number;
  /** Whether this player is the human. */
  isHuman: boolean;
  /** Number of cards in hand. */
  cardCount: number;
  /** Cards in hand (only populated for the human). */
  cards: Card[];
  /** Points scored this deal so far. */
  roundScore: number;
  /** Cumulative game score. */
  cumulativeScore: number;
  /** Number of tricks captured this deal. */
  trickCount: number;
}

/** A single card played to the current All Fours trick. */
export interface AllFoursTrickCard {
  playerIdx: number;
  card: Card;
}

/** All Fours hint payload (one of card/beg/run is set). */
export interface AllFoursHint {
  cardIndex?: number;
  beg?: boolean;
  run?: boolean;
  reason: string;
}

/** All Fours game configuration. */
export interface AllFoursConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A High/Low point award: the capturing player and the trump card, or -1 if unawarded. */
export interface AllFoursBreakdownAward {
  winnerIdx: number;
  card: Card | null;
}

/** The round-end point breakdown for All Fours (High/Low/Jack/Game). */
export interface AllFoursRoundBreakdown {
  high: AllFoursBreakdownAward;
  low: AllFoursBreakdownAward;
  /** Captor of the trump Jack, or -1 if no trump Jack was in play. */
  jack: { winnerIdx: number };
  /** Game point (most card pips): winner (-1 on tie/zero) and per-player pip totals. */
  game: { winnerIdx: number; points: number[] };
}

/** Full All Fours game state returned from the API. */
export interface AllFoursResponse extends BaseGameResponse {
  players: AllFoursPlayerData[];
  /** Current phase (0=Beg, 1=Gift, 2=Play, 3=TrickEnd, 4=RoundEnd, 5=GameEnd). */
  phase: number;
  roundNumber: number;
  trickNumber: number;
  dealerIdx: number;
  nonDealerIdx: number;
  currentPlayerIdx: number;
  trumpSuit: number;
  /** The turn-up card that set the provisional trump, or null. */
  turnUp: Card | null;
  /** Number of "run the cards" attempts this deal. */
  runCount: number;
  currentTrick: AllFoursTrickCard[];
  gameEndFlag: boolean;
  winnerIdx: number;
  leadPlayerIdx: number;
  validPlayIndices: number[];
  /** Present only at ROUND_END / GAME_END: the High/Low/Jack/Game point breakdown. */
  roundBreakdown?: AllFoursRoundBreakdown;
  config: AllFoursConfig;
  hint?: AllFoursHint;
}

// --- Guts ---

/** Guts game phase (0=Declare, 1=Result). */
export type GutsPhaseValue = 0 | 1;

/**
 * A Guts player's public/own state. `cards` is populated for the human and for
 * every player still `in` at showdown; `handName` is an i18n suffix
 * (`"pair"`, `"highcard"`, or `""`) set only when a hand is revealed.
 */
export interface GutsPlayer {
  id: number;
  isHuman: boolean;
  /** Remaining chips. */
  chips: number;
  /** Whether the player declared in (stayed) this round. */
  in: boolean;
  /** Whether the player has been eliminated (busted) from the match. */
  out: boolean;
  /** Chips this player has wagered / owes into the pot this round. */
  roundBet: number;
  cardCount: number;
  cards: Card[];
  /** The revealed hand-rank i18n suffix (`"pair"`, `"highcard"`), or empty. */
  handName?: string;
  /** Whether this player won the round's pot. */
  isWinner: boolean;
  /** Whether this player stayed in but lost and must match the pot. */
  isMatcher: boolean;
}

/** Guts local-rule configuration. */
export interface GutsConfig {
  /** Number of players at the table (2–7). */
  playerCount: number;
  /** Chips each player antes into the pot at the start of a round. */
  ante: number;
  /** Chips each player begins the match with. */
  startingChips: number;
  /** Number of rounds after which the richest player wins the match. */
  targetRounds: number;
}

/**
 * A suggested hint for Guts, computed by the backend. `declaration` is the
 * suggested call (0=out, 1=in) and `reason` is an i18n reason suffix
 * (`strong_hand` / `weak_hand`).
 */
export interface GutsHint {
  /** Suggested declaration: 0=out (fold), 1=in (stay). */
  declaration: number;
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Guts game state returned from the API.
 *
 * Guts is a fast multi-player pot-vying gambling game: every player antes, is
 * dealt 2 cards, and simultaneously declares "in" (stay) or "out" (fold). The
 * players who stayed reveal their hands; the best hand wins the pot, and every
 * other player who stayed must match the pot for the next round (which
 * carries over). If nobody stays, the pot carries to the next round. When
 * `targetRounds` rounds have been played the richest player wins the match.
 */
export interface GutsResponse extends BaseGameResponse {
  players: GutsPlayer[];
  /** Game phase: 0=Declare, 1=Result. */
  phase: GutsPhaseValue;
  roundNumber: number;
  /** Chips currently in the pot. */
  pot: number;
  /** Chips carried over from rounds where nobody stayed. */
  carryPot: number;
  /** Chips each player antes at the start of a round. */
  ante: number;
  /** The human's remaining chip stack. */
  chips: number;
  /** Winning seat index of the current round, or -1 for none. */
  winnerIdx: number;
  /** Winning seat index of the match, or -1 until it is decided. */
  matchWinnerIdx: number;
  /** The human's round result: 1=win, 0=none, -1=lose (matched). */
  result: number;
  /** Seat indices of players who stayed in but lost and must match the pot. */
  matchers: number[];
  gameEndFlag: boolean;
  hint?: GutsHint | null;
  config: GutsConfig;
}

// --- Anaconda (Pass the Trash) ---

/** Anaconda game phase (0=Pass, 1=Set, 2=Roll, 3=Result). */
export type AnacondaPhaseValue = 0 | 1 | 2 | 3;

/**
 * An Anaconda player's public/own state. `cards` holds the revealed cards: the
 * human sees their full hand, CPUs show only the `rollIndex`-length prefix
 * revealed so far during Roll (and the full 5 at showdown if still active).
 * `handName` is a poker-category i18n suffix (`"fourkind"`, `"flush"`, …) set
 * only when a full 5-card hand is revealed.
 */
export interface AnacondaPlayer {
  id: number;
  isHuman: boolean;
  /** Remaining chips. */
  chips: number;
  /** Whether the player folded out of the current round. */
  folded: boolean;
  /** Whether the player has been eliminated (busted) from the match. */
  out: boolean;
  /** Chips this player has wagered into the pot across the whole round. */
  roundBet: number;
  /** Chips this player has wagered on the current betting street. */
  streetBet: number;
  cardCount: number;
  /** Revealed cards (see interface doc). */
  cards: Card[];
  /** The revealed poker-category i18n suffix, or empty until a 5-card hand shows. */
  handName?: string;
  /** Whether this player won the round's pot. */
  isWinner: boolean;
}

/** Anaconda local-rule configuration. */
export interface AnacondaConfig {
  /** Number of players at the table (3–7). */
  playerCount: number;
  /** Chips each player antes into the pot at the start of a round. */
  ante: number;
  /** Chips each player begins the match with. */
  startingChips: number;
  /** Number of rounds after which the richest player wins the match. */
  targetRounds: number;
}

/**
 * A suggested hint for Anaconda, computed by the backend. `action` is the
 * suggested move (`"pass"` / `"keep"` / `"raise"` / `"call"` / `"fold"`),
 * `cardIndices` accompanies pass/keep suggestions, and `reason` is an i18n
 * reason suffix (`pass_weakest` / `keep_best` / `strong_hand` / `medium_hand` /
 * `weak_hand`).
 */
export interface AnacondaHint {
  /** Suggested action: pass/keep/raise/call/fold. */
  action: string;
  /** Card indices to pass or keep, when the suggestion is a pass/keep. */
  cardIndices?: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Anaconda (Pass the Trash) game state returned from the API.
 *
 * Anaconda is a poker pot game: each player is dealt 7 cards, then passes 3,
 * then 2, then 1 card to the left; keeps their best 5 (discarding 2); and
 * reveals them one at a time ("roll") with a betting round between reveals.
 * The best 5-card poker hand among the players still active wins the pot.
 */
export interface AnacondaResponse extends BaseGameResponse {
  players: AnacondaPlayer[];
  /** Game phase: 0=Pass, 1=Set, 2=Roll, 3=Result. */
  phase: AnacondaPhaseValue;
  roundNumber: number;
  /** Seat index of the current dealer. */
  dealerIdx: number;
  /** Seat index of the player to act. */
  currentPlayer: number;
  /** Cards still to pass this sub-round (3/2/1 during Pass, 0 otherwise). */
  passCount: number;
  /** Cards revealed so far during Roll (0–5). */
  rollIndex: number;
  /** Chips currently in the pot. */
  pot: number;
  /** The current bet to match on this betting street. */
  currentBet: number;
  /** Number of raises already made on the current street. */
  raiseCount: number;
  /** The maximum number of raises allowed per street. */
  maxRaises: number;
  /** Chips each player antes at the start of a round. */
  ante: number;
  /** The human's remaining chip stack. */
  chips: number;
  /** Winning seat index of the current round, or -1 for none. */
  winnerIdx: number;
  /** Winning seat index of the match, or -1 until it is decided. */
  matchWinnerIdx: number;
  /** The human's round result: 1=win, 0=none, -1=lose. */
  result: number;
  gameEndFlag: boolean;
  /** Whether it is the human's turn to act. */
  isHumanTurn: boolean;
  /** Whether the human may raise (raises remain and chips suffice). */
  canRaise: boolean;
  hint?: AnacondaHint | null;
  config: AnacondaConfig;
}

/** Bouillotte game phase (0=Betting, 1=Result). */
export type BouillottePhaseValue = 0 | 1;

/**
 * A Bouillotte player's public/own state. `cards` is populated for the human
 * and, at the result phase, for every player who has not folded; `handName` is
 * an i18n suffix (`"brelan"`, `"highcard"`, or `""`) set only when a hand is
 * revealed.
 */
export interface BouillottePlayer {
  id: number;
  isHuman: boolean;
  /** Remaining chips. */
  chips: number;
  /** Chips this player has wagered into the pot this round. */
  roundBet: number;
  /** Whether the player has folded out of the current round. */
  folded: boolean;
  /** Whether the player has been eliminated (busted) from the match. */
  out: boolean;
  cardCount: number;
  cards: Card[];
  /** The revealed hand-rank i18n suffix (`"brelan"`, `"highcard"`), or empty. */
  handName?: string;
  /** Whether this player won the round's pot. */
  isWinner: boolean;
}

/** Bouillotte local-rule configuration. */
export interface BouillotteConfig {
  /** Number of players at the table (3–4). */
  playerCount: number;
  /** Chips each player antes into the pot at the start of a round. */
  ante: number;
  /** Chips each player begins the match with. */
  startingChips: number;
  /** Number of rounds after which the richest player wins the match. */
  targetRounds: number;
}

/**
 * A suggested hint for Bouillotte, computed by the backend. `action` is the
 * suggested betting action (`"call"` / `"raise"` / `"fold"`) and `reason` is an
 * i18n reason suffix (`strong_hand` / `medium_hand` / `weak_hand`).
 */
export interface BouillotteHint {
  /** Suggested betting action: `"call"`, `"raise"`, or `"fold"`. */
  action: string;
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Bouillotte game state returned from the API.
 *
 * Bouillotte is an 18th-century French 3-card poker-vying pot game. Each player
 * antes, is dealt 3 cards, and a shared "retourne" card is turned up. Players
 * take turns to call, raise (vie), or fold; when betting closes the non-folded
 * players reveal their hands and the best hand (a brelan of three matching
 * ranks — counting the retourne — beats a high card) takes the pot. Chips
 * accumulate across rounds; after `targetRounds` rounds the richest player wins.
 */
export interface BouillotteResponse extends BaseGameResponse {
  players: BouillottePlayer[];
  /** Game phase: 0=Betting, 1=Result. */
  phase: BouillottePhaseValue;
  roundNumber: number;
  /** Chips currently in the pot. */
  pot: number;
  /** Chips each player antes at the start of a round. */
  ante: number;
  /** The human's remaining chip stack. */
  chips: number;
  /** The current bet each active player must match to stay in. */
  currentBet: number;
  /** Number of raises made this round. */
  raiseCount: number;
  /** Maximum raises permitted this round. */
  maxRaises: number;
  /** Seat index of the player to act. */
  currentPlayerIdx: number;
  /** Seat index of the dealer. */
  dealerIdx: number;
  /** The shared turned-up "retourne" card, or null before it is dealt. */
  retourne: Card | null;
  /** Whether it is the human's turn to act. */
  isHumanTurn: boolean;
  /** Whether the human may currently raise (vie). */
  canRaise: boolean;
  /** Winning seat index of the current round, or -1 for none. */
  winnerIdx: number;
  /** Winning seat index of the match, or -1 until it is decided. */
  matchWinnerIdx: number;
  /** The human's round result: 1=win, 0=none, -1=lose. */
  result: number;
  gameEndFlag: boolean;
  hint?: BouillotteHint | null;
  config: BouillotteConfig;
}

/** Michigan game phase (0=Bet, 1=Play, 2=Result). */
export type MichiganPhaseValue = 0 | 1 | 2;

/**
 * A Michigan player's public/own state. `cards` is populated for the human
 * during the play phase and revealed for every player at the result phase; CPU
 * hands are empty (`cards: []`, use `cardCount`) while the round is in progress.
 */
export interface MichiganPlayer {
  id: number;
  isHuman: boolean;
  /** Remaining chips. */
  chips: number;
  /** Chips this player has wagered across the boodles this round. */
  roundBet: number;
  cardCount: number;
  cards: Card[];
  /** Whether it is this player's turn to act. */
  isCurrent: boolean;
  /** Whether this player emptied their hand to end the round. */
  isWinner: boolean;
}

/**
 * A Michigan boodle — one of the four center "betting" cards (A♥, K♣, Q♦, J♠)
 * onto which players stake chips. When a player plays a card matching the
 * boodle they collect its chips.
 */
export interface MichiganBoodle {
  /** The fixed boodle card (A♥, K♣, Q♦, or J♠). */
  card: Card;
  /** Chips currently staked on this boodle. */
  chips: number;
  /** Seat index of the player who claimed the boodle's chips, or -1 if unclaimed. */
  claimedBy: number;
}

/** Michigan local-rule configuration. */
export interface MichiganConfig {
  /** Number of players at the table (3–8). */
  playerCount: number;
  /** Total chips each player distributes across the four boodles per round. */
  ante: number;
  /** Chips each player begins the match with. */
  startingChips: number;
  /** Number of rounds after which the richest player wins the match. */
  targetRounds: number;
}

/**
 * A suggested hint for Michigan, computed by the backend. `cardIndex` is the
 * hand index to play and `reason` is an i18n reason suffix (`forced` /
 * `claim_boodle` / `lead_low`).
 */
export interface MichiganHint {
  /** Hand index of the suggested card to play. */
  cardIndex: number;
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Michigan game state returned from the API.
 *
 * Michigan (Newmarket) is a "stops" chip-betting game. Players first stake chips
 * across four center "boodle" cards (A♥, K♣, Q♦, J♠), then play cards in
 * ascending same-suit sequences. Playing a card that matches a boodle wins its
 * chips; emptying your hand ends the round. After `targetRounds` rounds the
 * richest player wins the match.
 */
export interface MichiganResponse extends BaseGameResponse {
  players: MichiganPlayer[];
  boodles: MichiganBoodle[];
  /** Game phase: 0=Bet, 1=Play, 2=Result. */
  phase: MichiganPhaseValue;
  roundNumber: number;
  /** Total chips each player stakes across the boodles per round. */
  ante: number;
  /** The human's remaining chip stack. */
  chips: number;
  /** Chips the human must distribute across the four boodles this round. */
  betBudget: number;
  /** Whether the human has already placed their boodle bets this round. */
  humanBetPlaced: boolean;
  /** Seat index of the player to act. */
  currentPlayerIdx: number;
  /** Seat index of the dealer. */
  dealerIdx: number;
  /** Seat index of the player who leads the current sequence. */
  leadPlayerIdx: number;
  /** Suit of the active sequence (0=none/new sequence needed, 1–4=suit). */
  seqSuit: number;
  /** Display name of the active sequence's suit, or empty when a new one is needed. */
  seqSuitName: string;
  /** Highest card value played so far in the current run. */
  seqHighValue: number;
  /** Whether the current player must start a fresh sequence. */
  needNewSequence: boolean;
  /** Number of cards in the face-down dead hand. */
  deadHandCount: number;
  /** Whether it is the human's turn to act. */
  isHumanTurn: boolean;
  /** Legal hand indices the human may play this turn. */
  playableIndices: number[];
  /** Seat index of the player who emptied their hand, or -1 for none. */
  winnerIdx: number;
  /** Winning seat index of the match, or -1 until it is decided. */
  matchWinnerIdx: number;
  /** The human's round result: 1=win, 0=none, -1=lose. */
  result: number;
  gameEndFlag: boolean;
  hint?: MichiganHint | null;
  config: MichiganConfig;
}

/** Primero game phase (0=Betting, 1=Result). */
export type PrimeroPhaseValue = 0 | 1;

/**
 * A Primero player's public/own state. `cards` is populated for the human and,
 * at the result phase, for every player who has not folded; `handName` is an
 * i18n suffix (`"fluxus"` / `"supremus"` / `"primero"` / `"numerus"`, or `""`)
 * set only when a hand is revealed.
 */
export interface PrimeroPlayer {
  id: number;
  isHuman: boolean;
  /** Remaining chips. */
  chips: number;
  /** Chips this player has wagered into the pot this round. */
  roundBet: number;
  /** Whether the player has folded out of the current round. */
  folded: boolean;
  /** Whether the player has been eliminated (busted) from the match. */
  out: boolean;
  cardCount: number;
  cards: Card[];
  /** The revealed hand-rank i18n suffix (`"fluxus"` / `"supremus"` / `"primero"` / `"numerus"`), or empty. */
  handName?: string;
  /** Whether this player won the round's pot. */
  isWinner: boolean;
}

/** Primero local-rule configuration. */
export interface PrimeroConfig {
  /** Number of players at the table (2–6). */
  playerCount: number;
  /** Chips each player antes into the pot at the start of a round. */
  ante: number;
  /** Chips each player begins the match with. */
  startingChips: number;
  /** Number of rounds after which the richest player wins the match. */
  targetRounds: number;
}

/**
 * A suggested hint for Primero, computed by the backend. `action` is the
 * suggested betting action (`"call"` / `"raise"` / `"fold"`) and `reason` is an
 * i18n reason suffix (`strong_hand` / `medium_hand` / `weak_hand`).
 */
export interface PrimeroHint {
  /** Suggested betting action: `"call"`, `"raise"`, or `"fold"`. */
  action: string;
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Primero game state returned from the API.
 *
 * Primero is a Renaissance (16th-century) vying pot game and an ancestor of
 * poker. Each player antes, is dealt 4 cards, and players take turns to call,
 * raise (vie), or fold; when betting closes the non-folded players reveal their
 * hands and the best hand takes the pot. Hands rank (low→high): Numerus, then
 * Primero (one card of each suit), Supremus, and Fluxus (four of a suit). Chips
 * accumulate across rounds; after `targetRounds` rounds the richest player wins.
 * Unlike Bouillotte there is no shared "retourne" card.
 */
export interface PrimeroResponse extends BaseGameResponse {
  players: PrimeroPlayer[];
  /** Game phase: 0=Betting, 1=Result. */
  phase: PrimeroPhaseValue;
  roundNumber: number;
  /** Chips currently in the pot. */
  pot: number;
  /** Chips each player antes at the start of a round. */
  ante: number;
  /** The human's remaining chip stack. */
  chips: number;
  /** The current bet each active player must match to stay in. */
  currentBet: number;
  /** Number of raises made this round. */
  raiseCount: number;
  /** Maximum raises permitted this round. */
  maxRaises: number;
  /** Seat index of the player to act. */
  currentPlayerIdx: number;
  /** Seat index of the dealer. */
  dealerIdx: number;
  /** Whether it is the human's turn to act. */
  isHumanTurn: boolean;
  /** Whether the human may currently raise (vie). */
  canRaise: boolean;
  /** Winning seat index of the current round, or -1 for none. */
  winnerIdx: number;
  /** Winning seat index of the match, or -1 until it is decided. */
  matchWinnerIdx: number;
  /** The human's round result: 1=win, 0=none, -1=lose. */
  result: number;
  gameEndFlag: boolean;
  hint?: PrimeroHint | null;
  config: PrimeroConfig;
}
