import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { sixcardgolfApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, SixCardGolfResponse, SixCardGolfSlot } from '../types/card';
import { SixCardGolfPage } from './SixCardGolfPage';

vi.mock('../api/gameApi', () => ({
  sixcardgolfApi: { exec: vi.fn() },
  actionLogApi: { sixcardgolf: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(sixcardgolfApi.exec);

const card = (value: number): Card => ({ design: 'SPADE', value }) as unknown as Card;
const slot = (value: number): SixCardGolfSlot => ({ card: card(value), faceUp: true });

function makeState(overrides: Partial<SixCardGolfResponse> = {}): SixCardGolfResponse {
  const grid = [slot(5), slot(3), slot(7), slot(5), slot(9), slot(2)];
  return {
    players: [
      { id: 0, isHuman: true, grid, roundScore: 28, cumulativeScore: 28, allFaceUp: true },
      { id: 1, isHuman: false, grid: [...grid], roundScore: 10, cumulativeScore: 10, allFaceUp: true },
    ],
    phase: 3, // SCG_PHASE_ROUND_OVER
    roundNumber: 1,
    totalRounds: 9,
    currentPlayerIdx: 0,
    discardTop: card(4),
    drawPileCount: 20,
    drawnCard: null,
    drawnFromDiscard: false,
    canFlip: false,
    finalTurnTrigger: -1,
    gameEndFlag: false,
    winnerIdx: -1,
    config: { playerCount: 2, cpuDifficulty: 1, rounds: 9 },
    message: '',
    ...overrides,
  };
}

beforeEach(() => {
  localStorage.clear();
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('SixCardGolfPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<SixCardGolfPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith({ command: 'reset' }));
  });

  it('shows per-column score badges at round over, marking matched pairs', async () => {
    renderWithProviders(<SixCardGolfPage />);
    const breakdown = await screen.findByTestId('scg-column-scores');
    expect(breakdown).toBeInTheDocument();
    // Column 0 is a 5-over-5 pair → 0pt with success highlight.
    const col0 = screen.getByTestId('scg-column-score-0');
    expect(col0.className).toContain('bg-ds-success');
    // Column 1 (3 over 9) is not a pair → no success highlight.
    expect(screen.getByTestId('scg-column-score-1').className).not.toContain('bg-ds-success');
  });

  it('localizes the face-down grid slot aria-label', async () => {
    const faceDownGrid = [
      { card: null, faceUp: false } as unknown as SixCardGolfSlot,
      slot(3),
      slot(7),
      slot(5),
      slot(9),
      slot(2),
    ];
    mockExec.mockResolvedValue(
      makeState({
        phase: 1, // player turn
        players: [
          { id: 0, isHuman: true, grid: faceDownGrid, roundScore: 0, cumulativeScore: 0, allFaceUp: false },
          { id: 1, isHuman: false, grid: [...faceDownGrid], roundScore: 0, cumulativeScore: 0, allFaceUp: false },
        ],
      }),
    );
    renderWithProviders(<SixCardGolfPage />);
    // Position 1 (0-based index 0 → 1-based), face down → localized ja label.
    await waitFor(() => expect(screen.getAllByRole('button', { name: '位置1（裏向き）' }).length).toBeGreaterThan(0));
    // A face-up slot still reads its card name.
    expect(screen.getAllByRole('button', { name: '♠ 3' }).length).toBeGreaterThan(0);
  });

  it('shows the human column breakdown during active play, marking uncertain columns', async () => {
    const faceDownGrid = [
      slot(5),
      slot(3),
      slot(7),
      { card: null, faceUp: false } as unknown as SixCardGolfSlot, // column 0 bottom still hidden
      slot(9),
      slot(2),
    ];
    mockExec.mockResolvedValue(
      makeState({
        phase: 1, // player turn
        players: [
          { id: 0, isHuman: true, grid: faceDownGrid, roundScore: 0, cumulativeScore: 0, allFaceUp: false },
          { id: 1, isHuman: false, grid: [...faceDownGrid], roundScore: 0, cumulativeScore: 0, allFaceUp: false },
        ],
      }),
    );
    renderWithProviders(<SixCardGolfPage />);
    // The breakdown is now visible mid-play.
    await screen.findByTestId('scg-column-scores');
    // Column 0 has a hidden bottom card → uncertain "+?" display; column 1 is fully revealed.
    expect(screen.getByTestId('scg-column-score-0')).toHaveTextContent('+?');
    expect(screen.getByTestId('scg-column-score-1')).not.toHaveTextContent('+?');
  });

  describe('hinted slot highlight', () => {
    it('marks the slot the hint points at', async () => {
      vi.mocked(useGameHint).mockReturnValue({
        hint: { targetAction: 'swap', reason: 'hintReason.swapHigh', confidence: 'strong', targetPos: 1 },
        hintEnabled: true,
        setHintEnabled: vi.fn(),
      });
      renderWithProviders(<SixCardGolfPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());

      const hinted = document.querySelectorAll('[data-hinted="true"]');
      expect(hinted).toHaveLength(1);
    });

    it('marks nothing while hints are off', async () => {
      vi.mocked(useGameHint).mockReturnValue({
        hint: { targetAction: 'swap', reason: 'hintReason.swapHigh', confidence: 'strong', targetPos: 1 },
        hintEnabled: false,
        setHintEnabled: vi.fn(),
      });
      renderWithProviders(<SixCardGolfPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());
      expect(document.querySelectorAll('[data-hinted="true"]')).toHaveLength(0);
    });

    it('marks nothing when the hint carries no position', async () => {
      vi.mocked(useGameHint).mockReturnValue({
        hint: { targetAction: 'discard', reason: 'hintReason.discardHigh', confidence: 'strong' },
        hintEnabled: true,
        setHintEnabled: vi.fn(),
      });
      renderWithProviders(<SixCardGolfPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());
      expect(document.querySelectorAll('[data-hinted="true"]')).toHaveLength(0);
    });
  });

  // CUI には sd/sp/sr があり 2〜4 人戦も短縮ラウンドも遊べるのに、Web には設定
  // パネルが無く常に既定 (2人・9ラウンド) で始まっていた (#5616)。
  describe('settings', () => {
    it('starts a new game with the chosen player count', async () => {
      renderWithProviders(<SixCardGolfPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith({ command: 'reset' }));
      mockExec.mockClear();

      fireEvent.change(screen.getByTestId('scg-player-count'), { target: { value: '4' } });
      fireEvent.click(screen.getByTestId('scg-apply-settings'));

      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith({
          command: 'reset',
          config: { playerCount: 4, cpuDifficulty: 1, rounds: 9 },
        }),
      );
    });

    it('starts a new game with the chosen round count', async () => {
      renderWithProviders(<SixCardGolfPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith({ command: 'reset' }));
      mockExec.mockClear();

      fireEvent.change(screen.getByTestId('scg-rounds'), { target: { value: '3' } });
      fireEvent.click(screen.getByTestId('scg-apply-settings'));

      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith({
          command: 'reset',
          config: { playerCount: 2, cpuDifficulty: 1, rounds: 3 },
        }),
      );
    });

    it('offers every player count the domain accepts', async () => {
      renderWithProviders(<SixCardGolfPage />);
      // state が来るまではスケルトンなので、待たずに問い合わせると
      // 「選択肢が無い」ではなく「要素が無い」で落ちる。
      await waitFor(() => expect(screen.getByTestId('scg-player-count')).toBeInTheDocument());
      const opts = Array.from(screen.getByTestId('scg-player-count').querySelectorAll('option')).map((o) => o.value);
      // SixCardGolfConfig の PlayerCount は 2〜4。ここが増減したら合わせること。
      expect(opts).toEqual(['2', '3', '4']);
    });

    it('renders a grid for every seat when more than two play', async () => {
      const grid = [slot(5), slot(3), slot(7), slot(5), slot(9), slot(2)];
      mockExec.mockResolvedValue(
        makeState({
          players: [0, 1, 2, 3].map((id) => ({
            id,
            isHuman: id === 0,
            grid: [...grid],
            roundScore: 5,
            cumulativeScore: 5,
            allFaceUp: true,
          })),
          config: { playerCount: 4, cpuDifficulty: 1, rounds: 9 },
        }),
      );
      renderWithProviders(<SixCardGolfPage />);
      // 4 席ぶん描かれること。2 席で固定なら、ここで落ちる。
      // prefix は scg-seat- 。scg-player-count が同じ正規表現に引っかかると
      // 席が 1 つ多く数えられ、2 席固定でも通ってしまう。
      await waitFor(() => expect(screen.getAllByTestId(/^scg-seat-/).length).toBe(4));
    });
  });
});
