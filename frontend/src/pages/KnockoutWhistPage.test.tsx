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
  localStorage.clear();
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

  it('previews the next round hand size (one fewer card) and trump chooser at round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<KnockoutWhistPage />);
    const preview = await screen.findByTestId('kw-next-round-preview');
    // Default round-end hand size is 7, so the next round deals 6 cards; winner idx 0 is the human.
    expect(preview).toHaveTextContent('次ラウンド: 手札6枚 / 切り札選択: あなた');
  });

  it('flags the final round when the next hand size bottoms out at 1', async () => {
    mockExec.mockResolvedValue(makeKnockoutWhistState({ phase: 2, roundWinnerIdx: 1, handSize: 2 }));
    renderWithProviders(<KnockoutWhistPage />);
    const preview = await screen.findByTestId('kw-next-round-preview');
    expect(preview).toHaveTextContent('最終ラウンド: 手札1枚');
  });

  it('does not preview the next round at game end', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<KnockoutWhistPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());
    expect(screen.queryByTestId('kw-next-round-preview')).not.toBeInTheDocument();
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

  // Build a state whose human (seat 0) is knocked out, using a fresh players array so the
  // shared base fixture (referenced by makeKnockoutWhistState's shallow spread) is not mutated.
  const eliminatedHumanState = (overrides?: Parameters<typeof makeKnockoutWhistState>[0]) => {
    const base = makeKnockoutWhistState(overrides);
    return {
      ...base,
      players: base.players.map((p) => (p.isHuman ? { ...p, eliminated: true, dogbones: 0 } : p)),
    };
  };

  it('shows the spectator banner while the human is eliminated and the match continues', async () => {
    mockExec.mockResolvedValue(eliminatedHumanState({ isHumanTurn: false, currentPlayerIdx: 1, activeCount: 2 }));
    renderWithProviders(<KnockoutWhistPage />);
    const banner = await screen.findByTestId('kw-spectator-banner');
    expect(banner).toHaveAttribute('role', 'status');
    expect(banner).toHaveTextContent('観戦中');
    expect(banner).toHaveTextContent('残り 2人');
  });

  it('does not show the spectator banner while the human is still active', async () => {
    renderWithProviders(<KnockoutWhistPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByTestId('kw-spectator-banner')).not.toBeInTheDocument();
  });

  it('does not show the spectator banner once the game has ended', async () => {
    mockExec.mockResolvedValue(eliminatedHumanState({ phase: 3, gameEndFlag: true, winnerPlayer: 1, activeCount: 1 }));
    renderWithProviders(<KnockoutWhistPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('kw-spectator-banner')).not.toBeInTheDocument();
  });

  it('renders the leader badge with an opaque, high-contrast surface', async () => {
    // Default state: leadPlayerIdx 0 → the human is the leader.
    renderWithProviders(<KnockoutWhistPage />);
    const badge = await screen.findByText('リーダー');
    // Opaque surface token (badgeInfoColors) instead of the old translucent bg-white/20.
    expect(badge.className).toContain('bg-ds-surface');
    expect(badge.className).not.toContain('bg-white/20');
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndices: [0], reason: 'x' } });
    renderWithProviders(<KnockoutWhistPage />);
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
      hint: { cardIndex: 0, reason: 'x' },
      messageCode: 'knockoutWhist.hintRequested',
    });
    renderWithProviders(<KnockoutWhistPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });
});
