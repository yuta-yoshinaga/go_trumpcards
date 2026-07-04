import type {
  BasraResponse,
  BeziqueResponse,
  BouillotteResponse,
  CalabresellaResponse,
  CallBreakResponse,
  CinchResponse,
  CourtPieceResponse,
  DoppelkopfResponse,
  EcarteResponse,
  EscobaResponse,
  FortyFivesResponse,
  GongZhuResponse,
  GutsResponse,
  HeartsResponse,
  KingResponse,
  KlaverjasResponse,
  KnockoutWhistResponse,
  LooResponse,
  ManilleResponse,
  MariasResponse,
  MusResponse,
  NapResponse,
  PreferenceResponse,
  PrimeroResponse,
  ScoponeResponse,
  SedmaResponse,
  SheepsheadResponse,
  SoloWhistResponse,
  SpadesResponse,
  SpoilFiveResponse,
  SuecaResponse,
  TablanetResponse,
  TeenPattiResponse,
  ThreeCardBragResponse,
  TrenteEtQuaranteResponse,
  TressetteResponse,
  TuteResponse,
  TwentyNineResponse,
  TwoTenJackResponse,
  TysiacResponse,
} from '../types/card';

/** Base Hearts player data for player 0 (human). */
const heartsHumanPlayer = {
  id: 0,
  isHuman: true,
  cardCount: 13,
  cards: [
    { design: 'SPADE' as const, value: 1 },
    { design: 'HEART' as const, value: 11 },
  ],
  roundScore: 0,
  cumulativeScore: 0,
  trickCount: 0,
};

/** Base Hearts state used as the default for {@link makeHeartsState}. */
const baseHeartsState: HeartsResponse = {
  players: [
    heartsHumanPlayer,
    { id: 1, isHuman: false, cardCount: 13, cards: [], roundScore: 3, cumulativeScore: 10, trickCount: 1 },
    { id: 2, isHuman: false, cardCount: 13, cards: [], roundScore: 5, cumulativeScore: 20, trickCount: 2 },
    { id: 3, isHuman: false, cardCount: 13, cards: [], roundScore: 0, cumulativeScore: 5, trickCount: 0 },
  ],
  phase: 1,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  currentTrick: [],
  heartsBroken: false,
  passDirection: 0,
  gameEndFlag: false,
  winnerIdx: -1,
  leadPlayerIdx: 0,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 100, omnibusJD: false },
};

/**
 * Creates a {@link HeartsResponse} with sensible defaults.
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial HeartsResponse fields to override.
 * @returns A complete HeartsResponse suitable for use in tests.
 */
export function makeHeartsState(overrides?: Partial<HeartsResponse>): HeartsResponse {
  return { ...baseHeartsState, ...overrides };
}

/** Base Gong Zhu state used as the default for {@link makeGongZhuState}. */
const baseGongZhuState: GongZhuResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 13,
      cards: [
        { design: 'SPADE' as const, value: 12 },
        { design: 'DIAMOND' as const, value: 11 },
      ],
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    { id: 1, isHuman: false, cardCount: 13, cards: [], roundScore: 0, cumulativeScore: -30, trickCount: 1 },
    { id: 2, isHuman: false, cardCount: 13, cards: [], roundScore: 0, cumulativeScore: -10, trickCount: 2 },
    { id: 3, isHuman: false, cardCount: 13, cards: [], roundScore: 0, cumulativeScore: -50, trickCount: 0 },
  ],
  phase: 1,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  currentTrick: [],
  heartsBroken: false,
  exposed: { pig: false, sheep: false, ace: false, doubler: false },
  exposableIndices: [],
  gameEndFlag: false,
  winnerIdx: -1,
  leadPlayerIdx: 0,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 1000 },
};

/**
 * Creates a {@link GongZhuResponse} with sensible defaults.
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial GongZhuResponse fields to override.
 * @returns A complete GongZhuResponse suitable for use in tests.
 */
export function makeGongZhuState(overrides?: Partial<GongZhuResponse>): GongZhuResponse {
  return { ...baseGongZhuState, ...overrides };
}

/** Base Spades player data used by {@link makeSpadesState}. */
const spadesPlayers: SpadesResponse['players'] = [
  {
    id: 0,
    isHuman: true,
    cardCount: 13,
    cards: [
      { design: 'SPADE' as const, value: 1 },
      { design: 'HEART' as const, value: 11 },
    ],
    bid: 3,
    roundScore: 0,
    cumulativeScore: 0,
    trickCount: 0,
    bags: 0,
  },
  {
    id: 1,
    isHuman: false,
    cardCount: 13,
    cards: [],
    bid: 4,
    roundScore: 3,
    cumulativeScore: 10,
    trickCount: 1,
    bags: 2,
  },
  {
    id: 2,
    isHuman: false,
    cardCount: 13,
    cards: [],
    bid: 3,
    roundScore: 5,
    cumulativeScore: 20,
    trickCount: 2,
    bags: 1,
  },
  {
    id: 3,
    isHuman: false,
    cardCount: 13,
    cards: [],
    bid: 2,
    roundScore: 0,
    cumulativeScore: 5,
    trickCount: 0,
    bags: 0,
  },
];

/** Base Spades state used as the default for {@link makeSpadesState}. */
const baseSpadesState: SpadesResponse = {
  players: spadesPlayers,
  phase: 1,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  bidPlayerIdx: 0,
  currentTrick: [],
  spadesBroken: false,
  gameEndFlag: false,
  winnerIdx: -1,
  leadPlayerIdx: 0,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 500, nilBonus: 100, bagPenaltyThreshold: 10 },
};

/**
 * Creates a {@link SpadesResponse} with sensible defaults.
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial SpadesResponse fields to override.
 * @returns A complete SpadesResponse suitable for use in tests.
 */
export function makeSpadesState(overrides?: Partial<SpadesResponse>): SpadesResponse {
  return { ...baseSpadesState, ...overrides };
}

/** Base Call Break player data used by {@link makeCallBreakState}. */
const callBreakPlayers: CallBreakResponse['players'] = [
  {
    id: 0,
    isHuman: true,
    cardCount: 13,
    cards: [
      { design: 'SPADE' as const, value: 1 },
      { design: 'HEART' as const, value: 11 },
    ],
    bid: 3,
    roundScore: 0,
    cumulativeScore: 0,
    trickCount: 0,
  },
  {
    id: 1,
    isHuman: false,
    cardCount: 13,
    cards: [],
    bid: 4,
    roundScore: 0,
    cumulativeScore: 41,
    trickCount: 1,
  },
  {
    id: 2,
    isHuman: false,
    cardCount: 13,
    cards: [],
    bid: 3,
    roundScore: 0,
    cumulativeScore: 30,
    trickCount: 2,
  },
  {
    id: 3,
    isHuman: false,
    cardCount: 13,
    cards: [],
    bid: 2,
    roundScore: 0,
    cumulativeScore: -20,
    trickCount: 0,
  },
];

/** Base Call Break state for {@link makeCallBreakState}. */
const baseCallBreakState: CallBreakResponse = {
  players: callBreakPlayers,
  phase: 1,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  bidPlayerIdx: 0,
  currentTrick: [],
  spadesBroken: false,
  gameEndFlag: false,
  winnerIdx: -1,
  leadPlayerIdx: 0,
  message: '',
  config: { cpuDifficulty: 1, maxRounds: 5 },
  validPlayIndices: [],
};

/**
 * Creates a {@link CallBreakResponse} with sensible defaults. Scores are int×10
 * (e.g. 41 = 4.1 points).
 */
export function makeCallBreakState(overrides?: Partial<CallBreakResponse>): CallBreakResponse {
  return { ...baseCallBreakState, ...overrides };
}

/** Base Two Ten Jack player data used by {@link makeTwoTenJackState}. */
const twoTenJackPlayers: TwoTenJackResponse['players'] = [
  {
    id: 0,
    isHuman: true,
    cardCount: 13,
    cards: [
      { design: 'SPADE' as const, value: 1 },
      { design: 'HEART' as const, value: 11 },
    ],
    roundScore: 0,
    cumulativeScore: 0,
    trickCount: 0,
    capturedPoints: 0,
  },
  {
    id: 1,
    isHuman: false,
    cardCount: 13,
    cards: [],
    roundScore: 0,
    cumulativeScore: 0,
    trickCount: 0,
    capturedPoints: 0,
  },
  {
    id: 2,
    isHuman: false,
    cardCount: 13,
    cards: [],
    roundScore: 0,
    cumulativeScore: 0,
    trickCount: 0,
    capturedPoints: 0,
  },
  {
    id: 3,
    isHuman: false,
    cardCount: 13,
    cards: [],
    roundScore: 0,
    cumulativeScore: 0,
    trickCount: 0,
    capturedPoints: 0,
  },
];

