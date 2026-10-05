import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { bassetApi } from '../api/games/basset';
import i18n from '../i18n';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
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
    bankerHit: false,
    playerHit: false,
    payoutReceived: 0,
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
  playerCard: card('CLOVER', 4),
  turnsPlayed: 1,
  remaining: 50,
});
const decisionState = makeState({
  phase: 3,
  bet: { rank: 7, amount: 10, stage: 1 },
  bankerCard: card('SPADE', 3),
  playerCard: card('HEART', 7),
  hit: true,
  playerHit: true,
  turnsPlayed: 1,
  remaining: 50,
});
const roundEndState = makeState({
  phase: 4,
  bet: null,
  totalPayout: 20,
  payoutReceived: 20,
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

  it('shows remaining cards by rank and refreshes the counts after a deal', async () => {
    const initialCounts: number[] = Array.from({ length: 14 }, (_, i) => (i === 0 ? 0 : 4));
    const updatedCounts = [...initialCounts];
    updatedCounts[3] = 3;
    mockExec
      .mockResolvedValueOnce(makeState({ remainingByRank: initialCounts }))
      .mockResolvedValueOnce(makeState({ phase: 2, remaining: 50, remainingByRank: updatedCounts }));
    renderWithProviders(<BassetPage />);
    await screen.findByText('賭けなし');
    expect(screen.getByText('ランク別の残り札')).toBeInTheDocument();
    expect(screen.getByText('A: 4')).toBeInTheDocument();
    expect(screen.getByText('3: 4')).toBeInTheDocument();
    const previousLanguage = i18n.language;
    await i18n.changeLanguage('en');
    expect(screen.getByText('Cards remaining by rank')).toBeInTheDocument();
    expect(screen.getByText('3: 4')).toBeInTheDocument();
    await i18n.changeLanguage(previousLanguage);
    fireEvent.click(screen.getByRole('button', { name: '2枚めくる' }));
    expect(await screen.findByText('3: 3')).toBeInTheDocument();
  });

  it('gives the bet amount input a translated accessible name', async () => {
    const previousLanguage = i18n.language;
    renderWithProviders(<BassetPage />);
    try {
      expect(await screen.findByRole('spinbutton', { name: '賭け金' })).toBeInTheDocument();
      await i18n.changeLanguage('en');
      expect(screen.getByRole('spinbutton', { name: 'Bet amount' })).toBeInTheDocument();
    } finally {
      await i18n.changeLanguage(previousLanguage);
    }
  });

  it('shows the reset button and waits for confirmation before resetting', async () => {
    renderWithProviders(<BassetPage />);
    const reset = await screen.findByRole('button', { name: 'リセット' });
    mockExec.mockClear();

    fireEvent.click(reset);
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('reset');

    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows an error when a request fails after the game has loaded', async () => {
    mockExec.mockResolvedValue(bettingState);
    renderWithProviders(<BassetPage />);
    await screen.findByRole('button', { name: 'リセット' });

    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(await screen.findByRole('button', { name: '確認' }));

    expect(await screen.findByRole('alert')).toBeInTheDocument();
  });

  it('places the selected rank and amount as a bet', async () => {
    renderWithProviders(<BassetPage />);
    await screen.findByRole('button', { name: '賭ける' });
    const rankButtons = screen.getAllByRole('button').filter((button) => button.hasAttribute('aria-pressed'));
    expect(rankButtons).toHaveLength(13);
    expect(rankButtons.map((button) => button.textContent)).toEqual([
      'A',
      '2',
      '3',
      '4',
      '5',
      '6',
      '7',
      '8',
      '9',
      '10',
      'J',
      'Q',
      'K',
    ]);
    expect(screen.getByRole('button', { name: 'A' })).toHaveAttribute('aria-pressed', 'true');
    fireEvent.click(screen.getByRole('button', { name: 'K' }));
    expect(screen.getByRole('button', { name: 'K' })).toHaveAttribute('aria-pressed', 'true');
    fireEvent.change(screen.getByRole('spinbutton'), { target: { value: '25' } });
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '賭ける' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', { rank: 13, amount: 25 }));
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
    expect(screen.getByTestId('basset-turn-result')).toHaveTextContent(
      'どちらの札も賭けたランクと一致しませんでした。外れです。',
    );
  });

  it('explains banker losses in Japanese and English', async () => {
    const previousLanguage = i18n.language;
    mockExec.mockResolvedValueOnce(
      makeState({
        phase: 2,
        bankerCard: card('SPADE', 7),
        playerCard: card('HEART', 3),
        bankerHit: true,
        hit: true,
      }),
    );
    renderWithProviders(<BassetPage />);
    expect(await screen.findByTestId('basset-turn-result')).toHaveTextContent(
      '親札が賭けたランクと一致しました。賭け金を失いました。',
    );
    try {
      await i18n.changeLanguage('en');
      expect(screen.getByTestId('basset-turn-result')).toHaveTextContent(
        'The banker card matched your rank. Your bet was lost.',
      );
    } finally {
      await i18n.changeLanguage(previousLanguage);
    }
  });

  it('explains a player hit', async () => {
    mockExec.mockResolvedValueOnce(decisionState);
    renderWithProviders(<BassetPage />);
    expect(await screen.findByTestId('basset-turn-result')).toHaveTextContent(
      '子札が賭けたランクと一致しました。的中です。',
    );
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
    expect(await screen.findByTestId('basset-turn-result')).toHaveTextContent(
      '子札が賭けたランクと一致しました。的中です。',
    );
    expect(screen.queryByTestId('basset-payout-received')).not.toBeInTheDocument();
    const take = await screen.findByRole('button', { name: '配当を受け取る' });
    mockExec.mockClear();
    fireEvent.click(take);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('take'));
    expect(await screen.findByTestId('basset-payout-received')).toHaveTextContent('受取配当: 20チップ');
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
