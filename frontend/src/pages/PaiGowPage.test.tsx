import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, paigowApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, PaiGowResponse } from '../types/card';
import { cardAlt } from '../utils/cardAlt';
import { PaiGowPage } from './PaiGowPage';

vi.mock('../hooks/useGameHint');

vi.mock('../api/gameApi', () => ({
  paigowApi: { exec: vi.fn() },
  actionLogApi: { paigow: vi.fn() },
}));

const mockExec = vi.mocked(paigowApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

/** The 7 SET_HANDS fixture card buttons, in deal order, located by their cardAlt names. */
const getCardButtons = () =>
  setHandsPhaseState.playerCards.map((c) => screen.getByRole('button', { name: cardAlt(c) }));

const betPhaseState: PaiGowResponse = {
  playerCards: [],
  dealerCards: [],
  playerHighHand: [],
  playerLowHand: [],
  dealerHighHand: [],
  dealerLowHand: [],
  phase: 1,
  chips: 1000,
  bet: 0,
  result: 0,
  highHandResult: 0,
  lowHandResult: 0,
  payout: 0,
  commission: 0,
  playerHighRank: 0,
  playerLowRank: 0,
  dealerHighRank: 0,
  dealerLowRank: 0,
  message: '',
};

const setHandsPhaseState: PaiGowResponse = {
  ...betPhaseState,
  phase: 2,
  playerCards: [
    card('SPADE', 10),
    card('HEART', 11),
    card('DIAMOND', 13),
    card('CLOVER', 5),
    card('HEART', 7),
    card('SPADE', 3),
    card('DIAMOND', 9),
  ],
  bet: 100,
  chips: 900,
};

const endPhasePlayerWins: PaiGowResponse = {
  playerCards: [],
  dealerCards: [],
  playerHighHand: [card('SPADE', 10), card('HEART', 11), card('DIAMOND', 13), card('CLOVER', 5), card('HEART', 7)],
  playerLowHand: [card('SPADE', 3), card('DIAMOND', 9)],
  dealerHighHand: [card('CLOVER', 2), card('DIAMOND', 4), card('HEART', 6), card('SPADE', 8), card('CLOVER', 10)],
  dealerLowHand: [card('DIAMOND', 3), card('HEART', 5)],
  phase: 3,
  chips: 1190,
  bet: 100,
  result: 1,
  highHandResult: 1,
  lowHandResult: 1,
  payout: 200,
  commission: 10,
  playerHighRank: 0,
  playerLowRank: 0,
  dealerHighRank: 0,
  dealerLowRank: 0,
  message: '勝利！',
  messageCode: 'paigow.result.playerWins',
};

const endPhaseDealerWins: PaiGowResponse = {
  ...endPhasePlayerWins,
  result: -1,
  payout: 0,
  commission: 0,
  message: 'ディーラー勝利！',
  messageCode: 'paigow.result.dealerWins',
};

const endPhasePush: PaiGowResponse = {
  ...endPhasePlayerWins,
  result: 0,
  payout: 0,
  commission: 0,
  message: '引き分け！',
  messageCode: 'paigow.result.push',
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
});

