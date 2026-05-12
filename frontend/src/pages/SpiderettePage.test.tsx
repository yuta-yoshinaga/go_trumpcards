import { screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { spideretteApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, SpideretteResponse, SpideretteTableauCard } from '../types/card';
import { SpiderettePage } from './SpiderettePage';

vi.mock('../api/gameApi', () => ({
  spideretteApi: { exec: vi.fn() },
  actionLogApi: { spiderette: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const playSoundMock = vi.fn();
vi.mock('../providers/SoundProvider', async () => {
  const actual = await vi.importActual<typeof import('../providers/SoundProvider')>('../providers/SoundProvider');
  return {
    ...actual,
    useSound: () => ({ playSound: playSoundMock, muted: false, toggleMute: vi.fn() }),
  };
});

const mockSend = vi.mocked(spideretteApi.exec);

function makeTableau(cols: SpideretteTableauCard[][]): SpideretteTableauCard[][] {
  const result: SpideretteTableauCard[][] = [];
  for (let i = 0; i < 7; i++) {
    result.push(cols[i] ?? []);
  }
  return result;
}

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: SpideretteResponse = {
  tableau: makeTableau([
    [{ card: card('SPADE', 13), faceUp: true }],
    [
      { card: null, faceUp: false },
      { card: card('HEART', 5), faceUp: true },
    ],
    [],
    [],
    [],
    [],
    [],
  ]),
  stockCount: 24,
  completedSuits: 0,
  score: 500,
  phase: 0,
  moveCount: 5,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: SpideretteResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'spiderette.gameClear',
  messageParams: { moveCount: '42', score: '500' },
};

beforeEach(() => {
  mockSend.mockResolvedValue(playingState);
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  playSoundMock.mockClear();
});

describe('SpiderettePage', () => {
  it('renders skeleton when no state', () => {
    mockSend.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SpiderettePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders stock count', async () => {
    renderWithProviders(<SpiderettePage />);
    await waitFor(() => expect(screen.getByText(/山札/)).toBeInTheDocument());
    expect(screen.getByText(/\(24\)/)).toBeInTheDocument();
  });

  it('renders move count', async () => {
    renderWithProviders(<SpiderettePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 5/));
  });

  it('renders completed suits 0/4', async () => {
    renderWithProviders(<SpiderettePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/完成: 0\/4/));
  });

  it('shows game clear phase label', async () => {
    mockSend.mockResolvedValue(gameClearState);
    renderWithProviders(<SpiderettePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
  });
});
