import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { russianbankApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, RussianBankPlayer, RussianBankResponse } from '../types/card';
import { RussianBankPage } from './RussianBankPage';

vi.mock('../api/gameApi', () => ({
  russianbankApi: { exec: vi.fn() },
  actionLogApi: { russianbank: vi.fn() },
}));

const mockExec = vi.mocked(russianbankApi.exec);

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makePlayer(overrides: Partial<RussianBankPlayer> = {}): RussianBankPlayer {
  return {
    id: 0,
    isHuman: true,
    reserveCount: 13,
    reserveTop: card('DIAMOND', 7),
    handCount: 39,
    wasteCount: 0,
    wasteTop: undefined,
    stopPoints: 0,
    ...overrides,
  };
}

function makeState(overrides: Partial<RussianBankResponse> = {}): RussianBankResponse {
  return {
    phase: 1,
    currentPlayerIdx: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    isHumanTurn: true,
    canCallStop: false,
    canUndo: false,
    moveCount: 0,
    tableau: [[], [], [], []],
    foundations: [[], [], [], [], [], [], [], []],
    players: [makePlayer(), makePlayer({ id: 1, isHuman: false })],
    config: { cpuDifficulty: 1 },
    message: '',
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('RussianBankPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<RussianBankPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<RussianBankPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders both players and the board', async () => {
    renderWithProviders(<RussianBankPage />);
    await waitFor(() => expect(screen.getByTestId('player-0')).toBeInTheDocument());
    expect(screen.getByTestId('player-1')).toBeInTheDocument();
    expect(screen.getByTestId('tableau-0')).toBeInTheDocument();
    expect(screen.getByTestId('foundation-0')).toBeInTheDocument();
  });

  it('selects a source then sends it to a foundation', async () => {
    renderWithProviders(<RussianBankPage />);
    const reserve = await screen.findByTestId('reserve-0');
    fireEvent.click(reserve);
    const toFnd = await screen.findByTestId('to-foundation');
    mockExec.mockClear();
    fireEvent.click(toFnd);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pf', { zone: 0, fromOpp: false, col: 0 }));
  });

  it('labels slots by card + zone and marks the selected source aria-pressed', async () => {
    renderWithProviders(<RussianBankPage />);
    // Reserve top is ♦7 in the human's reserve.
    const reserve = await screen.findByTestId('reserve-0');
    expect(reserve).toHaveAttribute('aria-label', '♦ 7（自リザーブ）');
    expect(reserve).toHaveAttribute('aria-pressed', 'false');
    // An empty tableau column names its zone.
    expect(screen.getByTestId('tableau-0')).toHaveAttribute('aria-label', 'タブロー1（空き）');
    // Selecting the reserve flips aria-pressed and announces the source.
    fireEvent.click(reserve);
    expect(screen.getByTestId('reserve-0')).toHaveAttribute('aria-pressed', 'true');
    const src = screen.getByTestId('rb-selected-source');
    expect(src).toHaveAttribute('role', 'status');
    expect(src).toHaveTextContent('自リザーブ');
  });

  it('selects a source then moves it to a tableau column', async () => {
    renderWithProviders(<RussianBankPage />);
    const reserve = await screen.findByTestId('reserve-0');
    fireEvent.click(reserve);
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('tableau-2'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('mt', { zone: 0, fromOpp: false, col: 0, toCol: 2 }));
  });

  it('discards to end the turn', async () => {
    renderWithProviders(<RussianBankPage />);
    const btn = await screen.findByTestId('discard-button');
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('d'));
  });

  it('shows the stop button only when a violation can be caught', async () => {
    mockExec.mockResolvedValue(makeState({ canCallStop: true }));
    renderWithProviders(<RussianBankPage />);
    const btn = await screen.findByTestId('stop-button');
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('s'));
  });

  it('shows the undo button only when a move can be undone', async () => {
    mockExec.mockResolvedValue(makeState({ canUndo: true }));
    renderWithProviders(<RussianBankPage />);
    const btn = await screen.findByTestId('undo-button');
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('u'));
  });

  it('hides action controls when it is the CPU turn', async () => {
    mockExec.mockResolvedValue(makeState({ isHumanTurn: false, currentPlayerIdx: 1 }));
    renderWithProviders(<RussianBankPage />);
    await waitFor(() => expect(screen.getByTestId('player-0')).toBeInTheDocument());
    expect(screen.queryByTestId('discard-button')).not.toBeInTheDocument();
  });

  it('shows the game-over label at game end', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2, gameEndFlag: true, isHumanTurn: false, winnerIdx: 0 }));
    renderWithProviders(<RussianBankPage />);
    await waitFor(() => expect(screen.getAllByText('ゲーム終了').length).toBeGreaterThan(0));
    expect(screen.queryByTestId('discard-button')).not.toBeInTheDocument();
  });
});
