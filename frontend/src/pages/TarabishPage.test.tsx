import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tarabishApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, TarabishResponse } from '../types/card';
import { TarabishPage } from './TarabishPage';

vi.mock('../api/gameApi', () => ({
  tarabishApi: { exec: vi.fn() },
  actionLogApi: { tarabish: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(tarabishApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  team: id % 2,
  cardCount: 3,
  cards: id === 0 ? [card('HEART', 11), card('SPADE', 9), card('CLOVER', 1)] : [],
  meldPoints: 0,
  runLen: 0,
  hasBella: false,
  trickCount: 0,
  ...over,
});

function makeState(overrides: Partial<TarabishResponse> = {}): TarabishResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    roundNumber: 1,
    trickNumber: 0,
    trumpSuit: 3,
    upCard: card('HEART', 9),
    trumpTakerIdx: -1,
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
    config: { target: 500 },
    message: '',
    ...overrides,
  } as unknown as TarabishResponse;
}

/** A state where trump is settled and it is the human's turn to play. */
const playing = (over: Partial<TarabishResponse> = {}) =>
  makeState({ phase: 1, trumpTakerIdx: 0, ...over } as Partial<TarabishResponse>);

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('TarabishPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<TarabishPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **切り札の序列は盤面から読めない。** 常に出ていなければならない。
  it('always states the trump order', async () => {
    renderWithProviders(<TarabishPage />);
    expect(await screen.findByTestId('tb-order')).toHaveTextContent(/J\(Jass\)=20/);
    expect(screen.getByTestId('tb-order')).toHaveTextContent(/9\(Menel\)=14/);
  });

  it('offers both choices while bidding', async () => {
    renderWithProviders(<TarabishPage />);
    expect(await screen.findByTestId('tb-take-btn')).toBeInTheDocument();
    expect(screen.getByTestId('tb-pass-btn')).toBeInTheDocument();
  });

  // **親は見送れないので、見送りボタンを出さない。** 負のコントロール付き。
  it('hides the pass button when the human is the dealer', async () => {
    mockExec.mockResolvedValue(makeState({ dealerIdx: 0 }));
    renderWithProviders(<TarabishPage />);

    expect(await screen.findByTestId('tb-take-btn')).toBeInTheDocument();
    expect(screen.queryByTestId('tb-pass-btn')).not.toBeInTheDocument();
    expect(screen.getByTestId('tb-dealer-stuck')).toBeInTheDocument();
  });

  // 引き受けと見送りは別のコマンドを送る。
  it.each([
    ['tb-take-btn', 'take'],
    ['tb-pass-btn', 'pass'],
  ])('sends %s as the %s command', async (testId, command) => {
    renderWithProviders(<TarabishPage />);
    const btn = await screen.findByTestId(testId);
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('hides the bid buttons once trump is settled', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<TarabishPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('tb-take-btn')).not.toBeInTheDocument();
  });

  // 入札前は候補、決まったあとは切り札。両側を踏む。
  it('shows the turned card before trump is settled', async () => {
    renderWithProviders(<TarabishPage />);
    expect(await screen.findByTestId('tb-upcard')).toBeInTheDocument();
    expect(screen.queryByTestId('tb-trump')).not.toBeInTheDocument();
  });

  it('shows who took trump once it is settled', async () => {
    mockExec.mockResolvedValue(playing({ trumpTakerIdx: 2 } as Partial<TarabishResponse>));
    renderWithProviders(<TarabishPage />);
    expect(await screen.findByTestId('tb-trump')).toHaveTextContent('CPU2');
    expect(screen.queryByTestId('tb-upcard')).not.toBeInTheDocument();
  });

  it('plays the clicked card by its hand index', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<TarabishPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    mockExec.mockClear();
    fireEvent.click(cards[2]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  // **チーム番号とメルドは盤面から読めない。** 席ごとに出す。
  it('labels each seat with its team and meld', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [seat(0, { meldPoints: 70, runLen: 4, hasBella: true }), seat(1), seat(2), seat(3)],
      } as Partial<TarabishResponse>),
    );
    renderWithProviders(<TarabishPage />);

    expect(await screen.findByTestId('tb-seat-0')).toHaveTextContent('T0');
    expect(screen.getByTestId('tb-seat-0')).toHaveTextContent('ラン4枚');
    expect(screen.getByTestId('tb-seat-0')).toHaveTextContent('ベラ');
    expect(screen.getByTestId('tb-seat-1')).toHaveTextContent('T1');
    expect(screen.getByTestId('tb-seat-1')).toHaveTextContent('メルドなし');
    // 0 と 2 が味方であることが席表示から読める。
    expect(screen.getByTestId('tb-seat-2')).toHaveTextContent('T0');
  });

  it('shows the running team scores', async () => {
    mockExec.mockResolvedValue(makeState({ scores: [220, 140] } as Partial<TarabishResponse>));
    renderWithProviders(<TarabishPage />);
    expect(await screen.findByTestId('tb-score')).toHaveTextContent('220');
    expect(screen.getByTestId('tb-score')).toHaveTextContent('140');
  });

  it('advances the round when the next-round button is pressed', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<TarabishPage />);

    const btn = await screen.findByRole('button', { name: '次のラウンドへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<TarabishPage />);
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
      const { unmount } = renderWithProviders(<TarabishPage />);
      expect(await screen.findByText(expected)).toBeInTheDocument();
      unmount();
    }
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(playing({ currentPlayerIdx: 1 } as Partial<TarabishResponse>));
    renderWithProviders(<TarabishPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'take-trump', reason: 'hint.tarabishTakeTrump', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<TarabishPage />);
    expect(await screen.findByText(/引き受けてよいでしょう/)).toBeInTheDocument();
  });
});

