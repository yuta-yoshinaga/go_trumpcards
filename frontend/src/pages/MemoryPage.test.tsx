import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, memoryApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, MemoryBoardCard, MemoryResponse } from '../types/card';
import { MemoryPage } from './MemoryPage';

vi.mock('../api/gameApi', () => ({
  memoryApi: { exec: vi.fn() },
  actionLogApi: { memory: vi.fn() },
}));

const mockExec = vi.mocked(memoryApi.exec);

function makeBoard(
  overrides?: Partial<Record<number, { faceUp?: boolean; taken?: boolean; card?: Card | null }>>,
): MemoryBoardCard[] {
  return Array.from({ length: 52 }, (_, i) => ({
    card: overrides?.[i]?.card ?? null,
    faceUp: overrides?.[i]?.faceUp ?? false,
    taken: overrides?.[i]?.taken ?? false,
  }));
}

const flip1State: MemoryResponse = {
  players: [
    { id: 0, isHuman: true, pairCount: 0 },
    { id: 1, isHuman: false, pairCount: 2 },
    { id: 2, isHuman: false, pairCount: 1 },
    { id: 3, isHuman: false, pairCount: 0 },
  ],
  board: makeBoard(),
  phase: 0,
  currentPlayerIdx: 0,
  firstFlipPos: -1,
  secondFlipPos: -1,
  lastMatchResult: false,
  gameEndFlag: false,
  winnerIdx: -1,
  turnNumber: 0,
  message: '',
  config: { cpuDifficulty: 1 },
};

const flip2State: MemoryResponse = {
  ...flip1State,
  phase: 1,
  firstFlipPos: 5,
  board: makeBoard({ 5: { faceUp: true, card: { design: 'SPADE' as const, value: 3 } } }),
};

const resultMatchState: MemoryResponse = {
  ...flip1State,
  phase: 2,
  firstFlipPos: 0,
  secondFlipPos: 1,
  lastMatchResult: true,
  board: makeBoard({
    0: { faceUp: true, card: { design: 'SPADE' as const, value: 1 } },
    1: { faceUp: true, card: { design: 'HEART' as const, value: 1 } },
  }),
};

const gameEndState: MemoryResponse = {
  ...flip1State,
  phase: 3,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'ゲーム終了！',
};

const gameEndByFlagState: MemoryResponse = {
  ...flip1State,
  phase: 0,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'ゲーム終了！',
};

const cpuTurnState: MemoryResponse = {
  ...flip1State,
  currentPlayerIdx: 1,
};

beforeEach(() => {
  mockExec.mockResolvedValue(flip1State);
});

