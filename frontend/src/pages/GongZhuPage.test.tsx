import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { gongzhuApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeGongZhuState } from '../test/stateFactories';
import { GongZhuPage } from './GongZhuPage';

vi.mock('../api/gameApi', () => ({
  gongzhuApi: { exec: vi.fn() },
  actionLogApi: { gongzhu: vi.fn() },
}));

const mockExec = vi.mocked(gongzhuApi.exec);

const playPhaseState = makeGongZhuState();
const exposePhaseState = makeGongZhuState({ phase: 0, trickNumber: 0, exposableIndices: [0, 1] });
const trickEndState = makeGongZhuState({
  phase: 2,
  currentTrick: [
    { playerIdx: 0, card: { design: 'DIAMOND', value: 3 } },
    { playerIdx: 1, card: { design: 'HEART', value: 5 } },
  ],
});
const roundEndState = makeGongZhuState({ phase: 3 });
const gameEndState = makeGongZhuState({ phase: 4, gameEndFlag: true, winnerIdx: 0, message: 'ゲーム終了！' });
const cpuTurnState = makeGongZhuState({ currentPlayerIdx: 1 });

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
});

describe('GongZhuPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<GongZhuPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with config', async () => {
    renderWithProviders(<GongZhuPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 1000,
      }),
    );
  });

  it('renders play phase with human cards', async () => {
    renderWithProviders(<GongZhuPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♦ J')).toBeInTheDocument();
    });
  });

  it('renders expose phase with expose button', async () => {
    mockExec.mockResolvedValue(exposePhaseState);
    renderWithProviders(<GongZhuPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '公開しない' })).toBeInTheDocument());
  });

  it('names the exposable cards (index + card name) in a role=status hint', async () => {
    mockExec.mockResolvedValue(exposePhaseState); // exposableIndices [0, 1] → ♠ Q, ♦ J
    const { container } = renderWithProviders(<GongZhuPage />);
    await waitFor(() => expect(container.querySelector('[data-tutorial="gz-expose-area"]')).not.toBeNull());
    const hint = container.querySelector('[data-tutorial="gz-expose-area"]') as HTMLElement;
    expect(hint).toHaveAttribute('role', 'status');
    expect(hint.textContent).toContain('[0] ♠ Q');
    expect(hint.textContent).toContain('[1] ♦ J');
  });

  it('shows exposed point cards with localized symbols and an aria-label', async () => {
    mockExec.mockResolvedValue(makeGongZhuState({ exposed: { pig: true, sheep: true, ace: true, doubler: true } }));
    renderWithProviders(<GongZhuPage />);
    const summary = await screen.findByTestId('exposure-summary');
    expect(summary).toHaveTextContent('♠Q');
    expect(summary).toHaveTextContent('♦J');
    expect(summary).toHaveTextContent('♥A');
    expect(summary).toHaveTextContent('♣10');
    expect(summary).toHaveAttribute('aria-label');
  });

  it('shows the exposed-none label when nothing is exposed', async () => {
    mockExec.mockResolvedValue(makeGongZhuState({ exposed: { pig: false, sheep: false, ace: false, doubler: false } }));
    renderWithProviders(<GongZhuPage />);
    const summary = await screen.findByTestId('exposure-summary');
    expect(summary).toHaveTextContent('なし');
  });

  it('highlights the exposable cards during the expose phase', async () => {
    // Human hand has 2 cards; only index 0 is exposable so index 1 is the dimmed one.
    mockExec.mockResolvedValue(makeGongZhuState({ phase: 0, trickNumber: 0, exposableIndices: [0] }));
    const { container } = renderWithProviders(<GongZhuPage />);
    const hand = await waitFor(() => {
      const el = container.querySelector('[data-tutorial="gz-player-hand"]');
      if (!el) throw new Error('hand not rendered yet');
      return el as HTMLElement;
    });
    const cards = hand.querySelectorAll('button');
    // The exposable card (idx 0) carries the warning glow.
    expect((cards[0] as HTMLElement).style.boxShadow).toContain('rgba(232, 146, 58');
    // The non-exposable card (idx 1) is dimmed.
    expect(cards[1]).toHaveClass('opacity-60');
  });

  it('does not dim any card when there are no exposable cards', async () => {
    mockExec.mockResolvedValue(makeGongZhuState({ phase: 0, trickNumber: 0, exposableIndices: [] }));
    const { container } = renderWithProviders(<GongZhuPage />);
    const hand = await waitFor(() => {
      const el = container.querySelector('[data-tutorial="gz-player-hand"]');
      if (!el) throw new Error('hand not rendered yet');
      return el as HTMLElement;
    });
    for (const card of hand.querySelectorAll('button')) expect(card).not.toHaveClass('opacity-60');
  });

  it('expose button dispatches expose with empty selection', async () => {
    mockExec.mockResolvedValue(exposePhaseState);
    renderWithProviders(<GongZhuPage />);
    const btn = await screen.findByRole('button', { name: '公開しない' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('expose', []));
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<GongZhuPage />);
    const card = await screen.findByAltText('♠ Q');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 0));
  });

  it('renders trick end with next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<GongZhuPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with next round button', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<GongZhuPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('renders game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<GongZhuPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！')).toBeInTheDocument());
  });

  it('shows each player captured point cards summary in the score table', async () => {
    mockExec.mockResolvedValue(
      makeGongZhuState({
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 11,
            cards: [{ design: 'SPADE', value: 12 }],
            capturedPointCards: [
              { design: 'SPADE', value: 12 }, // pig
              { design: 'HEART', value: 5 },
              { design: 'HEART', value: 13 },
            ],
            roundScore: 0,
            cumulativeScore: 0,
            trickCount: 1,
          },
          {
            id: 1,
            isHuman: false,
            cardCount: 11,
            cards: [],
            capturedPointCards: [{ design: 'DIAMOND', value: 11 }], // sheep
            roundScore: 0,
            cumulativeScore: 0,
            trickCount: 1,
          },
          {
            id: 2,
            isHuman: false,
            cardCount: 12,
            cards: [],
            capturedPointCards: [],
            roundScore: 0,
            cumulativeScore: 0,
            trickCount: 0,
          },
          {
            id: 3,
            isHuman: false,
            cardCount: 12,
            cards: [],
            capturedPointCards: [],
            roundScore: 0,
            cumulativeScore: 0,
            trickCount: 0,
          },
        ],
      }),
    );
    renderWithProviders(<GongZhuPage />);
    const p0 = await screen.findByTestId('gz-captured-0');
    expect(p0).toHaveTextContent('♠Q');
    expect(p0).toHaveTextContent('♥×2');
    expect(p0).toHaveAttribute('aria-label');
    expect(screen.getByTestId('gz-captured-1')).toHaveTextContent('♦J');
    // A player with no captured point cards shows the "none" placeholder.
    expect(screen.getByTestId('gz-captured-2')).toHaveTextContent('なし');
  });

  it('does not show play button on CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<GongZhuPage />);
    await waitFor(() => expect(screen.getByAltText('♠ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  // **マストフォローの可視化 (#4812)。**どのカードが出せるかを画面が一切
  // 示しておらず、プレイヤーが自力で判断するしかなかった。
  it('disables the hand cards that cannot follow suit', async () => {
    mockExec.mockResolvedValue(makeGongZhuState({ playableIndices: [0] }));
    renderWithProviders(<GongZhuPage />);

    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // 出せない札は aria-disabled になる (PlayerHandSection の validIndices の規約)。
    await waitFor(() => expect(document.querySelectorAll('[aria-disabled="true"]').length).toBe(1));
  });
});
