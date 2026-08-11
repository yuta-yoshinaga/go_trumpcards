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
 *   - IndianRummy: internal/domain/IndianRummy.go  (IndianRummyPhaseDraw, IndianRummyPhaseDiscard, IndianRummyPhaseRoundEnd, IndianRummyPhaseGameEnd)
 *   - Machiavelli: internal/domain/Machiavelli.go  (MachiavelliPhaseTurn, MachiavelliPhaseRoundEnd, MachiavelliPhaseGameEnd)
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

/** 2-7 Triple Draw phase constants (sync: internal/domain/DeuceToSeven.go). */
export const DeuceToSevenPhase = {
  INIT: 0,
  DEAL: 1,
  BET: 2,
  DRAW: 3,
  SHOWDOWN: 4,
  END: 5,
} as const;

/** 2-7 Triple Draw betting action constants (sync: internal/domain/DeuceToSeven.go). */
export const DeuceToSevenAction = {
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

/** Gong Zhu phase constants (sync: internal/domain/GongZhu.go). */
export const GongZhuPhase = {
  EXPOSE: 0,
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

/** Agnes Sorel phase constants (sync: internal/domain/Agnes.go). */
export const AgnesPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Osmosis phase constants (sync: internal/domain/Osmosis.go). */
export const OsmosisPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Bristol phase constants (sync: internal/domain/Bristol.go). */
export const BristolPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** La Belle Lucie phase values mirroring the Go `LaBelleLuciePhase` enum. */
export const LaBelleLuciePhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Simple Simon phase values mirroring the Go `SimpleSimonPhase` enum. */
export const SimpleSimonPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Double Klondike phase values mirroring the Go `DoubleKlondikePhase` enum. */
export const DoubleKlondikePhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Black Hole phase values mirroring the Go `BlackHolePhase` enum. */
export const BlackHolePhase = {
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

/** Eight Off phase constants (sync: internal/domain/EightOff.go). */
export const EightOffPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Penguin phase constants (sync: internal/domain/Penguin.go). */
export const PenguinPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Seahaven Towers phase constants (sync: internal/domain/SeahavenTowers.go). */
export const SeahavenTowersPhase = {
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

/** Tressette phase constants (sync: internal/domain/Tressette.go). */
export const TressettePhase = {
  PLAY: 0,
  TRICK_END: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Sheepshead phase constants (sync: internal/domain/Sheepshead.go). */
export const SheepsheadPhase = {
  PICK: 0,
  BURY: 1,
  CALL: 2,
  PLAY: 3,
  TRICK_END: 4,
  ROUND_END: 5,
  GAME_END: 6,
} as const;

/** Mus phase constants (sync: internal/domain/Mus.go). */
export const MusPhase = {
  MUS: 0,
  DISCARD: 1,
  GRANDE: 2,
  CHICA: 3,
  PARES: 4,
  JUEGO: 5,
  SHOWDOWN: 6,
  ROUND_END: 7,
  GAME_END: 8,
} as const;

/** Mus betting action constants (sync: internal/domain/Mus.go). */
export const MusBetAction = {
  PASO: 0,
  ENVIDO: 1,
  ORDAGO: 2,
  QUIERO: 3,
  NO_QUIERO: 4,
} as const;

/** Doppelkopf phase constants (sync: internal/domain/Doppelkopf.go). */
export const DoppelkopfPhase = {
  PLAY: 0,
  TRICK_END: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Tute phase constants (sync: internal/domain/Tute.go). */
export const TutePhase = {
  PLAY: 0,
  TRICK_END: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Sueca phase constants (sync: internal/domain/Sueca.go). */
export const SuecaPhase = {
  PLAY: 0,
  TRICK_END: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Klaverjas phase constants (sync: internal/domain/Klaverjas.go). */
export const KlaverjasPhase = {
  PLAY: 0,
  TRICK_END: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Manille phase constants (sync: internal/domain/Manille.go). */
export const ManillePhase = {
  PLAY: 0,
  TRICK_END: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Sedma phase constants (sync: internal/domain/Sedma.go). */
export const SedmaPhase = {
  PLAY: 0,
  TRICK_END: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Mariáš phase constants (sync: internal/domain/Marias.go). */
export const MariasPhase = {
  PLAY: 0,
  TRICK_END: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Tysiąc (Thousand) phase constants (sync: internal/domain/Tysiac.go). */
export const TysiacPhase = {
  BID: 0,
  TALON: 1,
  PLAY: 2,
  TRICK_END: 3,
  ROUND_END: 4,
  GAME_END: 5,
} as const;

/** Calabresella (Terziglio) phase constants (sync: internal/domain/Calabresella.go). */
export const CalabresellaPhase = {
  BID: 0,
  DISCARD: 1,
  PLAY: 2,
  TRICK_END: 3,
  ROUND_END: 4,
  GAME_END: 5,
} as const;

/** Ombre (Hombre) phase constants (sync: internal/domain/Ombre.go). */
export const OmbrePhase = {
  BID: 0,
  PLAY: 1,
  TRICK_END: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Ulti (Ultimo) phase constants (sync: internal/domain/Ulti.go). */
export const UltiPhase = {
  BID: 0,
  DISCARD: 1,
  PLAY: 2,
  TRICK_END: 3,
  ROUND_END: 4,
  GAME_END: 5,
} as const;

/** French Tarot (フレンチタロット) phase constants (sync: internal/domain/FrenchTarot.go). */
export const FrenchTarotPhase = {
  BID: 0,
  CHIEN: 1,
  PLAY: 2,
  TRICK_END: 3,
  ROUND_END: 4,
  GAME_END: 5,
} as const;

/** Scarto (スカルト) phase constants (sync: internal/domain/Scarto.go). */
export const ScartoPhase = {
  SCARTO: 0,
  PLAY: 1,
  TRICK_END: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Königrufen (ケーニッヒルーフェン) phase constants (sync: internal/domain/Koenigrufen.go). */
export const KoenigrufenPhase = {
  BID: 0,
  CALL: 1,
  TALON: 2,
  PLAY: 3,
  TRICK_END: 4,
  ROUND_END: 5,
  GAME_END: 6,
} as const;

/** Cego (チェゴ) phase constants (sync: internal/domain/Cego.go). */
export const CegoPhase = {
  BID: 0,
  CONTRACT: 1,
  EXCHANGE: 2,
  PLAY: 3,
  TRICK_END: 4,
  ROUND_END: 5,
  GAME_END: 6,
} as const;

/** Cinch phase constants (sync: internal/domain/Cinch.go). */
export const CinchPhase = {
  BID: 0,
  NAME_TRUMP: 1,
  PLAY: 2,
  TRICK_END: 3,
  ROUND_END: 4,
  GAME_END: 5,
} as const;

/** Loo (Lanterloo) phase constants (sync: internal/domain/Loo.go). */
export const LooPhase = {
  DECIDE: 0,
  PLAY: 1,
  TRICK_END: 2,
  ROUND_END: 3,
} as const;

/** Basra (Bastra) phase constants (sync: internal/domain/Basra.go). */
export const BasraPhase = {
  PLAY: 0,
  GAME_END: 1,
} as const;

/** Koi-Koi (こいこい) phase constants (sync: internal/domain/KoiKoi.go). */
export const KoiKoiPhase = {
  PLAY: 0,
  KOIKOI_DECISION: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Hachi-Hachi (八八) phase constants (sync: internal/domain/HachiHachi.go). */
export const HachiHachiPhase = {
  PLAY: 0,
  ROUND_END: 1,
  GAME_END: 2,
} as const;

/** Go-Stop (Godori / ゴーストップ) phase constants (sync: internal/domain/GoStop.go). */
export const GoStopPhase = {
  PLAY: 0,
  GO_DECISION: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Tablanet (Tablić) phase constants (sync: internal/domain/Tablanet.go). */
export const TablanetPhase = {
  PLAY: 0,
  GAME_END: 1,
} as const;

/** Knockout Whist phase constants (sync: internal/domain/KnockoutWhist.go). */
export const KnockoutWhistPhase = {
  PLAY: 0,
  TRICK_END: 1,
  ROUND_END: 2,
  GAME_END: 3,
  TRUMP_SELECT: 4,
} as const;

/** Spoil Five phase constants (sync: internal/domain/SpoilFive.go). */
export const SpoilFivePhase = {
  PLAY: 0,
  TRICK_END: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Solo Whist phase constants (sync: internal/domain/SoloWhist.go). */
export const SoloWhistPhase = {
  BID: 0,
  PLAY: 1,
  TRICK_END: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Solo Whist contract constants (sync: internal/domain/SoloWhist.go). */
export const SoloWhistContract = {
  PASS: 0,
  SOLO: 1,
  MISERE: 2,
  ABUNDANCE: 3,
} as const;

/** Auction Forty-Fives phase constants (sync: internal/domain/FortyFives.go). */
export const FortyFivesPhase = {
  BID: 0,
  PLAY: 1,
  TRICK_END: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Twenty-Nine (29) phase constants (sync: internal/domain/TwentyNine.go). */
export const TwentyNinePhase = {
  BID: 0,
  PLAY: 1,
  TRICK_END: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Court Piece (Rang) phase constants (sync: internal/domain/CourtPiece.go). */
export const CourtPiecePhase = {
  TRUMP_DECLARATION: 0,
  PLAY: 1,
  TRICK_END: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Bezique phase constants (sync: internal/domain/Bezique.go). */
export const BeziquePhase = {
  PLAY: 0,
  MELD: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Écarté phase constants (sync: internal/domain/Ecarte.go). */
export const EcartePhase = {
  EXCHANGE: 0,
  PLAY: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Écarté Exchange-phase negotiation sub-step constants (sync: internal/domain/Ecarte.go). */
export const EcarteNegStep = {
  ELDER_DECIDE: 0,
  DEALER_RESPOND: 1,
  ELDER_DISCARD: 2,
  DEALER_DISCARD: 3,
} as const;

/** Three Card Brag phase constants (sync: internal/domain/ThreeCardBrag.go). */
export const ThreeCardBragPhase = {
  BETTING: 0,
  SHOWDOWN: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Teen Patti phase constants (sync: internal/domain/TeenPatti.go). */
export const TeenPattiPhase = {
  BETTING: 0,
  SIDE_SHOW: 1,
  SHOWDOWN: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Spoons phase constants (sync: internal/domain/Spoons.go). */
export const SpoonsPhase = {
  PASS: 0,
  GRAB: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Cuckoo phase constants (sync: internal/domain/Cuckoo.go). */
export const CuckooPhase = {
  TURN: 0,
  REFUSE: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Kemps phase constants (sync: internal/domain/Kemps.go). */
export const KempsPhase = {
  EXCHANGE: 0,
  DECLARE: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Kemps signal type constants (sync: internal/domain/KempsConfig.go). 0=Sound, 1=Blink. */
export const KempsSignal = {
  SOUND: 0,
  BLINK: 1,
} as const;

/** Ganjifa phase constants (sync: internal/domain/Ganjifa.go). There is no Bid phase — trump is auto-declared. */
export const GanjifaPhase = {
  PLAY: 0,
  TRICK_END: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/**
 * Aluette game phases.
 *
 * There is no bidding and no discard, so play starts the moment the cards are
 * dealt — the numbering does not line up with the tarot games it sits beside.
 */
export const AluettePhase = {
  PLAY: 0,
  TRICK_END: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Union of Aluette phase values. */
export type AluettePhaseType = (typeof AluettePhase)[keyof typeof AluettePhase];

/**
 * Minchiate phase constants (sync: internal/domain/Minchiate.go).
 *
 * **Scarto comes first, not a Bid.** The dealer buries the 13 surplus cards
 * before any trick is played; there is no bidding phase in this game.
 */
export const MinchiatePhase = {
  SCARTO: 0,
  PLAY: 1,
  TRICK_END: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/**
 * Tarocchini phase constants (sync: internal/domain/Tarocchini.go).
 *
 * **Scarto comes first, not a Bid.** The dealer buries the 2 surplus cards
 * before any trick is played; there is no bidding phase in this game.
 */
export const TarocchiniPhase = {
  SCARTO: 0,
  PLAY: 1,
  TRICK_END: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Préférence phase constants (sync: internal/domain/Preference.go). */
export const PreferencePhase = {
  BID: 0,
  PLAY: 1,
  TRICK_END: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Préférence contract constants (sync: internal/domain/Preference.go). Outranking order is Pass < Six < Misère < Seven < Eight. */
export const PreferenceContract = {
  PASS: 0,
  SIX: 1,
  MISERE: 2,
  SEVEN: 3,
  EIGHT: 4,
} as const;

/** Vira phase constants (sync: internal/domain/Vira.go). */
export const ViraPhase = {
  BID: 0,
  PLAY: 1,
  TRICK_END: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/**
 * Vira contract constants (sync: internal/domain/Vira.go).
 *
 * Outranking order is Pass < Gask < Solo < Misère < Vira. **Misère sits between
 * Solo and Vira rather than at the bottom** — it is a real contract worth 6,
 * not a way out of bidding.
 */
export const ViraContract = {
  PASS: 0,
  GASK: 1,
  SOLO: 2,
  MISERE: 3,
  VIRA: 4,
} as const;

/** Nap (Napoleon) phase constants (sync: internal/domain/Nap.go). */
export const NapPhase = {
  BID: 0,
  PLAY: 1,
  TRICK_END: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Nap (Napoleon) contract / bid constants (sync: internal/domain/Nap.go). The value equals the declared trick count. */
export const NapContract = {
  PASS: 0,
  TWO: 2,
  THREE: 3,
  FOUR: 4,
  NAP: 5,
} as const;

/** Call Break phase constants (sync: internal/domain/CallBreak.go). */
export const CallBreakPhase = {
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

/** Wizard phase constants (sync: internal/domain/Wizard.go). */
export const WizardPhase = {
  BID: 0,
  PLAY: 1,
  TRICK_END: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Ninety-Nine phase constants (sync: internal/domain/NinetyNine.go). */
export const NinetyNinePhase = {
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

/** Prší phase constants (sync: internal/domain/Prsi.go). */
export const PrsiPhase = {
  PLAY: 0,
  GAME_END: 1,
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

/** Macau phase constants (sync: internal/domain/Macau.go). */
export const MacauPhase = {
  PLAY: 0,
  CHOOSE_SUIT: 1,
  MUST_DECLARE: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Mao phase constants (sync: internal/domain/Mao.go). */
export const MaoPhase = {
  PLAY: 0,
  CHOOSE_SUIT: 1,
  MUST_DECLARE: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Gin Rummy phase constants (sync: internal/domain/GinRummy.go). */
export const GinRummyPhase = {
  DRAW: 0,
  DISCARD: 1,
  LAYOFF: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Indian Rummy phase constants (sync: internal/domain/IndianRummy.go). */
export const IndianRummyPhase = {
  DRAW: 0,
  DISCARD: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Machiavelli phase constants (sync: internal/domain/Machiavelli.go). */
export const MachiavelliPhase = {
  TURN: 0,
  ROUND_END: 1,
  GAME_END: 2,
} as const;

/** Panguingue / Pan phase constants (sync: internal/domain/Pan.go). */
export const PanPhase = {
  DRAW: 0,
  PLAY: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Conquian phase constants (sync: internal/domain/Conquian.go). */
export const ConquianPhase = {
  DRAW: 0,
  MELD: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Chinchón phase constants (sync: internal/domain/Chinchon.go). */
export const ChinchonPhase = {
  DRAW: 0,
  DISCARD: 1,
  LAYOFF: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Three Thirteen phase constants (sync: internal/domain/ThreeThirteen.go). */
export const ThreeThirteenPhase = {
  DRAW: 0,
  DISCARD: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Tonk phase constants (sync: internal/domain/Tonk.go). */
export const TonkPhase = {
  DRAW: 0,
  DISCARD: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Thirty-One phase constants (sync: internal/domain/ThirtyOne.go). */
export const ThirtyOnePhase = {
  DRAW: 0,
  DISCARD: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Yaniv phase constants (sync: internal/domain/Yaniv.go). */
export const YanivPhase = {
  DISCARD: 0,
  DRAW: 1,
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

/** 500 (Five Hundred) phase constants (sync: internal/domain/FiveHundred.go). */
export const FiveHundredPhase = {
  BID: 0,
  KITTY_EXCHANGE: 1,
  PLAY: 2,
  TRICK_END: 3,
  ROUND_END: 4,
  GAME_END: 5,
} as const;

/** 500 (Five Hundred) contract kind constants (sync: internal/domain/FiveHundred.go). */
export const FiveHundredContract = {
  NONE: 0,
  SUIT: 1,
  NO_TRUMP: 2,
  MISERE: 3,
  OPEN_MISERE: 4,
} as const;

/** Bid Whist phase constants (sync: internal/domain/BidWhist.go). */
export const BidWhistPhase = {
  BID: 0,
  TRUMP_DECLARATION: 1,
  KITTY_EXCHANGE: 2,
  PLAY: 3,
  TRICK_END: 4,
  ROUND_END: 5,
  GAME_END: 6,
} as const;

/** Bid Whist bid direction constants (sync: internal/domain/BidWhist.go). */
export const BidWhistDirection = {
  UPTOWN: 0,
  DOWNTOWN: 1,
  NO_TRUMP: 2,
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

/** Four Card Poker phase constants (sync: internal/domain/FourCardPoker.go). */
export const FourCardPokerPhase = {
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

/** Casino Hold'em phase constants (sync: internal/domain/CasinoHoldem.go). */
export const CasinoHoldemPhase = {
  BET: 1,
  FLOP: 2,
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

/** Jass (Schieber) phase constants (sync: internal/domain/Jass.go). */
export const JassPhase = {
  BID_TRUMP: 0,
  BID_PARTNER: 1,
  PLAY: 2,
  TRICK_END: 3,
  ROUND_END: 4,
  GAME_END: 5,
} as const;

/** Watten phase constants (sync: internal/domain/Watten.go). */
export const WattenPhase = {
  DECLARE: 0,
  PLAY: 1,
  RESPOND: 2,
  TRICK_END: 3,
  ROUND_END: 4,
  GAME_END: 5,
} as const;

/** Gaigel phase constants (sync: internal/domain/Gaigel.go). */
export const GaigelPhase = {
  PLAY: 0,
  TRICK_END: 1,
  ROUND_END: 2,
  GAME_END: 3,
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

/** Samba phase constants (sync: internal/domain/Samba.go). */
export const SambaPhase = {
  DRAW: 0,
  MELD: 1,
  DISCARD: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Hand and Foot phase constants (sync: internal/domain/HandAndFoot.go). */
export const HandAndFootPhase = {
  DRAW: 0,
  MELD: 1,
  DISCARD: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Burraco phase constants (sync: internal/domain/Burraco.go). */
export const BurracoPhase = {
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

/** Piquet phase constants (sync: internal/domain/Piquet.go). */
export const PiquetPhase = {
  EXCHANGE: 0,
  DECLARATION: 1,
  PLAY: 2,
  SCORE: 3,
  GAME_END: 4,
} as const;

/** Piquet declaration kind constants (sync: internal/domain/Piquet.go). */
export const PiquetDeclarationKind = {
  POINT: 0,
  SEQUENCE: 1,
  SET: 2,
} as const;

/** Piquet exchange turn constants (sync: internal/domain/Piquet.go). */
export const PiquetExchangeTurn = {
  ELDER: 0,
  YOUNGER: 1,
  DONE: 2,
} as const;

/** Golf Solitaire phase constants (sync: internal/domain/Golf.go). */
export const GolfPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Aces Up phase constants (sync: internal/domain/AcesUp.go). */
export const AcesUpPhase = {
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

/** Five Card Stud phase constants (sync: internal/domain/FiveCardStud.go). */
export const FiveCardStudPhase = {
  INIT: 0,
  SECOND_STREET: 1,
  THIRD_STREET: 2,
  FOURTH_STREET: 3,
  FIFTH_STREET: 4,
  SHOWDOWN: 5,
  END: 6,
  REBUY: 7,
} as const;

/** Five Card Stud rebuy phase type constants (sync: internal/domain/FiveCardStud.go). */
export const FiveCardStudRebuyPhaseType = {
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

/** Chinese Poker phase constants (sync: internal/domain/ChinesePoker.go). */
export const ChinesePokerPhase = {
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

/** Forty and Eight phase constants (sync: internal/domain/FortyAndEight.go). */
export const FortyAndEightPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Sultan of Turkey phase constants (sync: internal/domain/Sultan.go). */
export const SultanPhase = {
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

/** Auld Lang Syne phase constants (sync: internal/domain/AuldLangSyne.go). */
export const AuldLangSynePhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Sir Tommy phase constants (sync: internal/domain/SirTommy.go). */
/** Four Seasons phase constants (sync: internal/domain/FourSeasons.go). */
export const FourSeasonsPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Colorado phase constants (sync: internal/domain/Colorado.go). */
export const ColoradoPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Cribbage Squares phase constants (sync: internal/domain/CribbageSquares.go). */
export const CribbageSquaresPhase = {
  PLAYING: 0,
  COMPLETE: 1,
} as const;

/** Diplomat phase constants (sync: internal/domain/Diplomat.go). */
export const DiplomatPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Royal Cotillion phase constants (sync: internal/domain/RoyalCotillion.go). */
export const RoyalCotillionPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Crazy Quilt phase constants (sync: internal/domain/CrazyQuilt.go). */
export const CrazyQuiltPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

export const SirTommyPhase = {
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

/** Cruel phase constants (sync: internal/domain/Cruel.go). */
export const CruelPhase = {
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

/** Wasp phase constants (sync: internal/domain/Wasp.go). */
export const WaspPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Easthaven phase constants (sync: internal/domain/Easthaven.go). */
export const EasthavenPhase = {
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

/** Catch the Ten phase constants (sync: internal/domain/CatchTen.go). */
export const CatchTenPhase = {
  PLAY: 0,
  TRICK_END: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Briscola phase constants (sync: internal/domain/Briscola.go). */
export const BriscolaPhase = {
  PLAY: 0,
  TRICK_END: 1,
  GAME_END: 2,
} as const;

/** Schnapsen phase constants (sync: internal/domain/Schnapsen.go). */
export const SchnapsenPhase = {
  PLAY: 0,
  TRICK_END: 1,
  GAME_END: 2,
} as const;

/**
 * Sergeant Major (8-5-3) phase constants (sync: internal/domain/SergeantMajor.go).
 *
 * **There is no bidding phase.** The 8, 5 and 3 targets are fixed by seat, so
 * the dealer's only choice is trump — and then the kitty discard.
 */
/**
 * Honeymoon Bridge phase constants (sync: internal/domain/HoneymoonBridge.go).
 *
 * **DRAW is a real trick-playing phase, not a deal animation.** Thirteen
 * no-trump tricks are played that score nothing; each one just hands the
 * winner and then the loser a card from the stock.
 */
export const HoneymoonBridgePhase = {
  DRAW: 0,
  BID: 1,
  PLAY: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

export const SergeantMajorPhase = {
  TRUMP: 0,
  DISCARD: 1,
  PLAY: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/**
 * Hasenpfeffer phase constants (sync: internal/domain/Hasenpfeffer.go).
 *
 * **The declarer takes the blind before play**, so a discard phase of its own
 * sits between the auction and the first trick.
 */
export const HasenpfefferPhase = {
  BID: 0,
  DISCARD: 1,
  PLAY: 2,
  HAND_END: 3,
  GAME_END: 4,
} as const;

/**
 * 3-2-5 phase constants (sync: internal/domain/TeenDoPaanch.go).
 *
 * **There is no bidding phase.** The 3, 2 and 5 targets are assigned at the
 * start of each round, so the only declaration is trump.
 */
export const TeenDoPaanchPhase = {
  TRUMP: 0,
  PLAY: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/**
 * Bhabhi phase constants (sync: internal/domain/Bhabhi.go).
 *
 * **There are no hands.** The whole deck is dealt once and play runs until
 * only one player still holds cards, so there is nothing between PLAY and
 * GAME_END.
 */
export const BhabhiPhase = {
  PLAY: 0,
  GAME_END: 1,
} as const;

/**
 * Mendikot phase constants (sync: internal/domain/Mendikot.go).
 *
 * **There is no trump phase.** Trump is set by whichever card the first player
 * who cannot follow suit chooses to play, so the hand starts in PLAY.
 */
export const MendikotPhase = {
  PLAY: 0,
  HAND_END: 1,
  GAME_END: 2,
} as const;

/**
 * Shelem phase constants (sync: internal/domain/Shelem.go).
 *
 * **What is bid is the score itself**, not a number of tricks, and the winner
 * takes a four-card widow before naming trump — hence a discard phase of its
 * own between bidding and play.
 */
export const ShelemPhase = {
  BID: 0,
  DISCARD: 1,
  PLAY: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/**
 * Hokm phase constants (sync: internal/domain/Hokm.go).
 *
 * **A hand does not play out all thirteen tricks** -- the first partnership to
 * seven takes it and the rest of the cards are never played.
 */
export const HokmPhase = {
  /** The hakem declares trump from their first five cards. */
  TRUMP: 0,
  PLAY: 1,
  HAND_END: 2,
  GAME_END: 3,
} as const;

/**
 * Israeli Whist phase constants (sync: internal/domain/IsraeliWhist.go).
 *
 * **Bidding happens twice**: the auction settles trump and a quota for whoever
 * wins it, then everyone calls their own target separately.
 */
export const IsraeliWhistPhase = {
  AUCTION: 0,
  BID: 1,
  PLAY: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Estimation phase constants (sync: internal/domain/Estimation.go). */
export const EstimationPhase = {
  /** The dealer chooses the trump suit before anyone calls. */
  TRUMP: 0,
  BID: 1,
  PLAY: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/**
 * Estimation call kinds (sync: internal/domain/EstimationPlayer.go).
 *
 * The kind decides how far the score swings: a Dash Call is a flat ±23 and
 * Risk — the highest call at the table — doubles whatever it would otherwise
 * have been worth.
 */
export const EstimationCall = {
  NORMAL: 0,
  DASH: 1,
  RISK: 2,
} as const;

/** Baloot phase constants (sync: internal/domain/Baloot.go). */
export const BalootPhase = {
  /** Before play: each player in turn declares Sun, Hokom, or passes. */
  DECLARE: 0,
  PLAY: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/**
 * Baloot mode constants (sync: internal/domain/Baloot.go).
 *
 * **The mode selects the rank order itself**, not merely whether a trump
 * exists: Sun runs A>10>K>Q>J>9>8>7 with no trump, while Hokom gives the
 * trump suit J>9>A>10>K>Q>8>7.
 */
export const BalootMode = {
  NONE: 0,
  SUN: 1,
  HOKOM: 2,
} as const;

/** Tarabish phase constants (sync: internal/domain/Tarabish.go). */
export const TarabishPhase = {
  /** Before play: each player in turn may take the turned suit as trump. */
  BID: 0,
  PLAY: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Rams phase constants (sync: internal/domain/Rams.go). */
export const RamsPhase = {
  /** Before play: each player chooses to enter the round or drop. */
  DECIDE: 0,
  PLAY: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Reversis phase constants (sync: internal/domain/Reversis.go). */
export const ReversisPhase = {
  PLAY: 0,
  ROUND_END: 1,
  GAME_END: 2,
} as const;

/** Polignac phase constants (sync: internal/domain/Polignac.go). */
export const PolignacPhase = {
  /** Before play: the human may declare capot. */
  DECLARE: 0,
  PLAY: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Slobberhannes phase constants (sync: internal/domain/Slobberhannes.go). */
export const SlobberhannesPhase = {
  PLAY: 0,
  ROUND_END: 1,
  GAME_END: 2,
} as const;

/** German Whist phase constants (sync: internal/domain/GermanWhist.go). */
export const GermanWhistPhase = {
  /** First 13 tricks — played for the face-up card, and they do NOT score. */
  DRAW: 0,
  /** Second 13 tricks — every trick counts. */
  SCORING: 1,
  GAME_END: 2,
} as const;

/** Truco phase constants (sync: internal/domain/Truco.go). */
export const TrucoPhase = {
  PLAY: 0,
  RESPOND: 1,
  TRICK_END: 2,
  HAND_END: 3,
  GAME_END: 4,
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

/** Oicho-Kabu phase constants (sync: internal/domain/OichoKabu.go). */
export const OichoKabuPhase = {
  BET: 1,
  DRAW: 2,
  END: 3,
} as const;

/** Trente et Quarante (Rouge et Noir) phase constants (sync: internal/domain/TrenteEtQuarante.go). Betting immediately deals both rows and resolves. */
export const TrenteEtQuarantePhase = {
  BET: 0,
  RESULT: 1,
} as const;

/** Trente et Quarante bet-type constants (sync: internal/domain/TrenteEtQuaranteConfig.go). */
export const TrenteEtQuaranteBetType = {
  NOIR: 0,
  ROUGE: 1,
  COULEUR: 2,
  INVERSE: 3,
} as const;

/** Trente et Quarante winning-row constants (sync: internal/domain/TrenteEtQuarante.go). A row index, not a color. */
export const TrenteEtQuaranteWinningRow = {
  NONE: -1,
  NOIR: 0,
  ROUGE: 1,
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

/** Oasis Poker phase constants (sync: internal/domain/OasisPoker.go). */
export const OasisPokerPhase = {
  BET: 1,
  EXCHANGE: 2,
  ACTION: 3,
  END: 4,
} as const;

/** Russian Poker phase constants (sync: internal/domain/RussianPoker.go). */
export const RussianPokerPhase = {
  BET: 1,
  ACTION: 2,
  SELECT: 3,
  POST_ACTION: 4,
  FORCE_QUALIFY: 5,
  END: 6,
} as const;

/** Beleaguered Castle phase constants (sync: internal/domain/BeleagueredCastle.go). */
export const BeleagueredCastlePhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Bisley phase constants (sync: internal/domain/Bisley.go). */
export const BisleyPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Napoleon's Square phase constants (sync: internal/domain/NapoleonsSquare.go). */
export const NapoleonsSquarePhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Grandfather's Clock phase constants (sync: internal/domain/GrandfathersClock.go). */
export const GrandfathersClockPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Niu Niu phase constants (sync: internal/domain/NiuNiu.go). */
export const NiuNiuPhase = {
  BET: 1,
  END: 2,
} as const;

/** Bura phase constants (sync: internal/domain/Bura.go). */
export const BuraPhase = {
  PLAY: 0,
  GAME_END: 1,
} as const;

/** Mushi phase constants (sync: internal/domain/Mushi.go). */
export const MushiPhase = {
  PLAY: 0,
  SELECT: 1,
  WILD_SELECT: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

/** Toepen phase constants (sync: internal/domain/Toepen.go). */
export const ToepenPhase = {
  PLAY: 0,
  RESPOND: 1,
  HAND_END: 2,
  GAME_END: 3,
} as const;

/** Chinese Ten phase constants (sync: internal/domain/ChineseTen.go). */
export const ChineseTenPhase = {
  PLAY: 0,
  SELECT: 1,
  GAME_END: 2,
} as const;

/** Skitgubbe phase constants (sync: internal/domain/Skitgubbe.go). */
export const SkitgubbePhase = {
  COLLECT: 0,
  SHED: 1,
  GAME_END: 2,
} as const;

/** Laugh and Lie Down phase constants (sync: internal/domain/LaughAndLieDown.go). */
export const LaughAndLieDownPhase = {
  PLAY: 0,
  GAME_END: 1,
} as const;

/** Sjavs phase constants (sync: internal/domain/Sjavs.go). */
export const SjavsPhase = {
  BID: 0,
  PLAY: 1,
  HAND_END: 2,
  GAME_END: 3,
} as const;

/** Loba phase constants (sync: internal/domain/Loba.go). */
export const LobaPhase = {
  DRAW: 0,
  ACT: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Le Nain Jaune phase constants (sync: internal/domain/NainJaune.go). */
/** Vint phase constants (sync: internal/domain/Vint.go). */
export const VintPhase = {
  BID: 0,
  PLAY: 1,
  HAND_END: 2,
  GAME_END: 3,
} as const;

/** Bid Euchre phase constants (sync: internal/domain/BidEuchre.go). */
export const BidEuchrePhase = {
  BID: 0,
  CHOOSE_TRUMP: 1,
  PLAY: 2,
  HAND_END: 3,
  GAME_END: 4,
} as const;

/** Literature phase constants (sync: internal/domain/Literature.go). */
export const LiteraturePhase = {
  PLAY: 0,
  GAME_END: 1,
} as const;

/** Sheng Ji phase constants (sync: internal/domain/ShengJi.go). */
export const ShengJiPhase = {
  DECLARE: 0,
  KITTY: 1,
  PLAY: 2,
  HAND_END: 3,
  GAME_END: 4,
} as const;

/** Guandan phase constants (sync: internal/domain/Guandan.go). */
export const GuandanPhase = {
  TRIBUTE: 0,
  PLAY: 1,
  HAND_END: 2,
  GAME_END: 3,
} as const;

/** Karnöffel phase constants (sync: internal/domain/Karnoffel.go). */
export const KarnoffelPhase = {
  PLAY: 0,
  HAND_END: 1,
  GAME_END: 2,
} as const;

/** Six-Bid Solo phase constants (sync: internal/domain/SixBidSolo.go). */
export const SixBidSoloPhase = {
  BID: 0,
  DECLARE: 1,
  PLAY: 2,
  HAND_END: 3,
  GAME_END: 4,
} as const;

/** Boston phase constants (sync: internal/domain/Boston.go). */
export const BostonPhase = {
  BID: 0,
  CALL_PARTNER: 1,
  PLAY: 2,
  HAND_END: 3,
  GAME_END: 4,
} as const;

/** Kaiser phase constants (sync: internal/domain/Kaiser.go). */
export const KaiserPhase = {
  BID: 0,
  DISCARD: 1,
  PLAY: 2,
  HAND_END: 3,
  GAME_END: 4,
} as const;

/** Klaberjass phase constants (sync: internal/domain/Klaberjass.go). */
export const KlaberjassPhase = {
  BID_TURN_UP: 0,
  BID_FREE: 1,
  SCHMEISS: 2,
  PLAY: 3,
  HAND_END: 4,
  GAME_END: 5,
} as const;

/** Kille phase constants (sync: internal/domain/Kille.go). */
export const KillePhase = {
  EXCHANGE: 0,
  SHOWDOWN: 1,
  GAME_END: 2,
} as const;

export const NainJaunePhase = {
  PLAY: 0,
  DEAL_END: 1,
  GAME_END: 2,
} as const;

/** Pope Joan phase constants (sync: internal/domain/PopeJoan.go). */
export const PopeJoanPhase = {
  PLAY: 0,
  DEAL_END: 1,
  GAME_END: 2,
} as const;

/** Poch phase constants (sync: internal/domain/Poch.go). */
export const PochPhase = {
  STAKING: 0,
  POCHEN: 1,
  STOPS: 2,
  DEAL_END: 3,
  GAME_END: 4,
} as const;

/** Zwicker phase constants (sync: internal/domain/Zwicker.go). */
export const ZwickerPhase = {
  PLAY: 0,
  ROUND_END: 1,
  GAME_END: 2,
} as const;

/** Desmoche phase constants (sync: internal/domain/Desmoche.go). */
export const DesmochePhase = {
  DRAW: 0,
  ACT: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Trex phase constants (sync: internal/domain/Trex.go). */
export const TrexPhase = {
  CHOOSE: 0,
  PLAY: 1,
  DEAL_END: 2,
  GAME_END: 3,
} as const;

/** Trex contract constants (sync: internal/domain/Trex.go). */
export const TrexContract = {
  KING_OF_HEARTS: 0,
  DIAMONDS: 1,
  QUEENS: 2,
  TRICKS: 3,
  DOMINOES: 4,
  NONE: 5,
} as const;

/** Sette e Mezzo phase constants (sync: internal/domain/SetteEMezzo.go). */
export const SetteEMezzoPhase = {
  BET: 1,
  PLAYER_TURN: 2,
  BANKER_TURN: 3,
  END: 4,
} as const;

/** Pontoon phase constants (sync: internal/domain/Pontoon.go). */
export const PontoonPhase = {
  BET: 1,
  PLAYER_TURN: 2,
  BANKER_TURN: 3,
  END: 4,
} as const;

/** Pontoon hand ranks (sync: internal/domain/Pontoon.go). */
export const PontoonRank = {
  BUST: 0,
  POINTS: 1,
  FIVE_CARD: 2,
  PONTOON: 3,
} as const;

/** Braid phase constants (sync: internal/domain/Braid.go). */
export const BraidPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Terrace phase constants (sync: internal/domain/Terrace.go). */
export const TerracePhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Congress phase constants (sync: internal/domain/Congress.go). */
export const CongressPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** American Toad phase constants (sync: internal/domain/AmericanToad.go). */
export const AmericanToadPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Windmill phase constants (sync: internal/domain/Windmill.go). */
export const WindmillPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Duchess phase constants (sync: internal/domain/Duchess.go). */
export const DuchessPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Miss Milligan phase constants (sync: internal/domain/MissMilligan.go). */
export const MissMilliganPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Streets and Alleys phase constants (sync: internal/domain/StreetsAndAlleys.go). */
export const StreetsAndAlleysPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** King Albert phase constants (sync: internal/domain/KingAlbert.go). */
export const KingAlbertPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Flower Garden phase constants (sync: internal/domain/FlowerGarden.go). */
export const FlowerGardenPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Tarneeb phase constants (sync: internal/domain/Tarneeb.go). */
export const TarneebPhase = {
  BID: 0,
  TRUMP_DECLARATION: 1,
  PLAY: 2,
  TRICK_END: 3,
  ROUND_END: 4,
  GAME_END: 5,
} as const;

/** High Card Flush phase constants (sync: internal/domain/HighCardFlush.go). */
export const HighCardFlushPhase = {
  BET: 1,
  ACTION: 2,
  END: 3,
} as const;

/** Gaps / Montana phase constants (sync: internal/domain/Gaps.go). */
export const GapsPhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

/** Rummy 500 phase constants (sync: internal/domain/Rummy500.go). */
export const Rummy500Phase = {
  DRAW: 0,
  PLAY: 1,
  ROUND_END: 2,
  GAME_END: 3,
} as const;

/** Cuarenta phase constants (sync: internal/domain/Cuarenta.go). */
export const CuarentaPhase = {
  PLAY: 0,
  ROUND_END: 1,
  GAME_END: 2,
} as const;

/** Faro phase constants (sync: internal/domain/Faro.go). */
export const FaroPhase = {
  BETTING: 1,
  TURN: 2,
  CALL: 3,
  ROUND_END: 4,
  GAME_END: 5,
} as const;

/** Open Face Chinese Poker (OFC) phase constants (sync: internal/domain/OpenFaceChinese.go). */
export const OpenFaceChinesePhase = {
  PLACING: 0,
  ROUND_END: 1,
  GAME_END: 2,
} as const;

/** Russian Bank (Crapette) phase values mirroring the Go `RussianBankPhase` enum. */
export const RussianBankPhase = {
  IDLE: 0,
  PLAYING: 1,
  GAME_END: 2,
} as const;

/** Beggar-My-Neighbour phase constants (sync: internal/domain/BeggarMyNeighbour.go). */
export const BeggarMyNeighbourPhase = {
  PLAY: 0,
  PAY_PENALTY: 1,
  COLLECT: 2,
  GAME_END: 3,
} as const;

/** All Fours (Seven Up) phase constants (sync: internal/domain/AllFours.go). */
export const AllFoursPhase = {
  BEG: 0,
  GIFT: 1,
  PLAY: 2,
  TRICK_END: 3,
  ROUND_END: 4,
  GAME_END: 5,
} as const;

/** Guts phase constants (sync: internal/domain/Guts.go). */
export const GutsPhase = {
  DECLARE: 0,
  RESULT: 1,
} as const;

/** Guts declaration constants (sync: internal/domain/Guts.go). 0=out (fold), 1=in (stay). */
export const GutsDeclaration = {
  OUT: 0,
  IN: 1,
} as const;

/** Anaconda phase constants (sync: internal/domain/Anaconda.go). */
export const AnacondaPhase = {
  PASS: 0,
  SET: 1,
  ROLL: 2,
  RESULT: 3,
} as const;

/** Bouillotte phase constants (sync: internal/domain/Bouillotte.go). */
export const BouillottePhase = {
  BETTING: 0,
  RESULT: 1,
} as const;

/** Primero phase constants (sync: internal/domain/Primero.go). */
export const PrimeroPhase = {
  BETTING: 0,
  RESULT: 1,
} as const;

/** Michigan phase constants (sync: internal/domain/Michigan.go). 0=Bet, 1=Play, 2=Result. */
export const MichiganPhase = {
  BET: 0,
  PLAY: 1,
  RESULT: 2,
} as const;

/** Rook (ルーク) phase constants (sync: internal/domain/Rook.go). */
export const RookPhase = {
  BID: 0,
  NEST_EXCHANGE: 1,
  PLAY: 2,
  TRICK_END: 3,
  ROUND_END: 4,
  GAME_END: 5,
} as const;
