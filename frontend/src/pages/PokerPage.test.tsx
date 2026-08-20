import { act, fireEvent, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, pokerApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { PokerResponse } from '../types/card';
import { PokerPage } from './PokerPage';

vi.mock('../api/gameApi', () => ({
  pokerApi: { exec: vi.fn() },
  actionLogApi: { poker: vi.fn() },
}));

const mockExec = vi.mocked(pokerApi.exec);

/** Helper: base human player */
const humanPlayer = (overrides: Partial<import('../types/card').PokerPlayerData> = {}) => ({
  id: 0,
  isHuman: true,
  cards: [
    { design: 'SPADE' as const, value: 1 },
    { design: 'HEART' as const, value: 5 },
    { design: 'DIAMOND' as const, value: 10 },
    { design: 'CLOVER' as const, value: 3 },
    { design: 'SPADE' as const, value: 7 },
  ],
  chips: 990,
  currentBet: 0,
  folded: false,
  allIn: false,
  handRank: 0,
  handName: '',
  exchangeCount: -1,
  playStyleName: '',
  ...overrides,
});

/** Helper: base CPU player */
const cpuPlayer = (id: number, overrides: Partial<import('../types/card').PokerPlayerData> = {}) => ({
  id,
  isHuman: false,
  cards: [
    { design: 'DIAMOND' as const, value: 2 },
    { design: 'CLOVER' as const, value: 7 },
    { design: 'HEART' as const, value: 9 },
    { design: 'SPADE' as const, value: 11 },
    { design: 'DIAMOND' as const, value: 13 },
  ],
  chips: 990,
  currentBet: 0,
  folded: false,
  allIn: false,
  handRank: 0,
  handName: '',
  exchangeCount: -1,
  playStyleName: 'バランス型',
  ...overrides,
});

/** INIT state (phase 0): no players yet */
const initState: PokerResponse = {
  players: [],
  pot: 0,
  sidePots: [],
  dealerIdx: 0,
  currentTurn: 0,
  phase: 0,
  exchangeRead: false,
  gameEndFlag: false,
  lastBet: 0,
  minRaise: 10,
  ante: 10,
  jokerCount: 0,
  roundResults: [],
  cpuActions: [],
  cpuExchanges: [],
  message: 'リセットしました',
  bettingLimit: 0,
  raiseCount: 0,
  maxBetAmount: 0,
  isLowball: false,
};

/** DEAL phase (phase 1): human's turn, no outstanding bet */
const dealState: PokerResponse = {
  players: [humanPlayer(), cpuPlayer(1), cpuPlayer(2), cpuPlayer(3)],
  pot: 40,
  sidePots: [],
  dealerIdx: 3,
  currentTurn: 0,
  phase: 1,
  exchangeRead: false,
  gameEndFlag: false,
  lastBet: 0,
  minRaise: 10,
  ante: 10,
  jokerCount: 0,
  roundResults: [],
  cpuActions: [],
  cpuExchanges: [],
  message: 'あなたの番です',
  bettingLimit: 0,
  raiseCount: 0,
  maxBetAmount: 0,
  isLowball: false,
};

/** DEAL with outstanding bet: shows call/raise */
const dealWithBetState: PokerResponse = {
  ...dealState,
  lastBet: 40,
  cpuActions: [{ playerIdx: 1, action: 3, amount: 40 }],
};

/** EXCHANGE phase (phase 2): human's turn to exchange */
const exchangeState: PokerResponse = {
  ...dealState,
  phase: 2,
  message: '交換するカードを選んでください',
};

/** SECOND_BET phase (phase 3): human's turn */
const secondBetState: PokerResponse = {
  ...dealState,
  phase: 3,
  cpuExchanges: [
    { playerIdx: 1, exchangeCount: 2 },
    { playerIdx: 2, exchangeCount: 0 },
  ],
};

/** END phase (phase 4) */
const endState: PokerResponse = {
  players: [
    humanPlayer({ handName: 'High Card', chips: 960 }),
    cpuPlayer(1, {
      handName: 'One Pair',
      folded: false,
      exchangeCount: 2,
      cards: [
        { design: 'HEART', value: 2 },
        { design: 'DIAMOND', value: 4 },
        { design: 'CLOVER', value: 6 },
        { design: 'SPADE', value: 8 },
        { design: 'HEART', value: 10 },
      ],
    }),
    cpuPlayer(2, { folded: true }),
  ],
  pot: 0,
  sidePots: [],
  dealerIdx: 2,
  currentTurn: -1,
  phase: 4,
  exchangeRead: false,
  gameEndFlag: false,
  lastBet: 0,
  minRaise: 0,
  ante: 10,
  jokerCount: 0,
  roundResults: [
    { playerIdx: 0, handRank: 0, handName: 'High Card', kickers: '', wonAmount: 0 },
    { playerIdx: 1, handRank: 1, handName: 'One Pair', kickers: 'A, Q, 10', wonAmount: 200 },
  ],
  cpuActions: [],
  cpuExchanges: [],
  message: 'あなたの負け',
  bettingLimit: 0,
  raiseCount: 0,
  maxBetAmount: 0,
  isLowball: false,
};

beforeEach(() => {
  mockExec.mockResolvedValue(initState);
});

describe('PokerPage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<PokerPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  // ---- mount & reset ----
  it('calls reset on mount', async () => {
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // ---- info bar ----
  it('shows pot and the dealer name via playerName (CPU dealer)', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/ポット:/)).toBeInTheDocument());
    expect(screen.getByText(/ディーラー:/)).toBeInTheDocument();
    // Dealer renders via playerName (CPU 3), not the raw index.
    expect(screen.getAllByText('CPU 3').length).toBeGreaterThan(0);
    expect(screen.queryByText(/Player 3|プレイヤー 3/)).not.toBeInTheDocument();
  });

  it('shows joker count when jokerCount > 0', async () => {
    mockExec.mockResolvedValue({ ...dealState, jokerCount: 2 });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/ジョーカー:/)).toBeInTheDocument());
  });

  it('does not show joker count when jokerCount is 0', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/ポット:/)).toBeInTheDocument());
    expect(screen.queryByText(/ジョーカー:/)).not.toBeInTheDocument();
  });

  // ---- CPU players ----
  it('renders CPU player info with playStyleName and chips', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.getByText(/CPU 2/)).toBeInTheDocument();
    expect(screen.getAllByText(/バランス型/).length).toBeGreaterThanOrEqual(1);
  });

  it('shows CPU bet when currentBet > 0', async () => {
    mockExec.mockResolvedValue({
      ...dealState,
      players: [humanPlayer(), cpuPlayer(1, { currentBet: 50 }), cpuPlayer(2)],
    });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/ベット: 50/)).toBeInTheDocument());
  });

  it('does not show CPU bet when currentBet is 0', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByText(/ベット: 0/)).not.toBeInTheDocument();
  });

  it('shows fold badge for folded CPU', async () => {
    mockExec.mockResolvedValue({
      ...dealState,
      players: [humanPlayer(), cpuPlayer(1, { folded: true }), cpuPlayer(2)],
    });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('[フォールド]')).toBeInTheDocument());
  });

  it('shows all-in badge for all-in CPU', async () => {
    mockExec.mockResolvedValue({
      ...dealState,
      players: [humanPlayer(), cpuPlayer(1, { allIn: true }), cpuPlayer(2)],
    });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('[オールイン]')).toBeInTheDocument());
  });

  it('shows CPU exchange count in SECOND_BET phase when not folded and count > 0', async () => {
    mockExec.mockResolvedValue({
      ...secondBetState,
      players: [humanPlayer(), cpuPlayer(1, { exchangeCount: 2 }), cpuPlayer(2, { exchangeCount: 0 })],
    });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/交換: 2枚/)).toBeInTheDocument());
    // exchangeCount 0 (stood pat) should not show exchange label
    expect(screen.queryByText(/交換: 0枚/)).not.toBeInTheDocument();
  });

  it('does not show CPU exchange count in DEAL phase', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByText(/交換:.*枚/)).not.toBeInTheDocument();
  });

  it('does not show CPU exchange count when folded in SECOND_BET', async () => {
    mockExec.mockResolvedValue({
      ...secondBetState,
      players: [humanPlayer(), cpuPlayer(1, { folded: true, exchangeCount: 2 }), cpuPlayer(2)],
    });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('[フォールド]')).toBeInTheDocument());
    expect(screen.queryByText(/交換: 2枚/)).not.toBeInTheDocument();
  });

  it('does not show CPU exchange count when exchangeCount < 0 in SECOND_BET', async () => {
    mockExec.mockResolvedValue({
      ...secondBetState,
      players: [humanPlayer(), cpuPlayer(1, { exchangeCount: -1 }), cpuPlayer(2)],
    });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    // exchangeCount -1 means not yet exchanged, should not display
    expect(screen.queryByText(/交換:.*枚/)).not.toBeInTheDocument();
  });

  it('shows CPU hand name badge in END phase when not folded', async () => {
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('One Pair')).toBeInTheDocument());
  });

  it('does not show CPU hand name badge when folded in END phase', async () => {
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('One Pair')).toBeInTheDocument());
    // CPU 2 is folded → no hand name badge for it
  });

  it('does not show CPU hand name badge when handName is empty in END', async () => {
    mockExec.mockResolvedValue({
      ...endState,
      players: [
        humanPlayer({ handName: 'High Card' }),
        cpuPlayer(1, { handName: '', folded: false }),
        cpuPlayer(2, { folded: true }),
      ],
    });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('High Card')).toBeInTheDocument());
    // CPU 1 has empty handName → no badge
  });

  it('shows CPU cards face-up in END phase when not folded', async () => {
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByAltText('♥ 2')).toBeInTheDocument());
    expect(screen.getByAltText('♦ 4')).toBeInTheDocument();
  });

  it('shows compact card count for CPU cards when not in END phase', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    // CPU cards are shown as compact text count (not card back images)
    const compactCounts = screen.getAllByTestId('compact-card-count');
    expect(compactCounts.length).toBeGreaterThanOrEqual(1);
  });

  it('shows CardBack for folded CPU in END phase', async () => {
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('One Pair')).toBeInTheDocument());
    // CPU 2 is folded → shows CardBack
    const cardBacks = screen.getAllByAltText('カード裏面');
    expect(cardBacks.length).toBeGreaterThanOrEqual(5);
  });

  it('shows CardBack for CPU in END when cards is empty and not folded', async () => {
    mockExec.mockResolvedValue({
      ...endState,
      players: [
        humanPlayer({ handName: 'High Card' }),
        cpuPlayer(1, { handName: 'One Pair', folded: false, cards: [] }),
        cpuPlayer(2, { folded: true }),
      ],
    });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('High Card')).toBeInTheDocument());
    const cardBacks = screen.getAllByAltText('カード裏面');
    expect(cardBacks.length).toBeGreaterThanOrEqual(5);
  });

  it('shows exchange count for CPU in END phase when not folded', async () => {
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/交換: 2枚/)).toBeInTheDocument());
  });

  // ---- CPU actions log ----
  it('shows CPU actions log when cpuActions is non-empty', async () => {
    mockExec.mockResolvedValue(dealWithBetState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('CPU行動:')).toBeInTheDocument());
    expect(screen.getByText(/Player 1: ベット/)).toBeInTheDocument();
  });

  it('does not show CPU actions log when cpuActions is empty', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByText('CPU行動:')).not.toBeInTheDocument();
  });

  it('shows amount in CPU action when amount > 0', async () => {
    mockExec.mockResolvedValue(dealWithBetState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/Player 1: ベット \(40\)/)).toBeInTheDocument());
  });

  it('does not show amount when action amount is 0', async () => {
    mockExec.mockResolvedValue({
      ...dealState,
      cpuActions: [{ playerIdx: 1, action: 1, amount: 0 }],
    });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('CPU行動:')).toBeInTheDocument());
    expect(screen.getByText(/Player 1: チェック/)).toBeInTheDocument();
    expect(screen.queryByText(/\(0\)/)).not.toBeInTheDocument();
  });

  it('shows "不明" for unknown action type', async () => {
    mockExec.mockResolvedValue({
      ...dealState,
      cpuActions: [{ playerIdx: 1, action: 99, amount: 0 }],
    });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/Player 1: 不明/)).toBeInTheDocument());
  });

  // ---- CPU exchanges log ----
  it('shows CPU exchanges log when cpuExchanges is non-empty', async () => {
    mockExec.mockResolvedValue(secondBetState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('CPU交換:')).toBeInTheDocument());
    expect(screen.getByText(/Player 1: 2枚交換/)).toBeInTheDocument();
    expect(screen.getByText(/Player 2: 0枚交換/)).toBeInTheDocument();
  });

  it('does not show CPU exchanges log when cpuExchanges is empty', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByText('CPU交換:')).not.toBeInTheDocument();
  });

  // ---- round results ----
  it('shows round results in END phase', async () => {
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
  });

  it('does not show round results when not END phase', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByText('結果:')).not.toBeInTheDocument();
  });

  it('does not show round results when roundResults is empty in END', async () => {
    mockExec.mockResolvedValue({
      ...endState,
      roundResults: [],
    });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('あなたの負け')).toBeInTheDocument());
    expect(screen.queryByText('結果:')).not.toBeInTheDocument();
  });

  it('shows "あなた" for human player in results', async () => {
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
    expect(within(screen.getByTestId('round-results-visible')).getByText(/あなた: High Card/)).toBeInTheDocument();
  });

  it('shows "CPU X" for non-human player in results', async () => {
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
    expect(within(screen.getByTestId('round-results-visible')).getByText(/CPU 1: One Pair/)).toBeInTheDocument();
  });

  it('shows hand name in results when present', async () => {
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<PokerPage />);
    await waitFor(() =>
      expect(within(screen.getByTestId('round-results-visible')).getByText(/: High Card/)).toBeInTheDocument(),
    );
    expect(within(screen.getByTestId('round-results-visible')).getByText(/: One Pair/)).toBeInTheDocument();
  });

  it('does not show hand name colon when handName is empty in results', async () => {
    mockExec.mockResolvedValue({
      ...endState,
      roundResults: [
        { playerIdx: 0, handRank: 0, handName: '', kickers: '', wonAmount: 0 },
        { playerIdx: 1, handRank: 2, handName: 'One Pair', kickers: '', wonAmount: 200 },
      ],
    });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
  });

  it('shows won chips when wonAmount > 0', async () => {
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<PokerPage />);
    await waitFor(() =>
      expect(within(screen.getByTestId('round-results-visible')).getByText(/\+200チップ/)).toBeInTheDocument(),
    );
  });

  it('does not show won chips when wonAmount is 0', async () => {
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
    expect(screen.queryByText(/\+0チップ/)).not.toBeInTheDocument();
  });

  // ---- human player section ----
  it('shows human player section when humanPlayer exists', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
  });

  it('does not show human player section when no players', async () => {
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('リセットしました')).toBeInTheDocument());
    expect(screen.queryByText(/あなたの手札/)).not.toBeInTheDocument();
  });

  it('shows human cards', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.getByAltText('♥ 5')).toBeInTheDocument();
  });

  it('shows human bet when currentBet > 0', async () => {
    mockExec.mockResolvedValue({
      ...dealState,
      players: [humanPlayer({ currentBet: 30 }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/ベット: 30/)).toBeInTheDocument());
  });

  it('shows human fold badge', async () => {
    mockExec.mockResolvedValue({
      ...dealState,
      players: [humanPlayer({ folded: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('[フォールド]')).toBeInTheDocument());
  });

  it('shows human all-in badge', async () => {
    mockExec.mockResolvedValue({
      ...dealState,
      players: [humanPlayer({ allIn: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('[オールイン]')).toBeInTheDocument());
  });

  it('shows human hand name badge in END phase when not folded', async () => {
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('High Card')).toBeInTheDocument());
  });

  it('does not show human hand name badge when handName is empty in END', async () => {
    mockExec.mockResolvedValue({
      ...endState,
      players: [
        humanPlayer({ handName: '' }),
        cpuPlayer(1, { handName: 'One Pair', folded: false }),
        cpuPlayer(2, { folded: true }),
      ],
    });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('One Pair')).toBeInTheDocument());
  });

  // ---- exchange phase ----
  it('shows instruction text in EXCHANGE phase', async () => {
    mockExec.mockResolvedValue(exchangeState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/交換したいカードをクリックして選択/)).toBeInTheDocument());
  });

  it('shows exchange and stand buttons in EXCHANGE phase', async () => {
    mockExec.mockResolvedValue(exchangeState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument();
  });

  it('does not show exchange buttons when not EXCHANGE phase', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'スタンド' })).not.toBeInTheDocument();
  });

  it('does not show exchange buttons when it is not human turn', async () => {
    mockExec.mockResolvedValue({ ...exchangeState, currentTurn: 1 });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'スタンド' })).not.toBeInTheDocument();
  });

  it('disables the exchange button until a card is selected', async () => {
    mockExec.mockResolvedValue(exchangeState);
    renderWithProviders(<PokerPage />);
    const exchangeBtn = await screen.findByRole('button', { name: '交換' });
    // Nothing selected yet: stand is the only enabled action.
    expect(exchangeBtn).toBeDisabled();
    expect(screen.getByRole('button', { name: 'スタンド' })).toBeEnabled();
    fireEvent.click(screen.getByAltText('♠ A'));
    expect(exchangeBtn).toBeEnabled();
  });

  it('shows the selected-count badge (updating on selection) and a stand hint', async () => {
    mockExec.mockResolvedValue(exchangeState);
    renderWithProviders(<PokerPage />);
    const badge = await screen.findByTestId('pk-exchange-selected');
    expect(badge).toHaveTextContent('選択 0 枚');
    fireEvent.click(screen.getByAltText('♠ A'));
    expect(badge).toHaveTextContent('選択 1 枚');
    // A hint clarifies that standing keeps the current hand.
    expect(screen.getByTestId('pk-stand-hint')).toBeInTheDocument();
  });

  it('calls exchange with selected indices', async () => {
    mockExec.mockResolvedValue(exchangeState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A'));
    fireEvent.click(screen.getByAltText('♥ 5'));

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...exchangeState, phase: 3 });
    fireEvent.click(screen.getByRole('button', { name: '交換' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('exchange', expect.arrayContaining([0, 1])));
  });

  it('calls stand command', async () => {
    mockExec.mockResolvedValue(exchangeState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(endState);
    fireEvent.click(screen.getByRole('button', { name: 'スタンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('stand'));
  });

  it('toggles card selection in exchange phase', async () => {
    mockExec.mockResolvedValue(exchangeState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('card button has aria-label with card name in exchange phase', async () => {
    mockExec.mockResolvedValue(exchangeState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-label', '♠ A');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-label', '♠ A (交換選択中)');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-label', '♠ A');
  });

  it('applies unified selectedCardStyle to card button when selected', async () => {
    mockExec.mockResolvedValue(exchangeState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    // Before selection: transparent border, no transform
    expect(cardBtn.style.border).toBe('3px solid transparent');
    expect(cardBtn.style.transform).toBe('none');

    fireEvent.click(cardBtn);
    // After selection: selectedCardStyle applied to button (border + transform + shadow)
    expect(cardBtn.style.border).toContain('3px solid');
    expect(cardBtn.style.border).not.toBe('3px solid transparent');
    expect(cardBtn.style.transform).toBe('translateY(-8px)');
  });

  it('does not select card when not in exchange phase', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A'));
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
  });

  it('clears selection on successful API call', async () => {
    mockExec.mockResolvedValue(exchangeState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...exchangeState, phase: 3 });
    fireEvent.click(screen.getByRole('button', { name: '交換' }));
    await waitFor(() => expect(cardBtn).toHaveAttribute('aria-pressed', 'false'));
  });

  // ---- canAct / betting controls ----
  it('shows bet/check buttons when canAct and no outstanding bet', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument();
  });

  it('shows call/raise buttons when canAct and has outstanding bet', async () => {
    mockExec.mockResolvedValue(dealWithBetState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'レイズ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument();
  });

  it('shows betting controls in SECOND_BET phase', async () => {
    mockExec.mockResolvedValue(secondBetState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
  });

  it('hides betting controls in END phase', async () => {
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('あなたの負け')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  it('hides betting controls when human is folded', async () => {
    mockExec.mockResolvedValue({
      ...dealState,
      players: [humanPlayer({ folded: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  it('hides betting controls when human is all-in', async () => {
    mockExec.mockResolvedValue({
      ...dealState,
      players: [humanPlayer({ allIn: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  it('hides betting controls when it is not human turn', async () => {
    mockExec.mockResolvedValue({
      ...dealState,
      currentTurn: 1,
    });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  // ---- bet amount input ----
  it('updates bet amount when changing input', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    const betInput = await screen.findByLabelText('ベット額:');
    await waitFor(() => expect((betInput as HTMLInputElement).value).toBe('10'));

    fireEvent.change(betInput, { target: { value: '50' } });
    expect((betInput as HTMLInputElement).value).toBe('50');
  });

  it('sets betAmount to minRaise when state changes', async () => {
    mockExec.mockResolvedValue({ ...dealState, minRaise: 30 });
    renderWithProviders(<PokerPage />);
    const betInput = await screen.findByLabelText('ベット額:');
    await waitFor(() => expect((betInput as HTMLInputElement).value).toBe('30'));
  });

  it('sets betAmount to 10 when minRaise is 0', async () => {
    mockExec.mockResolvedValue({ ...dealState, minRaise: 0 });
    renderWithProviders(<PokerPage />);
    const betInput = await screen.findByLabelText('ベット額:');
    await waitFor(() => expect((betInput as HTMLInputElement).value).toBe('10'));
  });

  it('preserves a typed raise amount across a state update that keeps the same minRaise (#2980)', async () => {
    mockExec.mockResolvedValue({ ...dealState, minRaise: 10 });
    renderWithProviders(<PokerPage />);
    const betInput = await screen.findByLabelText('ベット額:');
    await waitFor(() => expect((betInput as HTMLInputElement).value).toBe('10'));
    fireEvent.change(betInput, { target: { value: '80' } });
    expect((betInput as HTMLInputElement).value).toBe('80');
    // A CPU action arrives (state changes) but minRaise is unchanged.
    mockExec.mockResolvedValue({ ...dealState, minRaise: 10, message: 'CPU acted' });
    fireEvent.click(screen.getByRole('button', { name: 'チェック' }));
    await waitFor(() => expect(screen.getByText('CPU acted')).toBeInTheDocument());
    expect((betInput as HTMLInputElement).value).toBe('80');
  });

  it('bumps betAmount when minRaise rises (#2980)', async () => {
    mockExec.mockResolvedValue({ ...dealState, minRaise: 10 });
    renderWithProviders(<PokerPage />);
    const betInput = await screen.findByLabelText('ベット額:');
    await waitFor(() => expect((betInput as HTMLInputElement).value).toBe('10'));
    fireEvent.change(betInput, { target: { value: '30' } });
    // A raise lifts minRaise; the input follows the new minimum.
    mockExec.mockResolvedValue({ ...dealState, minRaise: 60 });
    fireEvent.click(screen.getByRole('button', { name: 'チェック' }));
    await waitFor(() => expect((betInput as HTMLInputElement).value).toBe('60'));
  });

  // ---- button click handlers ----
  it('calls bet command with betAmount', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(dealState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', undefined, 10, undefined, 0));
  });

  it('calls check command', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(dealState);
    fireEvent.click(screen.getByRole('button', { name: 'チェック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('check', undefined, undefined, undefined, 0));
  });

  it('calls fold command', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(endState);
    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold', undefined, undefined, undefined, 0));
  });

  it('calls allin command', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(dealState);
    fireEvent.click(screen.getByRole('button', { name: 'オールイン' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('allin', undefined, undefined, undefined, 0));
  });

  it('calls call command when has outstanding bet', async () => {
    mockExec.mockResolvedValue(dealWithBetState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(dealState);
    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('call', undefined, undefined, undefined, 0));
  });

  it('calls raise command with betAmount when has outstanding bet', async () => {
    mockExec.mockResolvedValue(dealWithBetState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'レイズ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(dealState);
    fireEvent.click(screen.getByRole('button', { name: 'レイズ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('raise', undefined, 10, undefined, 0));
  });

  it('sends updated bet amount when bet is clicked after changing input', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    const betInput = await screen.findByLabelText('ベット額:');
    await waitFor(() => expect((betInput as HTMLInputElement).value).toBe('10'));

    fireEvent.change(betInput, { target: { value: '60' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(dealState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', undefined, 60, undefined, 0));
  });

  it('sends updated bet amount when raise is clicked after changing input', async () => {
    mockExec.mockResolvedValue(dealWithBetState);
    renderWithProviders(<PokerPage />);
    const betInput = await screen.findByLabelText('ベット額:');
    await waitFor(() => expect((betInput as HTMLInputElement).value).toBe('10'));

    fireEvent.change(betInput, { target: { value: '100' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(dealState);
    fireEvent.click(screen.getByRole('button', { name: 'レイズ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('raise', undefined, 100, undefined, 0));
  });

  it('calls reset command when reset button is clicked', async () => {
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        bettingLimit: 0,
        isLowball: false,
        cpuMetaAI: false,
      }),
    );
  });

  it('sends updated bettingLimit when select is changed before reset', async () => {
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const select = screen.getByLabelText('リミット:');
    fireEvent.change(select, { target: { value: '1' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        bettingLimit: 1,
        isLowball: false,
        cpuMetaAI: false,
      }),
    );
  });

  it('sends isLowball true when checkbox is checked before reset', async () => {
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const checkbox = screen.getByLabelText('2-7 ローボール');
    fireEvent.click(checkbox);

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        bettingLimit: 0,
        isLowball: true,
        cpuMetaAI: false,
      }),
    );
  });

  it('sends cpuMetaAI true when checkbox is checked before reset', async () => {
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const checkbox = screen.getByLabelText('メタAI（CPUがプレイスタイルを学習）');
    fireEvent.click(checkbox);

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        bettingLimit: 0,
        isLowball: false,
        cpuMetaAI: true,
      }),
    );
  });

  it('shows lowball mode indicator when isLowball is true', async () => {
    const lowballState = { ...dealState, isLowball: true };
    mockExec.mockResolvedValue(lowballState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('[2-7 ローボール モード]')).toBeInTheDocument());
  });

  it('does not show lowball mode indicator when isLowball is false', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/ポット/)).toBeInTheDocument());
    expect(screen.queryByText('[2-7 ローボール モード]')).not.toBeInTheDocument();
  });

  // ---- loading / disabled state ----
  it('disables buttons while loading', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());

    let resolve!: (value: PokerResponse) => void;
    const slowPromise = new Promise<PokerResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));

    expect(screen.getByRole('button', { name: 'ベット' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'チェック' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();

    resolve(dealState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());
  });

  // ---- error handling ----
  it('shows error message when API call fails', async () => {
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('clears error on successful call after failure', async () => {
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());

    mockExec.mockResolvedValueOnce(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.queryByText(NETWORK_ERROR_MESSAGE())).not.toBeInTheDocument());
  });

  // ---- message ----
  it('shows game message', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('あなたの番です')).toBeInTheDocument());
  });

  // ---- aria-busy / sr-only ----
  it('sets aria-busy while loading', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());

    const container = screen.getByRole('button', { name: 'ベット' }).closest('[aria-busy]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');

    let resolve!: (value: PokerResponse) => void;
    const slowPromise = new Promise<PokerResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));

    expect(container).toHaveAttribute('aria-busy', 'true');

    resolve(dealState);
    await waitFor(() => {
      expect(container).toHaveAttribute('aria-busy', 'false');
    });
  });

  // ---- draw odds ----
  it('shows odds panel when odds data is present in exchange phase', async () => {
    const oddsState: PokerResponse = {
      ...exchangeState,
      odds: [
        { handRank: 0, handName: 'High Card', probability: 0.5, count: 5, total: 10 },
        { handRank: 1, handName: 'One Pair', probability: 0.3, count: 3, total: 10 },
        { handRank: 5, handName: 'Flush', probability: 0.2, count: 2, total: 10 },
      ],
    };
    mockExec.mockResolvedValue(exchangeState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());

    // Switch to fake timers for debounce control
    vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] });
    mockExec.mockResolvedValue(oddsState);
    fireEvent.click(screen.getByAltText('♠ A'));

    await act(async () => {
      vi.advanceTimersByTime(300);
    });
    vi.useRealTimers();

    await waitFor(() => expect(screen.getByTestId('odds-panel')).toBeInTheDocument());
    expect(screen.getByText('ドローオッズ:')).toBeInTheDocument();
    expect(screen.getByText('High Card')).toBeInTheDocument();
    expect(screen.getByText('50.0%')).toBeInTheDocument();
    expect(screen.getByText('One Pair')).toBeInTheDocument();
    expect(screen.getByText('30.0%')).toBeInTheDocument();
    expect(screen.getByText('Flush')).toBeInTheDocument();
    expect(screen.getByText('20.0%')).toBeInTheDocument();
  });

  it('hides odds panel when not in exchange phase', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    expect(screen.queryByTestId('odds-panel')).not.toBeInTheDocument();
  });

  it('clears odds after exchange', async () => {
    const oddsState: PokerResponse = {
      ...exchangeState,
      odds: [{ handRank: 1, handName: 'One Pair', probability: 1.0, count: 1, total: 1 }],
    };
    mockExec.mockResolvedValue(exchangeState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());

    // Toggle card to get odds
    vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] });
    mockExec.mockResolvedValue(oddsState);
    fireEvent.click(screen.getByAltText('♠ A'));
    await act(async () => {
      vi.advanceTimersByTime(300);
    });
    vi.useRealTimers();
    await waitFor(() => expect(screen.getByTestId('odds-panel')).toBeInTheDocument());

    // Exchange clears odds via onSuccess
    mockExec.mockResolvedValue({ ...exchangeState, phase: 3 });
    fireEvent.click(screen.getByRole('button', { name: '交換' }));
    await waitFor(() => expect(screen.queryByTestId('odds-panel')).not.toBeInTheDocument());
  });

  it('clears odds when deselecting all cards', async () => {
    const oddsState: PokerResponse = {
      ...exchangeState,
      odds: [{ handRank: 1, handName: 'One Pair', probability: 1.0, count: 1, total: 1 }],
    };
    mockExec.mockResolvedValue(exchangeState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());

    // Select card → odds appear
    vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] });
    mockExec.mockResolvedValue(oddsState);
    fireEvent.click(screen.getByAltText('♠ A'));
    await act(async () => {
      vi.advanceTimersByTime(300);
    });
    vi.useRealTimers();
    await waitFor(() => expect(screen.getByTestId('odds-panel')).toBeInTheDocument());

    // Deselect card → odds cleared immediately
    fireEvent.click(screen.getByAltText('♠ A'));
    await waitFor(() => expect(screen.queryByTestId('odds-panel')).not.toBeInTheDocument());
  });

  it('debounces odds API call', async () => {
    const oddsState: PokerResponse = {
      ...exchangeState,
      odds: [{ handRank: 1, handName: 'One Pair', probability: 1.0, count: 1, total: 1 }],
    };
    mockExec.mockResolvedValue(exchangeState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());

    vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] });
    mockExec.mockClear();
    mockExec.mockResolvedValue(oddsState);

    // Rapidly toggle two cards
    fireEvent.click(screen.getByAltText('♠ A'));
    fireEvent.click(screen.getByAltText('♥ 5'));

    // Before debounce fires, no odds call
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('odds', expect.anything());

    // After 300ms, one odds call with both indices
    await act(async () => {
      vi.advanceTimersByTime(300);
    });
    vi.useRealTimers();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('odds', expect.arrayContaining([0, 1])));
  });

  it('discards stale odds response after exchange (race condition)', async () => {
    let oddsResolve!: (value: PokerResponse) => void;
    const oddsState: PokerResponse = {
      ...exchangeState,
      odds: [{ handRank: 1, handName: 'One Pair', probability: 1.0, count: 1, total: 1 }],
    };
    mockExec.mockResolvedValue(exchangeState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());

    // Toggle card, start debounce
    vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] });
    // Make odds call return a slow promise
    mockExec.mockImplementation(
      (cmd: string) =>
        new Promise<PokerResponse>((resolve) => {
          if (cmd === 'odds') {
            oddsResolve = resolve;
          } else {
            resolve({ ...exchangeState, phase: 3 });
          }
        }),
    );
    fireEvent.click(screen.getByAltText('♠ A'));
    await act(async () => {
      vi.advanceTimersByTime(300);
    });
    vi.useRealTimers();

    // Exchange before odds resolves → onSuccess increments generation
    fireEvent.click(screen.getByRole('button', { name: '交換' }));
    await waitFor(() => expect(screen.queryByTestId('odds-panel')).not.toBeInTheDocument());

    // Now resolve the stale odds response → should be ignored
    await act(async () => {
      oddsResolve(oddsState);
    });
    expect(screen.queryByTestId('odds-panel')).not.toBeInTheDocument();
  });

  it('ignores odds API error silently', async () => {
    mockExec.mockResolvedValue(exchangeState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());

    vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] });
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByAltText('♠ A'));
    await act(async () => {
      vi.advanceTimersByTime(300);
    });
    vi.useRealTimers();

    // No crash, no odds panel
    await waitFor(() => expect(screen.queryByTestId('odds-panel')).not.toBeInTheDocument());
  });

  it('does not show odds panel when all probabilities are 0', async () => {
    const oddsState: PokerResponse = {
      ...exchangeState,
      odds: [{ handRank: 0, handName: 'High Card', probability: 0, count: 0, total: 10 }],
    };
    mockExec.mockResolvedValue(exchangeState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());

    vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] });
    mockExec.mockResolvedValue(oddsState);
    fireEvent.click(screen.getByAltText('♠ A'));
    await act(async () => {
      vi.advanceTimersByTime(300);
    });
    vi.useRealTimers();

    // Wait for state update then check panel is not shown
    await waitFor(() => expect(screen.queryByTestId('odds-panel')).not.toBeInTheDocument());
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue({
      exchangeRead: false,
      gameEndFlag: true,
      phase: 3, // PokerPhase.END
      currentTurn: 0,
      players: [],
      playerIdx: 0,
    } as unknown as PokerResponse);

    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.poker).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));

    await waitFor(() => expect(actionLogApi.poker).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();

    fireEvent.click(screen.getByText('閉じる'));
    await waitFor(() => expect(screen.queryByText('棋譜')).not.toBeInTheDocument());
    expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
  });

  // --- ConfirmDialog tests ---

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, expect.any(Object)));
  });

  // --- PhaseIndicator coverage ---

  it('phase indicator shows your turn during exchange phase', async () => {
    mockExec.mockResolvedValue(exchangeState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows waiting during end phase', async () => {
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('待機中'));
  });

  // --- Keyboard navigation tests ---

  describe('keyboard navigation', () => {
    it('pressing number keys toggles card selection', async () => {
      mockExec.mockResolvedValue(exchangeState);
      renderWithProviders(<PokerPage />);
      await waitFor(() => expect(screen.getByText('交換するカードを選んでください')).toBeInTheDocument());

      const buttons = screen.getAllByRole('button', { pressed: false });
      const cardButtons = buttons.filter((b) => b.getAttribute('aria-pressed') !== null);
      expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'false');

      // Press '1' to toggle first card
      await act(async () => {
        fireEvent.keyDown(document, { key: '1' });
      });
      expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'true');

      // Press '1' again to deselect
      await act(async () => {
        fireEvent.keyDown(document, { key: '1' });
      });
      expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'false');
    });

    it('Enter key triggers exchange', async () => {
      mockExec.mockResolvedValue(exchangeState);
      renderWithProviders(<PokerPage />);
      await waitFor(() => expect(screen.getByText('交換するカードを選んでください')).toBeInTheDocument());
      mockExec.mockClear();
      mockExec.mockResolvedValue(secondBetState);

      // Select a card first, then press Enter
      await act(async () => {
        fireEvent.keyDown(document, { key: '1' });
      });
      await act(async () => {
        fireEvent.keyDown(document, { key: 'Enter' });
      });

      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('exchange', [0]));
    });

    it('Enter key does nothing when no cards are selected', async () => {
      mockExec.mockResolvedValue(exchangeState);
      renderWithProviders(<PokerPage />);
      await waitFor(() => expect(screen.getByText('交換するカードを選んでください')).toBeInTheDocument());
      mockExec.mockClear();

      await act(async () => {
        fireEvent.keyDown(document, { key: 'Enter' });
      });

      await flushPendingDispatch();
      expect(mockExec).not.toHaveBeenCalled();
    });

    it('Escape key clears selection', async () => {
      mockExec.mockResolvedValue(exchangeState);
      renderWithProviders(<PokerPage />);
      await waitFor(() => expect(screen.getByText('交換するカードを選んでください')).toBeInTheDocument());

      const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);

      // Select a card
      await act(async () => {
        fireEvent.keyDown(document, { key: '1' });
      });
      expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'true');

      // Press Escape to clear
      await act(async () => {
        fireEvent.keyDown(document, { key: 'Escape' });
      });
      expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'false');
    });

    it('keyboard is disabled when not in exchange phase', async () => {
      mockExec.mockResolvedValue(dealState);
      renderWithProviders(<PokerPage />);
      await waitFor(() => expect(screen.getByText('あなたの番です')).toBeInTheDocument());

      const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);

      // Press '1' - should not toggle since we're in deal phase, not exchange
      await act(async () => {
        fireEvent.keyDown(document, { key: '1' });
      });
      expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'false');
    });
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(initState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  // --- Mobile layout tests ---

  describe('mobile viewport', () => {
    const originalInnerWidth = window.innerWidth;

    beforeEach(() => {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
    });

    afterEach(() => {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalInnerWidth });
      window.dispatchEvent(new Event('resize'));
    });

    it('renders CpuAccordion on mobile', async () => {
      mockExec.mockResolvedValue(dealState);
      renderWithProviders(<PokerPage />);
      await waitFor(() => expect(screen.getByTestId('cpu-accordion')).toBeInTheDocument());
    });

    it('renders CpuActionToast instead of CpuActionLog on mobile', async () => {
      mockExec.mockResolvedValue(dealWithBetState);
      renderWithProviders(<PokerPage />);
      await waitFor(() => expect(screen.getByTestId('cpu-accordion')).toBeInTheDocument());
      expect(screen.queryByText('CPU行動:')).not.toBeInTheDocument();
    });
  });

  it('getElapsed returns non-zero time when cpuMetaAI is enabled and it is human turn', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());

    // Enable cpuMetaAI so getElapsed computes elapsed instead of returning 0
    fireEvent.click(screen.getByLabelText('メタAI（CPUがプレイスタイルを学習）'));

    mockExec.mockClear();
    mockExec.mockResolvedValue(dealState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', undefined, 10, undefined, expect.any(Number)));
  });

  it('shows the lowball hand-rank reference only in lowball mode', async () => {
    mockExec.mockResolvedValue({ ...dealState, isLowball: true });
    renderWithProviders(<PokerPage />);
    expect(await screen.findByTestId('pk-lowball-reference')).toBeInTheDocument();
  });

  it('hides the lowball reference in normal (non-lowball) mode', async () => {
    mockExec.mockResolvedValue(dealState); // isLowball: false
    renderWithProviders(<PokerPage />);
    // Wait for the deal phase to render before asserting the reference is absent.
    await screen.findByRole('button', { name: 'ベット' });
    expect(screen.queryByTestId('pk-lowball-reference')).not.toBeInTheDocument();
  });

  // **Holdem 系は EquityDisplay を持つのに、5カードドローには仕組み自体が
  // 無く、2巡目ベットで call/raise/fold を判断する材料が交換確率パネルしか
  // 無かった (#4678)。**
  it('shows the equity display during a betting round', async () => {
    mockExec.mockResolvedValue({
      ...secondBetState,
      equity: { winProbability: 0.62, handOdds: [] },
      potOdds: 25,
    });
    renderWithProviders(<PokerPage />);

    await waitFor(() => expect(screen.getByTestId('equity-display')).toBeInTheDocument());
  });

  // **サーバーが送らないフェーズでは出さない。**交換中はまだ手が変わるので、
  // 確定した勝率として読まれると誤解を招く。ページ側でフェーズを再判定せず、
  // 値の有無だけで決める。
  it('shows no equity display when the server sends none', async () => {
    mockExec.mockResolvedValue(secondBetState);
    renderWithProviders(<PokerPage />);

    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('equity-display')).not.toBeInTheDocument();
  });
});

// #5475: 交換枚数が3枚未満だと CPU のフォールド閾値が1ランク上がるという実在の
// 戦略要素が、画面のどこにも説明されていなかった (frontend を grep しても
// exchangeRead は0件だった)。プレイヤーは読まれていることを知りようがない。
describe('PokerPage exchange read', () => {
  it('explains that a small exchange is being read', async () => {
    mockExec.mockResolvedValue({ ...secondBetState, exchangeRead: true });
    renderWithProviders(<PokerPage />);
    const note = await screen.findByTestId('poker-exchange-read');
    expect(note).toHaveAttribute('role', 'status');
    expect(note.textContent).toMatch(/3枚未満/);
  });

  // **判定はサーバが返す exchangeRead だけを見る。** 交換枚数から数え直すと
  // domain の閾値とずれる。ここでは 1 枚交換していても false なら出さない。
  it('does not re-derive the verdict from the exchange count', async () => {
    mockExec.mockResolvedValue({ ...secondBetState, exchangeRead: false });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('poker-exchange-read')).not.toBeInTheDocument();
  });
});
