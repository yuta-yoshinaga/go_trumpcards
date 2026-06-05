import type {
  CallBreakResponse,
  GongZhuResponse,
  HeartsResponse,
  SpadesResponse,
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
