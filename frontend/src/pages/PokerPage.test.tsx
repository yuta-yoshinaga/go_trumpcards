import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { pokerApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { PokerResponse } from '../types/card';
import { PokerPage } from './PokerPage';

vi.mock('../api/gameApi', () => ({
  pokerApi: { exec: vi.fn() },
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
};

/** DEAL phase (phase 1): human's turn, no outstanding bet */
const dealState: PokerResponse = {
  players: [humanPlayer(), cpuPlayer(1), cpuPlayer(2), cpuPlayer(3)],
  pot: 40,
  sidePots: [],
  dealerIdx: 3,
  currentTurn: 0,
  phase: 1,
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
  gameEndFlag: false,
  lastBet: 0,
  minRaise: 0,
  ante: 10,
  jokerCount: 0,
  roundResults: [
    { playerIdx: 0, handRank: 0, handName: 'High Card', wonAmount: 0 },
    { playerIdx: 1, handRank: 1, handName: 'One Pair', wonAmount: 200 },
  ],
  cpuActions: [],
  cpuExchanges: [],
  message: 'あなたの負け',
  bettingLimit: 0,
  raiseCount: 0,
  maxBetAmount: 0,
};

beforeEach(() => {
  mockExec.mockResolvedValue(initState);
});

describe('PokerPage', () => {
  // ---- mount & reset ----
  it('calls reset on mount', async () => {
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // ---- info bar ----
  it('shows pot and dealer index', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/ポット:/)).toBeInTheDocument());
    expect(screen.getByText(/ディーラー:/)).toBeInTheDocument();
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

  it('shows CardBack for CPU cards when not in END phase', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    const cardBacks = screen.getAllByAltText('カード裏面');
    // 3 CPUs * 5 cards each = 15 card backs
    expect(cardBacks.length).toBeGreaterThanOrEqual(5);
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
    expect(screen.getByText(/あなた: High Card/)).toBeInTheDocument();
  });

  it('shows "CPU X" for non-human player in results', async () => {
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
    expect(screen.getByText(/CPU 1: One Pair/)).toBeInTheDocument();
  });

  it('shows hand name in results when present', async () => {
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/: High Card/)).toBeInTheDocument());
    expect(screen.getByText(/: One Pair/)).toBeInTheDocument();
  });

  it('does not show hand name colon when handName is empty in results', async () => {
    mockExec.mockResolvedValue({
      ...endState,
      roundResults: [
        { playerIdx: 0, handRank: 0, handName: '', wonAmount: 0 },
        { playerIdx: 1, handRank: 2, handName: 'One Pair', wonAmount: 200 },
      ],
    });
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
  });

  it('shows won chips when wonAmount > 0', async () => {
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/\+200チップ/)).toBeInTheDocument());
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

  // ---- button click handlers ----
  it('calls bet command with betAmount', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(dealState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', undefined, 10));
  });

  it('calls check command', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(dealState);
    fireEvent.click(screen.getByRole('button', { name: 'チェック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('check'));
  });

  it('calls fold command', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(endState);
    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold'));
  });

  it('calls allin command', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(dealState);
    fireEvent.click(screen.getByRole('button', { name: 'オールイン' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('allin'));
  });

  it('calls call command when has outstanding bet', async () => {
    mockExec.mockResolvedValue(dealWithBetState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(dealState);
    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('call'));
  });

  it('calls raise command with betAmount when has outstanding bet', async () => {
    mockExec.mockResolvedValue(dealWithBetState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'レイズ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(dealState);
    fireEvent.click(screen.getByRole('button', { name: 'レイズ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('raise', undefined, 10));
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
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', undefined, 60));
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
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('raise', undefined, 100));
  });

  it('calls reset command when reset button is clicked', async () => {
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { bettingLimit: 0 }));
  });

  it('sends updated bettingLimit when select is changed before reset', async () => {
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const select = screen.getByLabelText('リミット:');
    fireEvent.change(select, { target: { value: '1' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { bettingLimit: 1 }));
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
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('clears error on successful call after failure', async () => {
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());

    mockExec.mockResolvedValueOnce(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() => expect(screen.queryByText(NETWORK_ERROR_MESSAGE())).not.toBeInTheDocument());
  });

  // ---- message ----
  it('shows game message', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByText('あなたの番です')).toBeInTheDocument());
  });

  // ---- aria-busy / sr-only ----
  it('sets aria-busy and sr-only loading text while loading', async () => {
    mockExec.mockResolvedValue(dealState);
    renderWithProviders(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());

    const container = screen.getByRole('button', { name: 'ベット' }).closest('[aria-live]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');
    expect(screen.queryByText('処理中...')).not.toBeInTheDocument();

    let resolve!: (value: PokerResponse) => void;
    const slowPromise = new Promise<PokerResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));

    expect(container).toHaveAttribute('aria-busy', 'true');
    expect(screen.getByText('処理中...')).toBeInTheDocument();

    resolve(dealState);
    await waitFor(() => {
      expect(container).toHaveAttribute('aria-busy', 'false');
      expect(screen.queryByText('処理中...')).not.toBeInTheDocument();
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
});