/** Base Two Ten Jack state used as the default for {@link makeTwoTenJackState}. */
const baseTwoTenJackState: TwoTenJackResponse = {
  players: twoTenJackPlayers,
  phase: 1,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  declarerIdx: 0,
  trumpSuit: 1,
  currentTrick: [],
  gameEndFlag: false,
  winnerTeam: -1,
  leadPlayerIdx: 0,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 50 },
};

/**
 * Creates a {@link TwoTenJackResponse} with sensible defaults for tests.
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial TwoTenJackResponse fields to override.
 * @returns A complete TwoTenJackResponse suitable for use in tests.
 */
export function makeTwoTenJackState(overrides?: Partial<TwoTenJackResponse>): TwoTenJackResponse {
  return { ...baseTwoTenJackState, ...overrides };
}

/** Base Tressette game state (play phase, human's turn) for tests. */
const baseTressetteState: TressetteResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 10,
      cards: [
        { design: 'SPADE' as const, value: 3 },
        { design: 'DIAMOND' as const, value: 13 },
      ],
      trickCount: 0,
      teamId: 0,
    },
    { id: 1, isHuman: false, cardCount: 10, cards: [], trickCount: 0, teamId: 1 },
    { id: 2, isHuman: false, cardCount: 10, cards: [], trickCount: 0, teamId: 0 },
    { id: 3, isHuman: false, cardCount: 10, cards: [], trickCount: 0, teamId: 1 },
  ],
  phase: 0,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  currentTrick: [],
  leadPlayerIdx: 0,
  teamScores: [0, 0],
  teamRoundThirds: [0, 0],
  playableIndices: [0, 1],
  gameEndFlag: false,
  winnerTeam: -1,
  message: '',
  config: { cpuDifficulty: 1, targetPoints: 21 },
};

/**
 * Creates a Tressette game state for testing with sensible defaults.
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial TressetteResponse fields to override.
 * @returns A complete TressetteResponse suitable for use in tests.
 */
export function makeTressetteState(overrides?: Partial<TressetteResponse>): TressetteResponse {
  return { ...baseTressetteState, ...overrides };
}

/** Base Sheepshead state used as the default for {@link makeSheepsheadState}. Defaults to a human Play turn. */
const baseSheepsheadState: SheepsheadResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 6,
      cards: [
        { design: 'SPADE' as const, value: 1 },
        { design: 'DIAMOND' as const, value: 13 },
      ],
      trickCount: 0,
      chips: 100,
    },
    { id: 1, isHuman: false, cardCount: 6, cards: [], trickCount: 0, chips: 100 },
    { id: 2, isHuman: false, cardCount: 6, cards: [], trickCount: 0, chips: 100 },
    { id: 3, isHuman: false, cardCount: 6, cards: [], trickCount: 0, chips: 100 },
    { id: 4, isHuman: false, cardCount: 6, cards: [], trickCount: 0, chips: 100 },
  ],
  phase: 3,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  leadPlayerIdx: 0,
  dealerIdx: 4,
  currentTrick: [],
  blindCount: 0,
  buried: [],
  pickerIdx: 0,
  partnerIdx: -1,
  calledSuit: 1,
  partnerRevealed: false,
  passCount: 0,
  callableSuits: [],
  playableIndices: [0, 1],
  roundPickerPoints: 0,
  roundMultiplier: 1,
  roundPickerWon: false,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
  config: { cpuDifficulty: 1, baseChips: 1, startChips: 100, targetChips: 200 },
};

/**
 * Creates a {@link SheepsheadResponse} with sensible defaults (a human Play turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial SheepsheadResponse fields to override.
 * @returns A complete SheepsheadResponse suitable for use in tests.
 */
export function makeSheepsheadState(overrides?: Partial<SheepsheadResponse>): SheepsheadResponse {
  return { ...baseSheepsheadState, ...overrides };
}

/** Base Mus state used as the default for {@link makeMusState}. Defaults to a human Grande bet turn. */
const baseMusState: MusResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 4,
      cards: [
        { design: 'SPADE' as const, value: 1 },
        { design: 'CLOVER' as const, value: 12 },
        { design: 'HEART' as const, value: 12 },
        { design: 'DIAMOND' as const, value: 7 },
      ],
      teamScore: 5,
    },
    { id: 1, isHuman: false, cardCount: 4, cards: [], teamScore: 3 },
    { id: 2, isHuman: false, cardCount: 4, cards: [], teamScore: 5 },
    { id: 3, isHuman: false, cardCount: 4, cards: [], teamScore: 3 },
  ],
  phase: 2,
  roundNumber: 1,
  manoIdx: 0,
  betTeam: -1,
  pendingStake: 0,
  lastBettorTeam: -1,
  musTurn: 0,
  discardTurn: 0,
  musCycle: 0,
  amarrakos: [5, 3],
  results: [
    { kind: 0, stake: 0, team: -1 },
    { kind: 1, stake: 0, team: -1 },
    { kind: 2, stake: 0, team: -1 },
    { kind: 3, stake: 0, team: -1 },
  ],
  gameEndFlag: false,
  winnerTeam: -1,
  humanTeam: 0,
  isHumanTurn: true,
  canPaso: true,
  canEnvido: true,
  canOrdago: true,
  canQuiero: false,
  canNoQuiero: false,
  message: '',
  config: { cpuDifficulty: 1, targetAmarrakos: 40 },
};

/**
 * Creates a {@link MusResponse} with sensible defaults (a human Grande bet turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial MusResponse fields to override.
 * @returns A complete MusResponse suitable for use in tests.
 */
export function makeMusState(overrides?: Partial<MusResponse>): MusResponse {
  return { ...baseMusState, ...overrides };
}

/** Base Doppelkopf state used as the default for {@link makeDoppelkopfState}. Defaults to a human Play turn. */
const baseDoppelkopfState: DoppelkopfResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 12,
      cards: [
        { design: 'HEART' as const, value: 10 },
        { design: 'DIAMOND' as const, value: 13 },
      ],
      trickCount: 0,
      chips: 20,
      isRe: false,
    },
    { id: 1, isHuman: false, cardCount: 12, cards: [], trickCount: 0, chips: 20, isRe: false },
    { id: 2, isHuman: false, cardCount: 12, cards: [], trickCount: 0, chips: 20, isRe: false },
    { id: 3, isHuman: false, cardCount: 12, cards: [], trickCount: 0, chips: 20, isRe: false },
  ],
  phase: 0,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  leadPlayerIdx: 0,
  dealerIdx: 3,
  currentTrick: [],
  reTeam: [false, false, false, false],
  soloRe: false,
  teamsRevealed: false,
  reAnnounced: false,
  kontraAnnounced: false,
  canAnnounce: true,
  youAreRe: true,
  playableIndices: [0, 1],
  roundRePoints: 0,
  roundReWon: false,
  roundGamePoints: 0,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
  config: { cpuDifficulty: 1, baseChips: 2, startChips: 20, targetChips: 40 },
};

/**
 * Creates a {@link DoppelkopfResponse} with sensible defaults (a human Play turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial DoppelkopfResponse fields to override.
 * @returns A complete DoppelkopfResponse suitable for use in tests.
 */
export function makeDoppelkopfState(overrides?: Partial<DoppelkopfResponse>): DoppelkopfResponse {
  return { ...baseDoppelkopfState, ...overrides };
}

/** Base Tute state used as the default for {@link makeTuteState}. Defaults to a human Play turn. */
const baseTuteState: TuteResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 10,
      cards: [
        { design: 'HEART' as const, value: 12 },
        { design: 'HEART' as const, value: 13 },
        { design: 'SPADE' as const, value: 1 },
      ],
      trickCount: 0,
      teamScore: 0,
    },
    { id: 1, isHuman: false, cardCount: 10, cards: [], trickCount: 0, teamScore: 0 },
    { id: 2, isHuman: false, cardCount: 10, cards: [], trickCount: 0, teamScore: 0 },
    { id: 3, isHuman: false, cardCount: 10, cards: [], trickCount: 0, teamScore: 0 },
  ],
  phase: 0,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  leadPlayerIdx: 0,
  dealerIdx: 3,
  trumpSuit: 4,
  currentTrick: [],
  declaredSuits: [false, false, false, false, false],
  teamScores: [0, 0],
  roundTeamPoints: [0, 0],
  canDeclareMarriage: false,
  canDeclareTute: false,
  playableIndices: [0, 1, 2],
  gameEndFlag: false,
  winnerTeam: -1,
  isHumanTurn: true,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1, targetPoints: 121 },
};

