import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { gutsApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeGutsState } from '../test/stateFactories';
import * as gutsGuideUtils from '../utils/gutsGuideUtils';
import { GutsPage } from './GutsPage';

vi.mock('../api/gameApi', () => ({
  gutsApi: { exec: vi.fn() },
  actionLogApi: { guts: vi.fn() },
}));

const mockExec = vi.mocked(gutsApi.exec);

const declareState = makeGutsState({ phase: 0 });
const resultState = makeGutsState({
  phase: 1,
  winnerIdx: 0,
  result: 1,
  matchers: [1],
  players: [
    {
      id: 0,
      isHuman: true,
      chips: 230,
      in: true,
      out: false,
      roundBet: 10,
      cardCount: 2,
      cards: [
        { design: 'SPADE', value: 13 },
        { design: 'SPADE', value: 12 },
      ],
      handName: 'highcard',
      isWinner: true,
      isMatcher: false,
    },
    {
      id: 1,
      isHuman: false,
      chips: 170,
      in: true,
      out: false,
      roundBet: 90,
      cardCount: 2,
      cards: [
        { design: 'HEART', value: 5 },
        { design: 'CLOVER', value: 3 },
      ],
      handName: 'highcard',
      isWinner: false,
      isMatcher: true,
    },
    {
      id: 2,
      isHuman: false,
      chips: 200,
      in: false,
      out: false,
      roundBet: 0,
      cardCount: 2,
      cards: [],
      isWinner: false,
      isMatcher: false,
    },
    {
      id: 3,
      isHuman: false,
      chips: 200,
      in: false,
      out: false,
      roundBet: 0,
      cardCount: 2,
      cards: [],
      isWinner: false,
      isMatcher: false,
    },
  ],
});
const gameEndState = makeGutsState({
  phase: 1,
  gameEndFlag: true,
  matchWinnerIdx: 0,
  winnerIdx: 0,
  result: 1,
  message: 'ゲーム終了！ あなたの勝利です！',
});

beforeEach(() => {
  localStorage.clear();
  mockExec.mockReset();
  mockExec.mockResolvedValue(declareState);
});

