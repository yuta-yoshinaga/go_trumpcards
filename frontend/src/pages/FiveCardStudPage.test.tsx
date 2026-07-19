import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fiveCardStudApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { FiveCardStudPlayerData, FiveCardStudResponse } from '../types/card';
import { FiveCardStudPage } from './FiveCardStudPage';

vi.mock('../api/gameApi', () => ({
  fiveCardStudApi: { exec: vi.fn() },
  actionLogApi: { fivecardstud: vi.fn() },
}));

const mockExec = vi.mocked(fiveCardStudApi.exec);

/** Helper: base human player (1 hole card + up to 4 door cards). */
const humanPlayer = (overrides: Partial<FiveCardStudPlayerData> = {}): FiveCardStudPlayerData => ({
  id: 0,
  isHuman: true,
  holeCards: [{ design: 'SPADE', value: 1 }],
  doorCards: [
    { design: 'CLOVER', value: 5 },
    { design: 'SPADE', value: 8 },
    { design: 'HEART', value: 3 },
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

/** Helper: base CPU player. */
const cpuPlayer = (id: number, overrides: Partial<FiveCardStudPlayerData> = {}): FiveCardStudPlayerData => ({
  id,
  isHuman: false,
  holeCards: [],
  doorCards: [
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
  playStyleName: 'タイト',
  totalHands: 0,
  vpip: 0,
  pfr: 0,
  threeBet: 0,
  af: '-',
  ...overrides,
});

const baseState: Omit<FiveCardStudResponse, 'players' | 'phase' | 'message'> = {
  communityCard: null,
  pot: 0,
  sidePots: [],
  dealerIdx: 0,
  currentTurn: 0,
  gameEndFlag: false,
  lastBet: 0,
  minRaise: 20,
  roundResults: [],
  cpuActions: [],
  handCount: 0,
  ante: 5,
  bringIn: 10,
  smallBet: 20,
  bigBet: 40,
  tournamentMode: false,
  anteLevelHands: 10,
  anteMultiplier: 200,
  bettingLimit: 0,
  raiseCount: 0,
  maxBetAmount: 0,
  tableSize: 4,
  bringInPlayerIdx: 0,
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
};

/** INIT state (phase 0): no players yet */
const initState: FiveCardStudResponse = { ...baseState, players: [], phase: 0, message: '' };

/** SECOND_STREET (phase 1): human's turn, no outstanding bet */
const secondStreetState: FiveCardStudResponse = {
  ...baseState,
  players: [humanPlayer(), cpuPlayer(1), cpuPlayer(2), cpuPlayer(3)],
  pot: 30,
  dealerIdx: 3,
  currentTurn: 0,
  phase: 1,
  message: 'あなたの番です',
};

/** SECOND_STREET with outstanding bet: shows call/raise instead of bet/check */
const secondStreetWithBetState: FiveCardStudResponse = {
  ...secondStreetState,
  lastBet: 40,
  cpuActions: [{ playerIdx: 1, action: 3, amount: 40 }],
};

/** SHOWDOWN (phase 5) */
const showdownState: FiveCardStudResponse = {
  ...baseState,
  players: [
    humanPlayer({ handName: 'ワンペア', currentBet: 0, chips: 950 }),
    cpuPlayer(1, { handName: 'ツーペア', folded: false, holeCards: [{ design: 'SPADE', value: 5 }] }),
    cpuPlayer(2, { folded: true }),
  ],
  pot: 0,
  dealerIdx: 2,
  currentTurn: -1,
  phase: 5,
  roundResults: [
    { playerIdx: 0, handRank: 1, handName: 'ワンペア', kickers: 'A', bestHand: [], wonAmount: 0, mucked: false },
    { playerIdx: 1, handRank: 2, handName: 'ツーペア', kickers: '8', bestHand: [], wonAmount: 200, mucked: false },
  ],
  message: 'CPU 1 の勝ち',
};

/** END (phase 6) where the human takes the pot (positive wonAmount). */
const endHumanWinState: FiveCardStudResponse = {
  ...showdownState,
  phase: 6,
  gameEndFlag: true,
  players: [
    humanPlayer({ handName: 'ツーペア', chips: 1200 }),
    cpuPlayer(1, { handName: 'ワンペア', holeCards: [{ design: 'SPADE', value: 5 }] }),
    cpuPlayer(2, { folded: true }),
  ],
  roundResults: [
    { playerIdx: 0, handRank: 2, handName: 'ツーペア', kickers: 'A', bestHand: [], wonAmount: 200, mucked: false },
    { playerIdx: 1, handRank: 1, handName: 'ワンペア', kickers: '5', bestHand: [], wonAmount: 0, mucked: false },
  ],
  message: 'あなたの勝ち',
};

/** END (phase 6) where a CPU wins and the human gets nothing. */
const endHumanLossState: FiveCardStudResponse = {
  ...showdownState,
  phase: 6,
  gameEndFlag: true,
  roundResults: [
    { playerIdx: 0, handRank: 1, handName: 'ワンペア', kickers: 'A', bestHand: [], wonAmount: 0, mucked: false },
    { playerIdx: 1, handRank: 2, handName: 'ツーペア', kickers: '8', bestHand: [], wonAmount: 200, mucked: false },
  ],
  message: 'CPU 1 の勝ち',
};

beforeEach(() => {
  mockExec.mockResolvedValue(initState);
});

describe('FiveCardStudPage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<FiveCardStudPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<FiveCardStudPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows "初期化中" when phase is INIT', async () => {
    renderWithProviders(<FiveCardStudPage />);
    await waitFor(() => expect(screen.getByText('初期化中')).toBeInTheDocument());
  });

  it('shows known phase name for SECOND_STREET', async () => {
    mockExec.mockResolvedValue(secondStreetState);
    renderWithProviders(<FiveCardStudPage />);
    await waitFor(() => expect(screen.getByText('セカンドストリート')).toBeInTheDocument());
  });

  it('shows pot and the dealer name via playerName (CPU dealer)', async () => {
    mockExec.mockResolvedValue(secondStreetState);
    renderWithProviders(<FiveCardStudPage />);
    await waitFor(() => expect(screen.getByText(/ポット:/)).toBeInTheDocument());
    expect(screen.getByText(/ディーラー:/)).toBeInTheDocument();
    expect(screen.getAllByText('CPU 3').length).toBeGreaterThan(0);
  });

  it('renders CPU player info and shows fold/all-in badges', async () => {
    mockExec.mockResolvedValue({
      ...secondStreetState,
      players: [humanPlayer(), cpuPlayer(1, { folded: true }), cpuPlayer(2, { allIn: true })],
    });
    renderWithProviders(<FiveCardStudPage />);
    await waitFor(() => expect(screen.getByText('[フォールド]')).toBeInTheDocument());
    expect(screen.getByText('[オールイン]')).toBeInTheDocument();
    expect(screen.getAllByText(/タイト/).length).toBeGreaterThanOrEqual(1);
  });

  it('shows door and hole card labels and human cards', async () => {
    mockExec.mockResolvedValue(secondStreetState);
    renderWithProviders(<FiveCardStudPage />);
    await waitFor(() => expect(screen.getByAltText('♣ 5')).toBeInTheDocument());
    expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    expect(screen.getAllByText('ドアカード').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('ホールカード').length).toBeGreaterThanOrEqual(1);
  });

  it('highlights the latest door card for the human and each CPU', async () => {
    mockExec.mockResolvedValue(secondStreetState);
    renderWithProviders(<FiveCardStudPage />);
    // Human's newest door card (last of 4) gets a ring highlight.
    const humanLatest = await screen.findByTestId('latest-door-human');
    expect(humanLatest.className).toContain('ring-2');
    expect(humanLatest.className).toContain('motion-safe:animate-pulse');
    // The newest door card is the last in the array (♦ 7), not an earlier one.
    expect(humanLatest.querySelector('img')).toHaveAttribute('alt', '♦ 7');
    // Each CPU also has exactly one highlighted latest door card.
    expect(screen.getByTestId('latest-door-cpu-1')).toBeInTheDocument();
    expect(screen.getByTestId('latest-door-cpu-2')).toBeInTheDocument();
  });

  it('reveals CPU hand name and hole card during showdown and shows round results', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<FiveCardStudPage />);
    await waitFor(() => expect(screen.getByText('ツーペア')).toBeInTheDocument());
    expect(screen.getByAltText('♠ 5')).toBeInTheDocument();
    expect(screen.getByText('結果:')).toBeInTheDocument();
  });

  it('shows bet/check buttons when canAct and no outstanding bet', async () => {
    mockExec.mockResolvedValue(secondStreetState);
    renderWithProviders(<FiveCardStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument();
  });

  it('shows call/raise buttons when canAct and has outstanding bet', async () => {
    mockExec.mockResolvedValue(secondStreetWithBetState);
    renderWithProviders(<FiveCardStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'レイズ' })).toBeInTheDocument();
  });

  it('hides betting controls when not active phase', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<FiveCardStudPage />);
    await waitFor(() => expect(screen.getByText('ショーダウン')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  it('calls bet command with betAmount', async () => {
    mockExec.mockResolvedValue(secondStreetState);
    renderWithProviders(<FiveCardStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(secondStreetState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 20, undefined, 0));
  });

  it('calls fold command', async () => {
    mockExec.mockResolvedValue(secondStreetState);
    renderWithProviders(<FiveCardStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(secondStreetState);
    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold', undefined, undefined, 0));
  });

  it('calls reset command when reset button is clicked', async () => {
    renderWithProviders(<FiveCardStudPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuMetaAI: false }));
  });

  it('shows error message when API call fails', async () => {
    renderWithProviders(<FiveCardStudPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('shows muck/show controls and calls muck command', async () => {
    mockExec.mockResolvedValue({ ...showdownState, muckAvailable: true, phase: 5 });
    renderWithProviders(<FiveCardStudPage />);
    await waitFor(() => expect(screen.getByTestId('muck-controls')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ショー' })).toBeInTheDocument();
    mockExec.mockClear();
    mockExec.mockResolvedValue(showdownState);
    fireEvent.click(screen.getByRole('button', { name: 'マック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('muck'));
  });

  it('shows rebuy controls and calls rebuy command', async () => {
    const rebuyState: FiveCardStudResponse = {
      ...baseState,
      players: [humanPlayer({ chips: 0 }), cpuPlayer(1)],
      phase: 7,
      message: '',
      rebuyPhaseType: 1,
      rebuyChips: 1000,
      rebuyMaxCount: 3,
      rebuyCounts: [1, 0],
    };
    mockExec.mockResolvedValue(rebuyState);
    renderWithProviders(<FiveCardStudPage />);
    await waitFor(() => expect(screen.getByTestId('rebuy-controls')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リバイ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('rebuy'));
  });

  it('shows addon controls and toggles cpuMetaAI on reset', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      players: [humanPlayer(), cpuPlayer(1)],
      phase: 7,
      message: '',
      rebuyPhaseType: 2,
      addonChips: 1500,
    });
    renderWithProviders(<FiveCardStudPage />);
    await waitFor(() => expect(screen.getByTestId('addon-controls')).toBeInTheDocument());

    const metaAICheckbox = screen.getByRole('checkbox', { name: 'メタAI（CPUがプレイスタイルを学習）' });
    fireEvent.click(metaAICheckbox);
    expect(metaAICheckbox).toBeChecked();

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuMetaAI: true }));
  });

  it('plays the win celebration when the human wins the pot', async () => {
    mockExec.mockResolvedValue(endHumanWinState);
    renderWithProviders(<FiveCardStudPage />);
    expect(await screen.findByTestId('win-celebration')).toBeInTheDocument();
  });

  it('does not celebrate when the human loses the hand', async () => {
    mockExec.mockResolvedValue(endHumanLossState);
    renderWithProviders(<FiveCardStudPage />);
    await waitFor(() => expect(screen.queryByTestId('skeleton')).not.toBeInTheDocument());
    // Outwait the celebration's 400ms delay so a wrongly-fired overlay would be visible.
    await act(() => new Promise((resolve) => setTimeout(resolve, 600)));
    expect(screen.queryByTestId('win-celebration')).not.toBeInTheDocument();
  });
});
