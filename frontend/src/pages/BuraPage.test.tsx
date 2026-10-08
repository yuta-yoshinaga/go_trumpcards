import { cleanup, fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, buraApi } from '../api/gameApi';
import i18n from '../i18n';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BuraPlayer, BuraResponse, Card, CardDesign } from '../types/card';
import { BuraPage } from './BuraPage';

vi.mock('../api/gameApi', () => ({
  buraApi: { exec: vi.fn() },
  actionLogApi: { bura: vi.fn() },
}));

const mockExec = vi.mocked(buraApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function human(overrides?: Partial<BuraPlayer>): BuraPlayer {
  return {
    id: 0,
    isHuman: true,
    cardCount: 3,
    cards: [card('SPADE', 1), card('SPADE', 10), card('CLOVER', 7)],
    points: 0,
    hidden: false,
    ...overrides,
  };
}

function cpu(overrides?: Partial<BuraPlayer>): BuraPlayer {
  // A hidden seat arrives with a count and NO cards. Building the fixture this
  // way means a page that reads `cards` for the opponent renders nothing,
  // rather than quietly working off data the server never sends.
  return { id: 1, isHuman: false, cardCount: 3, cards: [], points: 0, hidden: true, ...overrides };
}

function makeState(overrides?: Partial<BuraResponse>): BuraResponse {
  return {
    players: [human(), cpu()],
    phase: 0,
    trickNumber: 0,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    currentLead: [],
    trumpSuit: 2,
    trumpCard: card('HEART', 7),
    stockRemaining: 29,
    winThreshold: 31,
    gameEndFlag: false,
    winnerIdx: -1,
    isDraw: false,
    winningCombinations: ['bura', 'moscow', 'littleMoscow', 'molodka'],
    message: '',
    ...overrides,
  };
}

describe('BuraPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('keeps the opponent hand face down during play', async () => {
    mockExec.mockResolvedValue(
      makeState({
        gameEndFlag: false,
        players: [human(), cpu({ cardCount: 2, cards: [card('HEART', 1), card('DIAMOND', 10)], hidden: false })],
      }),
    );
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const opponentHandContainer = screen.getByRole('img', { name: /CPU 1.*手札/ });
    const opponentHand = within(opponentHandContainer);
    expect(opponentHand.getAllByAltText('カード裏面')).toHaveLength(2);
    expect(screen.queryByAltText('♥ A')).not.toBeInTheDocument();
    expect(screen.queryByAltText('♦ 10')).not.toBeInTheDocument();
    expect(opponentHandContainer).not.toHaveAttribute('role', 'group');

    // The human's three cards remain buttons; opponent backs are not.
    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    expect(cardButtons).toHaveLength(3);
  });

  it('reveals the opponent hand after the game ends', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 1,
        gameEndFlag: true,
        players: [human(), cpu({ cards: [card('HEART', 1), card('DIAMOND', 10)], hidden: false })],
      }),
    );
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const opponentHand = within(screen.getByRole('group', { name: 'CPU 1の手札（公開）' }));
    expect(opponentHand.getByRole('img', { name: '♥ A' })).toBeInTheDocument();
    expect(opponentHand.getByRole('img', { name: '♦ 10' })).toBeInTheDocument();
    expect(screen.queryAllByAltText('カード裏面')).toHaveLength(0);
  });

  it('uses the lead player name for each opponent score and hand', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [human(), cpu({ id: 1, points: 12 }), cpu({ id: 2, points: 8, cardCount: 2 })],
        currentLead: [card('DIAMOND', 13)],
        leadPlayerIdx: 2,
      }),
    );
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    expect(screen.getByText('12').parentElement).toHaveTextContent('CPU 1: 12');
    expect(screen.getByText('8').parentElement).toHaveTextContent('CPU 2: 8');
    expect(screen.getByText('CPU 2が出しました')).toBeInTheDocument();
    expect(screen.getByText('CPU 1の手札 3 枚')).toBeInTheDocument();
    expect(screen.getByText('CPU 2の手札 2 枚')).toBeInTheDocument();
    expect(screen.getByRole('img', { name: 'CPU 1の手札 3 枚（裏向き）' })).toBeInTheDocument();
    expect(screen.getByRole('img', { name: 'CPU 2の手札 2 枚（裏向き）' })).toBeInTheDocument();
  });

  it('formats opponent names in English for scores, hands, and leads', async () => {
    await i18n.changeLanguage('en');
    mockExec.mockResolvedValue(
      makeState({
        players: [human(), cpu({ id: 1, points: 12 }), cpu({ id: 2, points: 8, cardCount: 2 })],
        currentLead: [card('DIAMOND', 13)],
        leadPlayerIdx: 2,
      }),
    );

    try {
      renderWithProviders(<BuraPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());

      expect(screen.getByText('12').parentElement).toHaveTextContent('CPU 1: 12');
      expect(screen.getByText('8').parentElement).toHaveTextContent('CPU 2: 8');
      expect(screen.getByText("CPU 1's hand: 3 cards")).toBeInTheDocument();
      expect(screen.getByText("CPU 2's hand: 2 cards")).toBeInTheDocument();
      expect(screen.getByText('Led by CPU 2')).toBeInTheDocument();
      expect(screen.getByRole('img', { name: "CPU 2's hand, 2 cards face down" })).toBeInTheDocument();
    } finally {
      await i18n.changeLanguage('ja');
    }
  });

  it('identifies the lead player visually and in each card label, and hides it with no lead', async () => {
    mockExec.mockResolvedValue(
      makeState({ currentLead: [card('DIAMOND', 13), card('DIAMOND', 12)], leadPlayerIdx: 1 }),
    );
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    expect(screen.getByText('CPU 1が出しました')).toBeInTheDocument();
    expect(screen.getByAltText('CPU 1が出したカード: ♦ K')).toBeInTheDocument();
    expect(screen.getByAltText('CPU 1が出したカード: ♦ Q')).toBeInTheDocument();

    cleanup();
    mockExec.mockResolvedValue(makeState({ currentLead: [], leadPlayerIdx: 0 }));
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText('あなたがリードします')).toBeInTheDocument();
    expect(screen.queryByText(/が出しました$/)).not.toBeInTheDocument();
  });

  it('plays the selected cards and clears the selection', async () => {
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[1]);
    expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'true');

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', [0, 1]));
  });

  it('asks the player to select a card when nothing is selected', async () => {
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    expect(screen.getByRole('button', { name: '出す' })).toBeDisabled();
    expect(document.getElementById('bura-play-reason')).toHaveTextContent('出すカードを選んでください');
  });

  it('refuses a mixed-suit lead', async () => {
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    // index 0 is a spade, index 2 a club -- illegal together as a lead.
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[2]);

    expect(screen.getByRole('button', { name: '出す' })).toBeDisabled();
    expect(document.getElementById('bura-play-reason')).toHaveTextContent('リードは同じスートのカードを選んでください');
    expect(screen.getByRole('button', { name: '出す' })).toHaveAttribute('aria-describedby', 'bura-play-reason');
  });

  it('requires a response to match the lead count exactly', async () => {
    mockExec.mockResolvedValue(
      makeState({ currentLead: [card('DIAMOND', 13), card('DIAMOND', 12)], leadPlayerIdx: 1, currentPlayerIdx: 0 }),
    );
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    const respond = screen.getByRole('button', { name: '2 枚で受ける' });

    fireEvent.click(cardButtons[0]);
    expect(respond).toBeDisabled();
    expect(document.getElementById('bura-play-reason')).toHaveTextContent('リードに合わせて2枚選んでください');

    // Two cards of DIFFERENT suits are a legal response -- only a lead has to
    // be one suit. This is the case a "same suit" check would wrongly block.
    fireEvent.click(cardButtons[2]);
    expect(respond).toBeEnabled();
    expect(document.getElementById('bura-play-reason')).toBeEmptyDOMElement();
    expect(respond).not.toHaveAttribute('aria-describedby');
  });

  it('explains when a response selects more cards than the lead', async () => {
    mockExec.mockResolvedValue(
      makeState({ currentLead: [card('DIAMOND', 13)], leadPlayerIdx: 1, currentPlayerIdx: 0 }),
    );
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[1]);

    expect(screen.getByRole('button', { name: '1 枚で受ける' })).toBeDisabled();
    expect(document.getElementById('bura-play-reason')).toHaveTextContent('リードに合わせて1枚だけ選んでください');
  });

  it('explains a lead selection that exceeds the three-card limit', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [human({ cards: [card('SPADE', 1), card('SPADE', 10), card('SPADE', 11), card('SPADE', 12)] }), cpu()],
      }),
    );
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    for (const button of cardButtons) fireEvent.click(button);
    expect(screen.getByRole('button', { name: '出す' })).toBeDisabled();
    expect(document.getElementById('bura-play-reason')).toHaveTextContent('リードは同じスートから3枚まで選べます');
  });

  it('claims and declares through their own commands', async () => {
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '31点を宣言' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('claim'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '役を宣言' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare'));
  });

  it('shows the claim risk until the human reaches the threshold', async () => {
    renderWithProviders(<BuraPage />);
    expect(await screen.findByText('到達していなければその場で負けます')).toBeInTheDocument();

    cleanup();
    mockExec.mockResolvedValue(makeState({ players: [human({ points: 31 }), cpu()] }));
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '31点を宣言' })).toBeInTheDocument());
    expect(screen.queryByText('到達していなければその場で負けます')).not.toBeInTheDocument();
  });

  it('reports a draw distinctly from a loss', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: 1, gameEndFlag: true, isDraw: true, winnerIdx: -1, messageCode: 'bura.draw' }),
    );
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByText('流局 — 誰も宣言しませんでした')).toBeInTheDocument();
  });

  it('shows the trump suit once the indicator has been drawn', async () => {
    mockExec.mockResolvedValue(makeState({ trumpCard: undefined, trumpSuit: 2 }));
    renderWithProviders(<BuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByText('切札 ハート（指示カードは引かれた）')).toBeInTheDocument();
  });
});

