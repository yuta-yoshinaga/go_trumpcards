import { act, fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, dramahaApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { DramahaResponse } from '../types/card';
import { DramahaPage } from './DramahaPage';

vi.mock('../api/gameApi', () => ({
  dramahaApi: { exec: vi.fn() },
  actionLogApi: { dramaha: vi.fn() },
}));

const mockUseIsLargeDesktop = vi.fn<() => boolean>().mockReturnValue(false);
vi.mock('../hooks/useCardDimensions', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../hooks/useCardDimensions')>();
  return {
    ...actual,
    useIsLargeDesktop: () => mockUseIsLargeDesktop(),
  };
});

const mockExec = vi.mocked(dramahaApi.exec);

/** Helper: base human player */
const humanPlayer = (overrides: Partial<import('../types/card').DramahaPlayerData> = {}) => ({
  id: 0,
  isHuman: true,
  // Five hole cards -- Dramaha deals five, not Omaha's four.
  // Omaha half (with the flop below): tens and fives = two pair.
  // Draw half (these five as dealt): A K 10 5 3 unpaired = high card.
  cards: [
    { design: 'SPADE' as const, value: 1 },
    { design: 'HEART' as const, value: 13 },
    { design: 'DIAMOND' as const, value: 10 },
    { design: 'CLOVER' as const, value: 5 },
    { design: 'SPADE' as const, value: 3 },
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
const cpuPlayer = (id: number, overrides: Partial<import('../types/card').DramahaPlayerData> = {}) => ({
  id,
  isHuman: false,
  cards: [
    { design: 'DIAMOND' as const, value: 2 },
    { design: 'CLOVER' as const, value: 7 },
    { design: 'SPADE' as const, value: 9 },
    { design: 'HEART' as const, value: 4 },
    { design: 'HEART' as const, value: 6 },
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
const initState: DramahaResponse = {
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
const preFlopState: DramahaResponse = {
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
const preFlopWithBetState: DramahaResponse = {
  ...preFlopState,
  lastBet: 40,
  cpuActions: [{ playerIdx: 1, action: 3, amount: 40 }],
};

/** FLOP (phase 2) with community cards */
const flopState: DramahaResponse = {
  ...preFlopState,
  phase: 2,
  communityCards: [
    { design: 'SPADE', value: 10 },
    { design: 'HEART', value: 5 },
    { design: 'DIAMOND', value: 8 },
  ],
};

/** SHOWDOWN (phase 5) */
const showdownState: DramahaResponse = {
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
        { design: 'CLOVER', value: 6 },
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
const endState: DramahaResponse = {
  ...showdownState,
  phase: 6,
  message: 'Game over.',
};

beforeEach(() => {
  mockExec.mockResolvedValue(initState);
  mockUseIsLargeDesktop.mockReturnValue(false);
});

describe('DramahaPage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<DramahaPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  // ---- mount & reset ----
  it('calls reset on mount', async () => {
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // ---- phase name display ----
  it('shows "初期化中" when phase is INIT (not in PHASE_NAMES)', async () => {
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('初期化中')).toBeInTheDocument());
  });

  it('shows known phase name for PRE_FLOP', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('プリフロップ')).toBeInTheDocument());
  });

  // ---- info bar ----
  it('shows pot and the dealer name via playerName (CPU dealer)', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/ポット:/)).toBeInTheDocument());
    expect(screen.getByText(/ディーラー:/)).toBeInTheDocument();
    // Dealer renders via playerName (CPU 3), not the raw index.
    expect(screen.getAllByText('CPU 3').length).toBeGreaterThan(0);
    expect(screen.queryByText(/Player 3|プレイヤー 3/)).not.toBeInTheDocument();
  });

  // ---- community cards ----
  it('shows 5 CardBack placeholders when communityCards is empty', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('コミュニティカード')).toBeInTheDocument());
    const cardBacks = screen.getAllByAltText('カード裏面');
    // 5 community placeholders + 5 face-down cards for each of the 3 CPUs.
    expect(cardBacks.length).toBeGreaterThanOrEqual(5);
  });

  it('shows CardImage when communityCards has cards', async () => {
    mockExec.mockResolvedValue(flopState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 10')).toBeInTheDocument());
    expect(screen.getByAltText('♥ 5')).toBeInTheDocument();
    expect(screen.getByAltText('♦ 8')).toBeInTheDocument();
  });

  // ---- Both halves of the split ----
  it('shows both halves of the split, which read differently on the same five cards', async () => {
    // Hole A♠ K♥ T♦ 5♣ 3♠ + board T♠ 5♥ 8♦.
    //   Omaha half: T♦ 5♣ + T♠ 5♥ 8♦ = two pair.
    //   Draw half:  A K T 5 3 as dealt = high card.
    mockExec.mockResolvedValue(flopState);
    renderWithProviders(<DramahaPage />);
    await screen.findByTestId('dramaha-hands');
    expect(screen.getByTestId('dramaha-omaha-hand-name')).toHaveTextContent('ツーペア');
    expect(screen.getByTestId('dramaha-draw-hand-name')).toHaveTextContent('ハイカード');
  });

  it('leaves the draw half alone when the board changes the Omaha half', async () => {
    // A third ten on the turn -> the Omaha half becomes a full house. The draw
    // half is the five hole cards and does not read the board at all.
    mockExec.mockResolvedValue({
      ...flopState,
      phase: 3,
      communityCards: [
        { design: 'SPADE', value: 10 },
        { design: 'HEART', value: 5 },
        { design: 'DIAMOND', value: 8 },
        { design: 'CLOVER', value: 10 },
      ],
    });
    renderWithProviders(<DramahaPage />);
    await screen.findByTestId('dramaha-hands');
    expect(screen.getByTestId('dramaha-omaha-hand-name')).toHaveTextContent('フルハウス');
    expect(screen.getByTestId('dramaha-draw-hand-name')).toHaveTextContent('ハイカード');
  });

  it('shows the draw half pre-flop, when there is no Omaha half yet', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    await screen.findByTestId('dramaha-hands');
    expect(screen.getByTestId('dramaha-draw-hand-name')).toHaveTextContent('ハイカード');
    expect(screen.getByTestId('dramaha-omaha-hand-name')).toHaveTextContent('まだ判定できません');
  });

  it('names the draw half from a paired holding on an unhelpful board', async () => {
    mockExec.mockResolvedValue({
      ...flopState,
      players: [
        humanPlayer({
          cards: [
            { design: 'SPADE', value: 7 },
            { design: 'HEART', value: 7 },
            { design: 'DIAMOND', value: 12 },
            { design: 'CLOVER', value: 4 },
            { design: 'SPADE', value: 6 },
          ],
        }),
        cpuPlayer(1),
        cpuPlayer(2),
      ],
    });
    renderWithProviders(<DramahaPage />);
    await screen.findByTestId('dramaha-hands');
    expect(screen.getByTestId('dramaha-draw-hand-name')).toHaveTextContent('ワンペア');
  });

  it('shows neither half once the human has folded', async () => {
    mockExec.mockResolvedValue({
      ...flopState,
      players: [humanPlayer({ folded: true }), cpuPlayer(1), cpuPlayer(2), cpuPlayer(3)],
    });
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 10')).toBeInTheDocument());
    expect(screen.getByTestId('dramaha-omaha-hand-name')).toHaveTextContent('まだ判定できません');
    expect(screen.getByTestId('dramaha-draw-hand-name')).toHaveTextContent('まだ判定できません');
  });

  it('states that the pot always splits', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    const badge = await screen.findByTestId('dramaha-split-rule-badge');
    expect(badge).toHaveTextContent('ポットは常に二分');
  });

  // ---- CPU players ----
  it('renders CPU player info with playStyleName and chips', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.getByText(/CPU 2/)).toBeInTheDocument();
    expect(screen.getAllByText(/タイト/).length).toBeGreaterThanOrEqual(1);
  });

  it('shows CPU bet when currentBet > 0', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer(), cpuPlayer(1, { currentBet: 50 }), cpuPlayer(2)],
    });
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/ベット: 50/)).toBeInTheDocument());
  });

  it('does not show CPU bet when currentBet is 0', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByText(/ベット: 0/)).not.toBeInTheDocument();
  });

  it('shows fold badge for folded CPU', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer(), cpuPlayer(1, { folded: true }), cpuPlayer(2)],
    });
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('[フォールド]')).toBeInTheDocument());
  });

  it('shows all-in badge for all-in CPU', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer(), cpuPlayer(1, { allIn: true }), cpuPlayer(2)],
    });
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('[オールイン]')).toBeInTheDocument());
  });

  it('shows CPU hand name badge during showdown when not folded', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<DramahaPage />);
    // Scoped to the CPU area: the human's own Omaha-half badge reads the same hand name.
    await waitFor(() => expect(within(screen.getByTestId('cpu-accordion')).getByText('ツーペア')).toBeInTheDocument());
  });

  it('highlights exactly 2 hole + 3 board cards as the best-5 at showdown', async () => {
    mockExec.mockResolvedValue(showdownState);
    const { container } = renderWithProviders(<DramahaPage />);
    await screen.findByTestId('dramaha-hands');
    // Dramaha must-use-2 rule → exactly 2 hole and 3 board cards highlighted.
    expect(container.querySelectorAll('[data-best5-hole]')).toHaveLength(2);
    expect(container.querySelectorAll('[data-best5-board]')).toHaveLength(3);
  });

  it('does not show CPU hand name badge when CPU is folded in showdown', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<DramahaPage />);
    const cpus = await waitFor(() => screen.getByTestId('cpu-accordion'));
    // CPU 1 shows a hand name; folded CPU 2 shows none, so exactly one badge.
    expect(within(cpus).getAllByText(/ツーペア|ワンペア/)).toHaveLength(1);
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
    renderWithProviders(<DramahaPage />);
    const cpus = await waitFor(() => screen.getByTestId('cpu-accordion'));
    // CPU 1's handName is empty and CPU 2 is folded -> no badge in the CPU area.
    expect(within(cpus).queryByText(/ツーペア|ワンペア/)).not.toBeInTheDocument();
  });

  it('shows CPU cards face-up during showdown when not folded', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 5')).toBeInTheDocument());
    expect(screen.getByAltText('♥ 8')).toBeInTheDocument();
    expect(screen.getByAltText('♦ 3')).toBeInTheDocument();
    expect(screen.getByAltText('♣ J')).toBeInTheDocument();
  });

  it('shows CardBack for CPU cards when not in showdown', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    const cardBacks = screen.getAllByAltText('カード裏面');
    // 5 community placeholders + 5 face-down cards for each of the 3 CPUs.
    expect(cardBacks.length).toBeGreaterThanOrEqual(5);
  });

  it('shows CardBack for folded CPU in showdown (isShowdown && !p.folded is false)', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<DramahaPage />);
    // CPU 2 is folded -> shows CardBack
    await screen.findByTestId('dramaha-hands');
    // CardBacks exist for CPU 2 (5 backs)
    const cardBacks = screen.getAllByAltText('カード裏面');
    expect(cardBacks.length).toBeGreaterThanOrEqual(5);
  });

  // ---- CPU actions log ----
  it('shows CPU actions log when cpuActions is non-empty', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('CPU行動:')).toBeInTheDocument());
    expect(screen.getByText(/Player 1: ベット/)).toBeInTheDocument();
  });

  it('does not show CPU actions log when cpuActions is empty', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByText('CPU行動:')).not.toBeInTheDocument();
  });

  it('shows amount in CPU action when amount > 0', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/Player 1: ベット \(40\)/)).toBeInTheDocument());
  });

  it('does not show amount when action amount is 0', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      cpuActions: [{ playerIdx: 1, action: 1, amount: 0 }],
    });
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('CPU行動:')).toBeInTheDocument());
    expect(screen.getByText(/Player 1: チェック/)).toBeInTheDocument();
    expect(screen.queryByText(/\(0\)/)).not.toBeInTheDocument();
  });

  it('shows "不明" for unknown action type', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      cpuActions: [{ playerIdx: 1, action: 99, amount: 0 }],
    });
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/Player 1: 不明/)).toBeInTheDocument());
  });

  // ---- round results ----
  it('shows round results in showdown', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
  });

  it('does not show round results when not in showdown', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByText('結果:')).not.toBeInTheDocument();
  });

  it('does not show round results when roundResults is empty in showdown', async () => {
    mockExec.mockResolvedValue({
      ...showdownState,
      roundResults: [],
    });
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('ショーダウン')).toBeInTheDocument());
    expect(screen.queryByText('結果:')).not.toBeInTheDocument();
  });

  it('shows "あなた" for human player in results', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
    expect(within(screen.getByTestId('round-results-visible')).getByText(/あなた: ワンペア/)).toBeInTheDocument();
  });

  it('shows "CPU X" for non-human player in results', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
    expect(within(screen.getByTestId('round-results-visible')).getByText(/CPU 1: ツーペア/)).toBeInTheDocument();
  });

  it('shows hand name in results when present', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<DramahaPage />);
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
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
  });

  it('shows won chips when wonAmount > 0', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() =>
      expect(within(screen.getByTestId('round-results-visible')).getByText(/\+200チップ/)).toBeInTheDocument(),
    );
  });

  it('does not show won chips when wonAmount is 0', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
    expect(screen.queryByText(/\+0チップ/)).not.toBeInTheDocument();
  });

  // ---- human player section ----
  it('shows human player section when humanPlayer exists', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
  });

  it('does not show human player section when no players', async () => {
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('初期化中')).toBeInTheDocument());
    expect(screen.queryByText(/あなたの手札/)).not.toBeInTheDocument();
  });

  it('shows human cards when cards exist', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.getByAltText('♥ K')).toBeInTheDocument();
    expect(screen.getByAltText('♦ 10')).toBeInTheDocument();
    expect(screen.getByAltText('♣ 5')).toBeInTheDocument();
    expect(screen.getByAltText('♠ 3')).toBeInTheDocument();
  });

  it('shows CardBack for human when cards is empty and not folded', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ cards: [] }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    // Human has no cards and has not folded -> 5 face-down cards for the human too.
    const cardBacks = screen.getAllByAltText('カード裏面');
    expect(cardBacks.length).toBeGreaterThanOrEqual(5);
  });

  it('does not show CardBack for human when folded and cards empty', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ cards: [], folded: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    // Folded human with no cards -> !humanPlayer.folded is false -> no CardBacks from human
    // CardBacks come from community (5) + CPU (2*5=10) = 15
    const cardBacks = screen.getAllByAltText('カード裏面');
    expect(cardBacks).toHaveLength(15);
  });

  it('shows human bet when currentBet > 0', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ currentBet: 30 }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/ベット: 30/)).toBeInTheDocument());
  });

  it('shows human fold badge', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ folded: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('[フォールド]')).toBeInTheDocument());
  });

  it('shows human all-in badge', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ allIn: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('[オールイン]')).toBeInTheDocument());
  });

  it('names both of the human halves at showdown', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<DramahaPage />);
    await screen.findByTestId('dramaha-hands');
    // The board pairs the tens and fives; the five hole cards on their own do not pair.
    expect(screen.getByTestId('dramaha-omaha-hand-name')).toHaveTextContent('ツーペア');
    expect(screen.getByTestId('dramaha-draw-hand-name')).toHaveTextContent('ハイカード');
  });

  // ---- message ----
  it('shows game message', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('あなたの番です')).toBeInTheDocument());
  });

  // ---- canAct / betting controls ----
  it('shows bet/check buttons when canAct and no outstanding bet', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument();
  });

  it('shows call/raise buttons when canAct and has outstanding bet', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'レイズ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument();
  });

  it('hides betting controls when not active phase', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('ショーダウン')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'チェック' })).not.toBeInTheDocument();
  });

  it('hides betting controls when human is folded', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ folded: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  it('hides betting controls when human is all-in', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ allIn: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  it('hides betting controls when it is not human turn', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      currentTurn: 1,
    });
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  // ---- bet amount input ----
  it('updates bet amount when changing input', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    const betInput = await screen.findByLabelText('ベット額:');
    await waitFor(() => expect((betInput as HTMLInputElement).value).toBe('20'));

    fireEvent.change(betInput, { target: { value: '50' } });
    expect((betInput as HTMLInputElement).value).toBe('50');
  });

  // ---- button click handlers ----
  it('calls bet command with betAmount', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 20, undefined, 0));
  });

  it('calls check command', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'チェック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('check', undefined, undefined, 0));
  });

  it('calls fold command', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold', undefined, undefined, 0));
  });

  it('calls allin command', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'オールイン' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('allin', undefined, undefined, 0));
  });

  it('calls call command when has outstanding bet', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('call', undefined, undefined, 0));
  });

  it('calls raise command with betAmount when has outstanding bet', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'レイズ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'レイズ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('raise', 20, undefined, 0));
  });

  it('calls reset command when reset button is clicked', async () => {
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuMetaAI: false }));
  });

  it('uses outline style for reset button', async () => {
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const resetBtn = screen.getByRole('button', { name: 'リセット' });
    expect(resetBtn.className).toContain('bg-transparent');
    expect(resetBtn.className).toContain('border');
  });

  // ---- loading / disabled state ----
  it('disables buttons while loading', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());

    let resolve!: (value: DramahaResponse) => void;
    const slowPromise = new Promise<DramahaResponse>((res) => {
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
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('clears error on successful call after failure', async () => {
    renderWithProviders(<DramahaPage />);
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
    renderWithProviders(<DramahaPage />);
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
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('ショーダウン')).toBeInTheDocument());
    // CPU 1 not folded but cards empty -> falls to CardBack branch
    const cardBacks = screen.getAllByAltText('カード裏面');
    expect(cardBacks.length).toBeGreaterThanOrEqual(5);
  });

  // ---- bet amount used by raise ----
  it('sends updated bet amount when raise is clicked after changing input', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    renderWithProviders(<DramahaPage />);
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
    renderWithProviders(<DramahaPage />);
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
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());

    const container = screen.getByRole('button', { name: 'ベット' }).closest('[aria-busy]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');

    let resolve!: (value: DramahaResponse) => void;
    const slowPromise = new Promise<DramahaResponse>((res) => {
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
    renderWithProviders(<DramahaPage />);
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
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByTestId('hud-stats')).not.toBeInTheDocument();
  });

  it('shows HUD stats for human when totalHands > 0', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ totalHands: 3, vpip: 33, pfr: 0, threeBet: 0, af: '-' }), cpuPlayer(1), cpuPlayer(2)],
    });
    renderWithProviders(<DramahaPage />);
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
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByTestId('hud-stats')).not.toBeInTheDocument();
  });

  // ---- SB/BB info bar ----
  it('shows SB/BB in info bar', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
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
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/ハンド#/)).toBeInTheDocument());
    expect(screen.getByText(/レベルアップ:5ハンド毎/)).toBeInTheDocument();
    expect(screen.getByText(/20\/40/)).toBeInTheDocument();
  });

  it('does not show tournament info when tournamentMode is false', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/SB\/BB:/)).toBeInTheDocument());
    expect(screen.queryByText(/ハンド#/)).not.toBeInTheDocument();
    expect(screen.queryByText(/レベルアップ:/)).not.toBeInTheDocument();
  });

  // ---- rebuy phase ----
  it('shows rebuy prompt and buttons in rebuy phase', async () => {
    const rebuyState: DramahaResponse = {
      ...preFlopState,
      phase: 7,
      rebuyPhaseType: 1,
      rebuyChips: 1000,
      rebuyMaxCount: 3,
      rebuyCounts: [0, 0, 0, 0],
      addonChips: 1500,
    };
    mockExec.mockResolvedValue(rebuyState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/リバイしますか/)).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'リバイ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument();
  });

  it('calls rebuy command when rebuy accept button is clicked', async () => {
    const rebuyState: DramahaResponse = {
      ...preFlopState,
      phase: 7,
      rebuyPhaseType: 1,
      rebuyChips: 1000,
      rebuyMaxCount: 3,
      rebuyCounts: [0, 0, 0, 0],
    };
    mockExec.mockResolvedValue(rebuyState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リバイ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'リバイ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('rebuy'));
  });

  it('calls skiprebuy command when rebuy skip button is clicked', async () => {
    const rebuyState: DramahaResponse = {
      ...preFlopState,
      phase: 7,
      rebuyPhaseType: 1,
      rebuyChips: 1000,
      rebuyMaxCount: 3,
      rebuyCounts: [0, 0, 0, 0],
    };
    mockExec.mockResolvedValue(rebuyState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByTestId('rebuy-controls')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    const rebuyControls = screen.getByTestId('rebuy-controls');
    fireEvent.click(within(rebuyControls).getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skiprebuy'));
  });

  // ---- addon phase ----
  it('shows addon prompt and buttons in addon phase', async () => {
    const addonState: DramahaResponse = {
      ...preFlopState,
      phase: 7,
      rebuyPhaseType: 2,
      addonChips: 1500,
      rebuyCounts: [0, 0, 0, 0],
    };
    mockExec.mockResolvedValue(addonState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/アドオンしますか/)).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'アドオン' })).toBeInTheDocument();
  });

  it('calls addon command when addon accept button is clicked', async () => {
    const addonState: DramahaResponse = {
      ...preFlopState,
      phase: 7,
      rebuyPhaseType: 2,
      addonChips: 1500,
      rebuyCounts: [0, 0, 0, 0],
    };
    mockExec.mockResolvedValue(addonState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'アドオン' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'アドオン' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('addon'));
  });

  it('calls skipaddon command when addon skip button is clicked', async () => {
    const addonState: DramahaResponse = {
      ...preFlopState,
      phase: 7,
      rebuyPhaseType: 2,
      addonChips: 1500,
      rebuyCounts: [0, 0, 0, 0],
    };
    mockExec.mockResolvedValue(addonState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText(/アドオンしますか/)).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    const addonControls = screen.getByTestId('addon-controls');
    fireEvent.click(within(addonControls).getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skipaddon'));
  });

  it('shows REBUY phase name in info bar', async () => {
    const rebuyState: DramahaResponse = {
      ...preFlopState,
      phase: 7,
      rebuyPhaseType: 1,
      rebuyChips: 1000,
      rebuyMaxCount: 3,
      rebuyCounts: [0, 0, 0, 0],
    };
    mockExec.mockResolvedValue(rebuyState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('リバイ/アドオン')).toBeInTheDocument());
  });

  it('does not show rebuy controls when not in rebuy phase', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('プリフロップ')).toBeInTheDocument());
    expect(screen.queryByText(/リバイしますか/)).not.toBeInTheDocument();
    expect(screen.queryByText(/アドオンしますか/)).not.toBeInTheDocument();
  });

  // ---- muck phase ----
  it('shows muck controls when phase=SHOWDOWN and muckAvailable=true', async () => {
    const muckState: DramahaResponse = {
      ...showdownState,
      muckAvailable: true,
    };
    mockExec.mockResolvedValue(muckState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByTestId('muck-controls')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'マック' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ショー' })).toBeInTheDocument();
  });

  it('calls muck command when muck button is clicked', async () => {
    const muckState: DramahaResponse = {
      ...showdownState,
      muckAvailable: true,
    };
    mockExec.mockResolvedValue(muckState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'マック' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(showdownState);
    fireEvent.click(screen.getByRole('button', { name: 'マック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('muck'));
  });

  it('calls show command when show button is clicked', async () => {
    const muckState: DramahaResponse = {
      ...showdownState,
      muckAvailable: true,
    };
    mockExec.mockResolvedValue(muckState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ショー' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(showdownState);
    fireEvent.click(screen.getByRole('button', { name: 'ショー' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('show'));
  });

  it('does not show muck controls when muckAvailable is false', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('ショーダウン')).toBeInTheDocument());
    expect(screen.queryByTestId('muck-controls')).not.toBeInTheDocument();
  });

  it('does not show muck controls in END phase even if muckAvailable is true', async () => {
    const endMuckState: DramahaResponse = {
      ...endState,
      muckAvailable: true,
    };
    mockExec.mockResolvedValue(endMuckState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('結果')).toBeInTheDocument());
    expect(screen.queryByTestId('muck-controls')).not.toBeInTheDocument();
  });

  // ---- ConfirmDialog on reset ----
  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuMetaAI: false }));
  });

  it('sends cpuMetaAI true when checkbox is checked before reset', async () => {
    renderWithProviders(<DramahaPage />);
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
      renderWithProviders(<DramahaPage />);
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
      renderWithProviders(<DramahaPage />);
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
      renderWithProviders(<DramahaPage />);
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
      renderWithProviders(<DramahaPage />);
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
      renderWithProviders(<DramahaPage />);
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
      renderWithProviders(<DramahaPage />);
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
      renderWithProviders(<DramahaPage />);
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
      renderWithProviders(<DramahaPage />);
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
      renderWithProviders(<DramahaPage />);
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
    } as unknown as DramahaResponse);

    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.dramaha).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));

    await waitFor(() => expect(actionLogApi.dramaha).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();

    fireEvent.click(screen.getByText('閉じる'));
    await waitFor(() => expect(screen.queryByText('棋譜')).not.toBeInTheDocument());
    expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
  });

  // ---- Learning mode ----
  describe('learning mode', () => {
    const stateWithEquity: DramahaResponse = {
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
      renderWithProviders(<DramahaPage />);
      await waitFor(() => expect(screen.getByTestId('learning-mode-toggle')).toBeInTheDocument());
    });

    it('does not show equity display when learning mode is off', async () => {
      mockExec.mockResolvedValue(stateWithEquity);
      renderWithProviders(<DramahaPage />);
      await waitFor(() => expect(screen.getByTestId('learning-mode-toggle')).toBeInTheDocument());
      expect(screen.queryByTestId('equity-display')).not.toBeInTheDocument();
    });

    it('shows equity display when learning mode is on and equity data exists', async () => {
      mockExec.mockResolvedValue(stateWithEquity);
      renderWithProviders(<DramahaPage />);
      await waitFor(() => expect(screen.getByTestId('learning-mode-toggle')).toBeInTheDocument());

      fireEvent.click(screen.getByLabelText('ラーニングモード'));
      expect(screen.getByTestId('equity-display')).toBeInTheDocument();
    });

    it('hides equity display when learning mode is toggled off', async () => {
      mockExec.mockResolvedValue(stateWithEquity);
      renderWithProviders(<DramahaPage />);
      await waitFor(() => expect(screen.getByTestId('learning-mode-toggle')).toBeInTheDocument());

      fireEvent.click(screen.getByLabelText('ラーニングモード'));
      expect(screen.getByTestId('equity-display')).toBeInTheDocument();

      fireEvent.click(screen.getByLabelText('ラーニングモード'));
      expect(screen.queryByTestId('equity-display')).not.toBeInTheDocument();
    });

    it('does not show equity display when equity is not in state', async () => {
      mockExec.mockResolvedValue(preFlopState);
      renderWithProviders(<DramahaPage />);
      await waitFor(() => expect(screen.getByTestId('learning-mode-toggle')).toBeInTheDocument());

      fireEvent.click(screen.getByLabelText('ラーニングモード'));
      expect(screen.queryByTestId('equity-display')).not.toBeInTheDocument();
    });
  });

  // The settings toggles are wrapped in a 44px-tall <label> so the whole row is
  // a tap target (DESIGN.md Interactive Element Minimum Size, issue #4368).
  // Clicking the label's *text* must therefore flip the checkbox -- that is the
  // behaviour the tap target buys, and it was previously untested here.
  describe('settings toggles are driven by their full label row', () => {
    it.each(['ヒント表示', 'メタAI（CPUがプレイスタイルを学習）'])('toggles %s', async (label) => {
      mockExec.mockResolvedValue(preFlopState);
      renderWithProviders(<DramahaPage />);
      const box = await waitFor(() => screen.getByLabelText(label) as HTMLInputElement);
      const before = box.checked;
      fireEvent.click(screen.getByText(label));
      expect((screen.getByLabelText(label) as HTMLInputElement).checked).toBe(!before);
    });
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(initState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('renders poker table layout (no accordion) on large desktop', async () => {
    mockUseIsLargeDesktop.mockReturnValue(true);
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
    expect(screen.queryByTestId('cpu-accordion')).not.toBeInTheDocument();
  });

  it('wraps settings in collapsible details element', async () => {
    mockExec.mockResolvedValue(preFlopState);
    const { container } = renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
    const allSummaries = container.querySelectorAll('details summary');
    const settingsSummary = Array.from(allSummaries).find((s) => s.textContent?.includes('設定'));
    expect(settingsSummary).toBeTruthy();
  });
});

/** DRAW (phase 8): the exchange round between the flop betting and the turn. */
const drawState: DramahaResponse = {
  ...flopState,
  phase: 8,
};

describe('DramahaPage draw round', () => {
  it('names the draw phase, which Hold-em has no phase 8 for', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('ドロー')).toBeInTheDocument());
  });

  it('shows the draw controls and says the draw happens only once', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<DramahaPage />);
    await screen.findByTestId('draw-controls');
    expect(screen.getByTestId('dramaha-draw-once')).toHaveTextContent('1回');
    expect(screen.getByTestId('dramaha-draw-standpat')).toBeInTheDocument();
  });

  it('offers no betting action during the draw round', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<DramahaPage />);
    await screen.findByTestId('draw-controls');
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'チェック' })).not.toBeInTheDocument();
  });

  it('shows no draw controls outside the draw round', async () => {
    mockExec.mockResolvedValue(flopState);
    renderWithProviders(<DramahaPage />);
    await screen.findByTestId('dramaha-hands');
    expect(screen.queryByTestId('draw-controls')).not.toBeInTheDocument();
  });

  it('shows no draw controls for a seat that has folded', async () => {
    mockExec.mockResolvedValue({
      ...drawState,
      players: [humanPlayer({ folded: true }), cpuPlayer(1), cpuPlayer(2), cpuPlayer(3)],
    });
    renderWithProviders(<DramahaPage />);
    await waitFor(() => expect(screen.getByText('ドロー')).toBeInTheDocument());
    expect(screen.queryByTestId('draw-controls')).not.toBeInTheDocument();
  });

  it('marks a clicked hole card as selected and counts it', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<DramahaPage />);
    const card = await screen.findByTestId('dramaha-hole-card-1');
    expect(card).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(card);
    await waitFor(() => expect(screen.getByTestId('dramaha-hole-card-1')).toHaveAttribute('aria-pressed', 'true'));
    expect(screen.getByTestId('dramaha-draw-selected')).toHaveTextContent('1');
  });

  it('unselects a card clicked twice', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<DramahaPage />);
    const card = await screen.findByTestId('dramaha-hole-card-1');

    fireEvent.click(card);
    await waitFor(() => expect(screen.getByTestId('dramaha-hole-card-1')).toHaveAttribute('aria-pressed', 'true'));
    fireEvent.click(screen.getByTestId('dramaha-hole-card-1'));
    await waitFor(() => expect(screen.getByTestId('dramaha-hole-card-1')).toHaveAttribute('aria-pressed', 'false'));
  });

  it('sends the selected positions 0-based, as the endpoint reads them', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<DramahaPage />);
    await screen.findByTestId('draw-controls');

    // The 1st and 3rd cards on screen -> indices 0 and 2.
    fireEvent.click(screen.getByTestId('dramaha-hole-card-0'));
    fireEvent.click(screen.getByTestId('dramaha-hole-card-2'));
    await waitFor(() => expect(screen.getByTestId('dramaha-draw-exchange')).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(drawState);
    fireEvent.click(screen.getByTestId('dramaha-draw-exchange'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw', undefined, { indices: [0, 2] }));
  });

  it('sends an empty list when standing pat', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<DramahaPage />);
    await screen.findByTestId('draw-controls');

    mockExec.mockClear();
    mockExec.mockResolvedValue(drawState);
    fireEvent.click(screen.getByTestId('dramaha-draw-standpat'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw', undefined, { indices: [] }));
  });

  it('stands pat with an empty list even after cards were ticked and unticked', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<DramahaPage />);
    await screen.findByTestId('draw-controls');

    fireEvent.click(screen.getByTestId('dramaha-hole-card-3'));
    await waitFor(() => expect(screen.getByTestId('dramaha-hole-card-3')).toHaveAttribute('aria-pressed', 'true'));
    fireEvent.click(screen.getByTestId('dramaha-hole-card-3'));
    await waitFor(() => expect(screen.getByTestId('dramaha-hole-card-3')).toHaveAttribute('aria-pressed', 'false'));

    mockExec.mockClear();
    mockExec.mockResolvedValue(drawState);
    fireEvent.click(screen.getByTestId('dramaha-draw-standpat'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw', undefined, { indices: [] }));
  });

  it('cannot exchange with nothing selected', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<DramahaPage />);
    await screen.findByTestId('draw-controls');
    expect(screen.getByTestId('dramaha-draw-exchange')).toBeDisabled();
  });

  it('clears the selection once the draw has been sent', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<DramahaPage />);
    await screen.findByTestId('draw-controls');

    fireEvent.click(screen.getByTestId('dramaha-hole-card-0'));
    await waitFor(() => expect(screen.getByTestId('dramaha-draw-selected')).toHaveTextContent('1'));

    fireEvent.click(screen.getByTestId('dramaha-draw-exchange'));
    await waitFor(() => expect(screen.getByTestId('dramaha-draw-selected')).toHaveTextContent('0'));
  });

  it('does not carry a selection out of the draw round', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<DramahaPage />);
    await screen.findByTestId('draw-controls');

    fireEvent.click(screen.getByTestId('dramaha-hole-card-0'));
    await waitFor(() => expect(screen.getByTestId('dramaha-hole-card-0')).toHaveAttribute('aria-pressed', 'true'));

    // The turn arrives: the draw round is over and the tick must not survive.
    mockExec.mockResolvedValue({ ...flopState, phase: 3 });
    fireEvent.click(screen.getByTestId('dramaha-draw-standpat'));
    await waitFor(() => expect(screen.queryByTestId('draw-controls')).not.toBeInTheDocument());

    mockExec.mockResolvedValue(drawState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await screen.findByTestId('draw-controls');
    expect(screen.getByTestId('dramaha-hole-card-0')).toHaveAttribute('aria-pressed', 'false');
  });

  it('cannot tick a hole card outside the draw round', async () => {
    mockExec.mockResolvedValue(flopState);
    renderWithProviders(<DramahaPage />);
    const card = await screen.findByTestId('dramaha-hole-card-0');
    expect(card).toBeDisabled();
    expect(card).not.toHaveAttribute('aria-pressed');
  });
});

