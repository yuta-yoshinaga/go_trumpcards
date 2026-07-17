import { fireEvent, screen, waitFor } from '@testing-library/react';
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
    // Cards now render as real images with localized alt text (suit + rank).
    await waitFor(() => expect(screen.getByAltText('♠ K')).toBeInTheDocument());
    expect(screen.getByAltText('♥ A')).toBeInTheDocument();
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
    const wonBadge = await screen.findByTestId('piquet-meld-badge');
    // Human (elder, idx 0) was scoredBy:0 above → won → success palette.
    expect(wonBadge).toHaveClass('text-ds-success', 'border-ds-success');
  });

  it('renders the translated per-player stats line', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<PiquetPage />);
    // ja: "手札: 12 | トリック: 0 | ラウンド: 0 | マッチ: 0" (PlayerCard.playerStats)
    await waitFor(() => expect(screen.getAllByText(/手札: 12/).length).toBeGreaterThan(0));
  });

  it('renders the declaration list with translated tied and fallback-scorer labels', async () => {
    const claim = { length: 0, topRank: 0, pipTotal: 0, suit: 0, cards: [] };
    mockExec.mockResolvedValue(
      makeState({
        phase: PiquetPhase.DECLARATION,
        declStage: PiquetDeclarationKind.SET,
        declResults: [
          {
            kind: PiquetDeclarationKind.POINT,
            elderClaim: claim,
            youngerClaim: claim,
            winner: -1,
            scoredBy: -1,
            score: 0,
          },
          {
            kind: PiquetDeclarationKind.SEQUENCE,
            elderClaim: claim,
            youngerClaim: claim,
            winner: -1,
            scoredBy: 5, // neither elder (0) nor younger (1) → "?" fallback label
            score: 3,
          },
        ],
      }),
    );
    renderWithProviders(<PiquetPage />);
    await waitFor(() => expect(screen.getByText(/引き分け/)).toBeInTheDocument()); // declTied
    expect(screen.getByText(/\? \+3/)).toBeInTheDocument(); // declScored with "?" scorer
  });

  it('exposes the declaration list as an additions-only live log for screen readers', async () => {
    const claim = { length: 0, topRank: 0, pipTotal: 0, suit: 0, cards: [] };
    mockExec.mockResolvedValue(
      makeState({
        phase: PiquetPhase.DECLARATION,
        declStage: PiquetDeclarationKind.SEQUENCE,
        declResults: [
          {
            kind: PiquetDeclarationKind.POINT,
            elderClaim: claim,
            youngerClaim: claim,
            winner: 0,
            scoredBy: 0,
            score: 4,
          },
        ],
      }),
    );
    renderWithProviders(<PiquetPage />);
    const log = await screen.findByTestId('piquet-declaration-list');
    // role="log" (default aria-atomic="false") announces only newly-appended
    // results, not the whole list on every update.
    expect(log).toHaveAttribute('role', 'log');
    expect(log).toHaveAttribute('aria-live', 'polite');
  });

  it('renders the translated trick header during the play phase', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: PiquetPhase.PLAY,
        currentTrick: [{ playerIdx: 0, card: { design: 'SPADE', value: 13 } }],
      }),
    );
    renderWithProviders(<PiquetPage />);
    // Exact match isolates the TrickView header from the "トリック: 0" stats line.
    await waitFor(() => expect(screen.getByText('トリック')).toBeInTheDocument()); // trickHeader
    expect(screen.getByText('P0')).toBeInTheDocument(); // TrickView player label (cards now render as images)
  });

  it('shows the meld badge in the lost palette when the opponent scores', async () => {
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
            winner: 1,
            scoredBy: 1,
            score: 2,
          },
        ],
      }),
    );
    renderWithProviders(<PiquetPage />);
    const lostBadge = await screen.findByTestId('piquet-meld-badge');
    // scoredBy:1 (younger) → human lost → error palette (border signal, readable text).
    expect(lostBadge).toHaveClass('text-ds-text-primary', 'border-ds-error');
  });

  it('shows the hint button while the human can act and dispatches the hint command', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<PiquetPage />);
    const hintBtn = await screen.findByRole('button', { name: 'ヒント' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(makeState());
    fireEvent.click(hintBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('h'));
  });

  it('renders the play hint text when a card suggestion is present', async () => {
    mockExec.mockResolvedValue(makeState({ phase: PiquetPhase.PLAY, hint: { cardIndex: 3, reason: 'lowest' } }));
    renderWithProviders(<PiquetPage />);
    const hint = await screen.findByTestId('piquet-hint');
    expect(hint).toHaveTextContent('[3]');
  });
});
