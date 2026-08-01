import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { literatureApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CardDesign, LiteraturePlayer, LiteratureResponse } from '../types/card';
import { LiteraturePhase } from '../types/phases';
import { LiteraturePage } from './LiteraturePage';

vi.mock('../api/gameApi', () => ({
  literatureApi: { exec: vi.fn() },
  actionLogApi: { literature: vi.fn() },
}));

const mockExec = vi.mocked(literatureApi.exec);

const card = (design: CardDesign, value: number) => ({ design, value });

/** The six cards of half-suit `half`, in the order the server sends them. */
function halfSuitCards(half: number) {
  const suits: CardDesign[] = ['SPADE', 'CLOVER', 'HEART', 'DIAMOND'];
  const design = suits[Math.floor(half / 2)] ?? 'SPADE';
  const ranks = half % 2 === 0 ? [2, 3, 4, 5, 6, 7] : [9, 10, 11, 12, 13, 1];
  return ranks.map((v) => card(design, v));
}

function seat(id: number, isHuman: boolean, overrides?: Partial<LiteraturePlayer>): LiteraturePlayer {
  return {
    id,
    isHuman,
    team: id % 2,
    cardCount: 8,
    cards: isHuman ? [card('SPADE', 2), card('HEART', 9)] : [],
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function makeState(overrides?: Partial<LiteratureResponse>): LiteratureResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false), seat(4, false), seat(5, false)],
    phase: LiteraturePhase.PLAY,
    currentPlayerIdx: 0,
    halfSuits: [0, 0, 0, 0, 0, 0, 0, 0],
    halfSuitCards: Array.from({ length: 8 }, (_, h) => halfSuitCards(h)),
    asks: [],
    claims: [],
    lastAsk: null,
    lastClaim: null,
    teamHalfSuits: [0, 0],
    cancelledCount: 0,
    openCount: 8,
    winThreshold: 5,
    halfSuitCnt: 8,
    gameEndFlag: false,
    winnerTeam: -1,
    message: '',
    config: { cpuDifficulty: 0 },
    ...overrides,
  };
}

