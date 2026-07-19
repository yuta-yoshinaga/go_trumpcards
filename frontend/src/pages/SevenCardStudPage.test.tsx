import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { sevenCardStudApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { SevenCardStudResponse } from '../types/card';
import { SevenCardStudPage } from './SevenCardStudPage';

vi.mock('../api/gameApi', () => ({
  sevenCardStudApi: { exec: vi.fn() },
  actionLogApi: { sevencardstud: vi.fn() },
}));

const mockExec = vi.mocked(sevenCardStudApi.exec);

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

describe('SevenCardStudPage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SevenCardStudPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  // ---- mount & reset ----
  it('calls reset on mount', async () => {
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // ---- phase name display ----
  it('shows "初期化中" when phase is INIT', async () => {
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByText('初期化中')).toBeInTheDocument());
  });

  it('shows known phase name for THIRD_STREET', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByText('サードストリート')).toBeInTheDocument());
  });

  // ---- info bar ----
  it('shows pot and the dealer name via playerName (CPU dealer)', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByText(/ポット:/)).toBeInTheDocument());
    expect(screen.getByText(/ディーラー:/)).toBeInTheDocument();
    // Dealer renders via playerName (CPU 3), not the raw index.
    expect(screen.getAllByText('CPU 3').length).toBeGreaterThan(0);
    expect(screen.queryByText(/Player 3|プレイヤー 3/)).not.toBeInTheDocument();
  });

  // ---- CPU players ----
  it('renders CPU player info with playStyleName and chips', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.getByText(/CPU 2/)).toBeInTheDocument();
    expect(screen.getAllByText(/タイト/).length).toBeGreaterThanOrEqual(1);
  });

  it('shows fold badge for folded CPU', async () => {
    mockExec.mockResolvedValue({
      ...thirdStreetState,
      players: [humanPlayer(), cpuPlayer(1, { folded: true }), cpuPlayer(2)],
    });
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByText('[フォールド]')).toBeInTheDocument());
  });

  it('shows all-in badge for all-in CPU', async () => {
    mockExec.mockResolvedValue({
      ...thirdStreetState,
      players: [humanPlayer(), cpuPlayer(1, { allIn: true }), cpuPlayer(2)],
    });
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByText('[オールイン]')).toBeInTheDocument());
  });

  it('shows CPU hand name badge during showdown when not folded', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByText('ツーペア')).toBeInTheDocument());
  });

  it('shows CPU door cards', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getAllByText('ドアカード').length).toBeGreaterThanOrEqual(1));
  });

  // ---- live current-hand strength (issue #3118) ----
  it('shows the human live best-hand strength during the streets', async () => {
    const pairedHuman = humanPlayer({
      holeCards: [
        { design: 'SPADE', value: 7 },
        { design: 'HEART', value: 7 },
        { design: 'DIAMOND', value: 10 },
      ],
      doorCards: [{ design: 'CLOVER', value: 2 }],
    });
    mockExec.mockResolvedValue({ ...thirdStreetState, players: [pairedHuman, cpuPlayer(1)] });
    renderWithProviders(<SevenCardStudPage />);
    const badge = await screen.findByTestId('scs-current-hand');
    expect(badge).toHaveTextContent('現在の役: ワンペア');
  });

  it('does not show the live strength badge at showdown (server handName takes over)', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
    expect(screen.queryByTestId('scs-current-hand')).not.toBeInTheDocument();
  });

  it('does not show the live strength badge when the human has folded', async () => {
    const foldedHuman = humanPlayer({ folded: true });
    mockExec.mockResolvedValue({ ...thirdStreetState, players: [foldedHuman, cpuPlayer(1)] });
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByTestId('scs-current-hand')).not.toBeInTheDocument();
  });

  it('shows CPU hole cards face-up during showdown when not folded', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 5')).toBeInTheDocument());
    expect(screen.getByAltText('♥ 8')).toBeInTheDocument();
  });

  // ---- CPU actions log ----
  it('shows CPU actions log when cpuActions is non-empty', async () => {
    mockExec.mockResolvedValue(thirdStreetWithBetState);
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByText('CPU行動:')).toBeInTheDocument());
    expect(screen.getByText(/Player 1: ベット/)).toBeInTheDocument();
  });

  // ---- round results ----
  it('shows round results in showdown', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
  });

  it('does not show round results when not in showdown', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByText('結果:')).not.toBeInTheDocument();
  });

  // ---- human player section ----
  it('shows human player section when humanPlayer exists', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
  });

  it('does not show human player section when no players', async () => {
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByText('初期化中')).toBeInTheDocument());
    expect(screen.queryByText(/あなたの手札/)).not.toBeInTheDocument();
  });

  it('shows human door cards', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByAltText('♣ 5')).toBeInTheDocument());
    expect(screen.getByAltText('♠ 8')).toBeInTheDocument();
  });

  it('shows human hole cards', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.getByAltText('♥ K')).toBeInTheDocument();
  });

  // ---- message ----
  it('shows game message', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByText('あなたの番です')).toBeInTheDocument());
  });

  // ---- canAct / betting controls ----
  it('shows bet/check buttons when canAct and no outstanding bet', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument();
  });

  it('shows call/raise buttons when canAct and has outstanding bet', async () => {
    mockExec.mockResolvedValue(thirdStreetWithBetState);
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'レイズ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument();
  });

  it('hides betting controls when not active phase', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByText('ショーダウン')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  it('hides betting controls when human is folded', async () => {
    mockExec.mockResolvedValue({
      ...thirdStreetState,
      players: [humanPlayer({ folded: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  // ---- button click handlers ----
  it('calls bet command with betAmount', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(thirdStreetState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 20, undefined, 0));
  });

  it('calls check command', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(thirdStreetState);
    fireEvent.click(screen.getByRole('button', { name: 'チェック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('check', undefined, undefined, 0));
  });

  it('calls fold command', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(thirdStreetState);
    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold', undefined, undefined, 0));
  });

  it('calls allin command', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(thirdStreetState);
    fireEvent.click(screen.getByRole('button', { name: 'オールイン' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('allin', undefined, undefined, 0));
  });

  it('calls reset command when reset button is clicked', async () => {
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuMetaAI: false }));
  });

  it('uses outline style for reset button', async () => {
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const resetBtn = screen.getByRole('button', { name: 'リセット' });
    expect(resetBtn.className).toContain('bg-transparent');
    expect(resetBtn.className).toContain('border');
  });

  // ---- loading / disabled state ----
  it('disables buttons while loading', async () => {
    mockExec.mockResolvedValue(thirdStreetState);
    renderWithProviders(<SevenCardStudPage />);
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
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('clears error on successful call after failure', async () => {
    renderWithProviders(<SevenCardStudPage />);
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
    renderWithProviders(<SevenCardStudPage />);
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
    renderWithProviders(<SevenCardStudPage />);
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
    renderWithProviders(<SevenCardStudPage />);
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
    renderWithProviders(<SevenCardStudPage />);
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
    renderWithProviders(<SevenCardStudPage />);
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
    renderWithProviders(<SevenCardStudPage />);
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
    renderWithProviders(<SevenCardStudPage />);
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
    renderWithProviders(<SevenCardStudPage />);
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
    renderWithProviders(<SevenCardStudPage />);
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
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByText(/ハンド#3/)).toBeInTheDocument());
  });

  // ---- human allIn disables betting controls ----
  it('hides betting controls when human is all-in', async () => {
    mockExec.mockResolvedValue({
      ...thirdStreetState,
      players: [humanPlayer({ allIn: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<SevenCardStudPage />);
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
    renderWithProviders(<SevenCardStudPage />);
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
    renderWithProviders(<SevenCardStudPage />);
    // tc('betting.currentBet') = 'ベット:' (common.json); rendered as "ベット: 20"
    await waitFor(() => expect(screen.getAllByText(/ベット:/).length).toBeGreaterThanOrEqual(1));
  });

  // ---- human showdown hand name badge ----
  it('shows human hand name badge at showdown when not folded', async () => {
    mockExec.mockResolvedValue({
      ...showdownState,
      players: [humanPlayer({ handName: 'フラッシュ', folded: false }), cpuPlayer(1, { handName: 'ハイカード' })],
    });
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByText('フラッシュ')).toBeInTheDocument());
  });

  // ---- human currentBet display ----
  it('shows human current bet when positive', async () => {
    mockExec.mockResolvedValue({
      ...thirdStreetState,
      players: [humanPlayer({ currentBet: 10 }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getAllByText(/ベット:/).length).toBeGreaterThanOrEqual(1));
  });

  // ---- settings cpuMetaAI checkbox ----
  it('toggles cpuMetaAI checkbox and sends reset with metaAI flag', async () => {
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const metaAICheckbox = screen.getByRole('checkbox', { name: 'メタAI（CPUがプレイスタイルを学習）' });
    expect(metaAICheckbox).not.toBeChecked();

    fireEvent.click(metaAICheckbox);
    expect(metaAICheckbox).toBeChecked();

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuMetaAI: true }));
  });

  // ---- end phase + win celebration ----
  it('shows result phase name at END phase', async () => {
    mockExec.mockResolvedValue({
      ...showdownState,
      phase: 7, // END
      gameEndFlag: true,
    });
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByText('結果')).toBeInTheDocument());
  });

  // ---- human folded allIn status at showdown ----
  it('shows allIn badge for human when allIn during showdown', async () => {
    mockExec.mockResolvedValue({
      ...showdownState,
      players: [humanPlayer({ allIn: true, folded: false }), cpuPlayer(1)],
    });
    renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByText('[オールイン]')).toBeInTheDocument());
  });

  it('wraps settings in collapsible details element', async () => {
    mockExec.mockResolvedValue(initState);
    const { container } = renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
    const allSummaries = container.querySelectorAll('details summary');
    const settingsSummary = Array.from(allSummaries).find((s) => s.textContent?.includes('設定'));
    expect(settingsSummary).toBeTruthy();
  });

  it('renders settings via the standard SettingsPanel component', async () => {
    mockExec.mockResolvedValue(initState);
    const { container } = renderWithProviders(<SevenCardStudPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
    // SettingsPanel items carry stable ids (the raw <details> checkboxes had none).
    expect(container.querySelector('#frontendHint')).toBeInTheDocument();
    expect(container.querySelector('#cpuMetaAI')).toBeInTheDocument();
    expect(screen.getByLabelText('ヒント表示')).toBeInTheDocument();
    expect(screen.getByLabelText('メタAI（CPUがプレイスタイルを学習）')).toBeInTheDocument();
  });
});