describe('PaiGowPage', () => {
  it('renders bet phase on mount', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
  });

  it('renders skeleton before state loads', () => {
    mockExec.mockReturnValue(new Promise(() => {})); // never resolves
    renderWithProviders(<PaiGowPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('does not expand card area with flex-1 during bet phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByTestId('card-area')).not.toHaveClass('flex-1');
  });

  it('expands card area with flex-1 during set hands phase', async () => {
    mockExec.mockResolvedValue(setHandsPhaseState);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'セット' })).toBeInTheDocument());
    expect(screen.getByTestId('card-area')).toHaveClass('flex-1');
  });

  it('shows set hands phase with 7 cards and set button', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(setHandsPhaseState);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: 'セット' })).toBeInTheDocument());
    // 7 card buttons should be present
    expect(getCardButtons()).toHaveLength(7);
  });

  it('shows an accessible foul-rule help in the set-hands phase', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(setHandsPhaseState);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: 'セット' })).toBeInTheDocument());
    const help = screen.getByTestId('foul-rule-help');
    expect(help).toHaveTextContent('ハイ/ローの分け方は？');
    // A native <details>/<summary> is keyboard-focusable.
    expect(help.querySelector('summary')).toBeInTheDocument();
  });

  it('labels each card button with its cardAlt name, not a raw index', async () => {
    mockExec.mockResolvedValue(setHandsPhaseState);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'セット' })).toBeInTheDocument());
    for (const c of setHandsPhaseState.playerCards) {
      expect(screen.getByRole('button', { name: cardAlt(c) })).toHaveAttribute('aria-pressed', 'false');
    }
    expect(screen.queryByRole('button', { name: /Card \d/ })).not.toBeInTheDocument();
  });

  it('set button is disabled until 2 cards are selected', async () => {
    mockExec.mockResolvedValue(setHandsPhaseState);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'セット' })).toBeInTheDocument());

    expect(screen.getByRole('button', { name: 'セット' })).toBeDisabled();

    // Select 2 cards
    const cardButtons = getCardButtons();
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[1]);

    expect(screen.getByRole('button', { name: 'セット' })).not.toBeDisabled();
  });

  it('calls API with set command and selected indices', async () => {
    mockExec.mockResolvedValueOnce(setHandsPhaseState).mockResolvedValueOnce(endPhasePlayerWins);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'セット' })).toBeInTheDocument());

    // Pick a non-foul split: low = [C5, S3] keeps the 13 in the high hand.
    const cardButtons = getCardButtons();
    fireEvent.click(cardButtons[3]);
    fireEvent.click(cardButtons[5]);

    fireEvent.click(screen.getByRole('button', { name: 'セット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('set', undefined, 3, 5));
  });

  it('auto-set button fills a legal non-foul split and enables Set', async () => {
    mockExec.mockResolvedValueOnce(setHandsPhaseState).mockResolvedValueOnce(endPhasePlayerWins);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByTestId('auto-set-button')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('auto-set-button'));

    // House-way max-low split for the singleton fixture is {S10, H11} = indices [0, 1].
    const cardButtons = getCardButtons();
    expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'true');
    expect(cardButtons[1]).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByTestId('foul-warning')).toHaveTextContent('');
    expect(screen.getByTestId('set-hands-button')).toBeEnabled();

    fireEvent.click(screen.getByTestId('set-hands-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('set', undefined, 0, 1));
  });

  it('lets the player re-select manually after an auto-set', async () => {
    mockExec.mockResolvedValue(setHandsPhaseState);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByTestId('auto-set-button')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('auto-set-button'));
    const cardButtons = getCardButtons();
    expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'true');

    // Deselect one auto-picked card — the auto choice is not locked in.
    fireEvent.click(cardButtons[0]);
    expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'false');
    expect(cardButtons[1]).toHaveAttribute('aria-pressed', 'true');
  });

  it('disables the auto-set button when a joker is present', async () => {
    mockExec.mockResolvedValue({
      ...setHandsPhaseState,
      playerCards: [
        card('JOKER', 0),
        card('HEART', 13),
        card('SPADE', 2),
        card('DIAMOND', 3),
        card('CLOVER', 4),
        card('SPADE', 5),
        card('HEART', 7),
      ],
    });
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByTestId('auto-set-button')).toBeInTheDocument());
    expect(screen.getByTestId('auto-set-button')).toBeDisabled();
  });

  it('blocks Set and shows a foul warning when the low hand outranks the high hand', async () => {
    mockExec.mockResolvedValue(setHandsPhaseState);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'セット' })).toBeInTheDocument());

    // Before selecting, the assertive live region is present but empty (so a
    // later foul — and its clearing — are both announced).
    const foulRegion = screen.getByTestId('foul-warning');
    expect(foulRegion).toHaveAttribute('aria-live', 'assertive');
    expect(foulRegion).toHaveTextContent('');

    // Low = [D13, S3] (top 13); high = [S10, H11, C5, H7, D9] (top 11) → foul.
    const cardButtons = getCardButtons();
    fireEvent.click(cardButtons[2]);
    fireEvent.click(cardButtons[5]);

    expect(foulRegion).not.toHaveTextContent('');
    expect(screen.getByTestId('set-hands-button')).toBeDisabled();

    // Deselecting a card clears the foul → the region empties (announcing the change).
    fireEvent.click(cardButtons[5]);
    await waitFor(() => expect(screen.getByTestId('foul-warning')).toHaveTextContent(''));
  });

  it('deselects a card when clicked again', async () => {
    mockExec.mockResolvedValue(setHandsPhaseState);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'セット' })).toBeInTheDocument());

    const cardButtons = getCardButtons();
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[1]);
    // Deselect first card
    fireEvent.click(cardButtons[0]);

    // Should be disabled again (only 1 selected)
    expect(screen.getByRole('button', { name: 'セット' })).toBeDisabled();
  });

  it('does not allow selecting more than 2 cards', async () => {
    mockExec.mockResolvedValue(setHandsPhaseState);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'セット' })).toBeInTheDocument());

    const cardButtons = getCardButtons();
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[1]);
    fireEvent.click(cardButtons[2]); // should be ignored

    // Third card should not be pressed
    expect(cardButtons[2]).toHaveAttribute('aria-pressed', 'false');
  });

  it('shows end phase with player wins', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByText('勝利！')).toBeInTheDocument());
    expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument();
  });

  it('shows end phase with dealer wins', async () => {
    mockExec.mockResolvedValue(endPhaseDealerWins);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByText('ディーラー勝利！')).toBeInTheDocument());
  });

  it('shows end phase with push', async () => {
    mockExec.mockResolvedValue(endPhasePush);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByText('引き分け！')).toBeInTheDocument());
  });

  it('shows payout and commission in end phase', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());
    expect(screen.getByText(/配当: 200/)).toBeInTheDocument();
    expect(screen.getByText(/コミッション: 10/)).toBeInTheDocument();
  });

  it('can change bet amount', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    const betInput = screen.getByLabelText('ベット');
    fireEvent.change(betInput, { target: { value: '200' } });

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 200));
  });

  it('steppers adjust the bet by the step and submit the new amount', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    // Default bet is 100; the +10 stepper raises it to 110.
    fireEvent.click(screen.getByRole('button', { name: 'ベット +10' }));
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 110));
  });

  it('prevents an out-of-range bet: disables submit and shows an alert', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    // A non-multiple-of-10 amount is invalid.
    const betInput = screen.getByLabelText('ベット');
    fireEvent.change(betInput, { target: { value: '15' } });

    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ベット' })).toBeDisabled();

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('bet', 15);
  });

  it('resets after end phase', async () => {
    mockExec
      .mockResolvedValueOnce(betPhaseState)
      .mockResolvedValueOnce(endPhasePlayerWins)
      .mockResolvedValueOnce(betPhaseState);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('next game button in end phase triggers reset immediately', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(betPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('next game button shows no confirm dialog at end phase', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('shows network error', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockRejectedValueOnce(new Error('Network'));
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('shows action log', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    vi.mocked(actionLogApi.paigow).mockResolvedValue({ entries: [] as never[] });
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    fireEvent.click(screen.getByText('棋譜を見る'));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());
  });

  it('renders player and dealer hands in end phase', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<PaiGowPage />);
    // Player high (5) + player low (2) + dealer high (5) + dealer low (2) = 14
    await waitFor(() => expect(screen.getAllByRole('img').length).toBe(14));
  });

  // --- Keyboard navigation tests ---

  it('pressing b triggers bet in BET phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(setHandsPhaseState);
    fireEvent.keyDown(document, { key: 'b' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 100));
  });

  it('pressing r triggers reset in END phase', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByText(/勝利/)).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(betPhaseState);
    fireEvent.keyDown(document, { key: 'r' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('pressing b does not trigger bet in END phase', async () => {
    mockExec.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByText(/勝利/)).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'b' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('pressing r does not trigger reset in BET phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'r' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('shows hint tooltip when hintEnabled is true and hint is available', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'set', reason: 'hintReason.strongBothHands', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    mockExec.mockResolvedValue(setHandsPhaseState);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('hides hint tooltip when hintEnabled is false', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'set', reason: 'hintReason.strongBothHands', confidence: 'strong' },
      hintEnabled: false,
      setHintEnabled: vi.fn(),
    });
    mockExec.mockResolvedValue(setHandsPhaseState);
    renderWithProviders(<PaiGowPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'セット' })).toBeInTheDocument());
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();
  });
});
