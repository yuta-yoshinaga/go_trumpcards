import { fireEvent, screen, waitFor } from '@testing-library/react';
import i18n from 'i18next';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, piquetApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { PiquetResponse } from '../types/card';
import { PiquetDeclarationKind, PiquetExchangeTurn, PiquetPhase } from '../types/phases';
import { PiquetPage } from './PiquetPage';

vi.mock('../api/gameApi', () => ({
  piquetApi: { exec: vi.fn() },
  actionLogApi: { piquet: vi.fn() },
}));

const mockExec = vi.mocked(piquetApi.exec);
const mockActionLog = vi.mocked(actionLogApi.piquet);

function makeState(overrides: Partial<PiquetResponse> = {}): PiquetResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 12,
        cards: [
          { design: 'SPADE', value: 13 },
          { design: 'HEART', value: 1 },
        ],
        trickCount: 0,
        declScore: 0,
        trickScore: 0,
        bonusScore: 0,
        roundScore: 0,
        matchScore: 0,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 12,
        cards: [],
        trickCount: 0,
        declScore: 0,
        trickScore: 0,
        bonusScore: 0,
        roundScore: 0,
        matchScore: 0,
      },
    ],
    phase: PiquetPhase.EXCHANGE,
    dealNumber: 1,
    dealsPerPartie: 6,
    elderIdx: 0,
    youngerIdx: 1,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    trickNumber: 0,
    tricksWon: [0, 0],
    exchangeTurn: PiquetExchangeTurn.ELDER,
    elderExchangedCnt: 0,
    youngerExchangedCnt: 0,
    elderTalon: [],
    youngerTalon: [],
    elderRevealedTalon: [],
    youngerRevealedTalon: [],
    carteBlanche: [false, false],
    declStage: PiquetDeclarationKind.POINT,
    declResults: [],
    currentTrick: [],
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: { dealsPerPartie: 6 },
    ...overrides,
  };
}

