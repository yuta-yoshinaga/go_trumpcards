import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { cucumberApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CucumberResponse } from '../types/card';
import { CucumberPage } from './CucumberPage';

vi.mock('../api/gameApi', () => ({
  cucumberApi: { exec: vi.fn() },
  actionLogApi: { cucumber: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(cucumberApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const hand = [card('SPADE', 3), card('HEART', 10), card('CLOVER', 12), card('DIAMOND', 5)];

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 7,
  cards: id === 0 ? hand : [],
  penalty: 0,
  ...over,
});

function makeState(overrides: Partial<CucumberResponse> = {}): CucumberResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    validPlays: [1, 2],
    highestInTrick: 9,
    forced: false,
    currentTrick: [{ playerIdx: 1, card: card('DIAMOND', 9) }],
    currentPlayerIdx: 0,
    leadPlayerIdx: 1,
    trickNumber: 2,
    roundNumber: 3,
    lastTrickWinnerIdx: -1,
    lastPenalty: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    config: { playerCnt: 4, targetScore: 30 },
    message: '',
    ...overrides,
  } as unknown as CucumberResponse;
}

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('CucumberPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<CucumberPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **スート無関係・失点は最終トリックだけ、が規則そのもの。**
  it('states the comparison rule and that only the last trick scores', async () => {
    renderWithProviders(<CucumberPage />);
    const rule = await screen.findByTestId('cu-rule');
    expect(rule).toHaveTextContent(/スートは一切関係ありません/);
    expect(rule).toHaveTextContent(/最終トリックを取った1人だけ/);
  });

  // **超えるべきランクが盤面の全て。** スートが無い以上、唯一の手がかり。
  it('shows the rank to beat, or that you lead', async () => {
    const { unmount } = renderWithProviders(<CucumberPage />);
    expect(await screen.findByTestId('cu-threshold')).toHaveTextContent('9');
    unmount();

    mockExec.mockResolvedValue(makeState({ highestInTrick: 0, currentTrick: [] }));
    renderWithProviders(<CucumberPage />);
    expect(await screen.findByTestId('cu-threshold')).toHaveTextContent(/リード/);
  });

  it('shows every hand size and penalty', async () => {
    mockExec.mockResolvedValue(makeState({ players: [seat(0, { penalty: 12 }), seat(1), seat(2), seat(3)] }));
    renderWithProviders(<CucumberPage />);
    const s0 = await screen.findByTestId('cu-seat-0');
    expect(s0).toHaveTextContent('手札7枚');
    expect(s0).toHaveTextContent('失点12点');
  });

  it('marks who took the last trick and for how much', async () => {
    const { unmount } = renderWithProviders(<CucumberPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByTestId('cu-seat-2')).not.toHaveTextContent(/最終トリック/);
    unmount();

    mockExec.mockResolvedValue(makeState({ lastTrickWinnerIdx: 2, lastPenalty: 11 }));
    renderWithProviders(<CucumberPage />);
    expect(await screen.findByTestId('cu-seat-2')).toHaveTextContent(/最終トリックで11点/);
  });

  it('plays the clicked card by its hand index', async () => {
    renderWithProviders(<CucumberPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す$/ });
    mockExec.mockClear();
    fireEvent.click(cards[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1));
  });

  // **「選べる」と「決まっている」を言い分けます。**
  it('distinguishes a forced play from a choice of high cards', async () => {
    const { unmount } = renderWithProviders(<CucumberPage />);
    expect(await screen.findByTestId('cu-status')).toHaveTextContent(/9 より高い札/);
    unmount();

    // **合法手が1つでも forced とは限らない。** サーバのフラグを使う。
    mockExec.mockResolvedValue(makeState({ validPlays: [1], forced: false }));
    const second = renderWithProviders(<CucumberPage />);
    expect(await screen.findByTestId('cu-status')).toHaveTextContent(/9 より高い札/);
    second.unmount();

    mockExec.mockResolvedValue(makeState({ validPlays: [0], forced: true }));
    renderWithProviders(<CucumberPage />);
    expect(await screen.findByTestId('cu-status')).toHaveTextContent(/いちばん低い札/);
  });

  it('says you may lead with anything', async () => {
    mockExec.mockResolvedValue(makeState({ highestInTrick: 0, currentTrick: [], validPlays: [0, 1, 2, 3] }));
    renderWithProviders(<CucumberPage />);
    expect(await screen.findByTestId('cu-status')).toHaveTextContent(/どの札でも出せます/);
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
    renderWithProviders(<CucumberPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す$/ });
    expect(cards[0]).toBeDisabled();
    expect(screen.queryByTestId('cu-status')).not.toBeInTheDocument();
  });

  // **失点はラウンドに1回だけの出来事。** 配り直す前に読ませます。
  it('shows the round penalty and deals the next round on request', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, lastTrickWinnerIdx: 1, lastPenalty: 13 }));
    renderWithProviders(<CucumberPage />);
    expect(await screen.findByTestId('cu-round-end')).toHaveTextContent(/CPU1 が最終トリックを取り/);

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('cu-next-btn'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('reports who finished with the fewest penalty points', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        gameEndFlag: true,
        winnerIdx: 0,
        players: [seat(0, { penalty: 8 }), seat(1), seat(2), seat(3)],
      }),
    );
    const { unmount } = renderWithProviders(<CucumberPage />);
    const banner = await screen.findByTestId('cu-result');
    expect(banner).toHaveTextContent(/あなたの勝ち/);
    expect(banner).toHaveTextContent('8');
    unmount();

    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        gameEndFlag: true,
        winnerIdx: 2,
        players: [seat(0), seat(1), seat(2, { penalty: 4 }), seat(3)],
      }),
    );
    renderWithProviders(<CucumberPage />);
    expect(await screen.findByTestId('cu-result')).toHaveTextContent(/CPU2 が失点 4 点/);
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<CucumberPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '投了' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('resets with the chosen table size and target score', async () => {
    renderWithProviders(<CucumberPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    // **サーバは 3..6 しか受けない。** 弾かれる値を並べると黙って既定に戻される。
    const options = [...screen.getByTestId('cu-players-select').querySelectorAll('option')].map((o) => o.value);
    expect(options).toEqual(['3', '4', '5', '6']);

    fireEvent.change(screen.getByTestId('cu-players-select'), { target: { value: '6' } });
    fireEvent.change(screen.getByTestId('cu-target-select'), { target: { value: '50' } });
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { playerCnt: 6, targetScore: 50 }));
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'card-1', reason: 'hint.cucumberBeat', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<CucumberPage />);
    expect(await screen.findByTestId('hint-tooltip')).toHaveTextContent(/いちばん低い札です/);
  });
});
