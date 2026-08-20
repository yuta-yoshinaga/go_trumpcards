import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { hokmApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, HokmResponse } from '../types/card';
import { HokmPage } from './HokmPage';

vi.mock('../api/gameApi', () => ({
  hokmApi: { exec: vi.fn() },
  actionLogApi: { hokm: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(hokmApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  team: id % 2,
  cardCount: 3,
  cards: id === 0 ? [card('HEART', 11), card('SPADE', 9), card('CLOVER', 1)] : [],
  isHakem: id === 0,
  trickCount: 0,
  ...over,
});

function makeState(overrides: Partial<HokmResponse> = {}): HokmResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    handNumber: 1,
    trickNumber: 0,
    trumpSuit: 0,
    hakemIdx: 0,
    scores: [0, 0],
    teamTricks: [0, 0],
    tricksToWin: 7,
    lastHandKot: false,
    lastHandWinner: -1,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    currentTrick: [],
    validPlays: [0, 1, 2],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { target: 7 },
    message: '',
    ...overrides,
  } as unknown as HokmResponse;
}

/** A state where trump is settled and it is the human's turn to play. */
const playing = (over: Partial<HokmResponse> = {}) =>
  makeState({ phase: 1, trumpSuit: 3, ...over } as Partial<HokmResponse>);

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('HokmPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<HokmPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **13トリック打ち切らないことは盤面から読めない。** 7先取の競り合いを常に出す。
  it('leads with the race to seven tricks', async () => {
    mockExec.mockResolvedValue(playing({ teamTricks: [5, 2] } as Partial<HokmResponse>));
    renderWithProviders(<HokmPage />);

    const race = await screen.findByTestId('hk-race');
    expect(race).toHaveTextContent('5');
    expect(race).toHaveTextContent('2');
    expect(race).toHaveTextContent('7');
  });

  it('offers all four trump suits to the hakem', async () => {
    renderWithProviders(<HokmPage />);
    for (const suit of [1, 2, 3, 4]) {
      expect(await screen.findByTestId(`hk-trump-${suit.toString()}-btn`)).toBeInTheDocument();
    }
  });

  it('hides the trump buttons when the human is not the hakem', async () => {
    mockExec.mockResolvedValue(makeState({ hakemIdx: 2 }));
    renderWithProviders(<HokmPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('hk-trump-1-btn')).not.toBeInTheDocument();
  });

  // **切り札は4番目の引数で送る。** 位置がずれると別の値として届く。
  it.each([1, 2, 3, 4])('sends trump suit %s', async (suit) => {
    renderWithProviders(<HokmPage />);
    const btn = await screen.findByTestId(`hk-trump-${suit.toString()}-btn`);
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trump', undefined, undefined, suit));
  });

  it('plays the clicked card by its hand index', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<HokmPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    mockExec.mockClear();
    fireEvent.click(cards[2]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  // **親と味方は盤面から読めない。** 席ごとに出す。
  it('marks the hakem and each seat team', async () => {
    mockExec.mockResolvedValue(
      makeState({
        hakemIdx: 1,
        players: [
          seat(0, { isHakem: false, trickCount: 3 }),
          seat(1, { isHakem: true }),
          seat(2, { isHakem: false }),
          seat(3, { isHakem: false }),
        ],
      } as Partial<HokmResponse>),
    );
    renderWithProviders(<HokmPage />);

    expect(await screen.findByTestId('hk-seat-1')).toHaveTextContent('親');
    expect(screen.getByTestId('hk-seat-0')).not.toHaveTextContent('親');
    expect(screen.getByTestId('hk-seat-0')).toHaveTextContent('T0');
    expect(screen.getByTestId('hk-seat-1')).toHaveTextContent('T1');
    expect(screen.getByTestId('hk-seat-2')).toHaveTextContent('T0');
    expect(screen.getByTestId('hk-seat-0')).toHaveTextContent('3');
  });

  // **Kot は2点。** 何が起きたか言わないと得点が飛んで見える。両側を踏む。
  it('explains how the hand ended', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2, lastHandWinner: 0, lastHandKot: true }));
    const { unmount } = renderWithProviders(<HokmPage />);
    expect(await screen.findByTestId('hk-hand-result')).toHaveTextContent(/Kot/);
    unmount();

    mockExec.mockResolvedValue(makeState({ phase: 2, lastHandWinner: 1, lastHandKot: false }));
    renderWithProviders(<HokmPage />);
    const result = await screen.findByTestId('hk-hand-result');
    expect(result).toHaveTextContent('7');
    expect(result).not.toHaveTextContent(/Kot/);
  });

  it('shows the running hand points', async () => {
    mockExec.mockResolvedValue(makeState({ scores: [4, 2] } as Partial<HokmResponse>));
    renderWithProviders(<HokmPage />);
    expect(await screen.findByTestId('hk-score')).toHaveTextContent('4');
    expect(screen.getByTestId('hk-score')).toHaveTextContent('2');
  });

  // 切り札は未定と確定の両側を踏む。
  it('shows the trump once declared', async () => {
    const { unmount } = renderWithProviders(<HokmPage />);
    expect(await screen.findByTestId('hk-trump')).toHaveTextContent(/未定/);
    unmount();

    mockExec.mockResolvedValue(playing({ trumpSuit: 4 } as Partial<HokmResponse>));
    renderWithProviders(<HokmPage />);
    expect(await screen.findByTestId('hk-trump')).toHaveTextContent('♦');
  });

  it('advances to the next hand when the button is pressed', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2, lastHandWinner: 0 }));
    renderWithProviders(<HokmPage />);

    const btn = await screen.findByRole('button', { name: '次のハンドへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<HokmPage />);
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
      const { unmount } = renderWithProviders(<HokmPage />);
      expect(await screen.findByText(expected)).toBeInTheDocument();
      unmount();
    }
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(playing({ currentPlayerIdx: 1 } as Partial<HokmResponse>));
    renderWithProviders(<HokmPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'trump-3', reason: 'hint.hokmDeclareTrump', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<HokmPage />);
    expect(await screen.findByText(/いちばん長いスート/)).toBeInTheDocument();
  });
});

// **親は負けたときだけ交代する** (#5753)。次に自分が切り札を選べるかを
// 左右するのに、次ハンドが始まって親バッジが動くまで分からなかった。
describe('HokmPage hakem hand-over notice', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  it('says the hakem moves after a losing hand', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        lastHandWinner: 1,
        lastHandKot: false,
        lastHandHakemChanged: true,
      } as Partial<HokmResponse>),
    );
    renderWithProviders(<HokmPage />);
    expect(await screen.findByTestId('hk-hakem-change')).toHaveTextContent('親が次の席に移ります');
  });

  it('says the hakem keeps the deal after a winning hand', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        lastHandWinner: 0,
        lastHandKot: false,
        lastHandHakemChanged: false,
      } as Partial<HokmResponse>),
    );
    renderWithProviders(<HokmPage />);
    expect(await screen.findByTestId('hk-hakem-change')).toHaveTextContent('親は交代しません');
  });

  // Kot でも通常勝利でも同じように出る (受け入れ条件4)。
  it('combines with the Kot result', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        lastHandWinner: 1,
        lastHandKot: true,
        lastHandHakemChanged: true,
      } as Partial<HokmResponse>),
    );
    renderWithProviders(<HokmPage />);
    const result = await screen.findByTestId('hk-hand-result');
    expect(result).toHaveTextContent('Kot');
    expect(result).toHaveTextContent('親が次の席に移ります');
  });
});
