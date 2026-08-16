import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { germanwhistApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, GermanWhistResponse } from '../types/card';
import { GermanWhistPage } from './GermanWhistPage';

vi.mock('../api/gameApi', () => ({
  germanwhistApi: { exec: vi.fn() },
  actionLogApi: { germanwhist: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(germanwhistApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<GermanWhistResponse> = {}): GermanWhistResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 3,
        cards: [card('SPADE', 1), card('HEART', 10), card('DIAMOND', 11)],
        trickCount: 2,
        scoringTricks: 0,
      },
      { id: 1, isHuman: false, cardCount: 3, cards: [], trickCount: 1, scoringTricks: 0 },
    ],
    phase: 0,
    trickNumber: 3,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    currentTrick: [],
    trumpSuit: 3,
    upCard: card('DIAMOND', 12),
    stockCount: 20,
    validPlays: [0, 1, 2],
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  } as unknown as GermanWhistResponse;
}

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('GermanWhistPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<GermanWhistPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('plays the clicked card by its hand index', async () => {
    renderWithProviders(<GermanWhistPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cards = await screen.findAllByRole('button', { name: /を出す|^Play / });
    mockExec.mockClear();
    fireEvent.click(cards[1]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1));
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<GermanWhistPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '投了' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  // 投了ボタンは終局後に消える。負のコントロール付き。
  it('hides give-up once the game is over', async () => {
    mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: 2, winnerIdx: 0 }));
    renderWithProviders(<GermanWhistPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    await waitFor(() => expect(screen.queryByRole('button', { name: '投了' })).not.toBeInTheDocument());
  });

  // **前半と後半で表示が入れ替わる。**得点になるのは後半だけ、という
  // このゲームの肝が画面に出ていなければ意味がない。
  it('labels the first half as non-scoring', async () => {
    renderWithProviders(<GermanWhistPage />);
    expect(await screen.findByTestId('gw-phase')).toHaveTextContent('前半');
    expect(screen.getByText(/前半のトリックは得点になりません/)).toBeInTheDocument();
  });

  it('labels the second half as scoring', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, upCard: undefined, stockCount: 0 }));
    renderWithProviders(<GermanWhistPage />);
    expect(await screen.findByTestId('gw-phase')).toHaveTextContent('後半');
    expect(screen.queryByText(/前半のトリックは得点になりません/)).not.toBeInTheDocument();
  });

  // 表向きの札は前半の主役。尽きたら「なし」に変わる。
  it('shows the face-up card in the first half and reports its absence later', async () => {
    const { unmount } = renderWithProviders(<GermanWhistPage />);
    await waitFor(() => expect(screen.getByTestId('gw-upcard')).toHaveTextContent('表向きの札'));
    unmount();

    mockExec.mockResolvedValue(makeState({ phase: 1, upCard: undefined, stockCount: 0 }));
    renderWithProviders(<GermanWhistPage />);
    await waitFor(() => expect(screen.getByTestId('gw-upcard')).toHaveTextContent(/表向きの札なし/));
  });

  it('renders the result banner for each outcome', async () => {
    for (const [winnerIdx, expected] of [
      [0, /あなたの勝ち/],
      [1, /CPUの勝ち/],
      [-1, /引き分け/],
    ] as const) {
      mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: 2, winnerIdx }));
      const { unmount } = renderWithProviders(<GermanWhistPage />);
      expect(await screen.findByText(expected)).toBeInTheDocument();
      unmount();
    }
  });

  // 手番でなければ札は押せない。
  it('disables the hand while it is the CPU turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
    renderWithProviders(<GermanWhistPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'card-1', reason: 'hint.germanWhistDuck', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<GermanWhistPage />);
    expect(await screen.findByText(/わざと負けましょう/)).toBeInTheDocument();
  });

  // **前半は得点が両者 0 のまま。** trickCount を出さないと、13 トリックが
  // どちらに有利に進んでいるか画面から読めない (#5744)。CUI は最初から両方出している。
  it('shows the running trick count for both seats, not just the scoring ones', async () => {
    // 既定の fixture は前半 (phase 0) で、得点はどちらも 0、通算は 2 対 1。
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<GermanWhistPage />);

    const human = await screen.findByTestId('gw-human-tricks');
    const cpu = screen.getByTestId('gw-cpu-tricks');

    // 通算と得点が別々に読めること。得点だけなら 0 対 0 で区別がつかない。
    expect(human).toHaveTextContent('獲得トリック: 2');
    expect(cpu).toHaveTextContent('獲得トリック: 1');
    expect(human).toHaveTextContent('得点トリック: 0');
    expect(cpu).toHaveTextContent('得点トリック: 0');
  });

  // `cpu` は players.find(...) なので undefined になりうる。フォールバックが
  // 効かないと画面に `NaN` や空欄が出る。
  it('falls back to zero when the CPU seat is missing', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 3,
            cards: [card('SPADE', 1)],
            trickCount: 5,
            scoringTricks: 2,
          },
        ],
      }),
    );
    renderWithProviders(<GermanWhistPage />);

    const cpu = await screen.findByTestId('gw-cpu-tricks');
    expect(cpu).toHaveTextContent('獲得トリック: 0');
    expect(cpu).not.toHaveTextContent('NaN');
    // 人間側は通常どおり出ること (フォールバックが全体を潰していない)。
    expect(screen.getByTestId('gw-human-tricks')).toHaveTextContent('獲得トリック: 5');
  });

  it('keeps both numbers distinguishable in the second half', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 1,
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 3,
            cards: [card('SPADE', 1)],
            trickCount: 9,
            scoringTricks: 4,
          },
          { id: 1, isHuman: false, cardCount: 3, cards: [], trickCount: 7, scoringTricks: 3 },
        ],
      }),
    );
    renderWithProviders(<GermanWhistPage />);

    const human = await screen.findByTestId('gw-human-tricks');
    expect(human).toHaveTextContent('獲得トリック: 9');
    expect(human).toHaveTextContent('得点トリック: 4');
    // 同じ数字が両方に使われていたら見分けられない。
    expect(human).not.toHaveTextContent('獲得トリック: 4');
  });
});