// **切り札だけ点数表が入れ替わるのがこの系統の肝** (#5749)。同じ J でも
// 切り札なら 20 点 (Jass)、そうでなければ 2 点。暗算させるとパートナーに
// 寄せる札を間違える。
describe('TarabishPage card points', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  it('scores the trump jack and nine by the trump table', async () => {
    // 切り札 = ♥。手札は ♥J(Jass) / ♥9(Menel) / ♠J / ♠9。
    mockExec.mockResolvedValue(
      makeState({
        trumpSuit: 3,
        players: [
          seat(0, {
            cards: [card('HEART', 11), card('HEART', 9), card('SPADE', 11), card('SPADE', 9)],
            cardCount: 4,
          }),
          seat(1),
          seat(2),
          seat(3),
        ],
      }),
    );
    renderWithProviders(<TarabishPage />);

    expect(await screen.findByTestId('tb-points-0')).toHaveTextContent('20');
    expect(screen.getByTestId('tb-points-1')).toHaveTextContent('14');
    // 同じランクでも切り札でなければ別の表。
    expect(screen.getByTestId('tb-points-2')).toHaveTextContent('2');
    expect(screen.getByTestId('tb-points-3')).toHaveTextContent('0');
  });

  it('says the points in the accessible name too', async () => {
    mockExec.mockResolvedValue(
      makeState({
        trumpSuit: 3,
        players: [seat(0, { cards: [card('HEART', 11)], cardCount: 1 }), seat(1), seat(2), seat(3)],
      }),
    );
    renderWithProviders(<TarabishPage />);
    expect(await screen.findByRole('button', { name: '♥ J（20点）を出す' })).toBeInTheDocument();
  });

  // **切り札が決まるまで点は定まらない。**入札中に出すと嘘になる。
  it('shows no points until a trump suit is called', async () => {
    mockExec.mockResolvedValue(
      makeState({
        trumpSuit: 0,
        players: [seat(0, { cards: [card('HEART', 11)], cardCount: 1 }), seat(1), seat(2), seat(3)],
      }),
    );
    renderWithProviders(<TarabishPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('tb-points-0')).not.toBeInTheDocument();
  });
});