/**
 * Creates a {@link TuteResponse} with sensible defaults (a human Play turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial TuteResponse fields to override.
 * @returns A complete TuteResponse suitable for use in tests.
 */
export function makeTuteState(overrides?: Partial<TuteResponse>): TuteResponse {
  return { ...baseTuteState, ...overrides };
}

/** Base Sueca state used as the default for {@link makeSuecaState}. Defaults to a human Play turn. */
const baseSuecaState: SuecaResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 10,
      cards: [
        { design: 'HEART' as const, value: 12 },
        { design: 'HEART' as const, value: 13 },
        { design: 'SPADE' as const, value: 1 },
      ],
      trickCount: 0,
      teamGamePoints: 0,
    },
    { id: 1, isHuman: false, cardCount: 10, cards: [], trickCount: 0, teamGamePoints: 0 },
    { id: 2, isHuman: false, cardCount: 10, cards: [], trickCount: 0, teamGamePoints: 0 },
    { id: 3, isHuman: false, cardCount: 10, cards: [], trickCount: 0, teamGamePoints: 0 },
  ],
  phase: 0,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  leadPlayerIdx: 0,
  dealerIdx: 3,
  trumpSuit: 4,
  currentTrick: [],
  teamGamePoints: [0, 0],
  roundCardPoints: [0, 0],
  roundWinnerTeam: -1,
  roundGamePoints: 0,
  playableIndices: [0, 1, 2],
  gameEndFlag: false,
  winnerTeam: -1,
  isHumanTurn: true,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1, targetGamePoints: 4 },
};

/**
 * Creates a {@link SuecaResponse} with sensible defaults (a human Play turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial SuecaResponse fields to override.
 * @returns A complete SuecaResponse suitable for use in tests.
 */
export function makeSuecaState(overrides?: Partial<SuecaResponse>): SuecaResponse {
  return { ...baseSuecaState, ...overrides };
}

/** Base Klaverjas state used as the default for {@link makeKlaverjasState}. Defaults to a human Play turn. */
const baseKlaverjasState: KlaverjasResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 8,
      cards: [
        { design: 'HEART' as const, value: 12 },
        { design: 'HEART' as const, value: 13 },
        { design: 'SPADE' as const, value: 1 },
      ],
      trickCount: 0,
      teamScore: 0,
    },
    { id: 1, isHuman: false, cardCount: 8, cards: [], trickCount: 0, teamScore: 0 },
    { id: 2, isHuman: false, cardCount: 8, cards: [], trickCount: 0, teamScore: 0 },
    { id: 3, isHuman: false, cardCount: 8, cards: [], trickCount: 0, teamScore: 0 },
  ],
  phase: 0,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  leadPlayerIdx: 0,
  dealerIdx: 3,
  trumpSuit: 4,
  currentTrick: [],
  teamScores: [0, 0],
  roundCardPoints: [0, 0],
  roundRoem: [0, 0],
  playableIndices: [0, 1, 2],
  gameEndFlag: false,
  winnerTeam: -1,
  isHumanTurn: true,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1, targetPoints: 1501 },
};

/**
 * Creates a {@link KlaverjasResponse} with sensible defaults (a human Play turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial KlaverjasResponse fields to override.
 * @returns A complete KlaverjasResponse suitable for use in tests.
 */
export function makeKlaverjasState(overrides?: Partial<KlaverjasResponse>): KlaverjasResponse {
  return { ...baseKlaverjasState, ...overrides };
}

/** Base Manille state used as the default for {@link makeManilleState}. Defaults to a human Play turn. */
const baseManilleState: ManilleResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 8,
      cards: [
        { design: 'HEART' as const, value: 12 },
        { design: 'HEART' as const, value: 13 },
        { design: 'SPADE' as const, value: 1 },
      ],
      trickCount: 0,
      teamScore: 0,
    },
    { id: 1, isHuman: false, cardCount: 8, cards: [], trickCount: 0, teamScore: 0 },
    { id: 2, isHuman: false, cardCount: 8, cards: [], trickCount: 0, teamScore: 0 },
    { id: 3, isHuman: false, cardCount: 8, cards: [], trickCount: 0, teamScore: 0 },
  ],
  phase: 0,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  leadPlayerIdx: 0,
  dealerIdx: 3,
  trumpSuit: 4,
  currentTrick: [],
  teamScores: [0, 0],
  roundCardPoints: [0, 0],
  playableIndices: [0, 1, 2],
  gameEndFlag: false,
  winnerTeam: -1,
  isHumanTurn: true,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1, targetPoints: 101 },
};

/**
 * Creates a {@link ManilleResponse} with sensible defaults (a human Play turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial ManilleResponse fields to override.
 * @returns A complete ManilleResponse suitable for use in tests.
 */
export function makeManilleState(overrides?: Partial<ManilleResponse>): ManilleResponse {
  return { ...baseManilleState, ...overrides };
}

/** Base Sedma state used as the default for {@link makeSedmaState}. Defaults to a human Play turn. No trump suit. */
const baseSedmaState: SedmaResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 8,
      cards: [
        { design: 'HEART' as const, value: 12 },
        { design: 'HEART' as const, value: 13 },
        { design: 'SPADE' as const, value: 1 },
      ],
      trickCount: 0,
      teamScore: 0,
    },
    { id: 1, isHuman: false, cardCount: 8, cards: [], trickCount: 0, teamScore: 0 },
    { id: 2, isHuman: false, cardCount: 8, cards: [], trickCount: 0, teamScore: 0 },
    { id: 3, isHuman: false, cardCount: 8, cards: [], trickCount: 0, teamScore: 0 },
  ],
  phase: 0,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  leadPlayerIdx: 0,
  dealerIdx: 3,
  currentTrick: [],
  teamScores: [0, 0],
  roundCardPoints: [0, 0],
  playableIndices: [0, 1, 2],
  gameEndFlag: false,
  winnerTeam: -1,
  isHumanTurn: true,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1, targetPoints: 101 },
};

/**
 * Creates a {@link SedmaResponse} with sensible defaults (a human Play turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial SedmaResponse fields to override.
 * @returns A complete SedmaResponse suitable for use in tests.
 */
export function makeSedmaState(overrides?: Partial<SedmaResponse>): SedmaResponse {
  return { ...baseSedmaState, ...overrides };
}

/** Base Mariáš state used as the default for {@link makeMariasState}. Defaults to a human Play turn. */
const baseMariasState: MariasResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 10,
      cards: [
        { design: 'HEART' as const, value: 12 },
        { design: 'HEART' as const, value: 13 },
        { design: 'SPADE' as const, value: 1 },
      ],
      trickCount: 0,
      score: 0,
      isSoloist: true,
    },
    { id: 1, isHuman: false, cardCount: 10, cards: [], trickCount: 0, score: 0, isSoloist: false },
    { id: 2, isHuman: false, cardCount: 10, cards: [], trickCount: 0, score: 0, isSoloist: false },
  ],
  phase: 0,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  leadPlayerIdx: 0,
  dealerIdx: 2,
  soloistIdx: 0,
  trumpSuit: 3,
  currentTrick: [],
  playerScores: [0, 0, 0],
  roundCardPoints: [0, 0, 0],
  roundMarriage: [0, 0, 0],
  lastTrickWinner: -1,
  playableIndices: [0, 1, 2],
  gameEndFlag: false,
  winnerPlayer: -1,
  isHumanTurn: true,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1, targetPoints: 10 },
};

/**
 * Creates a {@link MariasResponse} with sensible defaults (a human Play turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial MariasResponse fields to override.
 * @returns A complete MariasResponse suitable for use in tests.
 */
export function makeMariasState(overrides?: Partial<MariasResponse>): MariasResponse {
  return { ...baseMariasState, ...overrides };
}

