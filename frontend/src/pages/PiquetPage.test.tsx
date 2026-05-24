import { screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { piquetApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { PiquetResponse } from '../types/card';
import { PiquetDeclarationKind, PiquetExchangeTurn, PiquetPhase } from '../types/phases';
import { PiquetPage } from './PiquetPage';

vi.mock('../api/gameApi', () => ({
  piquetApi: { exec: vi.fn() },
  actionLogApi: { piquet: vi.fn() },
}));

const mockExec = vi.mocked(piquetApi.exec);

function makeState(overrides: Partial<PiquetResponse> = {}): PiquetResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 12,
        cards: [
          { design: 'SPADE', value: 13 },
          { design: 'HEART', value: 1 },
        ],
        trickCount: 0,
        declScore: 0,
        trickScore: 0,
        bonusScore: 0,
        roundScore: 0,
        matchScore: 0,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 12,
        cards: [],
        trickCount: 0,
        declScore: 0,
        trickScore: 0,
        bonusScore: 0,
        roundScore: 0,
        matchScore: 0,
      },
    ],
    phase: PiquetPhase.EXCHANGE,
    dealNumber: 1,
    dealsPerPartie: 6,
    elderIdx: 0,
    youngerIdx: 1,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    trickNumber: 0,
    tricksWon: [0, 0],
    exchangeTurn: PiquetExchangeTurn.ELDER,
    elderExchangedCnt: 0,
    youngerExchangedCnt: 0,
    elderTalon: [],
    youngerTalon: [],
    elderRevealedTalon: [],
    youngerRevealedTalon: [],
    carteBlanche: [false, false],
    declStage: PiquetDeclarationKind.POINT,
    declResults: [],
    currentTrick: [],
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: { cpuDifficulty: 1, dealsPerPartie: 6 },
    ...overrides,
  };
}

describe('PiquetPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<PiquetPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders deal info', async () => {
    mockExec.mockResolvedValue(makeState({ dealNumber: 2 }));
    renderWithProviders(<PiquetPage />);
    // Look for "2" + "/" + "6" pattern from the deal header
    await waitFor(() => expect(screen.getByText(/2.*\/.*6|ディール/i)).toBeInTheDocument());
  });

  it('renders elder + younger labels', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<PiquetPage />);
    await waitFor(() => expect(screen.getAllByText(/Elder|Younger/).length).toBeGreaterThan(0));
  });

  it('shows the human hand cards', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<PiquetPage />);
    await waitFor(() => expect(screen.getByText('K♠')).toBeInTheDocument());
    expect(screen.getByText('A♥')).toBeInTheDocument();
  });

  it('renders reset button', async () => {
    mockExec.mockResolvedValue(makeState({ phase: PiquetPhase.DECLARATION }));
    renderWithProviders(<PiquetPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /リセット|Reset/i })).toBeInTheDocument());
  });

  it('highlights human meld badge when a new declaration result arrives', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: PiquetPhase.DECLARATION,
        declStage: PiquetDeclarationKind.SEQUENCE,
        declResults: [
          {
            kind: PiquetDeclarationKind.POINT,
            elderClaim: {
              length: 2,
              topRank: 13,
              pipTotal: 21,
              suit: 0,
              cards: [
                { design: 'SPADE', value: 13 },
                { design: 'HEART', value: 1 },
              ],
            },
            youngerClaim: { length: 0, topRank: 0, pipTotal: 0, suit: 0, cards: [] },
            winner: 0,
            scoredBy: 0,
            score: 2,
          },
        ],
      }),
    );
    renderWithProviders(<PiquetPage />);
    await waitFor(() => expect(screen.getByTestId('piquet-meld-badge')).toBeInTheDocument());
  });
});
