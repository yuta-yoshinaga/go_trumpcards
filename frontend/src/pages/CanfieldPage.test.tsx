import { screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { canfieldApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CanfieldResponse, Card, CardDesign } from '../types/card';
import { CanfieldPage } from './CanfieldPage';

vi.mock('../api/gameApi', () => ({
  canfieldApi: { exec: vi.fn() },
  actionLogApi: { canfield: vi.fn() },
}));

const mockExec = vi.mocked(canfieldApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: CanfieldResponse = {
  tableau: [
    [{ card: card('SPADE', 7) }],
    [{ card: card('HEART', 8) }],
    [{ card: card('CLOVER', 9) }],
    [{ card: card('DIAMOND', 10) }],
  ],
  reserve: [card('SPADE', 3)],
  stockCount: 34,
  waste: [card('HEART', 4)],
  foundation: [[card('SPADE', 5)], [], [], []],
  baseRank: 5,
  phase: 0,
  moveCount: 0,
  canUndo: false,
  message: '',
  messageCode: 'canfield.playing',
};

const gameClearState: CanfieldResponse = {
  ...playingState,
  phase: 1,
  moveCount: 42,
  messageCode: 'canfield.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: CanfieldResponse = {
  ...playingState,
  phase: 2,
  messageCode: 'canfield.gameOver',
};

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(playingState);
});

describe('CanfieldPage', () => {
  it('renders heading', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows base rank', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(screen.getByText(/ベースランク/)).toBeInTheDocument());
  });

  it('shows game clear phase', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームクリア').length).toBeGreaterThan(0));
  });

  it('shows game over phase', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームオーバー').length).toBeGreaterThan(0));
  });

  it('clicks draw button (stock) to fire draw command', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const stockButton = screen.getByRole('button', { name: /山札/ });
    stockButton.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('moves reserve to foundation', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: /リザーブ→組札/ });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'reserve' }, { zone: 'foundation' }));
  });

  it('moves waste to foundation', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: /ウェイスト→組札/ });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'foundation' }));
  });

  it('hint button triggers hint command', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: 'ヒント' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('autocomplete button triggers autocomplete command', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: '自動完成' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('giveup button triggers giveup command', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: 'ギブアップ' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('shows error alert when API fails on mount', async () => {
    mockExec.mockRejectedValue(new Error('network error'));
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('hides game action buttons after game over (only reset remains)', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: /reset|リセット/i }).length).toBeGreaterThan(0);
  });

  it('undo button fires undo command when canUndo is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const undoBtn = screen.getByRole('button', { name: '元に戻す' });
    undoBtn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('undo button is disabled when canUndo is false', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const undoBtn = screen.getByRole('button', { name: '元に戻す' });
    expect(undoBtn).toBeDisabled();
  });

  it('renders empty waste and reserve placeholders', async () => {
    mockExec.mockResolvedValue({ ...playingState, waste: [], reserve: [] });
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // The reserve/waste sections still render without their top cards
    expect(screen.getByText(/リザーブ: 0/)).toBeInTheDocument();
  });

  it('tableau-to-tableau button fires move command with correct args', async () => {
    renderWithProviders(<CanfieldPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // Column 0 → column 1 button (each non-empty column has 3 →Tj buttons)
    const btn = screen.getAllByRole('button', { name: /→T1/ })[0];
    btn.click();
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'move',
        { zone: 'tableau', col: 0, cardIndex: 0 },
        { zone: 'tableau', col: 1 },
      ),
    );
  });
});
