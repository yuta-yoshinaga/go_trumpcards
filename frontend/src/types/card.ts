/** Card suit design identifier. */
export type CardDesign = 'SPADE' | 'CLOVER' | 'HEART' | 'DIAMOND' | 'JOKER';

/** A playing card with suit design and numeric value. */
export interface Card {
  design: CardDesign;
  value: number;
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

/** Hold'em round result for a single player. */
export interface HoldemResult {
  playerIdx: number;
  handRank: number;
  handName: string;
  kickers: string;
  bestHand: Card[];
  wonAmount: number;
  mucked: boolean;
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
  equity?: HoldemEquity;
  potOdds?: number;
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
  score: number;
  scoringMode: number;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  hint?: KlondikeHint;
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
  score: number;
  difficulty: number;
  message: string;
  messageCode?: string;
  messageParams?: Record<string, string>;
  hint?: SpiderHint;
}