/** Base King state used as the default for {@link makeKingState}. Defaults to a human Play turn. */
const baseKingState: KingResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 13,
      cards: [
        { design: 'HEART' as const, value: 12 },
        { design: 'HEART' as const, value: 13 },
        { design: 'SPADE' as const, value: 1 },
      ],
      trickCount: 0,
      totalScore: 0,
    },
    { id: 1, isHuman: false, cardCount: 13, cards: [], trickCount: 0, totalScore: 0 },
    { id: 2, isHuman: false, cardCount: 13, cards: [], trickCount: 0, totalScore: 0 },
    { id: 3, isHuman: false, cardCount: 13, cards: [], trickCount: 0, totalScore: 0 },
  ],
  phase: 'play',
  dealNumber: 0,
  totalDeals: 7,
  dealerIdx: 0,
  currentTurn: 0,
  currentContract: 0,
  trumpSuit: -1,
  trickNumber: 1,
  currentTrick: [],
  lastTrick: [],
  lastTrickWinner: -1,
  usedContracts: [false, false, false, false, false, false, false],
  playableIndices: [0, 1, 2],
  gameEndFlag: false,
  config: { cpuDifficulty: 1 },
  roundWinners: [],
  lastDealDetail: null,
  isHumanTurn: true,
  hint: null,
  message: '',
};

/**
 * Creates a {@link KingResponse} with sensible defaults (a human Play turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial KingResponse fields to override.
 * @returns A complete KingResponse suitable for use in tests.
 */
export function makeKingState(overrides?: Partial<KingResponse>): KingResponse {
  return { ...baseKingState, ...overrides };
}

/** Base Tysiąc state used as the default for {@link makeTysiacState}. Defaults to a human Play turn. */
const baseTysiacState: TysiacResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 7,
      cards: [
        { design: 'HEART' as const, value: 12 },
        { design: 'HEART' as const, value: 13 },
        { design: 'SPADE' as const, value: 1 },
      ],
      trickCount: 0,
      score: 0,
      isDeclarer: true,
    },
    { id: 1, isHuman: false, cardCount: 7, cards: [], trickCount: 0, score: 0, isDeclarer: false },
    { id: 2, isHuman: false, cardCount: 7, cards: [], trickCount: 0, score: 0, isDeclarer: false },
  ],
  phase: 2,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  leadPlayerIdx: 0,
  dealerIdx: 2,
  forehandIdx: 0,
  declarerIdx: 0,
  contract: 100,
  currentBid: 100,
  trumpSuit: 3,
  currentTrick: [],
  playerScores: [0, 0, 0],
  roundCardPoints: [0, 0, 0],
  roundMarriage: [0, 0, 0],
  lastTrickWinner: -1,
  playableIndices: [0, 1, 2],
  gameEndFlag: false,
  winnerPlayer: -1,
  isHumanTurn: true,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1, targetPoints: 1000 },
};

/**
 * Creates a {@link TysiacResponse} with sensible defaults (a human Play turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial TysiacResponse fields to override.
 * @returns A complete TysiacResponse suitable for use in tests.
 */
export function makeTysiacState(overrides?: Partial<TysiacResponse>): TysiacResponse {
  return { ...baseTysiacState, ...overrides };
}

/** Base Calabresella state used as the default for {@link makeCalabresellaState}. Defaults to a human Play turn. */
const baseCalabresellaState: CalabresellaResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 12,
      cards: [
        { design: 'HEART' as const, value: 12 },
        { design: 'HEART' as const, value: 13 },
        { design: 'SPADE' as const, value: 1 },
      ],
      trickCount: 0,
      score: 0,
      isSoloist: true,
      roundThirds: 0,
    },
    { id: 1, isHuman: false, cardCount: 12, cards: [], trickCount: 0, score: 0, isSoloist: false, roundThirds: 0 },
    { id: 2, isHuman: false, cardCount: 12, cards: [], trickCount: 0, score: 0, isSoloist: false, roundThirds: 0 },
  ],
  phase: 2,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  currentBidderIdx: 0,
  leadPlayerIdx: 0,
  dealerIdx: 2,
  forehandIdx: 0,
  soloistIdx: 0,
  winningBid: 1,
  currentTrick: [],
  playerScores: [0, 0, 0],
  roundThirds: [0, 0, 0],
  lastTrickWinner: -1,
  playableIndices: [0, 1, 2],
  gameEndFlag: false,
  winnerPlayer: -1,
  isHumanTurn: true,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1, targetPoints: 21 },
};

/**
 * Creates a {@link CalabresellaResponse} with sensible defaults (a human Play turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial CalabresellaResponse fields to override.
 * @returns A complete CalabresellaResponse suitable for use in tests.
 */
export function makeCalabresellaState(overrides?: Partial<CalabresellaResponse>): CalabresellaResponse {
  return { ...baseCalabresellaState, ...overrides };
}

/** Base Cinch state used as the default for {@link makeCinchState}. Defaults to a human Play turn (trump = ♠). */
const baseCinchState: CinchResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 9,
      cards: [
        { design: 'HEART' as const, value: 12 },
        { design: 'HEART' as const, value: 13 },
        { design: 'SPADE' as const, value: 1 },
      ],
      trickCount: 0,
      bid: 6,
      totalScore: 0,
    },
    { id: 1, isHuman: false, cardCount: 9, cards: [], trickCount: 0, bid: 0, totalScore: 0 },
    { id: 2, isHuman: false, cardCount: 9, cards: [], trickCount: 0, bid: 0, totalScore: 0 },
    { id: 3, isHuman: false, cardCount: 9, cards: [], trickCount: 0, bid: 0, totalScore: 0 },
  ],
  phase: 2,
  roundNumber: 1,
  trickNumber: 1,
  totalTricks: 9,
  dealerIdx: 3,
  currentTurn: 0,
  bidPlayerIdx: 0,
  currentBid: 6,
  bidWinnerIdx: 0,
  trumpSuit: 1,
  currentTrick: [],
  lastTrick: [],
  lastTrickWinner: -1,
  playableIndices: [0, 1, 2],
  gameEndFlag: false,
  winnerIdx: -1,
  roundWinners: [],
  lastDealDetail: null,
  isHumanTurn: true,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 21 },
};

/**
 * Creates a {@link CinchResponse} with sensible defaults (a human Play turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial CinchResponse fields to override.
 * @returns A complete CinchResponse suitable for use in tests.
 */
export function makeCinchState(overrides?: Partial<CinchResponse>): CinchResponse {
  return { ...baseCinchState, ...overrides };
}

/** Base Loo state used as the default for {@link makeLooState}. Defaults to a human Play turn (trump = ♠). */
const baseLooState: LooResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 5,
      cards: [
        { design: 'HEART' as const, value: 12 },
        { design: 'HEART' as const, value: 13 },
        { design: 'SPADE' as const, value: 1 },
      ],
      trickCount: 0,
      playing: true,
      chips: -3,
    },
    { id: 1, isHuman: false, cardCount: 5, cards: [], trickCount: 0, playing: true, chips: -3 },
    { id: 2, isHuman: false, cardCount: 5, cards: [], trickCount: 0, playing: true, chips: -3 },
    { id: 3, isHuman: false, cardCount: 5, cards: [], trickCount: 0, playing: false, chips: -3 },
  ],
  phase: 1,
  roundNumber: 1,
  trickNumber: 1,
  totalTricks: 5,
  dealerIdx: 3,
  currentTurn: 0,
  decidePlayerIdx: 0,
  trumpSuit: 1,
  turnUp: { design: 'SPADE' as const, value: 7 },
  pot: 12,
  potStart: 12,
  currentTrick: [],
  lastTrick: [],
  lastTrickWinner: -1,
  playableIndices: [0, 1, 2],
  gameEndFlag: false,
  lastDealDetail: null,
  isHumanTurn: true,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1, ante: 3 },
};

/**
 * Creates a {@link LooResponse} with sensible defaults (a human Play turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial LooResponse fields to override.
 * @returns A complete LooResponse suitable for use in tests.
 */
export function makeLooState(overrides?: Partial<LooResponse>): LooResponse {
  return { ...baseLooState, ...overrides };
}

