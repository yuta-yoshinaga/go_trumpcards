import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { madrassoApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeMadrassoState } from '../test/stateFactories';
import { MadrassoPage } from './MadrassoPage';

vi.mock('../api/gameApi', () => ({
  madrassoApi: { exec: vi.fn() },
  actionLogApi: { madrasso: vi.fn() },
}));

const mockExec = vi.mocked(madrassoApi.exec);

const playPhaseState = makeMadrassoState();
const trickEndState = makeMadrassoState({
  phase: 1,
  currentTrick: [
    { playerIdx: 0, card: { design: 'SPADE', value: 3 } },
    { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
  ],
});
const roundEndState = makeMadrassoState({ phase: 2 });
const gameEndState = makeMadrassoState({
  phase: 3,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'ゲーム終了！ チームAの勝ち！',
});
const cpuTurnState = makeMadrassoState({ currentPlayerIdx: 1 });

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
});

describe('MadrassoPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<MadrassoPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with config', async () => {
    renderWithProviders(<MadrassoPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        targetPoints: 21,
      }),
    );
  });

  it('renders play phase with human cards', async () => {
    renderWithProviders(<MadrassoPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♦ K')).toBeInTheDocument();
    });
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<MadrassoPage />);
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
    renderWithProviders(<MadrassoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with next round button', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<MadrassoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('renders game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<MadrassoPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ チームAの勝ち！')).toBeInTheDocument());
  });

  it('does not show play button on CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<MadrassoPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  // **1/3 点の 3 ドット表示はこのゲームには無い。** カード点は整数で、
  // 1 ディールで 121 点ある。クローン元 (トレセッテ) の遺物を残すと、
  // 「2/3 たまった」という意味の無い表示になる。
  it('renders the round card points out of the deal total', async () => {
    mockExec.mockResolvedValue(makeMadrassoState({ teamRoundPoints: [70, 51] }));
    renderWithProviders(<MadrassoPage />);

    const team0 = await screen.findByTestId('tr-round-points-0');
    expect(team0).toHaveTextContent('70 / 121');
    expect(screen.getByTestId('tr-round-points-1')).toHaveTextContent('51 / 121');

    // 3 ドットの表示は残っていないこと。
    expect(screen.queryByTestId('tr-thirds-0')).not.toBeInTheDocument();
  });

  it('shows the empty previous-trick message at the start of a round', async () => {
    mockExec.mockResolvedValue(makeMadrassoState({ lastTrick: [], lastTrickWinner: -1 }));
    renderWithProviders(<MadrassoPage />);

    const viewer = await screen.findByTestId('tr-previous-trick');
    expect(viewer).toHaveTextContent('前のトリック');
    expect(viewer).toHaveTextContent('このラウンドにはまだ前のトリックがありません');
  });

  it('renders the previous trick cards and winner once a trick has resolved', async () => {
    mockExec.mockResolvedValue(
      makeMadrassoState({
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
    renderWithProviders(<MadrassoPage />);

    const viewer = await screen.findByTestId('tr-previous-trick');
    // Winner label is rendered from the previousTrickWinner i18n key.
    expect(viewer).toHaveTextContent('が獲得');
    // The winning card carries the WIN badge.
    expect(viewer.querySelector('[data-testid="trick-winner-badge"]')).not.toBeNull();
  });

  it('renders a collapsible card-point legend with the scoring values', async () => {
    mockExec.mockResolvedValue(makeMadrassoState());
    renderWithProviders(<MadrassoPage />);

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
    mockExec.mockResolvedValue(makeMadrassoState({ playableIndices: [1] }));
    renderWithProviders(<MadrassoPage />);

    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // 手札は SPADE 3 (index 0) と DIAMOND 13 (index 1)。合法なのは index 1 だけ。
    const cards = screen.getAllByRole('button').filter((b) => b.hasAttribute('aria-pressed'));
    expect(cards).toHaveLength(2);
    expect(cards[0]).toHaveAttribute('aria-disabled', 'true');
    expect(cards[1]).not.toHaveAttribute('aria-disabled', 'true');
  });

  it('leaves every card enabled when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeMadrassoState({ currentPlayerIdx: 1, playableIndices: [] }));
    renderWithProviders(<MadrassoPage />);

    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    const cards = screen.getAllByRole('button').filter((b) => b.hasAttribute('aria-pressed'));
    expect(cards).toHaveLength(2);
    for (const card of cards) {
      expect(card).not.toHaveAttribute('aria-disabled', 'true');
    }
  });
});