describe('LiteraturePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<LiteraturePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **勝利には5組。**4組では決着しないことを画面に書く。
  it('states that winning takes five, not four', async () => {
    renderWithProviders(<LiteraturePage />);
    await waitFor(() => expect(screen.getByTestId('literature-threshold-note')).toBeInTheDocument());
    const note = screen.getByTestId('literature-threshold-note');
    expect(note).toHaveTextContent('8組の過半数=5組');
    expect(note).toHaveTextContent('4組では相手も4組になり得るので決着しません');
    // **無効はどちらのものにもならない。**
    expect(note).toHaveTextContent('無効になった組はどちらのものにもなりません');
  });

  // **味方の手札も見えない。**推理が成立する前提。
  it('says that even a teammate hand is hidden', async () => {
    renderWithProviders(<LiteraturePage />);
    await waitFor(() => expect(screen.getByTestId('literature-hidden-note')).toBeInTheDocument());
    expect(screen.getByTestId('literature-hidden-note')).toHaveTextContent('味方の手札も見えません');
  });

  // **席は交互。**味方に要求できない規則の前提。
  it('shows six alternating seats', async () => {
    renderWithProviders(<LiteraturePage />);
    await waitFor(() => expect(screen.getAllByTestId('literature-player')).toHaveLength(6));
    const players = screen.getAllByTestId('literature-player');
    expect(players[0]).toHaveTextContent('チーム0');
    expect(players[1]).toHaveTextContent('チーム1');
    expect(players[2]).toHaveTextContent('チーム0');
  });

  // **要求できるのは相手チームのみ。**味方は選択肢に出さない。
  it('offers only opponents as ask targets', async () => {
    renderWithProviders(<LiteraturePage />);
    await waitFor(() => expect(screen.getByLabelText(/相手/)).toBeInTheDocument());
    const select = screen.getByLabelText(/相手/) as HTMLSelectElement;
    expect(Array.from(select.options).map((o) => o.value)).toEqual(['1', '3', '5']);
    expect(screen.getByTestId('literature-ask-rules')).toHaveTextContent('相手チームにのみ');
  });

  // **手札の尽きた相手には訊けない。**選択肢から外す。
  it('drops an opponent with no cards left', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          seat(0, true),
          seat(1, false, { cardCount: 0 }),
          seat(2, false),
          seat(3, false),
          seat(4, false),
          seat(5, false),
        ],
      }),
    );
    renderWithProviders(<LiteraturePage />);
    await waitFor(() => expect(screen.getByLabelText(/相手/)).toBeInTheDocument());
    expect(Array.from((screen.getByLabelText(/相手/) as HTMLSelectElement).options).map((o) => o.value)).toEqual([
      '3',
      '5',
    ]);
  });

  it('asks a specific card of a specific opponent', async () => {
    renderWithProviders(<LiteraturePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '要求する' })).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText(/相手/), { target: { value: '3' } });
    // 先頭は ♠ 低位の 2。
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '要求する' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('ask', { target: 3, suit: 1, value: 2 }));
  });

  // **決着した組の札は要求できない。**
  it('drops settled half-suits from the askable cards', async () => {
    mockExec.mockResolvedValue(makeState({ halfSuits: [1, 0, 0, 0, 0, 0, 0, 0], openCount: 7 }));
    renderWithProviders(<LiteraturePage />);
    await waitFor(() => expect(screen.getByLabelText(/^札/)).toBeInTheDocument());
    const options = Array.from((screen.getByLabelText(/^札/) as HTMLSelectElement).options);
    // ♠ 低位 (2-7) は消え、残り 7 組 × 6 枚 = 42 枚。
    expect(options).toHaveLength(42);
    expect(options.map((o) => o.textContent)).not.toContain('♠2');
  });

  // **宣言は6枚すべての所在を申告する。**候補は自チームの席だけ。
  it('claims a half-suit by placing all six cards with own-team seats', async () => {
    renderWithProviders(<LiteraturePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '宣言する' })).toBeInTheDocument());

    const holderSelects = screen.getAllByLabelText(/^♠[0-9]+$/);
    expect(holderSelects).toHaveLength(6);
    // 自チームの席だけ。
    expect(Array.from((holderSelects[0] as HTMLSelectElement).options).map((o) => o.value)).toEqual(['0', '2', '4']);

    fireEvent.change(holderSelects[0], { target: { value: '2' } });
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '宣言する' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('claim', { halfSuit: 0, holders: [2, 0, 0, 0, 0, 0] }));
  });

  // **無効は「相手に渡る」とは違う。**宣言の説明に書く。
  it('explains that misplacing within your own team cancels the claim', async () => {
    renderWithProviders(<LiteraturePage />);
    await waitFor(() => expect(screen.getByTestId('literature-claim-rules')).toBeInTheDocument());
    const rules = screen.getByTestId('literature-claim-rules');
    expect(rules).toHaveTextContent('自チーム内で言い間違えると無効');
    expect(rules).toHaveTextContent('相手が1枚でも持っていれば相手の獲得');
  });

  // **無効を帰属の一覧でも区別する。**
  it('shows half-suit ownership with cancelled distinguished', async () => {
    mockExec.mockResolvedValue(
      makeState({ halfSuits: [1, 2, 3, 0, 0, 0, 0, 0], teamHalfSuits: [1, 1], cancelledCount: 1, openCount: 5 }),
    );
    renderWithProviders(<LiteraturePage />);
    await waitFor(() => expect(screen.getByTestId('literature-halfsuits')).toBeInTheDocument());
    const panel = screen.getByTestId('literature-halfsuits');
    expect(panel).toHaveTextContent('[0] ♠ 低位 (2-7): チーム0');
    expect(panel).toHaveTextContent('[1] ♠ 高位 (9-A): チーム1');
    expect(panel).toHaveTextContent('[2] ♣ 低位 (2-7): 無効');
    expect(panel).toHaveTextContent('[3] ♣ 高位 (9-A): 未決');
  });

  // **要求の履歴は公開情報。**的中と空振りを区別して出す。
  it('shows the public ask history', async () => {
    mockExec.mockResolvedValue(
      makeState({
        asks: [
          { from: 0, to: 1, card: card('SPADE', 3), success: true },
          { from: 1, to: 0, card: card('HEART', 9), success: false },
        ],
      }),
    );
    renderWithProviders(<LiteraturePage />);
    await waitFor(() => expect(screen.getByTestId('literature-history')).toBeInTheDocument());
    const history = screen.getByTestId('literature-history');
    expect(history).toHaveTextContent('全員に公開');
    expect(history).toHaveTextContent('的中');
    expect(history).toHaveTextContent('空振り');
  });

  // **同数で終わることがある。**無効が絡むため。
  it('reports the outcome, a level finish included', async () => {
    for (const [team, text] of [
      [0, /あなたのチームの勝利です！/],
      [1, /相手チームの勝利です。/],
      [-1, /同数で勝者なしです。/],
    ] as const) {
      mockExec.mockResolvedValue(makeState({ phase: LiteraturePhase.GAME_END, gameEndFlag: true, winnerTeam: team }));
      renderWithProviders(<LiteraturePage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });

  // 相手の手番なら操作は出さない。
  it('hides the controls when it is not your turn', async () => {
    mockExec.mockResolvedValue(
      makeState({
        currentPlayerIdx: 1,
        players: [
          seat(0, true, { isCurrentTurn: false }),
          seat(1, false, { isCurrentTurn: true }),
          seat(2, false),
          seat(3, false),
          seat(4, false),
          seat(5, false),
        ],
      }),
    );
    renderWithProviders(<LiteraturePage />);
    await waitFor(() => expect(screen.getAllByTestId('literature-player')).toHaveLength(6));
    expect(screen.queryByRole('button', { name: '要求する' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '宣言する' })).not.toBeInTheDocument();
  });
});
