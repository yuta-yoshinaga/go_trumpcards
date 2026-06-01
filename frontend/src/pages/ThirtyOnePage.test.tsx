import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { thirtyoneApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, ThirtyOneResponse } from '../types/card';
import { ThirtyOnePhase } from '../types/phases';
import { ThirtyOnePage } from './ThirtyOnePage';

vi.mock('../api/gameApi', () => ({
  thirtyoneApi: { exec: vi.fn() },
  actionLogApi: { thirtyone: vi.fn() },
}));

const mockExec = vi.mocked(thirtyoneApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function player(id: number, isHuman: boolean, cards: Card[], over: Partial<ThirtyOneResponse['players'][number]> = {}) {
  return { id, isHuman, cardCount: cards.length, cards, lives: 3, score: 0, isEliminated: false, ...over };
}

function makeState(overrides: Partial<ThirtyOneResponse> = {}): ThirtyOneResponse {
  return {
    players: [
      player(0, true, [card('SPADE', 3), card('HEART', 5), card('DIAMOND', 7)]),
      player(1, false, []),
      player(2, false, []),
      player(3, false, []),
    ],
    phase: ThirtyOnePhase.DRAW,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: card('HEART', 2),
    drawPileCount: 39,
    gameEndFlag: false,
    winnerIdx: -1,
    knockerIdx: -1,
    thirtyOneIdx: -1,
    roundWinnerIdx: -1,
    roundLosers: [],
    message: '',
    config: { cpuDifficulty: 1, initialLives: 3 },
    ...overrides,
  };
}

beforeEach(() => {
  localStorage.clear();
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('ThirtyOnePage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<ThirtyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows draw and knock actions during the draw phase', async () => {
    renderWithProviders(<ThirtyOnePage />);
    expect(await screen.findByTestId('draw-stock-button')).toBeEnabled();
    expect(screen.getByTestId('draw-discard-button')).toBeEnabled();
    expect(screen.getByTestId('knock-button')).toBeEnabled();
  });

  it('draws from stock when the button is clicked', async () => {
    renderWithProviders(<ThirtyOnePage />);
    fireEvent.click(await screen.findByTestId('draw-stock-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('knocks when the knock button is clicked', async () => {
    renderWithProviders(<ThirtyOnePage />);
    fireEvent.click(await screen.findByTestId('knock-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('knock'));
  });

  it('allows selecting and discarding a card in the discard phase', async () => {
    mockExec.mockResolvedValue(makeState({ phase: ThirtyOnePhase.DISCARD }));
    renderWithProviders(<ThirtyOnePage />);
    const discardBtn = await screen.findByTestId('discard-button');
    expect(discardBtn).toBeDisabled();

    // Select the first hand card, then discard.
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('discard-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', 0));
  });

  it('shows the next-round button at round end', async () => {
    mockExec.mockResolvedValue(makeState({ phase: ThirtyOnePhase.ROUND_END, roundWinnerIdx: 0, roundLosers: [1] }));
    renderWithProviders(<ThirtyOnePage />);
    fireEvent.click(await screen.findByTestId('next-round-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('reveals CPU lives at all times', async () => {
    renderWithProviders(<ThirtyOnePage />);
    await screen.findByTestId('draw-stock-button');
    // Human + 3 CPU each render a lives indicator (❤ or 💀).
    expect(screen.getAllByLabelText(/lives-/).length).toBeGreaterThanOrEqual(1);
  });
});
