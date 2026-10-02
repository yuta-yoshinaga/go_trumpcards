import { screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { crazyPineappleApi, pineappleApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { HoldemPlayerData, PineappleResponse } from '../types/card';
import { CrazyPineapplePage } from './CrazyPineapplePage';

vi.mock('../api/gameApi', () => ({
  pineappleApi: { exec: vi.fn() },
  crazyPineappleApi: { exec: vi.fn() },
  actionLogApi: { crazypineapple: vi.fn() },
}));

const mockCrazyExec = vi.mocked(crazyPineappleApi.exec);
const mockPineappleExec = vi.mocked(pineappleApi.exec);

const humanPlayer = (overrides: Partial<HoldemPlayerData> = {}): HoldemPlayerData => ({
  id: 0,
  isHuman: true,
  cards: [
    { design: 'SPADE', value: 1 },
    { design: 'HEART', value: 13 },
    { design: 'DIAMOND', value: 7 },
  ],
  chips: 980,
  currentBet: 0,
  folded: false,
  allIn: false,
  handRank: 0,
  handName: '',
  bestHand: [],
  playStyleName: '',
  totalHands: 0,
  vpip: 0,
  pfr: 0,
  threeBet: 0,
  af: '-',
  ...overrides,
});

const cpuPlayer = (id: number, overrides: Partial<HoldemPlayerData> = {}): HoldemPlayerData => ({
  id,
  isHuman: false,
  cards: [
    { design: 'DIAMOND', value: 2 },
    { design: 'CLOVER', value: 7 },
  ],
  chips: 1000,
  currentBet: 0,
  folded: false,
  allIn: false,
  handRank: 0,
  handName: '',
  bestHand: [],
  playStyleName: '',
  totalHands: 0,
  vpip: 0,
  pfr: 0,
  threeBet: 0,
  af: '-',
  ...overrides,
});

const initState: PineappleResponse = {
  players: [],
  communityCards: [],
  pot: 0,
  sidePots: [],
  dealerIdx: 0,
  currentTurn: 0,
  phase: 0,
  gameEndFlag: false,
  lastBet: 0,
  minRaise: 20,
  raiseCount: 0,
  maxBetAmount: 0,
  roundResults: [],
  cpuActions: [],
  message: '',
  handCount: 0,
  smallBlind: 5,
  bigBlind: 10,
  bettingLimit: 0,
  tournamentMode: false,
  blindLevelHands: 10,
  blindMultiplier: 150,
  rebuyEnabled: false,
  rebuyMaxCount: 2,
  rebuyChips: 1000,
  rebuyPeriodHands: 5,
  rebuyCounts: [],
  rebuyAvailable: false,
  addonEnabled: false,
  addonChips: 500,
  addonAfterHand: 5,
  addonUsed: [],
  addonAvailable: false,
  rebuyPhaseType: 0,
  muckAvailable: false,
  tableSize: 4,
  isDiscardPhase: false,
  discardDone: [],
  initialDealCount: 3,
  liveBestHand: '',
};

/** Discard phase state — set after the flop betting round in Crazy Pineapple. */
const discardState: PineappleResponse = {
  ...initState,
  players: [humanPlayer(), cpuPlayer(1), cpuPlayer(2), cpuPlayer(3)],
  pot: 60,
  dealerIdx: 3,
  currentTurn: 0,
  phase: 8,
  isDiscardPhase: true,
  discardDone: [false, true, true, true],
  communityCards: [
    { design: 'SPADE', value: 10 },
    { design: 'HEART', value: 5 },
    { design: 'DIAMOND', value: 8 },
  ],
  message: '',
  handCount: 1,
};

describe('CrazyPineapplePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCrazyExec.mockResolvedValue(initState);
    mockPineappleExec.mockResolvedValue(initState);
  });

  it('calls crazyPineappleApi.exec on mount (NOT pineappleApi)', async () => {
    renderWithProviders(<CrazyPineapplePage />);
    await waitFor(() => {
      expect(mockCrazyExec).toHaveBeenCalledWith('reset');
    });
    expect(mockPineappleExec).not.toHaveBeenCalled();
  });

  it('renders discard controls during discard phase (after flop betting in Crazy mode)', async () => {
    mockCrazyExec.mockResolvedValue(discardState);
    renderWithProviders(<CrazyPineapplePage />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
    // The discard prompt comes from the crazypineapple i18n namespace, which
    // mirrors the pineapple namespace key. This confirms the variant is wired
    // through useGamePageSetup → useTranslation('crazypineapple').
    expect(screen.getByText('捨てるカードを選択してください')).toBeInTheDocument();
  });

  it('forewarns the upcoming discard with a banner during the flop betting round', async () => {
    const flopState: PineappleResponse = {
      ...initState,
      players: [humanPlayer(), cpuPlayer(1), cpuPlayer(2), cpuPlayer(3)],
      pot: 60,
      currentTurn: 0,
      phase: 2, // PineapplePhase.FLOP
      communityCards: [
        { design: 'SPADE', value: 10 },
        { design: 'HEART', value: 5 },
        { design: 'DIAMOND', value: 8 },
      ],
      handCount: 1,
    };
    mockCrazyExec.mockResolvedValue(flopState);
    renderWithProviders(<CrazyPineapplePage />);
    await waitFor(() => expect(screen.getByTestId('cp-discard-upcoming-banner')).toBeInTheDocument());
    expect(screen.getByTestId('cp-discard-upcoming-banner')).toHaveTextContent(
      'フロップベット終了後にカードを1枚捨てます',
    );
  });
});
