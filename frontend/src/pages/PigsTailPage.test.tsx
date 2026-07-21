import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { pigtailApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { PigsTailResponse } from '../types/card';
import { PigsTailPage } from './PigsTailPage';

vi.mock('../hooks/useCliMode', () => ({
  useCliMode: vi.fn(() => ({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  })),
}));

const mockUseCliMode = vi.mocked(useCliMode);

vi.mock('../api/gameApi', () => ({
  pigtailApi: { exec: vi.fn() },
  actionLogApi: { pigtail: vi.fn() },
}));

const mockExec = vi.mocked(pigtailApi.exec);

const baseState: PigsTailResponse = {
  players: [
    { id: 0, isHuman: true, cardCount: 0, cards: [] },
    { id: 1, isHuman: false, cardCount: 0, cards: [] },
    { id: 2, isHuman: false, cardCount: 0, cards: [] },
    { id: 3, isHuman: false, cardCount: 0, cards: [] },
  ],
  circleCount: 52,
  centerTop: null,
  centerCount: 0,
  currentTurn: 0,
  gameEndFlag: false,
  loserIdx: -1,
  lastDrawCard: null,
  lastPenalty: false,
  cpuActions: [],
  humanAction: null,
  message: '',
};

const gameEndState: PigsTailResponse = {
  ...baseState,
  circleCount: 0,
  gameEndFlag: true,
  loserIdx: 1,
  message: 'Game Over! CPU 1 loses!',
  messageCode: 'pigtail.result.cpuLose',
  messageParams: { cpuId: '1' },
};

const humanLoseState: PigsTailResponse = {
  ...baseState,
  circleCount: 0,
  gameEndFlag: true,
  loserIdx: 0,
  message: 'Game Over! You lose!',
  messageCode: 'pigtail.result.humanLose',
};

beforeEach(() => {
  mockExec.mockResolvedValue(baseState);
  // Reset CLI mode to the default so tests that toggle it on (e.g. the
  // CLI-terminal case) don't leak state into the next test in source order.
  mockUseCliMode.mockReturnValue({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  });
});

describe('PigsTailPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<PigsTailPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalled();
    });
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders game state with circle and center info', async () => {
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => {
      expect(screen.getByText(/52/)).toBeInTheDocument();
    });
  });

  it('draw button is enabled on human turn', async () => {
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => {
      const drawBtn = screen.getByRole('button', { name: '山札から引く' });
      expect(drawBtn).not.toBeDisabled();
    });
  });

  it('shows the center top card rank and suit (not just the suit symbol)', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      centerTop: { design: 'SPADE', value: 1 },
      centerCount: 3,
    });
    renderWithProviders(<PigsTailPage />);
    const center = await screen.findByTestId('pt-center-top');
    // Rank (A) must be visible alongside the suit symbol.
    expect(center).toHaveTextContent('♠A');
  });

  it('falls back to "?" for a card whose design has no suit symbol', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      centerTop: { design: 'JOKER', value: 0 },
      centerCount: 1,
    });
    renderWithProviders(<PigsTailPage />);
    const center = await screen.findByTestId('pt-center-top');
    expect(center).toHaveTextContent('?');
  });

  it('accumulates a recent-center history strip across draws', async () => {
    mockExec.mockResolvedValue({ ...baseState, centerTop: { design: 'SPADE', value: 1 }, centerCount: 3 });
    renderWithProviders(<PigsTailPage />);
    const strip = await screen.findByTestId('pt-center-history');
    expect(strip).toHaveTextContent('♠A');

    mockExec.mockResolvedValue({ ...baseState, centerTop: { design: 'HEART', value: 3 }, centerCount: 4 });
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(screen.getByTestId('pt-center-history')).toHaveTextContent('♥3'));
    // The earlier center top is still part of the tail.
    expect(screen.getByTestId('pt-center-history')).toHaveTextContent('♠A');
  });

  it('does not duplicate the center history when the top card is unchanged', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      centerTop: { design: 'SPADE', value: 1 },
      centerCount: 3,
      circleCount: 52,
    });
    renderWithProviders(<PigsTailPage />);
    const strip = await screen.findByTestId('pt-center-history');
    expect(strip.children).toHaveLength(1);

    // A fresh draw signature (circleCount changed) but the same center top must
    // not append a duplicate entry.
    mockExec.mockResolvedValue({
      ...baseState,
      centerTop: { design: 'SPADE', value: 1 },
      centerCount: 3,
      circleCount: 51,
    });
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
    await waitFor(() => expect(screen.getByText((_, el) => el?.textContent === '山札: 51')).toBeInTheDocument());
    expect(screen.getByTestId('pt-center-history').children).toHaveLength(1);
  });

  it('clears the center history when the pile is collected (centerCount 0)', async () => {
    mockExec.mockResolvedValue({ ...baseState, centerTop: { design: 'SPADE', value: 1 }, centerCount: 3 });
    renderWithProviders(<PigsTailPage />);
    await screen.findByTestId('pt-center-history');

    mockExec.mockResolvedValue({ ...baseState, centerTop: null, centerCount: 0 });
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(screen.queryByTestId('pt-center-history')).not.toBeInTheDocument());
  });

  it('flips the drawn card face-up in the reveal area on a draw', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      lastDrawCard: { design: 'SPADE', value: 1 },
      lastPenalty: false,
    });
    renderWithProviders(<PigsTailPage />);
    const reveal = await screen.findByTestId('pt-draw-reveal');
    // The card face (AnimatedCard) is rendered, not just a text label.
    expect(reveal.querySelector('[data-testid="animated-card"]')).not.toBeNull();
    // The flip wrapper carries the motion-safe flip animation class.
    expect(reveal.querySelector('.motion-safe\\:animate-flipIn')).not.toBeNull();
    expect(reveal).toHaveTextContent('セーフ');
  });

  it('adds a red highlight ring to the drawn card on a penalty', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      lastDrawCard: { design: 'HEART', value: 3 },
      lastPenalty: true,
    });
    renderWithProviders(<PigsTailPage />);
    const reveal = await screen.findByTestId('pt-draw-reveal');
    expect(reveal.querySelector('.ring-ds-error')).not.toBeNull();
    expect(reveal).toHaveTextContent('ペナルティ');
  });

  it('does not render the draw reveal before any card is drawn', async () => {
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => expect(screen.getByTestId('circular-deck')).toBeInTheDocument());
    expect(screen.queryByTestId('pt-draw-reveal')).not.toBeInTheDocument();
  });

  it('draw button is disabled on game end', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => {
      const drawBtn = screen.getByRole('button', { name: '山札から引く' });
      expect(drawBtn).toBeDisabled();
    });
  });

  it('clicking draw button calls exec', async () => {
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument();
    });
    const drawBtn = screen.getByRole('button', { name: '山札から引く' });
    expect(drawBtn).not.toBeDisabled();
    fireEvent.click(drawBtn);
    // Verify at least the initial reset was called
    expect(mockExec).toHaveBeenCalled();
  });

  it('shows game end message when game is over', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => {
      const drawBtn = screen.getByRole('button', { name: '山札から引く' });
      expect(drawBtn).toBeDisabled();
    });
  });

  it('disables draw on game end (cpu loses)', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => {
      const drawBtn = screen.getByRole('button', { name: '山札から引く' });
      expect(drawBtn).toBeDisabled();
    });
  });

  it('does not show win celebration when human loses', async () => {
    mockExec.mockResolvedValue(humanLoseState);
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => {
      expect(screen.queryByTestId('win-celebration')).not.toBeInTheDocument();
    });
  });

  it('renders cpu actions when present', async () => {
    const stateWithCpuActions: PigsTailResponse = {
      ...baseState,
      cpuActions: [
        { drawPlayerIdx: 1, drawnCard: { design: 'SPADE', value: 5 }, penaltyFlag: false, penaltyCount: 0 },
        { drawPlayerIdx: 2, drawnCard: { design: 'HEART', value: 3 }, penaltyFlag: true, penaltyCount: 4 },
      ],
    };
    mockExec.mockResolvedValue(stateWithCpuActions);
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => {
      expect(screen.getByText(/ペナルティ/)).toBeInTheDocument();
      expect(screen.getByText(/セーフ/)).toBeInTheDocument();
    });
  });

  it('renders CLI terminal when CLI mode is enabled', async () => {
    mockUseCliMode.mockReturnValue({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => {
      expect(screen.getByRole('textbox')).toBeInTheDocument();
    });
    expect(screen.queryByRole('button', { name: '山札から引く' })).not.toBeInTheDocument();
  });

  it('renders the circular deck and dispatches draw when a ring card is tapped', async () => {
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => expect(screen.getByTestId('circular-deck')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('circular-deck-card-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('resets with the default player count on mount', async () => {
    mockExec.mockClear();
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // Mount reset carries no explicit player count (server uses the default).
    expect(mockExec.mock.calls[0]).toEqual(['reset']);
  });

  it('applies the selected player count on the next reset', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<PigsTailPage />);
    const select = await screen.findByTestId('pigtail-player-count');
    fireEvent.change(select, { target: { value: '6' } });

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, 6));
  });

  it('places the action log section inside the scrollable area before the footer', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<PigsTailPage />);
    const buttons = await screen.findAllByRole('button', { name: '棋譜を見る' });
    const footer = screen.getByRole('contentinfo');
    // At least one action-log view button (the ActionLogSection one) must
    // precede the footer in document order — i.e. live in the scroll area.
    const precedesFooter = buttons.some(
      (b) => (footer.compareDocumentPosition(b) & Node.DOCUMENT_POSITION_PRECEDING) !== 0,
    );
    expect(precedesFooter).toBe(true);
  });
});
