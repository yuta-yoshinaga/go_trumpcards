import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { presidentApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, PresidentResponse } from '../types/card';
import { PresidentPage } from './PresidentPage';

vi.mock('../api/gameApi', () => ({
  presidentApi: { exec: vi.fn() },
  actionLogApi: { president: vi.fn() },
}));

const mockExec = vi.mocked(presidentApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<PresidentResponse> = {}): PresidentResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        isFinished: false,
        rank: -1,
        cardCount: 3,
        cards: [card('SPADE', 3), card('HEART', 5), card('DIAMOND', 7)],
      },
      { id: 1, isHuman: false, isFinished: false, rank: -1, cardCount: 13, cards: [] },
      { id: 2, isHuman: false, isFinished: false, rank: -1, cardCount: 13, cards: [] },
      { id: 3, isHuman: false, isFinished: false, rank: -1, cardCount: 13, cards: [] },
    ],
    currentTurn: 0,
    tableCards: [],
    lastPlayPlayerIdx: -1,
    gameEndFlag: false,
    revolutionActive: false,
    config: {
      revolutionEnabled: true,
      cardExchangeEnabled: true,
      passFieldFlushEnabled: true,
      cpuDifficulty: 1,
    },
    exchangeActions: [],
    cpuActions: [],
    humanAction: null,
    message: '',
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('PresidentPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders the human hand', async () => {
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('hand-card-1')).toBeInTheDocument();
    expect(screen.getByTestId('hand-card-2')).toBeInTheDocument();
  });

  it('enables Play button when a card is selected', async () => {
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    const playButton = screen.getByTestId('play-button');
    expect(playButton).toBeDisabled();

    fireEvent.click(screen.getByTestId('hand-card-0'));
    await waitFor(() => expect(screen.getByTestId('play-button')).not.toBeDisabled());
  });

  it('calls play on Play click', async () => {
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('play-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', [0]));
  });

  it('calls pass on Pass click', async () => {
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByTestId('pass-button')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('pass-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', []));
  });

  it('shows revolution banner when active', async () => {
    mockExec.mockResolvedValue(makeState({ revolutionActive: true }));
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByText(/革命中/)).toBeInTheDocument());
  });

  it('disables action buttons when it is not human turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentTurn: 1 }));
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByTestId('pass-button')).toBeDisabled());
    expect(screen.getByTestId('play-button')).toBeDisabled();
  });
});
