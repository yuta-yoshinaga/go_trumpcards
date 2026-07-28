import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ultimatetexasholdemApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, UltimateTexasHoldemResponse } from '../types/card';
import { UltimateTexasHoldemPage } from './UltimateTexasHoldemPage';

vi.mock('../api/gameApi', () => ({
  ultimatetexasholdemApi: { exec: vi.fn() },
  actionLogApi: { ultimatetexasholdem: vi.fn() },
}));

const mockApi = vi.mocked(ultimatetexasholdemApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });
const maskedCard = { design: '' as CardDesign, value: 0 };

const betPhaseState: UltimateTexasHoldemResponse = {
  playerHand: [],
  dealerHand: [],
  community: [],
  phase: 1,
  chips: 1000,
  anteBet: 0,
  blindBet: 0,
  tripsBet: 0,
  playBet: 0,
  folded: false,
  result: 0,
  dealerQualified: false,
  antePayout: 0,
  blindPayout: 0,
  playPayout: 0,
  tripsPayout: 0,
  totalPayout: 0,
  playerHandRank: 0,
  dealerHandRank: 0,
  message: '',
};

const preFlopState: UltimateTexasHoldemResponse = {
  ...betPhaseState,
  phase: 2,
  playerHand: [card('SPADE', 1), card('SPADE', 13)],
  dealerHand: [maskedCard, maskedCard],
  anteBet: 100,
  blindBet: 100,
  chips: 800,
};

const flopState: UltimateTexasHoldemResponse = {
  ...preFlopState,
  phase: 3,
  community: [card('SPADE', 12), card('SPADE', 11), card('SPADE', 10)],
};

const riverState: UltimateTexasHoldemResponse = {
  ...flopState,
  phase: 4,
  community: [card('SPADE', 12), card('SPADE', 11), card('SPADE', 10), card('CLOVER', 2), card('HEART', 4)],
};

const endPlayerWins: UltimateTexasHoldemResponse = {
  ...riverState,
  phase: 5,
  dealerHand: [card('HEART', 7), card('DIAMOND', 5)],
  playBet: 100,
  result: 1,
  dealerQualified: true,
  antePayout: 200,
  blindPayout: 100 + 100 * 500, // royal flush blind
  playPayout: 200,
  tripsPayout: 0,
  totalPayout: 200 + (100 + 100 * 500) + 200,
  playerHandRank: 9,
  dealerHandRank: 0,
  message: '勝利！',
  messageCode: 'ultimatetexasholdem.result.playerWins',
  chips: 60000,
};

const endDealerWins: UltimateTexasHoldemResponse = {
  ...endPlayerWins,
  result: -1,
  dealerQualified: true,
  antePayout: 0,
  blindPayout: 0,
  playPayout: 0,
  totalPayout: 0,
  playerHandRank: 0,
  dealerHandRank: 1,
  message: 'ディーラー勝利！',
  messageCode: 'ultimatetexasholdem.result.dealerWins',
};

const endFold: UltimateTexasHoldemResponse = {
  ...riverState,
  phase: 5,
  folded: true,
  result: -1,
  message: 'フォールド',
  messageCode: 'ultimatetexasholdem.result.fold',
};

const endPush: UltimateTexasHoldemResponse = {
  ...endPlayerWins,
  result: 0,
  antePayout: 100,
  blindPayout: 100,
  playPayout: 100,
  totalPayout: 300,
  message: '引き分け！',
  messageCode: 'ultimatetexasholdem.result.push',
};

const endDealerNotQualifiedAntePushes: UltimateTexasHoldemResponse = {
  ...endPlayerWins,
  dealerQualified: false,
  // Ante pushes (returned only) while Blind and Play still pay out normally.
  antePayout: 100,
  blindPayout: 100 + 100 * 500,
  playPayout: 200,
};

beforeEach(() => {
  vi.clearAllMocks();
});

afterEach(() => {
  localStorage.clear();
});

