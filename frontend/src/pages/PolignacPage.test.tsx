import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { polignacApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, PolignacResponse } from '../types/card';
import { PolignacPage } from './PolignacPage';

vi.mock('../api/gameApi', () => ({
  polignacApi: { exec: vi.fn() },
  actionLogApi: { polignac: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(polignacApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 3,
  cards: id === 0 ? [card('SPADE', 11), card('HEART', 10), card('CLOVER', 1)] : [],
  score: 0,
  roundPenalty: 0,
  trickCount: 0,
  declaredCapot: false,
  ...over,
});

function makeState(overrides: Partial<PolignacResponse> = {}): PolignacResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 1,
    roundNumber: 1,
    trickNumber: 2,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 0,
    capotIdx: -1,
    capotTricks: 0,
    currentTrick: [],
    validPlays: [0, 1, 2],
    gameEndFlag: false,
    winnerIdx: -1,
    config: { rounds: 4 },
    message: '',
    ...overrides,
  } as unknown as PolignacResponse;
}

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('PolignacPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<PolignacPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('plays the clicked card by its hand index', async () => {
    renderWithProviders(<PolignacPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    mockExec.mockClear();
    fireEvent.click(cards[2]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  // **失点ルールは盤面から読み取れない。** 常時出ていなければならない。
  it('always states the jack penalty rule', async () => {
    renderWithProviders(<PolignacPage />);
    expect(await screen.findByTestId('pg-penalty-rule')).toHaveTextContent(/♠J（Polignac）は2失点/);
  });

  // 宣言フェーズでだけ capot / pass が出る。両側を踏む。
  it('offers capot and pass during the declaration phase', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 0 }));
    renderWithProviders(<PolignacPage />);

    expect(await screen.findByTestId('pg-capot-btn')).toBeInTheDocument();
    expect(screen.getByTestId('pg-pass-btn')).toBeInTheDocument();
  });

  it('hides capot and pass once play has started', async () => {
    renderWithProviders(<PolignacPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('pg-capot-btn')).not.toBeInTheDocument();
    expect(screen.queryByTestId('pg-pass-btn')).not.toBeInTheDocument();
  });

  // capot と pass は別のコマンドを送る。取り違えるとラウンドの行方が変わる。
  it.each([
    ['pg-capot-btn', 'capot'],
    ['pg-pass-btn', 'pass'],
  ])('sends %s as the %s command', async (testId, command) => {
    mockExec.mockResolvedValue(makeState({ phase: 0 }));
    renderWithProviders(<PolignacPage />);

    const btn = await screen.findByTestId(testId);
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  // capot 宣言中は全員の狙いが変わるので、必ず知らせる。負のコントロール付き。
  it('announces a declared capot with its progress', async () => {
    mockExec.mockResolvedValue(
      makeState({
        capotIdx: 2,
        capotTricks: 3,
        players: [seat(0), seat(1), seat(2, { declaredCapot: true }), seat(3)],
      }),
    );
    renderWithProviders(<PolignacPage />);

    expect(await screen.findByTestId('pg-capot-banner')).toHaveTextContent('3/8');
    expect(screen.getByTestId('pg-seat-2')).toHaveTextContent('[capot]');
  });

  it('says nothing about capot when nobody declared', async () => {
    renderWithProviders(<PolignacPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('pg-capot-banner')).not.toBeInTheDocument();
  });

  it('shows both the running total and the round penalty per seat', async () => {
    mockExec.mockResolvedValue(
      makeState({ players: [seat(0, { score: 4, roundPenalty: 2 }), seat(1), seat(2), seat(3)] }),
    );
    renderWithProviders(<PolignacPage />);
    expect(await screen.findByTestId('pg-seat-0')).toHaveTextContent('累計4失点（今ラウンド2）');
  });

  it('advances the round when the next-round button is pressed', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<PolignacPage />);

    const btn = await screen.findByRole('button', { name: '次のラウンドへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<PolignacPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '投了' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('renders the result banner for each outcome', async () => {
    for (const [winnerIdx, expected] of [
      [0, /あなたの勝ち/],
      [1, /CPU1 の勝ち/],
      [-1, /同点/],
    ] as const) {
      mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: 3, winnerIdx }));
      const { unmount } = renderWithProviders(<PolignacPage />);
      expect(await screen.findByText(expected)).toBeInTheDocument();
      unmount();
    }
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
    renderWithProviders(<PolignacPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'card-1', reason: 'hint.polignacAvoidJack', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<PolignacPage />);
    expect(await screen.findByText(/ジャックが乗っています/)).toBeInTheDocument();
  });
});

// **合計失点だけでは、♠J を踏んだのか他を 2 枚拾ったのかが分からない** (#5746)。
// 姉妹ゲームの Slobberhannes / Reversis は取った印付き札を個別に出している。
describe('PolignacPage jack breakdown', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  it('lists the jacks a seat took, spade first', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [seat(0, { roundPenalty: 3, takenJackSuits: [1, 3] }), seat(1), seat(2), seat(3)],
      }),
    );
    renderWithProviders(<PolignacPage />);

    const jacks = await screen.findByTestId('pg-jacks-0');
    expect(jacks).toHaveTextContent('♠J');
    expect(jacks).toHaveTextContent('♥J');
    // ♠ が先。重い方から読めないと内訳の意味が薄い。
    expect(jacks.textContent?.indexOf('♠J')).toBeLessThan(jacks.textContent?.indexOf('♥J') ?? -1);
    // 読み上げは失点まで言う。
    expect(jacks).toHaveTextContent('スペードのジャック（2失点）');
    expect(jacks).toHaveTextContent('ハートのジャック（1失点）');
  });

  it('shows nothing for a seat that took no jack', async () => {
    mockExec.mockResolvedValue(makeState({ players: [seat(0, { takenJackSuits: [] }), seat(1), seat(2), seat(3)] }));
    renderWithProviders(<PolignacPage />);
    await waitFor(() => expect(screen.getByTestId('pg-seat-0')).toBeInTheDocument());
    expect(screen.queryByTestId('pg-jacks-0')).not.toBeInTheDocument();
  });

  it('emphasises the spade jack and not the others', async () => {
    mockExec.mockResolvedValue(
      makeState({ players: [seat(0, { takenJackSuits: [1, 4] }), seat(1), seat(2), seat(3)] }),
    );
    renderWithProviders(<PolignacPage />);
    const jacks = await screen.findByTestId('pg-jacks-0');
    const emphasised = Array.from(jacks.querySelectorAll('span')).filter((el) =>
      el.className.includes('text-ds-error'),
    );
    expect(emphasised).toHaveLength(1);
    expect(emphasised[0]).toHaveTextContent('♠J');
  });
});
