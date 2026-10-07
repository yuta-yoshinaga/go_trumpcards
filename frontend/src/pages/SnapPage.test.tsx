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
    previousCard: card('HEART', 7),
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

  it('explains that player count and CPU reaction changes apply after the next reset', async () => {
    renderWithProviders(<SnapPage />);
    await screen.findByTestId('sp-rule');
    fireEvent.click(screen.getByText('設定', { selector: 'summary' }));

    const playerSelect = screen.getByTestId('sp-players-select');
    const difficultySelect = screen.getByTestId('sp-difficulty-select');
    expect(playerSelect).toHaveAttribute('aria-describedby', 'snap-players-description');
    expect(difficultySelect).toHaveAttribute('aria-describedby', 'snap-difficulty-description');
    expect(screen.getAllByText('設定変更は次のリセット後、次のゲームから反映されます。')).toHaveLength(2);
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

  it('shows the previous and top cards with labels, and omits the previous card for a single card pile', async () => {
    renderWithProviders(<SnapPage />);
    const pile = await screen.findByTestId('sp-pile');
    expect(pile).toHaveTextContent('直前の札');
    expect(pile).toHaveTextContent('最上札');
    expect(pile.querySelectorAll('img')).toHaveLength(2);
  });

  it('does not render a comparison card when only one card is in the pile', async () => {
    mockExec.mockResolvedValue(makeState({ centerPileSize: 1, previousCard: undefined }));
    renderWithProviders(<SnapPage />);
    const pile = await screen.findByTestId('sp-pile');
    expect(pile).toHaveTextContent('最上札');
    expect(pile).not.toHaveTextContent('直前の札');
    expect(pile.querySelectorAll('img')).toHaveLength(1);
  });

  it('structures seat summaries as a list with a heading for each seat', async () => {
    renderWithProviders(<SnapPage />);
    const seatList = await screen.findByRole('list');
    expect(seatList).toHaveClass('list-none');
    expect(seatList.querySelectorAll(':scope > li')).toHaveLength(2);
    expect(screen.getByRole('heading', { name: 'あなた', level: 2 })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'CPU1', level: 2 })).toBeInTheDocument();
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

  it.each([
    [1, /CPUの宣言予約中/],
    [2, /CPUのめくり予約中/],
  ])('shows the booked CPU action and approximate remaining time for kind %s', async (pendingKind, expected) => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    vi.setSystemTime(10_000);
    mockExec.mockResolvedValue(makeState({ pendingKind, pendingDeadlineMs: 12_400 }));
    renderWithProviders(<SnapPage />);
    const status = await screen.findByTestId('sp-pending');
    expect(status).toHaveTextContent(expected);
    expect(status).toHaveTextContent(/3秒/);
    if (pendingKind === 1) expect(status).toHaveTextContent(/期限前にスナップ/);
  });

  it('starts a new booking countdown from the current time', async () => {
    vi.useFakeTimers();
    vi.setSystemTime(10_000);
    mockExec.mockImplementation(async (command) =>
      makeState({
        pendingKind: command === 'step' ? 1 : 0,
        pendingDeadlineMs: command === 'step' ? Date.now() + 3_000 : 0,
      }),
    );
    renderWithProviders(<SnapPage />);
    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
      await Promise.resolve();
    });
    expect(mockExec).toHaveBeenCalledWith('reset');
    expect(screen.queryByTestId('sp-pending')).not.toBeInTheDocument();

    vi.setSystemTime(70_000);
    await act(async () => {
      fireEvent.click(screen.getByTestId('sp-step-btn'));
      await Promise.resolve();
      await Promise.resolve();
      await Promise.resolve();
    });

    const status = screen.getByTestId('sp-pending');
    expect(status).toHaveTextContent(/CPUの宣言予約中/);
    // Not /3秒/: that also matches the stale 63秒 this test guards against.
    expect(status).toHaveTextContent(/(^|\D)3秒/);
  });

  it('keeps the live announcement unchanged while the visible countdown ticks', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    vi.setSystemTime(10_000);
    mockExec.mockResolvedValue(makeState({ pendingKind: 2, pendingDeadlineMs: 12_400 }));
    renderWithProviders(<SnapPage />);
    const announcement = await screen.findByTestId('sp-pending-announcement');
    const startingText = announcement.textContent;
    expect(startingText).toMatch(/CPUのめくりが予約されました/);
    expect(screen.getByTestId('sp-pending')).toHaveTextContent(/3秒/);

    await act(async () => vi.advanceTimersByTime(1_000));

    expect(announcement).toHaveTextContent(startingText ?? '');
    expect(screen.getByTestId('sp-pending')).toHaveTextContent(/2秒/);
  });

  it('announces when a CPU booking starts', async () => {
    mockExec.mockImplementation(async (command) =>
      makeState({ pendingKind: command === 'step' ? 1 : 0, pendingDeadlineMs: Date.now() + 2_000 }),
    );
    renderWithProviders(<SnapPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByTestId('sp-pending-announcement')).toBeEmptyDOMElement();

    fireEvent.click(screen.getByTestId('sp-step-btn'));

    expect(await screen.findByTestId('sp-pending-announcement')).toHaveTextContent(/CPUの宣言が予約されました/);
  });

  it('clears the booked action display when the reservation is resolved', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    mockExec.mockImplementation(async (command) =>
      makeState({ pendingKind: command === 'reset' ? 1 : 0, pendingDeadlineMs: Date.now() + 1_000 }),
    );
    renderWithProviders(<SnapPage />);
    expect(await screen.findByTestId('sp-pending')).toHaveTextContent(/CPUの宣言予約中/);
    await act(async () => vi.advanceTimersByTime(150));
    expect(screen.queryByTestId('sp-pending')).not.toBeInTheDocument();
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

  // 札切れの席にだけ sp-out-{id} が出て、ストックを持つ席には出ない。
  it('shows out-of-cards badge only for zero-stock seats', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          { id: 0, isHuman: true, stockSize: 0 },
          { id: 1, isHuman: false, stockSize: 5 },
        ],
      }),
    );
    renderWithProviders(<SnapPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    // 札切れ席に sp-out-0 が出る
    const badge = await screen.findByTestId('sp-out-0');
    expect(badge).toBeInTheDocument();
    // 文言が実際の日本語「札切れ」に解決していること（i18n が機能していること）
    expect(badge).toHaveTextContent('札切れ');

    // ストックを持つ席 1 には出ない（否定コントロール）
    expect(screen.queryByTestId('sp-out-1')).not.toBeInTheDocument();
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
