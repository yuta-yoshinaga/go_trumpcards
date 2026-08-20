import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { goFishApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, GoFishResponse } from '../types/card';
import { GoFishPhase } from '../types/phases';
import { GoFishPage } from './GoFishPage';

vi.mock('../api/gameApi', () => ({
  goFishApi: { exec: vi.fn() },
  actionLogApi: { gofish: vi.fn() },
}));

const mockExec = vi.mocked(goFishApi.exec);

const humanCards: Card[] = [
  { design: 'SPADE', value: 7 },
  { design: 'HEART', value: 7 },
  { design: 'DIAMOND', value: 3 },
];

const baseState: GoFishResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 3,
      cards: humanCards,
      bookCount: 0,
      books: [],
    },
    { id: 1, isHuman: false, cardCount: 5, cards: [], bookCount: 1, books: [] },
    { id: 2, isHuman: false, cardCount: 4, cards: [], bookCount: 0, books: [] },
    { id: 3, isHuman: false, cardCount: 3, cards: [], bookCount: 2, books: [] },
  ],
  phase: GoFishPhase.PLAY,
  currentTurn: 0,
  gameEndFlag: false,
  winnerIdx: -1,
  turnNumber: 1,
  deckRemaining: 30,
  lastAsk: null,
  cpuActions: [],
  humanAction: null,
  message: '',
  config: { cpuDifficulty: 1 },
};

const cpuTurnState: GoFishResponse = {
  ...baseState,
  currentTurn: 1,
};

const gameEndState: GoFishResponse = {
  ...baseState,
  phase: GoFishPhase.GAME_END,
  gameEndFlag: true,
  winnerIdx: 0,
};

const gameEndByFlagState: GoFishResponse = {
  ...baseState,
  phase: GoFishPhase.PLAY,
  gameEndFlag: true,
  winnerIdx: 2,
};

beforeEach(() => {
  mockExec.mockResolvedValue(baseState);
});

afterEach(() => {
  localStorage.clear();
});