describe('BuraPage action log during play', () => {
  it('opens the action log with its current entries before the game ends', async () => {
    mockExec.mockResolvedValue(makeState({ gameEndFlag: false }));
    vi.mocked(actionLogApi.bura).mockResolvedValueOnce({
      entries: [
        {
          turnNumber: 1,
          playerIdx: -1,
          actionType: 'deal',
          detail: '切り札のスートは 2 です',
          detailCode: 'bura.log.deal',
          detailParams: { suit: '2' },
          cards: [],
        },
      ],
    });
    renderWithProviders(<BuraPage />);
    await screen.findByRole('button', { name: '棋譜を見る' });
    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    expect(await screen.findByText(/切り札のスートは 2 です/)).toBeInTheDocument();
  });
});

// #5568: claim には失敗時のリスクが title で出ているのに、declare は何が「役」か
// どこにも書かれておらず、押しても何も起きない理由が分からなかった。
describe('BuraPage declare tooltip', () => {
  it('lists every combination the server recognises', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<BuraPage />);
    const btn = await screen.findByRole('button', { name: '役を宣言' });
    const title = btn.getAttribute('title') ?? '';
    expect(title).toContain('ブラ');
    expect(title).toContain('モスクワ');
    expect(title).toContain('小モスクワ');
    expect(title).toContain('モロトカ');
  });

  // **役の一覧はサーバから。**画面側で数え直していないことを、別の一覧を
  // 返して確かめる。
  it('renders only what the server sends', async () => {
    mockExec.mockResolvedValue({ ...makeState(), winningCombinations: ['bura'] });
    renderWithProviders(<BuraPage />);
    const btn = await screen.findByRole('button', { name: '役を宣言' });
    const title = btn.getAttribute('title') ?? '';
    expect(title).toContain('ブラ');
    expect(title).not.toContain('モロトカ');
  });

  // 訳の無いキーはキー名を出すより落とす。役の説明として読めないので。
  it('drops a combination it has no wording for', async () => {
    mockExec.mockResolvedValue({ ...makeState(), winningCombinations: ['bura', 'notARealCombo'] });
    renderWithProviders(<BuraPage />);
    const btn = await screen.findByRole('button', { name: '役を宣言' });
    expect(btn.getAttribute('title') ?? '').not.toContain('notARealCombo');
  });

  // claim 側は変えていない (受け入れ条件3)。
  it('leaves the claim warning alone', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<BuraPage />);
    const claim = await screen.findByRole('button', { name: /点を宣言/ });
    expect(claim).toHaveAttribute('title', '到達していなければその場で負けます');
  });
});
