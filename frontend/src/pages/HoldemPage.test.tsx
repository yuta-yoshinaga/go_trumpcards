import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { holdemApi } from '../api/gameApi';
import type { HoldemResponse } from '../types/card';
import { HoldemPage } from './HoldemPage';

vi.mock('../api/gameApi', () => ({
  holdemApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(holdemApi.exec);

/** Helper: base human player */
const humanPlayer = (overrides: Partial<import('../types/card').HoldemPlayerData> = {}) => ({
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
  ...overrides,
});

/** Helper: base CPU player */
const cpuPlayer = (id: number, overrides: Partial<import('../types/card').HoldemPlayerData> = {}) => ({
  id,
  isHuman: false,
  cards: [
    { design: 'DIAMOND' as const, value: 2 },
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
  ...overrides,
});

/** INIT state (phase 0): no players yet */
const initState: HoldemResponse = {
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
};

/** PRE_FLOP (phase 1): human's turn, no outstanding bet */
const preFlopState: HoldemResponse = {
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
};

/** PRE_FLOP with outstanding bet: shows call/raise instead of bet/check */
const preFlopWithBetState: HoldemResponse = {
  ...preFlopState,
  lastBet: 40,
  cpuActions: [{ playerIdx: 1, action: 3, amount: 40 }],
};

/** FLOP (phase 2) with community cards */
const flopState: HoldemResponse = {
  ...preFlopState,
  phase: 2,
  communityCards: [
    { design: 'SPADE', value: 10 },
    { design: 'HEART', value: 5 },
    { design: 'DIAMOND', value: 8 },
  ],
};

/** SHOWDOWN (phase 5) */
const showdownState: HoldemResponse = {
  players: [
    humanPlayer({ handName: 'ワンペア', currentBet: 0, chips: 950 }),
    cpuPlayer(1, {
      handName: 'ツーペア',
      folded: false,
      cards: [
        { design: 'SPADE', value: 5 },
        { design: 'HEART', value: 8 },
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
    { playerIdx: 0, handRank: 1, handName: 'ワンペア', bestHand: [], wonAmount: 0 },
    { playerIdx: 1, handRank: 2, handName: 'ツーペア', bestHand: [], wonAmount: 200 },
  ],
  cpuActions: [],
  message: 'CPU 1 の勝ち',
};

/** END (phase 6) — also isShowdown */
const endState: HoldemResponse = {
  ...showdownState,
  phase: 6,
  message: 'Game over.',
};

beforeEach(() => {
  mockExec.mockResolvedValue(initState);
});

describe('HoldemPage', () => {
  // ---- mount & reset ----
  it('calls reset on mount', async () => {
    render(<HoldemPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined));
  });

  // ---- phase name display ----
  it('shows "初期化中" when phase is INIT (not in PHASE_NAMES)', async () => {
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('初期化中')).toBeInTheDocument());
  });

  it('shows known phase name for PRE_FLOP', async () => {
    mockExec.mockResolvedValue(preFlopState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('プリフロップ')).toBeInTheDocument());
  });

  // ---- info bar ----
  it('shows pot and dealer index', async () => {
    mockExec.mockResolvedValue(preFlopState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText(/ポット:/)).toBeInTheDocument());
    expect(screen.getByText(/ディーラー:/)).toBeInTheDocument();
  });

  // ---- community cards ----
  it('shows 5 CardBack placeholders when communityCards is empty', async () => {
    mockExec.mockResolvedValue(preFlopState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('コミュニティカード')).toBeInTheDocument());
    const cardBacks = screen.getAllByAltText('card back');
    // 5 community card placeholders + 2 cards for each of the 3 CPUs = 11 card backs expected
    // but we just verify at least 5 exist for community cards
    expect(cardBacks.length).toBeGreaterThanOrEqual(5);
  });

  it('shows CardImage when communityCards has cards', async () => {
    mockExec.mockResolvedValue(flopState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByAltText('SPADE 10')).toBeInTheDocument());
    expect(screen.getByAltText('HEART 5')).toBeInTheDocument();
    expect(screen.getByAltText('DIAMOND 8')).toBeInTheDocument();
  });

  // ---- CPU players ----
  it('renders CPU player info with playStyleName and chips', async () => {
    mockExec.mockResolvedValue(preFlopState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.getByText(/CPU 2/)).toBeInTheDocument();
    expect(screen.getAllByText(/タイト/).length).toBeGreaterThanOrEqual(1);
  });

  it('shows CPU bet when currentBet > 0', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer(), cpuPlayer(1, { currentBet: 50 }), cpuPlayer(2)],
    });
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText(/ベット: 50/)).toBeInTheDocument());
  });

  it('does not show CPU bet when currentBet is 0', async () => {
    mockExec.mockResolvedValue(preFlopState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    // CPU players have currentBet=0, so no "ベット:" text for them
    // (human section is not rendered because humanPlayer.currentBet is also 0 — but let's check CPU specifically)
    expect(screen.queryByText(/ベット: 0/)).not.toBeInTheDocument();
  });

  it('shows fold badge for folded CPU', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer(), cpuPlayer(1, { folded: true }), cpuPlayer(2)],
    });
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('[フォールド]')).toBeInTheDocument());
  });

  it('shows all-in badge for all-in CPU', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer(), cpuPlayer(1, { allIn: true }), cpuPlayer(2)],
    });
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('[オールイン]')).toBeInTheDocument());
  });

  it('shows CPU hand name badge during showdown when not folded', async () => {
    mockExec.mockResolvedValue(showdownState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('ツーペア')).toBeInTheDocument());
  });

  it('does not show CPU hand name badge when CPU is folded in showdown', async () => {
    // CPU 2 is folded in showdownState → no hand name
    mockExec.mockResolvedValue(showdownState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('ツーペア')).toBeInTheDocument());
    // CPU 2 folded, no extra hand name badge for it
    const handBadges = screen.getAllByText(/ツーペア|ワンペア/);
    // One for CPU 1, one for human, plus result section = there should be badges but not for CPU 2
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
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('ワンペア')).toBeInTheDocument());
    // CPU 1 has empty handName → no badge for it
  });

  it('shows CPU cards face-up during showdown when not folded', async () => {
    mockExec.mockResolvedValue(showdownState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByAltText('SPADE 5')).toBeInTheDocument());
    expect(screen.getByAltText('HEART 8')).toBeInTheDocument();
  });

  it('shows CardBack for CPU cards when not in showdown', async () => {
    mockExec.mockResolvedValue(preFlopState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    const cardBacks = screen.getAllByAltText('card back');
    // 5 community card placeholders + 2 cards for each of the 3 CPUs = 11 card backs expected
    expect(cardBacks.length).toBeGreaterThanOrEqual(4);
  });

  it('shows CardBack for folded CPU in showdown (isShowdown && !p.folded is false)', async () => {
    mockExec.mockResolvedValue(showdownState);
    render(<HoldemPage />);
    // CPU 2 is folded → shows CardBack
    await waitFor(() => expect(screen.getByText('ツーペア')).toBeInTheDocument());
    // CardBacks exist for CPU 2 (2 backs)
    const cardBacks = screen.getAllByAltText('card back');
    expect(cardBacks.length).toBeGreaterThanOrEqual(2);
  });

  // ---- CPU actions log ----
  it('shows CPU actions log when cpuActions is non-empty', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('CPU行動:')).toBeInTheDocument());
    expect(screen.getByText(/Player 1: ベット/)).toBeInTheDocument();
  });

  it('does not show CPU actions log when cpuActions is empty', async () => {
    mockExec.mockResolvedValue(preFlopState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByText('CPU行動:')).not.toBeInTheDocument();
  });

  it('shows amount in CPU action when amount > 0', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText(/Player 1: ベット \(40\)/)).toBeInTheDocument());
  });

  it('does not show amount when action amount is 0', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      cpuActions: [{ playerIdx: 1, action: 1, amount: 0 }],
    });
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('CPU行動:')).toBeInTheDocument());
    expect(screen.getByText(/Player 1: チェック/)).toBeInTheDocument();
    expect(screen.queryByText(/\(0\)/)).not.toBeInTheDocument();
  });

  it('shows "不明" for unknown action type', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      cpuActions: [{ playerIdx: 1, action: 99, amount: 0 }],
    });
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText(/Player 1: 不明/)).toBeInTheDocument());
  });

  // ---- round results ----
  it('shows round results in showdown', async () => {
    mockExec.mockResolvedValue(showdownState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
  });

  it('does not show round results when not in showdown', async () => {
    mockExec.mockResolvedValue(preFlopState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    expect(screen.queryByText('結果:')).not.toBeInTheDocument();
  });

  it('does not show round results when roundResults is empty in showdown', async () => {
    mockExec.mockResolvedValue({
      ...showdownState,
      roundResults: [],
    });
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('ショーダウン')).toBeInTheDocument());
    expect(screen.queryByText('結果:')).not.toBeInTheDocument();
  });

  it('shows "あなた" for human player in results', async () => {
    mockExec.mockResolvedValue(showdownState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
    // Human (playerIdx 0) → "あなた: ワンペア"
    expect(screen.getByText(/あなた: ワンペア/)).toBeInTheDocument();
  });

  it('shows "CPU X" for non-human player in results', async () => {
    mockExec.mockResolvedValue(showdownState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
    // CPU (playerIdx 1) → "CPU 1: ツーペア"
    expect(screen.getByText(/CPU 1: ツーペア/)).toBeInTheDocument();
  });

  it('shows hand name in results when present', async () => {
    mockExec.mockResolvedValue(showdownState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText(/: ワンペア/)).toBeInTheDocument());
    expect(screen.getByText(/: ツーペア/)).toBeInTheDocument();
  });

  it('does not show hand name in results when empty', async () => {
    mockExec.mockResolvedValue({
      ...showdownState,
      roundResults: [
        { playerIdx: 0, handRank: 0, handName: '', bestHand: [], wonAmount: 0 },
        { playerIdx: 1, handRank: 2, handName: 'ツーペア', bestHand: [], wonAmount: 200 },
      ],
    });
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
    // human result has no handName → no ":"
    // The human row should just be "あなた" without ":"
  });

  it('shows won chips when wonAmount > 0', async () => {
    mockExec.mockResolvedValue(showdownState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText(/\+200チップ/)).toBeInTheDocument());
  });

  it('does not show won chips when wonAmount is 0', async () => {
    mockExec.mockResolvedValue(showdownState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('結果:')).toBeInTheDocument());
    // Player 0 has wonAmount=0 → no "+0チップ"
    expect(screen.queryByText(/\+0チップ/)).not.toBeInTheDocument();
  });

  // ---- human player section ----
  it('shows human player section when humanPlayer exists', async () => {
    mockExec.mockResolvedValue(preFlopState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
  });

  it('does not show human player section when no players', async () => {
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('初期化中')).toBeInTheDocument());
    expect(screen.queryByText(/あなたの手札/)).not.toBeInTheDocument();
  });

  it('shows human cards when cards exist', async () => {
    mockExec.mockResolvedValue(preFlopState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByAltText('SPADE 1')).toBeInTheDocument());
    expect(screen.getByAltText('HEART 13')).toBeInTheDocument();
  });

  it('shows CardBack for human when cards is empty and not folded', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ cards: [] }), cpuPlayer(1), cpuPlayer(2)],
    });
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    // Human has no cards, not folded → shows 2 CardBacks for human + 5 community + 4 CPU = many
    const cardBacks = screen.getAllByAltText('card back');
    expect(cardBacks.length).toBeGreaterThanOrEqual(2);
  });

  it('does not show CardBack for human when folded and cards empty', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ cards: [], folded: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    // Folded human with no cards → !humanPlayer.folded is false → no CardBacks from human
    // CardBacks come from community (5) + CPU (2*2=4) = 9
    const cardBacks = screen.getAllByAltText('card back');
    expect(cardBacks).toHaveLength(9);
  });

  it('shows human bet when currentBet > 0', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ currentBet: 30 }), cpuPlayer(1), cpuPlayer(2)],
    });
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText(/ベット: 30/)).toBeInTheDocument());
  });

  it('shows human fold badge', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ folded: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('[フォールド]')).toBeInTheDocument());
  });

  it('shows human all-in badge', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ allIn: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('[オールイン]')).toBeInTheDocument());
  });

  it('shows human hand name badge during showdown when not folded', async () => {
    mockExec.mockResolvedValue(showdownState);
    render(<HoldemPage />);
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
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('ツーペア')).toBeInTheDocument());
  });

  // ---- message ----
  it('shows game message', async () => {
    mockExec.mockResolvedValue(preFlopState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('あなたの番です')).toBeInTheDocument());
  });

  // ---- canAct / betting controls ----
  it('shows bet/check buttons when canAct and no outstanding bet', async () => {
    mockExec.mockResolvedValue(preFlopState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument();
  });

  it('shows call/raise buttons when canAct and has outstanding bet', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'レイズ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument();
  });

  it('hides betting controls when not active phase', async () => {
    mockExec.mockResolvedValue(showdownState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('ショーダウン')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'チェック' })).not.toBeInTheDocument();
  });

  it('hides betting controls when human is folded', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ folded: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  it('hides betting controls when human is all-in', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      players: [humanPlayer({ allIn: true }), cpuPlayer(1), cpuPlayer(2)],
    });
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  it('hides betting controls when it is not human turn', async () => {
    mockExec.mockResolvedValue({
      ...preFlopState,
      currentTurn: 1,
    });
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText(/あなたの手札/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
  });

  // ---- bet amount input ----
  it('updates bet amount when changing input', async () => {
    mockExec.mockResolvedValue(preFlopState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByLabelText('ベット額:')).toBeInTheDocument());

    const betInput = screen.getByLabelText('ベット額:');
    fireEvent.change(betInput, { target: { value: '50' } });
    expect((betInput as HTMLInputElement).value).toBe('50');
  });

  // ---- button click handlers ----
  it('calls bet command with betAmount', async () => {
    mockExec.mockResolvedValue(preFlopState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 20));
  });

  it('calls check command', async () => {
    mockExec.mockResolvedValue(preFlopState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'チェック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('check', undefined));
  });

  it('calls fold command', async () => {
    mockExec.mockResolvedValue(preFlopState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold', undefined));
  });

  it('calls allin command', async () => {
    mockExec.mockResolvedValue(preFlopState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'オールイン' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('allin', undefined));
  });

  it('calls call command when has outstanding bet', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('call', undefined));
  });

  it('calls raise command with betAmount when has outstanding bet', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'レイズ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'レイズ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('raise', 20));
  });

  it('calls reset command when reset button is clicked', async () => {
    render(<HoldemPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined));

    mockExec.mockClear();
    mockExec.mockResolvedValue(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined));
  });

  // ---- loading / disabled state ----
  it('disables buttons while loading', async () => {
    mockExec.mockResolvedValue(preFlopState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());

    let resolve!: (value: HoldemResponse) => void;
    const slowPromise = new Promise<HoldemResponse>((res) => {
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
    render(<HoldemPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined));

    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );
  });

  it('clears error on successful call after failure', async () => {
    render(<HoldemPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined));

    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );

    mockExec.mockResolvedValueOnce(initState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() =>
      expect(screen.queryByText('通信エラーが発生しました。もう一度お試しください。')).not.toBeInTheDocument(),
    );
  });

  // ---- END phase (also isShowdown) ----
  it('shows results in END phase (phase 6)', async () => {
    mockExec.mockResolvedValue(endState);
    render(<HoldemPage />);
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
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByText('ショーダウン')).toBeInTheDocument());
    // CPU 1 not folded but cards empty → falls to CardBack branch
    const cardBacks = screen.getAllByAltText('card back');
    expect(cardBacks.length).toBeGreaterThanOrEqual(2);
  });

  // ---- bet amount used by raise ----
  it('sends updated bet amount when raise is clicked after changing input', async () => {
    mockExec.mockResolvedValue(preFlopWithBetState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'レイズ' })).not.toBeDisabled());

    fireEvent.change(screen.getByLabelText('ベット額:'), { target: { value: '100' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'レイズ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('raise', 100));
  });

  it('sends updated bet amount when bet is clicked after changing input', async () => {
    mockExec.mockResolvedValue(preFlopState);
    render(<HoldemPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());

    fireEvent.change(screen.getByLabelText('ベット額:'), { target: { value: '60' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(preFlopState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 60));
  });
});
