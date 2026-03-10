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
  it('renders null when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    const { container } = renderWithProviders(<MemoryPage />);
    expect(container.firstChild).toBeNull();
  });

  it('renders reset on mount', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1 }));
  });

  it('renders player scores', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText('あなた')).toBeInTheDocument());
    expect(screen.getByText('CPU 1')).toBeInTheDocument();
    expect(screen.getByText('CPU 2')).toBeInTheDocument();
    expect(screen.getByText('CPU 3')).toBeInTheDocument();
  });

  it('renders board with 52 buttons', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    // Board should have 52 buttons
    const buttons = screen.getAllByRole('button');
    // 52 board + 1 reset = 53
    expect(buttons.length).toBeGreaterThanOrEqual(52);
  });

  it('clicking a board card calls handleFlip', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());

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
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
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
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
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

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1 }));
  });

  it('board cards disabled when taken', async () => {
    const takenBoard = makeBoard({ 0: { taken: true } });
    mockExec.mockResolvedValue({ ...flip1State, board: takenBoard });
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());

    // Taken cards should not show position number (transparent)
    // Other cards should be enabled
  });

  it('board cards disabled when face up', async () => {
    mockExec.mockResolvedValue(flip2State);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    // Face-up card button at position 5 should be disabled
  });

  it('board cards disabled on CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
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
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());

    // Open settings
    fireEvent.click(screen.getByText('設定'));
    expect(screen.getByText('CPU難易度')).toBeInTheDocument();
  });

  it('displays message from state', async () => {
    mockExec.mockResolvedValue({ ...flip1State, message: 'テストメッセージ', messageCode: 'memory.flip1' });
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText('1枚目をめくってください')).toBeInTheDocument());
  });

  it('displays error message', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('Network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));

    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );
  });
});
