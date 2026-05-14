import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { seahaventowersApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, SeahavenTowersResponse } from '../types/card';
import { SeahavenTowersPage } from './SeahavenTowersPage';

vi.mock('../api/gameApi', () => ({
  seahaventowersApi: { exec: vi.fn() },
  actionLogApi: { seahaventowers: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(seahaventowersApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: SeahavenTowersResponse = {
  tableau: [[card('SPADE', 13)], [card('SPADE', 12)], [], [], [], [], [], [], [], []],
  reservedCells: [null, null],
  foundation: [[], [], [], []],
  phase: 0,
  moveCount: 5,
  canUndo: true,
  isStalemate: false,
  message: '',
  messageCode: 'seahaventowers.playing',
};

const gameClearState: SeahavenTowersResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'seahaventowers.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: SeahavenTowersResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'seahaventowers.gameOver',
};

const withFoundationState: SeahavenTowersResponse = {
  ...playingState,
  foundation: [[card('SPADE', 1)], [], [card('HEART', 1), card('HEART', 2)], []],
};

const withReservedCardState: SeahavenTowersResponse = {
  ...playingState,
  reservedCells: [card('DIAMOND', 7), null],
};

const withHintState: SeahavenTowersResponse = {
  ...playingState,
  hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'tableau', toCol: 2 },
};

beforeEach(() => {
  mockExec.mockResolvedValue(playingState);
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
});

describe('SeahavenTowersPage', () => {
  it('renders skeleton when state is null', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SeahavenTowersPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders empty tableau columns with K placeholder', async () => {
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    const kElements = screen.getAllByText('K');
    expect(kElements.length).toBeGreaterThanOrEqual(1);
  });

  it('renders foundation piles with all four suit symbols', async () => {
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    expect(screen.getByText('♣')).toBeInTheDocument();
    expect(screen.getByText('♥')).toBeInTheDocument();
    expect(screen.getByText('♦')).toBeInTheDocument();
  });

  it('renders empty foundation with A placeholder', async () => {
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    const aElements = screen.getAllByText('A');
    expect(aElements.length).toBeGreaterThanOrEqual(1);
  });

  it('renders foundation with cards', async () => {
    mockExec.mockResolvedValue(withFoundationState);
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    const imgs = screen.getAllByRole('img');
    expect(imgs.length).toBeGreaterThanOrEqual(1);
  });

  it('renders both reserved cells (empty)', async () => {
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    const emptyButtons = screen.getAllByText('空');
    expect(emptyButtons.length).toBe(2);
  });

  it('renders reserved cell with card occupied', async () => {
    mockExec.mockResolvedValue(withReservedCardState);
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(screen.getByAltText('♦ 7')).toBeInTheDocument();
    const emptyButtons = screen.getAllByText('空');
    expect(emptyButtons.length).toBe(1); // one remaining empty reserved cell
  });

  it('renders playing phase buttons', async () => {
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オートコンプリート' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument();
  });

  it('undo button disabled when canUndo is false', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: false });
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled());
  });

  it('hint button triggers hint API call', async () => {
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValueOnce(withHintState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('autocomplete button triggers autocomplete API call', async () => {
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'オートコンプリート' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValueOnce(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'オートコンプリート' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('renders game-clear state', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
  });

  it('renders game-over state', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームオーバー/).length).toBeGreaterThan(0));
  });
});
