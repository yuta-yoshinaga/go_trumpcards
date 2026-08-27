import { act, fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, baccaratApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BaccaratResponse, Card, CardDesign } from '../types/card';
import { BaccaratPage } from './BaccaratPage';

vi.mock('../hooks/useGameHint');

vi.mock('../api/gameApi', () => ({
  baccaratApi: { exec: vi.fn() },
  actionLogApi: { baccarat: vi.fn() },
}));

const mockExec = vi.mocked(baccaratApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const betPhaseState: BaccaratResponse = {
  playerHand: [],
  bankerHand: [],
  playerHandValue: 0,
  bankerHandValue: 0,
  phase: 1,
  chips: 1000,
  betAmount: 0,
  betType: 0,
  result: 0,
  payout: 0,
  history: [],
  playerPairBet: 0,
  bankerPairBet: 0,
  sideBetResults: [],
  message: '',
};

const endPhasePlayerWins: BaccaratResponse = {
  playerHand: [card('SPADE', 9), card('HEART', 3)],
  bankerHand: [card('CLOVER', 5), card('DIAMOND', 2)],
  playerHandValue: 2,
  bankerHandValue: 7,
  phase: 2,
  chips: 1100,
  betAmount: 100,
  betType: 0,
  result: 1,
  payout: 200,
  history: [0],
  playerPairBet: 0,
  bankerPairBet: 0,
  sideBetResults: [],
  message: 'プレイヤーの勝ち！',
  messageCode: 'baccarat.result.playerWins',
};

const endPhaseBankerWins: BaccaratResponse = {
  ...endPhasePlayerWins,
  result: -1,
  payout: 0,
  history: [1],
  message: 'バンカーの勝ち！',
  messageCode: 'baccarat.result.bankerWins',
};

const endPhaseTie: BaccaratResponse = {
  ...endPhasePlayerWins,
  result: 0,
  payout: 900,
  betType: 2,
  history: [2],
  message: '引き分け！',
  messageCode: 'baccarat.result.tie',
};

const errorState: BaccaratResponse = {
  ...betPhaseState,
  message: 'Invalid bet amount.',
};

const endPhaseWithSideBets: BaccaratResponse = {
  ...endPhasePlayerWins,
  playerPairBet: 10,
  bankerPairBet: 20,
  sideBetResults: [
    { betType: 1, resultType: 1, resultName: 'Pair', betAmount: 10, payout: 120 },
    { betType: 2, resultType: 0, resultName: '', betAmount: 20, payout: 0 },
  ],
};

const endPhaseWithHistory: BaccaratResponse = {
  ...endPhasePlayerWins,
  history: [0, 1, 0, 0, 1, 2, 1],
};

const endPhaseWithDragonTail: BaccaratResponse = {
  ...endPhasePlayerWins,
  history: [0, 0, 0, 0, 0, 0, 0, 1], // 7 player wins (dragon tail) + 1 banker
};

const endPhaseWithLeadingTie: BaccaratResponse = {
  ...endPhasePlayerWins,
  history: [2, 0, 1], // tie first, then player, then banker
};

const endPhaseWithOnlyTies: BaccaratResponse = {
  ...endPhasePlayerWins,
  history: [2, 2, 2], // only ties
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
});

