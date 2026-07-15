import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { basraApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeBasraState } from '../test/stateFactories';
import { BasraPage } from './BasraPage';

vi.mock('../api/gameApi', () => ({
  basraApi: { exec: vi.fn() },
  actionLogApi: { basra: vi.fn() },
}));

const mockExec = vi.mocked(basraApi.exec);

const playPhaseState = makeBasraState();
const cpuTurnState = makeBasraState({ currentTurn: 1, isHumanTurn: false });
const gameEndState = makeBasraState({
  phase: 1,
  gameEndFlag: true,
  winners: [0],
  players: [
    { id: 0, isHuman: true, cardCount: 0, cards: [], capturedCount: 30, basraCount: 2, score: 27 },
    { id: 1, isHuman: false, cardCount: 0, cards: [], capturedCount: 10, basraCount: 0, score: 5 },
    { id: 2, isHuman: false, cardCount: 0, cards: [], capturedCount: 8, basraCount: 0, score: 4 },
    { id: 3, isHuman: false, cardCount: 0, cards: [], capturedCount: 4, basraCount: 0, score: 1 },
  ],
  lastDealDetail: {
    cards: { 0: 30, 1: 10, 2: 8, 3: 4 },
    aces: { 0: 2, 1: 1, 2: 1, 3: 0 },
    basras: { 0: 2, 1: 0, 2: 0, 3: 0 },
    hasSevenDiamonds: 0,
    hasTenDiamonds: 0,
    mostCards: 0,
    gained: { 0: 27, 1: 5, 2: 4, 3: 1 },
  },
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('BasraPage', () => {
  it('renders skeleton fallback when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<BasraPage />);
    // The skeleton fallback marks its container aria-busy.
    expect(document.querySelector('[aria-busy]')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<BasraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders the play phase with table and hand cards', async () => {
    renderWithProviders(<BasraPage />);
    await waitFor(() => {
      expect(screen.getByTestId('hand-card-0')).toBeInTheDocument();
      expect(screen.getByTestId('table-card-0')).toBeInTheDocument();
    });
  });

  it('capturing dispatches play with the selected hand and table indices', async () => {
    renderWithProviders(<BasraPage />);
    const handCard = await screen.findByTestId('hand-card-0');
    fireEvent.click(handCard);
    fireEvent.click(screen.getByTestId('table-card-0'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '捕獲' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0, tableIndices: [0] }));
  });

  it('marks capture candidates and selected table cards with shape + aria (not colour)', async () => {
    renderWithProviders(<BasraPage />);
    const table0 = await screen.findByTestId('table-card-0');
    // Before selecting a hand card, table-card-0 is just its name.
    expect(table0).toHaveAttribute('aria-label', '♠ 5');
    // Selecting hand card 0 makes table card 0 a capture candidate (✓ badge + aria).
    fireEvent.click(screen.getByTestId('hand-card-0'));
    expect(screen.getByTestId('table-card-0')).toHaveAttribute('aria-label', '♠ 5、キャプチャ可能');
    expect(screen.getByTestId('table-card-0')).toHaveTextContent('✓');
    // Selecting the table card flips to selected (● badge + aria-pressed).
    fireEvent.click(screen.getByTestId('table-card-0'));
    expect(screen.getByTestId('table-card-0')).toHaveAttribute('aria-label', '♠ 5、選択済み');
    expect(screen.getByTestId('table-card-0')).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByTestId('table-card-0')).toHaveTextContent('●');
  });

  it('trailing (no table cards) dispatches play with an empty capture set', async () => {
    renderWithProviders(<BasraPage />);
    const handCard = await screen.findByTestId('hand-card-1');
    fireEvent.click(handCard);
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 1, tableIndices: [] }));
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<BasraPage />);
    await waitFor(() => expect(screen.getByTestId('basra-prompt')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('renders the game-end result with final scores and the new-game button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<BasraPage />);
    await waitFor(() => expect(screen.getByTestId('basra-result')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '新しいゲーム' })).toBeInTheDocument();
  });

  it('clicking new game dispatches nextround', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<BasraPage />);
    const nextBtn = await screen.findByRole('button', { name: '新しいゲーム' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(gameEndState);
    fireEvent.click(nextBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });
});
