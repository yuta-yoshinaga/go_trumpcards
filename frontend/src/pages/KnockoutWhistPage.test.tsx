import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { knockoutWhistApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeKnockoutWhistState } from '../test/stateFactories';
import { KnockoutWhistPage } from './KnockoutWhistPage';

vi.mock('../api/gameApi', () => ({
  knockoutWhistApi: { exec: vi.fn() },
  actionLogApi: { knockoutwhist: vi.fn() },
}));

const mockExec = vi.mocked(knockoutWhistApi.exec);

const playPhaseState = makeKnockoutWhistState();
const trickEndState = makeKnockoutWhistState({
  phase: 1,
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 12 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 7 } },
  ],
});
const roundEndState = makeKnockoutWhistState({ phase: 2, roundWinnerIdx: 0 });
const gameEndState = makeKnockoutWhistState({
  phase: 3,
  gameEndFlag: true,
  winnerPlayer: 0,
  message: 'ゲーム終了！ あなたの勝ち！',
});
const cpuTurnState = makeKnockoutWhistState({ currentPlayerIdx: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('KnockoutWhistPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<KnockoutWhistPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<KnockoutWhistPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1 },
      }),
    );
  });

  it('renders the play phase with the human cards', async () => {
    renderWithProviders(<KnockoutWhistPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<KnockoutWhistPage />);
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
    renderWithProviders(<KnockoutWhistPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<KnockoutWhistPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<KnockoutWhistPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<KnockoutWhistPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('renders trump-select suit buttons and dispatches selecttrump', async () => {
    const trumpSelectState = makeKnockoutWhistState({ phase: 4, roundWinnerIdx: 0 });
    mockExec.mockResolvedValue(trumpSelectState);
    renderWithProviders(<KnockoutWhistPage />);
    const heartBtn = await screen.findByTestId('knockoutwhist-trump-3');
    mockExec.mockClear();
    mockExec.mockResolvedValue(trumpSelectState);
    fireEvent.click(heartBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('selecttrump', { trumpSuit: 3 }));
  });

  it('greys out an eliminated player panel with a readable (not too faint) dim', async () => {
    const eliminatedState = makeKnockoutWhistState();
    eliminatedState.players[1] = { ...eliminatedState.players[1], eliminated: true, dogbones: 0 };
    mockExec.mockResolvedValue(eliminatedState);
    const { container } = renderWithProviders(<KnockoutWhistPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.getAllByText(/脱落/).length).toBeGreaterThan(0);
    // Eliminated rows keep the strike-through but use a lighter dim for WCAG-AA legibility.
    const rows = container.querySelectorAll('[data-eliminated="true"]');
    expect(rows.length).toBeGreaterThan(0);
    for (const row of rows) {
      expect(row.className).toContain('opacity-70');
      expect(row.className).not.toContain('opacity-40');
    }
  });

  it('renders the leader badge with an opaque, high-contrast surface and an aria-label', async () => {
    // Default state: leadPlayerIdx 0 → the human is the leader.
    renderWithProviders(<KnockoutWhistPage />);
    const badge = await screen.findByLabelText('リーダー');
    // Opaque surface token (badgeInfoColors) instead of the old translucent bg-white/20.
    expect(badge.className).toContain('bg-ds-surface');
    expect(badge.className).not.toContain('bg-white/20');
  });
});