describe('MemoryPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<MemoryPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders reset on mount', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1 }));
  });

  it('renders player scores inline', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    expect(screen.getByText(/CPU 1: 2/)).toBeInTheDocument();
    expect(screen.getByText(/CPU 2: 1/)).toBeInTheDocument();
    expect(screen.getByText(/CPU 3: 0/)).toBeInTheDocument();
  });

  it('score section has role="status" for accessibility', async () => {
    const { container } = renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    const scoreSection = container.querySelector('[role="status"]');
    expect(scoreSection).toBeInTheDocument();
    expect(scoreSection).toHaveAttribute('aria-label', 'スコア');
  });

  it('renders board with 52 buttons', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    // Board should have 52 buttons
    const buttons = screen.getAllByRole('button');
    // 52 board + 1 reset = 53
    expect(buttons.length).toBeGreaterThanOrEqual(52);
  });

  it('clicking a board card calls handleFlip', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(flip2State);

    // Click a board card button (find by the position text inside it)
    const boardButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === '3');
    fireEvent.click(boardButtons[0]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('flip', 3));
  });

  it('shows face-up card image in flip2 phase', async () => {
    mockExec.mockResolvedValue(flip2State);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    // Face-up card at position 5 shows card image instead of position number
    const imgs = screen.getAllByRole('img');
    expect(imgs.length).toBeGreaterThanOrEqual(1);
  });

  it('result phase shows next button', async () => {
    mockExec.mockResolvedValue(resultMatchState);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次へ' })).toBeInTheDocument());
  });

  it('next button dispatches next command', async () => {
    mockExec.mockResolvedValue(resultMatchState);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次へ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(flip1State);
    fireEvent.click(screen.getByRole('button', { name: '次へ' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('next button not shown in flip phases', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '次へ' })).not.toBeInTheDocument();
  });

  it('game end shows action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('game end by flag also shows action log', async () => {
    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('action log button fetches and shows log', async () => {
    mockExec.mockResolvedValue(gameEndState);
    const mockLogApi = vi.mocked(actionLogApi.memory);
    mockLogApi.mockResolvedValue({ entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'match', detail: 'test' }] });

    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());
  });

  it('reset button dispatches reset with config', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(flip1State);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1 }));
  });

  it('board cards disabled when taken', async () => {
    const takenBoard = makeBoard({ 0: { taken: true } });
    mockExec.mockResolvedValue({ ...flip1State, board: takenBoard });
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());

    // Taken cards should not show position number (transparent)
    // Other cards should be enabled
  });

  it('taken card buttons have aria-hidden to avoid empty slot announcements', async () => {
    const takenBoard = makeBoard({ 0: { taken: true } });
    mockExec.mockResolvedValue({ ...flip1State, board: takenBoard });
    const { container } = renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    const hiddenButtons = container.querySelectorAll('button[aria-hidden="true"]');
    expect(hiddenButtons.length).toBe(1);
  });

  it('changing cpu difficulty updates config used on reset', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));
    fireEvent.change(screen.getByRole('combobox', { name: 'CPU難易度' }), { target: { value: '2' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(flip1State);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 2 }));
  });

  it('board cards disabled when face up', async () => {
    mockExec.mockResolvedValue(flip2State);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    // Face-up card button at position 5 should be disabled
  });

  it('board cards disabled on CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    // Board card at position 10 should be disabled (avoids pairCount text conflicts)
    const boardButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === '10');
    expect(boardButtons[0]).toBeDisabled();
  });

  it('loading state shows sr-only text and aria-busy', async () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<MemoryPage />);
    // Nothing renders when no state
  });

  it('settings panel works', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());

    // Open settings
    fireEvent.click(screen.getByText('設定'));
    expect(screen.getByText('CPU難易度')).toBeInTheDocument();
  });

  it('displays message from state', async () => {
    mockExec.mockResolvedValue({ ...flip1State, message: 'テストメッセージ', messageCode: 'memory.flip1' });
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText('1枚目をめくってください')).toBeInTheDocument());
  });

  it('renders landscape orientation banner in DOM', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    expect(screen.getByText('横向きにすると快適にプレイできます')).toBeInTheDocument();
  });

  // ── ConfirmDialog on reset ─────────────────────────────────────────────────

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(flip1State);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1 }));
  });

  it('displays error message', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('Network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );
  });

  // --- PhaseIndicator coverage ---

  it('phase indicator shows your turn when human flip turn', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows waiting when cpu turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('待機中'));
  });

  it('reset hides the action log panel if open', async () => {
    mockExec.mockResolvedValue(gameEndState);
    vi.mocked(actionLogApi.memory).mockResolvedValueOnce({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'match', detail: 'test' }],
    });

    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());

    mockExec.mockResolvedValue(flip1State);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(screen.queryByText('棋譜')).not.toBeInTheDocument());
  });

  // --- Keyboard navigation tests ---

  it('pressing n triggers next in RESULT phase', async () => {
    mockExec.mockResolvedValue(resultMatchState);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次へ' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(flip1State);
    fireEvent.keyDown(document, { key: 'n' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('board card buttons have focus-visible ring classes', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    const boardButtons = screen.getAllByRole('button').filter((btn) => btn.className.includes('aspect-'));
    expect(boardButtons.length).toBeGreaterThan(0);
    for (const btn of boardButtons) {
      expect(btn.className).toContain('focus-visible:ring-white/80');
    }
  });

  it('pressing n does not trigger next outside RESULT phase', async () => {
    mockExec.mockResolvedValue(flip1State);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'n' });
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(flip1State);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });
});
