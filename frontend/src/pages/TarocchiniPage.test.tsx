import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tarocchiniApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
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

  it('renders CPU difficulty as a select and match length as a number input', async () => {
    renderWithProviders(<TarocchiniPage />);
    fireEvent.click(await screen.findByText('設定'));

    expect(screen.getByRole('combobox', { name: 'CPU難易度' })).toBeInTheDocument();
    expect(screen.getByRole('spinbutton', { name: 'ラウンド数' })).toBeInTheDocument();
  });

  it('accepts multiples of four for the target rounds', async () => {
    renderWithProviders(<TarocchiniPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'Re ♠' })).toBeInTheDocument());
    const rounds = screen.getByLabelText('ラウンド数') as HTMLInputElement;
    expect(rounds).toHaveAttribute('min', '4');
    expect(rounds).toHaveAttribute('step', '4');
    expect(screen.getByText('4以上の4の倍数を指定してください。')).toBeInTheDocument();
    fireEvent.change(rounds, { target: { value: '16' } });
    expect(rounds.value).toBe('16');
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

  // #5716: ラウンド得点は「トリック数 + 最終トリック +2 + スカルト加点」の合算。
  // トリック数だけでは teamScores の増分と突き合わせて検算できなかった。
  it('breaks the round score down into its three parts', async () => {
    mockExec.mockResolvedValue(
      makeTarocchiniState({
        phase: 3,
        isHumanTurn: false,
        playableIndices: [],
        roundTricks: [5, 4, 3, 3],
        lastTrickWinner: 1, // 席1 = チーム1
        dealerIdx: 0, // 席0 = チーム0
        scartoCount: 2,
        roundBreakdown: [
          { tricks: 8, lastTrickBonus: 0, scartoBonus: 2, total: 10 },
          { tricks: 5, lastTrickBonus: 2, scartoBonus: 0, total: 7 },
        ],
      }),
    );
    renderWithProviders(<TarocchiniPage />);

    const lastTrick = await screen.findByTestId('tarocchini-last-trick-bonus');
    expect(lastTrick).toHaveTextContent('チーム1');
    expect(lastTrick).toHaveTextContent('+2');
    const scarto = screen.getByTestId('tarocchini-scarto-bonus');
    expect(scarto).toHaveTextContent('チーム0');
    expect(scarto).toHaveTextContent('+2');
    expect(screen.getByTestId('tarocchini-team-round-breakdown-0')).toHaveTextContent(
      'チーム0: トリック 8 / 最終トリック +0 / スカルト +2 / 計 10',
    );
    expect(screen.getByTestId('tarocchini-team-round-breakdown-1')).toHaveTextContent(
      'チーム1: トリック 5 / 最終トリック +2 / スカルト +0 / 計 7',
    );
  });

  it('omits the breakdown lines that do not apply', async () => {
    mockExec.mockResolvedValue(
      makeTarocchiniState({
        phase: 3,
        isHumanTurn: false,
        playableIndices: [],
        roundTricks: [5, 4, 3, 3],
        lastTrickWinner: -1,
        scartoCount: 0,
      }),
    );
    renderWithProviders(<TarocchiniPage />);
    await screen.findByText('ラウンド結果');

    expect(screen.queryByTestId('tarocchini-last-trick-bonus')).not.toBeInTheDocument();
    expect(screen.queryByTestId('tarocchini-scarto-bonus')).not.toBeInTheDocument();
  });

  // ディーラーによるスカルト（不要札の捨て札）が完了している場合のみ捨て札枚数を表示する
  it('renders tarocchini-scarto-done when scartoCount > 0 and hides it when 0', async () => {
    mockExec.mockResolvedValue(scartoState);
    const { unmount } = renderWithProviders(<TarocchiniPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('tarocchini-scarto-done')).not.toBeInTheDocument();
    unmount();

    mockExec.mockResolvedValue(playState);
    renderWithProviders(<TarocchiniPage />);
    await waitFor(() => expect(screen.getByTestId('tarocchini-scarto-done')).toBeInTheDocument());
    expect(screen.getByTestId('tarocchini-scarto-done')).toHaveTextContent('スカルト済み: 2枚');
  });

  describe('scarto phase', () => {
    it('disables trumps and the Matto with a reason while allowing ordinary cards', async () => {
      const scartoHand = makeTarocchiniState({
        ...scartoState,
        players: scartoState.players.map((player) =>
          player.isHuman
            ? {
                ...player,
                cards: [
                  ...player.cards,
                  {
                    design: 'JOKER' as const,
                    value: 0,
                    glyph: '★',
                    label: 'Matto',
                    color: 'gold',
                    deck: 'tarot' as const,
                  },
                  {
                    design: 'HEART' as const,
                    value: 14,
                    glyph: '♥',
                    label: 'Re',
                    color: 'red',
                    deck: 'tarot' as const,
                  },
                ],
              }
            : player,
        ),
      });
      mockExec.mockResolvedValue(scartoHand);
      renderWithProviders(<TarocchiniPage />);

      const cards = await screen.findAllByRole('button', { name: /^(Re ♠|20 ✦|Papa ✦|Matto ★|Re ♥)$/ });
      expect(cards[0]).not.toHaveAttribute('aria-disabled');
      expect(cards[1]).toHaveAttribute('aria-disabled', 'true');
      expect(cards[1]).toHaveAttribute('title', '切札とマットはスカルトできません');
      expect(cards[2]).toHaveAttribute('aria-disabled', 'true');
      expect(cards[3]).toHaveAttribute('aria-disabled', 'true');
      expect(cards[4]).not.toHaveAttribute('aria-disabled');

      fireEvent.click(cards[1]);
      fireEvent.click(cards[2]);
      fireEvent.click(cards[3]);
      fireEvent.click(cards[0]);
      fireEvent.click(cards[4]);
      expect(screen.getByTestId('tarocchini-scarto-prompt')).toHaveTextContent('2/2');
      await flushPendingDispatch();
      expect(mockExec).not.toHaveBeenCalledWith('scarto', expect.anything());

      fireEvent.click(screen.getByRole('button', { name: '捨てる' }));
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('scarto', { cardIndices: [0, 4] }));
    });

    it('prompts for exactly two cards and dispatches both', async () => {
      // Keep this interaction test's two selected indices legal ordinary cards.
      // The fixture's JOKER cards model Tarocchini trumps and cannot be buried.
      const ordinaryScartoState = makeTarocchiniState({
        ...scartoState,
        players: scartoState.players.map((player) =>
          player.isHuman
            ? {
                ...player,
                cards: player.cards.map((card, index) =>
                  index === 1
                    ? { design: 'HEART' as const, value: 13, glyph: '✦', label: '20', color: 'red', deck: 'tarot' }
                    : card,
                ),
              }
            : player,
        ),
      });
      mockExec.mockResolvedValue(ordinaryScartoState);
      renderWithProviders(<TarocchiniPage />);
      expect(await screen.findByTestId('tarocchini-scarto-prompt')).toHaveTextContent('2');
      expect(screen.getByTestId('tarocchini-scarto-prompt')).toHaveTextContent('0/2');

      const bury = screen.getByRole('button', { name: '捨てる' });
      expect(bury).toBeDisabled();

      // 1 枚だけでは足りない。
      fireEvent.click(screen.getByRole('button', { name: 'Re ♠' }));
      expect(screen.getByRole('button', { name: '捨てる' })).toBeDisabled();
      expect(screen.getByTestId('tarocchini-scarto-prompt')).toHaveTextContent('1/2');

      fireEvent.click(screen.getByRole('button', { name: '20 ✦' }));
      mockExec.mockClear();
      mockExec.mockResolvedValue(ordinaryScartoState);
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

  // 催促はフェーズや手番が変わったときに現れるテキスト。領域が無いと、あるいは
  // 領域が催促と**同時に**生えると、スクリーンリーダには何も届かない (#6880)。
  it('announces the tarocchini-scarto-prompt from the always-mounted live region', async () => {
    mockExec.mockResolvedValue(scartoState);
    renderWithProviders(<TarocchiniPage />);

    const live = await screen.findByTestId('tarocchini-prompt-live');
    expect(live).toHaveAttribute('role', 'status');
    expect(live).toHaveAttribute('aria-live', 'polite');
    // 隣に置いただけの実装は属性の検査を通る。**中にあること**を見る。
    expect(live).toContainElement(await screen.findByTestId('tarocchini-scarto-prompt'));
  });
});
