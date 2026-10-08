import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { casinoholdemApi } from '../api/gameApi';
import i18n from '../i18n';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, CasinoHoldemResponse } from '../types/card';
import { CasinoHoldemPage } from './CasinoHoldemPage';

vi.mock('../api/gameApi', () => ({
  casinoholdemApi: { exec: vi.fn() },
  actionLogApi: { casinoholdem: vi.fn() },
}));

const mockApi = vi.mocked(casinoholdemApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });
const maskedCard = { design: '' as CardDesign, value: 0 };

const betPhaseState: CasinoHoldemResponse = {
  playerHand: [],
  dealerHand: [],
  community: [],
  phase: 1,
  chips: 1000,
  anteBet: 0,
  bonusBet: 0,
  callBet: 0,
  result: 0,
  dealerQualify: false,
  antePayout: 0,
  callPayout: 0,
  bonusPayout: 0,
  totalPayout: 0,
  netChange: 0,
  playerHandRank: 0,
  dealerHandRank: 0,
  message: '',
};

const flopState: CasinoHoldemResponse = {
  ...betPhaseState,
  phase: 2,
  playerHand: [card('SPADE', 1), card('SPADE', 13)],
  dealerHand: [maskedCard, maskedCard],
  community: [card('SPADE', 12), card('SPADE', 11), card('SPADE', 10)],
  anteBet: 100,
  chips: 900,
  callBet: 200,
  callWinRate: 0.75,
};

const endPlayerWins: CasinoHoldemResponse = {
  ...flopState,
  phase: 3,
  dealerHand: [card('HEART', 7), card('DIAMOND', 5)],
  community: [card('SPADE', 12), card('SPADE', 11), card('SPADE', 10), card('CLOVER', 2), card('HEART', 4)],
  callBet: 200,
  result: 1,
  dealerQualify: true,
  antePayout: 100 + 100 * 100, // Royal flush ante
  callPayout: 400,
  totalPayout: 100 + 100 * 100 + 400,
  playerHandRank: 9,
  dealerHandRank: 0,
  message: '勝利！',
  messageCode: 'casinoholdem.result.playerWins',
  chips: 10800,
};

const endDealerWins: CasinoHoldemResponse = {
  ...endPlayerWins,
  result: -1,
  dealerQualify: true,
  antePayout: 0,
  callPayout: 0,
  totalPayout: 0,
  playerHandRank: 0,
  dealerHandRank: 1,
  message: 'ディーラー勝利！',
  messageCode: 'casinoholdem.result.dealerWins',
};

const endFold: CasinoHoldemResponse = {
  ...flopState,
  phase: 3,
  callBet: 0,
  result: -1,
  message: 'フォールド',
  messageCode: 'casinoholdem.result.fold',
};

const endNoQualify: CasinoHoldemResponse = {
  ...endPlayerWins,
  dealerQualify: false,
  callPayout: 200, // call pushes
};

beforeEach(() => {
  vi.clearAllMocks();
});

afterEach(() => {
  localStorage.clear();
});