describe('BaccaratPage', () => {
  it('renders bet phase on mount', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
  });

  it('renders the side-bet inputs inside a collapsed details section', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    const details = screen.getByTestId('baccarat-sidebet-details');
    // Collapsed by default (no `open` attribute) but the inputs still live inside it.
    expect(details).not.toHaveAttribute('open');
    expect(within(details).getByText('サイドベット（任意）')).toBeInTheDocument();
    expect(within(details).getByLabelText('プレイヤーペア')).toBeInTheDocument();
    expect(within(details).getByLabelText('バンカーペア')).toBeInTheDocument();
  });

  it('auto-expands the side-bet details once a side bet is non-zero', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    const details = screen.getByTestId('baccarat-sidebet-details');
    expect(details).not.toHaveAttribute('open');
    // Setting a side bet must keep it visible (not hidden behind the collapsed summary).
    fireEvent.change(screen.getByLabelText('プレイヤーペア'), { target: { value: '10' } });
    expect(details).toHaveAttribute('open');
  });

  it('does not expand card area with flex-1 during bet phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('ベットしてゲーム開始')).toBeInTheDocument());
    expect(screen.getByTestId('card-area')).not.toHaveClass('flex-1');
  });

  it('expands card area with flex-1 during end phase', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('プレイヤーの勝ち！')).toBeInTheDocument());
    expect(screen.getByTestId('card-area')).toHaveClass('flex-1');
  });

  it('renders skeleton before state loads', () => {
    mockExec.mockReturnValue(new Promise(() => {})); // never resolves
    renderWithProviders(<BaccaratPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('shows end phase with player wins', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(endPhasePlayerWins);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByText('プレイヤーの勝ち！')).toBeInTheDocument());
    // Payout is gated behind the staged-reveal final step (#1892).
    await waitFor(() => expect(screen.getByText('配当: 200')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument();
  });

  it('shows end phase with banker wins', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(endPhaseBankerWins);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByText('バンカーの勝ち！')).toBeInTheDocument());
  });

  it('shows end phase with tie', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(endPhaseTie);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByText('引き分け！')).toBeInTheDocument());
    // Payout is gated behind the staged-reveal final step (#1892).
    await waitFor(() => expect(screen.getByText('配当: 900')).toBeInTheDocument());
  });

  it('can change bet amount and bet type', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    const amountInput = screen.getByLabelText('ベット額:');
    fireEvent.change(amountInput, { target: { value: '200' } });

    const select = screen.getByRole('combobox');
    fireEvent.change(select, { target: { value: '1' } }); // banker

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 200, 1, 0, 0));
  });

  it('can set side bet amounts', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    const ppInput = screen.getByLabelText('プレイヤーペア');
    fireEvent.change(ppInput, { target: { value: '10' } });

    const bpInput = screen.getByLabelText('バンカーペア');
    fireEvent.change(bpInput, { target: { value: '20' } });

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 100, 0, 10, 20));
  });

  it('resets after end phase', async () => {
    mockExec
      .mockResolvedValueOnce(betPhaseState)
      .mockResolvedValueOnce(endPhasePlayerWins)
      .mockResolvedValueOnce(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // ── Next game button at end phase ──────────────────────────────────────────

  it('next game button does not show confirm dialog', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(endPhasePlayerWins);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('next game button executes reset directly', async () => {
    mockExec
      .mockResolvedValueOnce(betPhaseState)
      .mockResolvedValueOnce(endPhasePlayerWins)
      .mockResolvedValueOnce(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows error message', async () => {
    mockExec.mockResolvedValue(errorState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('Invalid bet amount.')).toBeInTheDocument());
  });

  it('shows network error', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockRejectedValueOnce(new Error('Network'));
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('shows action log', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(endPhasePlayerWins);
    vi.mocked(actionLogApi.baccarat).mockResolvedValue({ entries: [] as never[] });
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    fireEvent.click(screen.getByText('棋譜を見る'));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());
  });

  it('renders player and banker cards', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getAllByRole('img').length).toBe(4));
    expect(screen.getByText(/値: 2/)).toBeInTheDocument();
    expect(screen.getByText(/値: 7/)).toBeInTheDocument();
    expect(screen.getByText('🟡')).toBeInTheDocument();
    expect(screen.getByText('🔴')).toBeInTheDocument();
  });

  // --- Big Road tests ---

  it('renders Big Road grid when history is present', async () => {
    mockExec.mockResolvedValue(endPhaseWithHistory);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByTestId('big-road')).toBeInTheDocument());
    expect(screen.getByText('罫線')).toBeInTheDocument();
    expect(screen.getByText('履歴クリア')).toBeInTheDocument();
  });

  it('does not render Big Road when history is empty', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.queryByTestId('big-road')).not.toBeInTheDocument();
    expect(screen.queryByText('罫線')).not.toBeInTheDocument();
  });

  it('renders Big Road with dragon tail (7+ same side)', async () => {
    mockExec.mockResolvedValue(endPhaseWithDragonTail);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByTestId('big-road')).toBeInTheDocument());
    // Dragon tail: 7 player wins should create overflow columns, then banker in separate column
    expect(screen.getByText('罫線')).toBeInTheDocument();
  });

  it('renders Big Road with leading tie (tie dropped silently)', async () => {
    mockExec.mockResolvedValue(endPhaseWithLeadingTie);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByTestId('big-road')).toBeInTheDocument());
  });

  it('wraps the Big Road in a bounded, horizontally scrollable container', async () => {
    mockExec.mockResolvedValue(endPhaseWithHistory);
    renderWithProviders(<BaccaratPage />);
    const scroller = await screen.findByTestId('big-road-scroll');
    expect(scroller).toHaveClass('overflow-x-auto');
  });

  it('auto-scrolls the Big Road to the latest (right-most) results', async () => {
    // jsdom does no layout, so fake a wide scroll width and assert the effect
    // pins scrollLeft to the right edge.
    Object.defineProperty(HTMLElement.prototype, 'scrollWidth', {
      configurable: true,
      get() {
        return 500;
      },
    });
    mockExec.mockResolvedValue(endPhaseWithHistory);
    renderWithProviders(<BaccaratPage />);
    const scroller = await screen.findByTestId('big-road-scroll');
    await waitFor(() => expect(scroller.scrollLeft).toBe(500));
    delete (HTMLElement.prototype as unknown as { scrollWidth?: number }).scrollWidth;
  });

  it('does not render Big Road when history is only ties', async () => {
    mockExec.mockResolvedValue(endPhaseWithOnlyTies);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('プレイヤーの勝ち！')).toBeInTheDocument());
    // Only ties means columns is empty, so BigRoadGrid returns null
    // But history.length > 0, so the title and clear button should show
    expect(screen.queryByTestId('big-road')).not.toBeInTheDocument();
  });

  it('calls clearhistory when clear button is clicked', async () => {
    mockExec.mockResolvedValueOnce(endPhaseWithHistory).mockResolvedValueOnce(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('履歴クリア')).toBeInTheDocument());

    fireEvent.click(screen.getByText('履歴クリア'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('clearhistory'));
  });

  // --- Shoe stats panel tests ---

  it('renders shoe stats counts, rates, and streak from history', async () => {
    // history [0,1,0,0,1,2,1] -> P3 B3 T1 of 7; streak trailing banker (tie ignored) = 2
    mockExec.mockResolvedValue(endPhaseWithHistory);
    renderWithProviders(<BaccaratPage />);
    const panel = await screen.findByTestId('baccarat-shoe-stats');
    expect(within(panel).getByText('P 3回 (43%)')).toBeInTheDocument();
    expect(within(panel).getByText('B 3回 (43%)')).toBeInTheDocument();
    expect(within(panel).getByText('T 1回 (14%)')).toBeInTheDocument();
    expect(within(panel).getByText('現在 バンカー 2連勝')).toBeInTheDocument();
  });

  it('does not render shoe stats when history is empty', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.queryByTestId('baccarat-shoe-stats')).not.toBeInTheDocument();
  });

  it('omits the streak label for a tie-only history', async () => {
    mockExec.mockResolvedValue(endPhaseWithOnlyTies);
    renderWithProviders(<BaccaratPage />);
    const panel = await screen.findByTestId('baccarat-shoe-stats');
    expect(within(panel).getByText('T 3回 (100%)')).toBeInTheDocument();
    expect(within(panel).queryByText(/連勝/)).not.toBeInTheDocument();
  });

  // --- Side bet tests ---

  it('shows side bet results in end phase', async () => {
    mockExec.mockResolvedValue(endPhaseWithSideBets);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByTestId('side-bet-results')).toBeInTheDocument());
    expect(screen.getByText(/プレイヤーペア.*ペア.*配当: 120/)).toBeInTheDocument();
    expect(screen.getByText(/バンカーペア.*ペアなし.*配当: 0/)).toBeInTheDocument();
  });

  it('does not show side bet results when empty', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('プレイヤーの勝ち！')).toBeInTheDocument());
    expect(screen.queryByTestId('side-bet-results')).not.toBeInTheDocument();
  });

  // --- Keyboard navigation tests ---

  it('pressing b triggers bet in BET phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(endPhasePlayerWins);
    fireEvent.keyDown(document, { key: 'b' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 100, 0, 0, 0));
  });

  it('pressing r triggers reset in END phase', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText(/プレイヤーの勝ち/)).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(betPhaseState);
    fireEvent.keyDown(document, { key: 'r' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('pressing b does not trigger bet in END phase', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText(/プレイヤーの勝ち/)).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'b' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('pressing r does not trigger reset in BET phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'r' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  // --- Hint UI tests ---

  it('shows hint tooltip when hintEnabled is true and hint is available', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'banker', reason: 'hintReason.bankerBestOdds', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('hides hint tooltip when hintEnabled is false', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'banker', reason: 'hintReason.bankerBestOdds', confidence: 'strong' },
      hintEnabled: false,
      setHintEnabled: vi.fn(),
    });
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();
  });

  it('hint settings checkbox calls setHintEnabled on toggle', async () => {
    const mockSetHintEnabled = vi.fn();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: mockSetHintEnabled });
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    const checkbox = screen.getByRole('checkbox', { name: 'ヒント表示' });
    fireEvent.click(checkbox);
    expect(mockSetHintEnabled).toHaveBeenCalledWith(true);
  });

  it('shows a Rebet button at end-phase after a bet has been placed, replaying with the same amount', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(endPhasePlayerWins);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByTestId('bac-rebet-button')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValueOnce(betPhaseState);
    mockExec.mockResolvedValueOnce(endPhasePlayerWins);
    fireEvent.click(screen.getByTestId('bac-rebet-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 100, 0, 0, 0));
  });

  it('does not show the Rebet button at end-phase when chips are insufficient to replay', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...endPhasePlayerWins, chips: 50 });
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    expect(screen.queryByTestId('bac-rebet-button')).not.toBeInTheDocument();
  });

  it("snapshots the bet when the 'b' keyboard shortcut is used so Rebet is available at end phase", async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(endPhasePlayerWins);
    fireEvent.keyDown(document, { key: 'b' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 100, 0, 0, 0));
    await waitFor(() => expect(screen.getByTestId('bac-rebet-button')).toBeInTheDocument());
  });

  it("the 'e' keyboard shortcut fires Rebet at end phase", async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());

    mockExec.mockResolvedValue(endPhasePlayerWins);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByTestId('bac-rebet-button')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValueOnce(betPhaseState);
    mockExec.mockResolvedValueOnce(endPhasePlayerWins);
    fireEvent.keyDown(document, { key: 'e' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 100, 0, 0, 0));
  });

  it('staggers the third-card reveal: 3rd card hidden until its timer fires, payout gated until last step', async () => {
    vi.useFakeTimers();
    try {
      const endWithThirdCards: BaccaratResponse = {
        ...endPhasePlayerWins,
        playerHand: [card('SPADE', 9), card('HEART', 3), card('CLOVER', 2)],
        bankerHand: [card('DIAMOND', 4), card('SPADE', 2), card('HEART', 6)],
        playerHandValue: 4,
        bankerHandValue: 2,
      };
      mockExec.mockResolvedValue(endWithThirdCards);
      renderWithProviders(<BaccaratPage />);
      // Step 1: initial 2 + 2 cards.
      await vi.waitFor(() => {
        expect(screen.getByTestId('bac-player-cards').childElementCount).toBe(2);
      });
      expect(screen.getByTestId('bac-banker-cards').childElementCount).toBe(2);
      expect(screen.queryByTestId('bac-payout')).not.toBeInTheDocument();
      // Step 2: player's third card lands.
      await act(async () => {
        await vi.advanceTimersByTimeAsync(600);
      });
      expect(screen.getByTestId('bac-player-cards').childElementCount).toBe(3);
      expect(screen.getByTestId('bac-banker-cards').childElementCount).toBe(2);
      // Step 3: banker's third card lands.
      await act(async () => {
        await vi.advanceTimersByTimeAsync(600);
      });
      expect(screen.getByTestId('bac-banker-cards').childElementCount).toBe(3);
      expect(screen.queryByTestId('bac-payout')).not.toBeInTheDocument();
      // Step 4: payout appears.
      await act(async () => {
        await vi.advanceTimersByTimeAsync(600);
      });
      expect(screen.getByTestId('bac-payout')).toBeInTheDocument();
    } finally {
      vi.useRealTimers();
    }
  });

  it('header hand value tracks the visible slice, not the final server value (no spoilers)', async () => {
    vi.useFakeTimers();
    try {
      // Player: [9,3,2] → final 4. Two-card total: 9+3 = 12 % 10 = 2.
      // Banker: [4,2,6] → final 2. Two-card total: 4+2 = 6.
      const endWithThirdCards: BaccaratResponse = {
        ...endPhasePlayerWins,
        playerHand: [card('SPADE', 9), card('HEART', 3), card('CLOVER', 2)],
        bankerHand: [card('DIAMOND', 4), card('SPADE', 2), card('HEART', 6)],
        playerHandValue: 4,
        bankerHandValue: 2,
      };
      mockExec.mockResolvedValue(endWithThirdCards);
      renderWithProviders(<BaccaratPage />);
      await vi.waitFor(() => {
        expect(screen.getByTestId('bac-player-cards').childElementCount).toBe(2);
      });
      // Header values are split across sibling text nodes ("プレイヤー" + space + "値: N"),
      // so query the row by data-testid and assert on its full textContent instead.
      const playerHeader = screen.getByTestId('bac-player-cards').previousElementSibling as HTMLElement;
      const bankerHeader = screen.getByTestId('bac-banker-cards').previousElementSibling as HTMLElement;
      // Step 1: two-card totals (not the final server values 4 / 2).
      expect(playerHeader.textContent).toContain('値: 2');
      expect(bankerHeader.textContent).toContain('値: 6');
      // Step 2: player's third card lands → 9+3+2 = 14 % 10 = 4. Banker still at 6.
      await act(async () => {
        await vi.advanceTimersByTimeAsync(600);
      });
      expect(playerHeader.textContent).toContain('値: 4');
      expect(bankerHeader.textContent).toContain('値: 6');
      // Step 3: banker's third card lands → 4+2+6 = 12 % 10 = 2. Player stays at 4.
      await act(async () => {
        await vi.advanceTimersByTimeAsync(600);
      });
      expect(playerHeader.textContent).toContain('値: 4');
      expect(bankerHeader.textContent).toContain('値: 2');
    } finally {
      vi.useRealTimers();
    }
  });

  it('no third-card stagger when both hands stand on 2 cards (skips straight to payout)', async () => {
    vi.useFakeTimers();
    try {
      mockExec.mockResolvedValue(endPhasePlayerWins);
      renderWithProviders(<BaccaratPage />);
      await vi.waitFor(() => {
        expect(screen.getByTestId('bac-player-cards').childElementCount).toBe(2);
      });
      // No third-card delays; only the final payout reveal at +600ms.
      await act(async () => {
        await vi.advanceTimersByTimeAsync(600);
      });
      expect(screen.getByTestId('bac-payout')).toBeInTheDocument();
    } finally {
      vi.useRealTimers();
    }
  });

  it('payout table is rendered as a collapsible details element in bet phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    const { container } = renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByText('配当表')).toBeInTheDocument());
    const details = container.querySelector('details');
    expect(details).toBeInTheDocument();
    const summary = details?.querySelector('summary');
    expect(summary).toHaveTextContent('配当表');
  });

  it('states the big road as text, not colour alone', async () => {
    mockExec.mockResolvedValue(endPhaseWithHistory);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(screen.getByTestId('big-road')).toBeInTheDocument());
    // Cells are coloured circles; the CUI already prints the same run as P/B/T.
    const summary = screen.getByTestId('big-road-summary');
    expect(summary).toHaveClass('sr-only');
    // history [0,1,0,0,1,2,1] = P B P P B T B, and the tie must not be dropped.
    expect(summary.textContent).toContain('プレイヤー');
    expect(summary.textContent).toContain('バンカー');
    expect(summary.textContent).toContain('タイ');
  });

  it('renders no big road summary before any hand is decided', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BaccaratPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('big-road-summary')).not.toBeInTheDocument();
  });
});
