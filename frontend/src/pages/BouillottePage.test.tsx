import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { bouillotteApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeBouillotteState } from '../test/stateFactories';
import { BouillottePage } from './BouillottePage';

vi.mock('../api/gameApi', () => ({
  bouillotteApi: { exec: vi.fn() },
  actionLogApi: { bouillotte: vi.fn() },
}));

const mockExec = vi.mocked(bouillotteApi.exec);

const bettingState = makeBouillotteState({ phase: 0, isHumanTurn: true, canRaise: true });
const cpuTurnState = makeBouillotteState({ phase: 0, isHumanTurn: false, canRaise: false });
const resultState = makeBouillotteState({
  phase: 1,
  isHumanTurn: false,
  winnerIdx: 0,
  result: 1,
  players: [
    {
      id: 0,
      isHuman: true,
      chips: 230,
      roundBet: 40,
      folded: false,
      out: false,
      cardCount: 3,
      cards: [
        { design: 'SPADE', value: 13 },
        { design: 'SPADE', value: 12 },
        { design: 'HEART', value: 5 },
      ],
      handName: 'highcard',
      isWinner: true,
    },
    {
      id: 1,
      isHuman: false,
      chips: 170,
      roundBet: 40,
      folded: false,
      out: false,
      cardCount: 3,
      cards: [
        { design: 'HEART', value: 4 },
        { design: 'CLOVER', value: 3 },
        { design: 'DIAMOND', value: 2 },
      ],
      handName: 'highcard',
      isWinner: false,
    },
    {
      id: 2,
      isHuman: false,
      chips: 190,
      roundBet: 10,
      folded: true,
      out: false,
      cardCount: 3,
      cards: [],
      isWinner: false,
    },
    {
      id: 3,
      isHuman: false,
      chips: 190,
      roundBet: 10,
      folded: true,
      out: false,
      cardCount: 3,
      cards: [],
      isWinner: false,
    },
  ],
});
const gameEndState = makeBouillotteState({
  phase: 1,
  isHumanTurn: false,
  gameEndFlag: true,
  matchWinnerIdx: 0,
  winnerIdx: 0,
  result: 1,
  message: 'ゲーム終了！ あなたの勝利です！',
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(bettingState);
});

