import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { balootApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BalootResponse, Card } from '../types/card';
import { BalootPage } from './BalootPage';

vi.mock('../api/gameApi', () => ({
  balootApi: { exec: vi.fn() },
  actionLogApi: { baloot: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(balootApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  team: id % 2,
  cardCount: 3,
  cards: id === 0 ? [card('HEART', 11), card('SPADE', 9), card('CLOVER', 1)] : [],
  hasBaloot: false,
  declared: false,
  trickCount: 0,
  ...over,
});

function makeState(overrides: Partial<BalootResponse> = {}): BalootResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    mode: 0,
    roundNumber: 1,
    trickNumber: 0,
    trumpSuit: 0,
    declarerIdx: -1,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    // 既定では親を 2 番にして、人間が見送れる状況にする。
    dealerIdx: 2,
    scores: [0, 0],
    roundPoints: [0, 0],
    currentTrick: [],
    validPlays: [0, 1, 2],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { target: 152 },
    message: '',
    ...overrides,
  } as unknown as BalootResponse;
}

/** A state where the mode is settled and it is the human's turn to play. */
const playing = (over: Partial<BalootResponse> = {}) =>
  makeState({ phase: 1, mode: 1, declarerIdx: 0, ...over } as Partial<BalootResponse>);

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('BalootPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<BalootPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('says the mode is undecided before anyone declares', async () => {
    renderWithProviders(<BalootPage />);
    expect(await screen.findByTestId('bl-mode')).toHaveTextContent(/モード未定/);
  });

  // **序列はモードで入れ替わる。** 有効な方だけを出し、他方は出さない。
  it('prints the Sun order under Sun', async () => {
    mockExec.mockResolvedValue(playing({ mode: 1 } as Partial<BalootResponse>));
    renderWithProviders(<BalootPage />);

    const order = await screen.findByTestId('bl-order');
    expect(order).toHaveTextContent(/A=11 > 10 > K=4/);
    expect(order).not.toHaveTextContent(/J=20/);
  });

  it('prints the Hokom order and the trump suit under Hokom', async () => {
    mockExec.mockResolvedValue(playing({ mode: 2, trumpSuit: 3 } as Partial<BalootResponse>));
    renderWithProviders(<BalootPage />);

    const order = await screen.findByTestId('bl-order');
    expect(order).toHaveTextContent(/J=20 > 9=14/);
    expect(order).not.toHaveTextContent(/A=11 > 10 > K=4 > Q=3 > J=2/);
    expect(screen.getByTestId('bl-mode')).toHaveTextContent('♥');
  });

  it('offers Sun, all four Hokom suits and pass while declaring', async () => {
    renderWithProviders(<BalootPage />);
    expect(await screen.findByTestId('bl-sun-btn')).toBeInTheDocument();
    for (const suit of [1, 2, 3, 4]) {
      expect(screen.getByTestId(`bl-hokom-${suit.toString()}-btn`)).toBeInTheDocument();
    }
    expect(screen.getByTestId('bl-pass-btn')).toBeInTheDocument();
  });

  // **親は見送れないので、見送りボタンを出さない。** 負のコントロール付き。
  it('hides the pass button when the human is the dealer', async () => {
    mockExec.mockResolvedValue(makeState({ dealerIdx: 0 }));
    renderWithProviders(<BalootPage />);

    expect(await screen.findByTestId('bl-sun-btn')).toBeInTheDocument();
    expect(screen.queryByTestId('bl-pass-btn')).not.toBeInTheDocument();
    expect(screen.getByTestId('bl-dealer-stuck')).toBeInTheDocument();
  });

  it.each([
    ['bl-sun-btn', 'sun'],
    ['bl-pass-btn', 'pass'],
  ])('sends %s as the %s command', async (testId, command) => {
    renderWithProviders(<BalootPage />);
    const btn = await screen.findByTestId(testId);
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  // **Hokom は選んだスートを 4 番目の引数で送る。** 位置がずれると
  // プレイヤーが選んでいないスートが切り札になる。
  it.each([1, 2, 3, 4])('sends hokom with suit %s', async (suit) => {
    renderWithProviders(<BalootPage />);
    const btn = await screen.findByTestId(`bl-hokom-${suit.toString()}-btn`);
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hokom', undefined, undefined, suit));
  });

  it('hides the declaration buttons once the mode is settled', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<BalootPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('bl-sun-btn')).not.toBeInTheDocument();
    expect(screen.queryByTestId('bl-hokom-1-btn')).not.toBeInTheDocument();
  });

  it('plays the clicked card by its hand index', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<BalootPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    mockExec.mockClear();
    fireEvent.click(cards[2]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  // **チーム番号と Baloot 役は盤面から読めない。** 席ごとに出す。
  it('labels each seat with its team and Baloot bonus', async () => {
    mockExec.mockResolvedValue(
      makeState({ players: [seat(0, { hasBaloot: true }), seat(1), seat(2), seat(3)] } as Partial<BalootResponse>),
    );
    renderWithProviders(<BalootPage />);

    expect(await screen.findByTestId('bl-seat-0')).toHaveTextContent('T0');
    expect(screen.getByTestId('bl-seat-0')).toHaveTextContent(/Baloot/);
    expect(screen.getByTestId('bl-seat-1')).toHaveTextContent('T1');
    expect(screen.getByTestId('bl-seat-1')).toHaveTextContent('役なし');
    // 0 と 2 が味方であることが席表示から読める。
    expect(screen.getByTestId('bl-seat-2')).toHaveTextContent('T0');
  });

  it('shows the running team scores', async () => {
    mockExec.mockResolvedValue(makeState({ scores: [120, 90] } as Partial<BalootResponse>));
    renderWithProviders(<BalootPage />);
    expect(await screen.findByTestId('bl-score')).toHaveTextContent('120');
    expect(screen.getByTestId('bl-score')).toHaveTextContent('90');
  });

  it('advances the round when the next-round button is pressed', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<BalootPage />);

    const btn = await screen.findByRole('button', { name: '次のラウンドへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<BalootPage />);
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
      mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: 3, winnerTeam }));
      const { unmount } = renderWithProviders(<BalootPage />);
      expect(await screen.findByText(expected)).toBeInTheDocument();
      unmount();
    }
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(playing({ currentPlayerIdx: 1 } as Partial<BalootResponse>));
    renderWithProviders(<BalootPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'declare-hokom-3', reason: 'hint.balootDeclareHokom', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<BalootPage />);
    expect(await screen.findByText(/Hokom が有利/)).toBeInTheDocument();
  });
});
