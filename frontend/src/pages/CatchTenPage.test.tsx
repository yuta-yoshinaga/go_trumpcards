import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { catchtenApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CatchTenResponse } from '../types/card';
import { CatchTenPage } from './CatchTenPage';

vi.mock('../api/gameApi', () => ({
  catchtenApi: { exec: vi.fn() },
  actionLogApi: { catchten: vi.fn() },
}));

const mockExec = vi.mocked(catchtenApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<CatchTenResponse> = {}): CatchTenResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 3,
        cards: [card('SPADE', 1), card('HEART', 5), card('DIAMOND', 9)],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        team: 0,
      },
      { id: 1, isHuman: false, cardCount: 3, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0, team: 1 },
      { id: 2, isHuman: false, cardCount: 3, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0, team: 0 },
      { id: 3, isHuman: false, cardCount: 3, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0, team: 1 },
    ],
    phase: 0,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    currentTrick: [],
    trumpSuit: 0,
    dealerIdx: 0,
    teamScores: [0, 0],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 1, pointLimit: 41 },
    ...overrides,
  };
}

const playState = makeState();
const gameEndState = makeState({ phase: 3, gameEndFlag: true, winnerTeam: 0 });

beforeEach(() => {
  mockExec.mockResolvedValue(playState);
});

describe('CatchTenPage', () => {
  it('calls reset on mount with default config', async () => {
    renderWithProviders(<CatchTenPage />);
    // useTrickGameBase fires the mount reset with four positional args.
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 1, pointLimit: 41 }),
    );
  });

  // The phase key map must hold bare keys; usePhaseNames adds the `phase.`
  // prefix itself, so a prefixed key resolved to the literal
  // "phase.phase.play" on screen. See issue #4374.
  it('renders the translated phase name, not the raw i18n key', async () => {
    renderWithProviders(<CatchTenPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('プレイ'));
    expect(screen.getByTestId('phase-indicator')).not.toHaveTextContent('phase.');
  });

  it('renders CPU stats as a structured definition list with labeled fields', async () => {
    renderWithProviders(<CatchTenPage />);
    // Each CPU stat block is a <dl>; every field is a term/definition pair so
    // screen readers announce hand, team and scores independently instead of
    // one pipe-joined string.
    await waitFor(() => expect(screen.getAllByRole('term').length).toBeGreaterThan(0));
    // The sr-only labels give each value its own accessible name.
    expect(screen.getAllByText('手札').length).toBeGreaterThan(0);
    expect(screen.getAllByRole('definition').length).toBeGreaterThan(0);
    // The decorative pipe separators are hidden from assistive tech.
    const pipes = screen.getAllByText('|');
    for (const pipe of pipes) {
      expect(pipe).toHaveAttribute('aria-hidden', 'true');
    }
  });

  it('advances to the next trick when pressing n at trick end', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1 }));
    renderWithProviders(<CatchTenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'n' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('advances to the next round when pressing n at round end', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<CatchTenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'n' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('shows mid-game リセット button that opens confirm dialog', async () => {
    renderWithProviders(<CatchTenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('executes reset after confirm dialog is accepted', async () => {
    renderWithProviders(<CatchTenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1, pointLimit: 41 }),
    );
  });

  it('shows 次のゲーム and fires reset directly at game end (no confirm)', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CatchTenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, expect.any(Object)));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('color-codes the human player team badge (team 0 → info)', async () => {
    renderWithProviders(<CatchTenPage />);
    const humanTeam = await screen.findByTestId('catchten-human-team');
    expect(humanTeam.querySelector('span')).toHaveClass('text-ds-info');
  });

  it('color-codes the score-table team rows (team 0 info, team 1 error)', async () => {
    renderWithProviders(<CatchTenPage />);
    await waitFor(() => expect(screen.getAllByText('チーム 0').length).toBeGreaterThan(0));
    // Score-table cells render the team label inside a colored chip span.
    const team0Chips = screen.getAllByText('チーム 0').filter((el) => el.className.includes('text-ds-info'));
    const team1Chips = screen.getAllByText('チーム 1').filter((el) => el.className.includes('text-ds-error'));
    expect(team0Chips.length).toBeGreaterThan(0);
    expect(team1Chips.length).toBeGreaterThan(0);
  });

  it('plays the win celebration when the human team wins', async () => {
    mockExec.mockResolvedValue(gameEndState); // winnerTeam: 0 = human team
    renderWithProviders(<CatchTenPage />);
    expect(await screen.findByTestId('win-celebration')).toBeInTheDocument();
  });

  it('does not celebrate when the CPU team wins', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 3, gameEndFlag: true, winnerTeam: 1 }));
    renderWithProviders(<CatchTenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    // Outwait the celebration's 400ms delay so a wrongly-fired overlay would be visible.
    await act(() => new Promise((resolve) => setTimeout(resolve, 600)));
    expect(screen.queryByTestId('win-celebration')).not.toBeInTheDocument();
  });

  it('does not celebrate on a draw (winnerTeam -1)', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 3, gameEndFlag: true, winnerTeam: -1 }));
    renderWithProviders(<CatchTenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    await act(() => new Promise((resolve) => setTimeout(resolve, 600)));
    expect(screen.queryByTestId('win-celebration')).not.toBeInTheDocument();
  });

  it('names the recommended card (suit + rank) in the hint text', async () => {
    renderWithProviders(<CatchTenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    // cardIndex 0 in the human hand is ♠ A.
    mockExec.mockResolvedValueOnce(makeState({ hint: { cardIndex: 0, reason: 'lead_strong' } }));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(/♠ A/)).toBeInTheDocument());
  });
});