describe('UltimateTexasHoldemPage', () => {
  it('renders bet phase on mount', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
  });

  it('renders skeleton before state loads', () => {
    mockApi.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<UltimateTexasHoldemPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('shows pre-flop with play 4x / 3x / check buttons', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(preFlopState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ 4×' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'プレイ 3×' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument();
  });

  it('shows flop with play 2x / check buttons after preflop check', async () => {
    mockApi.mockResolvedValueOnce(preFlopState).mockResolvedValueOnce(flopState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ 4×' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'チェック' }));
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ 2×' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument();
  });

  it('shows river with play 1x / fold buttons after flop check', async () => {
    mockApi.mockResolvedValueOnce(flopState).mockResolvedValueOnce(riverState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ 2×' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'チェック' }));
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ 1×' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
  });

  it('sends multiplier 4 when Play 4x is pressed', async () => {
    mockApi.mockResolvedValue(preFlopState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ 4×' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'プレイ 4×' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('play', undefined, undefined, 4));
  });

  it('sends multiplier 3 when Play 3x is pressed', async () => {
    mockApi.mockResolvedValue(preFlopState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ 3×' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'プレイ 3×' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('play', undefined, undefined, 3));
  });

  it('shows end phase with player wins', async () => {
    mockApi.mockResolvedValue(endPlayerWins);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('勝利！')).toBeInTheDocument());
    expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument();
  });

  it('shows dealer-qualified label at end phase', async () => {
    mockApi.mockResolvedValue(endPlayerWins);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText(/クオリファイ（ペア以上）/)).toBeInTheDocument());
  });

  it('shows dealer-not-qualified label at end phase', async () => {
    mockApi.mockResolvedValue(endDealerNotQualifiedAntePushes);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText(/クオリファイなし/)).toBeInTheDocument());
  });

  it('shows end phase with dealer wins', async () => {
    mockApi.mockResolvedValue(endDealerWins);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('ディーラー勝利！')).toBeInTheDocument());
  });

  it('shows end phase with fold', async () => {
    mockApi.mockResolvedValue(endFold);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('フォールド')).toBeInTheDocument());
  });

  it('shows end phase with push', async () => {
    mockApi.mockResolvedValue(endPush);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('引き分け！')).toBeInTheDocument());
  });

  it('changes ante and trips amounts when betting', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    const anteInput = screen.getByLabelText('アンテ');
    fireEvent.change(anteInput, { target: { value: '200' } });

    const tripsInput = screen.getByLabelText('トリップス');
    fireEvent.change(tripsInput, { target: { value: '20' } });

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 200, 20));
  });

  it('shows network error', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockRejectedValueOnce(new Error('Network'));
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('renders board and player cards on flop', async () => {
    mockApi.mockResolvedValue(flopState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    // 2 player face-up + 3 community face-up + 2 dealer face-down (back image)
    await waitFor(() => expect(screen.getAllByRole('img').length).toBeGreaterThanOrEqual(5));
    expect(screen.getByText('🟡')).toBeInTheDocument();
    expect(screen.getByText('🔴')).toBeInTheDocument();
    expect(screen.getByText('🃏')).toBeInTheDocument();
  });

  it('renders hint toggle checkbox', async () => {
    mockApi.mockResolvedValue(preFlopState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ 4×' })).toBeInTheDocument());
    expect(screen.getByRole('checkbox')).toBeInTheDocument();
  });

  it('renders hint tooltip when the checkbox is enabled', async () => {
    mockApi.mockResolvedValue(preFlopState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ 4×' })).toBeInTheDocument());

    const checkbox = screen.getByRole('checkbox') as HTMLInputElement;
    fireEvent.click(checkbox);
    expect(checkbox.checked).toBe(true);
    // Pre-flop with hole cards A♠/K♠ -> suited Broadway -> hint should render.
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('sends multiplier 2 when Play 2× is pressed on the flop', async () => {
    mockApi.mockResolvedValue(flopState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ 2×' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'プレイ 2×' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('play', undefined, undefined, 2));
  });

  it('sends multiplier 1 when Play 1× is pressed on the river', async () => {
    mockApi.mockResolvedValue(riverState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ 1×' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'プレイ 1×' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('play', undefined, undefined, 1));
  });

  it('sends fold when Fold is pressed on the river', async () => {
    mockApi.mockResolvedValue(riverState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('fold'));
  });

  it('renders the action-log button at end phase', async () => {
    mockApi.mockResolvedValue(endPlayerWins);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('勝利！')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument();
  });

  it('caps the ante input max at half the available chips', async () => {
    mockApi.mockResolvedValue({ ...betPhaseState, chips: 400 });
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 400')).toBeInTheDocument());

    const anteInput = screen.getByLabelText('アンテ') as HTMLInputElement;
    expect(anteInput.max).toBe('200');
  });

  it('clamps the ante input when chips drop below the current ante', async () => {
    // Initial render with 1000 chips, default ante 100. Then chips drop to 100.
    mockApi.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce({ ...betPhaseState, chips: 100 });
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    // Force a re-fetch by clicking bet (which the mock returns as a low-chip state).
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByText('チップ: 100')).toBeInTheDocument());

    const anteInput = screen.getByLabelText('アンテ') as HTMLInputElement;
    // Math.floor(100 / 2) === 50, so the input cannot stay at 100.
    expect(Number(anteInput.value)).toBeLessThanOrEqual(50);
  });

  it('pulses the 3x button and not the 4x when the pre-flop hand is moderate', async () => {
    // K-7 offsuit → moderate per utHoldemPreflopStrength.
    mockApi.mockResolvedValue({ ...preFlopState, playerHand: [card('SPADE', 13), card('HEART', 7)] });
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByTestId('play-3x')).toBeInTheDocument());

    const container = screen.getByTestId('play-4x').closest('[data-preflop-strength]') as HTMLElement;
    expect(container).toHaveAttribute('data-preflop-strength', 'moderate');
    expect(screen.getByTestId('play-3x').className).toContain('animate-pulse');
    expect(screen.getByTestId('play-4x').className).not.toContain('animate-pulse');
  });

  it('does not pulse any button when the pre-flop hand is weak', async () => {
    // 7-2 offsuit → weak per utHoldemPreflopStrength.
    mockApi.mockResolvedValue({ ...preFlopState, playerHand: [card('SPADE', 7), card('HEART', 2)] });
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByTestId('play-4x')).toBeInTheDocument());

    const container = screen.getByTestId('play-4x').closest('[data-preflop-strength]') as HTMLElement;
    expect(container).toHaveAttribute('data-preflop-strength', 'weak');
    expect(screen.getByTestId('play-4x').className).not.toContain('animate-pulse');
    expect(screen.getByTestId('play-3x').className).not.toContain('animate-pulse');
  });

  it('shows a strong pre-flop evaluation in success color (A + suited K → 4×)', async () => {
    mockApi.mockResolvedValue(preFlopState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    const evalText = await screen.findByTestId('uth-preflop-eval');
    expect(evalText).toHaveTextContent('強い手札 → 4× 推奨');
    expect(evalText.className).toContain('text-ds-success');
  });

  it('shows a moderate pre-flop evaluation in warning color (K-7 offsuit → 3×)', async () => {
    mockApi.mockResolvedValue({ ...preFlopState, playerHand: [card('SPADE', 13), card('HEART', 7)] });
    renderWithProviders(<UltimateTexasHoldemPage />);
    const evalText = await screen.findByTestId('uth-preflop-eval');
    expect(evalText).toHaveTextContent('微妙な手札 → 3× 推奨');
    expect(evalText.className).toContain('text-ds-warning');
  });

  it('shows a weak pre-flop evaluation in muted color (7-2 offsuit → check)', async () => {
    mockApi.mockResolvedValue({ ...preFlopState, playerHand: [card('SPADE', 7), card('HEART', 2)] });
    renderWithProviders(<UltimateTexasHoldemPage />);
    const evalText = await screen.findByTestId('uth-preflop-eval');
    expect(evalText).toHaveTextContent('弱い手札 → チェック推奨');
    expect(evalText.className).toContain('text-ds-text-muted');
  });

  it('displays the combined bet total (ante*2 + trips) before submitting', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    // Default ante 100 (×2) + trips 0 = 200 committed of 1000 chips.
    expect(screen.getByTestId('uth-bet-total')).toHaveTextContent('200 / 1000');

    fireEvent.change(screen.getByLabelText('トリップス'), { target: { value: '150' } });
    expect(screen.getByTestId('uth-bet-total')).toHaveTextContent('350 / 1000');
  });

  it('submits when ante*2 + trips is within chips', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText('アンテ'), { target: { value: '300' } });
    fireEvent.change(screen.getByLabelText('トリップス'), { target: { value: '400' } });
    // 300*2 + 400 = 1000 === chips → valid boundary.
    const betButton = screen.getByRole('button', { name: 'ベット' });
    expect(betButton).not.toBeDisabled();
    expect(screen.queryByTestId('uth-bet-error')).not.toBeInTheDocument();

    fireEvent.click(betButton);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 300, 400));
  });

  it('disables submit and shows an alert when ante*2 + trips exceeds chips', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    // Ante 100 (×2 = 200) + trips 900 = 1100 > 1000 chips.
    fireEvent.change(screen.getByLabelText('トリップス'), { target: { value: '900' } });

    const alert = await screen.findByTestId('uth-bet-error');
    expect(alert).toHaveAttribute('role', 'alert');
    expect(screen.getByRole('button', { name: 'ベット' })).toBeDisabled();
    expect(screen.getByLabelText('トリップス')).toHaveAttribute('aria-invalid', 'true');

    // A blocked over-balance bet must not reach the server.
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await flushPendingDispatch();
    expect(mockApi).not.toHaveBeenCalledWith('bet', expect.anything(), expect.anything());
  });

  it('re-clamps trips immediately when the ante increases past the combined limit', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    // Trips at 800 (max while ante=100: 1000 - 200).
    const tripsInput = screen.getByLabelText('トリップス') as HTMLInputElement;
    fireEvent.change(tripsInput, { target: { value: '800' } });
    expect(Number(tripsInput.value)).toBe(800);

    // Raise ante to 400 → new max trips = 1000 - 800 = 200; trips must re-cap.
    fireEvent.change(screen.getByLabelText('アンテ'), { target: { value: '400' } });
    await waitFor(() => expect(Number(tripsInput.value)).toBe(200));
    expect(screen.queryByTestId('uth-bet-error')).not.toBeInTheDocument();
  });
});

