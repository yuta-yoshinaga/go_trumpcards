import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { aluetteApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeAluetteState } from '../test/stateFactories';
import { AluettePage } from './AluettePage';

vi.mock('../api/gameApi', () => ({
  aluetteApi: { exec: vi.fn() },
  actionLogApi: { aluette: vi.fn() },
}));

const mockExec = vi.mocked(aluetteApi.exec);

const playState = makeAluetteState();
const cpuTurnState = makeAluetteState({ isHumanTurn: false, currentPlayerIdx: 1, playableIndices: [] });
const trickEndState = makeAluetteState({ phase: 1, isHumanTurn: false, playableIndices: [] });
const roundEndState = makeAluetteState({
  phase: 2,
  isHumanTurn: false,
  playableIndices: [],
  roundTricks: [3, 1, 0, 1],
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playState);
});

describe('AluettePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<AluettePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<AluettePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 1, targetPoints: 5 } }),
    );
  });

  // **序列表は常時表示。**強さは値ではなく札ごとに決まるので、6 枚を覚えるまでは
  // 手札を強さ順に並べることすらできない。
  it('always shows the luette table, strongest first', async () => {
    renderWithProviders(<AluettePage />);
    const legend = await screen.findByTestId('aluette-luettes');
    expect(legend).toHaveTextContent('ムッシュー');
    expect(legend).toHaveTextContent('プチ・ヌフ');
    const text = legend.textContent ?? '';
    expect(text.indexOf('ムッシュー')).toBeLessThan(text.indexOf('プチ・ヌフ'));
  });

  // 序列表はサーバーが送る。空で来ても落ちてはならない。
  it('survives a response with no luette table', async () => {
    mockExec.mockResolvedValue(makeAluetteState({ luettes: [] }));
    renderWithProviders(<AluettePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByTestId('aluette-luettes')).toBeInTheDocument();
  });

  // **手札のリュエットは名指しする。**カード画像だけでは ♦3 が最強札とは読めない。
  it('names the luettes the human is holding', async () => {
    renderWithProviders(<AluettePage />);
    const held = await screen.findByTestId('aluette-hand-luettes');
    expect(held).toHaveTextContent('[0] ムッシュー');
    expect(held).toHaveTextContent('[2] グラン・ヌフ');
    // 同じ 3 でも剣の3 はリュエットではない。値だけで名前を付けてはならない。
    expect(held).not.toHaveTextContent('[1]');
  });

  it('says so when the hand holds no luette', async () => {
    mockExec.mockResolvedValue(
      makeAluetteState({
        players: [
          {
            ...playState.players[0],
            cards: [
              { design: 'SPADE', value: 3 },
              { design: 'CLOVER', value: 2 },
            ],
            cardCount: 2,
          },
          ...playState.players.slice(1),
        ],
      }),
    );
    renderWithProviders(<AluettePage />);
    expect(await screen.findByText('手札にリュエットはありません')).toBeInTheDocument();
  });

  // 個人戦ではなくチーム戦。対面がパートナーであることが画面に出ていないと、
  // 味方のトリックを奪う手が「正しく」見えてしまう。
  it('shows team scores rather than per-player scores', async () => {
    mockExec.mockResolvedValue(makeAluetteState({ teamScores: [3, 1] }));
    renderWithProviders(<AluettePage />);
    const scores = await screen.findByTestId('aluette-team-scores');
    expect(scores).toHaveTextContent('3');
    expect(scores).toHaveTextContent('1');
  });

  it('plays the selected card', async () => {
    renderWithProviders(<AluettePage />);
    const playButton = await screen.findByRole('button', { name: '出す' });
    expect(playButton).toBeDisabled();

    fireEvent.click(screen.getByRole('button', { name: '♦ 3' }));
    mockExec.mockClear();
    mockExec.mockResolvedValue(playState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('hides the play control when it is not the human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<AluettePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('advances to the next trick', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<AluettePage />);
    const button = await screen.findByRole('button', { name: '次のトリック' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(trickEndState);
    fireEvent.click(button);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('advances to the next mene and shows the trick tally', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<AluettePage />);
    const button = await screen.findByRole('button', { name: '次のメーヌ' });
    expect(screen.getByText('メーヌ結果')).toBeInTheDocument();
    mockExec.mockClear();
    mockExec.mockResolvedValue(roundEndState);
    fireEvent.click(button);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  // **捨て札も入札もこのゲームには無い。**タロー系から写した痕跡が残っていないこと。
  it('has neither bidding nor a discard control', async () => {
    renderWithProviders(<AluettePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: 'パス' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '捨てる' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-0')).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playState, hint: { cardIndices: [0], reason: 'play_luette' } });
    renderWithProviders(<AluettePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  // **押したときは出る。**押していない側だけを見ていると、`isRequestedHint` を
  // 定数 false にしても通ってしまう。真の分岐も踏んでおく。
  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...playState,
      hint: { cardIndices: [0], reason: 'play_luette' },
      messageCode: 'aluette.hintRequested',
    });
    renderWithProviders(<AluettePage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });
});
