export type CardDesign = 'SPADE' | 'CLOVER' | 'HEART' | 'DIAMOND' | 'JOKER'

export interface Card {
  design: CardDesign
  value: number
}

export interface BlackJackPlayer {
  score: number
  cards: Card[]
}

export interface BlackJackResponse {
  dealer: BlackJackPlayer
  player: BlackJackPlayer
  message: string
}

export interface PokerPlayer {
  cards: Card[]
  handName: string
}

export type PokerPhase = 0 | 1 | 2

export interface PokerResponse {
  phase: PokerPhase
  player: PokerPlayer
  dealer: PokerPlayer
  message: string
}

export interface OldMaidPlayerData {
  id: number
  isHuman: boolean
  isFinished: boolean
  cardCount: number
  cards: Card[]
}

export interface CpuAction {
  drawPlayerIdx: number
  drawFromIdx: number
  drawnCard: Card | null
  discardedPairs: number
  discardedCards?: Card[]
}

export interface OldMaidResponse {
  players: OldMaidPlayerData[]
  currentTurn: number
  nextDrawTargetIdx: number
  gameEndFlag: boolean
  hasDrawn: boolean
  lastDrawPlayerIdx: number
  lastDrawFromIdx: number
  lastDrawCard: Card | null
  lastDiscardedPairs: number
  lastDiscardedCards?: Card[]
  cpuActions: CpuAction[]
  message: string
}

export interface DaifugoPlayerData {
  id: number
  isHuman: boolean
  isFinished: boolean
  rank: number
  cardCount: number
  cards: Card[]
}

export interface DaifugoAction {
  playerIdx: number
  playedCards: Card[] | null // null = pass
}

export interface DaifugoResponse {
  players: DaifugoPlayerData[]
  currentTurn: number
  tableCards: Card[]
  lastPlayPlayerIdx: number
  gameEndFlag: boolean
  cpuActions: DaifugoAction[]
  humanAction: DaifugoAction | null
  message: string
}

export interface SevensPlayerData {
  id: number
  isHuman: boolean
  isFinished: boolean
  rank: number
  cardCount: number
  passesUsed: number
  maxPasses: number
  cards: Card[]
}

export interface SevensAction {
  playerIdx: number
  playedCard: Card | null // null = pass
}

export interface SevensResponse {
  players: SevensPlayerData[]
  currentTurn: number
  tableMinVals: number[] // index 0 unused; 1=SPADE, 2=CLOVER, 3=HEART, 4=DIAMOND
  tableMaxVals: number[]
  gameEndFlag: boolean
  cpuActions: SevensAction[]
  humanAction: SevensAction | null
  message: string
}