describe('GoFishPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<GoFishPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with default config', async () => {
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 1 }));
  });

  it('renders player cards in footer', async () => {
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getAllByRole('button', { name: /♠ 7|♥ 7|♦ 3/ })).toHaveLength(3));
  });

  it('renders CPU player areas', async () => {
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByText(/CPU 2/)).toBeInTheDocument());
  });

  it('renders deck remaining', async () => {
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByText(/30/)).toBeInTheDocument());
  });

  it('shows ask button on human turn', async () => {
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '要求する' })).toBeInTheDocument());
  });

  it('ask button is disabled when target or rank not selected', async () => {
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '要求する' })).toBeDisabled());
  });

  it('ask button is enabled when both target and rank selected', async () => {
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByText(/CPU 2/)).toBeInTheDocument());

    // Select opponent
    const cpuButtons = screen.getAllByRole('button', { name: /CPU/ });
    fireEvent.click(cpuButtons[0]);

    // Select rank by clicking a card
    const cardButton = screen.getByRole('button', { name: /♠ 7/ });
    fireEvent.click(cardButton);

    expect(screen.getByRole('button', { name: '要求する' })).not.toBeDisabled();
  });

  it('calls ask command when ask button clicked', async () => {
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByText(/CPU 2/)).toBeInTheDocument());

    const cpuButtons = screen.getAllByRole('button', { name: /CPU/ });
    fireEvent.click(cpuButtons[0]);
    fireEvent.click(screen.getByRole('button', { name: /♠ 7/ }));

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseState);
    fireEvent.click(screen.getByRole('button', { name: '要求する' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('ask', expect.any(Number), 7));
  });

  // **どのキーが誰なのかを一覧から読み取れること。**相手が何人いても全行が
  // 「対象のプレイヤーを選ぶ」で、盤面と数え合わせるしかなかった (#4862)。
  it('names each opponent in the shortcut list', async () => {
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByText(/CPU 2/)).toBeInTheDocument());
    const panel = screen.getByTestId('go-fish-kbd-shortcuts');
    fireEvent.click(panel.querySelector('summary') as HTMLElement);

    expect(within(panel).getByText('CPU 1 を選ぶ')).toBeInTheDocument();
    expect(within(panel).getByText('CPU 2 を選ぶ')).toBeInTheDocument();
    expect(within(panel).queryByText('対象のプレイヤーを選ぶ')).not.toBeInTheDocument();
  });

  it('keyboard: number key selects an opponent and arrows cycle the rank, then "a" asks', async () => {
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByText(/CPU 2/)).toBeInTheDocument());
    // "1" picks the first opponent (CPU 1) and announces it.
    fireEvent.keyDown(document.body, { key: '1' });
    await waitFor(() => expect(screen.getByTestId('gf-kbd-announce').textContent).toMatch(/CPU 1/));
    // ArrowRight cycles to the first hand rank (3) and announces it.
    fireEvent.keyDown(document.body, { key: 'ArrowRight' });
    await waitFor(() => expect(screen.getByTestId('gf-kbd-announce').textContent).toMatch(/3/));
    // A second ArrowRight advances from the current rank to the next one (7).
    fireEvent.keyDown(document.body, { key: 'ArrowRight' });
    await waitFor(() => expect(screen.getByTestId('gf-kbd-announce').textContent).toMatch(/7/));
    mockExec.mockClear();
    mockExec.mockResolvedValue(baseState);
    // "a" asks the selected opponent for the selected rank.
    fireEvent.keyDown(document.body, { key: 'a' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('ask', 1, 7));
  });

  it('keyboard: "a" also asks once a target and rank are chosen', async () => {
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByText(/CPU 2/)).toBeInTheDocument());
    fireEvent.keyDown(document.body, { key: '2' }); // CPU 2
    fireEvent.keyDown(document.body, { key: 'ArrowLeft' }); // wraps to the last rank (7)
    mockExec.mockClear();
    mockExec.mockResolvedValue(baseState);
    fireEvent.keyDown(document.body, { key: 'a' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('ask', 2, 7));
  });

  it('keyboard: bindings are inert on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByText(/CPU 2/)).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: '1' });
    fireEvent.keyDown(document.body, { key: 'ArrowRight' });
    fireEvent.keyDown(document.body, { key: 'a' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('ask', expect.anything(), expect.anything());
    expect(screen.getByTestId('gf-kbd-announce').textContent).toBe('');
  });

  it('rank buttons appear on human turn', async () => {
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByText('7')).toBeInTheDocument());
  });

  it('hides ask button on CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByText(/CPU 2/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '要求する' })).not.toBeInTheDocument();
  });

  it('opponent toggle deselects on second click', async () => {
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByText(/CPU 2/)).toBeInTheDocument());

    const cpuButtons = screen.getAllByRole('button', { name: /CPU/ });
    fireEvent.click(cpuButtons[0]);
    expect(cpuButtons[0]).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(cpuButtons[0]);
    expect(cpuButtons[0]).toHaveAttribute('aria-pressed', 'false');
  });

  it('rank button toggle deselects on second click', async () => {
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByText('7')).toBeInTheDocument());

    const rankButton = screen.getByRole('button', { name: '7' });
    fireEvent.click(rankButton);
    expect(rankButton).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(rankButton);
    expect(rankButton).toHaveAttribute('aria-pressed', 'false');
  });

  it('shows error alert on API failure', async () => {
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('reset button triggers confirmation dialog', async () => {
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('button', { name: '確認' })).toBeInTheDocument();
  });

  it('confirms reset and calls reset command', async () => {
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 1 }));
  });

  it('renders game end state', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: '要求する' })).not.toBeInTheDocument());
  });

  it('renders game end state via gameEndFlag', async () => {
    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: '要求する' })).not.toBeInTheDocument());
  });

  it('settings change updates difficulty for reset', async () => {
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByText(/CPU 2/)).toBeInTheDocument());

    fireEvent.click(screen.getByText('ゴーフィッシュ 設定'));
    const select = screen.getByRole('combobox');
    fireEvent.change(select, { target: { value: '2' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 2 }));
  });

  it('loading state disables reset button', async () => {
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    let resolve!: (value: GoFishResponse) => void;
    const slow = new Promise<GoFishResponse>((r) => {
      resolve = r;
    });
    mockExec.mockReturnValueOnce(slow);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();
    resolve(baseState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
  });

  it('renders hint toggle checkbox', async () => {
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByText(/CPU 2/)).toBeInTheDocument());
    expect(screen.getByRole('checkbox')).toBeInTheDocument();
  });

  it('renders the hint toggle as a SettingsPanel item next to the difficulty select', async () => {
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByText(/CPU 2/)).toBeInTheDocument());
    // The toggle is a SettingsPanel item with a stable id (the old inline checkbox had none).
    const checkbox = screen.getByRole('checkbox');
    expect(checkbox).toHaveAttribute('id', 'frontendHint');
    expect(screen.getByLabelText('ヒント表示')).toBe(checkbox);
    expect(screen.getByRole('combobox')).toHaveAttribute('id', 'cpuDifficulty');
  });

  it('shows HintTooltip when hint is enabled and human has a valid turn', async () => {
    localStorage.setItem('hint_enabled_gofish', 'true');
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByRole('checkbox')).toBeChecked());
    // HintTooltip renders when hint is active (human turn, PLAY phase, cards in hand)
    await waitFor(() => expect(screen.getByRole('checkbox')).toBeInTheDocument());
    // Verify HintTooltip is rendered (it contains the hint reason text)
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  // #5518: 文言は「最も多く持っているランク」と言うのに、どの札のことか
  // 盤面には出ていなかった。
  it('marks the cards of the rank the hint names, and only those', async () => {
    localStorage.setItem('hint_enabled_gofish', 'true');
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());

    // 手札は ♠7 ♥7 ♦3 -- 7 が2枚なので 7 の2枚だけに印が付く。
    expect(screen.getByRole('button', { name: '♠ 7' })).toHaveAttribute('data-hint-card', 'true');
    expect(screen.getByRole('button', { name: '♥ 7' })).toHaveAttribute('data-hint-card', 'true');
    expect(screen.getByRole('button', { name: '♦ 3' })).not.toHaveAttribute('data-hint-card');
    // **リングはインラインで置く。**手札には selectedCardStyle の
    // `boxShadow: 'none'` が乗るので、Tailwind の ring-* (同じ box-shadow) は
    // 未選択のあいだ潰される。印だけ見ていても気づけない。
    expect(screen.getByRole('button', { name: '♠ 7' }).style.outline).toContain('var(--color-ds-warning)');
    expect(screen.getByRole('button', { name: '♦ 3' }).style.outline).toBe('');
    // ツールチップもランクを名指しする。
    expect(screen.getByTestId('hint-tooltip')).toHaveTextContent('7');
  });

  it('marks nothing while the hint is switched off', async () => {
    localStorage.setItem('hint_enabled_gofish', 'false');
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ 7' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '♠ 7' })).not.toHaveAttribute('data-hint-card');
  });

  it('shows books display when human has books', async () => {
    const stateWithBooks: GoFishResponse = {
      ...baseState,
      players: [
        {
          ...baseState.players[0],
          bookCount: 1,
          books: [{ rank: 7, cards: [] }],
        },
        ...baseState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(stateWithBooks);
    renderWithProviders(<GoFishPage />);
    await waitFor(() => expect(screen.getAllByText(/ブック.*1/).length).toBeGreaterThanOrEqual(1));
  });
});
