export type CardDesign = 'SPADE' | 'CLOVER' | 'HEART' | 'DIAMOND' | 'JOKER';

export interface Card {
  design: CardDesign;
  value: number;
}

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

export interface BlackJackPlayer {
  score?: number;
  cards?: Card[];
  chips: number;
}

export type BlackJackPhase = 1 | 2 | 3 | 4 | 5;

export interface BlackJackResponse {
  dealer: BlackJackPlayer;
  player: BlackJackPlayer;
  hands?: BlackJackHand[];
  currentHandIdx: number;
  phase: BlackJackPhase;
  insuranceBet: number;
  insuranceAvailable: boolean;
  message: string;
  hintEnabled: boolean;
  suggestedAction: number;
  deckCount: number;
}

export interface PokerPlayer {
  cards: Card[];
  handRank: number;
  handName: string;
  chips: number;
  bet: number;
}

export type PokerPhase = 0 | 1 | 2 | 3 | 4;

export interface PokerResponse {
  phase: PokerPhase;
  player: PokerPlayer;
  dealer: PokerPlayer;
  message: string;
  pot: number;
  ante: number;
}

export interface OldMaidPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  cardCount: number;
  cards: Card[];
}

export interface CpuAction {
  drawPlayerIdx: number;
  drawFromIdx: number;
  drawnCard: Card | null;
  discardedPairs: number;
  discardedCards?: Card[];
}

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
  cpuHighlightedCardIdx: number;
  removedCard: Card | null;
  mode: number;
  message: string;
}

export interface DaifugoPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  rank: number;
  cardCount: number;
  cards: Card[];
}

export interface DaifugoAction {
  playerIdx: number;
  playedCards: Card[] | null; // null = pass
}

export interface DaifugoConfig {
  jokerCount: number;
  eightCutEnabled: boolean;
  suitLockEnabled: boolean;
  elevenBackEnabled: boolean;
  sequenceEnabled: boolean;
  cardExchangeEnabled: boolean;
  fiveSkipEnabled: boolean;
  sevenPassEnabled: boolean;
  tenDiscardEnabled: boolean;
  spadeThreeEnabled: boolean;
  capitalFallEnabled: boolean;
}

export type DaifugoConfigInput = DaifugoConfig;

export interface DaifugoExchangeAction {
  fromPlayerIdx: number;
  toPlayerIdx: number;
  cards: Card[];
}

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
  pendingAction: 'none' | 'sevenPass' | 'tenDiscard';
  pendingActionTarget: number;
}

export interface SevensPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  rank: number;
  cardCount: number;
  passesUsed: number;
  maxPasses: number;
  cards: Card[];
}

export interface SevensAction {
  playerIdx: number;
  playedCard: Card | null; // null = pass
  targetSuit: number;
  targetValue: number;
}

export interface SevensConfig {
  tunnelEnabled: boolean;
  jokerCount: number;
  cpuStrategy: boolean;
}

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
}

export interface DoubtPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  cardCount: number;
  cards: Card[];
}

export interface DoubtCpuAction {
  playerIdx: number;
  claimedValue: number;
  cardCount: number;
  isBluff: boolean;
}

export interface DoubtDoubtResult {
  doubterIdx: number;
  cardPlayerIdx: number;
  wasLying: boolean;
  loserIdx: number;
  cardCount: number;
  revealedCards: Card[];
}

export interface DoubtConfig {
  doubtWindowSec: number;
  cpuMemoryLevel: number; // 0=Easy, 1=Normal, 2=Hard
}

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
  doubtWindowSec: number;
}

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
}

export interface HoldemCpuAction {
  playerIdx: number;
  action: number;
  amount: number;
}

export interface HoldemResult {
  playerIdx: number;
  handRank: number;
  handName: string;
  bestHand: Card[];
  wonAmount: number;
}

export interface HoldemSidePot {
  amount: number;
  eligiblePlayers: number[];
}

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
  roundResults: HoldemResult[];
  cpuActions: HoldemCpuAction[];
  message: string;
}