describe('DramahaPage showdown split', () => {
  /** Human takes the Omaha half, CPU 1 the draw half. */
  const splitShowdown: DramahaResponse = {
    ...showdownState,
    roundResults: [
      {
        playerIdx: 0,
        handRank: 1,
        handName: 'ワンペア',
        kickers: '',
        bestHand: [],
        wonAmount: 100,
        mucked: false,
        hiWonAmount: 100,
        lowWonAmount: 0,
      },
      {
        playerIdx: 1,
        handRank: 2,
        handName: 'ツーペア',
        kickers: '',
        bestHand: [],
        wonAmount: 100,
        mucked: false,
        hiWonAmount: 0,
        lowWonAmount: 100,
      },
    ],
  };

  it('says which half of the split each seat took', async () => {
    mockExec.mockResolvedValue(splitShowdown);
    renderWithProviders(<DramahaPage />);
    const table = await screen.findByTestId('dramaha-split-results');
    expect(within(table).getByTestId('dramaha-split-result-0')).toHaveTextContent('オマハ側');
    expect(within(table).getByTestId('dramaha-split-result-0')).not.toHaveTextContent('ドロー側');
    expect(within(table).getByTestId('dramaha-split-result-1')).toHaveTextContent('ドロー側');
    expect(within(table).getByTestId('dramaha-split-result-1')).not.toHaveTextContent('オマハ側');
  });

  it('flags a scoop when one seat takes both halves', async () => {
    mockExec.mockResolvedValue({
      ...splitShowdown,
      roundResults: [
        { ...splitShowdown.roundResults[0], wonAmount: 200, hiWonAmount: 100, lowWonAmount: 100 },
        { ...splitShowdown.roundResults[1], wonAmount: 0, hiWonAmount: 0, lowWonAmount: 0 },
      ],
    });
    renderWithProviders(<DramahaPage />);
    const table = await screen.findByTestId('dramaha-split-results');
    expect(within(table).getByTestId('dramaha-scoop')).toBeInTheDocument();
    expect(within(table).getByTestId('dramaha-split-result-0')).toHaveTextContent('オマハ側 + ドロー側');
  });

  it('does not flag a scoop for a seat that took only one half', async () => {
    mockExec.mockResolvedValue(splitShowdown);
    renderWithProviders(<DramahaPage />);
    const table = await screen.findByTestId('dramaha-split-results');
    expect(within(table).queryByTestId('dramaha-scoop')).not.toBeInTheDocument();
  });

  it('says so when a seat took neither half', async () => {
    mockExec.mockResolvedValue({
      ...splitShowdown,
      roundResults: [
        { ...splitShowdown.roundResults[0], wonAmount: 0, hiWonAmount: 0, lowWonAmount: 0 },
        { ...splitShowdown.roundResults[1], wonAmount: 200, hiWonAmount: 100, lowWonAmount: 100 },
      ],
    });
    renderWithProviders(<DramahaPage />);
    const table = await screen.findByTestId('dramaha-split-results');
    expect(within(table).getByTestId('dramaha-split-result-0')).toHaveTextContent('獲得なし');
  });

  it('shows no split table before showdown', async () => {
    mockExec.mockResolvedValue(flopState);
    renderWithProviders(<DramahaPage />);
    await screen.findByTestId('dramaha-hands');
    expect(screen.queryByTestId('dramaha-split-results')).not.toBeInTheDocument();
  });
});
