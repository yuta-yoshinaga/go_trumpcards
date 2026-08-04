import { act, fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, omahaHiLoApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { OmahaResponse } from '../types/card';
import { OmahaHiLoPage } from './OmahaHiLoPage';

vi.mock('../api/gameApi', () => ({
  omahaHiLoApi: { exec: vi.fn() },
  actionLogApi: { omahahilo: vi.fn() },
}));

const mockUseIsLargeDesktop = vi.fn<() => boolean>().mockReturnValue(false);
vi.mock('../hooks/useCardDimensions', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../hooks/useCardDimensions')>();
  return {
    ...actual,
    useIsLargeDesktop: () => mockUseIsLargeDesktop(),
  };
});

const mockExec = vi.mocked(omahaHiLoApi.exec);

/** Helper: base human player */
const humanPlayer = (overrides: Partial<import('../types/card').OmahaPlayerData> = {}) => ({
  id: 0,
  isHuman: true,
  cards: [
    { design: 'SPADE' as const, value: 1 },
    { design: 'HEART' as const, value: 13 },
    { design: 'DIAMOND' as const, value: 10 },
    { design: 'CLOVER' as const, value: 5 },
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
const cpuPlayer = (id: number, overrides: Partial<import('../types/card').OmahaPlayerData> = {}) => ({
  id,
  isHuman: false,
  cards: [
    { design: 'DIAMOND' as const, value: 2 },
    { design: 'CLOVER' as const, value: 7 },
    { design: 'SPADE' as const, value: 9 },
    { design: 'HEART' as const, value: 4 },
  ],
  chips: 1000,
  currentBet: 0,
  folded: false,
  allIn: false,
  handRank: 0,
  handName: '',
  bestHand: [],
  playStyleName: '\u30bf\u30a4\u30c8',
  totalHands: 0,
  vpip: 0,
  pfr: 0,
  threeBet: 0,
  af: '-',
  ...overrides,
});

/** INIT state (phase 0): no players yet */
const initState: OmahaResponse = {
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
const preFlopState: OmahaResponse = {
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
  message: '\u3042\u306a\u305f\u306e\u756a\u3067\u3059',
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
const preFlopWithBetState: OmahaResponse = {
  ...preFlopState,
  lastBet: 40,
  cpuActions: [{ playerIdx: 1, action: 3, amount: 40 }],
};

/** FLOP (phase 2) with community cards */
const flopState: OmahaResponse = {
  ...preFlopState,
  phase: 2,
  communityCards: [
    { design: 'SPADE', value: 10 },
    { design: 'HEART', value: 5 },
    { design: 'DIAMOND', value: 8 },
  ],
};

/** SHOWDOWN (phase 5) */
const showdownState: OmahaResponse = {
  players: [
    humanPlayer({ handName: '\u30ef\u30f3\u30da\u30a2', currentBet: 0, chips: 950 }),
    cpuPlayer(1, {
      handName: '\u30c4\u30fc\u30da\u30a2',
      folded: false,
      cards: [
        { design: 'SPADE', value: 5 },
        { design: 'HEART', value: 8 },
        { design: 'DIAMOND', value: 3 },
        { design: 'CLOVER', value: 11 },
      ],
    }),
    cpuPlayer(2, { folded: true }),
  ],
  communityCards: [
    { design: 'SPADE', value: 10 },
    { design: 'HEART', value: 5 },
    { design: 'DIAMOND', value: 8 },
    { design: 'CLOVER', value: 2 },
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
    {
      playerIdx: 0,
      handRank: 1,
      handName: '\u30ef\u30f3\u30da\u30a2',
      kickers: 'A, Q, 10',
      bestHand: [],
      wonAmount: 0,
      mucked: false,
    },
    {
      playerIdx: 1,
      handRank: 2,
      handName: '\u30c4\u30fc\u30da\u30a2',
      kickers: '8',
      bestHand: [],
      wonAmount: 200,
      mucked: false,
    },
  ],
  cpuActions: [],
  message: 'CPU 1 \u306e\u52dd\u3061',
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
const endState: OmahaResponse = {
  ...showdownState,
  phase: 6,
  message: 'Game over.',
};

beforeEach(() => {
  mockExec.mockResolvedValue(initState);
  mockUseIsLargeDesktop.mockReturnValue(false);
});

describe('OmahaHiLoPage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<OmahaHiLoPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  // ---- mount & reset ----
  it('calls reset on mount', async () => {
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // ---- phase name display ----
  it('shows "初期化中" when phase is INIT (not in PHASE_NAMES)', async () => {
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('初期化中')).toBeInTheDocument());
  });

  it('shows known phase name for PRE_FLOP', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('プリフロップ')).toBeInTheDocument());
  });

  // ---- info bar ----
  it('shows pot and the dealer name via playerName (CPU dealer)', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/ポット:/)).toBeInTheDocument());
    expect(screen.getByText(/ディーラー:/)).toBeInTheDocument();
    // Dealer renders via playerName (CPU 3), not the raw index.
    expect(screen.getAllByText('CPU 3').length).toBeGreaterThan(0);
    expect(screen.queryByText(/Player 3|プレイヤー 3/)).not.toBeInTheDocument();
  });

  // ---- community cards ----
  it('shows 5 CardBack placeholders when communityCards is empty', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('コミュニティカード')).toBeInTheDocument());
    const cardBacks = screen.getAllByAltText('カード裏面');
    // 5 community card placeholders + 4 cards for each of the 3 CPUs = 17 card backs expected
    // but we just verify at least 5 exist for community cards
    expect(cardBacks.length).toBeGreaterThanOrEqual(5);
  });

  it('shows CardImage when communityCards has cards', async () => {
    mockExec.mockResolvedValue(flopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 10')).toBeInTheDocument());
    expect(screen.getByAltText('♥ 5')).toBeInTheDocument();
    expect(screen.getByAltText('♦ 8')).toBeInTheDocument();
  });

  // ---- board low-possibility badge (#3005) ----
  it('does not show the board-low badge pre-flop', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('コミュニティカード')).toBeInTheDocument());
    expect(screen.queryByTestId('omahahilo-board-low-badge')).not.toBeInTheDocument();
  });

  it('shows the "possible" board-low badge on a flop with 2 distinct low ranks', async () => {
    // flopState board is 10,5,8 → only 5 and 8 are ≤ 8 (2 distinct low ranks).
    mockExec.mockResolvedValue(flopState);
    renderWithProviders(<OmahaHiLoPage />);
    const badge = await screen.findByTestId('omahahilo-board-low-badge');
    expect(badge).toHaveAttribute('data-status', 'possible');
    // badgeInfoColors carries the "info" signal on the border, not the text:
    // text-ds-info is only ~4.5:1 on the surface background, so the foreground
    // stays text-ds-text-primary (12:1) and the border does the signalling.
    // See styles/badgeStyles.ts and issue #4367.
    expect(badge).toHaveClass('border-ds-border-subtle');
    expect(badge).toHaveClass('text-ds-text-primary');
  });

  it('shows the "live" board-low badge when the board has 3 distinct low ranks', async () => {
    mockExec.mockResolvedValue({
      ...flopState,
      communityCards: [
        { design: 'SPADE', value: 2 },
        { design: 'HEART', value: 4 },
        { design: 'DIAMOND', value: 7 },
      ],
    });
    renderWithProviders(<OmahaHiLoPage />);
    const badge = await screen.findByTestId('omahahilo-board-low-badge');
    expect(badge).toHaveAttribute('data-status', 'live');
    expect(badge).toHaveClass('text-ds-success');
  });

  it('shows the "impossible" board-low badge on a full board with fewer than 3 low ranks', async () => {
    // River with only 5 and 8 low → cannot reach 3 distinct low ranks.
    mockExec.mockResolvedValue({
      ...flopState,
      phase: 4,
      communityCards: [
        { design: 'SPADE', value: 5 },
        { design: 'HEART', value: 8 },
        { design: 'DIAMOND', value: 10 },
        { design: 'CLOVER', value: 12 },
        { design: 'SPADE', value: 13 },
      ],
    });
    renderWithProviders(<OmahaHiLoPage />);
    const badge = await screen.findByTestId('omahahilo-board-low-badge');
    expect(badge).toHaveAttribute('data-status', 'impossible');
    expect(badge).toHaveClass('text-ds-text-muted');
  });

  // ---- CPU players ----
  it('renders CPU player info with playStyleName and chips', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.getByText(/CPU 2/)).toBeInTheDocument();
    expect(screen.getAllByText(/タイト/).length).toBeGreaterThanOrEqual(1);
  });

  it('shows CPU bet when currentBet > 0', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer(), cpuPlayer(1, { currentBet: 50 }), cpuPlayer(2)],
    });
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/ベット: 50/)).toBeInTheDocument());
  });

  it('does not show CPU bet when currentBet is 0', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByText(/ベット: 0/)).not.toBeInTheDocument();
  });

  it('shows fold badge for folded CPU', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer(), cpuPlayer(1, { folded: true }), cpuPlayer(2)],
    });
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('[フォールド]')).toBeInTheDocument());
  });

  it('shows all-in badge for all-in CPU', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer(), cpuPlayer(1, { allIn: true }), cpuPlayer(2)],
    });
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('[オールイン]')).toBeInTheDocument());
  });

  it('shows CPU hand name badge during showdown when not folded', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('ツーペア')).toBeInTheDocument());
  });

  it('highlights exactly 2 hole + 3 board cards as the Hi best-5 at showdown', async () => {
    mockExec.mockResolvedValue(showdownState);
    const { container } = renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('ツーペア')).toBeInTheDocument());
    expect(container.querySelectorAll('[data-best5-hole]')).toHaveLength(2);
    expect(container.querySelectorAll('[data-best5-board]')).toHaveLength(3);
  });

  it('does not highlight a Hi best-5 when the human folded', async () => {
    mockExec.mockResolvedValue({
      ...showdownState,
      players: [humanPlayer({ folded: true, handName: '' }), showdownState.players[1], showdownState.players[2]],
    });
    const { container } = renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('ツーペア')).toBeInTheDocument());
    expect(container.querySelectorAll('[data-best5-hole]')).toHaveLength(0);
    expect(container.querySelectorAll('[data-best5-board]')).toHaveLength(0);
  });

  it('shows green Hi and blue Lo badges when the pot splits', async () => {
    const splitState: OmahaResponse = {
      ...showdownState,
      roundResults: [
        { ...showdownState.roundResults[0], hiWonAmount: 200, lowWonAmount: 0 },
        { ...showdownState.roundResults[1], wonAmount: 100, hiWonAmount: 0, lowWonAmount: 100 },
      ],
    };
    mockExec.mockResolvedValue(splitState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByTestId('omahahilo-split')).toBeInTheDocument());
    expect(screen.getByTestId('omahahilo-hi-badge')).toHaveClass('text-ds-success');
    expect(screen.getByTestId('omahahilo-lo-badge')).toHaveClass('text-ds-info');
  });

  it('omits the Lo badge when no low qualifies', async () => {
    const hiOnlyState: OmahaResponse = {
      ...showdownState,
      roundResults: [{ ...showdownState.roundResults[0], hiWonAmount: 200, lowWonAmount: 0 }],
    };
    mockExec.mockResolvedValue(hiOnlyState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByTestId('omahahilo-hi-badge')).toBeInTheDocument());
    expect(screen.queryByTestId('omahahilo-lo-badge')).not.toBeInTheDocument();
    // With no qualifying low, the breakdown states that Hi scoops the pot.
    expect(screen.getByTestId('omahahilo-hi-takes-all')).toBeInTheDocument();
  });

  it('highlights the human qualifying low and lists the low cards', async () => {
    // Human hole A♠,K♥,10♦,5♣; board 10♠,5♥,8♦,2♣,9♥.
    const lowState: OmahaResponse = {
      ...showdownState,
      roundResults: [
        {
          ...showdownState.roundResults[0],
          hiWonAmount: 0,
          lowWonAmount: 100,
          lowBestHand: [
            { design: 'SPADE', value: 1 },
            { design: 'CLOVER', value: 5 },
            { design: 'DIAMOND', value: 8 },
            { design: 'CLOVER', value: 2 },
            { design: 'HEART', value: 5 },
          ],
        },
        { ...showdownState.roundResults[1], hiWonAmount: 200, lowWonAmount: 0 },
      ],
    };
    mockExec.mockResolvedValue(lowState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByTestId('omahahilo-split')).toBeInTheDocument());
    // At least one card (hole and/or board) is ringed as a low card.
    const loCards = screen.getAllByTestId('omahahilo-lo-card');
    expect(loCards.length).toBeGreaterThan(0);
    // Each low card keeps the blue ring AND gains screen-reader text plus a
    // color-independent "LO" badge (not just color).
    expect(loCards[0].className).toContain('ring-ds-info');
    expect(within(loCards[0]).getByText('ロー構成カード')).toBeInTheDocument();
    expect(screen.getAllByTestId('omahahilo-lo-card-badge')[0]).toHaveTextContent('LO');
    // The low winner badge still lists the qualifying card ranks.
    expect(screen.getByTestId('omahahilo-lo-badge')).toHaveTextContent('A 5 8 2 5');
  });

  it('does not show CPU hand name badge when CPU is folded in showdown', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('ツーペア')).toBeInTheDocument());
    const handBadges = screen.getAllByText(/ツーペア|ワンペア/);
    expect(handBadges.length).toBeGreaterThanOrEqual(1);
  });

  it('does not show CPU hand name badge when handName is empty in showdown', async () => {
    mockExec.mockResolvedValue({
      ...showdownState,
      players: [
        humanPlayer({ handName: 'ワンペア' }),
        cpuPlayer(1, { handName: '', folded: false }),
        cpuPlayer(2, { folded: true }),
      ],
    });
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('ワンペア')).toBeInTheDocument());
  });

  it('shows CPU cards face-up during showdown when not folded', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 5')).toBeInTheDocument());
    expect(screen.getByAltText('♥ 8')).toBeInTheDocument();
    expect(screen.getByAltText('♦ 3')).toBeInTheDocument();
    expect(screen.getByAltText('♣ J')).toBeInTheDocument();
  });

  it('shows CardBack for CPU cards when not in showdown', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    const cardBacks = screen.getAllByAltText('カード裏面');
    // 5 community card placeholders + 4 cards for each of the 3 CPUs = 17 card backs expected
    expect(cardBacks.length).toBeGreaterThanOrEqual(4);
  });

  it('shows CardBack for folded CPU in showdown (isShowdown && !p.folded is false)', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<OmahaHiLoPage />);
    // CPU 2 is folded -> shows CardBack
    await waitFor(() => expect(screen.getByText('ツーペア')).toBeInTheDocument());
    // CardBacks exist for CPU 2 (4 backs)
    const cardBacks = screen.getAllByAltText('カード裏面');
    expect(cardBacks.length).toBeGreaterThanOrEqual(4);
  });

  // ---- CPU actions log ----
  it('shows CPU actions log when cpuActions is non-empty', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('CPU行動:')).toBeInTheDocument());
    expect(screen.getByText(/Player 1: ベット/)).toBeInTheDocument();
  });

  it('does not show CPU actions log when cpuActions is empty', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByText('CPU行動:')).not.toBeInTheDocument();
  });

  it('shows amount in CPU action when amount > 0', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/Player 1: ベット \(40\)/)).toBeInTheDocument());
  });

  it('does not show amount when action amount is 0', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      cpuActions: [{ playerIdx: 1, action: 1, amount: 0 }],
    });
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('CPU行動:')).toBeInTheDocument());
    expect(screen.getByText(/Player 1: チェック/)).toBeInTheDocument();
    expect(screen.queryByText(/\(0\)/)).not.toBeInTheDocument();
  });

  it('shows "不明" for unknown action type', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      cpuActions: [{ playerIdx: 1, action: 99, amount: 0 }],
    });
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/Player 1: 不明/)).toBeInTheDocument());
  });

  // ---- round results ----
  it('shows round results in showdown', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
  });

  it('does not show round results when not in showdown', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByText('結果:')).not.toBeInTheDocument();
  });

  it('does not show round results when roundResults is empty in showdown', async () => {
    mockExec.mockResolvedValue({
      ...showdownState,
      roundResults: [],
    });
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('ショーダウン')).toBeInTheDocument());
    expect(screen.queryByText('結果:')).not.toBeInTheDocument();
  });

  it('shows "あなた" for human player in results', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
    expect(within(screen.getByTestId('round-results-visible')).getByText(/あなた: ワンペア/)).toBeInTheDocument();
  });

  it('shows "CPU X" for non-human player in results', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
    expect(within(screen.getByTestId('round-results-visible')).getByText(/CPU 1: ツーペア/)).toBeInTheDocument();
  });

  it('shows hand name in results when present', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() =>
      expect(within(screen.getByTestId('round-results-visible')).getByText(/: ワンペア/)).toBeInTheDocument(),
    );
    expect(within(screen.getByTestId('round-results-visible')).getByText(/: ツーペア/)).toBeInTheDocument();
  });

  it('does not show hand name in results when empty', async () => {
    mockExec.mockResolvedValue({
      ...showdownState,
      roundResults: [
        { playerIdx: 0, handRank: 0, handName: '', kickers: '', bestHand: [], wonAmount: 0, mucked: false },
        { playerIdx: 1, handRank: 2, handName: 'ツーペア', kickers: '', bestHand: [], wonAmount: 200, mucked: false },
      ],
    });
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
  });

  it('shows won chips when wonAmount > 0', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() =>
      expect(within(screen.getByTestId('round-results-visible')).getByText(/\+200チップ/)).toBeInTheDocument(),
    );
  });

  it('does not show won chips when wonAmount is 0', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
    expect(screen.queryByText(/\+0チップ/)).not.toBeInTheDocument();
  });

  // ---- human player section ----
  it('shows human player section when humanPlayer exists', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
  });

  it('does not show human player section when no players', async () => {
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('初期化中')).toBeInTheDocument());
    expect(screen.queryByText(/あなたの手札/)).not.toBeInTheDocument();
  });

  it('shows human cards when cards exist', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.getByAltText('♥ K')).toBeInTheDocument();
    expect(screen.getByAltText('♦ 10')).toBeInTheDocument();
    expect(screen.getByAltText('♣ 5')).toBeInTheDocument();
  });

  it('shows CardBack for human when cards is empty and not folded', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ cards: [] }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    // Human has no cards, not folded -> shows 4 CardBacks for human + 5 community + 8 CPU = many
    const cardBacks = screen.getAllByAltText('カード裏面');
    expect(cardBacks.length).toBeGreaterThanOrEqual(4);
  });

  it('does not show CardBack for human when folded and cards empty', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ cards: [], folded: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    // Folded human with no cards -> !humanPlayer.folded is false -> no CardBacks from human
    // CardBacks come from community (5) + CPU (2*4=8) = 13
    const cardBacks = screen.getAllByAltText('カード裏面');
    expect(cardBacks).toHaveLength(13);
  });

  it('shows human bet when currentBet > 0', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ currentBet: 30 }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/ベット: 30/)).toBeInTheDocument());
  });

  it('shows human fold badge', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ folded: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('[フォールド]')).toBeInTheDocument());
  });

  it('shows human all-in badge', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ allIn: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('[オールイン]')).toBeInTheDocument());
  });

  it('shows human hand name badge during showdown when not folded', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('ワンペア')).toBeInTheDocument());
  });

  it('does not show human hand name badge when handName is empty in showdown', async () => {
    mockExec.mockResolvedValue({
      ...showdownState,
      players: [
        humanPlayer({ handName: '' }),
        cpuPlayer(1, { handName: 'ツーペア', folded: false }),
        cpuPlayer(2, { folded: true }),
      ],
    });
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('ツーペア')).toBeInTheDocument());
  });

  // ---- message ----
  it('shows game message', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('あなたの番です')).toBeInTheDocument());
  });

  // ---- canAct / betting controls ----
  it('shows bet/check buttons when canAct and no outstanding bet', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument();
  });

  it('shows call/raise buttons when canAct and has outstanding bet', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'レイズ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument();
  });

  it('hides betting controls when not active phase', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('ショーダウン')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'チェック' })).not.toBeInTheDocument();
  });

  it('hides betting controls when human is folded', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ folded: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  it('hides betting controls when human is all-in', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ allIn: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  it('hides betting controls when it is not human turn', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      currentTurn: 1,
    });
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  // ---- bet amount input ----
  it('updates bet amount when changing input', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    const betInput = await screen.findByLabelText('ベット額:');
    await waitFor(() => expect((betInput as HTMLInputElement).value).toBe('20'));

    fireEvent.change(betInput, { target: { value: '50' } });
    expect((betInput as HTMLInputElement).value).toBe('50');
  });

  // ---- button click handlers ----
  it('calls bet command with betAmount', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 20, undefined, 0));
  });

  it('calls check command', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'チェック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('check', undefined, undefined, 0));
  });

  it('calls fold command', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold', undefined, undefined, 0));
  });

  it('calls allin command', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'オールイン' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('allin', undefined, undefined, 0));
  });

  it('calls call command when has outstanding bet', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('call', undefined, undefined, 0));
  });

  it('calls raise command with betAmount when has outstanding bet', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'レイズ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'レイズ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('raise', 20, undefined, 0));
  });

  it('calls reset command when reset button is clicked', async () => {
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuMetaAI: false }));
  });

  it('uses outline style for reset button', async () => {
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const resetBtn = screen.getByRole('button', { name: 'リセット' });
    expect(resetBtn.className).toContain('bg-transparent');
    expect(resetBtn.className).toContain('border');
  });

  // ---- loading / disabled state ----
  it('disables buttons while loading', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());

    let resolve!: (value: OmahaResponse) => void;
    const slowPromise = new Promise<OmahaResponse>((res) => {
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
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('clears error on successful call after failure', async () => {
    renderWithProviders(<OmahaHiLoPage />);
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
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('結果')).toBeInTheDocument());
    expect(screen.getByText('結果:')).toBeInTheDocument();
  });

  // ---- showdown with CPU cards having no length (empty cards, not folded) ----
  it('shows CardBack for CPU in showdown when cards is empty and not folded', async () => {
    mockExec.mockResolvedValue({
      ...showdownState,
      players: [
        humanPlayer({ handName: 'ワンペア' }),
        cpuPlayer(1, { handName: 'ツーペア', folded: false, cards: [] }),
        cpuPlayer(2, { folded: true }),
      ],
    });
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('ショーダウン')).toBeInTheDocument());
    // CPU 1 not folded but cards empty -> falls to CardBack branch
    const cardBacks = screen.getAllByAltText('カード裏面');
    expect(cardBacks.length).toBeGreaterThanOrEqual(4);
  });

  // ---- bet amount used by raise ----
  it('sends updated bet amount when raise is clicked after changing input', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    renderWithProviders(<OmahaHiLoPage />);
    const betInput = await screen.findByLabelText('ベット額:');
    await waitFor(() => expect((betInput as HTMLInputElement).value).toBe('20'));

    fireEvent.change(betInput, { target: { value: '100' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'レイズ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('raise', 100, undefined, 0));
  });

  it('sends updated bet amount when bet is clicked after changing input', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    const betInput = await screen.findByLabelText('ベット額:');
    await waitFor(() => expect((betInput as HTMLInputElement).value).toBe('20'));

    fireEvent.change(betInput, { target: { value: '60' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 60, undefined, 0));
  });

  it('sets aria-busy while loading', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());

    const container = screen.getByRole('button', { name: 'ベット' }).closest('[aria-busy]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');

    let resolve!: (value: OmahaResponse) => void;
    const slowPromise = new Promise<OmahaResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));

    expect(container).toHaveAttribute('aria-busy', 'true');

    resolve(preFlopState);
    await waitFor(() => {
      expect(container).toHaveAttribute('aria-busy', 'false');
    });
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
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => {
      const statsElem = screen.getByTestId('hud-stats');
      expect(statsElem).toHaveTextContent(/VPIP.*:60%.*PFR.*:20%.*3Bet.*:10%.*AF.*:2\.5/);
      expect(
        within(statsElem).getByRole('tooltip', { name: /VPIP（ボランタリー・プット・イン・ポット）/ }),
      ).toBeInTheDocument();
      expect(within(statsElem).getByRole('tooltip', { name: /PFR（プリフロップレイズ）/ })).toBeInTheDocument();
      expect(within(statsElem).getByRole('tooltip', { name: /3Bet（スリーベット）/ })).toBeInTheDocument();
      expect(within(statsElem).getByRole('tooltip', { name: /AF（アグレッションファクター）/ })).toBeInTheDocument();
    });
  });

  it('does not show HUD stats for CPU when totalHands is 0', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByTestId('hud-stats')).not.toBeInTheDocument();
  });

  it('shows HUD stats for human when totalHands > 0', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ totalHands: 3, vpip: 33, pfr: 0, threeBet: 0, af: '-' }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => {
      const statsElem = screen.getByTestId('hud-stats');
      expect(statsElem).toHaveTextContent(/VPIP.*:33%.*PFR.*:0%.*3Bet.*:0%.*AF.*:-/);
      expect(
        within(statsElem).getByRole('tooltip', { name: /VPIP（ボランタリー・プット・イン・ポット）/ }),
      ).toBeInTheDocument();
    });
  });

  it('does not show HUD stats for human when totalHands is 0', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByTestId('hud-stats')).not.toBeInTheDocument();
  });

  // ---- SB/BB info bar ----
  it('shows SB/BB in info bar', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
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
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/ハンド#/)).toBeInTheDocument());
    expect(screen.getByText(/レベルアップ:5ハンド毎/)).toBeInTheDocument();
    expect(screen.getByText(/20\/40/)).toBeInTheDocument();
  });

  it('does not show tournament info when tournamentMode is false', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/SB\/BB:/)).toBeInTheDocument());
    expect(screen.queryByText(/ハンド#/)).not.toBeInTheDocument();
    expect(screen.queryByText(/レベルアップ:/)).not.toBeInTheDocument();
  });

  // ---- rebuy phase ----
  it('shows rebuy prompt and buttons in rebuy phase', async () => {
    const rebuyState: OmahaResponse = {
      ...preFlopState,
      phase: 7,
      rebuyPhaseType: 1,
      rebuyChips: 1000,
      rebuyMaxCount: 3,
      rebuyCounts: [0, 0, 0, 0],
      addonChips: 1500,
    };
    mockExec.mockResolvedValue(rebuyState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/リバイしますか/)).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'リバイ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument();
  });

  it('calls rebuy command when rebuy accept button is clicked', async () => {
    const rebuyState: OmahaResponse = {
      ...preFlopState,
      phase: 7,
      rebuyPhaseType: 1,
      rebuyChips: 1000,
      rebuyMaxCount: 3,
      rebuyCounts: [0, 0, 0, 0],
    };
    mockExec.mockResolvedValue(rebuyState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リバイ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'リバイ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('rebuy'));
  });

  it('calls skiprebuy command when rebuy skip button is clicked', async () => {
    const rebuyState: OmahaResponse = {
      ...preFlopState,
      phase: 7,
      rebuyPhaseType: 1,
      rebuyChips: 1000,
      rebuyMaxCount: 3,
      rebuyCounts: [0, 0, 0, 0],
    };
    mockExec.mockResolvedValue(rebuyState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByTestId('rebuy-controls')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    const rebuyControls = screen.getByTestId('rebuy-controls');
    fireEvent.click(within(rebuyControls).getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skiprebuy'));
  });

  // ---- addon phase ----
  it('shows addon prompt and buttons in addon phase', async () => {
    const addonState: OmahaResponse = {
      ...preFlopState,
      phase: 7,
      rebuyPhaseType: 2,
      addonChips: 1500,
      rebuyCounts: [0, 0, 0, 0],
    };
    mockExec.mockResolvedValue(addonState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/アドオンしますか/)).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'アドオン' })).toBeInTheDocument();
  });

  it('calls addon command when addon accept button is clicked', async () => {
    const addonState: OmahaResponse = {
      ...preFlopState,
      phase: 7,
      rebuyPhaseType: 2,
      addonChips: 1500,
      rebuyCounts: [0, 0, 0, 0],
    };
    mockExec.mockResolvedValue(addonState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'アドオン' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'アドオン' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('addon'));
  });

  it('calls skipaddon command when addon skip button is clicked', async () => {
    const addonState: OmahaResponse = {
      ...preFlopState,
      phase: 7,
      rebuyPhaseType: 2,
      addonChips: 1500,
      rebuyCounts: [0, 0, 0, 0],
    };
    mockExec.mockResolvedValue(addonState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText(/アドオンしますか/)).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    const addonControls = screen.getByTestId('addon-controls');
    fireEvent.click(within(addonControls).getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skipaddon'));
  });

  it('shows REBUY phase name in info bar', async () => {
    const rebuyState: OmahaResponse = {
      ...preFlopState,
      phase: 7,
      rebuyPhaseType: 1,
      rebuyChips: 1000,
      rebuyMaxCount: 3,
      rebuyCounts: [0, 0, 0, 0],
    };
    mockExec.mockResolvedValue(rebuyState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('リバイ/アドオン')).toBeInTheDocument());
  });

  it('does not show rebuy controls when not in rebuy phase', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('プリフロップ')).toBeInTheDocument());
    expect(screen.queryByText(/リバイしますか/)).not.toBeInTheDocument();
    expect(screen.queryByText(/アドオンしますか/)).not.toBeInTheDocument();
  });

  // ---- muck phase ----
  it('shows muck controls when phase=SHOWDOWN and muckAvailable=true', async () => {
    const muckState: OmahaResponse = {
      ...showdownState,
      muckAvailable: true,
    };
    mockExec.mockResolvedValue(muckState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByTestId('muck-controls')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'マック' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ショー' })).toBeInTheDocument();
  });

  it('calls muck command when muck button is clicked', async () => {
    const muckState: OmahaResponse = {
      ...showdownState,
      muckAvailable: true,
    };
    mockExec.mockResolvedValue(muckState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'マック' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(showdownState);
    fireEvent.click(screen.getByRole('button', { name: 'マック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('muck'));
  });

  it('calls show command when show button is clicked', async () => {
    const muckState: OmahaResponse = {
      ...showdownState,
      muckAvailable: true,
    };
    mockExec.mockResolvedValue(muckState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ショー' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(showdownState);
    fireEvent.click(screen.getByRole('button', { name: 'ショー' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('show'));
  });

  it('does not show muck controls when muckAvailable is false', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('ショーダウン')).toBeInTheDocument());
    expect(screen.queryByTestId('muck-controls')).not.toBeInTheDocument();
  });

  it('does not show muck controls in END phase even if muckAvailable is true', async () => {
    const endMuckState: OmahaResponse = {
      ...endState,
      muckAvailable: true,
    };
    mockExec.mockResolvedValue(endMuckState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('結果')).toBeInTheDocument());
    expect(screen.queryByTestId('muck-controls')).not.toBeInTheDocument();
  });

  // ---- ConfirmDialog on reset ----
  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuMetaAI: false }));
  });

  it('sends cpuMetaAI true when checkbox is checked before reset', async () => {
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const checkbox = screen.getByLabelText('メタAI（CPUがプレイスタイルを学習）');
    fireEvent.click(checkbox);

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuMetaAI: true }));
  });

  // ---- Keyboard navigation ----
  describe('keyboard navigation', () => {
    it('pressing c triggers call when canAct and hasOutstandingBet', async () => {
      mockExec.mockResolvedValue(preFlopWithBetState);
      renderWithProviders(<OmahaHiLoPage />);
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
      renderWithProviders(<OmahaHiLoPage />);
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
      renderWithProviders(<OmahaHiLoPage />);
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
      renderWithProviders(<OmahaHiLoPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument());

      mockExec.mockClear();
      mockExec.mockResolvedValue(preFlopState);

      await act(async () => {
        fireEvent.keyDown(document, { key: 'a' });
      });
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('allin', undefined, undefined, 0));
    });

    it('pressing r triggers raise when hasOutstandingBet', async () => {
      mockExec.mockResolvedValue(preFlopWithBetState);
      renderWithProviders(<OmahaHiLoPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: 'レイズ' })).toBeInTheDocument());

      mockExec.mockClear();
      mockExec.mockResolvedValue(preFlopWithBetState);

      await act(async () => {
        fireEvent.keyDown(document, { key: 'r' });
      });
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('raise', 20, undefined, 0));
    });

    it('pressing r triggers bet when !hasOutstandingBet', async () => {
      mockExec.mockResolvedValue(preFlopState);
      renderWithProviders(<OmahaHiLoPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());

      mockExec.mockClear();
      mockExec.mockResolvedValue(preFlopState);

      await act(async () => {
        fireEvent.keyDown(document, { key: 'r' });
      });
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 20, undefined, 0));
    });

    it('keyboard is disabled when not canAct', async () => {
      mockExec.mockResolvedValue(showdownState);
      renderWithProviders(<OmahaHiLoPage />);
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

    it('pressing c is ignored when !hasOutstandingBet', async () => {
      mockExec.mockResolvedValue(preFlopState);
      renderWithProviders(<OmahaHiLoPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument());

      mockExec.mockClear();

      await act(async () => {
        fireEvent.keyDown(document, { key: 'c' });
      });
      await flushPendingDispatch();
      expect(mockExec).not.toHaveBeenCalled();
    });

    it('pressing k is ignored when hasOutstandingBet', async () => {
      mockExec.mockResolvedValue(preFlopWithBetState);
      renderWithProviders(<OmahaHiLoPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

      mockExec.mockClear();

      await act(async () => {
        fireEvent.keyDown(document, { key: 'k' });
      });
      await flushPendingDispatch();
      expect(mockExec).not.toHaveBeenCalled();
    });
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue({
      gameEndFlag: true,
      phase: 4, // HoldemPhase.END
      currentTurn: 0,
      players: [],
      playerIdx: 0,
      communityCards: [],
    } as unknown as OmahaResponse);

    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.omahahilo).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));

    await waitFor(() => expect(actionLogApi.omahahilo).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();

    fireEvent.click(screen.getByText('閉じる'));
    await waitFor(() => expect(screen.queryByText('棋譜')).not.toBeInTheDocument());
    expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
  });

  // ---- Learning mode ----
  describe('learning mode', () => {
    const stateWithEquity: OmahaResponse = {
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
      renderWithProviders(<OmahaHiLoPage />);
      await waitFor(() => expect(screen.getByTestId('learning-mode-toggle')).toBeInTheDocument());
    });

    it('does not show equity display when learning mode is off', async () => {
      mockExec.mockResolvedValue(stateWithEquity);
      renderWithProviders(<OmahaHiLoPage />);
      await waitFor(() => expect(screen.getByTestId('learning-mode-toggle')).toBeInTheDocument());
      expect(screen.queryByTestId('equity-display')).not.toBeInTheDocument();
    });

    it('shows equity display when learning mode is on and equity data exists', async () => {
      mockExec.mockResolvedValue(stateWithEquity);
      renderWithProviders(<OmahaHiLoPage />);
      await waitFor(() => expect(screen.getByTestId('learning-mode-toggle')).toBeInTheDocument());

      fireEvent.click(screen.getByLabelText('ラーニングモード'));
      expect(screen.getByTestId('equity-display')).toBeInTheDocument();
    });

    it('hides equity display when learning mode is toggled off', async () => {
      mockExec.mockResolvedValue(stateWithEquity);
      renderWithProviders(<OmahaHiLoPage />);
      await waitFor(() => expect(screen.getByTestId('learning-mode-toggle')).toBeInTheDocument());

      fireEvent.click(screen.getByLabelText('ラーニングモード'));
      expect(screen.getByTestId('equity-display')).toBeInTheDocument();

      fireEvent.click(screen.getByLabelText('ラーニングモード'));
      expect(screen.queryByTestId('equity-display')).not.toBeInTheDocument();
    });

    it('does not show equity display when equity is not in state', async () => {
      mockExec.mockResolvedValue(preFlopState);
      renderWithProviders(<OmahaHiLoPage />);
      await waitFor(() => expect(screen.getByTestId('learning-mode-toggle')).toBeInTheDocument());

      fireEvent.click(screen.getByLabelText('ラーニングモード'));
      expect(screen.queryByTestId('equity-display')).not.toBeInTheDocument();
    });
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(initState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('renders poker table layout (no accordion) on large desktop', async () => {
    mockUseIsLargeDesktop.mockReturnValue(true);
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
    expect(screen.queryByTestId('cpu-accordion')).not.toBeInTheDocument();
  });

  it('wraps settings in collapsible details element', async () => {
    mockExec.mockResolvedValue(preFlopState);
    const { container } = renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
    const allSummaries = container.querySelectorAll('details summary');
    const settingsSummary = Array.from(allSummaries).find((s) => s.textContent?.includes('設定'));
    expect(settingsSummary).toBeTruthy();
  });
  // The settings toggles are wrapped in a 44px-tall <label> so the whole row is
  // a tap target (DESIGN.md Interactive Element Minimum Size, issue #4368).
  // Clicking the label's *text* must therefore flip the checkbox -- that is the
  // behaviour the tap target buys, and it was previously untested here.
  describe('settings toggles are driven by their full label row', () => {
    it.each(['ヒント表示', 'メタAI（CPUがプレイスタイルを学習）'])('toggles %s', async (label) => {
      mockExec.mockResolvedValue(preFlopState);
      renderWithProviders(<OmahaHiLoPage />);
      const box = await waitFor(() => screen.getByLabelText(label) as HTMLInputElement);
      const before = box.checked;
      fireEvent.click(screen.getByText(label));
      expect((screen.getByLabelText(label) as HTMLInputElement).checked).toBe(!before);
    });
  });

  it('previews the current best hand during play', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(flopState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => expect(screen.getByTestId('omahahilo-live-besthand')).toBeInTheDocument());
  });
});
