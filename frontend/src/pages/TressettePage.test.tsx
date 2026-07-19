import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tressetteApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeTressetteState } from '../test/stateFactories';
import { TressettePage } from './TressettePage';

vi.mock('../api/gameApi', () => ({
  tressetteApi: { exec: vi.fn() },
  actionLogApi: { tressette: vi.fn() },
}));

const mockExec = vi.mocked(tressetteApi.exec);

const playPhaseState = makeTressetteState();
const trickEndState = makeTressetteState({
  phase: 1,
  currentTrick: [
    { playerIdx: 0, card: { design: 'SPADE', value: 3 } },
    { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
  ],
});
const roundEndState = makeTressetteState({ phase: 2 });
const gameEndState = makeTressetteState({
  phase: 3,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'ゲーム終了！ チームAの勝ち！',
});
const cpuTurnState = makeTressetteState({ currentPlayerIdx: 1 });

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
});

describe('TressettePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<TressettePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with config', async () => {
    renderWithProviders(<TressettePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        targetPoints: 21,
      }),
    );
  });

  it('renders play phase with human cards', async () => {
    renderWithProviders(<TressettePage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ 3')).toBeInTheDocument();
      expect(screen.getByAltText('♦ K')).toBeInTheDocument();
    });
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<TressettePage />);
    const card = await screen.findByAltText('♠ 3');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 0));
  });

  it('renders trick end with next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<TressettePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with next round button', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<TressettePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('renders game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<TressettePage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ チームAの勝ち！')).toBeInTheDocument());
  });

  it('does not show play button on CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<TressettePage />);
    await waitFor(() => expect(screen.getByAltText('♠ 3')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('renders a three-dot thirds indicator with the filled count and remaining tooltip', async () => {
    mockExec.mockResolvedValue(makeTressetteState({ teamRoundThirds: [2, 0] }));
    renderWithProviders(<TressettePage />);

    const team0 = await screen.findByTestId('tr-thirds-0');
    // 3 dots rendered; 2 filled (bg-ds-accent), 1 empty (border).
    const dots0 = team0.querySelectorAll('span[aria-hidden="true"]');
    expect(dots0).toHaveLength(3);
    expect(team0.querySelectorAll('.bg-ds-accent')).toHaveLength(2);
    // Tooltip shows the remaining thirds (3 - 2 = 1) and sr-only text the filled count.
    expect(team0).toHaveAttribute('title', 'ラウンド得点まであと1サーズ');
    expect(team0).toHaveTextContent('2/3');

    // Team 1 has 0 thirds → no filled dots.
    const team1 = screen.getByTestId('tr-thirds-1');
    expect(team1.querySelectorAll('.bg-ds-accent')).toHaveLength(0);
  });

  it('shows the empty previous-trick message at the start of a round', async () => {
    mockExec.mockResolvedValue(makeTressetteState({ lastTrick: [], lastTrickWinner: -1 }));
    renderWithProviders(<TressettePage />);

    const viewer = await screen.findByTestId('tr-previous-trick');
    expect(viewer).toHaveTextContent('前のトリック');
    expect(viewer).toHaveTextContent('このラウンドにはまだ前のトリックがありません');
  });

  it('renders the previous trick cards and winner once a trick has resolved', async () => {
    mockExec.mockResolvedValue(
      makeTressetteState({
        trickNumber: 2,
        lastTrick: [
          { playerIdx: 1, card: { design: 'SPADE', value: 3 } },
          { playerIdx: 2, card: { design: 'SPADE', value: 1 } },
          { playerIdx: 3, card: { design: 'SPADE', value: 5 } },
          { playerIdx: 0, card: { design: 'SPADE', value: 7 } },
        ],
        lastTrickWinner: 2,
      }),
    );
    renderWithProviders(<TressettePage />);

    const viewer = await screen.findByTestId('tr-previous-trick');
    // Winner label is rendered from the previousTrickWinner i18n key.
    expect(viewer).toHaveTextContent('が獲得');
    // The winning card carries the WIN badge.
    expect(viewer.querySelector('[data-testid="trick-winner-badge"]')).not.toBeNull();
  });

  it('renders a collapsible card-point legend with the scoring values', async () => {
    mockExec.mockResolvedValue(makeTressetteState());
    renderWithProviders(<TressettePage />);

    const legend = await screen.findByTestId('tr-point-legend');
    // Summary is always present; the point values render inside the details.
    expect(legend).toHaveTextContent('点数の凡例');
    expect(legend).toHaveTextContent('1点');
    expect(legend).toHaveTextContent('1/3点');
    expect(legend).toHaveTextContent('+1/3点');
    expect(legend).toHaveTextContent('0点');
  });
});