describe('CasinoHoldemPage', () => {
  it('renders bet phase on mount', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
    expect(screen.queryByTestId('ch-call-win-rate')).not.toBeInTheDocument();
  });

  it('renders skeleton before state loads', () => {
    mockApi.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<CasinoHoldemPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('toggles the hint tooltip through the settings panel', async () => {
    mockApi.mockResolvedValue({ ...flopState, playerHandRank: 1 });
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /コール/ })).toBeInTheDocument());

    // Hint is off by default, so no tooltip renders.
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    // Open the shared settings panel and enable the hint via its checkbox.
    fireEvent.click(screen.getByText('設定'));
    const checkbox = screen.getByRole('checkbox', { name: 'ヒント表示' });
    fireEvent.click(checkbox);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());

    // Disabling it hides the hint again.
    fireEvent.click(checkbox);
    await waitFor(() => expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument());
  });

  it('shows flop with call and fold buttons', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(flopState);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: /コール/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    expect(screen.getByTestId('ch-flop-call-bet')).toHaveTextContent('コール額: 200');
    expect(screen.getByTestId('ch-dealer-qualify-rule')).toHaveTextContent(
      'ディーラーはツーペア以上、または4以上のペアでクオリファイします',
    );
    expect(screen.getByTestId('ch-call-win-rate')).toHaveTextContent('75%');
  });

  it('shows the dealer qualification rule in English during the flop', async () => {
    const previousLanguage = i18n.language;
    await i18n.changeLanguage('en');
    try {
      mockApi.mockResolvedValue(flopState);
      renderWithProviders(<CasinoHoldemPage />);
      expect(await screen.findByTestId('ch-dealer-qualify-rule')).toHaveTextContent(
        'The dealer qualifies with two pair or better, or a pair of fours or better',
      );
    } finally {
      await i18n.changeLanguage(previousLanguage);
    }
  });

  it('shows the server-provided call amount for a different ante', async () => {
    mockApi.mockResolvedValue({ ...flopState, anteBet: 150, callBet: 300 });
    renderWithProviders(<CasinoHoldemPage />);
    const callAmount = await screen.findByTestId('ch-flop-call-bet');
    expect(callAmount).toHaveTextContent('コール額: 300');
  });

  it('does not show the flop call amount outside the flop phase', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    expect(screen.queryByTestId('ch-flop-call-bet')).not.toBeInTheDocument();
  });

  it('does not show the flop call amount in the end phase', async () => {
    mockApi.mockResolvedValue(endPlayerWins);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());
    expect(screen.getByTestId('payout-breakdown').querySelector('[data-testid="ch-flop-call-bet"]')).toBeNull();
    expect(screen.queryByTestId('ch-call-win-rate')).not.toBeInTheDocument();
  });

  it('shows the call win estimate in English at the flop', async () => {
    const previousLanguage = i18n.language;
    await i18n.changeLanguage('en');
    try {
      mockApi.mockResolvedValue(flopState);
      renderWithProviders(<CasinoHoldemPage />);
      expect(await screen.findByTestId('ch-call-win-rate')).toHaveTextContent('Estimated chance to win after calling:');
    } finally {
      await i18n.changeLanguage(previousLanguage);
    }
  });

  it('shows end phase with player wins', async () => {
    mockApi.mockResolvedValue(endPlayerWins);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByText('勝利！')).toBeInTheDocument());
    expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument();
    expect(screen.getByText('ディーラークオリファイ')).toBeInTheDocument();
  });

  it.each([
    ['increase', { ...endPlayerWins, netChange: 10200 }, '+10200'],
    ['decrease', { ...endDealerWins, netChange: -300 }, '-300'],
    ['no wager', { ...betPhaseState, phase: 3, netChange: 0 }, '増減なし'],
  ])('shows net chip change for %s', async (_name, state, expected) => {
    mockApi.mockResolvedValue(state);
    renderWithProviders(<CasinoHoldemPage />);
    expect(await screen.findByTestId('net-change')).toHaveTextContent(expected);
  });

  it('shows localized net chip change in English', async () => {
    const previousLanguage = i18n.language;
    await i18n.changeLanguage('en');
    try {
      mockApi.mockResolvedValue({ ...betPhaseState, phase: 3, netChange: 0 });
      renderWithProviders(<CasinoHoldemPage />);
      expect(await screen.findByTestId('net-change')).toHaveTextContent('Net change: No net change');
    } finally {
      await i18n.changeLanguage(previousLanguage);
    }
  });

  it('shows a positive bonus payout', async () => {
    mockApi.mockResolvedValue({ ...endPlayerWins, bonusBet: 10, bonusPayout: 50 });
    renderWithProviders(<CasinoHoldemPage />);
    expect(await screen.findByTestId('payout-breakdown')).toHaveTextContent('AAボーナス: 50');
  });

  it('shows zero payouts for losing showdown bets', async () => {
    mockApi.mockResolvedValue({ ...endDealerWins, bonusBet: 10, bonusPayout: 0 });
    renderWithProviders(<CasinoHoldemPage />);
    const breakdown = await screen.findByTestId('payout-breakdown');
    expect(breakdown).toHaveTextContent('アンテ: 払戻しなし (0)');
    expect(breakdown).toHaveTextContent('コール: 払戻しなし (0)');
    expect(breakdown).toHaveTextContent('AAボーナス: 払戻しなし (0)');
  });

  it('shows no bet for the bonus row when bonusBet is zero', async () => {
    mockApi.mockResolvedValue({ ...endDealerWins, bonusBet: 0, bonusPayout: 0 });
    renderWithProviders(<CasinoHoldemPage />);
    const breakdown = await screen.findByTestId('payout-breakdown');
    expect(breakdown).toHaveTextContent('AAボーナス: ベットなし');
    expect(breakdown).not.toHaveTextContent('AAボーナス: 払戻しなし');
  });

  it('distinguishes no call bet on fold from zero ante and bonus payouts', async () => {
    mockApi.mockResolvedValue({ ...endFold, anteBet: 100, bonusBet: 10 });
    renderWithProviders(<CasinoHoldemPage />);
    const breakdown = await screen.findByTestId('payout-breakdown');
    expect(breakdown).toHaveTextContent('アンテ: 払戻しなし (0)');
    expect(breakdown).toHaveTextContent('コール: ベットなし');
    expect(breakdown).toHaveTextContent('AAボーナス: 払戻しなし (0)');
  });

  it('shows end phase with dealer wins', async () => {
    mockApi.mockResolvedValue(endDealerWins);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByText('ディーラー勝利！')).toBeInTheDocument());
  });

  it('shows end phase with fold', async () => {
    mockApi.mockResolvedValue(endFold);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByText('フォールド')).toBeInTheDocument());
  });

  it('shows dealer no-qualify message after call', async () => {
    mockApi.mockResolvedValue(endNoQualify);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByText('ディーラーノークオリファイ')).toBeInTheDocument());
  });

  it('changes ante and bonus amounts', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    const anteInput = screen.getByLabelText('アンテ');
    fireEvent.change(anteInput, { target: { value: '200' } });

    const bonusInput = screen.getByLabelText('AAボーナス');
    fireEvent.change(bonusInput, { target: { value: '10' } });

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 200, 10));
  });

  it('previews the total chips needed through a call and announces any shortfall', async () => {
    mockApi.mockResolvedValue({ ...betPhaseState, chips: 350 });
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 350')).toBeInTheDocument());

    const preview = screen.getByTestId('ch-bet-total-preview');
    expect(preview).toHaveAttribute('role', 'status');
    expect(preview).toHaveAttribute('aria-live', 'polite');
    expect(preview).toHaveTextContent('必要チップ合計（アンテ・ボーナス・コール）: 300');
    expect(preview).not.toHaveTextContent('チップ不足');

    fireEvent.change(screen.getByLabelText('アンテ'), { target: { value: '120' } });
    fireEvent.change(screen.getByLabelText('AAボーナス'), { target: { value: '20' } });
    expect(preview).toHaveTextContent('必要チップ合計（アンテ・ボーナス・コール）: 380');
    expect(preview).toHaveTextContent('チップ不足: 30');
    expect(preview).toHaveClass('text-ds-error-text');
  });

  it('shows a validation error and disables Bet for a non-multiple-of-10 ante', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText('アンテ'), { target: { value: '15' } });
    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ベット' })).toBeDisabled();
  });

  it('steps the ante up and down with the stepper buttons', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    const ante = screen.getByLabelText('アンテ') as HTMLInputElement;
    fireEvent.click(screen.getByRole('button', { name: 'アンテ +10' }));
    expect(ante.value).toBe('110');
    fireEvent.click(screen.getByRole('button', { name: 'アンテ −10' }));
    expect(ante.value).toBe('100');
  });

  it('shows network error', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockRejectedValueOnce(new Error('Network'));
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('renders board and player cards in flop phase', async () => {
    mockApi.mockResolvedValue(flopState);
    renderWithProviders(<CasinoHoldemPage />);
    // 2 player face-up + 3 community face-up + 2 dealer face-down = 7 imgs
    await waitFor(() => expect(screen.getAllByRole('img').length).toBe(7));
    expect(screen.getByText('🟡')).toBeInTheDocument();
    expect(screen.getByText('🔴')).toBeInTheDocument();
    expect(screen.getByText('🃏')).toBeInTheDocument();
  });

  it('shows the player current hand at flop (computed client-side)', async () => {
    // flopState is A-K-Q-J-10 all spades → Royal Flush.
    mockApi.mockResolvedValue(flopState);
    renderWithProviders(<CasinoHoldemPage />);
    const hand = await screen.findByTestId('ch-flop-hand');
    expect(hand).toHaveTextContent('現在の役');
    expect(hand).toHaveTextContent('ロイヤルフラッシュ');
  });

  it('shows a non-royal current hand at flop (one pair)', async () => {
    // A♠ + 10♣ with board 10♥ 4♦ 2♠ → a pair of tens.
    const onePairFlop: CasinoHoldemResponse = {
      ...flopState,
      playerHand: [card('SPADE', 1), card('CLOVER', 10)],
      community: [card('HEART', 10), card('DIAMOND', 4), card('SPADE', 2)],
    };
    mockApi.mockResolvedValue(onePairFlop);
    renderWithProviders(<CasinoHoldemPage />);
    const hand = await screen.findByTestId('ch-flop-hand');
    expect(hand).toHaveTextContent('ワンペア');
  });

  it('does not show the flop current-hand readout in the bet phase', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /ベット/ })).toBeInTheDocument());
    expect(screen.queryByTestId('ch-flop-hand')).not.toBeInTheDocument();
  });

  it('renders hint toggle checkbox', async () => {
    mockApi.mockResolvedValue(flopState);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /コール/ })).toBeInTheDocument());
    expect(screen.getByRole('checkbox')).toBeInTheDocument();
  });

  it('shows HintTooltip when hint is enabled at flop', async () => {
    localStorage.setItem('hint_enabled_casinoholdem', 'true');
    mockApi.mockResolvedValue({ ...flopState, playerHandRank: 1 });
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('next game button executes reset', async () => {
    mockApi.mockResolvedValue(endPlayerWins);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });
});

// --- keyboard shortcut execution (#4429) ---
// Every key here is gated on a specific phase, so each case pairs the key with
// the fixture that enables it. Asserting the exact command *and* arguments is
// the point: a table that only checked "exec was called" would still pass if two
// keys were swapped, which is the defect class these tests exist to catch.
const kbdCases: [string, unknown[], CasinoHoldemResponse][] = [
  ['b', ['bet', 100, 0], betPhaseState],
  ['c', ['call'], flopState],
  ['f', ['fold'], flopState],
  ['r', ['reset'], endPlayerWins],
];

describe('CasinoHoldemPage keyboard shortcuts', () => {
  it.each(kbdCases)('pressing %s dispatches %j', async (key, expected, state) => {
    mockApi.mockResolvedValue(state);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockApi.mockClear();
    mockApi.mockResolvedValue(state);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith(...expected));
  });

  it('ignores a key whose phase gate is closed', async () => {
    // 'b' places the ante and is enabled only in the BET phase.
    mockApi.mockResolvedValue(flopState);
    renderWithProviders(<CasinoHoldemPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockApi.mockClear();
    fireEvent.keyDown(document, { key: 'b' });
    await flushPendingDispatch();
    expect(mockApi).not.toHaveBeenCalled();
  });
});
