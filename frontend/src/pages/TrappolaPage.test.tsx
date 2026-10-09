import { cleanup, fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { trappolaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeTrappolaState } from '../test/stateFactories';
import { TrappolaPage } from './TrappolaPage';

vi.mock('../api/gameApi', () => ({
  trappolaApi: { exec: vi.fn() },
  actionLogApi: { trappola: vi.fn() },
}));

const mockExec = vi.mocked(trappolaApi.exec);

const playPhaseState = makeTrappolaState();
const trickEndState = makeTrappolaState({
  phase: 1,
  currentTrick: [
    { playerIdx: 0, card: { design: 'SPADE', value: 3 } },
    { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
  ],
});
const roundEndState = makeTrappolaState({ phase: 2 });
const gameEndState = makeTrappolaState({
  phase: 3,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'ゲーム終了！ チームAの勝ち！',
});
const cpuTurnState = makeTrappolaState({ currentPlayerIdx: 1 });

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
});

describe('TrappolaPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<TrappolaPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with config', async () => {
    renderWithProviders(<TrappolaPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        targetPoints: 21,
      }),
    );
  });

  it('renders play phase with human cards', async () => {
    renderWithProviders(<TrappolaPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♦ K')).toBeInTheDocument();
    });
  });

  it('highlights the hinted hand cards and dims the others', async () => {
    renderWithProviders(<TrappolaPage />);

    await screen.findByAltText('♠ A');
    const cards = screen.getAllByRole('button').filter((button) => button.hasAttribute('aria-pressed'));
    expect(cards[0]).not.toHaveClass('opacity-60');
    expect(cards[1]).not.toHaveClass('opacity-60');

    mockExec.mockResolvedValueOnce(makeTrappolaState({ hint: { cardIndices: [1], reason: 'follow_win' } }));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => {
      expect(cards[1].style.boxShadow).toContain('rgba(232, 146, 58');
      expect(cards[0]).toHaveClass('opacity-60');
    });
  });

  it('exposes each team name as a row header for its score cells', async () => {
    renderWithProviders(<TrappolaPage />);

    const teamA = await screen.findByRole('rowheader', { name: 'チームA' });
    expect(teamA).toHaveAttribute('scope', 'row');
    expect(teamA.closest('tr')).toHaveTextContent('チームA');
    expect(teamA.closest('tr')).toHaveTextContent('0');
    expect(screen.getByRole('columnheader', { name: '得点' })).toBeInTheDocument();
    expect(screen.getByRole('columnheader', { name: '今ラウンド(サーズ)' })).toBeInTheDocument();
    expect(screen.getByRole('rowheader', { name: 'チームB' })).toHaveAttribute('scope', 'row');
  });

  it('shows the server target points and updates only after a reset response', async () => {
    renderWithProviders(<TrappolaPage />);

    expect(await screen.findByText('目標点: 21点')).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText('目標点'), { target: { value: '31' } });

    expect(screen.getByText('目標点: 21点')).toBeInTheDocument();
    expect(screen.queryByText('目標点: 31点')).not.toBeInTheDocument();

    mockExec.mockResolvedValue(makeTrappolaState({ config: { cpuDifficulty: 1, targetPoints: 31 } }));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(within(await screen.findByRole('alertdialog')).getByRole('button', { name: '確認' }));

    expect(await screen.findByText('目標点: 31点')).toBeInTheDocument();
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<TrappolaPage />);
    const card = await screen.findByAltText('♠ A');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 0));
  });

  it('renders trick end with next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<TrappolaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('shows the domain supplied current and resolved trick points with the winning team', async () => {
    mockExec.mockResolvedValue(
      makeTrappolaState({
        currentTrick: [{ playerIdx: 0, card: { design: 'SPADE', value: 1 } }],
        currentTrickThirds: 3,
      }),
    );
    renderWithProviders(<TrappolaPage />);
    expect(await screen.findByText('場札の得点: 3/3点')).toBeInTheDocument();

    cleanup();
    mockExec.mockResolvedValue(
      makeTrappolaState({
        phase: 1,
        trickNumber: 2,
        lastTrick: [{ playerIdx: 2, card: { design: 'SPADE', value: 1 } }],
        lastTrickWinner: 2,
        lastTrickThirds: 3,
      }),
    );
    renderWithProviders(<TrappolaPage />);
    expect(await screen.findByText('チームAがこのトリックで獲得: 3/3点')).toBeInTheDocument();
  });

  it('renders round end with next round button', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<TrappolaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('renders game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<TrappolaPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ チームAの勝ち！')).toBeInTheDocument());
  });

  it('does not show play button on CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<TrappolaPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('renders a three-dot thirds indicator with the filled count and remaining tooltip', async () => {
    mockExec.mockResolvedValue(makeTrappolaState({ teamRoundThirds: [2, 0] }));
    renderWithProviders(<TrappolaPage />);

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
    mockExec.mockResolvedValue(makeTrappolaState({ lastTrick: [], lastTrickWinner: -1 }));
    renderWithProviders(<TrappolaPage />);

    const viewer = await screen.findByTestId('tr-previous-trick');
    expect(viewer).toHaveTextContent('前のトリック');
    expect(viewer).toHaveTextContent('このラウンドにはまだ前のトリックがありません');
  });

  it('renders the previous trick cards and winner once a trick has resolved', async () => {
    mockExec.mockResolvedValue(
      makeTrappolaState({
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
    renderWithProviders(<TrappolaPage />);

    const viewer = await screen.findByTestId('tr-previous-trick');
    // Winner label is rendered from the previousTrickWinner i18n key.
    expect(viewer).toHaveTextContent('が獲得');
    // The winning card carries the WIN badge.
    expect(viewer.querySelector('[data-testid="trick-winner-badge"]')).not.toBeNull();
  });

  it('renders a collapsible card-point legend with the scoring values', async () => {
    mockExec.mockResolvedValue(makeTrappolaState());
    renderWithProviders(<TrappolaPage />);

    const legend = await screen.findByTestId('tr-point-legend');
    // Summary is always present; the point values render inside the details.
    expect(legend).toHaveTextContent('点数の凡例');
    expect(legend).toHaveTextContent('1点');
    expect(legend).toHaveTextContent('1/3点');
    expect(legend).toHaveTextContent('+1/3点');
    expect(legend).toHaveTextContent('0点');
  });

  // **合法手はサーバーが計算済みなのに画面が使っていなかった (#4718)。**
  // マストフォローに反する札もクリックできてしまい、エラーが返って初めて分かる。
  // Tute/Sueca と同じ validIndices 配線に揃える。
  it('dims the cards that must-follow forbids on the human turn', async () => {
    mockExec.mockResolvedValue(makeTrappolaState({ playableIndices: [1] }));
    renderWithProviders(<TrappolaPage />);

    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // 手札は SPADE 3 (index 0) と DIAMOND 13 (index 1)。合法なのは index 1 だけ。
    const cards = screen.getAllByRole('button').filter((b) => b.hasAttribute('aria-pressed'));
    expect(cards).toHaveLength(2);
    expect(cards[0]).toHaveAttribute('aria-disabled', 'true');
    expect(cards[1]).not.toHaveAttribute('aria-disabled', 'true');
  });

  it('leaves every card enabled when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeTrappolaState({ currentPlayerIdx: 1, playableIndices: [] }));
    renderWithProviders(<TrappolaPage />);

    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    const cards = screen.getAllByRole('button').filter((b) => b.hasAttribute('aria-pressed'));
    expect(cards).toHaveLength(2);
    for (const card of cards) {
      expect(card).not.toHaveAttribute('aria-disabled', 'true');
    }
  });
});
