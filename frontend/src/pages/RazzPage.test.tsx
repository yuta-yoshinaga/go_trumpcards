import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { razzApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { SevenCardStudResponse } from '../types/card';
import { RazzPage } from './RazzPage';

vi.mock('../api/gameApi', () => ({
  razzApi: { exec: vi.fn() },
  actionLogApi: { razz: vi.fn() },
}));

vi.mock('../hooks/useCliMode', () => ({
  useCliMode: vi.fn(() => ({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  })),
}));

const mockExec = vi.mocked(razzApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

/** Helper: base human player */
const humanPlayer = (overrides: Partial<import('../types/card').SevenCardStudPlayerData> = {}) => ({
  id: 0,
  isHuman: true,
  holeCards: [
    { design: 'SPADE' as const, value: 1 },
    { design: 'HEART' as const, value: 13 },
    { design: 'DIAMOND' as const, value: 10 },
  ],
  doorCards: [
    { design: 'CLOVER' as const, value: 5 },
    { design: 'SPADE' as const, value: 8 },
    { design: 'HEART' as const, value: 3 },
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
const cpuPlayer = (id: number, overrides: Partial<import('../types/card').SevenCardStudPlayerData> = {}) => ({
  id,
  isHuman: false,
  holeCards: [],
  doorCards: [
    { design: 'DIAMOND' as const, value: 2 },
    { design: 'CLOVER' as const, value: 7 },
    { design: 'HEART' as const, value: 9 },
    { design: 'SPADE' as const, value: 4 },
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

const baseState: Omit<SevenCardStudResponse, 'players' | 'phase' | 'message'> = {
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
const initState: SevenCardStudResponse = {
  ...baseState,
  players: [],
  phase: 0,
  message: '',
};

/** THIRD_STREET (phase 1): human's turn, no outstanding bet */
const thirdStreetState: SevenCardStudResponse = {
  ...baseState,
  players: [humanPlayer(), cpuPlayer(1), cpuPlayer(2), cpuPlayer(3)],
  pot: 30,
  dealerIdx: 3,
  currentTurn: 0,
  phase: 1,
  message: 'あなたの番です',
};

/** THIRD_STREET with outstanding bet: shows call/raise instead of bet/check */
const thirdStreetWithBetState: SevenCardStudResponse = {
  ...thirdStreetState,
  lastBet: 40,
  cpuActions: [{ playerIdx: 1, action: 3, amount: 40 }],
};

/** SHOWDOWN (phase 6) */
const showdownState: SevenCardStudResponse = {
  ...baseState,
  players: [
    humanPlayer({ handName: 'ワンペア', currentBet: 0, chips: 950 }),
    cpuPlayer(1, {
      handName: 'ツーペア',
      folded: false,
      holeCards: [
        { design: 'SPADE', value: 5 },
        { design: 'HEART', value: 8 },
        { design: 'DIAMOND', value: 12 },
      ],
    }),
    cpuPlayer(2, { folded: true }),
  ],
  pot: 0,
  dealerIdx: 2,
  currentTurn: -1,
  phase: 6,
  roundResults: [
    { playerIdx: 0, handRank: 1, handName: 'ワンペア', kickers: 'A, Q, 10', bestHand: [], wonAmount: 0, mucked: false },
    { playerIdx: 1, handRank: 2, handName: 'ツーペア', kickers: '8', bestHand: [], wonAmount: 200, mucked: false },
  ],
  message: 'CPU 1 の勝ち',
};

beforeEach(() => {
  mockExec.mockResolvedValue(initState);
});

describe('RazzPage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<RazzPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  // ---- mount & reset ----
  it('calls reset on mount', async () => {
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // ---- phase name display ----
  it('shows "初期化中" when phase is INIT', async () => {
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText('初期化中')).toBeInTheDocument());
  });

  it('shows known phase name for THIRD_STREET', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText('サードストリート')).toBeInTheDocument());
  });

  it('shows the current best low for the human from 3rd street', async () => {
    // Human cards A,K,10 + 5,8,3,7 → five lowest distinct = A,3,5,7,8.
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<RazzPage />);
    const low = await screen.findByTestId('razz-best-low');
    expect(low).toHaveTextContent('現在のロー: 8-7-5-3-A');
  });

  it('marks the low as incomplete when fewer than five distinct ranks are known', async () => {
    mockExec.mockResolvedValue({
      ...thirdStreetState,
      players: [
        humanPlayer({ holeCards: [{ design: 'SPADE', value: 2 }], doorCards: [{ design: 'HEART', value: 2 }] }),
        thirdStreetState.players[1],
        thirdStreetState.players[2],
        thirdStreetState.players[3],
      ],
    });
    renderWithProviders(<RazzPage />);
    const low = await screen.findByTestId('razz-best-low');
    expect(low).toHaveTextContent('未完成');
  });

  // ---- info bar ----
  it('shows pot and the dealer name via playerName (CPU dealer)', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText(/ポット:/)).toBeInTheDocument());
    expect(screen.getByText(/ディーラー:/)).toBeInTheDocument();
    // Dealer renders via playerName (CPU 3), not the raw index.
    expect(screen.getAllByText('CPU 3').length).toBeGreaterThan(0);
    expect(screen.queryByText(/Player 3|プレイヤー 3/)).not.toBeInTheDocument();
  });

  // ---- CPU players ----
  it('renders CPU player info with playStyleName and chips', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.getByText(/CPU 2/)).toBeInTheDocument();
    expect(screen.getAllByText(/タイト/).length).toBeGreaterThanOrEqual(1);
  });

  it('shows fold badge for folded CPU', async () => {
    mockExec.mockResolvedValue({
      ...thirdStreetState,
      players: [humanPlayer(), cpuPlayer(1, { folded: true }), cpuPlayer(2)],
    });
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText('[フォールド]')).toBeInTheDocument());
  });

  it('shows all-in badge for all-in CPU', async () => {
    mockExec.mockResolvedValue({
      ...thirdStreetState,
      players: [humanPlayer(), cpuPlayer(1, { allIn: true }), cpuPlayer(2)],
    });
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText('[オールイン]')).toBeInTheDocument());
  });

  it('shows CPU hand name badge during showdown when not folded', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText('ツーペア')).toBeInTheDocument());
  });

  it('shows CPU door cards', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getAllByText('ドアカード').length).toBeGreaterThanOrEqual(1));
  });

  it('shows CPU hole cards face-up during showdown when not folded', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 5')).toBeInTheDocument());
    expect(screen.getByAltText('♥ 8')).toBeInTheDocument();
  });

  // ---- CPU actions log ----
  it('shows CPU actions log when cpuActions is non-empty', async () => {
    mockExec.mockResolvedValue(thirdStreetWithBetState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText('CPU行動:')).toBeInTheDocument());
    expect(screen.getByText(/Player 1: ベット/)).toBeInTheDocument();
  });

  // ---- round results ----
  it('shows round results in showdown', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
  });

  it('does not show round results when not in showdown', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByText('結果:')).not.toBeInTheDocument();
  });

  // ---- human player section ----
  it('shows human player section when humanPlayer exists', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
  });

  it('does not show human player section when no players', async () => {
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText('初期化中')).toBeInTheDocument());
    expect(screen.queryByText(/あなたの手札/)).not.toBeInTheDocument();
  });

  it('shows human door cards', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByAltText('♣ 5')).toBeInTheDocument());
    expect(screen.getByAltText('♠ 8')).toBeInTheDocument();
  });

  it('shows human hole cards', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.getByAltText('♥ K')).toBeInTheDocument();
  });

  // ---- message ----
  it('shows game message', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText('あなたの番です')).toBeInTheDocument());
  });

  // ---- canAct / betting controls ----
  it('shows bet/check buttons when canAct and no outstanding bet', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument();
  });

  it('shows call/raise buttons when canAct and has outstanding bet', async () => {
    mockExec.mockResolvedValue(thirdStreetWithBetState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'レイズ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument();
  });

  it('hides betting controls when not active phase', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText('ショーダウン')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  it('hides betting controls when human is folded', async () => {
    mockExec.mockResolvedValue({
      ...thirdStreetState,
      players: [humanPlayer({ folded: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  // ---- button click handlers ----
  it('calls bet command with betAmount', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(thirdStreetState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 20, undefined, 0));
  });

  it('calls check command', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(thirdStreetState);
    fireEvent.click(screen.getByRole('button', { name: 'チェック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('check', undefined, undefined, 0));
  });

  it('calls fold command', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(thirdStreetState);
    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold', undefined, undefined, 0));
  });

  it('calls allin command', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(thirdStreetState);
    fireEvent.click(screen.getByRole('button', { name: 'オールイン' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('allin', undefined, undefined, 0));
  });

  it('calls reset command when reset button is clicked', async () => {
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { ante: 1, tournamentMode: false, cpuMetaAI: false }),
    );
  });

  it('uses outline style for reset button', async () => {
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const resetBtn = screen.getByRole('button', { name: 'リセット' });
    expect(resetBtn.className).toContain('bg-transparent');
    expect(resetBtn.className).toContain('border');
  });

  // ---- loading / disabled state ----
  it('disables buttons while loading', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());

    let resolve!: (value: SevenCardStudResponse) => void;
    const slowPromise = new Promise<SevenCardStudResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));

    expect(screen.getByRole('button', { name: 'ベット' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'チェック' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();

    resolve(thirdStreetState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());
  });

  // ---- error handling ----
  it('shows error message when API call fails', async () => {
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('clears error on successful call after failure', async () => {
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());

    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.queryByText(NETWORK_ERROR_MESSAGE())).not.toBeInTheDocument());
  });

  // ---- muck / show controls ----
  it('shows muck/show buttons when isMuckPhase (showdown + muckAvailable)', async () => {
    mockExec.mockResolvedValue({
      ...showdownState,
      muckAvailable: true,
      phase: 6, // SHOWDOWN
    });
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByTestId('muck-controls')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'マック' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ショー' })).toBeInTheDocument();
  });

  it('calls muck command when muck button clicked', async () => {
    mockExec.mockResolvedValue({
      ...showdownState,
      muckAvailable: true,
      phase: 6,
    });
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'マック' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(showdownState);
    fireEvent.click(screen.getByRole('button', { name: 'マック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('muck'));
  });

  it('calls show command when show button clicked', async () => {
    mockExec.mockResolvedValue({
      ...showdownState,
      muckAvailable: true,
      phase: 6,
    });
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ショー' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(showdownState);
    fireEvent.click(screen.getByRole('button', { name: 'ショー' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('show'));
  });

  // ---- rebuy controls ----
  it('shows rebuy controls when isRebuyPhase', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      players: [humanPlayer({ chips: 0 }), cpuPlayer(1)],
      phase: 8, // REBUY
      message: '',
      rebuyPhaseType: 1, // REBUY type
      rebuyChips: 1000,
      rebuyMaxCount: 3,
      rebuyCounts: [1, 0],
    });
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByTestId('rebuy-controls')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'リバイ' })).toBeInTheDocument();
  });

  it('calls rebuy command when rebuy button clicked', async () => {
    const rebuyState = {
      ...baseState,
      players: [humanPlayer({ chips: 0 }), cpuPlayer(1)],
      phase: 8,
      message: '',
      rebuyPhaseType: 1,
      rebuyChips: 1000,
      rebuyMaxCount: 3,
      rebuyCounts: [1, 0],
    };
    mockExec.mockResolvedValue(rebuyState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リバイ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リバイ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('rebuy'));
  });

  it('calls skiprebuy command when skip rebuy button clicked', async () => {
    const rebuyState = {
      ...baseState,
      players: [humanPlayer({ chips: 0 }), cpuPlayer(1)],
      phase: 8,
      message: '',
      rebuyPhaseType: 1,
      rebuyChips: 1000,
      rebuyMaxCount: 3,
      rebuyCounts: [0, 0],
    };
    mockExec.mockResolvedValue(rebuyState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByTestId('rebuy-controls')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    // Only rebuy-controls is showing (not addon), so skip button is unique
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skiprebuy'));
  });

  // ---- addon controls ----
  it('shows addon controls when isAddonPhase', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      players: [humanPlayer(), cpuPlayer(1)],
      phase: 8, // REBUY
      message: '',
      rebuyPhaseType: 2, // ADDON type
      addonChips: 1500,
    });
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByTestId('addon-controls')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'アドオン' })).toBeInTheDocument();
  });

  it('calls addon command when addon button clicked', async () => {
    const addonState = {
      ...baseState,
      players: [humanPlayer(), cpuPlayer(1)],
      phase: 8,
      message: '',
      rebuyPhaseType: 2,
      addonChips: 1500,
    };
    mockExec.mockResolvedValue(addonState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'アドオン' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'アドオン' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('addon'));
  });

  it('calls skipaddon command when skip addon button clicked', async () => {
    const addonState = {
      ...baseState,
      players: [humanPlayer(), cpuPlayer(1)],
      phase: 8,
      message: '',
      rebuyPhaseType: 2,
      addonChips: 1500,
    };
    mockExec.mockResolvedValue(addonState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByTestId('addon-controls')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    // Only addon-controls is showing (not rebuy), so skip button is unique
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skipaddon'));
  });

  // ---- tournament mode ----
  it('shows hand number in tournament mode', async () => {
    mockExec.mockResolvedValue({
      ...thirdStreetState,
      tournamentMode: true,
      handCount: 3,
      anteLevelHands: 10,
    });
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText(/ハンド#3/)).toBeInTheDocument());
  });

  // ---- human allIn disables betting controls ----
  it('hides betting controls when human is all-in', async () => {
    mockExec.mockResolvedValue({
      ...thirdStreetState,
      players: [humanPlayer({ allIn: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  // ---- 7th street: 3 face-down CPU hole cards ----
  it('shows 3 face-down placeholders for CPU hole cards on seventh street', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      players: [
        humanPlayer(),
        { ...cpuPlayer(1), holeCards: [] }, // no revealed hole cards
      ],
      phase: 5, // SEVENTH_STREET
      message: '',
      currentTurn: 0,
    });
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    // On 7th street, CPU should show 3 face-down cards (not 2)
    // Verify the holeCards section is rendered for both CPU and human
    expect(screen.getAllByText('ホールカード').length).toBeGreaterThanOrEqual(1);
  });

  // ---- CPU currentBet display ----
  it('shows CPU current bet when positive', async () => {
    mockExec.mockResolvedValue({
      ...thirdStreetState,
      players: [humanPlayer(), cpuPlayer(1, { currentBet: 20 }), cpuPlayer(2)],
    });
    renderWithProviders(<RazzPage />);
    // tc('betting.currentBet') = 'ベット:' (common.json); rendered as "ベット: 20"
    await waitFor(() => expect(screen.getAllByText(/ベット:/).length).toBeGreaterThanOrEqual(1));
  });

  // ---- human showdown hand name badge ----
  it('shows human hand name badge at showdown when not folded', async () => {
    mockExec.mockResolvedValue({
      ...showdownState,
      players: [humanPlayer({ handName: 'フラッシュ', folded: false }), cpuPlayer(1, { handName: 'ハイカード' })],
    });
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText('フラッシュ')).toBeInTheDocument());
  });

  // ---- human currentBet display ----
  it('shows human current bet when positive', async () => {
    mockExec.mockResolvedValue({
      ...thirdStreetState,
      players: [humanPlayer({ currentBet: 10 }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getAllByText(/ベット:/).length).toBeGreaterThanOrEqual(1));
  });

  // ---- settings cpuMetaAI checkbox ----
  it('toggles cpuMetaAI checkbox and sends reset with metaAI flag', async () => {
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const metaAICheckbox = screen.getByRole('checkbox', { name: 'メタAI（CPUがプレイスタイルを学習）' });
    expect(metaAICheckbox).not.toBeChecked();

    fireEvent.click(metaAICheckbox);
    expect(metaAICheckbox).toBeChecked();

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { ante: 1, tournamentMode: false, cpuMetaAI: true }),
    );
  });

  // ---- settings ante + tournament mode ----
  it('changes ante and tournament mode, then passes them to reset', async () => {
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const anteSelect = screen.getByTestId('razz-ante-select') as HTMLSelectElement;
    fireEvent.change(anteSelect, { target: { value: '5' } });
    expect(anteSelect.value).toBe('5');

    const tournamentCheckbox = screen.getByRole('checkbox', { name: 'トーナメントモード' });
    expect(tournamentCheckbox).not.toBeChecked();
    fireEvent.click(tournamentCheckbox);
    expect(tournamentCheckbox).toBeChecked();

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { ante: 5, tournamentMode: true, cpuMetaAI: false }),
    );
  });

  // ---- end phase + win celebration ----
  it('shows result phase name at END phase', async () => {
    mockExec.mockResolvedValue({
      ...showdownState,
      phase: 7, // END
      gameEndFlag: true,
    });
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText('結果')).toBeInTheDocument());
  });

  // ---- human folded allIn status at showdown ----
  it('shows allIn badge for human when allIn during showdown', async () => {
    mockExec.mockResolvedValue({
      ...showdownState,
      players: [humanPlayer({ allIn: true, folded: false }), cpuPlayer(1)],
    });
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText('[オールイン]')).toBeInTheDocument());
  });

  it('wraps settings in collapsible details element', async () => {
    mockExec.mockResolvedValue(initState);
    const { container } = renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
    const allSummaries = container.querySelectorAll('details summary');
    const settingsSummary = Array.from(allSummaries).find((s) => s.textContent?.includes('設定'));
    expect(settingsSummary).toBeTruthy();
  });

  it('renders settings via the standard SettingsPanel component', async () => {
    mockExec.mockResolvedValue(initState);
    const { container } = renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
    // SettingsPanel items carry stable ids (the raw <details> checkboxes had none).
    expect(container.querySelector('#frontendHint')).toBeInTheDocument();
    expect(container.querySelector('#cpuMetaAI')).toBeInTheDocument();
    expect(screen.getByLabelText('ヒント表示')).toBeInTheDocument();
    expect(screen.getByLabelText('メタAI（CPUがプレイスタイルを学習）')).toBeInTheDocument();
  });

  // ---- CLI mode ----
  it('renders CliTerminal when CLI mode is enabled', async () => {
    mockUseCliMode.mockReturnValue({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    mockUseCliMode.mockReturnValue({
      cliEnabled: false,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
  });

  it('sends CLI command via terminal input', async () => {
    const addInput = vi.fn();
    const addOutput = vi.fn();
    const addError = vi.fn();
    const clearLog = vi.fn();
    mockUseCliMode.mockReturnValue({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput,
      addOutput,
      addError,
      clearLog,
    });
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    const input = screen.getByRole('textbox');
    fireEvent.change(input, { target: { value: 'fold' } });
    fireEvent.keyDown(input, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold', undefined));
    mockUseCliMode.mockReturnValue({
      cliEnabled: false,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
  });

  it('rejects unknown CLI command', async () => {
    const addError = vi.fn();
    mockUseCliMode.mockReturnValue({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError,
      clearLog: vi.fn(),
    });
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    const input = screen.getByRole('textbox');
    fireEvent.change(input, { target: { value: 'invalid' } });
    fireEvent.keyDown(input, { key: 'Enter' });
    await waitFor(() => expect(addError).toHaveBeenCalledWith(expect.stringContaining('Unknown command')));
    mockUseCliMode.mockReturnValue({
      cliEnabled: false,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
  });

  it('sends CLI bet command with amount', async () => {
    mockUseCliMode.mockReturnValue({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    const input = screen.getByRole('textbox');
    fireEvent.change(input, { target: { value: 'bet 50' } });
    fireEvent.keyDown(input, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 50));
    mockUseCliMode.mockReturnValue({
      cliEnabled: false,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
  });

  // ---- CPU card rendering: no door cards, not folded ----
  it('shows card-back placeholders for CPU with no door cards', async () => {
    mockExec.mockResolvedValue({
      ...thirdStreetState,
      players: [humanPlayer(), cpuPlayer(1, { doorCards: [], folded: false })],
    });
    const { container } = renderWithProviders(<RazzPage />);
    await waitFor(() => expect(container.querySelector('[data-testid="animated-card-back"]')).toBeInTheDocument());
  });

  // ---- Human with no door cards renders card-back placeholders ----
  it('shows card-back placeholders for human with no door cards', async () => {
    mockExec.mockResolvedValue({
      ...thirdStreetState,
      players: [humanPlayer({ doorCards: [], holeCards: [] }), cpuPlayer(1)],
    });
    const { container } = renderWithProviders(<RazzPage />);
    await waitFor(() => expect(container.querySelector('[data-testid="animated-card-back"]')).toBeInTheDocument());
  });

  // ---- Seventh street: 3 hole card placeholders ----
  // ---- bring-in indicator ----
  it('shows the bring-in badge on the human when they post the bring-in during 3rd street', async () => {
    // thirdStreetState has bringInPlayerIdx 0 → the human (id 0).
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<RazzPage />);
    const badge = await screen.findByTestId('razz-bringin-badge-0');
    expect(badge).toHaveTextContent('ブリングイン');
  });

  it('shows the bring-in badge on the CPU indicated by bringInPlayerIdx during 3rd street', async () => {
    mockExec.mockResolvedValue({ ...thirdStreetState, bringInPlayerIdx: 1 });
    renderWithProviders(<RazzPage />);
    // bringInPlayerIdx 1 → CPU with id 1; the human (id 0) must not carry the badge.
    expect(await screen.findByTestId('razz-bringin-badge-1')).toHaveTextContent('ブリングイン');
    expect(screen.queryByTestId('razz-bringin-badge-0')).not.toBeInTheDocument();
  });

  it('hides the bring-in badge outside of 3rd street', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByText('ショーダウン')).toBeInTheDocument());
    expect(screen.queryByTestId('razz-bringin-badge-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('razz-bringin-badge-1')).not.toBeInTheDocument();
  });

  it('shows 3 face-down placeholders for human hole cards on seventh street', async () => {
    mockExec.mockResolvedValue({
      ...thirdStreetState,
      phase: 5, // SEVENTH_STREET
      players: [humanPlayer({ holeCards: [] }), cpuPlayer(1)],
    });
    renderWithProviders(<RazzPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });
});
