import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { manilleApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeManilleState } from '../test/stateFactories';
import { ManillePage } from './ManillePage';

vi.mock('../api/gameApi', () => ({
  manilleApi: { exec: vi.fn() },
  actionLogApi: { manille: vi.fn() },
}));

const mobileFlag = vi.hoisted(() => ({ value: false }));
vi.mock('../hooks/useCardDimensions', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../hooks/useCardDimensions')>();
  return {
    ...actual,
    useCardDimensions: () => ({ ...actual.useCardDimensions(), isMobile: mobileFlag.value }),
  };
});

const mockExec = vi.mocked(manilleApi.exec);

const playPhaseState = makeManilleState();
const trickEndState = makeManilleState({
  phase: 1,
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 12 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 13 } },
  ],
});
const roundEndState = makeManilleState({ phase: 2, roundCardPoints: [35, 25] });
const gameEndState = makeManilleState({
  phase: 3,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'ゲーム終了！ あなたのチームの勝ち！',
});
const cpuTurnState = makeManilleState({ currentPlayerIdx: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
  mobileFlag.value = false;
});

describe('ManillePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<ManillePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<ManillePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetPoints: 101 },
      }),
    );
  });

  it('renders the play phase with the human cards', async () => {
    renderWithProviders(<ManillePage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<ManillePage />);
    const card = await screen.findByAltText('♥ Q');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<ManillePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('shows the trick winner and team in a status banner at trick end', async () => {
    // leadPlayerIdx 0 = the human (Team A), so their own-team banner appears.
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<ManillePage />);
    const banner = await screen.findByTestId('manille-trick-winner');
    expect(banner).toHaveTextContent('あなた（チームA）がトリック獲得');
    expect(banner).toHaveClass('text-ds-accent');
    expect(banner).toHaveAttribute('role', 'status');
    expect(banner).toHaveAttribute('aria-live', 'polite');
  });

  it('shows an opponent-team win without own-team emphasis', async () => {
    mockExec.mockResolvedValue({ ...trickEndState, leadPlayerIdx: 1 });
    renderWithProviders(<ManillePage />);
    const banner = await screen.findByTestId('manille-trick-winner');
    expect(banner).toHaveTextContent('チームB');
    expect(banner).not.toHaveClass('text-ds-accent');
  });

  it('does not show the trick-winner banner during play', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<ManillePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('manille-trick-winner')).not.toBeInTheDocument();
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<ManillePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<ManillePage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたのチームの勝ち！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<ManillePage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('highlights the human player and their partner (same team) in the player list', async () => {
    renderWithProviders(<ManillePage />);
    // Human is player 0 → team 0. Even ids (0, 2) are own team; odd ids (1, 3) are opponents.
    await waitFor(() => expect(screen.getByTestId('manille-player-0')).toBeInTheDocument());
    expect(screen.getByTestId('manille-player-0')).toHaveAttribute('data-own-team', 'true');
    expect(screen.getByTestId('manille-player-2')).toHaveAttribute('data-own-team', 'true');
    expect(screen.getByTestId('manille-player-1')).not.toHaveAttribute('data-own-team');
    expect(screen.getByTestId('manille-player-3')).not.toHaveAttribute('data-own-team');
  });

  it('applies the same-team highlight in the mobile player list', async () => {
    mobileFlag.value = true;
    renderWithProviders(<ManillePage />);
    // Expand the <details> player list, then check the shared row renderer applied the flag.
    const summary = await screen.findByText('プレイヤー');
    fireEvent.click(summary);
    expect(screen.getByTestId('manille-player-0')).toHaveAttribute('data-own-team', 'true');
    expect(screen.getByTestId('manille-player-1')).not.toHaveAttribute('data-own-team');
  });
});