/** Base Basra state used as the default for {@link makeBasraState}. Defaults to a human Play turn. */
const baseBasraState: BasraResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 4,
      cards: [
        { design: 'HEART' as const, value: 5 },
        { design: 'SPADE' as const, value: 11 },
        { design: 'DIAMOND' as const, value: 7 },
        { design: 'CLOVER' as const, value: 3 },
      ],
      capturedCount: 0,
      basraCount: 0,
      score: 0,
    },
    { id: 1, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, basraCount: 0, score: 0 },
    { id: 2, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, basraCount: 0, score: 0 },
    { id: 3, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, basraCount: 0, score: 0 },
  ],
  phase: 0,
  roundNumber: 1,
  currentTurn: 0,
  tableCards: [
    { design: 'SPADE' as const, value: 5 },
    { design: 'HEART' as const, value: 9 },
  ],
  lastCaptureIdx: -1,
  remainingDeck: 32,
  playableIndices: [0, 1, 2, 3],
  captureOptions: { 0: [0] },
  winners: [],
  gameEndFlag: false,
  lastDealDetail: null,
  isHumanTurn: true,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1 },
};

/**
 * Creates a {@link BasraResponse} with sensible defaults (a human Play turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial BasraResponse fields to override.
 * @returns A complete BasraResponse suitable for use in tests.
 */
export function makeBasraState(overrides?: Partial<BasraResponse>): BasraResponse {
  return { ...baseBasraState, ...overrides };
}

/** Base Tablanet state used as the default for {@link makeTablanetState}. Defaults to a human Play turn. */
const baseTablanetState: TablanetResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 4,
      cards: [
        { design: 'HEART' as const, value: 5 },
        { design: 'SPADE' as const, value: 11 },
        { design: 'DIAMOND' as const, value: 7 },
        { design: 'CLOVER' as const, value: 3 },
      ],
      capturedCount: 0,
      tablaCount: 0,
      score: 0,
    },
    { id: 1, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, tablaCount: 0, score: 0 },
    { id: 2, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, tablaCount: 0, score: 0 },
    { id: 3, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, tablaCount: 0, score: 0 },
  ],
  phase: 0,
  roundNumber: 1,
  currentTurn: 0,
  tableCards: [
    { design: 'SPADE' as const, value: 5 },
    { design: 'HEART' as const, value: 9 },
  ],
  lastCaptureIdx: -1,
  remainingDeck: 32,
  playableIndices: [0, 1, 2, 3],
  captureOptions: { 0: [0] },
  winners: [],
  gameEndFlag: false,
  lastDealDetail: null,
  isHumanTurn: true,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1 },
};

/**
 * Creates a {@link TablanetResponse} with sensible defaults (a human Play turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial TablanetResponse fields to override.
 * @returns A complete TablanetResponse suitable for use in tests.
 */
export function makeTablanetState(overrides?: Partial<TablanetResponse>): TablanetResponse {
  return { ...baseTablanetState, ...overrides };
}

/** Base Trente et Quarante state used as the default for {@link makeTrenteEtQuaranteState}. Defaults to the bet phase. */
const baseTrenteEtQuaranteState: TrenteEtQuaranteResponse = {
  phase: 0, // TrenteEtQuarantePhase.BET
  roundNumber: 0,
  chips: 1000,
  currentBet: 0, // Noir
  stake: 0,
  noirRow: [],
  rougeRow: [],
  noirTotal: 0,
  rougeTotal: 0,
  winningRow: -1,
  firstCardRed: false,
  refait: false,
  result: 0,
  payout: 0,
  remainingDeck: 312,
  gameEndFlag: false,
  config: { defaultBet: 0 },
  message: '',
};

/**
 * Creates a {@link TrenteEtQuaranteResponse} with sensible defaults (the bet phase).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial TrenteEtQuaranteResponse fields to override.
 * @returns A complete TrenteEtQuaranteResponse suitable for use in tests.
 */
export function makeTrenteEtQuaranteState(overrides?: Partial<TrenteEtQuaranteResponse>): TrenteEtQuaranteResponse {
  return { ...baseTrenteEtQuaranteState, ...overrides };
}

/** Base Knockout Whist state used as the default for {@link makeKnockoutWhistState}. Defaults to a human Play turn (round 1, 7-card hand). */
const baseKnockoutWhistState: KnockoutWhistResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 7,
      cards: [
        { design: 'HEART' as const, value: 12 },
        { design: 'HEART' as const, value: 13 },
        { design: 'SPADE' as const, value: 1 },
      ],
      trickCount: 0,
      eliminated: false,
      dogbones: 1,
      roundTricks: 0,
    },
    { id: 1, isHuman: false, cardCount: 7, cards: [], trickCount: 0, eliminated: false, dogbones: 1, roundTricks: 0 },
    { id: 2, isHuman: false, cardCount: 7, cards: [], trickCount: 0, eliminated: false, dogbones: 1, roundTricks: 0 },
    { id: 3, isHuman: false, cardCount: 7, cards: [], trickCount: 0, eliminated: false, dogbones: 1, roundTricks: 0 },
  ],
  phase: 0,
  roundNumber: 1,
  handSize: 7,
  trickNumber: 1,
  currentPlayerIdx: 0,
  leadPlayerIdx: 0,
  dealerIdx: 3,
  trumpSuit: 3,
  roundWinnerIdx: -1,
  currentTrick: [],
  activeCount: 4,
  playableIndices: [0, 1, 2],
  gameEndFlag: false,
  winnerPlayer: -1,
  isHumanTurn: true,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1 },
};

/**
 * Creates a {@link KnockoutWhistResponse} with sensible defaults (a human Play turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial KnockoutWhistResponse fields to override.
 * @returns A complete KnockoutWhistResponse suitable for use in tests.
 */
export function makeKnockoutWhistState(overrides?: Partial<KnockoutWhistResponse>): KnockoutWhistResponse {
  return { ...baseKnockoutWhistState, ...overrides };
}

/** Base Spoil Five state used as the default for {@link makeSpoilFiveState}. Defaults to a human Play turn (round 1, 5-card hand, 5 players). */
const baseSpoilFiveState: SpoilFiveResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 5,
      cards: [
        { design: 'HEART' as const, value: 12 },
        { design: 'HEART' as const, value: 13 },
        { design: 'SPADE' as const, value: 1 },
      ],
      trickCount: 0,
      score: 0,
      roundTricks: 0,
    },
    { id: 1, isHuman: false, cardCount: 5, cards: [], trickCount: 0, score: 0, roundTricks: 0 },
    { id: 2, isHuman: false, cardCount: 5, cards: [], trickCount: 0, score: 0, roundTricks: 0 },
    { id: 3, isHuman: false, cardCount: 5, cards: [], trickCount: 0, score: 0, roundTricks: 0 },
    { id: 4, isHuman: false, cardCount: 5, cards: [], trickCount: 0, score: 0, roundTricks: 0 },
  ],
  phase: 0,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  leadPlayerIdx: 0,
  dealerIdx: 4,
  trumpSuit: 3,
  pot: 5,
  roundWinnerIdx: -1,
  currentTrick: [],
  playableIndices: [0, 1, 2],
  gameEndFlag: false,
  winnerPlayer: -1,
  isHumanTurn: true,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1, targetPoints: 30 },
};

/**
 * Creates a {@link SpoilFiveResponse} with sensible defaults (a human Play turn,
 * 5 players, a 5-card hand, and a pot). Any field can be overridden via the
 * `overrides` parameter.
 *
 * @param overrides - Partial SpoilFiveResponse fields to override.
 * @returns A complete SpoilFiveResponse suitable for use in tests.
 */
export function makeSpoilFiveState(overrides?: Partial<SpoilFiveResponse>): SpoilFiveResponse {
  return { ...baseSpoilFiveState, ...overrides };
}

/** Base Solo Whist state used as the default for {@link makeSoloWhistState}. Defaults to a human Bid turn. */
const baseSoloWhistState: SoloWhistResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 13,
      cards: [
        { design: 'HEART' as const, value: 12 },
        { design: 'HEART' as const, value: 13 },
        { design: 'SPADE' as const, value: 1 },
      ],
      trickCount: 0,
      score: 0,
      isDeclarer: false,
    },
    { id: 1, isHuman: false, cardCount: 13, cards: [], trickCount: 0, score: 0, isDeclarer: false },
    { id: 2, isHuman: false, cardCount: 13, cards: [], trickCount: 0, score: 0, isDeclarer: false },
    { id: 3, isHuman: false, cardCount: 13, cards: [], trickCount: 0, score: 0, isDeclarer: false },
  ],
  phase: 0,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  leadPlayerIdx: 0,
  dealerIdx: 3,
  declarerIdx: -1,
  contract: 0,
  trumpSuit: 0,
  bids: [0, 0, 0, 0],
  currentTrick: [],
  playerScores: [0, 0, 0, 0],
  roundTricks: [0, 0, 0, 0],
  playableIndices: [],
  gameEndFlag: false,
  winnerPlayer: -1,
  isHumanTurn: false,
  isHumanBidTurn: true,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1, targetPoints: 21 },
};

