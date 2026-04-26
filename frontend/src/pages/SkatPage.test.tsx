import { screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../test/renderWithProviders';
import { SkatPage } from './SkatPage';

vi.mock('../api/gameApi', async () => {
  const actual = await vi.importActual<typeof import('../api/gameApi')>('../api/gameApi');
  const sampleResponse = {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 10,
        cards: [],
        bid: 0,
        isDeclarer: false,
        cardPoints: 0,
        roundsWon: 0,
        roundsLost: 0,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 10,
        cards: [],
        bid: 0,
        isDeclarer: false,
        cardPoints: 0,
        roundsWon: 0,
        roundsLost: 0,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
      {
        id: 2,
        isHuman: false,
        cardCount: 10,
        cards: [],
        bid: 0,
        isDeclarer: false,
        cardPoints: 0,
        roundsWon: 0,
        roundsLost: 0,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
    ],
    phase: 0,
    roundNumber: 1,
    trickNumber: 0,
    currentPlayerIdx: -1,
    currentTrick: [],
    forehandIdx: 1,
    middlehandIdx: 2,
    rearhandIdx: 0,
    dealerIdx: 0,
    declarerIdx: -1,
    currentBid: 0,
    activeBidActorIdx: 0,
    gameType: 0,
    trumpSuit: 0,
    pickedSkat: false,
    declarerCardPoints: 0,
    defendersCardPoints: 0,
    winnerSide: -1,
    gameValue: 0,
    gameEndFlag: false,
    leadPlayerIdx: -1,
    message: '',
    config: { cpuDifficulty: 1, targetScore: 500 },
  };
  return {
    ...actual,
    skatApi: { exec: vi.fn().mockResolvedValue(sampleResponse) },
  };
});

describe('SkatPage', () => {
  it('renders the round info header after loading state', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => {
      expect(screen.getAllByText(/ラウンド|Round|スカート/i).length).toBeGreaterThan(0);
    });
  });
});