describe('PiquetPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<PiquetPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders deal info', async () => {
    mockExec.mockResolvedValue(makeState({ dealNumber: 2 }));
    renderWithProviders(<PiquetPage />);
    // Look for "2" + "/" + "6" pattern from the deal header
    await waitFor(() => expect(screen.getByText(/2.*\/.*6|ディール/i)).toBeInTheDocument());
  });

  it('renders elder + younger labels', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<PiquetPage />);
    await waitFor(() => expect(screen.getAllByText(/Elder|Younger/).length).toBeGreaterThan(0));
  });

  it('shows the human hand cards', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<PiquetPage />);
    // Cards now render as real images with localized alt text (suit + rank).
    await waitFor(() => expect(screen.getByAltText('♠ K')).toBeInTheDocument());
    expect(screen.getByAltText('♥ A')).toBeInTheDocument();
  });

  it('renders reset button', async () => {
    mockExec.mockResolvedValue(makeState({ phase: PiquetPhase.DECLARATION }));
    renderWithProviders(<PiquetPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /リセット|Reset/i })).toBeInTheDocument());
  });

  it('highlights human meld badge when a new declaration result arrives', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: PiquetPhase.DECLARATION,
        declStage: PiquetDeclarationKind.SEQUENCE,
        declResults: [
          {
            kind: PiquetDeclarationKind.POINT,
            elderClaim: {
              length: 2,
              topRank: 13,
              pipTotal: 21,
              suit: 0,
              cards: [
                { design: 'SPADE', value: 13 },
                { design: 'HEART', value: 1 },
              ],
            },
            youngerClaim: { length: 0, topRank: 0, pipTotal: 0, suit: 0, cards: [] },
            winner: 0,
            scoredBy: 0,
            score: 2,
          },
        ],
      }),
    );
    renderWithProviders(<PiquetPage />);
    const wonBadge = await screen.findByTestId('piquet-meld-badge');
    // Human (elder, idx 0) was scoredBy:0 above → won → success palette.
    expect(wonBadge).toHaveClass('text-ds-success', 'border-ds-success');
  });

  it('renders the translated per-player stats line', async () => {
    const state = makeState();
    Object.assign(state.players[0]!, { declScore: 4, trickScore: 8, bonusScore: 10, roundScore: 22 });
    Object.assign(state.players[1]!, { declScore: 3, trickScore: 6, bonusScore: 2, roundScore: 11 });
    mockExec.mockResolvedValue(state);
    renderWithProviders(<PiquetPage />);
    await waitFor(() => expect(screen.getAllByText(/手札: 12/).length).toBeGreaterThan(0));
    expect(screen.getByText(/宣言点: 4/)).toBeInTheDocument();
    expect(screen.getByText(/トリック点: 8/)).toBeInTheDocument();
    expect(screen.getByText(/ボーナス点: 10/)).toBeInTheDocument();
    expect(screen.getByText(/ラウンド: 22/)).toBeInTheDocument();
    expect(screen.getByText(/宣言点: 3/)).toBeInTheDocument();
    expect(screen.getByText(/トリック点: 6/)).toBeInTheDocument();
    expect(screen.getByText(/ボーナス点: 2/)).toBeInTheDocument();
    expect(screen.getByText(/ラウンド: 11/)).toBeInTheDocument();
  });

  it.each([
    [PiquetExchangeTurn.ELDER, 0],
    [PiquetExchangeTurn.YOUNGER, 1],
  ])('shows hand counts and the selected exchange preview for turn %s', async (exchangeTurn, humanIdx) => {
    const state = makeState({ exchangeTurn });
    if (humanIdx === 1) {
      state.players[0]!.isHuman = false;
      state.players[1]!.isHuman = true;
      state.players[1]!.cards = state.players[0]!.cards;
    }
    mockExec.mockResolvedValue(state);
    renderWithProviders(<PiquetPage />);
    await waitFor(() => {
      const region = screen.getByTestId('piquet-exchange-counts');
      expect(region).toHaveAttribute('aria-live', 'polite');
      expect(region).toHaveTextContent('現在の手札: 12枚 / 選択中: 0枚 / 交換後: 12枚');
    });
    fireEvent.click(await screen.findByAltText('♠ K'));
    expect(screen.getByTestId('piquet-exchange-counts')).toHaveTextContent(
      '現在の手札: 12枚 / 選択中: 1枚 / 交換後: 12枚',
    );
  });

  it('renders the declaration list with translated tied and fallback-scorer labels', async () => {
    const claim = { length: 0, topRank: 0, pipTotal: 0, suit: 0, cards: [] };
    mockExec.mockResolvedValue(
      makeState({
        phase: PiquetPhase.DECLARATION,
        declStage: PiquetDeclarationKind.SET,
        declResults: [
          {
            kind: PiquetDeclarationKind.POINT,
            elderClaim: claim,
            youngerClaim: claim,
            winner: -1,
            scoredBy: -1,
            score: 0,
          },
          {
            kind: PiquetDeclarationKind.SEQUENCE,
            elderClaim: claim,
            youngerClaim: claim,
            winner: -1,
            scoredBy: 5, // neither elder (0) nor younger (1) → "?" fallback label
            score: 3,
          },
        ],
      }),
    );
    renderWithProviders(<PiquetPage />);
    await waitFor(() => expect(screen.getByText(/引き分け/)).toBeInTheDocument()); // declTied
    expect(screen.getByText(/\? \+3/)).toBeInTheDocument(); // declScored with "?" scorer
  });

  it('exposes the declaration list as an additions-only live log for screen readers', async () => {
    const claim = { length: 0, topRank: 0, pipTotal: 0, suit: 0, cards: [] };
    mockExec.mockResolvedValue(
      makeState({
        phase: PiquetPhase.DECLARATION,
        declStage: PiquetDeclarationKind.SEQUENCE,
        declResults: [
          {
            kind: PiquetDeclarationKind.POINT,
            elderClaim: claim,
            youngerClaim: claim,
            winner: 0,
            scoredBy: 0,
            score: 4,
          },
        ],
      }),
    );
    renderWithProviders(<PiquetPage />);
    const log = await screen.findByTestId('piquet-declaration-list');
    // role="log" (default aria-atomic="false") announces only newly-appended
    // results, not the whole list on every update.
    expect(log).toHaveAttribute('role', 'log');
    expect(log).toHaveAttribute('aria-live', 'polite');
  });

  it('renders the translated trick header during the play phase', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: PiquetPhase.PLAY,
        currentTrick: [{ playerIdx: 0, card: { design: 'SPADE', value: 13 } }],
      }),
    );
    renderWithProviders(<PiquetPage />);
    // Exact match isolates the TrickView header from the "トリック: 0" stats line.
    await waitFor(() => expect(screen.getByText('トリック')).toBeInTheDocument()); // trickHeader
    expect((await screen.findByTestId('piquet-trick-card-0')).textContent).toContain('エルダー');
  });

  it('maps trick players to roles when elder is player 1, in English', async () => {
    await i18n.changeLanguage('en');
    mockExec.mockResolvedValue(
      makeState({
        phase: PiquetPhase.PLAY,
        elderIdx: 1,
        youngerIdx: 0,
        currentTrick: [{ playerIdx: 0, card: { design: 'SPADE', value: 13 } }],
      }),
    );
    renderWithProviders(<PiquetPage />);
    expect((await screen.findByTestId('piquet-trick-card-0')).textContent).toContain('Younger');
    await i18n.changeLanguage('ja');
  });

  it('maps trick players to roles when elder is player 0, in Japanese', async () => {
    await i18n.changeLanguage('ja');
    mockExec.mockResolvedValue(
      makeState({
        phase: PiquetPhase.PLAY,
        currentTrick: [
          { playerIdx: 0, card: { design: 'SPADE', value: 13 } },
          { playerIdx: 1, card: { design: 'HEART', value: 1 } },
        ],
      }),
    );
    renderWithProviders(<PiquetPage />);
    expect((await screen.findByTestId('piquet-trick-card-0')).textContent).toContain('エルダー');
    expect(screen.getByTestId('piquet-trick-card-1').textContent).toContain('ヤンガー');
  });

  it('shows the meld badge in the lost palette when the opponent scores', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: PiquetPhase.DECLARATION,
        declStage: PiquetDeclarationKind.SEQUENCE,
        declResults: [
          {
            kind: PiquetDeclarationKind.POINT,
            elderClaim: {
              length: 2,
              topRank: 13,
              pipTotal: 21,
              suit: 0,
              cards: [
                { design: 'SPADE', value: 13 },
                { design: 'HEART', value: 1 },
              ],
            },
            youngerClaim: { length: 0, topRank: 0, pipTotal: 0, suit: 0, cards: [] },
            winner: 1,
            scoredBy: 1,
            score: 2,
          },
        ],
      }),
    );
    renderWithProviders(<PiquetPage />);
    const lostBadge = await screen.findByTestId('piquet-meld-badge');
    // scoredBy:1 (younger) → human lost → error palette (border signal, readable text).
    expect(lostBadge).toHaveClass('text-ds-text-primary', 'border-ds-error');
  });

  it('shows the hint button while the human can act and dispatches the hint command', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<PiquetPage />);
    const hintBtn = await screen.findByRole('button', { name: 'ヒント' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(makeState());
    fireEvent.click(hintBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('h'));
  });

  it('renders the play hint text when a card suggestion is present', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: PiquetPhase.PLAY,
        hint: { cardIndex: 3, reason: 'lowest' },
        messageCode: 'piquet.hintAvailable',
      }),
    );
    renderWithProviders(<PiquetPage />);
    const hint = await screen.findByTestId('piquet-hint');
    expect(hint).toHaveTextContent('[3]');
  });

  it('renders the discard hint text when discard indices are suggested', async () => {
    mockExec.mockResolvedValue(
      makeState({ hint: { discardIndices: [1, 2], reason: 'lowest' }, messageCode: 'piquet.hintAvailable' }),
    );
    renderWithProviders(<PiquetPage />);
    const hint = await screen.findByTestId('piquet-hint');
    expect(hint).toHaveTextContent('1, 2');
  });

  // #5603: 出せないカードも押せて、サーバーに move を投げてエラーで返ってくる形
  // だった。マストフォローの相手は presenter が毎回計算して返しているので、
  // それをそのまま使う。
  it('disables the cards that legalPlayIndices leaves out', async () => {
    mockExec.mockResolvedValue(makeState({ phase: PiquetPhase.PLAY, legalPlayIndices: [1] }));
    renderWithProviders(<PiquetPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const cards = await screen.findAllByRole('button', { name: /♠|♥/ });
    const playable = cards.filter((c) => c.getAttribute('data-hint-action') === 'play');
    expect(playable).toHaveLength(2);
    // 共有の PlayerHandSection と同じ扱い: aria-disabled で**フォーカスは残す**。
    // HTML の disabled にすると、読み上げ利用者から札そのものが消える。
    expect(playable[0]).toHaveAttribute('aria-disabled', 'true');
    expect(playable[0]).toHaveAccessibleName(/♠.*出せない札/);
    expect(playable[0]).not.toBeDisabled();
    expect(playable[1]).not.toHaveAttribute('aria-disabled');
    expect(playable[1]).toHaveAccessibleName(/♥.*出せる札/);

    mockExec.mockClear();
    fireEvent.click(playable[0] as HTMLElement);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  // 交換フェーズには legalPlayIndices が載らない。**載っていないことを
  // 「どれも出せない」と読むと、手札が全部死ぬ。**
  it('leaves every card usable when no legal set is present', async () => {
    mockExec.mockResolvedValue(makeState({ phase: PiquetPhase.PLAY }));
    renderWithProviders(<PiquetPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const cards = await screen.findAllByRole('button', { name: /♠|♥/ });
    for (const c of cards.filter((b) => b.getAttribute('data-hint-action') === 'play')) {
      expect(c).not.toBeDisabled();
      expect(c).not.toHaveAttribute('aria-disabled');
    }
  });

  it('does not show the action log view button before the partie ends', async () => {
    mockExec.mockResolvedValue(makeState({ phase: PiquetPhase.PLAY }));
    renderWithProviders(<PiquetPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '棋譜を見る' })).not.toBeInTheDocument();
  });

  it('shows the action log view button at game end and renders fetched entries on click', async () => {
    mockExec.mockResolvedValue(makeState({ phase: PiquetPhase.GAME_END, gameEndFlag: true, winnerIdx: 0 }));
    mockActionLog.mockResolvedValue({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'play', detail: 'plays a card' }],
    });
    renderWithProviders(<PiquetPage />);
    const viewBtn = await screen.findByRole('button', { name: '棋譜を見る' });
    fireEvent.click(viewBtn);
    await waitFor(() => expect(mockActionLog).toHaveBeenCalled());
    expect(await screen.findByText(/plays a card/)).toBeInTheDocument();
  });

  // **ヒント経路はページ側からも踏む。**ファクトリ単体テストだけだと
  // `hintFactories` の登録行と、ページのトグル／ツールチップが一度も実行
  // されない（codecov が #4596 でその 3 ファイルを未到達として報告した）。
  it('turns the frontend hint on from the checkbox', async () => {
    localStorage.removeItem('hint_enabled_piquet');
    mockExec.mockResolvedValue(makeState({ phase: PiquetPhase.PLAY, hint: { cardIndex: 0, reason: 'lowest' } }));
    renderWithProviders(<PiquetPage />);

    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(toggle);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
  });

  it('shows the hint tooltip once the toggle is on', async () => {
    localStorage.setItem('hint_enabled_piquet', 'true');
    mockExec.mockResolvedValue(makeState({ phase: PiquetPhase.PLAY, hint: { cardIndex: 0, reason: 'lowest' } }));
    renderWithProviders(<PiquetPage />);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
  });

  it('keeps the tooltip hidden while the toggle is off', async () => {
    // 直前のテストが立てた localStorage を引き継がない。
    localStorage.removeItem('hint_enabled_piquet');
    mockExec.mockResolvedValue(makeState({ phase: PiquetPhase.PLAY, hint: { cardIndex: 0, reason: 'lowest' } }));
    renderWithProviders(<PiquetPage />);
    await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();
  });

  it('takes the declaration-kind label from the i18n bundle', async () => {
    // A sentinel proves the label is looked up: the old hardcoded 'Point' ignores it.
    const original = i18n.getResourceBundle('ja', 'piquet') as Record<string, unknown>;
    i18n.addResourceBundle('ja', 'piquet', { declKindPoint: '__POINT__' }, true, true);
    try {
      mockExec.mockResolvedValue(makeState({ phase: PiquetPhase.DECLARATION, declStage: PiquetDeclarationKind.POINT }));
      renderWithProviders(<PiquetPage />);
      await waitFor(() => expect(screen.getByText('次の宣言: __POINT__')).toBeInTheDocument());
    } finally {
      i18n.removeResourceBundle('ja', 'piquet');
      i18n.addResourceBundle('ja', 'piquet', original, true, true);
    }
  });

  it('falls back to ? for an unknown declaration kind', async () => {
    mockExec.mockResolvedValue(makeState({ phase: PiquetPhase.DECLARATION, declStage: 99 }));
    renderWithProviders(<PiquetPage />);
    await waitFor(() => expect(screen.getByText('次の宣言: ?')).toBeInTheDocument());
  });
});
