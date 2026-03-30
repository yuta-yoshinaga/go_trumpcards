import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { pineappleApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { PineappleResponse } from '../types/card';
import { PineapplePage } from './PineapplePage';

vi.mock('../api/gameApi', () => ({
  pineappleApi: { exec: vi.fn() },
  actionLogApi: { pineapple: vi.fn() },
}));

const mockExec = vi.mocked(pineappleApi.exec);

/** Helper: base human player */
const humanPlayer = (overrides: Partial<import('../types/card').HoldemPlayerData> = {}) => ({
  id: 0,
  isHuman: true,
  cards: [
    { design: 'SPADE' as const, value: 1 },
    { design: 'HEART' as const, value: 13 },
    { design: 'DIAMOND' as const, value: 7 },
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

/** Helper: base CPU player */
const cpuPlayer = (id: number, overrides: Partial<import('../types/card').HoldemPlayerData> = {}) => ({
  id,
  isHuman: false,
  cards: [
    { design: 'DIAMOND' as const, value: 2 },
    { design: 'CLOVER' as const, value: 7 },
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

/** INIT state (phase 0): no players yet */
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
  roundResults: [],
  cpuActions: [],
  message: '',
  handCount: 0,
  smallBlind: 5,
  bigBlind: 10,
  tournamentMode: false,
  blindLevelHands: 10,
  blindMultiplier: 200,
  bettingLimit: 0,
  raiseCount: 0,
  maxBetAmount: 0,
  tableSize: 4,
  rebuyPhaseType: 0,
  rebuyChips: 0,
  rebuyMaxCount: 0,
  rebuyCounts: [],
  addonChips: 0,
  rebuyAvailable: false,
  addonAvailable: false,
  rebuyEnabled: false,
  addonEnabled: false,
  rebuyPeriodHands: 0,
  addonAfterHand: 0,
  addonUsed: [],
  muckAvailable: false,
  isDiscardPhase: false,
  discardDone: [],
};

/** PRE_FLOP (phase 1): human's turn, no outstanding bet */
const preFlopState: PineappleResponse = {
  ...initState,
  players: [humanPlayer(), cpuPlayer(1), cpuPlayer(2), cpuPlayer(3)],
  pot: 30,
  dealerIdx: 3,
  currentTurn: 0,
  phase: 1,
  message: 'あなたの番です',
  handCount: 1,
  discardDone: [false, false, false, false],
};

/** Discard phase state */
const discardState: PineappleResponse = {
  ...preFlopState,
  phase: 8,
  isDiscardPhase: true,
  discardDone: [false, true, true, true],
  communityCards: [
    { design: 'SPADE', value: 10 },
    { design: 'HEART', value: 5 },
    { design: 'DIAMOND', value: 8 },
  ],
};

/** SHOWDOWN (phase 5) */
const showdownState: PineappleResponse = {
  ...initState,
  players: [
    humanPlayer({
      handName: 'ワンペア',
      currentBet: 0,
      chips: 950,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 13 },
      ],
    }),
    cpuPlayer(1, {
      handName: 'ツーペア',
      folded: false,
      cards: [
        { design: 'SPADE', value: 5 },
        { design: 'HEART', value: 8 },
      ],
    }),
    cpuPlayer(2, { folded: true }),
  ],
  communityCards: [
    { design: 'SPADE', value: 10 },
    { design: 'HEART', value: 5 },
    { design: 'DIAMOND', value: 8 },
    { design: 'CLOVER', value: 2 },
    { design: 'HEART', value: 9 },
  ],
  pot: 0,
  dealerIdx: 2,
  currentTurn: -1,
  phase: 5,
  roundResults: [
    { playerIdx: 0, handRank: 1, handName: 'ワンペア', kickers: 'A, Q, 10', bestHand: [], wonAmount: 0, mucked: false },
    { playerIdx: 1, handRank: 2, handName: 'ツーペア', kickers: '8', bestHand: [], wonAmount: 200, mucked: false },
  ],
  message: 'CPU 1 の勝ち',
  handCount: 1,
  discardDone: [true, true, true],
};

beforeEach(() => {
  mockExec.mockResolvedValue(initState);
});

describe('PineapplePage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<PineapplePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows phase name "初期化中" for INIT phase', async () => {
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByText('初期化中')).toBeInTheDocument());
  });

  it('shows betting controls during pre-flop when it is human turn', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByText('あなたの手札')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
  });

  it('renders discard controls during discard phase', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
    expect(screen.getByText('捨てるカードを選択してください')).toBeInTheDocument();
  });

  it('allows selecting a card and clicking discard', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());

    // Discard button should be disabled until a card is selected
    const discardBtn = screen.getByRole('button', { name: '1枚捨ててください。' });
    expect(discardBtn).toBeDisabled();

    // Click the first card
    const cardButtons = screen.getAllByRole('button').filter((btn) => btn.getAttribute('aria-pressed') !== null);
    fireEvent.click(cardButtons[0]);

    // Now discard button should be enabled
    await waitFor(() => expect(discardBtn).not.toBeDisabled());

    fireEvent.click(discardBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', undefined, { cardIdx: 0 }));
  });

  it('shows round results during showdown', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByText('ワンペア')).toBeInTheDocument());
  });

  it('shows reset button and triggers reset flow', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByText('あなたの手札')).toBeInTheDocument());

    const resetBtn = screen.getByRole('button', { name: 'リセット' });
    expect(resetBtn).toBeInTheDocument();
    fireEvent.click(resetBtn);
    expect(screen.getByText('本当にゲームをリセットしますか？')).toBeInTheDocument();
  });
});
