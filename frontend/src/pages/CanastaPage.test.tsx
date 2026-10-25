import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { canastaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CanastaPlayerData, CanastaResponse } from '../types/card';
import { CanastaPage } from './CanastaPage';

vi.mock('../api/gameApi', () => ({
  canastaApi: { exec: vi.fn() },
  actionLogApi: { canasta: vi.fn() },
}));
const mockExec = vi.mocked(canastaApi.exec);
const breakpointState = vi.hoisted(() => ({ enabled: false }));

vi.mock('../hooks/useCardDimensions', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../hooks/useCardDimensions')>();
  return {
    ...actual,
    useCardDimensions: () => {
      const dimensions = actual.useCardDimensions();
      return breakpointState.enabled ? { ...dimensions, isMobile: false } : dimensions;
    },
    useBreakpoint: () => {
      const breakpoint = actual.useBreakpoint();
      return breakpointState.enabled ? 'desktop' : breakpoint;
    },
  };
});

const basePlayers: CanastaPlayerData[] = [
  {
    id: 0,
    isHuman: true,
    cardCount: 15,
    cards: [
      { design: 'SPADE', value: 7 },
      { design: 'CLOVER', value: 7 },
      { design: 'HEART', value: 7 },
      { design: 'SPADE', value: 10 },
      { design: 'CLOVER', value: 10 },
    ],
    melds: [],
    red3Count: 0,
    red3s: [],
    roundScore: 0,
    cumulativeScore: 0,
    scoreBreakdown: { meldCards: 0, canastaBonus: 0, red3Bonus: 0, goOutBonus: 0, handPenalty: 0 },
    hasCanasta: false,
    hasInitMeld: false,
  },
  {
    id: 1,
    isHuman: false,
    cardCount: 15,
    cards: [],
    melds: [],
    red3Count: 0,
    red3s: [],
    roundScore: 0,
    cumulativeScore: 0,
    scoreBreakdown: { meldCards: 0, canastaBonus: 0, red3Bonus: 0, goOutBonus: 0, handPenalty: 0 },
    hasCanasta: false,
    hasInitMeld: false,
  },
];

const drawPhaseState: CanastaResponse = {
  minMeld: 50,
  players: basePlayers,
  phase: 0,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardTop: { design: 'SPADE', value: 5 },
  discardPile: [{ design: 'SPADE', value: 5 }],
  drawPileCount: 67,
  discardPileCount: 1,
  isFrozen: false,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
  messageCode: 'canasta.drawPhase',
  config: { cpuDifficulty: 1, pointLimit: 5000 },
};

const meldPhaseState: CanastaResponse = {
  ...drawPhaseState,
  phase: 1,
  messageCode: 'canasta.meldPhase',
};

const discardPhaseState: CanastaResponse = {
  ...drawPhaseState,
  phase: 2,
  messageCode: 'canasta.discardPhase',
};

const roundEndState: CanastaResponse = {
  ...drawPhaseState,
  phase: 3,
  messageCode: 'canasta.roundEnd',
};

const gameEndState: CanastaResponse = {
  ...drawPhaseState,
  phase: 4,
  gameEndFlag: true,
  winnerIdx: 0,
};

// Human holds one completed natural canasta (7 cards) and one incomplete meld (3 cards)
// so the compact stack + count badge rendering can be exercised.
const meldDisplayState: CanastaResponse = {
  ...meldPhaseState,
  players: [
    {
      ...basePlayers[0],
      hasCanasta: true,
      melds: [
        {
          rank: 7,
          isNatural: true,
          isCanasta: true,
          cards: Array.from({ length: 7 }, (_, i) => ({ design: i % 2 ? 'HEART' : 'SPADE', value: 7 })),
        },
        {
          rank: 10,
          isNatural: false,
          isCanasta: false,
          cards: [
            { design: 'SPADE', value: 10 },
            { design: 'CLOVER', value: 10 },
            { design: 'HEART', value: 10 },
          ],
        },
      ],
    },
    basePlayers[1],
  ],
};

