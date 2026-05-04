/** Card suit design identifier. */
export type CardDesign = 'SPADE' | 'CLOVER' | 'HEART' | 'DIAMOND' | 'JOKER';

/** A playing card with suit design and numeric value. */
export interface Card {
  design: CardDesign;
  value: number;
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
export interface BlackJackResponse {
  dealer: BlackJackPlayer;
  player: BlackJackPlayer;
  hands?: BlackJackHand[];
  currentHandIdx: number;
  phase: BlackJackPhase;
  insuranceBet: number;
  insuranceAvailable: boolean;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface PokerResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface BadugiResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  metaAI?: BadugiMetaAI;
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
export interface OldMaidResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface DaifugoResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  pendingAction: 'none' | 'sevenPass' | 'tenDiscard' | 'queenBomber';
  pendingActionTarget: number;
  reverseDirection: boolean;
  numberLocked: boolean;
  sequenceLocked: boolean;
  sortMode: number;
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
export interface SevensResponse {
  players: SevensPlayerData[];
  currentTurn: number;
  tableMinVals: number[]; // index 0 unused; 1=SPADE, 2=CLOVER, 3=HEART, 4=DIAMOND
  tableMaxVals: number[];
  tablePlaced: number[]; // bitmask per suit; bit i = value i placed
  config: SevensConfig;
  gameEndFlag: boolean;
  cpuActions: SevensAction[];
  humanAction: SevensAction | null;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface DoubtResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface HoldemResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
}

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
export interface HeartsResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  config: HeartsConfig;
  hint?: HeartsHint;
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
export interface SpadesResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  config: SpadesConfig;
  hint?: SpadesHint;
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
export interface PitchResponse {
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
  gameEndFlag: boolean;
  winnerIdx: number;
  leadPlayerIdx: number;
  validPlayIndices: number[];
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface TwoTenJackResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface CrazyEightsResponse {
  players: CrazyEightsPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  chosenSuit: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  config: CrazyEightsConfig;
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
export interface PageOneResponse {
  players: PageOnePlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface TonkResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  config: TonkConfig;
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
export interface SevenBridgeResponse {
  players: SevenBridgePlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  roundWinnerIdx: number;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  config: SevenBridgeConfig;
}

/** Full Gin Rummy game state returned from the API. */
export interface GinRummyResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  config: GinRummyConfig;
}

// --- Memory (神経衰弱) ---

/** Memory player data with pair count. */
export interface MemoryPlayerData {
  id: number;
  isHuman: boolean;
  pairCount: number;
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
export interface MemoryResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface KlondikeResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface CanfieldResponse {
  tableau: CanfieldTableauCard[][];
  reserve: Card[];
  stockCount: number;
  waste: Card[];
  foundation: Card[][];
  baseRank: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  hint?: CanfieldHint;
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
export interface FreeCellResponse {
  tableau: (Card | null)[][];
  freeCells: (Card | null)[];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  hint?: FreeCellHint;
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
export interface BaccaratResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface NapoleonResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  config: NapoleonConfig;
  hint?: NapoleonHint;
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
export interface SkatResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface ShitheadResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface SpiderResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  hint?: SpiderHint;
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
export interface IndianPokerResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface EuchreResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  config: EuchreConfig;
  hint?: EuchreHint;
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
export interface BridgeResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface PyramidResponse {
  pyramid: PyramidCard[][];
  stockCount: number;
  waste: Card[];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface TriPeaksResponse {
  layout: TriPeaksCard[][];
  stockCount: number;
  waste: Card[];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  hint?: TriPeaksHint;
}

/** Full Video Poker game state returned from the API. */
export interface VideoPokerResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface CribbageResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface OhHellResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
}

// --- Three Card Poker (スリーカードポーカー) ---

/** Three Card Poker API response. */
export interface ThreeCardResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
}

// --- Caribbean Stud Poker (カリビアンスタッドポーカー) ---

/** Caribbean Stud Poker API response. */
export interface CaribbeanStudResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
}

// --- Texas Hold'em Bonus Poker (テキサスホールデムボーナスポーカー) ---

/** Texas Hold'em Bonus Poker API response. */
export interface TexasHoldemBonusResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
}

// --- Pai Gow Poker (パイゴウポーカー) ---

/** Pai Gow Poker API response. */
export interface PaiGowResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface WarResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
}

/** Full Speed game state returned from the API. */
export interface SpeedResponse {
  players: SpeedPlayerData[];
  centerPiles: Card[];
  phase: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  cpuActions?: SpeedCpuAction[];
  hint?: SpeedHint;
  config: SpeedConfig;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface GoFishResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface CanastaResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  config: CanastaConfig;
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
export interface PinochleResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  config: PinochleConfig;
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
export interface GolfResponse {
  layout: GolfCard[][];
  stockCount: number;
  waste: Card[];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  hint?: GolfHint;
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
export interface PigsTailResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface SevenCardStudResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
}

// --- Clock Solitaire (クロックソリティア) ---

/** A card in a Clock Solitaire pile with face-up status. */
export interface ClockSolitaireCard {
  card: Card | null;
  faceUp: boolean;
}

/** Full Clock Solitaire game state returned from the API. */
export interface ClockSolitaireResponse {
  piles: ClockSolitaireCard[][];
  faceUpCount: number[];
  phase: number;
  stepCount: number;
  currentCard?: Card;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface DurakResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface FortyThievesResponse {
  tableau: FortyThievesTableauCard[][];
  stockCount: number;
  waste: Card[];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  hint?: FortyThievesHint;
}

/** Source or target zone for a Forty Thieves card move. */
export interface FortyThievesMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
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
export interface BakersDozenResponse {
  tableau: BakersDozenTableauCard[][];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface CalculationResponse {
  foundations: Card[][];
  wastes: Card[][];
  stockCount: number;
  stockTop?: Card;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface FiftyOneResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface YukonResponse {
  tableau: KlondikeTableauCard[][];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface RussianSolitaireResponse {
  tableau: KlondikeTableauCard[][];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  hint?: RussianSolitaireHint;
}

// --- Scorpion (スコーピオン) ---

/** A suggested move hint in Scorpion. */
export interface ScorpionHint {
  fromCol: number;
  cardIndex: number;
  toCol: number;
}

/** API response shape for a Scorpion game. */
export interface ScorpionResponse {
  tableau: KlondikeTableauCard[][];
  stockCount: number;
  completedSuits: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  hint?: ScorpionHint;
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
export interface AccordionResponse {
  piles: AccordionPile[];
  pileCount: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface TrashResponse {
  phase: number;
  current: number;
  players: [TrashPlayerState, TrashPlayerState];
  stockSize: number;
  discardSize: number;
  discardTop?: Card;
  pending?: Card;
  moveCount: number;
  winner: number;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface WhistResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  config: WhistConfig;
  hint?: WhistHint;
}

// --- Poker Squares (ポーカー・スクエア) ---

/** Single cell of the 5x5 Poker Squares board. */
export interface PokerSquaresBoardCell {
  /** Placed card, or `null` when the cell is empty. */
  card: Card | null;
}

/** Poker Squares API response. */
export interface PokerSquaresResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
}

// --- Let It Ride (レット・イット・ライド) ---

/** Let It Ride API response. */
export interface LetItRideResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
}

// --- Red Dog (レッドドッグ) ---

/** Red Dog API response. */
export interface RedDogResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
}

// --- Casino War (カジノウォー) ---

/** Casino War API response. */
export interface CasinoWarResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface PresidentResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface CassinoResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface SpiteAndMaliceResponse {
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
  hint?: SpiteAndMaliceHint;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
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
export interface NertzResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
}

/** Player snapshot for Slapjack. */
export interface SlapjackPlayerData {
  name: string;
  isHuman: boolean;
  stockSize: number;
}

/** Full Slapjack game state returned from the API. */
export interface SlapjackResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
}

/** Player snapshot for Egyptian Ratscrew. */
export interface EgyptianRatscrewPlayerData {
  name: string;
  isHuman: boolean;
  stockSize: number;
}

/** Full Egyptian Ratscrew game state returned from the API. */
export interface EgyptianRatscrewResponse {
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
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
}