describe('BouillottePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<BouillottePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<BouillottePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        playerCount: 4,
        ante: 10,
        startingChips: 200,
        targetRounds: 10,
      }),
    );
  });

  it('shows the betting action buttons on the human betting turn', async () => {
    renderWithProviders(<BouillottePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /レイズ／ヴィ/ })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド（降りる）' })).toBeInTheDocument();
  });

  it('dispatches bet call when the Call button is clicked', async () => {
    renderWithProviders(<BouillottePage />);
    const btn = await screen.findByRole('button', { name: 'コール' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 'call'));
  });

  it('dispatches bet raise when the Raise button is clicked', async () => {
    renderWithProviders(<BouillottePage />);
    const btn = await screen.findByRole('button', { name: /レイズ／ヴィ/ });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 'raise'));
  });

  it('dispatches bet fold when the Fold button is clicked', async () => {
    renderWithProviders(<BouillottePage />);
    const btn = await screen.findByRole('button', { name: 'フォールド（降りる）' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 'fold'));
  });

  it('hides action buttons when it is not the human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<BouillottePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /リセット|ゲームをリセット/ })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'コール' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'フォールド（降りる）' })).not.toBeInTheDocument();
  });

  it('hides the raise button when raising is not allowed', async () => {
    mockExec.mockResolvedValue(makeBouillotteState({ phase: 0, isHumanTurn: true, canRaise: false }));
    renderWithProviders(<BouillottePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: /レイズ／ヴィ/ })).not.toBeInTheDocument();
  });

  it('shows the next-round button at the result phase and dispatches nextround', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<BouillottePage />);
    const btn = await screen.findByRole('button', { name: '次のラウンド' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('hides betting buttons on the result phase', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<BouillottePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'コール' })).not.toBeInTheDocument();
  });

  it('renders the game-end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<BouillottePage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝利です！')).toBeInTheDocument());
  });

  // **サーバー由来の retourneMatch でリングが付く (#6494)。**
  // クライアント側の analyzeRetourneMatch を消したので、state.retourneMatch を
  // 直接読むことでリングが付くことを確認する。
  describe('retourneMatch ring and note (server-driven)', () => {
    it('highlights hand cards listed in state.retourneMatch.matchingIndices', async () => {
      // デフォルト: retourneMatch.matchingIndices = [0] → 手札 0 のみリング
      renderWithProviders(<BouillottePage />);
      await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
      expect(screen.getByTestId('hand-card-0').className).toContain('ring-ds-accent');
      expect(screen.getByTestId('hand-card-1').className).not.toContain('ring-ds-accent');
      expect(screen.getByTestId('hand-card-2').className).not.toContain('ring-ds-accent');
      // ルトゥルヌ自体も一致あるときリング。
      expect(screen.getByTestId('retourne-card').className).toContain('ring-ds-accent');
      // 1 枚だけ一致: noteKey = "" → ノートなし。
      expect(screen.queryByTestId('retourne-note')).not.toBeInTheDocument();
    });

    it('shows the brelan favori note when server sends noteKey "favori"', async () => {
      // サーバーが matchingIndices=[0,1], noteKey="favori" を返す場合。
      const favoriState = makeBouillotteState({
        phase: 0,
        isHumanTurn: true,
        retourne: { design: 'DIAMOND', value: 12 },
        players: [
          {
            id: 0,
            isHuman: true,
            chips: 190,
            roundBet: 10,
            folded: false,
            out: false,
            cardCount: 3,
            cards: [
              { design: 'SPADE', value: 12 },
              { design: 'HEART', value: 12 },
              { design: 'CLOVER', value: 8 },
            ],
            handName: 'brelan',
            isWinner: false,
          },
          ...bettingState.players.slice(1),
        ],
        retourneMatch: { matchingIndices: [0, 1], noteKey: 'favori' },
      });
      mockExec.mockResolvedValue(favoriState);
      renderWithProviders(<BouillottePage />);
      await waitFor(() => expect(screen.getByTestId('retourne-note')).toBeInTheDocument());
      expect(screen.getByTestId('retourne-note').textContent).toContain('ブルラン・ファヴォリ');
      expect(screen.getByTestId('hand-card-0').className).toContain('ring-ds-accent');
      expect(screen.getByTestId('hand-card-1').className).toContain('ring-ds-accent');
      expect(screen.getByTestId('hand-card-2').className).not.toContain('ring-ds-accent');
    });

    it('shows the carre note when server sends noteKey "carre"', async () => {
      const carreState = makeBouillotteState({
        retourneMatch: { matchingIndices: [0, 1, 2], noteKey: 'carre' },
      });
      mockExec.mockResolvedValue(carreState);
      renderWithProviders(<BouillottePage />);
      await waitFor(() => expect(screen.getByTestId('retourne-note')).toBeInTheDocument());
      expect(screen.getByTestId('retourne-note').textContent).toContain('カレ');
      expect(screen.getByTestId('hand-card-0').className).toContain('ring-ds-accent');
      expect(screen.getByTestId('hand-card-1').className).toContain('ring-ds-accent');
      expect(screen.getByTestId('hand-card-2').className).toContain('ring-ds-accent');
    });

    it('does not highlight any card when server sends empty matchingIndices', async () => {
      const noMatchState = makeBouillotteState({
        retourneMatch: { matchingIndices: [], noteKey: '' },
      });
      mockExec.mockResolvedValue(noMatchState);
      renderWithProviders(<BouillottePage />);
      await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
      expect(screen.getByTestId('hand-card-0').className).not.toContain('ring-ds-accent');
      expect(screen.getByTestId('retourne-card').className).not.toContain('ring-ds-accent');
      expect(screen.queryByTestId('retourne-note')).not.toBeInTheDocument();
    });

    it('shows no note when server sends noteKey ""', async () => {
      // noteKey が空文字の場合ノートは出ない。
      const oneMatchState = makeBouillotteState({
        retourneMatch: { matchingIndices: [0], noteKey: '' },
      });
      mockExec.mockResolvedValue(oneMatchState);
      renderWithProviders(<BouillottePage />);
      await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
      expect(screen.queryByTestId('retourne-note')).not.toBeInTheDocument();
    });
  });

  it('shows the pot odds and call amount when chips are owed to call', async () => {
    // pot 40, currentBet 20, human roundBet 10 -> call 10 into pot 40:
    // odds 10 / (40 + 10) = 20%, ratio 40:10 -> 4:1.
    const oddsState = makeBouillotteState({
      phase: 0,
      isHumanTurn: true,
      canRaise: true,
      pot: 40,
      currentBet: 20,
    });
    mockExec.mockResolvedValue(oddsState);
    renderWithProviders(<BouillottePage />);
    const odds = await screen.findByTestId('bouillotte-pot-odds');
    expect(odds.textContent).toContain('20%');
    expect(odds.textContent).toContain('4:1');
    // The Call button spells out the required amount.
    expect(screen.getByRole('button', { name: 'コール（10）' })).toBeInTheDocument();
  });

  it('shows the free-check state when nothing is owed to call', async () => {
    // Default betting state: currentBet 10 equals the human's roundBet 10 -> call 0.
    renderWithProviders(<BouillottePage />);
    const odds = await screen.findByTestId('bouillotte-pot-odds');
    expect(odds.textContent).toContain('コール不要');
    // With nothing owed the Call button shows the plain label (no amount).
    expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument();
  });

  // **レイズが消えた理由を書く。**回数上限とチップ不足を区別できないと、
  // 突然選択肢を奪われたように見える (#4924)。
  describe('raise cap readout', () => {
    it('shows the current count against the cap while raising is still open', async () => {
      mockExec.mockResolvedValue(
        makeBouillotteState({ phase: 0, isHumanTurn: true, canRaise: true, raiseCount: 1, maxRaises: 3 }),
      );
      renderWithProviders(<BouillottePage />);
      const readout = await screen.findByTestId('bouillotte-raise-count');
      expect(readout).toHaveTextContent('レイズ 1/3回');
      expect(readout).not.toHaveTextContent('上限');
    });

    // **枚数不足も理由として区別する。**上限に達していないのに押せないとき、
    // 「レイズ 1/3回」とだけ出すと「まだできる」と読めてしまう（レビュー指摘）。
    it('says the chips are short when the cap is not the reason', async () => {
      mockExec.mockResolvedValue(
        makeBouillotteState({
          phase: 0,
          isHumanTurn: true,
          canRaise: false,
          raiseCount: 1,
          maxRaises: 3,
          currentBet: 10,
          ante: 5,
          players: [
            { ...makeBouillotteState().players[0], isHuman: true, chips: 3, roundBet: 0 },
            ...makeBouillotteState().players.slice(1),
          ],
        }),
      );
      renderWithProviders(<BouillottePage />);
      const readout = await screen.findByTestId('bouillotte-raise-count');
      expect(readout).toHaveTextContent('チップが足りません');
      expect(readout).not.toHaveTextContent('レイズ 1/3回');
    });

    it('says the cap was reached rather than silently dropping the button', async () => {
      mockExec.mockResolvedValue(
        makeBouillotteState({ phase: 0, isHumanTurn: true, canRaise: false, raiseCount: 3, maxRaises: 3 }),
      );
      renderWithProviders(<BouillottePage />);
      const readout = await screen.findByTestId('bouillotte-raise-count');
      expect(readout).toHaveTextContent('レイズ上限（3回）に達しました');
      // ボタン自体は消えたままでよい。理由が読めることが要点。
      expect(screen.queryByRole('button', { name: /レイズ/ })).not.toBeInTheDocument();
    });
  });
});
