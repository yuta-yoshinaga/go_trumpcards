import { screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { spiteAndMaliceApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, SpiteAndMaliceResponse } from '../types/card';
import { SpiteAndMalicePage } from './SpiteAndMalicePage';

vi.mock('../api/gameApi', () => ({
  spiteAndMaliceApi: { exec: vi.fn() },
  actionLogApi: { spiteandmalice: vi.fn() },
}));

const mockExec = vi.mocked(spiteAndMaliceApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const baseState: SpiteAndMaliceResponse = {
  phase: 0,
  current: 0,
  players: [
    {
      hand: [card('SPADE', 5), card('HEART', 8), card('DIAMOND', 11), card('CLOVER', 13), card('SPADE', 2)],
      goalTop: card('HEART', 9),
      goalSize: 20,
      sides: [[], [], [], []],
      isCpu: false,
    },
    {
      hand: [],
      goalTop: card('CLOVER', 7),
      goalSize: 20,
      sides: [[], [], [], []],
      isCpu: true,
    },
  ],
  foundations: [[], [], [], []],
  foundationTops: [0, 0, 0, 0],
  stockSize: 60,
  completedSize: 0,
  moveCount: 0,
  winner: -1,
  goalSize: 20,
  cpuDifficulty: 1,
  message: '',
  messageCode: 'spiteandmalice.playing',
};

const cpuTurnState: SpiteAndMaliceResponse = { ...baseState, current: 1 };

const winState: SpiteAndMaliceResponse = {
  ...baseState,
  phase: 1,
  winner: 0,
  moveCount: 42,
  messageCode: 'spiteandmalice.win',
  messageParams: { moveCount: '42' },
};

const loseState: SpiteAndMaliceResponse = {
  ...baseState,
  phase: 1,
  winner: 1,
  moveCount: 42,
  messageCode: 'spiteandmalice.lose',
  messageParams: { moveCount: '42' },
};

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  mockExec.mockResolvedValue(baseState);
});

describe('SpiteAndMalicePage', () => {
  it('renders heading', async () => {
    renderWithProviders(<SpiteAndMalicePage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<SpiteAndMalicePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows move count label', async () => {
    renderWithProviders(<SpiteAndMalicePage />);
    await waitFor(() => expect(screen.getByText(/手数|Moves/)).toBeInTheDocument());
  });

  it('drives CPU turn automatically', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SpiteAndMalicePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('cpu'), { timeout: 2000 });
  });

  it('shows win phase label', async () => {
    mockExec.mockResolvedValue(winState);
    renderWithProviders(<SpiteAndMalicePage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームオーバー|Game Over/).length).toBeGreaterThan(0));
  });

  it('shows lose phase label', async () => {
    mockExec.mockResolvedValue(loseState);
    renderWithProviders(<SpiteAndMalicePage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームオーバー|Game Over/).length).toBeGreaterThan(0));
  });
});