describe('GutsPage', () => {
  it('shows the human turn instruction once', async () => {
    mockExec.mockResolvedValue(
      makeGutsState({ message: 'あなたの番です — イン（残る）かアウト（降りる）を選んでください。' }),
    );
    renderWithProviders(<GutsPage />);
    expect(
      await screen.findAllByText('あなたの番です — イン（残る）かアウト（降りる）を選んでください。'),
    ).toHaveLength(1);
  });

  it('shows the configured target rounds', async () => {
    for (const { targetRounds, text } of [
      { targetRounds: 4, text: 'ラウンド 1 / 4' },
      { targetRounds: 12, text: 'ラウンド 1 / 12' },
    ]) {
      mockExec.mockResolvedValueOnce(makeGutsState({ config: { ...makeGutsState().config, targetRounds } }));
      const { unmount } = renderWithProviders(<GutsPage />);
      await waitFor(() => expect(screen.getByText(text)).toBeInTheDocument());
      unmount();
    }
  });

  it('exposes each player as a list item with a combined accessible summary', async () => {
    const state = makeGutsState({
      phase: 1,
      players: [
        { ...resultState.players[0], chips: 230, roundBet: 10, handName: 'highcard' },
        ...resultState.players.slice(1),
      ],
    });
    mockExec.mockResolvedValue(state);
    renderWithProviders(<GutsPage />);

    const list = await screen.findByTestId('guts-player-list');
    const player = list.querySelector('li');
    expect(player).toHaveAccessibleName('あなた。チップ: 230。賭け: 10。勝者。ハイカード');
  });

  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<GutsPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<GutsPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        playerCount: 4,
        ante: 10,
        startingChips: 200,
        targetRounds: 10,
      }),
    );
  });

  it('keeps the hand in the footer and renders declaration guidance and reset in the play area', async () => {
    const { container } = renderWithProviders(<GutsPage />);
    const footerContent = await screen.findByTestId('game-footer-content');
    const playArea = container.querySelector('.flex-1.overflow-y-auto');

    expect(playArea).not.toBeNull();
    expect(footerContent.querySelector('[data-tutorial="guts-hand"]')).not.toBeNull();
    expect(within(footerContent).queryByTestId('guts-declare-guide')).not.toBeInTheDocument();
    expect(footerContent.querySelector('[data-tutorial="guts-reset-button"]')).toBeNull();
    expect(within(playArea as HTMLElement).getByTestId('guts-declare-guide')).toBeInTheDocument();
    expect(playArea?.querySelector('[data-tutorial="guts-reset-button"]')).not.toBeNull();
  });

  it('shows the declare action buttons on the declare phase', async () => {
    renderWithProviders(<GutsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'イン（残る）' })).toBeInTheDocument());
    const actions = within(screen.getByTestId('game-footer-actions'));
    expect(actions.getByRole('button', { name: 'イン（残る）' })).toBeInTheDocument();
    expect(actions.getByRole('button', { name: 'アウト（降りる）' })).toBeInTheDocument();
  });

  it('shows the human chips, round bet, and current pot beside the declare actions', async () => {
    const state = makeGutsState({
      phase: 0,
      pot: 75,
      players: [
        { ...declareState.players[0], isHuman: true, chips: 185, roundBet: 15 },
        ...declareState.players.slice(1),
      ],
    });
    mockExec.mockResolvedValue(state);
    renderWithProviders(<GutsPage />);

    const money = await screen.findByTestId('guts-declare-money');
    expect(money).toHaveTextContent('残高: 185');
    expect(money).toHaveTextContent('賭け額: 15');
    expect(money).toHaveTextContent('ポット: 75');
    expect(
      money.compareDocumentPosition(screen.getByRole('button', { name: 'イン（残る）' })) &
        Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
  });

  it('does not show declaration money after the phase changes to result', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<GutsPage />);
    await screen.findByRole('button', { name: '次のラウンド' });
    expect(screen.queryByTestId('guts-declare-money')).not.toBeInTheDocument();
  });

  it('dispatches declare in when the In button is clicked', async () => {
    renderWithProviders(<GutsPage />);
    const btn = await screen.findByRole('button', { name: 'イン（残る）' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare', 1));
  });

  it('dispatches declare out when the Out button is clicked', async () => {
    renderWithProviders(<GutsPage />);
    const btn = await screen.findByRole('button', { name: 'アウト（降りる）' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare', 0));
  });

  it('shows the next-round button at the result phase and dispatches nextround', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<GutsPage />);
    const btn = await screen.findByRole('button', { name: '次のラウンド' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('hides declare buttons on the result phase', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<GutsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'イン（残る）' })).not.toBeInTheDocument();
  });

  it('shows the matcher extra payment in the round result', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<GutsPage />);
    const line = await screen.findByTestId('guts-matcher-payment');
    // roundBet 90 minus ante 10 = 80 chips matched into the next pot.
    expect(line).toHaveTextContent('マッチ支払い: CPU 1 が -80');
  });

  it('shows a strong (high) win-chance guideline and the pair name for a paired hand', async () => {
    const pairState = makeGutsState({
      phase: 0,
      players: [
        {
          id: 0,
          isHuman: true,
          chips: 200,
          in: false,
          out: false,
          roundBet: 10,
          cardCount: 2,
          cards: [
            { design: 'SPADE', value: 9 },
            { design: 'HEART', value: 9 },
          ],
          isWinner: false,
          isMatcher: false,
        },
        ...declareState.players.slice(1),
      ],
    });
    mockExec.mockResolvedValue(pairState);
    renderWithProviders(<GutsPage />);
    const guide = await screen.findByTestId('guts-declare-guide');
    expect(guide).toHaveTextContent('手役: ペア');
    expect(screen.getByTestId('guts-guide-tier')).toHaveTextContent('高い');
    expect(screen.getByTestId('guts-guide-probability')).toHaveTextContent(/推定勝率: \d+\.\d+%/);
    expect(screen.getByText(/残っているCPUはランダムな手札/)).toBeInTheDocument();
    // The maximum match-loss payment is the smaller of the pot and chip balance.
    expect(screen.getByTestId('guts-guide-risk')).toHaveTextContent('敗北時の最大支払額: 40');
  });

  it('caps the declaration guide payment at the human chip balance', async () => {
    const shortStackState = makeGutsState({
      phase: 0,
      pot: 75,
      players: [{ ...declareState.players[0], isHuman: true, chips: 25 }, ...declareState.players.slice(1)],
    });
    mockExec.mockResolvedValue(shortStackState);
    renderWithProviders(<GutsPage />);

    expect(await screen.findByTestId('guts-guide-risk')).toHaveTextContent('敗北時の最大支払額: 25');
  });

  it('does not reevaluate the declaration guide on unrelated rerenders', async () => {
    const evaluateGuide = vi.spyOn(gutsGuideUtils, 'evaluateGutsGuide');
    renderWithProviders(<GutsPage />);
    await screen.findByTestId('guts-declare-guide');
    expect(evaluateGuide).toHaveBeenCalledTimes(1);

    fireEvent.click(screen.getByText('CLI'));
    fireEvent.click(screen.getByText('GUI'));

    expect(evaluateGuide).toHaveBeenCalledTimes(1);
    evaluateGuide.mockRestore();
  });

  it('shows a weak (low) win-chance guideline and the high-card name for a low hand', async () => {
    const weakState = makeGutsState({
      phase: 0,
      players: [
        {
          id: 0,
          isHuman: true,
          chips: 200,
          in: false,
          out: false,
          roundBet: 10,
          cardCount: 2,
          cards: [
            { design: 'SPADE', value: 2 },
            { design: 'HEART', value: 7 },
          ],
          isWinner: false,
          isMatcher: false,
        },
        ...declareState.players.slice(1),
      ],
    });
    mockExec.mockResolvedValue(weakState);
    renderWithProviders(<GutsPage />);
    const guide = await screen.findByTestId('guts-declare-guide');
    expect(guide).toHaveTextContent('手役: ハイカード');
    expect(screen.getByTestId('guts-guide-tier')).toHaveTextContent('低い');
    expect(screen.getByTestId('guts-guide-probability')).toHaveTextContent(/推定勝率: \d+\.\d+%/);
  });

  it('hides the declaration guideline on the result phase', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<GutsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.queryByTestId('guts-declare-guide')).not.toBeInTheDocument();
  });

  it('renders the game-end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<GutsPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝利です！')).toBeInTheDocument());
  });

  // **誰も残らなかったラウンドは表示が丸ごと無かった。**CUI は最初から
  // guts.result.carry で持ち越し額と連続回数を出している (#4847)。
  it('shows the carried pot when nobody stayed in', async () => {
    mockExec.mockResolvedValue(makeGutsState({ phase: 1, winnerIdx: -1, carryPot: 40, carryCount: 2 }));
    renderWithProviders(<GutsPage />);

    const panel = await screen.findByTestId('guts-carry-result');
    expect(panel).toHaveTextContent('ポット 40');
    expect(panel).toHaveTextContent('2 回連続');
  });

  it('shows the winner panel instead when the round had a winner', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<GutsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.queryByTestId('guts-carry-result')).not.toBeInTheDocument();
  });

  it('exposes the declare guide as a live region for screen readers', async () => {
    const pairState = makeGutsState({
      phase: 0,
      players: [
        {
          id: 0,
          isHuman: true,
          chips: 200,
          in: false,
          out: false,
          roundBet: 10,
          cardCount: 2,
          cards: [
            { design: 'SPADE', value: 9 },
            { design: 'HEART', value: 9 },
          ],
          isWinner: false,
          isMatcher: false,
        },
        ...declareState.players.slice(1),
      ],
    });
    mockExec.mockResolvedValue(pairState);
    const { container } = renderWithProviders(<GutsPage />);

    await waitFor(() => {
      // 属性の有無ではなく「読み上げられる中身」を見る ── ガイドの手役・tier・
      // マッチリスクが live region の**中に**あって初めて変化が伝わる。
      const liveRegion = container.querySelector('[data-testid="guts-declare-guide"][aria-live="polite"]');
      expect(liveRegion).not.toBeNull();
      expect(liveRegion).toHaveAttribute('role', 'status');
      expect(liveRegion).toHaveTextContent('手役: ペア');
      expect(liveRegion).toHaveTextContent('高い');
      expect(liveRegion).toHaveTextContent('敗北時の最大支払額: 40');
    });
  });

  it('shows the translated phase name, not the raw phase key', async () => {
    renderWithProviders(<GutsPage />);
    expect(await screen.findByText('コール', { selector: 'strong' })).toBeInTheDocument();
    expect(screen.queryByText('declare', { selector: 'strong' })).not.toBeInTheDocument();
  });

  it('renders out, in, and waiting player badges with and without hand names', async () => {
    const state = makeGutsState({
      phase: 0,
      players: [
        { ...declareState.players[0], in: true, handName: 'pair' },
        { ...declareState.players[1], id: 1, isHuman: false, in: true, out: true, handName: undefined },
        { ...declareState.players[2], id: 2, isHuman: false, in: true, out: false, handName: 'highcard' },
        { ...declareState.players[3], id: 3, isHuman: false, in: false, out: false, handName: undefined },
      ],
    });
    mockExec.mockResolvedValue(state);
    renderWithProviders(<GutsPage />);

    const players = await screen.findByTestId('guts-player-list');
    expect(players).toHaveTextContent('あなた — チップ: 200 · 賭け: 10 · [イン] · ペア');
    expect(players).toHaveTextContent('CPU 1 — チップ: 200 · 賭け: 10 · [脱落]');
    expect(players).toHaveTextContent('CPU 2 — チップ: 200 · 賭け: 10 · [イン] · ハイカード');
    expect(players).toHaveTextContent('CPU 3 — チップ: 200 · 賭け: 10 · [待機]');
  });

  it('shows only the hand label when the human has no cards', async () => {
    mockExec.mockResolvedValue(
      makeGutsState({
        players: [{ ...declareState.players[0], cards: [], handName: undefined }, ...declareState.players.slice(1)],
      }),
    );
    renderWithProviders(<GutsPage />);
    const footer = await screen.findByTestId('game-footer-content');
    expect(within(footer).getByText('あなたの手札')).toBeInTheDocument();
    expect(within(footer).queryByText(/あなたの手札 —/)).not.toBeInTheDocument();
  });

  it('skips a matcher payment when its player index is missing', async () => {
    mockExec.mockResolvedValue(makeGutsState({ phase: 1, winnerIdx: 0, matchers: [99] }));
    renderWithProviders(<GutsPage />);
    await screen.findByRole('button', { name: '次のラウンド' });
    expect(screen.queryByTestId('guts-matcher-payment')).not.toBeInTheDocument();
  });

  it('suppresses round actions when the game has ended', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<GutsPage />);
    await screen.findByText('ゲーム終了！ あなたの勝利です！');
    expect(screen.queryByRole('button', { name: '次のラウンド' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'イン（残る）' })).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument();
  });

  it.each([
    { name: 'human wins the match', matchWinnerIdx: 0, isHuman: true, result: 0, celebrates: true },
    { name: 'human wins the round', matchWinnerIdx: -1, isHuman: false, result: 1, celebrates: true },
    { name: 'human does not win', matchWinnerIdx: 1, isHuman: false, result: 0, celebrates: false },
  ])('sets the win state when $name', async ({ matchWinnerIdx, isHuman, result, celebrates }) => {
    const state = makeGutsState({
      phase: 1,
      gameEndFlag: true,
      matchWinnerIdx,
      result,
      players: [
        { ...declareState.players[0], isHuman: true },
        { ...declareState.players[1], isHuman, id: 1 },
        ...declareState.players.slice(2),
      ],
    });
    mockExec.mockResolvedValue(state);
    renderWithProviders(<GutsPage />);
    await screen.findByRole('button', { name: '次のゲーム' });
    if (celebrates) expect(await screen.findByTestId('win-celebration')).toBeInTheDocument();
    else expect(screen.queryByTestId('win-celebration')).not.toBeInTheDocument();
  });

  it.each([
    ['playerCount', '6', { playerCount: 6 }],
    ['ante', '20', { ante: 20 }],
    ['startingChips', '500', { startingChips: 500 }],
    ['targetRounds', '20', { targetRounds: 20 }],
  ])('applies the selected %s when the game resets', async (id, value, expected) => {
    renderWithProviders(<GutsPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, expect.any(Object)));
    mockExec.mockClear();
    fireEvent.change(document.getElementById(id as string) as HTMLSelectElement, { target: { value } });
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, expect.objectContaining(expected)));
  });

  it('executes reset after confirming the reset prompt', async () => {
    renderWithProviders(<GutsPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, expect.any(Object)));
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        playerCount: 4,
        ante: 10,
        startingChips: 200,
        targetRounds: 10,
      }),
    );
  });
});
