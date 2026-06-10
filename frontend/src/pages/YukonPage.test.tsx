import { screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { yukonApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, YukonResponse } from '../types/card';
import { YukonPage } from './YukonPage';

vi.mock('../api/gameApi', () => ({
  yukonApi: { exec: vi.fn() },
  actionLogApi: { yukon: vi.fn() },
}));

const mockExec = vi.mocked(yukonApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: YukonResponse = {
  tableau: [
    [{ card: card('SPADE', 13), faceUp: true }],
    [
      { card: null, faceUp: false },
      { card: card('HEART', 8), faceUp: true },
    ],
    [
      { card: null, faceUp: false },
      { card: null, faceUp: false },
      { card: card('CLOVER', 5), faceUp: true },
    ],
    [{ card: card('DIAMOND', 10), faceUp: true }],
    [{ card: card('SPADE', 3), faceUp: true }],
    [{ card: card('HEART', 7), faceUp: true }],
    [{ card: card('CLOVER', 2), faceUp: true }],
  ],
  foundation: [[], [], [], []],
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
  messageCode: 'yukon.playing',
};

const gameClearState: YukonResponse = {
  ...playingState,
  phase: 1,
  moveCount: 42,
  messageCode: 'yukon.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: YukonResponse = {
  ...playingState,
  phase: 2,
  messageCode: 'yukon.gameOver',
};

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(playingState);
});

describe('YukonPage', () => {
  it('renders heading', async () => {
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows move count', async () => {
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(screen.getByText(/手数/)).toBeInTheDocument());
  });

  it('shows game clear phase', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームクリア').length).toBeGreaterThan(0));
  });

  it('shows game over phase', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームオーバー').length).toBeGreaterThan(0));
  });

  it('hint button triggers hint command', async () => {
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: 'ヒント' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('embeds the hint in card aria-labels instead of a text panel', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromCol: 1, cardIndex: 1, toZone: 'tableau', toCol: 4 },
    });
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // The source (♥8) and target (♠3) cards carry the hint in their aria-labels.
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /ヒント: このカードを場札 4へ移動/ })).toBeInTheDocument(),
    );
    expect(screen.getByRole('button', { name: /ヒント: 移動先/ })).toBeInTheDocument();
    // The hint live region survives for screen readers but is visually hidden,
    // so it no longer squeezes the footer on mobile. It names the card since
    // "this card" is ambiguous when announced without focus context.
    const liveRegion = screen.getByRole('status');
    expect(liveRegion).toHaveClass('sr-only');
    expect(liveRegion).toHaveTextContent('♥ 8');
  });

  it('labels the hint source with the foundation destination', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 0 },
    });
    renderWithProviders(<YukonPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /ヒント: このカードを組札へ移動/ })).toBeInTheDocument(),
    );
  });

  it('autocomplete button triggers autocomplete command when all face-up', async () => {
    const readyState: YukonResponse = {
      ...playingState,
      tableau: [
        [{ card: card('SPADE', 13), faceUp: true }],
        [{ card: card('HEART', 8), faceUp: true }],
        [{ card: card('CLOVER', 5), faceUp: true }],
        [{ card: card('DIAMOND', 10), faceUp: true }],
        [{ card: card('SPADE', 3), faceUp: true }],
        [{ card: card('HEART', 7), faceUp: true }],
        [{ card: card('CLOVER', 2), faceUp: true }],
      ],
    };
    mockExec.mockResolvedValue(readyState);
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(readyState);
    const btn = screen.getByRole('button', { name: '自動完成' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('autocomplete button is disabled while face-down cards exist', async () => {
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeInTheDocument());
    expect(screen.getByTestId('autocomplete-button')).toBeDisabled();
  });

  it('giveup button triggers giveup command', async () => {
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: 'ギブアップ' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('shows error alert when API fails on mount', async () => {
    mockExec.mockRejectedValue(new Error('network error'));
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('hides game action buttons after game over (only reset remains)', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: /next game|次のゲーム/i }).length).toBeGreaterThan(0);
  });

  it('undo button fires undo command when canUndo is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const undoBtn = screen.getByRole('button', { name: '元に戻す' });
    undoBtn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('undo button is disabled when canUndo is false', async () => {
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const undoBtn = screen.getByRole('button', { name: '元に戻す' });
    expect(undoBtn).toBeDisabled();
  });

  it('renders empty tableau column placeholder', async () => {
    const stateWithEmpty = {
      ...playingState,
      tableau: [[], ...playingState.tableau.slice(1)],
    };
    mockExec.mockResolvedValue(stateWithEmpty);
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText('空')).toBeInTheDocument();
  });

  it('renders foundation suit labels', async () => {
    renderWithProviders(<YukonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText('♠')).toBeInTheDocument();
    expect(screen.getByText('♣')).toBeInTheDocument();
    expect(screen.getByText('♥')).toBeInTheDocument();
    expect(screen.getByText('♦')).toBeInTheDocument();
  });
});
