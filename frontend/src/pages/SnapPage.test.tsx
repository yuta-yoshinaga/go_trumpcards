import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { snapApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, SnapResponse } from '../types/card';
import { SnapPage } from './SnapPage';

vi.mock('../api/gameApi', () => ({
  snapApi: { exec: vi.fn() },
  actionLogApi: { snap: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(snapApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<SnapResponse> = {}): SnapResponse {
  return {
    phase: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    currentTurnIdx: 0,
    isHumanTurn: true,
    snapAvailable: false,
    centerPileSize: 3,
    topCard: card('SPADE', 7),
    players: [
      { id: 0, isHuman: true, stockSize: 24 },
      { id: 1, isHuman: false, stockSize: 25 },
    ],
    playerCnt: 2,
    cpuDifficulty: 1,
    pendingKind: 0,
    pendingDeadlineMs: 0,
    lastEventKind: 0,
    lastEventPlayerIdx: 0,
    message: '',
    ...overrides,
  } as unknown as SnapResponse;
}

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

afterEach(() => {
  vi.useRealTimers();
});

describe('SnapPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<SnapPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **トリガーが動くことが規則そのもの。**
  it('states the rule', async () => {
    renderWithProviders(<SnapPage />);
    expect(await screen.findByTestId('sp-rule')).toHaveTextContent(/直前に出た札と同じランク/);
  });

  it('shows the pile and every stock', async () => {
    renderWithProviders(<SnapPage />);
    expect(await screen.findByTestId('sp-pile')).toHaveTextContent('3');
    expect(screen.getByTestId('sp-seat-0')).toHaveTextContent('24');
    expect(screen.getByTestId('sp-seat-1')).toHaveTextContent('25');
  });

  it('says when the pile is empty', async () => {
    mockExec.mockResolvedValue(makeState({ centerPileSize: 0, topCard: undefined }));
    renderWithProviders(<SnapPage />);
    expect(await screen.findByTestId('sp-pile')).toHaveTextContent(/場札なし/);
  });

  // **成立しているかは一目で分かる必要がある。** 反射ゲームなので。
  it('announces only while a call would be correct', async () => {
    const { unmount } = renderWithProviders(<SnapPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('sp-available')).not.toBeInTheDocument();
    unmount();

    mockExec.mockResolvedValue(makeState({ snapAvailable: true }));
    renderWithProviders(<SnapPage />);
    expect(await screen.findByTestId('sp-available')).toBeInTheDocument();
  });

  it('turns a card when the step button is pressed', async () => {
    renderWithProviders(<SnapPage />);
    const btn = await screen.findByTestId('sp-step-btn');
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('step'));
  });

  it('disables the step button while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentTurnIdx: 1, isHumanTurn: false }));
    renderWithProviders(<SnapPage />);
    expect(await screen.findByTestId('sp-step-btn')).toBeDisabled();
  });

  // **宣言はいつでも押せる。** 成立していなければペナルティ——それが賭け。
  it('keeps the snap button pressable even when no call would be correct', async () => {
    renderWithProviders(<SnapPage />);
    const btn = await screen.findByTestId('sp-snap-btn');
    expect(btn).toBeEnabled();

    mockExec.mockClear();
    fireEvent.click(btn);
    // **席は送らない。** 送れると CPU に誤宣言させられる。
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('snap'));
  });

  it('keeps the snap button pressable on a CPU turn too', async () => {
    mockExec.mockResolvedValue(makeState({ currentTurnIdx: 1, isHumanTurn: false }));
    renderWithProviders(<SnapPage />);
    expect(await screen.findByTestId('sp-snap-btn')).toBeEnabled();
  });

  // **予約でゲートする。手番ではない。** CPU の宣言は人間の手番中にも予約される。
  it('polls tick while a CPU action is booked, even on the human turn', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    mockExec.mockResolvedValue(makeState({ pendingKind: 1, isHumanTurn: true, currentTurnIdx: 0 }));
    renderWithProviders(<SnapPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    await act(async () => {
      vi.advanceTimersByTime(350);
    });
    expect(mockExec).toHaveBeenCalledWith('tick');
  });

  it('does not poll when nothing is booked', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    mockExec.mockResolvedValue(makeState({ pendingKind: 0 }));
    renderWithProviders(<SnapPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    await act(async () => {
      vi.advanceTimersByTime(500);
    });
    expect(mockExec).not.toHaveBeenCalledWith('tick');
  });

  it('stops polling once the game ends', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    mockExec.mockResolvedValue(makeState({ pendingKind: 1, gameEndFlag: true, phase: 1, winnerIdx: 0 }));
    renderWithProviders(<SnapPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    await act(async () => {
      vi.advanceTimersByTime(500);
    });
    expect(mockExec).not.toHaveBeenCalledWith('tick');
  });

  // **直近に何が起きたかを出す。** 盤面だけでは誰が取ったのか読めない。
  it.each([
    [1, /めくりました/],
    [2, /総取り/],
    [3, /誤宣言/],
    [4, /尽きました/],
  ])('reports last event kind %s', async (kind, expected) => {
    mockExec.mockResolvedValue(makeState({ lastEventKind: kind, lastEventPlayerIdx: 1 }));
    renderWithProviders(<SnapPage />);
    expect(await screen.findByTestId('sp-event')).toHaveTextContent(expected);
  });

  it('shows no event line when nothing has happened', async () => {
    renderWithProviders(<SnapPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('sp-event')).not.toBeInTheDocument();
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<SnapPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '投了' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('renders the result banner for each outcome', async () => {
    for (const [winnerIdx, expected] of [
      [0, /あなたの勝ち/],
      [1, /CPU1 の勝ち/],
      [-1, /続けられなく/],
    ] as const) {
      mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: 1, winnerIdx }));
      const { unmount } = renderWithProviders(<SnapPage />);
      expect(await screen.findByTestId('sp-result')).toHaveTextContent(expected);
      unmount();
    }
  });

  it('hides the action buttons once the game ends', async () => {
    mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: 1, winnerIdx: 0 }));
    renderWithProviders(<SnapPage />);
    await screen.findByTestId('sp-result');
    expect(screen.queryByTestId('sp-step-btn')).not.toBeInTheDocument();
    expect(screen.queryByTestId('sp-snap-btn')).not.toBeInTheDocument();
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'snap', reason: 'hint.snapDeclare', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<SnapPage />);
    expect(await screen.findByTestId('hint-tooltip')).toHaveTextContent(/いま宣言/);
  });
});

// **反射ゲームの核心は相手の反応速度** (#5763)。ラベルだけでは何が変わるのか
// 分からず、単なる名前の選択に見えていた。
describe('SnapPage difficulty explanation', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  it('spells out how fast each difficulty reacts', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<SnapPage />);

    const select = await screen.findByTestId('sp-difficulty-select');
    const labels = Array.from(select.querySelectorAll('option')).map((o) => o.textContent ?? '');
    expect(labels[0]).toContain('1.4秒');
    expect(labels[1]).toContain('0.9秒');
    expect(labels[2]).toContain('0.5秒');
    // **値は変えない** (受け入れ条件3)。
    expect(Array.from(select.querySelectorAll('option')).map((o) => o.getAttribute('value'))).toEqual(['0', '1', '2']);
  });
});
