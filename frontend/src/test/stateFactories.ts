import type {
  CallBreakResponse,
  DoppelkopfResponse,
  GongZhuResponse,
  HeartsResponse,
  MusResponse,
  SheepsheadResponse,
  SpadesResponse,
  SuecaResponse,
  TressetteResponse,
  TuteResponse,
  TwoTenJackResponse,
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
