import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { shelemApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, ShelemResponse } from '../types/card';
import { ShelemPage } from './ShelemPage';

vi.mock('../api/gameApi', () => ({
  shelemApi: { exec: vi.fn() },
  actionLogApi: { shelem: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(shelemApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  team: id % 2,
  cardCount: 4,
  cards: id === 0 ? [card('HEART', 11), card('SPADE', 9), card('CLOVER', 1), card('DIAMOND', 5), card('SPADE', 2)] : [],
  bid: -1,
  passed: false,
  declaredShelem: false,
  trickCount: 0,
  ...over,
});

function makeState(overrides: Partial<ShelemResponse> = {}): ShelemResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    roundNumber: 1,
    trickNumber: 0,
    trumpSuit: 0,
    declarerIdx: -1,
    contract: 0,
    shelemBid: false,
    minBid: 100,
    widowSize: 4,
    discardCount: 4,
    scores: [0, 0],
    roundPoints: [0, 0],
    teamTricks: [0, 0],
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 3,
    currentTrick: [],
    validPlays: [0, 1, 2],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { target: 500 },
    message: '',
    ...overrides,
  } as unknown as ShelemResponse;
}

/** A state where the contract is settled and it is the human's turn to play. */
const playing = (over: Partial<ShelemResponse> = {}) =>
  makeState({
    phase: 2,
    trumpSuit: 3,
    declarerIdx: 1,
    contract: 130,
    widowSize: 0,
    ...over,
  } as Partial<ShelemResponse>);

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('ShelemPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<ShelemPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **点になるのは A/10/5 だけ。** 盤面から読めないので常に出す。
  it('always states the point cards', async () => {
    renderWithProviders(<ShelemPage />);
    const box = await screen.findByTestId('sh-points');
    expect(box).toHaveTextContent(/A/);
    expect(box).toHaveTextContent(/100/);
  });

  // **上回れる額だけを出す。** サーバが必ず拒否する額のボタンは作らない。
  it('offers only bids that beat the standing one', async () => {
    mockExec.mockResolvedValue(makeState({ minBid: 135 }));
    renderWithProviders(<ShelemPage />);

    expect(await screen.findByTestId('sh-bid-135-btn')).toBeInTheDocument();
    expect(screen.getByTestId('sh-bid-165-btn')).toBeInTheDocument();
    expect(screen.queryByTestId('sh-bid-130-btn')).not.toBeInTheDocument();
    expect(screen.queryByTestId('sh-bid-100-btn')).not.toBeInTheDocument();
  });

  it.each([100, 130, 165])('sends bid %s', async (bid) => {
    mockExec.mockResolvedValue(makeState({ minBid: 100 }));
    renderWithProviders(<ShelemPage />);
    const btn = await screen.findByTestId(`sh-bid-${bid.toString()}-btn`);
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', undefined, undefined, undefined, bid));
  });

  it.each([
    ['sh-shelem-btn', 'shelem'],
    ['sh-pass-btn', 'pass'],
  ])('sends %s as the %s command', async (testId, command) => {
    renderWithProviders(<ShelemPage />);
    const btn = await screen.findByTestId(testId);
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  // **誰も入札しないまま最後の1人になったら降りられない。** 負のコントロール付き。
  it('hides pass when the human is the last bidder standing', async () => {
    mockExec.mockResolvedValue(
      makeState({
        contract: 0,
        players: [seat(0), seat(1, { passed: true }), seat(2, { passed: true }), seat(3, { passed: true })],
      } as Partial<ShelemResponse>),
    );
    renderWithProviders(<ShelemPage />);

    expect(await screen.findByTestId('sh-must-bid')).toBeInTheDocument();
    expect(screen.queryByTestId('sh-pass-btn')).not.toBeInTheDocument();
    expect(screen.getByTestId('sh-bid-100-btn')).toBeEnabled();
  });

  it('still offers pass once a bid is standing', async () => {
    mockExec.mockResolvedValue(
      makeState({
        contract: 120,
        minBid: 125,
        declarerIdx: 2,
        players: [seat(0), seat(1, { passed: true }), seat(2, { bid: 120 }), seat(3, { passed: true })],
      } as Partial<ShelemResponse>),
    );
    renderWithProviders(<ShelemPage />);

    expect(await screen.findByTestId('sh-pass-btn')).toBeInTheDocument();
    expect(screen.queryByTestId('sh-must-bid')).not.toBeInTheDocument();
  });

  // **捨て札はちょうど4枚選ぶまで確定できない。** サーバが必ず拒否する。
  it('keeps the discard buttons disabled until four cards are picked', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: 1, declarerIdx: 0, contract: 130, widowSize: 0 } as Partial<ShelemResponse>),
    );
    renderWithProviders(<ShelemPage />);

    expect(await screen.findByTestId('sh-discard-1-btn')).toBeDisabled();

    const cards = await screen.findAllByRole('button', { name: /捨て札に選ぶ/ });
    for (let i = 0; i < 4; i++) fireEvent.click(cards[i]);
    expect(screen.getByTestId('sh-discard-1-btn')).toBeEnabled();

    // 1枚外すとまた押せなくなる。
    fireEvent.click(cards[0]);
    expect(screen.getByTestId('sh-discard-1-btn')).toBeDisabled();
  });

  // **捨て札は4つのインデックスとスートを一緒に送る。**
  it('sends the picked cards and the trump suit together', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: 1, declarerIdx: 0, contract: 130, widowSize: 0 } as Partial<ShelemResponse>),
    );
    renderWithProviders(<ShelemPage />);

    const cards = await screen.findAllByRole('button', { name: /捨て札に選ぶ/ });
    for (const i of [0, 2, 3, 4]) fireEvent.click(cards[i]);
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('sh-discard-3-btn'));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('discard', undefined, undefined, 3, undefined, [0, 2, 3, 4]),
    );
  });

  it('plays the clicked card by its hand index once play starts', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<ShelemPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    mockExec.mockClear();
    fireEvent.click(cards[2]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  // 契約は未定・通常・Shelem の3通りを踏む。
  it('shows the contract in each of its shapes', async () => {
    mockExec.mockResolvedValue(makeState());
    const first = renderWithProviders(<ShelemPage />);
    expect(await screen.findByTestId('sh-contract')).toHaveTextContent(/未定/);
    first.unmount();

    mockExec.mockResolvedValue(playing({ roundPoints: [45, 20] } as Partial<ShelemResponse>));
    const second = renderWithProviders(<ShelemPage />);
    expect(await screen.findByTestId('sh-contract')).toHaveTextContent('130');
    second.unmount();

    mockExec.mockResolvedValue(playing({ shelemBid: true } as Partial<ShelemResponse>));
    renderWithProviders(<ShelemPage />);
    expect(await screen.findByTestId('sh-contract')).toHaveTextContent(/Shelem/);
  });

  // 競りでの立場が席ごとに出る。
  it('labels each seat with its bidding standing', async () => {
    mockExec.mockResolvedValue(
      makeState({
        declarerIdx: 0,
        contract: 130,
        players: [seat(0, { bid: 130 }), seat(1, { passed: true }), seat(2, { bid: 120 }), seat(3)],
      } as Partial<ShelemResponse>),
    );
    renderWithProviders(<ShelemPage />);

    expect(await screen.findByTestId('sh-seat-0')).toHaveTextContent('落札 130');
    expect(screen.getByTestId('sh-seat-1')).toHaveTextContent('降り');
    expect(screen.getByTestId('sh-seat-2')).toHaveTextContent('入札 120');
    expect(screen.getByTestId('sh-seat-3')).toHaveTextContent('競り中');
  });

  it('advances the round when the next-round button is pressed', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 3 }));
    renderWithProviders(<ShelemPage />);

    const btn = await screen.findByRole('button', { name: '次のラウンドへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<ShelemPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '投了' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('renders the result banner for each outcome', async () => {
    for (const [winnerTeam, expected] of [
      [0, /あなたのチームの勝ち/],
      [1, /相手チームの勝ち/],
      [-1, /同点/],
    ] as const) {
      mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: 4, winnerTeam }));
      const { unmount } = renderWithProviders(<ShelemPage />);
      expect(await screen.findByText(expected)).toBeInTheDocument();
      unmount();
    }
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(playing({ currentPlayerIdx: 1 } as Partial<ShelemResponse>));
    renderWithProviders(<ShelemPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'bid-125', reason: 'hint.shelemBid', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<ShelemPage />);
    expect(await screen.findByText(/競り落とす価値があります/)).toBeInTheDocument();
  });
});
