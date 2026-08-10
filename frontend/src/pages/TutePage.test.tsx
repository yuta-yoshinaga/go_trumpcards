import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tuteApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeTuteState } from '../test/stateFactories';
import { TutePage } from './TutePage';

vi.mock('../api/gameApi', () => ({
  tuteApi: { exec: vi.fn() },
  actionLogApi: { tute: vi.fn() },
}));

const mockExec = vi.mocked(tuteApi.exec);

const playPhaseState = makeTuteState();
const marriageState = makeTuteState({ canDeclareMarriage: true });
const tuteDeclState = makeTuteState({ canDeclareTute: true });
const trickEndState = makeTuteState({
  phase: 1,
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 12 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 13 } },
  ],
});
const roundEndState = makeTuteState({ phase: 2, roundTeamPoints: [70, 60] });
const gameEndState = makeTuteState({
  phase: 3,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'ゲーム終了！ あなたのチームの勝ち！',
});
const cpuTurnState = makeTuteState({ currentPlayerIdx: 1 });

beforeEach(() => {
  localStorage.clear();
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('TutePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<TutePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<TutePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetPoints: 121 },
      }),
    );
  });

  it('renders the play phase with the human cards', async () => {
    renderWithProviders(<TutePage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<TutePage />);
    const card = await screen.findByAltText('♥ Q');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('shows marriage declaration buttons when allowed and dispatches a suit', async () => {
    mockExec.mockResolvedValue(marriageState);
    renderWithProviders(<TutePage />);
    // The button's accessible name is now the spoken aria-label (suit name + points);
    // heart is not the trump (♦) so the marriage is worth 20.
    const heartBtn = await screen.findByRole('button', { name: 'ハートのマリッジを宣言（20点）' });
    fireEvent.click(heartBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('marriage', { suit: 3 }));
  });

  it('labels a trump-suit marriage as worth 40 points', async () => {
    // Make heart (the human's K+Q suit) the trump so the marriage is worth 40.
    mockExec.mockResolvedValue(makeTuteState({ canDeclareMarriage: true, trumpSuit: 3 }));
    renderWithProviders(<TutePage />);
    expect(await screen.findByRole('button', { name: 'ハートのマリッジを宣言（40点）' })).toBeInTheDocument();
  });

  it('shows the Tute declaration button when allowed and dispatches', async () => {
    mockExec.mockResolvedValue(tuteDeclState);
    renderWithProviders(<TutePage />);
    const tuteBtn = await screen.findByRole('button', { name: 'Tute 宣言' });
    fireEvent.click(tuteBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('tute'));
  });

  it('groups declarations in a labelled fieldset and describes the Tute button', async () => {
    mockExec.mockResolvedValue(tuteDeclState);
    renderWithProviders(<TutePage />);
    const group = await screen.findByTestId('tute-declarations');
    expect(group.tagName).toBe('FIELDSET');
    // The <legend> names the group (no redundant aria-label).
    expect(group.querySelector('legend')).toHaveTextContent('宣言できる役');
    const tuteBtn = screen.getByTestId('tute-declare-button');
    expect(tuteBtn).toHaveAttribute('aria-describedby', 'tute-help-desc');
    expect(document.getElementById('tute-help-desc')).toHaveTextContent(/Tute/);
  });

  it('does not show declaration buttons when not allowed', async () => {
    renderWithProviders(<TutePage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: /結婚宣言/ })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Tute 宣言' })).not.toBeInTheDocument();
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<TutePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<TutePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<TutePage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたのチームの勝ち！')).toBeInTheDocument());
  });

  it('shows the declared-marriage readout reflecting declaredSuits', async () => {
    // Club (suit 2) declared; the other three suits remain undeclared.
    mockExec.mockResolvedValue(makeTuteState({ declaredSuits: [false, false, true, false, false] }));
    renderWithProviders(<TutePage />);
    const panel = await screen.findByTestId('tute-declared-marriages');
    expect(panel).toHaveTextContent('宣言済みマリッジ');
    // One suit declared, three not declared.
    expect(screen.getAllByText('宣言済')).toHaveLength(1);
    expect(screen.getAllByText('未宣言')).toHaveLength(3);
    expect(panel).toHaveTextContent('切り札スートは +40 点');
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<TutePage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndices: [0], marriage: 0, reason: 'x' } });
    renderWithProviders(<TutePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  // **押したときは出る。**押していない側だけを見ていると、`isRequestedHint` を
  // 定数 false にしても通ってしまう。真の分岐も踏んでおく。
  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      hint: { cardIndices: [0], marriage: 0, reason: 'x' },
      messageCode: 'tute.hintRequested',
    });
    renderWithProviders(<TutePage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });

  it('shows the running round points during play, not only at round end', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<TutePage />);
    // A marriage scores immediately, so the total has to be visible while playing.
    await waitFor(() => expect(screen.getByTestId('tute-running-points')).toBeInTheDocument());
  });
});
