import { screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { clocksolitaireApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, ClockSolitaireCard, ClockSolitaireResponse } from '../types/card';
import { ClockSolitairePage } from './ClockSolitairePage';

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
  clocksolitaireApi: { exec: vi.fn() },
  actionLogApi: { clocksolitaire: vi.fn() },
}));

const mockExec = vi.mocked(clocksolitaireApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeTestPiles(): ClockSolitaireCard[][] {
  const piles: ClockSolitaireCard[][] = [];
  for (let i = 0; i < 13; i++) {
    const pile: ClockSolitaireCard[] = [];
    for (let j = 0; j < 4; j++) {
      pile.push({ card: card('SPADE', (i % 13) + 1), faceUp: false });
    }
    piles.push(pile);
  }
  piles[12][3].faceUp = true;
  return piles;
}

function makeFaceUpCount(): number[] {
  const fuc = Array(13).fill(0);
  fuc[12] = 1;
  return fuc;
}

const playingState: ClockSolitaireResponse = {
  piles: makeTestPiles(),
  faceUpCount: makeFaceUpCount(),
  phase: 0,
  stepCount: 0,
  currentCard: card('SPADE', 5),
  message: '',
  messageCode: 'clocksolitaire.playing',
};

const gameClearState: ClockSolitaireResponse = {
  ...playingState,
  phase: 1,
  stepCount: 48,
  currentCard: undefined,
  messageCode: 'clocksolitaire.gameClear',
  messageParams: { stepCount: '48' },
};

const gameOverState: ClockSolitaireResponse = {
  ...playingState,
  phase: 2,
  stepCount: 30,
  currentCard: undefined,
  messageCode: 'clocksolitaire.gameOver',
};

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(playingState);
});

describe('ClockSolitairePage', () => {
  it('renders heading', async () => {
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('renders step count', async () => {
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByText(/ステップ数/)).toBeInTheDocument());
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows game clear phase name', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('ゲームクリア'));
  });

  it('shows game over phase name', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('ゲームオーバー'));
  });

  it('renders clock position labels', async () => {
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByText('12')).toBeInTheDocument());
    expect(screen.getByText('6')).toBeInTheDocument();
  });

  it('renders center K label', async () => {
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByText('K')).toBeInTheDocument());
  });

  it('renders current card section when playing', async () => {
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByText('手持ちカード')).toBeInTheDocument());
  });

  it('does not render current card section when game over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('ゲームオーバー'));
    expect(screen.queryByText('手持ちカード')).not.toBeInTheDocument();
  });

  it('renders face-up count for piles', async () => {
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getAllByText('0/4').length).toBeGreaterThan(0));
  });

  it('renders playing phase indicator', async () => {
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('プレイ中'));
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
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => {
      expect(screen.getByRole('textbox')).toBeInTheDocument();
    });
    expect(screen.queryByText('手持ちカード')).not.toBeInTheDocument();
  });

  // Helper: the CLI-mode test above leaves `mockUseCliMode` returning `cliEnabled: true`
  // because vi.clearAllMocks() only clears call history, not the implementation. Restore
  // the default cliEnabled: false at the top of each remaining test so the clock face renders.
  const restoreCliModeDefault = (): void => {
    mockUseCliMode.mockReturnValue({
      cliEnabled: false,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
  };

  it('marks the matching hour pile as the flight target while a 1-12 card is waiting', async () => {
    restoreCliModeDefault();
    // playingState already has currentCard = ♠5, so the 5 o'clock pile (index 4) should be the
    // single flight target. Other piles must not carry the data attribute.
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByText(/ステップ数/)).toBeInTheDocument());
    const targets = document.querySelectorAll('[data-flight-target="true"]');
    expect(targets).toHaveLength(1);
    const target = targets[0] as HTMLElement;
    expect(target.className).toContain('ring-ds-warning');
    expect(target.className).toContain('animate-pulse');
  });

  it('marks the center K pile as the flight target when a King is waiting', async () => {
    restoreCliModeDefault();
    mockExec.mockResolvedValue({ ...playingState, currentCard: card('SPADE', 13) });
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByText(/ステップ数/)).toBeInTheDocument());
    const targets = document.querySelectorAll('[data-flight-target="true"]');
    // Only the center K pile should match for value=13.
    expect(targets).toHaveLength(1);
    expect((targets[0] as HTMLElement).className).toContain('ring-ds-warning');
  });

  it('renders no flight target when there is no card awaiting placement', async () => {
    restoreCliModeDefault();
    mockExec.mockResolvedValue({ ...playingState, currentCard: undefined });
    renderWithProviders(<ClockSolitairePage />);
    await waitFor(() => expect(screen.getByText(/ステップ数/)).toBeInTheDocument());
    expect(document.querySelectorAll('[data-flight-target="true"]')).toHaveLength(0);
  });
});
