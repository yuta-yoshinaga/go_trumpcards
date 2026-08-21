import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fourteenoutApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, FourteenOutBoardCell, FourteenOutResponse } from '../types/card';
import { FourteenOutPage } from './FourteenOutPage';

vi.mock('../api/gameApi', () => ({
  fourteenoutApi: { exec: vi.fn() },
  actionLogApi: { fourteenout: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(fourteenoutApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

/** 12 列の盤を作る。指定した列だけ中身を持つ。 */
function columns(...cols: Array<Array<[CardDesign, number]>>): FourteenOutBoardCell[][] {
  return Array.from({ length: 12 }, (_, i) => (cols[i] ?? []).map(([d, v]) => ({ card: card(d, v) })));
}

// 列0 の末尾 ♠9 と 列1 の末尾 ♥5 で 14。列2 の ♣4 はどちらとも組めない。
const playingState: FourteenOutResponse = {
  columns: columns([['SPADE', 9]], [['HEART', 5]], [['CLOVER', 4]]),
  phase: 0,
  removedCount: 0,
  removablePairs: 1,
  canUndo: false,
  isStalemate: false,
  message: '',
  messageCode: 'fourteenout.playing',
};

const gameClearState: FourteenOutResponse = {
  ...playingState,
  phase: 1,
  removedCount: 52,
  messageCode: 'fourteenout.gameClear',
  messageParams: { removedCount: '52' },
};

const gameOverState: FourteenOutResponse = {
  ...playingState,
  phase: 2,
  messageCode: 'fourteenout.gameOver',
};

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  mockExec.mockResolvedValue(playingState);
});

describe('FourteenOutPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<FourteenOutPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the removed count in the header', async () => {
    renderWithProviders(<FourteenOutPage />);
    await waitFor(() => expect(screen.getByText(/0\/52/)).toBeInTheDocument());
  });

  // **12 列。**クローン元は 5x5 の 25 マス。
  it('renders twelve columns', async () => {
    renderWithProviders(<FourteenOutPage />);
    await waitFor(() => expect(screen.getByTestId('mc-col-0')).toBeInTheDocument());
    for (let i = 0; i < 12; i++) {
      expect(screen.getByTestId(`mc-col-${i}`)).toBeInTheDocument();
    }
    expect(screen.queryByTestId('mc-col-12')).not.toBeInTheDocument();
  });

  it('disables a cleared column', async () => {
    renderWithProviders(<FourteenOutPage />);
    await waitFor(() => expect(screen.getByTestId('mc-col-0')).toBeEnabled());
    // 列3 以降は空。
    expect(screen.getByTestId('mc-col-3')).toBeDisabled();
  });

  it('selects a column, then deselects it on a second click', async () => {
    renderWithProviders(<FourteenOutPage />);
    const col0 = await screen.findByTestId('mc-col-0');

    fireEvent.click(col0);
    await waitFor(() => expect(col0).toHaveAttribute('aria-pressed', 'true'));
    fireEvent.click(col0);
    await waitFor(() => expect(col0).toHaveAttribute('aria-pressed', 'false'));
  });

  // **サーバに渡すのは列番号 2 つ。**クローン元は (行,列) x2 を渡す。
  it('sends the two column numbers on the second click', async () => {
    renderWithProviders(<FourteenOutPage />);
    fireEvent.click(await screen.findByTestId('mc-col-0'));

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('mc-col-1'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('remove', 0, 1));
  });

  // **合計 14 になる列だけを立てる。**クローン元は「隣接かつ同ランク」で立てる。
  it('rings the column that completes the pair and dims the rest', async () => {
    renderWithProviders(<FourteenOutPage />);
    fireEvent.click(await screen.findByTestId('mc-col-0'));

    await waitFor(() => expect(screen.getByTestId('mc-col-1')).toHaveAttribute('data-pair-match', 'true'));
    // ♣4 は ♠9 と 13 にしかならない。
    expect(screen.getByTestId('mc-col-2')).not.toHaveAttribute('data-pair-match');
    expect(screen.getByTestId('mc-col-2')).toHaveAttribute('data-dimmed', 'true');
  });

  it('shows the removable-pair counter from the current board', async () => {
    renderWithProviders(<FourteenOutPage />);
    await waitFor(() => expect(screen.getByTestId('mc-removable-count')).toBeInTheDocument());
    expect(screen.getByTestId('mc-removable-count').textContent).toContain('1');
  });

  // **0 は敗北の合図。**クローン元では補充を促す合図だった。
  it('warns when no pair remains', async () => {
    mockExec.mockResolvedValue({ ...playingState, columns: columns([['SPADE', 2]], [['HEART', 3]]) });
    renderWithProviders(<FourteenOutPage />);
    await waitFor(() =>
      expect(screen.getByTestId('mc-removable-count')).toHaveAttribute('data-removable-zero', 'true'),
    );
  });

  it('does not warn while a pair is available', async () => {
    renderWithProviders(<FourteenOutPage />);
    await waitFor(() => expect(screen.getByTestId('mc-removable-count')).toBeInTheDocument());
    expect(screen.getByTestId('mc-removable-count')).not.toHaveAttribute('data-removable-zero');
  });

  // **補充ボタンは存在しない。**押せても盤が変わらない無言の no-op になる。
  it('offers no deal button', async () => {
    renderWithProviders(<FourteenOutPage />);
    await waitFor(() => expect(screen.getByTestId('mc-col-0')).toBeInTheDocument());
    expect(screen.queryByTestId('mc-deal-button')).not.toBeInTheDocument();
  });

  it('rings the hint-suggested columns', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'remove-0-1', reason: 'hint.removePair', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    } as unknown as ReturnType<typeof useGameHint>);
    renderWithProviders(<FourteenOutPage />);
    await waitFor(() => expect(screen.getByTestId('mc-col-0').className).toContain('ring-ds-warning'));
    expect(screen.getByTestId('mc-col-1').className).toContain('ring-ds-warning');
    expect(screen.getByTestId('mc-col-2').className).not.toContain('ring-ds-warning');
  });

  it('undo is disabled when canUndo is false, live when true', async () => {
    renderWithProviders(<FourteenOutPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /元に戻す/ })).toBeDisabled());

    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    fireEvent.click(screen.getByTestId('mc-col-0'));
    fireEvent.click(screen.getByTestId('mc-col-1'));
    await waitFor(() => expect(screen.getByRole('button', { name: /元に戻す/ })).toBeEnabled());
  });

  it('shows game clear', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<FourteenOutPage />);
    await waitFor(() => expect(screen.getByText(/52\/52/)).toBeInTheDocument());
  });

  it('shows game over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<FourteenOutPage />);
    await waitFor(() => expect(screen.queryByTestId('mc-deal-button')).not.toBeInTheDocument());
    expect(screen.getByTestId('mc-col-0')).toBeDisabled();
  });

  it('shows an error alert when the API fails on mount', async () => {
    mockExec.mockRejectedValue(new Error('network error'));
    renderWithProviders(<FourteenOutPage />);
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });
});
