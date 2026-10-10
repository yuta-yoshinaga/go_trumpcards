import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { beziqueApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeBeziqueState } from '../test/stateFactories';
import { BeziquePage } from './BeziquePage';

vi.mock('../api/gameApi', () => ({
  beziqueApi: { exec: vi.fn() },
  actionLogApi: { bezique: vi.fn() },
}));

const mockExec = vi.mocked(beziqueApi.exec);

// Default fixture: a human Play turn (seat 0, phase 1=Play in domain terms = 0 here).
const playPhaseState = makeBeziqueState({ phase: 0, currentPlayerIdx: 0 });
// A CPU turn (currentPlayerIdx is the CPU).
const cpuTurnState = makeBeziqueState({ phase: 0, currentPlayerIdx: 1 });
// A human Meld turn with two declarable melds.
const meldPhaseState = makeBeziqueState({
  phase: 1,
  currentPlayerIdx: 0,
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 4,
      cards: [
        { design: 'SPADE', value: 13 },
        { design: 'SPADE', value: 12 },
        { design: 'DIAMOND', value: 11 },
        { design: 'HEART', value: 1 },
      ],
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    { id: 1, isHuman: false, cardCount: 9, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
  ],
  availableMelds: [
    { type: 0, suit: 1, points: 20, cardIndices: [0, 1] },
    { type: 1, suit: -1, points: 40, cardIndices: [1, 2] },
  ],
});
const roundEndState = makeBeziqueState({
  phase: 2,
  dealPoints: [240, 160],
});
const gameEndState = makeBeziqueState({
  phase: 3,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'ゲーム終了！ あなたの勝利です (1010-820)！',
});
const endgameState = makeBeziqueState({ phase: 0, currentPlayerIdx: 0, stockRemaining: 0, isEndgame: true });
// Endgame human FOLLOW turn: opponent led ♠K; hand is ♠Q / ♦J / ♥A, trump ♥.
// Holding ♠ forces following ♠ (♠Q cannot beat ♠K, so it is the only legal card).
const endgameFollowState = makeBeziqueState({
  phase: 0,
  currentPlayerIdx: 0,
  stockRemaining: 0,
  isEndgame: true,
  trumpSuit: 3,
  currentTrick: [{ playerIdx: 1, card: { design: 'SPADE', value: 13 } }],
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('BeziquePage', () => {
  it('shows the translated no-trump label before a trump is declared', async () => {
    mockExec.mockResolvedValue(makeBeziqueState({ trumpSuit: 0 }));
    renderWithProviders(<BeziquePage />);
    expect(await screen.findByText('切り札: 未宣言')).toBeInTheDocument();
  });

  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<BeziquePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<BeziquePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { targetScore: 1000 },
      }),
    );
  });

  it('shows the deal-points trick/meld breakdown in the score sidebar', async () => {
    mockExec.mockResolvedValue(
      makeBeziqueState({ phase: 0, currentPlayerIdx: 0, dealPoints: [10, 0], dealMeldPoints: [4, 0] }),
    );
    renderWithProviders(<BeziquePage />);
    const breakdown = await screen.findByTestId('bezique-deal-breakdown-0');
    // trick = deal(10) - meld(4) = 6
    expect(breakdown).toHaveTextContent('6');
    expect(breakdown).toHaveTextContent('4');
  });

  it('shows who received the last-trick bonus in the deal result and omits it when absent', async () => {
    mockExec.mockResolvedValue(
      makeBeziqueState({ phase: 2, dealPoints: [10, 30], dealMeldPoints: [0, 0], lastTrickBonus: [0, 10] }),
    );
    renderWithProviders(<BeziquePage />);
    expect(await screen.findByTestId('bezique-last-trick-bonus-1')).toHaveTextContent('（うち最終トリック加点 10点）');
    expect(screen.queryByTestId('bezique-last-trick-bonus-0')).not.toBeInTheDocument();
  });

  it('does not show a last-trick bonus line when no bonus was scored', async () => {
    mockExec.mockResolvedValue(makeBeziqueState({ phase: 2, lastTrickBonus: [0, 0] }));
    renderWithProviders(<BeziquePage />);
    expect(await screen.findByText('ディール結果（獲得ポイント）')).toBeInTheDocument();
    expect(screen.queryByTestId('bezique-last-trick-bonus-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('bezique-last-trick-bonus-1')).not.toBeInTheDocument();
  });

  it('shows the literal trick point table', async () => {
    mockExec.mockResolvedValue(makeBeziqueState());
    renderWithProviders(<BeziquePage />);
    const points = await screen.findByTestId('bezique-trick-points');
    expect(points).toHaveTextContent('トリック得点: A=11 / 10=10 / K=4 / Q=3 / J=2 / その他=0');
  });

  it('renders the play phase with the human cards and the play button', async () => {
    renderWithProviders(<BeziquePage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♦ J')).toBeInTheDocument();
    });
    expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument();
  });

  it('dispatches play when a card is selected and the play button clicked', async () => {
    renderWithProviders(<BeziquePage />);
    const card = await screen.findByAltText('♠ Q');
    fireEvent.click(card);
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<BeziquePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('shows meld buttons and a skip button on a human meld turn', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BeziquePage />);
    await waitFor(() => expect(screen.getByTestId('meld-0')).toBeInTheDocument());
    expect(screen.getByTestId('meld-1')).toBeInTheDocument();
    expect(screen.getByTestId('meld-skip')).toBeInTheDocument();
  });

  // #5657: マッチ進捗は role="status" で読み上げられるのに、**メルド宣言フェーズに
  // 入ったこと自体**を知らせるライブリージョンが無かった。選べる状態になったのに
  // 画面を都度確かめるしかない。
  it('announces how many melds became declarable', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BeziquePage />);

    const live = await screen.findByTestId('bezique-meld-live');
    expect(live).toHaveAttribute('role', 'status');
    expect(live).toHaveAttribute('aria-live', 'polite');
    expect(live).toHaveTextContent('2');
    expect(screen.getAllByText('宣言できるメルドが2件あります')).toHaveLength(1);
  });

  it('announces meld availability through the atomic polite live region', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    const { container } = renderWithProviders(<BeziquePage />);

    const live = await waitFor(() => {
      const element = container.querySelector('[data-testid="bezique-meld-live"][role="status"][aria-live="polite"]');
      expect(element).not.toBeNull();
      return element;
    });
    expect(live).toHaveTextContent('2');
  });

  // 0件でもプレイヤーが選ぶ必要があるため通知する。
  it('announces that nothing can be declared when the list is empty', async () => {
    mockExec.mockResolvedValue(makeBeziqueState({ ...meldPhaseState, availableMelds: [] }));
    renderWithProviders(<BeziquePage />);

    const live = await screen.findByTestId('bezique-meld-live');
    expect(live).toHaveTextContent('宣言できるメルドはありません');
    // スキップは引き続き押せる。
    expect(screen.getByTestId('meld-skip')).toBeInTheDocument();
  });

  it('does not announce meld availability outside the human meld turn', async () => {
    mockExec.mockResolvedValue(
      makeBeziqueState({ phase: 1, currentPlayerIdx: 1, availableMelds: meldPhaseState.availableMelds }),
    );
    renderWithProviders(<BeziquePage />);

    const live = await screen.findByTestId('bezique-meld-live');
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(live).toBeEmptyDOMElement();
  });

  it('gives each meld button a suit-named aria-label inside a labelled group', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BeziquePage />);
    // Marriage of ♠ (type 0, suit 1) reads the suit by name, not the glyph.
    const marriage = await screen.findByTestId('meld-0');
    expect(marriage).toHaveAttribute('aria-label', 'スペードの結婚 (K+Q)を選択 +20点');
    // A non-suited meld (bezique) omits the suit.
    expect(screen.getByTestId('meld-1')).toHaveAttribute('aria-label', 'ベジーク (♠Q+♦J)を選択 +40点');
    // The melds are bundled in a labelled group.
    const group = screen.getByRole('group', { name: '宣言するメルドを選んでください:' });
    expect(group).toBeInTheDocument();
    expect(group).toContainElement(marriage);
  });

  it('dispatches a meld declaration when a meld button is clicked', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BeziquePage />);
    const meld1 = await screen.findByTestId('meld-1');
    fireEvent.click(meld1);
    fireEvent.click(await screen.findByTestId('meld-declare-selected'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(meldPhaseState);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('meld', { meldIndex: 1 }));
  });

  it('highlights and describes only the selected meld cards, updating on selection', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BeziquePage />);
    const marriage = await screen.findByTestId('meld-0');
    const bezique = screen.getByTestId('meld-1');
    fireEvent.click(marriage);
    expect(marriage).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: /♠ Q.*選択中のメルドの構成札/ })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /♥ A/ })).not.toHaveAccessibleName(/選択中のメルド/);
    fireEvent.click(bezique);
    expect(bezique).toHaveAttribute('aria-pressed', 'true');
    expect(marriage).toHaveAttribute('aria-pressed', 'false');
    expect(screen.getByRole('button', { name: /♦ J.*選択中のメルドの構成札/ })).toBeInTheDocument();
  });

  it('clears the selected meld when the available meld list changes', async () => {
    mockExec.mockResolvedValueOnce(meldPhaseState);
    renderWithProviders(<BeziquePage />);
    const marriage = await screen.findByTestId('meld-0');
    fireEvent.click(marriage);
    expect(marriage).toHaveAttribute('aria-pressed', 'true');
    mockExec.mockResolvedValueOnce(makeBeziqueState({ ...meldPhaseState, availableMelds: [] }));
    fireEvent.click(screen.getByTestId('meld-skip'));
    await waitFor(() => expect(marriage).not.toHaveAttribute('aria-pressed', 'true'));
    expect(screen.queryByRole('button', { name: /選択中のメルドの構成札/ })).not.toBeInTheDocument();
    expect(screen.queryByTestId('meld-declare-selected')).not.toBeInTheDocument();
  });

  it('dispatches skip when the skip-meld button is clicked', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BeziquePage />);
    const skip = await screen.findByTestId('meld-skip');
    mockExec.mockClear();
    mockExec.mockResolvedValue(meldPhaseState);
    fireEvent.click(skip);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skip'));
  });

  it('renders round end with the next deal button and the deal result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<BeziquePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のディール' })).toBeInTheDocument());
    expect(screen.getByText('ディール結果（獲得ポイント）')).toBeInTheDocument();
  });

  it('shows the endgame (phase 2) notice once the stock is empty', async () => {
    mockExec.mockResolvedValue(endgameState);
    renderWithProviders(<BeziquePage />);
    await waitFor(() => expect(screen.getByText(/フェーズ2/)).toBeInTheDocument());
  });

  it('announces the endgame once when the stock runs out, but not on the initial render', async () => {
    mockExec.mockResolvedValueOnce(playPhaseState).mockResolvedValue(endgameState);
    renderWithProviders(<BeziquePage />);

    const live = await screen.findByTestId('bezique-endgame-live');
    expect(live).toHaveAttribute('role', 'status');
    expect(live).toHaveAttribute('aria-live', 'polite');
    expect(live).toBeEmptyDOMElement();

    fireEvent.click(await screen.findByAltText('♠ Q'));
    fireEvent.click(screen.getByRole('button', { name: '出す' }));

    await waitFor(() => expect(screen.getByTestId('bezique-endgame-live')).toHaveTextContent('エンドゲーム開始'));
    const announced = screen.getByTestId('bezique-endgame-live');
    expect(announced).toHaveTextContent('マストフォロー');
    expect(announced).toHaveTextContent('メルドは宣言できません');
    expect(announced).toHaveTextContent('最終トリックは+10点');

    mockExec.mockClear();
    mockExec.mockResolvedValue(endgameState);
    fireEvent.click(screen.getByAltText('♠ Q'));
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
    expect(screen.getByTestId('bezique-endgame-live')).toHaveTextContent('エンドゲーム開始');
  });

  it('does not announce the endgame on the initial render when the stock is empty', async () => {
    mockExec.mockResolvedValue(endgameState);
    renderWithProviders(<BeziquePage />);
    expect(await screen.findByTestId('bezique-endgame-live')).toBeEmptyDOMElement();
  });

  it('rings only the legal cards during an endgame follow turn', async () => {
    mockExec.mockResolvedValue(endgameFollowState);
    renderWithProviders(<BeziquePage />);
    const legal = (await screen.findByAltText('♠ Q')).closest('button');
    const illegalDiamond = screen.getByAltText('♦ J').closest('button');
    const illegalHeart = screen.getByAltText('♥ A').closest('button');
    // ♠Q is the only legal follow → it alone carries the success ring marker.
    expect(legal).toHaveAttribute('data-legal', 'true');
    expect(illegalDiamond).not.toHaveAttribute('data-legal');
    expect(illegalHeart).not.toHaveAttribute('data-legal');
  });

  it('keeps an illegal card clickable during the endgame (backend still validates)', async () => {
    mockExec.mockResolvedValue(endgameFollowState);
    renderWithProviders(<BeziquePage />);
    const illegal = (await screen.findByAltText('♦ J')).closest('button');
    // Ring-only highlighting must never block clicks: the card stays enabled.
    expect(illegal).not.toHaveAttribute('aria-disabled');
    fireEvent.click(illegal as HTMLElement);
    expect(illegal).toHaveAttribute('aria-pressed', 'true');
    mockExec.mockClear();
    mockExec.mockResolvedValue(endgameFollowState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    // ♦J is index 1 in the hand; the play still dispatches so the server can reject it.
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 1 }));
  });

  it('does not ring any card before the endgame (all cards legal)', async () => {
    renderWithProviders(<BeziquePage />);
    const card = (await screen.findByAltText('♠ Q')).closest('button');
    expect(card).not.toHaveAttribute('data-legal');
  });

  it('shows the low-stock endgame warning when the stock nears empty', async () => {
    mockExec.mockResolvedValue(makeBeziqueState({ phase: 0, currentPlayerIdx: 0, stockRemaining: 3 }));
    renderWithProviders(<BeziquePage />);
    const warning = await screen.findByTestId('bezique-stock-warning');
    expect(warning).toHaveTextContent('3');
    expect(warning).toHaveAttribute('role', 'status');
  });

  it('does not show the low-stock warning while the stock is plentiful', async () => {
    // Default fixture has stockRemaining 30 (well above the threshold).
    renderWithProviders(<BeziquePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument());
    expect(screen.queryByTestId('bezique-stock-warning')).not.toBeInTheDocument();
  });

  it('drops the low-stock warning once the endgame begins', async () => {
    mockExec.mockResolvedValue(endgameState);
    renderWithProviders(<BeziquePage />);
    await waitFor(() => expect(screen.getByText(/フェーズ2/)).toBeInTheDocument());
    expect(screen.queryByTestId('bezique-stock-warning')).not.toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<BeziquePage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝利です (1010-820)！')).toBeInTheDocument());
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndex: 0, reason: 'x' } });
    renderWithProviders(<BeziquePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  // **押したときは出る。**押していない側だけを見ていると、`isRequestedHint` を
  // 定数 false にしても通ってしまう。真の分岐も踏んでおく。
  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      hint: { cardIndex: 0, reason: 'x' },
      messageCode: 'bezique.hintRequested',
    });
    renderWithProviders(<BeziquePage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });

  it('says the target is reached rather than counting down past it', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(makeBeziqueState({ matchScore: [1000, 10] }));
    renderWithProviders(<BeziquePage />);
    const reached = await screen.findByTestId('bezique-match-progress-0');
    expect(reached.textContent).not.toMatch(/-\d/);
    // The opponent is still counting down, and only the human row is a live region.
    expect(screen.getByTestId('bezique-match-progress-1')).not.toHaveAttribute('aria-live');
  });

  it('shows how far each player is from the match target', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<BeziquePage />);
    const row = await screen.findByTestId('bezique-match-progress-0');
    expect(row.textContent).toMatch(/\d+ \/ \d+/);
  });
});
