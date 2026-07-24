import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { briscolaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BriscolaResponse, Card } from '../types/card';
import { BriscolaPage } from './BriscolaPage';

vi.mock('../api/gameApi', () => ({
  briscolaApi: { exec: vi.fn() },
  actionLogApi: { briscola: vi.fn() },
}));

const mockExec = vi.mocked(briscolaApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<BriscolaResponse> = {}): BriscolaResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 3,
        cards: [card('SPADE', 1), card('HEART', 5), card('DIAMOND', 11)],
        points: 0,
        trickCount: 0,
      },
      { id: 1, isHuman: false, cardCount: 3, cards: [], points: 0, trickCount: 0 },
    ],
    phase: 0,
    trickNumber: 1,
    currentPlayerIdx: 0,
    currentTrick: [],
    trumpSuit: 1,
    trumpCard: card('SPADE', 13),
    dealerIdx: 0,
    leadPlayerIdx: 0,
    stockRemaining: 33,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: { cpuDifficulty: 0 },
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockResolvedValue(makeState());
});

describe('BriscolaPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<BriscolaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders header info (trick, stock, points)', async () => {
    renderWithProviders(<BriscolaPage />);
    await waitFor(() => expect(screen.getByText(/トリック: 1/)).toBeInTheDocument());
    expect(screen.getByText(/山札: 33/)).toBeInTheDocument();
    expect(screen.getByText(/得点/)).toBeInTheDocument();
  });

  it('exposes the tutorial target elements for the guided tour', async () => {
    // A card on the table so the trick area (conditionally rendered) is present.
    mockExec.mockResolvedValue(makeState({ currentTrick: [{ playerIdx: 1, card: card('CLOVER', 4) }] }));
    const { container } = renderWithProviders(<BriscolaPage />);
    await waitFor(() => expect(screen.getByText(/トリック: 1/)).toBeInTheDocument());
    for (const target of ['briscola-trump', 'briscola-trick', 'briscola-hand', 'briscola-score']) {
      expect(container.querySelector(`[data-tutorial="${target}"]`)).not.toBeNull();
    }
  });

  it('shows human hand as 3 play buttons', async () => {
    renderWithProviders(<BriscolaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ A を出す' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '♥ 5 を出す' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♦ J を出す' })).toBeInTheDocument();
  });

  it('renders hand and trump cards with the AnimatedCard component', async () => {
    renderWithProviders(<BriscolaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ A を出す' })).toBeInTheDocument());
    // 3 hand cards + 1 face-up trump card are rendered as animated cards.
    expect(screen.getAllByTestId('animated-card')).toHaveLength(4);
  });

  it('fires play with the selected card index when a card is clicked', async () => {
    renderWithProviders(<BriscolaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♥ 5 を出す' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '♥ 5 を出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1));
  });

  it('keyboard: a number key highlights a card and Enter plays it', async () => {
    renderWithProviders(<BriscolaPage />);
    const secondCard = await screen.findByRole('button', { name: '♥ 5 を出す' });
    mockExec.mockClear();
    // "2" highlights the second hand card (index 1) without playing it.
    fireEvent.keyDown(document.body, { key: '2' });
    await waitFor(() => expect(secondCard).toHaveAttribute('aria-pressed', 'true'));
    expect(mockExec).not.toHaveBeenCalled();
    // Enter plays the highlighted card.
    fireEvent.keyDown(document.body, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1));
  });

  it('keyboard: Escape clears the highlight and Enter then plays nothing', async () => {
    renderWithProviders(<BriscolaPage />);
    const firstCard = await screen.findByRole('button', { name: '♠ A を出す' });
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: '1' });
    await waitFor(() => expect(firstCard).toHaveAttribute('aria-pressed', 'true'));
    fireEvent.keyDown(document.body, { key: 'Escape' });
    await waitFor(() => expect(firstCard).toHaveAttribute('aria-pressed', 'false'));
    fireEvent.keyDown(document.body, { key: 'Enter' });
    expect(mockExec).not.toHaveBeenCalledWith('play', expect.anything());
  });

  it('keyboard: number keys do nothing on a CPU turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
    renderWithProviders(<BriscolaPage />);
    const firstCard = await screen.findByRole('button', { name: '♠ A を出す' });
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: '1' });
    fireEvent.keyDown(document.body, { key: 'Enter' });
    expect(firstCard).toHaveAttribute('aria-pressed', 'false');
    expect(mockExec).not.toHaveBeenCalledWith('play', expect.anything());
  });

  it('shows "Next trick" button on trick-end and dispatches next', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1 }));
    renderWithProviders(<BriscolaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリックへ' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のトリックへ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('shows youWin banner when winnerIdx is 0', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        gameEndFlag: true,
        winnerIdx: 0,
        players: [
          { id: 0, isHuman: true, cardCount: 0, cards: [], points: 70, trickCount: 10 },
          { id: 1, isHuman: false, cardCount: 0, cards: [], points: 50, trickCount: 10 },
        ],
      }),
    );
    renderWithProviders(<BriscolaPage />);
    await waitFor(() => expect(screen.getByText(/あなたの勝ち！.*70.*50/)).toBeInTheDocument());
  });

  it('shows cpuWin banner when winnerIdx is 1', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        gameEndFlag: true,
        winnerIdx: 1,
        players: [
          { id: 0, isHuman: true, cardCount: 0, cards: [], points: 50, trickCount: 10 },
          { id: 1, isHuman: false, cardCount: 0, cards: [], points: 70, trickCount: 10 },
        ],
      }),
    );
    renderWithProviders(<BriscolaPage />);
    await waitFor(() => expect(screen.getByText(/CPUの勝ち.*70|CPUの勝ち.*50.*70/)).toBeInTheDocument());
  });

  it('shows tie banner when winnerIdx is -1 and game ended', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        gameEndFlag: true,
        winnerIdx: -1,
        players: [
          { id: 0, isHuman: true, cardCount: 0, cards: [], points: 60, trickCount: 10 },
          { id: 1, isHuman: false, cardCount: 0, cards: [], points: 60, trickCount: 10 },
        ],
      }),
    );
    renderWithProviders(<BriscolaPage />);
    await waitFor(() => expect(screen.getByText(/引き分け/)).toBeInTheDocument());
  });

  it('hides trump card label when stock is exhausted', async () => {
    mockExec.mockResolvedValue(makeState({ trumpCard: undefined, stockRemaining: 0 }));
    renderWithProviders(<BriscolaPage />);
    await waitFor(() => expect(screen.getByText(/トランプ: 使い切り/)).toBeInTheDocument());
  });

  it('disables play buttons when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
    renderWithProviders(<BriscolaPage />);
    await waitFor(() => {
      const btn = screen.getByRole('button', { name: '♠ A を出す' });
      expect(btn).toBeDisabled();
    });
  });

  it('shows confirm dialog on reset click and runs reset on accept', async () => {
    renderWithProviders(<BriscolaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the hint button on the human turn and requests a hint', async () => {
    renderWithProviders(<BriscolaPage />);
    const hintBtn = await screen.findByRole('button', { name: 'ヒント' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(makeState());
    fireEvent.click(hintBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('renders the server hint text with the translated reason when present', async () => {
    mockExec.mockResolvedValue(makeState({ hint: { cardIndex: 2, reason: 'lead_trump' } }));
    renderWithProviders(<BriscolaPage />);
    const hint = await screen.findByTestId('briscola-hint');
    expect(hint).toHaveTextContent('切り札でリードして主導権を握りましょう');
  });
});