// --- keyboard shortcut execution (#4429) ---
// The digit keys are the play multipliers (4x/3x pre-flop, 2x on the flop, 1x on
// the river) and differ only by their fourth argument, so that argument is what
// each case pins.
const kbdCases: [string, unknown[], UltimateTexasHoldemResponse][] = [
  ['b', ['bet', 100, 0], betPhaseState],
  ['4', ['play', undefined, undefined, 4], preFlopState],
  ['3', ['play', undefined, undefined, 3], preFlopState],
  ['c', ['check'], preFlopState],
  ['2', ['play', undefined, undefined, 2], flopState],
  ['c', ['check'], flopState],
  ['1', ['play', undefined, undefined, 1], riverState],
  ['f', ['fold'], riverState],
  ['r', ['reset'], endPlayerWins],
];

describe('UltimateTexasHoldemPage keyboard shortcuts', () => {
  it.each(kbdCases)('pressing %s dispatches %j', async (key, expected, state) => {
    mockApi.mockResolvedValue(state);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockApi.mockClear();
    mockApi.mockResolvedValue(state);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith(...expected));
  });

  it('ignores a play multiplier that belongs to another street', async () => {
    // '2' is the flop bet; pre-flop offers only 4x and 3x.
    mockApi.mockResolvedValue(preFlopState);
    renderWithProviders(<UltimateTexasHoldemPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockApi.mockClear();
    fireEvent.keyDown(document, { key: '2' });
    await flushPendingDispatch();
    expect(mockApi).not.toHaveBeenCalled();
  });
});