/**
 * Creates a {@link SoloWhistResponse} with sensible defaults (a human Bid turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial SoloWhistResponse fields to override.
 * @returns A complete SoloWhistResponse suitable for use in tests.
 */
export function makeSoloWhistState(overrides?: Partial<SoloWhistResponse>): SoloWhistResponse {
  return { ...baseSoloWhistState, ...overrides };
}

/** Base Auction Forty-Fives state used as the default for {@link makeFortyFivesState}. A 4-player 2-team bidding trick-taker; defaults to a human Bid turn. */
const baseFortyFivesState: FortyFivesResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 5,
      cards: [
        { design: 'HEART' as const, value: 12 },
        { design: 'HEART' as const, value: 13 },
        { design: 'SPADE' as const, value: 1 },
      ],
      trickCount: 0,
      teamScore: 0,
      isDeclarer: false,
    },
    { id: 1, isHuman: false, cardCount: 5, cards: [], trickCount: 0, teamScore: 0, isDeclarer: false },
    { id: 2, isHuman: false, cardCount: 5, cards: [], trickCount: 0, teamScore: 0, isDeclarer: false },
    { id: 3, isHuman: false, cardCount: 5, cards: [], trickCount: 0, teamScore: 0, isDeclarer: false },
  ],
  phase: 0,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  leadPlayerIdx: 0,
  dealerIdx: 3,
  declarerIdx: -1,
  contract: 0,
  trumpSuit: 0,
  bids: [0, 0, 0, 0],
  currentTrick: [],
  teamScores: [0, 0],
  roundTeamPoints: [0, 0],
  playableIndices: [],
  gameEndFlag: false,
  winnerTeam: -1,
  isHumanTurn: false,
  isHumanBidTurn: true,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1, targetPoints: 45 },
};

/**
 * Creates a {@link FortyFivesResponse} with sensible defaults (a human Bid turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial FortyFivesResponse fields to override.
 * @returns A complete FortyFivesResponse suitable for use in tests.
 */
export function makeFortyFivesState(overrides?: Partial<FortyFivesResponse>): FortyFivesResponse {
  return { ...baseFortyFivesState, ...overrides };
}

/** Base Twenty-Nine (29) state used as the default for {@link makeTwentyNineState}. A 4-player 2-team hidden-trump bidding trick-taker; defaults to a human Bid turn. */
const baseTwentyNineState: TwentyNineResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 8,
      cards: [
        { design: 'HEART' as const, value: 12 },
        { design: 'HEART' as const, value: 13 },
        { design: 'SPADE' as const, value: 1 },
      ],
      trickCount: 0,
      teamScore: 0,
      isDeclarer: false,
    },
    { id: 1, isHuman: false, cardCount: 8, cards: [], trickCount: 0, teamScore: 0, isDeclarer: false },
    { id: 2, isHuman: false, cardCount: 8, cards: [], trickCount: 0, teamScore: 0, isDeclarer: false },
    { id: 3, isHuman: false, cardCount: 8, cards: [], trickCount: 0, teamScore: 0, isDeclarer: false },
  ],
  phase: 0,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  leadPlayerIdx: 0,
  dealerIdx: 3,
  declarerIdx: -1,
  contract: 0,
  trumpSuit: 0,
  trumpRevealed: false,
  bids: [0, 0, 0, 0],
  currentTrick: [],
  teamScores: [0, 0],
  roundTeamPoints: [0, 0],
  playableIndices: [],
  gameEndFlag: false,
  winnerTeam: -1,
  isHumanTurn: false,
  isHumanBidTurn: true,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1, targetPoints: 6 },
};

/**
 * Creates a {@link TwentyNineResponse} with sensible defaults (a human Bid turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial TwentyNineResponse fields to override.
 * @returns A complete TwentyNineResponse suitable for use in tests.
 */
export function makeTwentyNineState(overrides?: Partial<TwentyNineResponse>): TwentyNineResponse {
  return { ...baseTwentyNineState, ...overrides };
}

/** Base Court Piece (Rang) state used as the default for {@link makeCourtPieceState}. A 4-player 2-team called-trump trick-taker; defaults to a human TrumpDeclaration turn (the human is the caller). */
const baseCourtPieceState: CourtPieceResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      team: 0,
      cardCount: 13,
      cards: [
        { design: 'HEART' as const, value: 12 },
        { design: 'HEART' as const, value: 13 },
        { design: 'SPADE' as const, value: 1 },
      ],
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    { id: 1, isHuman: false, team: 1, cardCount: 13, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
    { id: 2, isHuman: false, team: 0, cardCount: 13, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
    { id: 3, isHuman: false, team: 1, cardCount: 13, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
  ],
  phase: 0,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  callerIdx: 0,
  trumpSuit: 0,
  currentTrick: [],
  teamScores: [0, 0],
  consecutiveWins: 0,
  lastWinnerTeam: -1,
  lastRoundCourt: false,
  gameEndFlag: false,
  winnerTeam: -1,
  leadPlayerIdx: 0,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 7 },
};

/**
 * Creates a {@link CourtPieceResponse} with sensible defaults (a human
 * TrumpDeclaration turn). Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial CourtPieceResponse fields to override.
 * @returns A complete CourtPieceResponse suitable for use in tests.
 */
export function makeCourtPieceState(overrides?: Partial<CourtPieceResponse>): CourtPieceResponse {
  return { ...baseCourtPieceState, ...overrides };
}

/** Base Bezique state used as the default for {@link makeBeziqueState}. A 2-player melding trick-taker; defaults to a human Play turn with stock remaining (phase 1). */
const baseBeziqueState: BeziqueResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 9,
      cards: [
        { design: 'SPADE' as const, value: 12 },
        { design: 'DIAMOND' as const, value: 11 },
        { design: 'HEART' as const, value: 1 },
      ],
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    { id: 1, isHuman: false, cardCount: 9, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
  ],
  dealPoints: [0, 0],
  dealMeldPoints: [0, 0],
  matchScore: [0, 0],
  phase: 0,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  leadPlayerIdx: 0,
  dealerIdx: 1,
  trumpSuit: 3,
  trumpCard: { design: 'HEART' as const, value: 7 },
  currentTrick: [],
  stockRemaining: 30,
  isEndgame: false,
  availableMelds: [],
  gameEndFlag: false,
  winnerIdx: -1,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1, targetScore: 1000 },
};

/**
 * Creates a {@link BeziqueResponse} with sensible defaults (a human Play turn,
 * phase 1, stock remaining). Any field can be overridden via the `overrides`
 * parameter.
 *
 * @param overrides - Partial BeziqueResponse fields to override.
 * @returns A complete BeziqueResponse suitable for use in tests.
 */
export function makeBeziqueState(overrides?: Partial<BeziqueResponse>): BeziqueResponse {
  return { ...baseBeziqueState, ...overrides };
}

/** Base Écarté state used as the default for {@link makeEcarteState}. A 2-player French trick game; defaults to a human Exchange turn at the ElderDecide sub-step (phase 0). */
const baseEcarteState: EcarteResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 5,
      cards: [
        { design: 'SPADE' as const, value: 13 },
        { design: 'DIAMOND' as const, value: 11 },
        { design: 'HEART' as const, value: 1 },
      ],
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    { id: 1, isHuman: false, cardCount: 5, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
  ],
  dealPoints: [0, 0],
  matchScore: [0, 0],
  phase: 0,
  negStep: 0,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  dealerIdx: 1,
  elderIdx: 0,
  leadPlayerIdx: 0,
  trumpSuit: 3,
  trumpCard: { design: 'HEART' as const, value: 13 },
  currentTrick: [],
  stockRemaining: 22,
  refusalByDealer: false,
  validPlays: [],
  gameEndFlag: false,
  winnerIdx: -1,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1, targetScore: 5 },
};

/**
 * Creates an {@link EcarteResponse} with sensible defaults (a human Exchange
 * turn at the ElderDecide sub-step, phase 0). Any field can be overridden via
 * the `overrides` parameter.
 *
 * @param overrides - Partial EcarteResponse fields to override.
 * @returns A complete EcarteResponse suitable for use in tests.
 */
