/**
 * Game phase enums for all games.
 *
 * These values must stay in sync with the backend domain constants:
 *   - Baccarat:    internal/domain/Baccarat.go    (BaccaratPhaseBet, BaccaratPhaseEnd)
 *   - BlackJack:   internal/domain/BlackJack.go   (BJPhaseBet, BJPhaseDeal, BJPhaseInsurance, BJPhaseAction, BJPhaseEnd)
 *   - CrazyEights: internal/domain/CrazyEights.go (CrazyEightsPhasePlay, CrazyEightsPhaseChooseSuit, CrazyEightsPhaseRoundEnd, CrazyEightsPhaseGameEnd)
 *   - Cribbage:    internal/domain/Cribbage.go    (CribbagePhaseDiscard, CribbagePhaseCut, CribbagePhasePegging, CribbagePhaseShow, CribbagePhaseRoundEnd, CribbagePhaseGameEnd)
 *   - Doubt:       internal/domain/Doubt.go       (DoubtPhasePlay, DoubtPhaseDoubt, DoubtPhaseEnd)
 *   - Bridge:      internal/domain/Bridge.go      (BridgePhaseBid, BridgePhasePlay, BridgePhaseTrickEnd, BridgePhaseRoundEnd, BridgePhaseGameEnd)
 *   - Euchre:      internal/domain/Euchre.go      (EuchrePhasePickUp, EuchrePhaseCallTrump, EuchrePhaseDiscard, EuchrePhasePlay, EuchrePhaseTrickEnd, EuchrePhaseRoundEnd, EuchrePhaseGameEnd)
 *   - FreeCell:    internal/domain/FreeCell.go    (FreeCellPhasePlaying, FreeCellPhaseGameClear, FreeCellPhaseGameOver)
 *   - GinRummy:    internal/domain/GinRummy.go    (GinRummyPhaseDraw, GinRummyPhaseDiscard, GinRummyPhaseLayoff, GinRummyPhaseRoundEnd, GinRummyPhaseGameEnd)
 *   - Hearts:      internal/domain/Hearts.go      (HeartsPhasePass, HeartsPhasePlay, HeartsPhaseTrickEnd, HeartsPhaseRoundEnd, HeartsPhaseGameEnd)
 *   - Holdem:      internal/domain/Holdem.go      (HoldemPhaseInit, HoldemPhasePreFlop, HoldemPhaseFlop, HoldemPhaseTurn, HoldemPhaseRiver, HoldemPhaseShowdown, HoldemPhaseEnd, HoldemPhaseRebuy)
 *   - IndianPoker: internal/domain/IndianPoker.go (IndianPokerPhaseInit, IndianPokerPhaseAnte, IndianPokerPhaseBetting, IndianPokerPhaseShowdown, IndianPokerPhaseEnd)
 *   - Klondike:    internal/domain/Klondike.go    (KlondikePhasePlaying, KlondikePhaseGameClear, KlondikePhaseGameOver)
 *   - Memory:      internal/domain/Memory.go      (MemoryPhaseFlip1, MemoryPhaseFlip2, MemoryPhaseResult, MemoryPhaseGameEnd)
 *   - Napoleon:    internal/domain/Napoleon.go    (NapoleonPhaseBid, NapoleonPhaseTrumpDeclaration, NapoleonPhaseKittyExchange, NapoleonPhasePlay, NapoleonPhaseTrickEnd, NapoleonPhaseRoundEnd, NapoleonPhaseGameEnd)
 *   - OhHell:      internal/domain/OhHell.go      (OhHellPhaseBid, OhHellPhasePlay, OhHellPhaseTrickEnd, OhHellPhaseRoundEnd, OhHellPhaseGameEnd)
 *   - Omaha:       (alias of HoldemPhase)
 *   - Poker:       internal/domain/Poker.go       (PokerPhaseInit, PokerPhaseDeal, PokerPhaseExchange, PokerPhaseSecondBet, PokerPhaseEnd)
 *   - Pyramid:     internal/domain/Pyramid.go     (PyramidPhasePlaying, PyramidPhaseGameClear, PyramidPhaseGameOver)
 *   - ShortDeck:   (alias of HoldemPhase)
 *   - Spades:      internal/domain/Spades.go      (SpadesPhaseBid, SpadesPhasePlay, SpadesPhaseTrickEnd, SpadesPhaseRoundEnd, SpadesPhaseGameEnd)
 *   - Spider:      internal/domain/Spider.go      (SpiderPhasePlaying, SpiderPhaseGameClear, SpiderPhaseGameOver)
 *   - ThreeCard:   internal/domain/ThreeCard.go   (ThreeCardPhaseBet, ThreeCardPhaseAction, ThreeCardPhaseEnd)
 *   - TriPeaks:    internal/domain/TriPeaks.go    (TriPeaksPhasePlaying, TriPeaksPhaseGameClear, TriPeaksPhaseGameOver)
 *   - VideoPoker:  internal/domain/VideoPoker.go  (VideoPokerPhaseBet, VideoPokerPhaseDraw, VideoPokerPhaseResult)
 */

/** BlackJack phase constants (sync: internal/domain/BlackJack.go). */
export const BjPhase = {
  BET: 1,
  DEAL: 2,
  INSURANCE: 3,
  ACTION: 4,
  END: 5,
  EARLY_SURRENDER: 6,
} as const;

/** Poker phase constants (sync: internal/domain/Poker.go). */
export const PokerPhase = {
  INIT: 0,
  DEAL: 1,
  EXCHANGE: 2,
  SECOND_BET: 3,
  END: 4,
} as const;

/** Poker action constants (sync: internal/domain/Poker.go). */
export const PokerAction = {
  FOLD: 0,
  CHECK: 1,
  CALL: 2,
  BET: 3,
  RAISE: 4,
  ALL_IN: 5,
} as const;

