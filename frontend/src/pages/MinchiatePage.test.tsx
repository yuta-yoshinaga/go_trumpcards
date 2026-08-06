import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { minchiateApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeMinchiateState } from '../test/stateFactories';
import { MINCHIATE_SURPLUS } from '../types/card';
import { MinchiatePage } from './MinchiatePage';

vi.mock('../api/gameApi', () => ({
  minchiateApi: { exec: vi.fn() },
  actionLogApi: { minchiate: vi.fn() },
}));

const mockExec = vi.mocked(minchiateApi.exec);

const playState = makeMinchiateState();
const scartoState = makeMinchiateState({
  phase: 0,
  isHumanTurn: false,
  isHumanScarto: true,
  scartoCount: 0,
  playableIndices: [],
});
const cpuTurnState = makeMinchiateState({ isHumanTurn: false, currentPlayerIdx: 1, playableIndices: [] });
const trickEndState = makeMinchiateState({ phase: 2, isHumanTurn: false, playableIndices: [] });
const roundEndState = makeMinchiateState({
  phase: 3,
  isHumanTurn: false,
  playableIndices: [],
  roundTricks: [5, 4, 3, 3],
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playState);
});

describe('MinchiatePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<MinchiatePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<MinchiatePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 1, targetRounds: 4 } }),
    );
  });

  // 切札が40枚あることは手札からは読めない唯一の情報なので、常時表示が必要。
  it('always states the 40-trump count', async () => {
    renderWithProviders(<MinchiatePage />);
    const note = await screen.findByTestId('minchiate-trump-count-note');
    expect(note).toHaveTextContent('切札は40枚');
  });

  // 個人戦ではなくチーム戦。対面がパートナーであることが画面に出ていないと、
  // 味方のトリックを奪う手が「正しく」見えてしまう。
  it('shows team scores rather than per-player scores', async () => {
    mockExec.mockResolvedValue(makeMinchiateState({ teamScores: [7, 4] }));
    renderWithProviders(<MinchiatePage />);
    const scores = await screen.findByTestId('minchiate-team-scores');
    expect(scores).toHaveTextContent('7');
    expect(scores).toHaveTextContent('4');
  });

  it('plays the selected card', async () => {
    renderWithProviders(<MinchiatePage />);
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
    renderWithProviders(<MinchiatePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  describe('scarto phase', () => {
    // 余剰ぶんの札を持つ手を組む。**枚数は定数から。**Tarocchini から写すと
    // 2 枚のままになるが、Minchiate は 13 枚選ばないと捨てられない。
    const scartoHandState = () =>
      makeMinchiateState({
        ...scartoState,
        players: [
          {
            ...scartoState.players[0],
            cardCount: MINCHIATE_SURPLUS,
            cards: Array.from({ length: MINCHIATE_SURPLUS }, (_, i) => ({
              design: 'SPADE' as const,
              value: i + 1,
              glyph: '\u2660',
              label: String(i + 1),
              color: 'black',
              deck: 'tarot',
            })),
          },
          ...scartoState.players.slice(1),
        ],
      });

    it('prompts for the full surplus and dispatches every index', async () => {
      mockExec.mockResolvedValue(scartoHandState());
      renderWithProviders(<MinchiatePage />);
      expect(await screen.findByTestId('minchiate-scarto-prompt')).toHaveTextContent(String(MINCHIATE_SURPLUS));

      expect(screen.getByRole('button', { name: '捨てる' })).toBeDisabled();

      // 1 枚足りないうちは押せない。
      for (let i = 1; i < MINCHIATE_SURPLUS; i++) {
        fireEvent.click(screen.getByRole('button', { name: `${i} \u2660` }));
      }
      expect(screen.getByRole('button', { name: '捨てる' })).toBeDisabled();

      fireEvent.click(screen.getByRole('button', { name: `${MINCHIATE_SURPLUS} \u2660` }));
      mockExec.mockClear();
      mockExec.mockResolvedValue(scartoHandState());
      fireEvent.click(screen.getByRole('button', { name: '捨てる' }));
      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith('scarto', {
          cardIndices: Array.from({ length: MINCHIATE_SURPLUS }, (_, i) => i),
        }),
      );
    });

    it('shows no scarto control when the dealer is a CPU', async () => {
      mockExec.mockResolvedValue(makeMinchiateState({ phase: 0, isHumanScarto: false, isHumanTurn: false }));
      renderWithProviders(<MinchiatePage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());
      expect(screen.queryByRole('button', { name: '捨てる' })).not.toBeInTheDocument();
    });
  });

  it('advances to the next trick', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<MinchiatePage />);
    const button = await screen.findByRole('button', { name: '次のトリック' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(trickEndState);
    fireEvent.click(button);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('advances to the next round and shows the trick tally', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<MinchiatePage />);
    const button = await screen.findByRole('button', { name: '次のラウンド' });
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
    mockExec.mockClear();
    mockExec.mockResolvedValue(roundEndState);
    fireEvent.click(button);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('has no bid controls — Minchiate has no bidding phase', async () => {
    renderWithProviders(<MinchiatePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: 'パス' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-0')).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playState, hint: { cardIndices: [0], reason: 'lead_trump' } });
    renderWithProviders(<MinchiatePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  // **押したときは出る。**押していない側だけを見ていると、`isRequestedHint` を
  // 定数 false にしても通ってしまう。真の分岐も踏んでおく。
  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...playState,
      hint: { cardIndices: [0], reason: 'lead_trump' },
      messageCode: 'minchiate.hintRequested',
    });
    renderWithProviders(<MinchiatePage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });

  // **CUI と CLI からは呼べるのに、盤面には要求する手段が無かった (#4819)。**
  // isRequestedHint は hintRequested の messageCode を待つので、コマンドを
  // 送らないと表示側は永遠に出ない (実質デッドコードだった)。
  it('asks the server for a hint and shows the answer', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<MinchiatePage />);
    const hintBtn = await screen.findByTestId('minchiate-hint-button');

    mockExec.mockClear();
    mockExec.mockResolvedValue({
      ...playState,
      messageCode: 'minchiate.hintRequested',
      hint: { cardIndices: [2], reason: 'lead_trump' },
    });
    fireEvent.click(hintBtn);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
    await waitFor(() => expect(screen.getByText(/\[2\]/)).toBeInTheDocument());
  });

  it('shows no hint line until one is requested', async () => {
    mockExec.mockResolvedValue({ ...playState, hint: { cardIndices: [2], reason: 'lead_trump' } });
    renderWithProviders(<MinchiatePage />);
    await waitFor(() => expect(screen.getByTestId('minchiate-hint-button')).toBeInTheDocument());
    expect(screen.queryByText(/\[2\]/)).not.toBeInTheDocument();
  });
});