export function makeEcarteState(overrides?: Partial<EcarteResponse>): EcarteResponse {
  return { ...baseEcarteState, ...overrides };
}

/** Base Three Card Brag state used as the default for {@link makeThreeCardBragState}. A 4-player British vying game; defaults to a human Betting turn (phase 0). */
const baseThreeCardBragState: ThreeCardBragResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      chips: 100,
      seen: false,
      folded: false,
      out: false,
      roundBet: 1,
      cardCount: 3,
      cards: [
        { design: 'SPADE' as const, value: 13 },
        { design: 'SPADE' as const, value: 12 },
        { design: 'SPADE' as const, value: 11 },
      ],
    },
    { id: 1, isHuman: false, chips: 100, seen: false, folded: false, out: false, roundBet: 1, cardCount: 3, cards: [] },
    { id: 2, isHuman: false, chips: 100, seen: false, folded: false, out: false, roundBet: 1, cardCount: 3, cards: [] },
    { id: 3, isHuman: false, chips: 100, seen: false, folded: false, out: false, roundBet: 1, cardCount: 3, cards: [] },
  ],
  pot: 4,
  stake: 1,
  phase: 0,
  roundNumber: 1,
  dealerIdx: 3,
  currentPlayerIdx: 0,
  roundWinnerIdx: -1,
  matchWinnerIdx: -1,
  isShowdown: false,
  canShow: false,
  gameEndFlag: false,
  isHumanTurn: true,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1, ante: 1, startingChips: 100 },
};

/**
 * Creates a {@link ThreeCardBragResponse} with sensible defaults (a human
 * Betting turn, phase 0). Any field can be overridden via the `overrides`
 * parameter.
 *
 * @param overrides - Partial ThreeCardBragResponse fields to override.
 * @returns A complete ThreeCardBragResponse suitable for use in tests.
 */
export function makeThreeCardBragState(overrides?: Partial<ThreeCardBragResponse>): ThreeCardBragResponse {
  return { ...baseThreeCardBragState, ...overrides };
}

/** Base Teen Patti state used as the default for {@link makeTeenPattiState}. The Indian variant of Three Card Brag; a 4-player vying game that defaults to a human Betting turn (phase 0). */
const baseTeenPattiState: TeenPattiResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      chips: 100,
      seen: false,
      folded: false,
      out: false,
      roundBet: 1,
      cardCount: 3,
      cards: [
        { design: 'SPADE' as const, value: 13 },
        { design: 'SPADE' as const, value: 12 },
        { design: 'SPADE' as const, value: 11 },
      ],
    },
    { id: 1, isHuman: false, chips: 100, seen: false, folded: false, out: false, roundBet: 1, cardCount: 3, cards: [] },
    { id: 2, isHuman: false, chips: 100, seen: false, folded: false, out: false, roundBet: 1, cardCount: 3, cards: [] },
    { id: 3, isHuman: false, chips: 100, seen: false, folded: false, out: false, roundBet: 1, cardCount: 3, cards: [] },
  ],
  pot: 4,
  stake: 1,
  phase: 0,
  roundNumber: 1,
  dealerIdx: 3,
  currentPlayerIdx: 0,
  roundWinnerIdx: -1,
  matchWinnerIdx: -1,
  isShowdown: false,
  canShow: false,
  canRequestSideShow: false,
  sideShowRequester: -1,
  sideShowTarget: -1,
  gameEndFlag: false,
  isHumanTurn: true,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1, ante: 1, startingChips: 100 },
};

/**
 * Creates a {@link TeenPattiResponse} with sensible defaults (a human Betting
 * turn, phase 0). Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial TeenPattiResponse fields to override.
 * @returns A complete TeenPattiResponse suitable for use in tests.
 */
export function makeTeenPattiState(overrides?: Partial<TeenPattiResponse>): TeenPattiResponse {
  return { ...baseTeenPattiState, ...overrides };
}

/** Base Préférence state used as the default for {@link makePreferenceState}. A 3-player Russian bidding trick-taker; defaults to a human Bid turn. */
const basePreferenceState: PreferenceResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 10,
      cards: [
        { design: 'HEART' as const, value: 12 },
        { design: 'HEART' as const, value: 13 },
        { design: 'SPADE' as const, value: 1 },
      ],
      trickCount: 0,
      score: 0,
      isDeclarer: false,
    },
    { id: 1, isHuman: false, cardCount: 10, cards: [], trickCount: 0, score: 0, isDeclarer: false },
    { id: 2, isHuman: false, cardCount: 10, cards: [], trickCount: 0, score: 0, isDeclarer: false },
  ],
  phase: 0,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  leadPlayerIdx: 0,
  dealerIdx: 2,
  declarerIdx: -1,
  contract: 0,
  trumpSuit: 0,
  bids: [0, 0, 0],
  currentTrick: [],
  playerScores: [0, 0, 0],
  roundTricks: [0, 0, 0],
  playableIndices: [],
  gameEndFlag: false,
  winnerPlayer: -1,
  isHumanTurn: false,
  isHumanBidTurn: true,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1, targetPoints: 30 },
};

/**
 * Creates a {@link PreferenceResponse} with sensible defaults (a human Bid turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial PreferenceResponse fields to override.
 * @returns A complete PreferenceResponse suitable for use in tests.
 */
export function makePreferenceState(overrides?: Partial<PreferenceResponse>): PreferenceResponse {
  return { ...basePreferenceState, ...overrides };
}

/** Base Nap (Napoleon) state used as the default for {@link makeNapState}. Defaults to a human Bid turn. */
const baseNapState: NapResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 5,
      cards: [
        { design: 'HEART' as const, value: 12 },
        { design: 'HEART' as const, value: 13 },
        { design: 'SPADE' as const, value: 1 },
      ],
      trickCount: 0,
      score: 0,
      isDeclarer: false,
    },
    { id: 1, isHuman: false, cardCount: 5, cards: [], trickCount: 0, score: 0, isDeclarer: false },
    { id: 2, isHuman: false, cardCount: 5, cards: [], trickCount: 0, score: 0, isDeclarer: false },
    { id: 3, isHuman: false, cardCount: 5, cards: [], trickCount: 0, score: 0, isDeclarer: false },
  ],
  phase: 0,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  leadPlayerIdx: 0,
  dealerIdx: 3,
  declarerIdx: -1,
  contract: 0,
  trumpSuit: 0,
  bids: [0, 0, 0, 0],
  currentTrick: [],
  playerScores: [0, 0, 0, 0],
  roundTricks: [0, 0, 0, 0],
  playableIndices: [],
  gameEndFlag: false,
  winnerPlayer: -1,
  isHumanTurn: false,
  isHumanBidTurn: true,
  hint: null,
  message: '',
  config: { cpuDifficulty: 1, targetPoints: 20 },
};

/**
 * Creates a {@link NapResponse} with sensible defaults (a human Bid turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial NapResponse fields to override.
 * @returns A complete NapResponse suitable for use in tests.
 */
export function makeNapState(overrides?: Partial<NapResponse>): NapResponse {
  return { ...baseNapState, ...overrides };
}

/** Base Scopone state used as the default for {@link makeScoponeState}. Defaults to a human play turn. */
const baseScoponeState: ScoponeResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      team: 0,
      handCount: 3,
      cards: [
        { design: 'SPADE' as const, value: 3 },
        { design: 'HEART' as const, value: 5 },
        { design: 'DIAMOND' as const, value: 7 },
      ],
      capturedCount: 0,
      scopaCount: 0,
    },
    { id: 1, isHuman: false, team: 1, handCount: 3, cards: [], capturedCount: 0, scopaCount: 0 },
    { id: 2, isHuman: false, team: 0, handCount: 3, cards: [], capturedCount: 0, scopaCount: 0 },
    { id: 3, isHuman: false, team: 1, handCount: 3, cards: [], capturedCount: 0, scopaCount: 0 },
  ],
  tableCards: [
    { design: 'SPADE' as const, value: 2 },
    { design: 'HEART' as const, value: 5 },
  ],
  phase: 'playerTurn',
  roundNumber: 1,
  currentTurn: 0,
  dealerIdx: 3,
  teamScores: [0, 0],
  lastCaptureIdx: -1,
  winnerTeam: -1,
  gameEndFlag: false,
  isHumanTurn: true,
  handCaptures: [[[1]], [], []],
  lastRoundDetail: null,
  config: { cpuDifficulty: 1, targetScore: 11 },
  message: '',
};