/** Texas Hold'em phase constants (sync: internal/domain/Holdem.go). */
export const HoldemPhase = {
  INIT: 0,
  PRE_FLOP: 1,
  FLOP: 2,
  TURN: 3,
  RIVER: 4,
  SHOWDOWN: 5,
  END: 6,
  REBUY: 7,
} as const;

/** Texas Hold'em rebuy phase type constants (sync: internal/domain/Holdem.go). */
export const HoldemRebuyPhaseType = {
  NONE: 0,
  REBUY: 1,
  ADDON: 2,
} as const;

/** Omaha Hold'em phase constants (same as Holdem). */
export const OmahaPhase = HoldemPhase;
/** Omaha Hold'em rebuy phase type constants (same as Holdem). */
export const OmahaRebuyPhaseType = HoldemRebuyPhaseType;

/** Pineapple Poker phase constants (extends Hold'em with DISCARD phase). */
export const PineapplePhase = {
  ...HoldemPhase,
  DISCARD: 8,
} as const;
/** Pineapple Poker rebuy phase type constants (same as Holdem). */
export const PineappleRebuyPhaseType = HoldemRebuyPhaseType;

/** Short Deck Hold'em phase constants (same as Holdem). */
export const ShortDeckPhase = HoldemPhase;
/** Short Deck Hold'em rebuy phase type constants (same as Holdem). */
export const ShortDeckRebuyPhaseType = HoldemRebuyPhaseType;

/** Hearts phase constants (sync: internal/domain/Hearts.go). */
export const HeartsPhase = {
  PASS: 0,
  PLAY: 1,
  TRICK_END: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Memory phase constants (sync: internal/domain/Memory.go). */
export const MemoryPhase = {
  FLIP1: 0,
  FLIP2: 1,
  RESULT: 2,
  GAME_END: 3,
} as const;

/** Klondike phase constants (sync: internal/domain/Klondike.go). */
export const KlondikePhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Klondike scoring mode constants (sync: internal/domain/Klondike.go). */
export const KlondikeScoringMode = {
  NONE: 0,
  VEGAS: 1,
} as const;

/** FreeCell phase constants (sync: internal/domain/FreeCell.go). */
export const FreeCellPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Spades phase constants (sync: internal/domain/Spades.go). */
export const SpadesPhase = {
  BID: 0,
  PLAY: 1,
  TRICK_END: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Oh Hell phase constants (sync: internal/domain/OhHell.go). */
export const OhHellPhase = {
  BID: 0,
  PLAY: 1,
  TRICK_END: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Crazy Eights phase constants (sync: internal/domain/CrazyEights.go). */
export const CrazyEightsPhase = {
  PLAY: 0,
  CHOOSE_SUIT: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Crazy Eights suit constants (sync: internal/domain/Card.go). */
export const CrazyEightsSuit = {
  SPADE: 1,
  CLOVER: 2,
  HEART: 3,
  DIAMOND: 4,
} as const;

/** Gin Rummy phase constants (sync: internal/domain/GinRummy.go). */
export const GinRummyPhase = {
  DRAW: 0,
  DISCARD: 1,
  LAYOFF: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Baccarat bet type constants (sync: internal/domain/Baccarat.go). */
export const BaccaratBetType = {
  PLAYER: 0,
  BANKER: 1,
  TIE: 2,
} as const;

/** Spider phase constants (sync: internal/domain/Spider.go). */
export const SpiderPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Napoleon phase constants (sync: internal/domain/Napoleon.go). */
export const NapoleonPhase = {
  BID: 0,
  TRUMP_DECLARATION: 1,
  KITTY_EXCHANGE: 2,
  PLAY: 3,
  TRICK_END: 4,
  ROUND_END: 5,
  GAME_END: 6,
} as const;

/** Baccarat phase constants (sync: internal/domain/Baccarat.go). */
export const BaccaratPhase = {
  BET: 1,
  END: 2,
} as const;

/** Indian Poker phase constants (sync: internal/domain/IndianPoker.go). */
export const IndianPokerPhase = {
  INIT: 0,
  ANTE: 1,
  BETTING: 2,
  SHOWDOWN: 3,
  END: 4,
} as const;

/** Video Poker phase constants (sync: internal/domain/VideoPoker.go). */
export const VideoPokerPhase = {
  BET: 1,
  DRAW: 2,
  RESULT: 3,
} as const;

/** Pyramid phase constants (sync: internal/domain/Pyramid.go). */
export const PyramidPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** TriPeaks phase constants (sync: internal/domain/TriPeaks.go). */
export const TriPeaksPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Cribbage phase constants (sync: internal/domain/Cribbage.go). */
export const CribbagePhase = {
  DISCARD: 0,
  CUT: 1,
  PEGGING: 2,
  SHOW: 3,
  ROUND_END: 4,
  GAME_END: 5,
} as const;

/** Euchre phase constants (sync: internal/domain/Euchre.go). */
export const EuchrePhase = {
  PICK_UP: 0,
  CALL_TRUMP: 1,
  DISCARD: 2,
  PLAY: 3,
  TRICK_END: 4,
  ROUND_END: 5,
  GAME_END: 6,
} as const;

/** Bridge phase constants (sync: internal/domain/Bridge.go). */
export const BridgePhase = {
  BID: 0,
  PLAY: 1,
  TRICK_END: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Three Card Poker phase constants (sync: internal/domain/ThreeCard.go). */
export const ThreeCardPhase = {
  BET: 1,
  ACTION: 2,
  END: 3,
} as const;

/** Speed phase constants (sync: internal/domain/Speed.go). */
export const SpeedPhase = {
  PLAY: 0,
  STUCK: 1,
  GAME_END: 2,
} as const;
