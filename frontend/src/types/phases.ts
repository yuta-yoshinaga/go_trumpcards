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
 *   - Pinochle:    internal/domain/Pinochle.go    (PinochlePhaseBid, PinochlePhaseTrump, PinochlePhaseMeld, PinochlePhasePlay, PinochlePhaseTrickEnd, PinochlePhaseRoundEnd, PinochlePhaseGameEnd)
 *   - Golf:        internal/domain/Golf.go        (GolfPhasePlaying, GolfPhaseGameClear, GolfPhaseGameOver)
 *   - PigsTail:    internal/domain/PigsTail.go    (gameEndFlag bool; local constants PIGTAIL_PHASE_PLAY/END in PigsTailPage.tsx)
 *   - SevenCardStud: internal/domain/SevenCardStud.go (SevenCardStudPhaseInit, SevenCardStudPhaseThirdStreet, ..., SevenCardStudPhaseEnd, SevenCardStudPhaseRebuy)
 *   - ClockSolitaire: internal/domain/ClockSolitaire.go (ClockSolitairePhasePlaying, ClockSolitairePhaseGameClear, ClockSolitairePhaseGameOver)
 *   - LetItRide:  internal/domain/LetItRide.go (LetItRidePhaseBet, LetItRidePhaseFirstDecision, LetItRidePhaseSecondDecision, LetItRidePhaseEnd)
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

/** Badugi phase constants (sync: internal/domain/Badugi.go). */
export const BadugiPhase = {
  INIT: 0,
  DEAL: 1,
  BET: 2,
  DRAW: 3,
  SHOWDOWN: 4,
  END: 5,
} as const;

