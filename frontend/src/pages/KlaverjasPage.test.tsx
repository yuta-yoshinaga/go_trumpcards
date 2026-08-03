import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { klaverjasApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeKlaverjasState } from '../test/stateFactories';
import { KlaverjasPage } from './KlaverjasPage';

vi.mock('../api/gameApi', () => ({
  klaverjasApi: { exec: vi.fn() },
  actionLogApi: { klaverjas: vi.fn() },
}));

const mockExec = vi.mocked(klaverjasApi.exec);

const playPhaseState = makeKlaverjasState();
const trickEndState = makeKlaverjasState({
  phase: 1,
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 12 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 13 } },
  ],
});
const roundEndState = makeKlaverjasState({ phase: 2, roundCardPoints: [70, 50], roundRoem: [20, 0] });
const gameEndState = makeKlaverjasState({
  phase: 3,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'ゲーム終了！ あなたのチームの勝ち！',
});
const cpuTurnState = makeKlaverjasState({ currentPlayerIdx: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('KlaverjasPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<KlaverjasPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<KlaverjasPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetPoints: 1501 },
      }),
    );
  });

  it('renders the play phase with the human cards', async () => {
    renderWithProviders(<KlaverjasPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<KlaverjasPage />);
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
    renderWithProviders(<KlaverjasPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('shows the winning team in a status banner at trick end', async () => {
    // leadPlayerIdx 0 -> Team A won the trick.
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<KlaverjasPage />);
    const banner = await screen.findByTestId('klaverjas-trick-winner');
    expect(banner).toHaveTextContent('チームA がトリック獲得');
    expect(banner).toHaveAttribute('role', 'status');
    expect(banner).toHaveAttribute('aria-live', 'polite');
  });

  it('attributes the trick to Team B when an odd-seat player leads next', async () => {
    mockExec.mockResolvedValue({ ...trickEndState, leadPlayerIdx: 1 });
    renderWithProviders(<KlaverjasPage />);
    await waitFor(() =>
      expect(screen.getByTestId('klaverjas-trick-winner')).toHaveTextContent('チームB がトリック獲得'),
    );
  });

  it('does not show the trick-winner banner during play', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<KlaverjasPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('klaverjas-trick-winner')).not.toBeInTheDocument();
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<KlaverjasPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
  });

  it('shows live Roem during play', async () => {
    mockExec.mockResolvedValue(makeKlaverjasState({ phase: 0, roundRoem: [40, 20] }));
    renderWithProviders(<KlaverjasPage />);
    const roem = await screen.findByTestId('klaverjas-roem');
    expect(roem).toHaveTextContent('40');
    expect(roem).toHaveTextContent('20');
  });

  it('exposes the live Roem panel as a polite live region', async () => {
    mockExec.mockResolvedValue(makeKlaverjasState({ phase: 0, roundRoem: [20, 0] }));
    renderWithProviders(<KlaverjasPage />);
    const roem = await screen.findByTestId('klaverjas-roem');
    expect(roem).toHaveAttribute('role', 'status');
    expect(roem).toHaveAttribute('aria-live', 'polite');
    // No pulse highlight before any increase.
    expect(roem.className).not.toContain('animate-pulse');
  });

  it('pulses the Roem panel when the combined Roem total increases', async () => {
    mockExec.mockResolvedValue(makeKlaverjasState({ phase: 0, roundRoem: [20, 0] }));
    renderWithProviders(<KlaverjasPage />);
    const card = await screen.findByAltText('♥ Q');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    // The play resolves to a state with higher Roem (20 → 40), triggering the pulse.
    mockExec.mockResolvedValue(makeKlaverjasState({ phase: 0, roundRoem: [40, 0] }));
    fireEvent.click(playBtn);
    await waitFor(() => expect(screen.getByTestId('klaverjas-roem').className).toContain('animate-pulse'));
  });

  it('falls back to 0 Roem when the array is empty', async () => {
    mockExec.mockResolvedValue(makeKlaverjasState({ phase: 0, roundRoem: [] }));
    renderWithProviders(<KlaverjasPage />);
    const roem = await screen.findByTestId('klaverjas-roem');
    expect(roem).toHaveTextContent('A=0');
    expect(roem).toHaveTextContent('B=0');
  });

  it('hides the live Roem block at round end (the round result repeats it)', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<KlaverjasPage />);
    await waitFor(() => expect(screen.getByText('ラウンド結果')).toBeInTheDocument());
    expect(screen.queryByTestId('klaverjas-roem')).not.toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<KlaverjasPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたのチームの勝ち！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<KlaverjasPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndices: [0], reason: 'x' } });
    renderWithProviders(<KlaverjasPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // バナーは推奨札の位置を `([0])` の形で含む。トグルのラベル (「ヒント表示」)
    // と紛れないよう、そこで判定する。
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  // **押したときは出る。**押していない側だけを見ていると、`isRequestedHint` を
  // 定数 false にしても通ってしまう。真の分岐も踏んでおく。
  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      // このページのバナーは `cardIndices` を並べる。`cardIndex` は型に無い。
      hint: { cardIndices: [0], reason: 'x' },
      messageCode: 'klaverjas.hintRequested',
    });
    renderWithProviders(<KlaverjasPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });
});