/**
 * Creates a {@link ScoponeResponse} with sensible defaults (a human play turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial ScoponeResponse fields to override.
 * @returns A complete ScoponeResponse suitable for use in tests.
 */
export function makeScoponeState(overrides?: Partial<ScoponeResponse>): ScoponeResponse {
  return { ...baseScoponeState, ...overrides };
}

/** Base Escoba state used as the default for {@link makeEscobaState}. Defaults to a human play turn. */
const baseEscobaState: EscobaResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      handCount: 3,
      cards: [
        { design: 'SPADE' as const, value: 7 },
        { design: 'HEART' as const, value: 5 },
        { design: 'DIAMOND' as const, value: 11 },
      ],
      capturedCount: 0,
      escobaCount: 0,
      score: 0,
    },
    { id: 1, isHuman: false, handCount: 3, cards: [], capturedCount: 0, escobaCount: 0, score: 0 },
    { id: 2, isHuman: false, handCount: 3, cards: [], capturedCount: 0, escobaCount: 0, score: 0 },
    { id: 3, isHuman: false, handCount: 3, cards: [], capturedCount: 0, escobaCount: 0, score: 0 },
  ],
  tableCards: [
    { design: 'SPADE' as const, value: 4 },
    { design: 'HEART' as const, value: 4 },
  ],
  phase: 'playerTurn',
  roundNumber: 1,
  currentTurn: 0,
  dealerIdx: 3,
  stockRemaining: 28,
  lastCaptureIdx: -1,
  winnerIdx: -1,
  gameEndFlag: false,
  isHumanTurn: true,
  handCaptures: [[[0, 1]], [], []],
  lastRoundDetail: null,
  config: { cpuDifficulty: 1, targetScore: 10 },
  message: '',
};

/**
 * Creates an {@link EscobaResponse} with sensible defaults (a human play turn).
 * Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial EscobaResponse fields to override.
 * @returns A complete EscobaResponse suitable for use in tests.
 */
export function makeEscobaState(overrides?: Partial<EscobaResponse>): EscobaResponse {
  return { ...baseEscobaState, ...overrides };
}

/** Base Guts state used as the default for {@link makeGutsState}. A 4-player pot-vying gambling game; defaults to the human Declare phase (phase 0). */
const baseGutsState: GutsResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      chips: 200,
      in: false,
      out: false,
      roundBet: 10,
      cardCount: 2,
      cards: [
        { design: 'SPADE' as const, value: 13 },
        { design: 'HEART' as const, value: 11 },
      ],
      isWinner: false,
      isMatcher: false,
    },
    {
      id: 1,
      isHuman: false,
      chips: 200,
      in: false,
      out: false,
      roundBet: 10,
      cardCount: 2,
      cards: [],
      isWinner: false,
      isMatcher: false,
    },
    {
      id: 2,
      isHuman: false,
      chips: 200,
      in: false,
      out: false,
      roundBet: 10,
      cardCount: 2,
      cards: [],
      isWinner: false,
      isMatcher: false,
    },
    {
      id: 3,
      isHuman: false,
      chips: 200,
      in: false,
      out: false,
      roundBet: 10,
      cardCount: 2,
      cards: [],
      isWinner: false,
      isMatcher: false,
    },
  ],
  phase: 0,
  roundNumber: 1,
  pot: 40,
  carryPot: 0,
  ante: 10,
  chips: 200,
  winnerIdx: -1,
  matchWinnerIdx: -1,
  result: 0,
  matchers: [],
  gameEndFlag: false,
  hint: null,
  config: { playerCount: 4, ante: 10, startingChips: 200, targetRounds: 10 },
  message: '',
};

/**
 * Creates a {@link GutsResponse} with sensible defaults (the human Declare
 * phase, phase 0). Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial GutsResponse fields to override.
 * @returns A complete GutsResponse suitable for use in tests.
 */
export function makeGutsState(overrides?: Partial<GutsResponse>): GutsResponse {
  return { ...baseGutsState, ...overrides };
}

/** Base Bouillotte state used as the default for {@link makeBouillotteState}. A 4-player 3-card pot-vying game; defaults to the human Betting phase (phase 0). */
const baseBouillotteState: BouillotteResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      chips: 190,
      roundBet: 10,
      folded: false,
      out: false,
      cardCount: 3,
      cards: [
        { design: 'SPADE' as const, value: 13 },
        { design: 'HEART' as const, value: 11 },
        { design: 'CLOVER' as const, value: 4 },
      ],
      handName: 'highcard',
      isWinner: false,
    },
    {
      id: 1,
      isHuman: false,
      chips: 190,
      roundBet: 10,
      folded: false,
      out: false,
      cardCount: 3,
      cards: [],
      isWinner: false,
    },
    {
      id: 2,
      isHuman: false,
      chips: 190,
      roundBet: 10,
      folded: false,
      out: false,
      cardCount: 3,
      cards: [],
      isWinner: false,
    },
    {
      id: 3,
      isHuman: false,
      chips: 190,
      roundBet: 10,
      folded: false,
      out: false,
      cardCount: 3,
      cards: [],
      isWinner: false,
    },
  ],
  phase: 0,
  roundNumber: 1,
  pot: 40,
  ante: 10,
  chips: 190,
  currentBet: 10,
  raiseCount: 0,
  maxRaises: 4,
  currentPlayerIdx: 0,
  dealerIdx: 0,
  retourne: { design: 'CLOVER' as const, value: 13 },
  isHumanTurn: true,
  canRaise: true,
  winnerIdx: -1,
  matchWinnerIdx: -1,
  result: 0,
  gameEndFlag: false,
  hint: null,
  config: { playerCount: 4, ante: 10, startingChips: 200, targetRounds: 10 },
  message: '',
};

/**
 * Creates a {@link BouillotteResponse} with sensible defaults (the human
 * Betting phase, phase 0). Any field can be overridden via the `overrides`
 * parameter.
 *
 * @param overrides - Partial BouillotteResponse fields to override.
 * @returns A complete BouillotteResponse suitable for use in tests.
 */
export function makeBouillotteState(overrides?: Partial<BouillotteResponse>): BouillotteResponse {
  return { ...baseBouillotteState, ...overrides };
}

/** Base Primero state used as the default for {@link makePrimeroState}. A 4-player 4-card pot-vying game; defaults to the human Betting phase (phase 0). */
const basePrimeroState: PrimeroResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      chips: 190,
      roundBet: 10,
      folded: false,
      out: false,
      cardCount: 4,
      cards: [
        { design: 'SPADE' as const, value: 13 },
        { design: 'HEART' as const, value: 11 },
        { design: 'CLOVER' as const, value: 7 },
        { design: 'DIAMOND' as const, value: 4 },
      ],
      handName: 'primero',
      isWinner: false,
    },
    {
      id: 1,
      isHuman: false,
      chips: 190,
      roundBet: 10,
      folded: false,
      out: false,
      cardCount: 4,
      cards: [],
      isWinner: false,
    },
    {
      id: 2,
      isHuman: false,
      chips: 190,
      roundBet: 10,
      folded: false,
      out: false,
      cardCount: 4,
      cards: [],
      isWinner: false,
    },
    {
      id: 3,
      isHuman: false,
      chips: 190,
      roundBet: 10,
      folded: false,
      out: false,
      cardCount: 4,
      cards: [],
      isWinner: false,
    },
  ],
  phase: 0,
  roundNumber: 1,
  pot: 40,
  ante: 10,
  chips: 190,
  currentBet: 10,
  raiseCount: 0,
  maxRaises: 4,
  currentPlayerIdx: 0,
  dealerIdx: 0,
  isHumanTurn: true,
  canRaise: true,
  winnerIdx: -1,
  matchWinnerIdx: -1,
  result: 0,
  gameEndFlag: false,
  hint: null,
  config: { playerCount: 4, ante: 10, startingChips: 200, targetRounds: 10 },
  message: '',
};

/**
 * Creates a {@link PrimeroResponse} with sensible defaults (the human Betting
 * phase, phase 0). Any field can be overridden via the `overrides` parameter.
 *
 * @param overrides - Partial PrimeroResponse fields to override.
 * @returns A complete PrimeroResponse suitable for use in tests.
 */
export function makePrimeroState(overrides?: Partial<PrimeroResponse>): PrimeroResponse {
  return { ...basePrimeroState, ...overrides };
}
