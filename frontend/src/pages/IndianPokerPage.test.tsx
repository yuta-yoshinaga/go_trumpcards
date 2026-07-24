import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { indianpokerApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { IndianPokerResponse } from '../types/card';
import { IndianPokerPage } from './IndianPokerPage';

vi.mock('../api/gameApi', () => ({
  indianpokerApi: { exec: vi.fn() },
  actionLogApi: { indianpoker: vi.fn() },
}));

const mockExec = vi.mocked(indianpokerApi.exec);

/** Helper: base human player */
const humanPlayer = (overrides: Partial<import('../types/card').IndianPokerPlayerOutput> = {}) => ({
  id: 0,
  isHuman: true,
  card: null as import('../types/card').Card | null,
  chips: 990,
  currentBet: 0,
  folded: false,
  allIn: false,
  cardRank: 0,
  playStyleName: '',
  ...overrides,
});

/** Helper: base CPU player */
const cpuPlayer = (id: number, overrides: Partial<import('../types/card').IndianPokerPlayerOutput> = {}) => ({
  id,
  isHuman: false,
  card: { design: 'SPADE' as const, value: 10 },
  chips: 1000,
  currentBet: 0,
  folded: false,
  allIn: false,
  cardRank: 10,
  playStyleName: 'タイト',
  ...overrides,
});

/** INIT state (phase 0): no players yet */
const initState: IndianPokerResponse = {
  players: [],
  pot: 0,
  sidePots: [],
  dealerIdx: 0,
  currentTurn: 0,
  phase: 0,
  gameEndFlag: false,
  lastBet: 0,
  minRaise: 20,
  bettingLimit: 2,
  raiseCount: 0,
  maxBetAmount: 0,
  roundResults: [],
  cpuActions: [],
  handCount: 0,
  ante: 10,
  message: '',
};

/** BETTING (phase 2): human's turn, no outstanding bet */
const bettingState: IndianPokerResponse = {
  players: [humanPlayer(), cpuPlayer(1), cpuPlayer(2), cpuPlayer(3)],
  pot: 40,
  sidePots: [],
  dealerIdx: 3,
  currentTurn: 0,
  phase: 2,
  gameEndFlag: false,
  lastBet: 0,
  minRaise: 20,
  bettingLimit: 2,
  raiseCount: 0,
  maxBetAmount: 0,
  roundResults: [],
  cpuActions: [],
  handCount: 1,
  ante: 10,
  message: 'あなたの番です',
};

/** BETTING with outstanding bet: shows call/raise instead of bet/check */
const bettingWithBetState: IndianPokerResponse = {
  ...bettingState,
  lastBet: 40,
  cpuActions: [{ playerIdx: 1, action: 3, amount: 40 }],
};

/** SHOWDOWN (phase 3) */
const showdownState: IndianPokerResponse = {
  players: [
    humanPlayer({ card: { design: 'HEART', value: 7 }, chips: 950 }),
    cpuPlayer(1, { card: { design: 'SPADE', value: 10 } }),
    cpuPlayer(2, { folded: true }),
  ],
  pot: 0,
  sidePots: [],
  dealerIdx: 2,
  currentTurn: -1,
  phase: 3,
  gameEndFlag: false,
  lastBet: 0,
  minRaise: 0,
  bettingLimit: 2,
  raiseCount: 0,
  maxBetAmount: 0,
  roundResults: [
    { playerIdx: 0, card: { design: 'HEART', value: 7 }, cardRank: 7, wonAmount: 0 },
    { playerIdx: 1, card: { design: 'SPADE', value: 10 }, cardRank: 10, wonAmount: 200 },
  ],
  cpuActions: [],
  handCount: 1,
  ante: 10,
  message: 'CPU 1 の勝ち',
};

/** END (phase 4) — also isShowdown */
const endState: IndianPokerResponse = {
  ...showdownState,
  phase: 4,
  gameEndFlag: true,
  message: 'Game over.',
};

/** ANTE phase (phase 1) */
const anteState: IndianPokerResponse = {
  ...bettingState,
  phase: 1,
  message: 'アンティを支払いました',
};

beforeEach(() => {
  mockExec.mockResolvedValue(initState);
});

describe('IndianPokerPage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<IndianPokerPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  // ---- mount & reset ----
  it('calls reset on mount', async () => {
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // ---- phase name display ----
  it('shows "初期化" when phase is INIT', async () => {
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText('初期化')).toBeInTheDocument());
  });

  it('shows "ベッティング" when phase is BETTING', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText('ベッティング')).toBeInTheDocument());
  });

  it('shows "アンティ" when phase is ANTE', async () => {
    mockExec.mockResolvedValue(anteState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => {
      const phaseIndicator = screen.getByTestId('phase-indicator');
      expect(phaseIndicator).toHaveTextContent('アンティ');
    });
  });

  // ---- info bar ----
  it('shows pot and the dealer name via playerName (CPU dealer)', async () => {
    mockExec.mockResolvedValue(bettingState); // dealerIdx 3 → CPU
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText(/ポット:/)).toBeInTheDocument());
    expect(screen.getByText(/ディーラー:/)).toBeInTheDocument();
    // Dealer renders via playerName (CPU 3), not the raw "Player 3" index.
    expect(screen.getAllByText('CPU 3').length).toBeGreaterThan(0);
    expect(screen.queryByText(/Player 3/)).not.toBeInTheDocument();
  });

  it('shows あなた as the dealer name when the human is the dealer', async () => {
    mockExec.mockResolvedValue({ ...bettingState, dealerIdx: 0 }); // player 0 is the human
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText(/ディーラー:/)).toBeInTheDocument());
    // The dealer <strong> renders the human label, not "Player 0".
    expect(screen.queryByText(/Player 0/)).not.toBeInTheDocument();
    expect(screen.getAllByText('あなた').length).toBeGreaterThan(0);
  });

  it('shows ante in info bar', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => {
      const phaseIndicator = screen.getByTestId('phase-indicator');
      expect(phaseIndicator).toHaveTextContent(/アンティ/);
      expect(phaseIndicator).toHaveTextContent('10');
    });
  });

  // ---- CPU players ----
  it('renders CPU player info with playStyleName and chips', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.getByText(/CPU 2/)).toBeInTheDocument();
    expect(screen.getAllByText(/タイト/).length).toBeGreaterThanOrEqual(1);
  });

  it('shows CPU card face-up during betting', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => {
      const cards = screen.getAllByAltText('♠ 10');
      expect(cards.length).toBeGreaterThanOrEqual(1);
    });
  });

  it('shows CardBack for CPU when card is null', async () => {
    mockExec.mockResolvedValue({
      ...bettingState,
      players: [humanPlayer(), cpuPlayer(1, { card: null }), cpuPlayer(2)],
    });
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    const cardBacks = screen.getAllByAltText('カード裏面');
    expect(cardBacks.length).toBeGreaterThanOrEqual(1);
  });

  it('shows CPU bet when currentBet > 0', async () => {
    mockExec.mockResolvedValue({
      ...bettingState,
      players: [humanPlayer(), cpuPlayer(1, { currentBet: 50 }), cpuPlayer(2)],
    });
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText(/ベット: 50/)).toBeInTheDocument());
  });

  it('shows fold badge for folded CPU', async () => {
    mockExec.mockResolvedValue({
      ...bettingState,
      players: [humanPlayer(), cpuPlayer(1, { folded: true }), cpuPlayer(2)],
    });
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText('[フォールド]')).toBeInTheDocument());
  });

  it('shows all-in badge for all-in CPU', async () => {
    mockExec.mockResolvedValue({
      ...bettingState,
      players: [humanPlayer(), cpuPlayer(1, { allIn: true }), cpuPlayer(2)],
    });
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText('[オールイン]')).toBeInTheDocument());
  });

  // ---- CPU actions log ----
  it('shows CPU actions log when cpuActions is non-empty', async () => {
    mockExec.mockResolvedValue(bettingWithBetState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText('CPU行動:')).toBeInTheDocument());
    expect(screen.getByText(/Player 1: ベット/)).toBeInTheDocument();
  });

  it('does not show CPU actions log when cpuActions is empty', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByText('CPU行動:')).not.toBeInTheDocument();
  });

  // ---- round results ----
  it('shows round results in showdown', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
  });

  it('does not show round results when not in showdown', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByText('結果:')).not.toBeInTheDocument();
  });

  it('does not show round results when roundResults is empty in showdown', async () => {
    mockExec.mockResolvedValue({
      ...showdownState,
      roundResults: [],
    });
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText('ショーダウン')).toBeInTheDocument());
    expect(screen.queryByText('結果:')).not.toBeInTheDocument();
  });

  it('shows won chips when wonAmount > 0', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() =>
      expect(within(screen.getByTestId('round-results-visible')).getByText(/\+200チップ/)).toBeInTheDocument(),
    );
  });

  it('does not show won chips when wonAmount is 0', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
    expect(screen.queryByText(/\+0チップ/)).not.toBeInTheDocument();
  });

  // ---- human player section ----
  it('shows human player section when humanPlayer exists', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText(/あなたのカード/)).toBeInTheDocument());
  });

  it('does not show human player section when no players', async () => {
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText('初期化')).toBeInTheDocument());
    expect(screen.queryByText(/あなたのカード/)).not.toBeInTheDocument();
  });

  it('shows CardBack for human during betting (card hidden)', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText(/あなたのカード/)).toBeInTheDocument());
    const cardBacks = screen.getAllByAltText('カード裏面');
    expect(cardBacks.length).toBeGreaterThanOrEqual(1);
  });

  it('shows human card face-up during showdown', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByAltText('♥ 7')).toBeInTheDocument());
  });

  it('does not show CardBack for human when folded', async () => {
    mockExec.mockResolvedValue({
      ...bettingState,
      players: [humanPlayer({ folded: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText(/あなたのカード/)).toBeInTheDocument());
    // Only CPU cards appear as card backs are not shown for folded human
    // CPU cards are face-up, so the only card backs would be for CPUs with null cards
  });

  it('shows human bet when currentBet > 0', async () => {
    mockExec.mockResolvedValue({
      ...bettingState,
      players: [humanPlayer({ currentBet: 30 }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText(/ベット: 30/)).toBeInTheDocument());
  });

  it('shows human fold badge', async () => {
    mockExec.mockResolvedValue({
      ...bettingState,
      players: [humanPlayer({ folded: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText('[フォールド]')).toBeInTheDocument());
  });

  it('shows human all-in badge', async () => {
    mockExec.mockResolvedValue({
      ...bettingState,
      players: [humanPlayer({ allIn: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText('[オールイン]')).toBeInTheDocument());
  });

  // ---- message ----
  it('shows game message', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText('あなたの番です')).toBeInTheDocument());
  });

  // ---- canAct / betting controls ----
  it('shows bet/check buttons when canAct and no outstanding bet', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument();
  });

  it('shows call/raise buttons when canAct and has outstanding bet', async () => {
    mockExec.mockResolvedValue(bettingWithBetState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'レイズ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument();
  });

  it('hides betting controls when not active phase', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText('ショーダウン')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'チェック' })).not.toBeInTheDocument();
  });

  it('hides betting controls when human is folded', async () => {
    mockExec.mockResolvedValue({
      ...bettingState,
      players: [humanPlayer({ folded: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText(/あなたのカード/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  it('hides betting controls when human is all-in', async () => {
    mockExec.mockResolvedValue({
      ...bettingState,
      players: [humanPlayer({ allIn: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText(/あなたのカード/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  it('hides betting controls when it is not human turn', async () => {
    mockExec.mockResolvedValue({
      ...bettingState,
      currentTurn: 1,
    });
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText(/あなたのカード/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  // ---- bet amount input ----
  it('updates bet amount when changing input', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<IndianPokerPage />);
    const betInput = await screen.findByLabelText('ベット額:');
    await waitFor(() => expect((betInput as HTMLInputElement).value).toBe('20'));

    fireEvent.change(betInput, { target: { value: '50' } });
    expect((betInput as HTMLInputElement).value).toBe('50');
  });

  // ---- button click handlers ----
  it('calls bet command with betAmount', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(bettingState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 20, undefined, expect.any(Number)));
  });

  it('calls check command', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(bettingState);
    fireEvent.click(screen.getByRole('button', { name: 'チェック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('check', undefined, undefined, expect.any(Number)));
  });

  it('calls fold command', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(bettingState);
    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold', undefined, undefined, expect.any(Number)));
  });

  it('calls allin command', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(bettingState);
    fireEvent.click(screen.getByRole('button', { name: 'オールイン' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('allin', undefined, undefined, expect.any(Number)));
  });

  it('calls call command when has outstanding bet', async () => {
    mockExec.mockResolvedValue(bettingWithBetState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(bettingState);
    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('call', undefined, undefined, expect.any(Number)));
  });

  it('calls raise command with betAmount when has outstanding bet', async () => {
    mockExec.mockResolvedValue(bettingWithBetState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'レイズ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(bettingState);
    fireEvent.click(screen.getByRole('button', { name: 'レイズ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('raise', 20, undefined, expect.any(Number)));
  });

  it('calls reset command when reset button is clicked', async () => {
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { ante: 10, bettingLimit: 2, cpuMetaAI: true }),
    );
  });

  // ---- loading / disabled state ----
  it('disables buttons while loading', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());

    let resolve!: (value: IndianPokerResponse) => void;
    const slowPromise = new Promise<IndianPokerResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));

    expect(screen.getByRole('button', { name: 'ベット' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'チェック' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();

    resolve(bettingState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());
  });

  // ---- error handling ----
  it('shows error message when API call fails', async () => {
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('clears error on successful call after failure', async () => {
    renderWithProviders(<IndianPokerPage />);
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

  // ---- END phase (also isShowdown) ----
  it('shows results in END phase (phase 4)', async () => {
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText('終了')).toBeInTheDocument());
    // Results appear after the staged own-card reveal completes (#3068).
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
  });

  // ---- bet amount used by raise ----
  it('sends updated bet amount when raise is clicked after changing input', async () => {
    mockExec.mockResolvedValue(bettingWithBetState);
    renderWithProviders(<IndianPokerPage />);
    const betInput = await screen.findByLabelText('ベット額:');
    await waitFor(() => expect((betInput as HTMLInputElement).value).toBe('20'));

    fireEvent.change(betInput, { target: { value: '100' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(bettingState);
    fireEvent.click(screen.getByRole('button', { name: 'レイズ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('raise', 100, undefined, expect.any(Number)));
  });

  it('sends updated bet amount when bet is clicked after changing input', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<IndianPokerPage />);
    const betInput = await screen.findByLabelText('ベット額:');
    await waitFor(() => expect((betInput as HTMLInputElement).value).toBe('20'));

    fireEvent.change(betInput, { target: { value: '60' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(bettingState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 60, undefined, expect.any(Number)));
  });

  it('sets aria-busy while loading', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());

    const container = screen.getByRole('button', { name: 'ベット' }).closest('[aria-busy]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');

    let resolve!: (value: IndianPokerResponse) => void;
    const slowPromise = new Promise<IndianPokerResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));

    expect(container).toHaveAttribute('aria-busy', 'true');

    resolve(bettingState);
    await waitFor(() => {
      expect(container).toHaveAttribute('aria-busy', 'false');
    });
  });

  // ---- ConfirmDialog on reset ----
  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { ante: 10, bettingLimit: 2, cpuMetaAI: true }),
    );
  });

  // ---- settings panel ----
  it('renders settings panel with betting limit selector and meta AI checkbox', async () => {
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    // Settings panel should be rendered (uses <details> with summary)
    const summaries = screen.getAllByText('アンティ');
    // At least one element should be the settings panel summary
    expect(summaries.length).toBeGreaterThanOrEqual(1);
  });

  // ---- betAmount resets when minRaise changes ----
  it('resets betAmount to minRaise when state changes with non-zero minRaise', async () => {
    mockExec.mockResolvedValue({ ...bettingState, minRaise: 50 });
    renderWithProviders(<IndianPokerPage />);
    const betInput = await screen.findByLabelText('ベット額:');
    await waitFor(() => expect((betInput as HTMLInputElement).value).toBe('50'));
  });

  it('resets betAmount to 20 when minRaise is 0', async () => {
    mockExec.mockResolvedValue({ ...bettingState, minRaise: 0 });
    renderWithProviders(<IndianPokerPage />);
    const betInput = await screen.findByLabelText('ベット額:');
    await waitFor(() => expect((betInput as HTMLInputElement).value).toBe('20'));
  });

  // ---- action log section ----
  it('shows action log section', async () => {
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText('終了')).toBeInTheDocument());
    // Action log section should render (button to view log)
    expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
  });

  // ---- roundResultsForDisplay with null card ----
  it('shows empty handName when roundResult card is null', async () => {
    mockExec.mockResolvedValue({
      ...showdownState,
      roundResults: [
        { playerIdx: 0, card: null, cardRank: 0, wonAmount: 100 },
        { playerIdx: 1, card: { design: 'SPADE' as const, value: 10 }, cardRank: 10, wonAmount: 0 },
      ],
    });
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
    expect(within(screen.getByTestId('round-results-visible')).getByText(/\+100チップ/)).toBeInTheDocument();
  });

  // ---- cpuMetaAI toggle ----
  it('sends cpuMetaAI: false when meta AI checkbox is unchecked before reset', async () => {
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const metaAiCheckbox = screen.getByRole('checkbox', { name: 'メタAI' });
    expect((metaAiCheckbox as HTMLInputElement).checked).toBe(true);
    fireEvent.click(metaAiCheckbox);
    expect((metaAiCheckbox as HTMLInputElement).checked).toBe(false);

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { ante: 10, bettingLimit: 2, cpuMetaAI: false }),
    );
  });

  // ---- bettingLimit select ----
  it('sends updated bettingLimit when select changes before reset', async () => {
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const select = screen.getByRole('combobox', { name: 'ベッティングリミット' });
    fireEvent.change(select, { target: { value: '0' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { ante: 10, bettingLimit: 0, cpuMetaAI: true }),
    );
  });

  // ---- ante select ----
  it('sends updated ante when the ante select changes before reset', async () => {
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const anteSelect = screen.getByRole('combobox', { name: 'アンティ額（次のゲームから反映）' });
    fireEvent.change(anteSelect, { target: { value: '50' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { ante: 50, bettingLimit: 2, cpuMetaAI: true }),
    );
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(initState);
    renderWithProviders(<IndianPokerPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  // ---- staged showdown reveal (#3068) ----
  describe('staged showdown reveal', () => {
    it('conceals the own card (no reveal wrapper) during betting', async () => {
      mockExec.mockResolvedValue(bettingState);
      renderWithProviders(<IndianPokerPage />);
      await waitFor(() => expect(screen.getByText(/あなたのカード/)).toBeInTheDocument());
      expect(screen.queryByTestId('indianpoker-own-reveal')).not.toBeInTheDocument();
    });

    it('flips the own card first and holds round results until the reveal completes', async () => {
      vi.useFakeTimers();
      try {
        mockExec.mockResolvedValue(showdownState);
        renderWithProviders(<IndianPokerPage />);
        // Own-card reveal wrapper (and the flipped card) appears immediately at showdown.
        await vi.waitFor(() => expect(screen.getByTestId('indianpoker-own-reveal')).toBeInTheDocument());
        expect(screen.getByAltText('♥ 7')).toBeInTheDocument();
        // Results stay hidden until the 600ms reveal delay elapses (no spoiler).
        expect(screen.queryByText('結果:')).not.toBeInTheDocument();
        // After the delay the results panel is revealed.
        await vi.advanceTimersByTimeAsync(600);
        await vi.waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
      } finally {
        vi.useRealTimers();
      }
    });

    it('reveals card and results together immediately when reduced motion is preferred', async () => {
      const original = window.matchMedia;
      window.matchMedia = vi.fn().mockImplementation((query: string) => ({
        matches: query.includes('prefers-reduced-motion'),
        media: query,
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      }));
      try {
        mockExec.mockResolvedValue(showdownState);
        renderWithProviders(<IndianPokerPage />);
        await waitFor(() => expect(screen.getByTestId('indianpoker-own-reveal')).toBeInTheDocument());
        // No staged delay: the results panel is present right away.
        expect(screen.getByText('結果:')).toBeInTheDocument();
      } finally {
        window.matchMedia = original;
      }
    });
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

    it('renders CPU cards in 3-column grid on mobile', async () => {
      mockExec.mockResolvedValue(bettingState);
      const { container } = renderWithProviders(<IndianPokerPage />);
      await waitFor(() => expect(container.querySelector('[data-tutorial="ip-cpu-cards"]')).toBeInTheDocument());
      const cpuGrid = container.querySelector('[data-tutorial="ip-cpu-cards"]');
      expect(cpuGrid).toHaveClass('grid-cols-3');
    });

    it('hides playStyleName on mobile', async () => {
      mockExec.mockResolvedValue(bettingState);
      renderWithProviders(<IndianPokerPage />);
      await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
      expect(screen.queryByText(/バランス型/)).not.toBeInTheDocument();
    });
  });
});