/** Badugi betting action constants (sync: internal/domain/Badugi.go). */
export const BadugiAction = {
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

/** Canfield phase constants (sync: internal/domain/Canfield.go). */
export const CanfieldPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
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

/** Pitch phase constants (sync: internal/domain/Pitch.go). */
export const PitchPhase = {
  BID: 0,
  PLAY: 1,
  TRICK_END: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Two Ten Jack phase constants (sync: internal/domain/TwoTenJack.go). */
export const TwoTenJackPhase = {
  DECLARE: 0,
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

/** Page One phase constants (sync: internal/domain/PageOne.go). */
export const PageOnePhase = {
  PLAY: 0,
  MUST_DECLARE: 1,
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

/** Tonk phase constants (sync: internal/domain/Tonk.go). */
export const TonkPhase = {
  DRAW: 0,
  DISCARD: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Seven Bridge phase constants (sync: internal/domain/SevenBridge.go). */
export const SevenBridgePhase = {
  DRAW: 0,
  PLAY: 1,
  ROUND_END: 2,
  GAME_END: 3,
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

/** Spiderette phase constants (sync: internal/domain/Spiderette.go). */
export const SpiderettePhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Skat phase constants (sync: internal/domain/Skat.go). */
export const SkatPhase = {
  BID: 0,
  SKAT_PICKUP: 1,
  DISCARD: 2,
  GAME_DECLARATION: 3,
  PLAY: 4,
  TRICK_END: 5,
  ROUND_END: 6,
  GAME_END: 7,
} as const;

/** Skat game type constants (sync: internal/domain/Skat.go). */
export const SkatGameType = {
  NONE: 0,
  SUIT: 1,
  GRAND: 2,
  NULL: 3,
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

/** Mighty phase constants (sync: internal/domain/Mighty.go). */
export const MightyPhase = {
  BID: 0,
  TRUMP_AND_FRIEND: 1,
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

/** Caribbean Stud Poker phase constants (sync: internal/domain/CaribbeanStud.go). */
export const CaribbeanStudPhase = {
  BET: 1,
  ACTION: 2,
  END: 3,
} as const;

/** Texas Hold'em Bonus Poker phase constants (sync: internal/domain/TexasHoldemBonus.go). */
export const TexasHoldemBonusPhase = {
  BET: 1,
  PRE_FLOP: 2,
  FLOP: 3,
  TURN: 4,
  END: 5,
} as const;

/** Ultimate Texas Hold'em phase constants (sync: internal/domain/UltimateTexasHoldem.go). */
export const UltimateTexasHoldemPhase = {
  BET: 1,
  PRE_FLOP: 2,
  FLOP: 3,
  RIVER: 4,
  END: 5,
} as const;

/** Belote phase constants (sync: internal/domain/Belote.go). */
export const BelotePhase = {
  BID_PICK_UP: 0,
  BID_CALL_TRUMP: 1,
  PLAY: 2,
  TRICK_END: 3,
  ROUND_END: 4,
  GAME_END: 5,
} as const;

/** Mississippi Stud phase constants (sync: internal/domain/MississippiStud.go). */
export const MississippiStudPhase = {
  ANTE: 1,
  THIRD_STREET: 2,
  FOURTH_STREET: 3,
  FIFTH_STREET: 4,
  END: 5,
} as const;

/** Doubt phase constants (sync: internal/domain/Doubt.go). */
export const DoubtPhase = {
  PLAY: 0,
  DOUBT: 1,
  END: 2,
} as const;

/** Speed phase constants (sync: internal/domain/Speed.go). */
export const SpeedPhase = {
  PLAY: 0,
  STUCK: 1,
  GAME_END: 2,
} as const;

/** War phase constants (sync: internal/domain/War.go). */
export const WarPhase = {
  REVEAL: 0,
  RESOLVED: 1,
  WAR_BURY: 2,
  GAME_END: 3,
} as const;

/** Go Fish phase constants (sync: internal/domain/GoFish.go). */
export const GoFishPhase = {
  PLAY: 0,
  GAME_END: 1,
} as const;

/** Canasta phase constants (sync: internal/domain/Canasta.go). */
export const CanastaPhase = {
  DRAW: 0,
  MELD: 1,
  DISCARD: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Pinochle phase constants (sync: internal/domain/Pinochle.go). */
export const PinochlePhase = {
  BID: 0,
  TRUMP: 1,
  MELD: 2,
  PLAY: 3,
  TRICK_END: 4,
  ROUND_END: 5,
  GAME_END: 6,
} as const;

/** Golf Solitaire phase constants (sync: internal/domain/Golf.go). */
export const GolfPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Seven Card Stud phase constants (sync: internal/domain/SevenCardStud.go). */
export const SevenCardStudPhase = {
  INIT: 0,
  THIRD_STREET: 1,
  FOURTH_STREET: 2,
  FIFTH_STREET: 3,
  SIXTH_STREET: 4,
  SEVENTH_STREET: 5,
  SHOWDOWN: 6,
  END: 7,
  REBUY: 8,
} as const;

/** Seven Card Stud rebuy phase type constants (sync: internal/domain/SevenCardStud.go). */
export const SevenCardStudRebuyPhaseType = {
  NONE: 0,
  REBUY: 1,
  ADDON: 2,
} as const;

/** Clock Solitaire phase constants (sync: internal/domain/ClockSolitaire.go). */
export const ClockSolitairePhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Pai Gow Poker phase constants (sync: internal/domain/PaiGow.go). */
export const PaiGowPhase = {
  BET: 1,
  SET_HANDS: 2,
  END: 3,
} as const;

/** Forty Thieves phase constants (sync: internal/domain/FortyThieves.go). */
export const FortyThievesPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Baker's Dozen phase constants (sync: internal/domain/BakersDozen.go). */
export const BakersDozenPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Calculation phase constants (sync: internal/domain/Calculation.go). */
export const CalculationPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Fifty-one phase constants (sync: internal/domain/FiftyOne.go). */
export const FiftyOnePhase = {
  PLAY: 0,
  GAME_END: 1,
} as const;

/** Yukon phase constants (sync: internal/domain/Yukon.go). */
export const YukonPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Russian Solitaire phase constants (sync: internal/domain/RussianSolitaire.go). */
export const RussianSolitairePhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Scorpion phase constants (sync: internal/domain/Scorpion.go). */
export const ScorpionPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Accordion phase constants (sync: internal/domain/Accordion.go). */
export const AccordionPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Trash phase constants (sync: internal/domain/Trash.go). */
export const TrashPhase = {
  PLAYER_TURN: 0,
  AWAIT_WILD: 1,
  GAME_OVER: 2,
} as const;

/** Spite & Malice phase constants (sync: internal/domain/SpiteAndMalice.go). */
export const SpiteAndMalicePhase = {
  PLAYING: 0,
  GAME_OVER: 1,
} as const;

/** Whist phase constants (sync: internal/domain/Whist.go). */
export const WhistPhase = {
  PLAY: 0,
  TRICK_END: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Let It Ride phase constants (sync: internal/domain/LetItRide.go). */
/** Poker Squares phase constants (sync: internal/domain/PokerSquares.go). */
export const PokerSquaresPhase = {
  PLAYING: 0,
  COMPLETE: 1,
} as const;

/** Monte Carlo Solitaire phase constants (sync: internal/domain/MonteCarlo.go). */
export const MonteCarloPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

export const LetItRidePhase = {
  BET: 1,
  FIRST_DECISION: 2,
  SECOND_DECISION: 3,
  END: 4,
} as const;

/** Red Dog phase constants (sync: internal/domain/RedDog.go). */
export const RedDogPhase = {
  BET: 1,
  INITIAL_DEALT: 2,
  SPREAD_DECISION: 3,
  PAIR_THIRD: 4,
  END: 5,
} as const;

/** Casino War phase constants (sync: internal/domain/CasinoWar.go). */
export const CasinoWarPhase = {
  BET: 1,
  INITIAL_DEALT: 2,
  TIE_DECISION: 3,
  WAR_DEALT: 4,
  END: 5,
} as const;

/** Dragon Tiger phase constants (sync: internal/domain/DragonTiger.go). */
export const DragonTigerPhase = {
  BET: 1,
  END: 2,
} as const;

/** Dragon Tiger bet-type constants (sync: internal/domain/DragonTiger.go). */
export const DragonTigerBetType = {
  DRAGON: 0,
  TIGER: 1,
  TIE: 2,
} as const;

/** Dragon Tiger Big Road history values (sync: internal/domain/DragonTiger.go). */
export const DragonTigerHistoryResult = {
  DRAGON: 0,
  TIGER: 1,
  TIE: 2,
} as const;

/** Blackjack Switch phase constants (sync: internal/domain/BlackJackSwitch.go). */
export const BlackJackSwitchPhase = {
  BET: 1,
  SWITCH: 2,
  ACTION: 3,
  END: 4,
} as const;

/** Blackjack Switch domain GameResult values (sync: internal/domain/BlackJack.go). */
export const BlackJackSwitchResult = {
  WIN: 1,
  DRAW: 0,
  LOSE: -1,
} as const;

/** Nertz / Pounce phase constants (sync: internal/domain/Nertz.go). */
export const NertzPhase = {
  IDLE: 0,
  PLAYING: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Slapjack phase constants (sync: internal/domain/Slapjack.go). */
export const SlapjackPhase = {
  PLAY: 0,
  GAME_END: 1,
} as const;

/** Slapjack pending CPU action kind (sync: internal/domain/Slapjack.go). */
export const SlapjackPendingKind = {
  NONE: 0,
  STEP: 1,
  SLAP: 2,
} as const;

/** Slapjack last-event kind for UI feedback (sync: internal/domain/Slapjack.go). */
export const SlapjackEventKind = {
  NONE: 0,
  STEP: 1,
  SLAP_CORRECT: 2,
  SLAP_WRONG: 3,
} as const;

/** Egyptian Ratscrew phase constants (sync: internal/domain/EgyptianRatscrew.go). */
export const EgyptianRatscrewPhase = {
  PLAY: 0,
  GAME_END: 1,
} as const;

/** Egyptian Ratscrew pending CPU action kind (sync: internal/domain/EgyptianRatscrew.go). */
export const EgyptianRatscrewPendingKind = {
  NONE: 0,
  STEP: 1,
  SLAP: 2,
} as const;

/** Egyptian Ratscrew last-event kind for UI feedback (sync: internal/domain/EgyptianRatscrew.go). */
export const EgyptianRatscrewEventKind = {
  NONE: 0,
  STEP: 1,
  SLAP_CORRECT: 2,
  SLAP_WRONG: 3,
  CHANCE_WIN: 4,
} as const;

/** Egyptian Ratscrew slap reason for UI feedback (sync: internal/domain/EgyptianRatscrew.go). */
export const EgyptianRatscrewSlapReason = {
  NONE: 0,
  PAIR: 1,
  SANDWICH: 2,
} as const;

/** Crescent Solitaire phase constants (sync: internal/domain/Crescent.go). */
export const CrescentPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;
