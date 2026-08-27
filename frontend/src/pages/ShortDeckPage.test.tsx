import { act, fireEvent, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, shortdeckApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { ShortDeckResponse } from '../types/card';
import { ShortDeckPage } from './ShortDeckPage';

vi.mock('../api/gameApi', () => ({
  shortdeckApi: { exec: vi.fn() },
  actionLogApi: { shortdeck: vi.fn() },
}));

const mockExec = vi.mocked(shortdeckApi.exec);

/** Helper: base human player */
const humanPlayer = (overrides: Partial<import('../types/card').ShortDeckPlayerData> = {}) => ({
  id: 0,
  isHuman: true,
  cards: [
    { design: 'SPADE' as const, value: 1 },
    { design: 'HEART' as const, value: 13 },
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
const cpuPlayer = (id: number, overrides: Partial<import('../types/card').ShortDeckPlayerData> = {}) => ({
  id,
  isHuman: false,
  cards: [
    { design: 'DIAMOND' as const, value: 6 },
    { design: 'CLOVER' as const, value: 7 },
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

/** INIT state (phase 0): no players yet */
const initState: ShortDeckResponse = {
  players: [],
  communityCards: [],
  pot: 0,
  sidePots: [],
  dealerIdx: 0,
  currentTurn: 0,
  phase: 0,
  gameEndFlag: false,
  lastBet: 0,
  minRaise: 20,
  roundResults: [],
  cpuActions: [],
  message: '',
  handCount: 0,
  smallBlind: 5,
  bigBlind: 10,
  tournamentMode: false,
  blindLevelHands: 10,
  blindMultiplier: 200,
  bettingLimit: 0,
  raiseCount: 0,
  maxBetAmount: 0,
  tableSize: 4,
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

/** PRE_FLOP (phase 1): human's turn, no outstanding bet */
const preFlopState: ShortDeckResponse = {
  players: [humanPlayer(), cpuPlayer(1), cpuPlayer(2), cpuPlayer(3)],
  communityCards: [],
  pot: 30,
  sidePots: [],
  dealerIdx: 3,
  currentTurn: 0,
  phase: 1,
  gameEndFlag: false,
  lastBet: 0,
  minRaise: 20,
  roundResults: [],
  cpuActions: [],
  message: 'あなたの番です',
  handCount: 1,
  smallBlind: 5,
  bigBlind: 10,
  tournamentMode: false,
  blindLevelHands: 10,
  blindMultiplier: 200,
  bettingLimit: 0,
  raiseCount: 0,
  maxBetAmount: 0,
  tableSize: 4,
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

/** PRE_FLOP with outstanding bet: shows call/raise instead of bet/check */
const preFlopWithBetState: ShortDeckResponse = {
  ...preFlopState,
  lastBet: 40,
  cpuActions: [{ playerIdx: 1, action: 3, amount: 40 }],
};

/** FLOP (phase 2) with community cards */
const flopState: ShortDeckResponse = {
  ...preFlopState,
  phase: 2,
  communityCards: [
    { design: 'SPADE', value: 10 },
    { design: 'HEART', value: 6 },
    { design: 'DIAMOND', value: 8 },
  ],
};

/** SHOWDOWN (phase 5) */
const showdownState: ShortDeckResponse = {
  players: [
    humanPlayer({ handName: 'ワンペア', currentBet: 0, chips: 950 }),
    cpuPlayer(1, {
      handName: 'ツーペア',
      folded: false,
      cards: [
        { design: 'SPADE', value: 6 },
        { design: 'HEART', value: 8 },
      ],
    }),
    cpuPlayer(2, { folded: true }),
  ],
  communityCards: [
    { design: 'SPADE', value: 10 },
    { design: 'HEART', value: 6 },
    { design: 'DIAMOND', value: 8 },
    { design: 'CLOVER', value: 9 },
    { design: 'HEART', value: 9 },
  ],
  pot: 0,
  sidePots: [],
  dealerIdx: 2,
  currentTurn: -1,
  phase: 5,
  gameEndFlag: false,
  lastBet: 0,
  minRaise: 0,
  roundResults: [
    { playerIdx: 0, handRank: 1, handName: 'ワンペア', kickers: 'A, Q, 10', bestHand: [], wonAmount: 0, mucked: false },
    { playerIdx: 1, handRank: 2, handName: 'ツーペア', kickers: '8', bestHand: [], wonAmount: 200, mucked: false },
  ],
  cpuActions: [],
  message: 'CPU 1 の勝ち',
  handCount: 1,
  smallBlind: 5,
  bigBlind: 10,
  tournamentMode: false,
  blindLevelHands: 10,
  blindMultiplier: 200,
  bettingLimit: 0,
  raiseCount: 0,
  maxBetAmount: 0,
  tableSize: 4,
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

/** END (phase 6) -- also isShowdown */
const endState: ShortDeckResponse = {
  ...showdownState,
  phase: 6,
  message: 'Game over.',
};

beforeEach(() => {
  mockExec.mockResolvedValue(initState);
});

describe('ShortDeckPage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<ShortDeckPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  // ---- mount & reset ----
  it('calls reset on mount', async () => {
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // ---- phase name display ----
  it('shows "初期化中" when phase is INIT (not in PHASE_NAMES)', async () => {
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText('初期化中')).toBeInTheDocument());
  });

  it('shows known phase name for PRE_FLOP', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText('プリフロップ')).toBeInTheDocument());
  });

  // ---- info bar ----
  it('shows pot and the dealer name via playerName (CPU dealer)', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText(/ポット:/)).toBeInTheDocument());
    expect(screen.getByText(/ディーラー:/)).toBeInTheDocument();
    // Dealer renders via playerName (CPU 3), not the raw index.
    expect(screen.getAllByText('CPU 3').length).toBeGreaterThan(0);
    expect(screen.queryByText(/Player 3|プレイヤー 3/)).not.toBeInTheDocument();
  });

  // ---- community cards ----
  it('shows 5 CardBack placeholders when communityCards is empty', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText('コミュニティカード')).toBeInTheDocument());
    const cardBacks = screen.getAllByAltText('カード裏面');
    expect(cardBacks.length).toBeGreaterThanOrEqual(5);
  });

  it('shows CardImage when communityCards has cards', async () => {
    mockExec.mockResolvedValue(flopState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 10')).toBeInTheDocument());
    expect(screen.getByAltText('♥ 6')).toBeInTheDocument();
    expect(screen.getByAltText('♦ 8')).toBeInTheDocument();
  });

  // ---- CPU players ----
  it('renders CPU player info with playStyleName and chips', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.getByText(/CPU 2/)).toBeInTheDocument();
    expect(screen.getAllByText(/タイト/).length).toBeGreaterThanOrEqual(1);
  });

  it('shows CPU bet when currentBet > 0', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer(), cpuPlayer(1, { currentBet: 50 }), cpuPlayer(2)],
    });
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText(/ベット: 50/)).toBeInTheDocument());
  });

  it('does not show CPU bet when currentBet is 0', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByText(/ベット: 0/)).not.toBeInTheDocument();
  });

  it('shows fold badge for folded CPU', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer(), cpuPlayer(1, { folded: true }), cpuPlayer(2)],
    });
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText('[フォールド]')).toBeInTheDocument());
  });

  it('shows all-in badge for all-in CPU', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer(), cpuPlayer(1, { allIn: true }), cpuPlayer(2)],
    });
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText('[オールイン]')).toBeInTheDocument());
  });

  it('shows CPU hand name badge during showdown when not folded', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<ShortDeckPage />);
    // 'ツーペア' appears in the CPU badge and also in the always-present hand-rank reference.
    await waitFor(() => expect(screen.getAllByText('ツーペア').length).toBeGreaterThanOrEqual(2));
  });

  it('appends the Short Deck rank-override marker to the human hand-name badge at showdown', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<ShortDeckPage />);
    const marker = await screen.findByTestId('shortdeck-handname-rule');
    expect(marker).toHaveAttribute('title', 'ショートデック特殊ルール：フラッシュ > フルハウス');
    expect(marker).toHaveAttribute('aria-label', 'ショートデック特殊ルール：フラッシュ > フルハウス');
  });

  it('toggles the rank-override note on tap so touch users can read it', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<ShortDeckPage />);
    const marker = await screen.findByTestId('shortdeck-handname-rule');
    // It is a real button (focusable, activatable), not a hover-only span.
    expect(marker.tagName).toBe('BUTTON');
    expect(screen.queryByTestId('shortdeck-rule-note')).not.toBeInTheDocument();
    fireEvent.click(marker);
    expect(screen.getByTestId('shortdeck-rule-note')).toHaveTextContent('ショートデック特殊ルール');
    fireEvent.click(marker);
    expect(screen.queryByTestId('shortdeck-rule-note')).not.toBeInTheDocument();
  });

  it('shows CPU cards face-up during showdown when not folded', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 6')).toBeInTheDocument());
    expect(screen.getByAltText('♥ 8')).toBeInTheDocument();
  });

  it('shows CardBack for CPU cards when not in showdown', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    const cardBacks = screen.getAllByAltText('カード裏面');
    expect(cardBacks.length).toBeGreaterThanOrEqual(4);
  });

  // ---- CPU actions log ----
  it('shows CPU actions log when cpuActions is non-empty', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText('CPU行動:')).toBeInTheDocument());
    expect(screen.getByText(/Player 1: ベット/)).toBeInTheDocument();
  });

  it('does not show CPU actions log when cpuActions is empty', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByText('CPU行動:')).not.toBeInTheDocument();
  });

  // ---- round results ----
  it('shows round results in showdown', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
  });

  it('does not show round results when not in showdown', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByText('結果:')).not.toBeInTheDocument();
  });

  // ---- human player section ----
  it('shows human player section when humanPlayer exists', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
  });

  it('does not show human player section when no players', async () => {
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText('初期化中')).toBeInTheDocument());
    expect(screen.queryByText(/あなたの手札/)).not.toBeInTheDocument();
  });

  it('shows human cards when cards exist', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.getByAltText('♥ K')).toBeInTheDocument();
  });

  it('shows CardBack for human when cards is empty and not folded', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ cards: [] }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    const cardBacks = screen.getAllByAltText('カード裏面');
    expect(cardBacks.length).toBeGreaterThanOrEqual(2);
  });

  it('shows human fold badge', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ folded: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText('[フォールド]')).toBeInTheDocument());
  });

  it('shows human all-in badge', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ allIn: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText('[オールイン]')).toBeInTheDocument());
  });

  it('shows human hand name badge during showdown when not folded', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<ShortDeckPage />);
    // 'ワンペア' appears in the human badge and also in the always-present hand-rank reference.
    await waitFor(() => expect(screen.getAllByText('ワンペア').length).toBeGreaterThanOrEqual(2));
  });

  it('shows the Short Deck hand-ranking reference with the flush>full-house note', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<ShortDeckPage />);
    const ref = await screen.findByTestId('sd-handrank-reference');
    expect(ref).toHaveTextContent('役の強さ');
    expect(ref).toHaveTextContent('フラッシュ');
    expect(ref).toHaveTextContent('フルハウス');
    expect(ref).toHaveTextContent('6-7-8-9-A');
  });

  // ---- message ----
  it('shows game message', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText('あなたの番です')).toBeInTheDocument());
  });

  // ---- canAct / betting controls ----
  it('shows bet/check buttons when canAct and no outstanding bet', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument();
  });

  it('shows call/raise buttons when canAct and has outstanding bet', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'レイズ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument();
  });

  it('hides betting controls when not active phase', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText('ショーダウン')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'チェック' })).not.toBeInTheDocument();
  });

  it('hides betting controls when human is folded', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ folded: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  it('hides betting controls when human is all-in', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ allIn: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  it('hides betting controls when it is not human turn', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      currentTurn: 1,
    });
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  // ---- bet amount input ----
  it('updates bet amount when changing input', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    const betInput = await screen.findByLabelText('ベット額:');
    await waitFor(() => expect((betInput as HTMLInputElement).value).toBe('20'));

    fireEvent.change(betInput, { target: { value: '50' } });
    expect((betInput as HTMLInputElement).value).toBe('50');
  });

  // ---- button click handlers ----
  it('calls bet command with betAmount', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 20, undefined, 0));
  });

  it('calls check command', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'チェック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('check', undefined, undefined, 0));
  });

  it('calls fold command', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold', undefined, undefined, 0));
  });

  it('calls allin command', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'オールイン' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('allin', undefined, undefined, 0));
  });

  it('calls call command when has outstanding bet', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('call', undefined, undefined, 0));
  });

  it('calls raise command with betAmount when has outstanding bet', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'レイズ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'レイズ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('raise', 20, undefined, 0));
  });

  it('calls reset command when reset button is clicked', async () => {
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuMetaAI: false, tournamentMode: false }),
    );
  });

  it('uses outline style for reset button', async () => {
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const resetBtn = screen.getByRole('button', { name: 'リセット' });
    expect(resetBtn.className).toContain('bg-transparent');
    expect(resetBtn.className).toContain('border');
  });

  // ---- loading / disabled state ----
  it('disables buttons while loading', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());

    let resolve!: (value: ShortDeckResponse) => void;
    const slowPromise = new Promise<ShortDeckResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));

    expect(screen.getByRole('button', { name: 'ベット' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'チェック' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();

    resolve(preFlopState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());
  });

  // ---- error handling ----
  it('shows error message when API call fails', async () => {
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('clears error on successful call after failure', async () => {
    renderWithProviders(<ShortDeckPage />);
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
  it('shows results in END phase (phase 6)', async () => {
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText('結果')).toBeInTheDocument());
    expect(screen.getByText('結果:')).toBeInTheDocument();
  });

  // ---- SB/BB info bar ----
  it('shows SB/BB in info bar', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText(/SB\/BB:/)).toBeInTheDocument());
    expect(screen.getByText(/5\/10/)).toBeInTheDocument();
  });

  // ---- tournament mode ----
  it('shows tournament info when tournamentMode is true', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      tournamentMode: true,
      handCount: 7,
      blindLevelHands: 5,
      smallBlind: 20,
      bigBlind: 40,
    });
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText(/ハンド#/)).toBeInTheDocument());
    expect(screen.getByText(/レベルアップ:5ハンド毎/)).toBeInTheDocument();
    expect(screen.getByText(/20\/40/)).toBeInTheDocument();
  });

  it('does not show tournament info when tournamentMode is false', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText(/SB\/BB:/)).toBeInTheDocument());
    expect(screen.queryByText(/ハンド#/)).not.toBeInTheDocument();
    expect(screen.queryByText(/レベルアップ:/)).not.toBeInTheDocument();
  });

  // ---- rebuy phase ----
  it('shows rebuy prompt and buttons in rebuy phase', async () => {
    const rebuyState: ShortDeckResponse = {
      ...preFlopState,
      phase: 7,
      rebuyPhaseType: 1,
      rebuyChips: 1000,
      rebuyMaxCount: 3,
      rebuyCounts: [0, 0, 0, 0],
      addonChips: 1500,
    };
    mockExec.mockResolvedValue(rebuyState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText(/リバイしますか/)).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'リバイ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument();
  });

  it('calls rebuy command when rebuy accept button is clicked', async () => {
    const rebuyState: ShortDeckResponse = {
      ...preFlopState,
      phase: 7,
      rebuyPhaseType: 1,
      rebuyChips: 1000,
      rebuyMaxCount: 3,
      rebuyCounts: [0, 0, 0, 0],
    };
    mockExec.mockResolvedValue(rebuyState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リバイ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'リバイ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('rebuy'));
  });

  it('calls skiprebuy command when rebuy skip button is clicked', async () => {
    const rebuyState: ShortDeckResponse = {
      ...preFlopState,
      phase: 7,
      rebuyPhaseType: 1,
      rebuyChips: 1000,
      rebuyMaxCount: 3,
      rebuyCounts: [0, 0, 0, 0],
    };
    mockExec.mockResolvedValue(rebuyState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByTestId('rebuy-controls')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    const rebuyControls = screen.getByTestId('rebuy-controls');
    fireEvent.click(within(rebuyControls).getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skiprebuy'));
  });

  // ---- addon phase ----
  it('shows addon prompt and buttons in addon phase', async () => {
    const addonState: ShortDeckResponse = {
      ...preFlopState,
      phase: 7,
      rebuyPhaseType: 2,
      addonChips: 1500,
      rebuyCounts: [0, 0, 0, 0],
    };
    mockExec.mockResolvedValue(addonState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText(/アドオンしますか/)).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'アドオン' })).toBeInTheDocument();
  });

  it('calls addon command when addon accept button is clicked', async () => {
    const addonState: ShortDeckResponse = {
      ...preFlopState,
      phase: 7,
      rebuyPhaseType: 2,
      addonChips: 1500,
      rebuyCounts: [0, 0, 0, 0],
    };
    mockExec.mockResolvedValue(addonState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'アドオン' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'アドオン' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('addon'));
  });

  it('calls skipaddon command when addon skip button is clicked', async () => {
    const addonState: ShortDeckResponse = {
      ...preFlopState,
      phase: 7,
      rebuyPhaseType: 2,
      addonChips: 1500,
      rebuyCounts: [0, 0, 0, 0],
    };
    mockExec.mockResolvedValue(addonState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText(/アドオンしますか/)).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    const addonControls = screen.getByTestId('addon-controls');
    fireEvent.click(within(addonControls).getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skipaddon'));
  });

  // ---- muck phase ----
  it('shows muck controls when phase=SHOWDOWN and muckAvailable=true', async () => {
    const muckState: ShortDeckResponse = {
      ...showdownState,
      muckAvailable: true,
    };
    mockExec.mockResolvedValue(muckState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByTestId('muck-controls')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'マック' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ショー' })).toBeInTheDocument();
  });

  it('calls muck command when muck button is clicked', async () => {
    const muckState: ShortDeckResponse = {
      ...showdownState,
      muckAvailable: true,
    };
    mockExec.mockResolvedValue(muckState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'マック' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(showdownState);
    fireEvent.click(screen.getByRole('button', { name: 'マック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('muck'));
  });

  it('calls show command when show button is clicked', async () => {
    const muckState: ShortDeckResponse = {
      ...showdownState,
      muckAvailable: true,
    };
    mockExec.mockResolvedValue(muckState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ショー' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(showdownState);
    fireEvent.click(screen.getByRole('button', { name: 'ショー' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('show'));
  });

  it('does not show muck controls when muckAvailable is false', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText('ショーダウン')).toBeInTheDocument());
    expect(screen.queryByTestId('muck-controls')).not.toBeInTheDocument();
  });

  // ---- ConfirmDialog on reset ----
  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuMetaAI: false, tournamentMode: false }),
    );
  });

  it('sends cpuMetaAI true when checkbox is checked before reset', async () => {
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const checkbox = screen.getByLabelText('メタAI（CPUがプレイスタイルを学習）');
    fireEvent.click(checkbox);

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuMetaAI: true, tournamentMode: false }),
    );
  });

  // ---- Keyboard navigation ----
  describe('keyboard navigation', () => {
    it('pressing c triggers call when canAct and hasOutstandingBet', async () => {
      mockExec.mockResolvedValue(preFlopWithBetState);
      renderWithProviders(<ShortDeckPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

      mockExec.mockClear();
      mockExec.mockResolvedValue(preFlopWithBetState);

      await act(async () => {
        fireEvent.keyDown(document, { key: 'c' });
      });
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('call', undefined, undefined, 0));
    });

    it('pressing k triggers check when canAct and !hasOutstandingBet', async () => {
      mockExec.mockResolvedValue(preFlopState);
      renderWithProviders(<ShortDeckPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument());

      mockExec.mockClear();
      mockExec.mockResolvedValue(preFlopState);

      await act(async () => {
        fireEvent.keyDown(document, { key: 'k' });
      });
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('check', undefined, undefined, 0));
    });

    it('pressing f triggers fold when canAct', async () => {
      mockExec.mockResolvedValue(preFlopState);
      renderWithProviders(<ShortDeckPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

      mockExec.mockClear();
      mockExec.mockResolvedValue(preFlopState);

      await act(async () => {
        fireEvent.keyDown(document, { key: 'f' });
      });
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold', undefined, undefined, 0));
    });

    it('pressing a triggers allin when canAct', async () => {
      mockExec.mockResolvedValue(preFlopState);
      renderWithProviders(<ShortDeckPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument());

      mockExec.mockClear();
      mockExec.mockResolvedValue(preFlopState);

      await act(async () => {
        fireEvent.keyDown(document, { key: 'a' });
      });
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('allin', undefined, undefined, 0));
    });

    it('keyboard is disabled when not canAct', async () => {
      mockExec.mockResolvedValue(showdownState);
      renderWithProviders(<ShortDeckPage />);
      await waitFor(() => expect(screen.getByText('ショーダウン')).toBeInTheDocument());

      mockExec.mockClear();

      await act(async () => {
        fireEvent.keyDown(document, { key: 'c' });
        fireEvent.keyDown(document, { key: 'k' });
        fireEvent.keyDown(document, { key: 'f' });
        fireEvent.keyDown(document, { key: 'a' });
      });
      await flushPendingDispatch();
      expect(mockExec).not.toHaveBeenCalled();
    });
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue({
      gameEndFlag: true,
      phase: 4,
      currentTurn: 0,
      players: [],
      playerIdx: 0,
      communityCards: [],
    } as unknown as ShortDeckResponse);

    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.shortdeck).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));

    await waitFor(() => expect(actionLogApi.shortdeck).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();

    fireEvent.click(screen.getByText('閉じる'));
    await waitFor(() => expect(screen.queryByText('棋譜')).not.toBeInTheDocument());
    expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
  });

  // ---- Learning mode ----
  describe('learning mode', () => {
    const stateWithEquity: ShortDeckResponse = {
      ...preFlopState,
      equity: {
        winProbability: 0.75,
        handOdds: [
          { handRank: 0, handName: 'High Card', probability: 0.1 },
          { handRank: 1, handName: 'One Pair', probability: 0.9 },
        ],
      },
      potOdds: 33.3,
    };

    it('shows learning mode toggle', async () => {
      mockExec.mockResolvedValue(preFlopState);
      renderWithProviders(<ShortDeckPage />);
      await waitFor(() => expect(screen.getByTestId('learning-mode-toggle')).toBeInTheDocument());
    });

    it('does not show equity display when learning mode is off', async () => {
      mockExec.mockResolvedValue(stateWithEquity);
      renderWithProviders(<ShortDeckPage />);
      await waitFor(() => expect(screen.getByTestId('learning-mode-toggle')).toBeInTheDocument());
      expect(screen.queryByTestId('equity-display')).not.toBeInTheDocument();
    });

    it('shows equity display when learning mode is on and equity data exists', async () => {
      mockExec.mockResolvedValue(stateWithEquity);
      renderWithProviders(<ShortDeckPage />);
      await waitFor(() => expect(screen.getByTestId('learning-mode-toggle')).toBeInTheDocument());

      fireEvent.click(screen.getByLabelText('ラーニングモード'));
      expect(screen.getByTestId('equity-display')).toBeInTheDocument();
    });

    it('hides equity display when learning mode is toggled off', async () => {
      mockExec.mockResolvedValue(stateWithEquity);
      renderWithProviders(<ShortDeckPage />);
      await waitFor(() => expect(screen.getByTestId('learning-mode-toggle')).toBeInTheDocument());

      fireEvent.click(screen.getByLabelText('ラーニングモード'));
      expect(screen.getByTestId('equity-display')).toBeInTheDocument();

      fireEvent.click(screen.getByLabelText('ラーニングモード'));
      expect(screen.queryByTestId('equity-display')).not.toBeInTheDocument();
    });

    it('does not show equity display when equity is not in state', async () => {
      mockExec.mockResolvedValue(preFlopState);
      renderWithProviders(<ShortDeckPage />);
      await waitFor(() => expect(screen.getByTestId('learning-mode-toggle')).toBeInTheDocument());

      fireEvent.click(screen.getByLabelText('ラーニングモード'));
      expect(screen.queryByTestId('equity-display')).not.toBeInTheDocument();
    });
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  // ---- HUD stats ----
  it('shows HUD stats for CPU when totalHands > 0', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [
        humanPlayer(),
        cpuPlayer(1, { totalHands: 5, vpip: 60, pfr: 20, threeBet: 10, af: '2.5' }),
        cpuPlayer(2),
      ],
    });
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => {
      const statsElem = screen.getByTestId('hud-stats');
      expect(statsElem).toHaveTextContent(/VPIP.*:60%.*PFR.*:20%.*3Bet.*:10%.*AF.*:2\.5/);
    });
  });

  it('does not show HUD stats for CPU when totalHands is 0', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByTestId('hud-stats')).not.toBeInTheDocument();
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(initState);
    renderWithProviders(<ShortDeckPage />);
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
      mockExec.mockResolvedValue(preFlopState);
      renderWithProviders(<ShortDeckPage />);
      await waitFor(() => expect(screen.getByTestId('cpu-accordion')).toBeInTheDocument());
    });

    it('renders sticky community cards on mobile', async () => {
      mockExec.mockResolvedValue(preFlopState);
      const { container } = renderWithProviders(<ShortDeckPage />);
      await waitFor(() => expect(screen.getByTestId('cpu-accordion')).toBeInTheDocument());
      const stickyDiv = container.querySelector('[data-tutorial="sd-community-cards"]');
      expect(stickyDiv).toHaveClass('sticky');
    });
  });

  it('wraps settings in collapsible details element', async () => {
    mockExec.mockResolvedValue(preFlopState);
    const { container } = renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
    const allSummaries = container.querySelectorAll('details summary');
    const settingsSummary = Array.from(allSummaries).find((s) => s.textContent?.includes('設定'));
    expect(settingsSummary).toBeTruthy();
  });

  it('renders the Flush > Full House rule reminder chip near the community cards', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<ShortDeckPage />);
    const chip = await screen.findByTestId('shortdeck-rank-watermark');
    expect(chip).toBeInTheDocument();
    // Visible text now comes from the i18n key (ja locale in tests); suit
    // symbols stay locale-independent.
    expect(chip.textContent).toContain('フラッシュ');
    expect(chip.textContent).toContain('フルハウス');
    expect(chip.textContent).toContain('♣♠♥♦');
    // Uses the design-system warning state-badge tokens, not opacity-modified ones.
    expect(chip.className).toContain('border-ds-warning');
    expect(chip.className).toContain('text-ds-warning');
    expect(chip.className).not.toContain('bg-ds-warning/');
    expect(chip.className).not.toContain('border-ds-warning/');
  });

  it('highlights the five cards that made the winning hand', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(document.querySelectorAll('[data-best5-board]').length).toBeGreaterThan(0));
    const marked =
      document.querySelectorAll('[data-best5-board]').length + document.querySelectorAll('[data-best5-hole]').length;
    expect(marked).toBe(5);
  });

  it('highlights the Short Deck wheel, which the standard evaluator cannot see', async () => {
    localStorage.clear();
    mockExec.mockReset();
    // A♠ 7♥ against 6♥ 8♦ 9♣ K♠ Q♦: A-6-7-8-9 is a straight here but nothing in
    // standard poker, so a Hold'em evaluator would light up A K Q 9 8 instead.
    mockExec.mockResolvedValue({
      ...showdownState,
      players: [
        humanPlayer({
          handName: 'ストレート',
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'HEART', value: 7 },
          ],
        }),
        ...showdownState.players.slice(1),
      ],
      communityCards: [
        { design: 'HEART', value: 6 },
        { design: 'DIAMOND', value: 8 },
        { design: 'CLOVER', value: 9 },
        { design: 'SPADE', value: 13 },
        { design: 'DIAMOND', value: 12 },
      ],
    });
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(document.querySelectorAll('[data-best5-hole]').length).toBe(2));
    // Both hole cards belong to the wheel; the K and Q on the board do not.
    expect(document.querySelectorAll('[data-best5-board]').length).toBe(3);
  });

  // The page already rendered hand/level progress behind state.tournamentMode,
  // but nothing could switch it on: the display existed for a state the player
  // could not reach.
  it('sends tournamentMode when the setting is switched on before a reset', async () => {
    mockExec.mockResolvedValue(initState);
    renderWithProviders(<ShortDeckPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    fireEvent.click(screen.getByTestId('sd-tournament-toggle'));
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    // The toggle is only worth having if the value reaches the reset call.
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, expect.objectContaining({ tournamentMode: true })),
    );
  });
});
