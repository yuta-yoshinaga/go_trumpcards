import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tarocchiniApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeTarocchiniState } from '../test/stateFactories';
import { TarocchiniPage } from './TarocchiniPage';

vi.mock('../api/gameApi', () => ({
  tarocchiniApi: { exec: vi.fn() },
  actionLogApi: { tarocchini: vi.fn() },
}));

const mockExec = vi.mocked(tarocchiniApi.exec);

const playState = makeTarocchiniState();
const scartoState = makeTarocchiniState({
  phase: 0,
  isHumanTurn: false,
  isHumanScarto: true,
  scartoCount: 0,
  playableIndices: [],
});
const cpuTurnState = makeTarocchiniState({ isHumanTurn: false, currentPlayerIdx: 1, playableIndices: [] });
const trickEndState = makeTarocchiniState({ phase: 2, isHumanTurn: false, playableIndices: [] });
const roundEndState = makeTarocchiniState({
  phase: 3,
  isHumanTurn: false,
  playableIndices: [],
  roundTricks: [5, 4, 3, 3],
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playState);
});

describe('TarocchiniPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<TarocchiniPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<TarocchiniPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 1, targetRounds: 4 } }),
    );
  });

  // 後出し優先は手札からは読めない唯一の情報なので、常時表示されている必要がある。
  it('always states the papi rule', async () => {
    renderWithProviders(<TarocchiniPage />);
    const note = await screen.findByTestId('tarocchini-papi-note');
    expect(note).toHaveTextContent('後から出した方が勝ち');
  });

  // 個人戦ではなくチーム戦。対面がパートナーであることが画面に出ていないと、
  // 味方のトリックを奪う手が「正しく」見えてしまう。
  it('shows team scores rather than per-player scores', async () => {
    mockExec.mockResolvedValue(makeTarocchiniState({ teamScores: [7, 4] }));
    renderWithProviders(<TarocchiniPage />);
    const scores = await screen.findByTestId('tarocchini-team-scores');
    expect(scores).toHaveTextContent('7');
    expect(scores).toHaveTextContent('4');
  });

  it('plays the selected card', async () => {
    renderWithProviders(<TarocchiniPage />);
    const playButton = await screen.findByRole('button', { name: '出す' });
    expect(playButton).toBeDisabled();

    fireEvent.click(screen.getByRole('button', { name: 'Re ♠' }));
    mockExec.mockClear();
    mockExec.mockResolvedValue(playState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('hides the play control when it is not the human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<TarocchiniPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  describe('scarto phase', () => {
    it('prompts for exactly two cards and dispatches both', async () => {
      mockExec.mockResolvedValue(scartoState);
      renderWithProviders(<TarocchiniPage />);
      expect(await screen.findByTestId('tarocchini-scarto-prompt')).toHaveTextContent('2');

      const bury = screen.getByRole('button', { name: '捨てる' });
      expect(bury).toBeDisabled();

      // 1 枚だけでは足りない。
      fireEvent.click(screen.getByRole('button', { name: 'Re ♠' }));
      expect(screen.getByRole('button', { name: '捨てる' })).toBeDisabled();

      fireEvent.click(screen.getByRole('button', { name: '20 ✦' }));
      mockExec.mockClear();
      mockExec.mockResolvedValue(scartoState);
      fireEvent.click(screen.getByRole('button', { name: '捨てる' }));
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('scarto', { cardIndices: [0, 1] }));
    });

    it('shows no scarto control when the dealer is a CPU', async () => {
      mockExec.mockResolvedValue(makeTarocchiniState({ phase: 0, isHumanScarto: false, isHumanTurn: false }));
      renderWithProviders(<TarocchiniPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());
      expect(screen.queryByRole('button', { name: '捨てる' })).not.toBeInTheDocument();
    });
  });

  it('advances to the next trick', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<TarocchiniPage />);
    const button = await screen.findByRole('button', { name: '次のトリック' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(trickEndState);
    fireEvent.click(button);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('advances to the next round and shows the trick tally', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<TarocchiniPage />);
    const button = await screen.findByRole('button', { name: '次のラウンド' });
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
    mockExec.mockClear();
    mockExec.mockResolvedValue(roundEndState);
    fireEvent.click(button);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('has no bid controls — Tarocchini has no bidding phase', async () => {
    renderWithProviders(<TarocchiniPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: 'パス' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-0')).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playState, hint: { cardIndices: [0], reason: 'play_papa' } });
    renderWithProviders(<TarocchiniPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  // **押したときは出る。**押していない側だけを見ていると、`isRequestedHint` を
  // 定数 false にしても通ってしまう。真の分岐も踏んでおく。
  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...playState,
      hint: { cardIndices: [0], reason: 'play_papa' },
      messageCode: 'tarocchini.hintRequested',
    });
    renderWithProviders(<TarocchiniPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });

  // **CUI と CLI からは呼べるのに、盤面には要求する手段が無かった (#4820)。**
  // isRequestedHint は hintRequested の messageCode を待つので、コマンドを
  // 送らないと表示側は永遠に出ない (実質デッドコードだった)。
  it('asks the server for a hint and shows the answer', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<TarocchiniPage />);
    const hintBtn = await screen.findByTestId('tarocchini-hint-button');

    mockExec.mockClear();
    mockExec.mockResolvedValue({
      ...playState,
      messageCode: 'tarocchini.hintRequested',
      hint: { cardIndices: [2], reason: 'lead_trump' },
    });
    fireEvent.click(hintBtn);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
    await waitFor(() => expect(screen.getByText(/\[2\]/)).toBeInTheDocument());
  });

  it('shows no hint line until one is requested', async () => {
    // hint が載っていても messageCode が無ければ出さない (常時表示にしない)。
    mockExec.mockResolvedValue({ ...playState, hint: { cardIndices: [2], reason: 'lead_trump' } });
    renderWithProviders(<TarocchiniPage />);
    await waitFor(() => expect(screen.getByTestId('tarocchini-hint-button')).toBeInTheDocument());
    expect(screen.queryByText(/\[2\]/)).not.toBeInTheDocument();
  });
});