afterEach(() => {
  Object.defineProperty(window, 'innerWidth', { configurable: true, writable: true, value: 1024 });
});

describe('CanastaPage', () => {
  it('renders turn guidance in the footer on desktop', async () => {
    Object.defineProperty(window, 'innerWidth', { configurable: true, writable: true, value: 1024 });
    mockExec.mockResolvedValue({
      ...drawPhaseState,
      message: 'Footer guidance',
      messageCode: undefined,
      messageParams: undefined,
    });
    renderWithProviders(<CanastaPage />);
    const message = await screen.findByText('Footer guidance');
    expect(message.closest('[data-testid="game-footer-message"]')).toBeInTheDocument();
  });

  it('keeps turn guidance in the play area on mobile', async () => {
    Object.defineProperty(window, 'innerWidth', { configurable: true, writable: true, value: 375 });
    mockExec.mockResolvedValue({
      ...drawPhaseState,
      message: 'Mobile guidance',
      messageCode: undefined,
      messageParams: undefined,
    });
    renderWithProviders(<CanastaPage />);
    const message = await screen.findByText('Mobile guidance');
    expect(message.closest('[role="status"]')).toBeInTheDocument();
    expect(screen.queryByTestId('game-footer-message')).not.toBeInTheDocument();
  });

  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(drawPhaseState);
  });

  it('keeps the player meld area beside the hand on large desktop', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1280 });
    window.dispatchEvent(new Event('resize'));
    try {
      mockExec.mockResolvedValue({
        ...drawPhaseState,
        players: [
          {
            ...basePlayers[0],
            cards: Array.from({ length: 15 }, (_, i) => ({ design: 'SPADE' as const, value: (i % 13) + 1 })),
          },
          basePlayers[1],
        ],
      });
      renderWithProviders(<CanastaPage />);
      await screen.findByTestId('phase-indicator');
      const footer = document.querySelector('[data-testid="game-footer-content"]');
      expect(footer?.querySelector('[data-tutorial="ca-meld-area"]')).toBeInTheDocument();
      expect(footer?.querySelector('[data-tutorial="ca-player-hand"]')).toBeInTheDocument();
      expect(footer?.querySelector('[data-testid="ca-player-hand"]')).toBeInTheDocument();
      expect(footer?.querySelectorAll('[data-testid^="hand-card-"]')).toHaveLength(15);
      expect(footer?.firstElementChild).toHaveClass('flex', 'items-start');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
      window.dispatchEvent(new Event('resize'));
    }
  });

  it('renders human red threes and their Japanese label in the large desktop meld area', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1280 });
    window.dispatchEvent(new Event('resize'));
    try {
      mockExec.mockResolvedValue({
        ...meldPhaseState,
        players: [{ ...basePlayers[0], red3Count: 1, red3s: [{ design: 'HEART', value: 3 }] }, basePlayers[1]],
      });
      renderWithProviders(<CanastaPage />);
      await screen.findByTestId('phase-indicator');
      const footer = document.querySelector('[data-testid="game-footer-content"]');
      const meldArea = footer?.querySelector('[data-tutorial="ca-meld-area"]');
      expect(meldArea).toHaveTextContent('赤3');
      expect(meldArea?.querySelectorAll('[data-testid="animated-card"]')).toHaveLength(1);
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
      window.dispatchEvent(new Event('resize'));
    }
  });

  it('does not mark the CPU meld area as the tutorial target on non-large desktop', async () => {
    breakpointState.enabled = true;
    try {
      mockExec.mockResolvedValue({
        ...meldPhaseState,
        players: [
          basePlayers[0],
          {
            ...basePlayers[1],
            melds: [
              {
                rank: 7,
                isNatural: true,
                isCanasta: false,
                cards: [
                  { design: 'SPADE', value: 7 },
                  { design: 'HEART', value: 7 },
                  { design: 'CLOVER', value: 7 },
                ],
              },
            ],
          },
        ],
      });
      renderWithProviders(<CanastaPage />);
      await screen.findByTestId('phase-indicator');

      const meldAreas = document.querySelectorAll('[data-tutorial="ca-meld-area"]');
      expect(meldAreas).toHaveLength(1);
      const cpuMeldArea = document.querySelector('[data-testid="ca-meld-1-0"]')?.parentElement;
      expect(cpuMeldArea).toBeInTheDocument();
      expect(cpuMeldArea).toHaveTextContent('CPU');
      expect(cpuMeldArea).not.toHaveAttribute('data-tutorial', 'ca-meld-area');
      expect(cpuMeldArea?.querySelector('[data-testid="ca-meld-1-0"]')).toBeInTheDocument();
    } finally {
      breakpointState.enabled = false;
    }
  });

  it('renders the standard card hand on tablet', async () => {
    breakpointState.enabled = true;
    try {
      renderWithProviders(<CanastaPage />);
      await screen.findByTestId('phase-indicator');
      const hand = document.querySelector<HTMLElement>('[data-tutorial="ca-player-hand"]');
      expect(hand).not.toBeNull();
      expect(hand).toHaveClass('flex', 'flex-wrap');
      expect(hand?.querySelector('[data-testid="ca-player-hand"]')).toBeNull();
      expect(hand?.querySelectorAll('button[aria-pressed]')).toHaveLength(basePlayers[0].cards.length);
    } finally {
      breakpointState.enabled = false;
    }
  });

  it('stacks mobile hand rows when the hand spans multiple rows', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    window.dispatchEvent(new Event('resize'));
    try {
      mockExec.mockResolvedValue({
        ...drawPhaseState,
        players: [
          {
            ...basePlayers[0],
            cards: Array.from({ length: 15 }, (_, i) => ({ design: 'SPADE' as const, value: (i % 13) + 1 })),
          },
          basePlayers[1],
        ],
      });
      renderWithProviders(<CanastaPage />);
      await screen.findByTestId('phase-indicator');
      const hand = document.querySelector<HTMLElement>('[data-tutorial="ca-player-hand"]');
      expect(hand).not.toBeNull();
      if (!hand) throw new Error('Expected the Canasta player hand');
      await waitFor(() => expect(hand.querySelectorAll('[data-testid="hand-row"]').length).toBeGreaterThanOrEqual(2));
      for (const row of Array.from(hand.querySelectorAll<HTMLElement>('[data-testid="hand-row"]')).slice(1)) {
        expect(row.style.marginTop).toMatch(/^-/);
      }
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
      window.dispatchEvent(new Event('resize'));
    }
  });

  it('allows selecting a mobile hand card', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    window.dispatchEvent(new Event('resize'));
    try {
      renderWithProviders(<CanastaPage />);
      const hand = await screen.findByTestId('phase-indicator');
      expect(hand).toBeInTheDocument();
      const card = screen
        .getAllByRole('button', { pressed: false })
        .find((button) => button.hasAttribute('aria-pressed'));
      expect(card).toBeDefined();
      if (!card) throw new Error('Expected an unselected hand card');
      fireEvent.click(card);
      expect(card).toHaveAttribute('aria-pressed', 'true');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
      window.dispatchEvent(new Event('resize'));
    }
  });

  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<CanastaPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('shows server score breakdown only at round or game end, including hand penalty as a negative value', async () => {
    const completedState: CanastaResponse = {
      ...roundEndState,
      players: [
        {
          ...basePlayers[0],
          roundScore: 120,
          scoreBreakdown: { meldCards: 70, canastaBonus: 50, red3Bonus: 100, goOutBonus: 100, handPenalty: 35 },
        },
        {
          ...basePlayers[1],
          roundScore: -20,
          scoreBreakdown: { meldCards: 15, canastaBonus: 0, red3Bonus: 0, goOutBonus: 0, handPenalty: 35 },
        },
      ],
    };
    mockExec.mockResolvedValue(completedState);
    const view = renderWithProviders(<CanastaPage />);
    expect(await screen.findAllByText('メルド札')).toHaveLength(2);
    expect(screen.getByText('70')).toBeInTheDocument();
    expect(screen.getAllByText('100')).toHaveLength(2);
    expect(screen.getAllByText('−35')).toHaveLength(2);

    view.unmount();
    mockExec.mockResolvedValue(drawPhaseState);
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getAllByText('ラウンド').length).toBeGreaterThan(0));
    expect(screen.queryByText('メルド札')).not.toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<CanastaPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        cpuDifficulty: 1,
        pointLimit: 5000,
      }),
    );
  });

  it('shows draw phase buttons', async () => {
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    expect(
      within(screen.getByTestId('game-footer-actions')).getByRole('button', { name: '山札から引く' }),
    ).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument();
  });

  it('keeps the hand in the footer and puts reset and keyboard shortcuts in the play area', async () => {
    renderWithProviders(<CanastaPage />);

    const footerContent = await screen.findByTestId('game-footer-content');
    const playArea = document.querySelector('.flex-1.overflow-y-auto');
    expect(playArea).not.toBeNull();
    expect(footerContent.querySelector('[data-tutorial="ca-player-hand"]')).toBeInTheDocument();
    expect(within(footerContent).queryByTestId('canasta-kbd-shortcuts')).not.toBeInTheDocument();
    expect(within(playArea as HTMLElement).getByTestId('canasta-kbd-shortcuts')).toBeInTheDocument();
    expect(within(footerContent).queryByRole('button', { name: 'リセット' })).not.toBeInTheDocument();
    expect(within(playArea as HTMLElement).getByRole('button', { name: 'リセット' })).toBeInTheDocument();
  });

  it('keeps the human meld tutorial target visible when no melds or red threes exist', async () => {
    mockExec.mockResolvedValue(drawPhaseState);
    renderWithProviders(<CanastaPage />);

    await screen.findByText(/メルド/);
    expect(document.querySelector('[data-tutorial="ca-meld-area"]')).toBeInTheDocument();
    expect(screen.queryByTestId(/^ca-meld-\d+-\d+$/)).not.toBeInTheDocument();
    expect(screen.queryByText('赤3')).not.toBeInTheDocument();
  });

  it('calls drawstock command when button clicked', async () => {
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(meldPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('shows meld phase buttons', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'メルドする' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument();
  });

  it('shows the initial-meld minimum and selected total in the meld phase', async () => {
    mockExec.mockResolvedValue({ ...meldPhaseState, minMeld: 90 }); // score 0 → min 50; hasInitMeld false
    renderWithProviders(<CanastaPage />);
    const info = await screen.findByTestId('ca-meld-points');
    expect(info).toHaveTextContent('初回メルド最低点: 90');
    expect(info).toHaveTextContent('選択合計: 0');
  });

  it('shows only the selected total once the initial meld is made', async () => {
    mockExec.mockResolvedValue({
      ...meldPhaseState,
      players: [{ ...basePlayers[0], hasInitMeld: true }, basePlayers[1]],
    });
    renderWithProviders(<CanastaPage />);
    const info = await screen.findByTestId('ca-meld-points');
    expect(info).toHaveTextContent('選択合計: 0');
    expect(info).not.toHaveTextContent('初回メルド最低点');
  });

  it('announces selected meld points in a permanent live region', async () => {
    mockExec.mockResolvedValue({ ...meldPhaseState, minMeld: 90 });
    renderWithProviders(<CanastaPage />);

    const announcement = await screen.findByTestId('ca-meld-points-announcement');
    expect(announcement).toHaveAttribute('role', 'status');
    expect(announcement).toHaveAttribute('aria-live', 'polite');
    expect(announcement).toHaveTextContent('初回メルド最低点: 90 / 選択合計: 0');
    expect(announcement).toHaveClass('sr-only');

    fireEvent.click(screen.getByRole('button', { name: '♠ 7' }));
    await waitFor(() => expect(announcement).toHaveTextContent('初回メルド最低点: 90 / 選択合計: 5'));
    fireEvent.click(screen.getByRole('button', { name: '♠ 7' }));
    await waitFor(() => expect(announcement).toHaveTextContent('初回メルド最低点: 90 / 選択合計: 0'));
  });

  // **最低点に届かない選択でボタンが押せてしまい、サーバのバリデーションで
  // 弾かれて初めて気づく形だった** (#6165)。警告テキストは既に出ていた。
  it('blocks the meld while the selection is under the initial minimum', async () => {
    mockExec.mockResolvedValue(meldPhaseState); // 累計 0 → 最低点 50
    renderWithProviders(<CanastaPage />);
    await screen.findByTestId('ca-meld-points');

    // 7 が 3 枚 = 15 点。初回最低点 50 に届かない。
    fireEvent.click(screen.getByRole('button', { name: '♠ 7' }));
    fireEvent.click(screen.getByRole('button', { name: '♣ 7' }));
    fireEvent.click(screen.getByRole('button', { name: '♥ 7' }));

    await waitFor(() => expect(screen.getByTestId('ca-meld-points')).toHaveTextContent('選択合計: 15'));
    expect(screen.getByRole('button', { name: 'メルドする' })).toBeDisabled();
  });

  // **キーボードの確定も同じ門番を通す** (#6165 レビュー指摘)。ボタンだけ
  // 無効化しても Enter が素通りすれば、直したはずのサーバ拒否が残る。
  it('does not meld from the keyboard while the selection is under the minimum', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<CanastaPage />);
    await screen.findByTestId('ca-meld-points');

    fireEvent.click(screen.getByRole('button', { name: '♠ 7' }));
    fireEvent.click(screen.getByRole('button', { name: '♣ 7' }));
    fireEvent.click(screen.getByRole('button', { name: '♥ 7' }));
    await waitFor(() => expect(screen.getByTestId('ca-meld-points')).toHaveTextContent('選択合計: 15'));

    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'Enter' });
    await waitFor(() => expect(screen.getByTestId('ca-meld-points')).toBeInTheDocument());
    // **どの引数でも meld を送っていないこと。** 引数まで書くと、形が変わった
    // ときに「呼ばれていない」と読み違える。
    expect(mockExec.mock.calls.some((c) => c[0] === 'meld')).toBe(false);

    // **キーボード自体は生きている。** 上の断言が「Enter が届いていない」で
    // 通っていないことを、押せる場面で確かめる。
    fireEvent.click(screen.getByRole('button', { name: '♠ 10' }));
    await waitFor(() => expect(screen.getByTestId('ca-meld-points')).toHaveTextContent('選択合計: 25'));
  });

  // **既に上がっているチームには最低点が無い。** ここまで無効化すると、
  // 開いた後のメルドが打てなくなる。
  it('keeps the meld live once the initial meld is made', async () => {
    mockExec.mockResolvedValue({
      ...meldPhaseState,
      players: [{ ...basePlayers[0], hasInitMeld: true }, basePlayers[1]],
    });
    renderWithProviders(<CanastaPage />);
    await screen.findByTestId('ca-meld-points');

    fireEvent.click(screen.getByRole('button', { name: '♠ 7' }));
    fireEvent.click(screen.getByRole('button', { name: '♣ 7' }));
    fireEvent.click(screen.getByRole('button', { name: '♥ 7' }));

    await waitFor(() => expect(screen.getByTestId('ca-meld-points')).toHaveTextContent('選択合計: 15'));
    expect(screen.getByRole('button', { name: 'メルドする' })).toBeEnabled();
  });

  // 3 枚未満の既存の無効化条件はそのまま。
  it('still blocks a meld of fewer than three cards', async () => {
    mockExec.mockResolvedValue({
      ...meldPhaseState,
      players: [{ ...basePlayers[0], hasInitMeld: true }, basePlayers[1]],
    });
    renderWithProviders(<CanastaPage />);
    await screen.findByTestId('ca-meld-points');

    fireEvent.click(screen.getByRole('button', { name: '♠ 7' }));
    fireEvent.click(screen.getByRole('button', { name: '♣ 7' }));
    expect(screen.getByRole('button', { name: 'メルドする' })).toBeDisabled();
  });

  it('calls skipmeld command when skip button clicked', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skipmeld'));
  });

  it('shows discard phase buttons', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨てる' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '上がる' })).toBeInTheDocument();
  });

  it('disables go out and shows the reason while the player has no canasta', async () => {
    mockExec.mockResolvedValue(discardPhaseState); // hasCanasta: false
    renderWithProviders(<CanastaPage />);
    const goOut = await screen.findByTestId('ca-go-out-button');
    expect(goOut).toBeDisabled();
    const reason = screen.getByTestId('ca-go-out-reason');
    expect(reason).toHaveTextContent('ゴーアウトするには完成したカナスタ（7枚メルド）が1組以上必要です');
    expect(goOut).toHaveAttribute('aria-describedby', 'ca-go-out-reason');
  });

  it('enables go out and hides the reason once a canasta is completed', async () => {
    mockExec.mockResolvedValue({
      ...discardPhaseState,
      players: [{ ...basePlayers[0], hasCanasta: true }, basePlayers[1]],
    });
    renderWithProviders(<CanastaPage />);
    const goOut = await screen.findByTestId('ca-go-out-button');
    expect(goOut).toBeEnabled();
    expect(screen.queryByTestId('ca-go-out-reason')).not.toBeInTheDocument();
    expect(goOut).not.toHaveAttribute('aria-describedby');
  });

  it('shows next round button at round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('shows win celebration at game end', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
  });

  it('renders hint toggle checkbox', async () => {
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    expect(screen.getByRole('checkbox', { name: 'ヒント表示' })).toBeInTheDocument();
  });

  it('shows HintTooltip when hint is enabled in draw phase', async () => {
    localStorage.setItem('hint_enabled_canasta', 'true');
    // drawPhaseState: human turn (currentPlayerIdx=0), DRAW phase → returns drawStock hint
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('shows 次のゲーム at game-end and fires reset directly (no confirm)', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        cpuDifficulty: 1,
        pointLimit: 5000,
      }),
    );
    expect(screen.queryByRole('button', { name: '確認' })).not.toBeInTheDocument();
  });

  it('shows the disabled reason for draw-from-discard when no cards are selected', async () => {
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    const reason = screen.getByTestId('ca-draw-discard-reason');
    expect(reason).toHaveTextContent('手札からトップカードと同ランクの2枚を選択してください');
  });

  it('renders the frozen badge and reason when the discard pile is frozen', async () => {
    mockExec.mockResolvedValue({ ...drawPhaseState, isFrozen: true });
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByTestId('ca-frozen-badge')).toBeInTheDocument());
    expect(screen.getByTestId('ca-draw-discard-reason')).toHaveTextContent(
      'フリーズ中はワイルドカードでの代用ができません',
    );
  });

  it('keeps the frozen status label together while the status row can wrap', async () => {
    mockExec.mockResolvedValue({ ...drawPhaseState, isFrozen: true });
    renderWithProviders(<CanastaPage />);
    const frozenLabel = await screen.findByText('[フリーズ中]');
    expect(frozenLabel.parentElement).toHaveClass('flex', 'flex-wrap');
    expect(frozenLabel).toHaveClass('shrink-0', 'whitespace-nowrap');
  });

  it('switches the reason to selectOneMore once the player selects one card', async () => {
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByTestId('ca-draw-discard-reason')).toBeInTheDocument());
    const handCards = screen.getAllByRole('button', { pressed: false }).filter((b) => b.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    expect(screen.getByTestId('ca-draw-discard-reason')).toHaveTextContent('もう1枚選択してください');
  });

  it('clears the reason and enables the draw button for a legal natural pair', async () => {
    // The pair must be two natural cards of the discard top's rank, so the top is
    // a 7 to match the two sevens the fixture hand opens with.
    mockExec.mockResolvedValue({ ...drawPhaseState, discardTop: { design: 'SPADE', value: 7 } });
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    const handCards = screen.getAllByRole('button', { pressed: false }).filter((b) => b.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[1]);
    expect(screen.queryByTestId('ca-draw-discard-reason')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: '捨て札を取る' })).not.toBeDisabled();
    expect(screen.getByRole('button', { name: '捨て札を取る' })).not.toHaveAttribute('aria-describedby');
  });

  it('warns when more than 2 cards are selected', async () => {
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    const handCards = screen.getAllByRole('button', { pressed: false }).filter((b) => b.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[1]);
    fireEvent.click(handCards[2]);
    expect(screen.getByTestId('ca-draw-discard-reason')).toHaveTextContent('選択は2枚までです');
    expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeDisabled();
  });

  it('keeps the frozen warning visible while the player has only picked one card', async () => {
    mockExec.mockResolvedValue({ ...drawPhaseState, isFrozen: true });
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByTestId('ca-frozen-badge')).toBeInTheDocument());
    const handCards = screen.getAllByRole('button', { pressed: false }).filter((b) => b.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    expect(screen.getByTestId('ca-draw-discard-reason')).toHaveTextContent(
      'フリーズ中はワイルドカードでの代用ができません',
    );
  });

  it('renders melds compactly with a count badge and canasta label, collapsed by default', async () => {
    mockExec.mockResolvedValue(meldDisplayState);
    renderWithProviders(<CanastaPage />);

    const canastaMeld = await screen.findByTestId('ca-meld-0-0');
    // Collapsed by default so a meld occupies ~one card of height.
    expect(canastaMeld).not.toHaveAttribute('open');
    // Count badge shows the number of cards in the meld.
    expect(screen.getByTestId('ca-meld-badge-0-0')).toHaveTextContent('7');
    expect(screen.getByTestId('ca-meld-badge-0-1')).toHaveTextContent('3');
    expect(document.querySelector('[data-tutorial="ca-meld-area"]')).toBeInTheDocument();
    // Canasta type (with the ★) is visible at a glance for the completed canasta.
    expect(within(canastaMeld).getByText(/ナチュラルカナスタ/)).toBeInTheDocument();
  });

  it('renders the Japanese label for a mixed canasta', async () => {
    mockExec.mockResolvedValue({
      ...meldDisplayState,
      players: [
        {
          ...basePlayers[0],
          hasCanasta: true,
          melds: [
            {
              rank: 7,
              isNatural: false,
              isCanasta: true,
              cards: Array.from({ length: 7 }, (_, i) => ({
                design: i % 2 ? ('HEART' as const) : ('SPADE' as const),
                value: 7,
              })),
            },
          ],
        },
        basePlayers[1],
      ],
    });
    renderWithProviders(<CanastaPage />);

    const mixedMeld = await screen.findByTestId('ca-meld-0-0');
    expect(within(mixedMeld).getByText(/ミックスカナスタ/)).toBeInTheDocument();
  });

  it('expands a meld to reveal every constituent card when the summary is clicked', async () => {
    mockExec.mockResolvedValue(meldDisplayState);
    renderWithProviders(<CanastaPage />);

    const canastaMeld = (await screen.findByTestId('ca-meld-0-0')) as HTMLDetailsElement;
    expect(canastaMeld.open).toBe(false);
    const summary = canastaMeld.querySelector('summary');
    if (!summary) throw new Error('meld summary not found');
    fireEvent.click(summary);
    expect(canastaMeld.open).toBe(true);
  });

  afterEach(() => {
    localStorage.clear();
  });

  it('refuses a two-card selection the server would reject', async () => {
    localStorage.clear();
    mockExec.mockReset();
    // Discard top is a 5; the hand's sevens are two cards but not a legal pair.
    mockExec.mockResolvedValue(drawPhaseState);
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    const handCards = screen.getAllByRole('button', { pressed: false }).filter((b) => b.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[1]);
    // Counting to two used to be enough to enable it.
    expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeDisabled();
    expect(screen.getByTestId('ca-draw-discard-reason')).toHaveTextContent('同じランク');
  });

  it('names the wildcard as the problem when one is picked', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue({
      ...drawPhaseState,
      discardTop: { design: 'SPADE', value: 7 },
      players: [
        {
          ...drawPhaseState.players[0],
          cards: [
            { design: 'SPADE', value: 7 },
            { design: 'HEART', value: 2 },
          ],
        },
        ...drawPhaseState.players.slice(1),
      ],
    });
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    const handCards = screen.getAllByRole('button', { pressed: false }).filter((b) => b.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[1]);
    expect(screen.getByTestId('ca-draw-discard-reason')).toHaveTextContent('ワイルドカード');
  });

  it('renders CPU hand card count in Japanese at round end without raw "cards"', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CanastaPage />);
    const cpuHandInfo = await screen.findByText('CPU 1: 15枚');
    expect(cpuHandInfo).toBeInTheDocument();
    expect(cpuHandInfo).not.toHaveTextContent('cards');
    expect(screen.queryByText(/15 cards/)).not.toBeInTheDocument();
  });

  it('renders every discard pile card in oldest-to-newest order', async () => {
    mockExec.mockResolvedValue({
      ...drawPhaseState,
      discardTop: { design: 'HEART', value: 8 },
      discardPile: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 8 },
      ],
    });
    renderWithProviders(<CanastaPage />);
    const pile = await screen.findByTestId('ca-discard-pile');
    const cards = within(pile).getAllByTestId('animated-card');
    expect(cards).toHaveLength(2);
    expect(cards[0].parentElement).not.toHaveAttribute('data-discard-top');
    expect(cards[1].parentElement).toHaveAttribute('data-discard-top', 'true');
    expect(cards[1].parentElement).toHaveTextContent('捨て札の一番上');
    expect(within(pile).getByAltText('♠ A')).toBeInTheDocument();
    expect(within(pile).getByAltText('♥ 8')).toBeInTheDocument();
  });

  it('shows the frozen badge with a multi-card discard pile', async () => {
    mockExec.mockResolvedValue({
      ...drawPhaseState,
      isFrozen: true,
      discardPile: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 8 },
      ],
    });
    renderWithProviders(<CanastaPage />);
    const pile = await screen.findByTestId('ca-discard-pile');
    expect(within(pile).getAllByTestId('animated-card')).toHaveLength(2);
    expect(within(pile).getByTestId('ca-frozen-badge')).toBeInTheDocument();
  });

  it('renders the discard pile container when the pile is empty', async () => {
    mockExec.mockResolvedValue({ ...drawPhaseState, discardTop: null, discardPile: [], discardPileCount: 0 });
    renderWithProviders(<CanastaPage />);
    const pile = await screen.findByTestId('ca-discard-pile');
    expect(within(pile).getByText('捨て札')).toBeInTheDocument();
    expect(within(pile).queryByTestId('animated-card')).not.toBeInTheDocument();
  });

  it('shows the translated phase name, not the raw phase key', async () => {
    renderWithProviders(<CanastaPage />);
    expect(await screen.findByText('ドロー', { selector: 'strong' })).toBeInTheDocument();
    expect(screen.queryByText('draw', { selector: 'strong' })).not.toBeInTheDocument();
  });
});
