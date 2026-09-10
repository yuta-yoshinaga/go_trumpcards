import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { bassetApi } from '../api/games/basset';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card } from '../types/common';
import type { BassetResponse } from '../types/games/basset';
import { BassetPage } from './BassetPage';

vi.mock('../api/games/basset', () => ({
  bassetApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(bassetApi.exec);

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<BassetResponse> = {}): BassetResponse {
  return {
    phase: 1,
    chips: 1000,
    bet: null,
    bankerCard: null,
    playerCard: null,
    hit: false,
    turnsPlayed: 0,
    turnsTotal: 25,
    remaining: 52,
    remainingByRank: Array.from({ length: 14 }, (_, i) => (i === 0 ? 0 : 4)),
    totalPayout: 0,
    gameEndFlag: false,
    message: '',
    ...overrides,
  };
}

const bettingState = makeState();
const turnState = makeState({
  phase: 2,
  bet: { rank: 7, amount: 10, stage: 1 },
  bankerCard: card('SPADE', 3),
  turnsPlayed: 1,
  remaining: 50,
});
const decisionState = makeState({
  phase: 3,
  bet: { rank: 7, amount: 10, stage: 1 },
  bankerCard: card('SPADE', 3),
  playerCard: card('HEART', 7),
  hit: true,
  turnsPlayed: 1,
  remaining: 50,
});
const roundEndState = makeState({
  phase: 4,
  bet: null,
  totalPayout: 20,
  chips: 1020,
});
const gameEndState = makeState({ phase: 5, gameEndFlag: true, chips: 0 });

beforeEach(() => {
  localStorage.clear();
  localStorage.setItem('tutorial_no_suggest', 'true');
  mockExec.mockReset();
  mockExec.mockResolvedValue(bettingState);
});

describe('BassetPage', () => {
  it('renders the skeleton while the initial request is pending', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<BassetPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('resets on mount and renders the betting state', async () => {
    renderWithProviders(<BassetPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText('チップ: 1000')).toBeInTheDocument();
    expect(screen.getByText('ターン: 0/25')).toBeInTheDocument();
    expect(screen.getByText('残り: 52')).toBeInTheDocument();
    expect(screen.getByText('賭けなし')).toBeInTheDocument();
  });

  it('places the selected rank and amount as a bet', async () => {
    renderWithProviders(<BassetPage />);
    await screen.findByRole('button', { name: '賭ける' });
    fireEvent.change(screen.getAllByRole('spinbutton')[0], { target: { value: '7' } });
    fireEvent.change(screen.getAllByRole('spinbutton')[1], { target: { value: '25' } });
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '賭ける' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', { rank: 7, amount: 25 }));
  });

  it('deals two cards and renders the turn result', async () => {
    mockExec.mockResolvedValueOnce(bettingState).mockResolvedValueOnce(turnState);
    renderWithProviders(<BassetPage />);
    await screen.findByRole('button', { name: '2枚めくる' });
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '2枚めくる' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
    expect(await screen.findByText('賭け: 7 / 10 / 1')).toBeInTheDocument();
    expect(screen.getByText('親の札: 3')).toBeInTheDocument();
  });

  it('shows both take and paroli after a hit', async () => {
    mockExec.mockResolvedValue(decisionState);
    renderWithProviders(<BassetPage />);
    expect(await screen.findByRole('button', { name: '配当を受け取る' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'パロリ' })).toBeInTheDocument();
    expect(screen.getByText('子の札: 7')).toBeInTheDocument();
  });

  it('raises the displayed paroli stage from 1 to 7 to 15 to 33 to 67', async () => {
    const stages = [1, 7, 15, 33, 67];
    mockExec.mockResolvedValueOnce(decisionState);
    for (const stage of stages.slice(1)) {
      mockExec.mockResolvedValueOnce(makeState({ ...decisionState, bet: { rank: 7, amount: 10, stage } }));
    }
    renderWithProviders(<BassetPage />);
    expect(await screen.findByText('賭け: 7 / 10 / 1')).toBeInTheDocument();
    for (const stage of stages.slice(1)) {
      fireEvent.click(screen.getByRole('button', { name: 'パロリ' }));
      expect(await screen.findByText(`賭け: 7 / 10 / ${stage}`)).toBeInTheDocument();
    }
  });

  it('takes the winnings and returns to the next-deal state', async () => {
    mockExec.mockResolvedValueOnce(decisionState).mockResolvedValueOnce(roundEndState);
    renderWithProviders(<BassetPage />);
    const take = await screen.findByRole('button', { name: '配当を受け取る' });
    mockExec.mockClear();
    fireEvent.click(take);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('take'));
    expect(await screen.findByRole('button', { name: '次のディール' })).toBeInTheDocument();
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のディール' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('shows the game-end state', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<BassetPage />);
    expect(await screen.findByTestId('phase-indicator')).toHaveTextContent('gameEnd');
    expect(screen.queryByRole('button', { name: '賭ける' })).not.toBeInTheDocument();
  });

  it('toggles to CLI mode and submits a command', async () => {
    renderWithProviders(<BassetPage />);
    await screen.findByRole('button', { name: 'CLIモードに切り替え' });
    fireEvent.click(screen.getByRole('button', { name: 'CLIモードに切り替え' }));
    const input = screen.getByRole('textbox', { name: 'コマンドを入力...' });
    fireEvent.change(input, { target: { value: 'paroli' } });
    mockExec.mockClear();
    fireEvent.keyDown(input, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('paroli'));
  });
});
