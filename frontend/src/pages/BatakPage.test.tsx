import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { batakApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeBatakState } from '../test/stateFactories';
import type { BatakResponse } from '../types/card';
import { BatakPage } from './BatakPage';

vi.mock('../api/gameApi', () => ({
  batakApi: { exec: vi.fn() },
  actionLogApi: { batak: vi.fn() },
}));

const mockExec = vi.mocked(batakApi.exec);

const playPhaseState = makeBatakState({ validPlayIndices: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12] });

const bidPhaseState = makeBatakState({
  phase: 0,
  bidPlayerIdx: 0,
  declarerIdx: -1,
  highBid: 0,
  minLegalBid: 5,
  players: makeBatakState().players.map((p) => ({ ...p, bid: -1 })),
});

const bidPhaseCpuTurnState = makeBatakState({
  ...bidPhaseState,
  bidPlayerIdx: 1,
});

const trickEndState = makeBatakState({
  phase: 2,
  currentTrick: [
    { playerIdx: 0, card: { design: 'DIAMOND', value: 3 } },
    { playerIdx: 1, card: { design: 'HEART', value: 5 } },
  ],
});

const roundEndState = makeBatakState({ phase: 3 });

const gameEndState = makeBatakState({
  phase: 4,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
});

const gameEndByFlagState = makeBatakState({
  phase: 1,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
});

const spadesBrokenState = makeBatakState({ spadesBroken: true });

const cpuTurnState = makeBatakState({ currentPlayerIdx: 1 });

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
});

describe('BatakPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<BatakPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with default config', async () => {
    renderWithProviders(<BatakPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        maxRounds: 5,
      }),
    );
  });

  it('renders play phase with human cards', async () => {
    renderWithProviders(<BatakPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♥ J')).toBeInTheDocument();
    });
  });

  it('renders bid buttons only from minLegalBid to 13, and does not show buttons below minLegalBid', async () => {
    const customBidState = makeBatakState({
      ...bidPhaseState,
      minLegalBid: 6,
      highBid: 5,
    });
    mockExec.mockResolvedValue(customBidState);
    renderWithProviders(<BatakPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'ビッド' })).toBeInTheDocument();
    });
    // Buttons below minLegalBid (1..5) MUST NOT be in the document!
    expect(screen.queryByTestId('bid-option-1')).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-option-2')).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-option-3')).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-option-4')).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-option-5')).not.toBeInTheDocument();
    // Buttons from minLegalBid to 13 are present, defaulting to minLegalBid pressed.
    expect(screen.getByTestId('bid-option-6')).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByTestId('bid-option-13')).toBeInTheDocument();
    expect(screen.getByTestId('bid-pass')).toBeInTheDocument();
  });

  it('renders only the pass button and no bid number buttons when minLegalBid is 0', async () => {
    const passOnlyState = makeBatakState({
      ...bidPhaseState,
      minLegalBid: 0,
      highBid: 13,
    });
    mockExec.mockResolvedValue(passOnlyState);
    renderWithProviders(<BatakPage />);
    await waitFor(() => {
      expect(screen.getByTestId('bid-pass')).toBeInTheDocument();
    });
    expect(screen.queryByRole('button', { name: 'ビッド' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-option-5')).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-option-13')).not.toBeInTheDocument();
  });

  it('calls bid command with 0 when pass button is clicked', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<BatakPage />);
    await waitFor(() => expect(screen.getByTestId('bid-pass')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByTestId('bid-pass'));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 0));
  });

  it('shows bid phase instruction when human bid turn', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<BatakPage />);
    await waitFor(() => {
      expect(screen.getByText('ビッド宣言 (5-13 または パス)')).toBeInTheDocument();
    });
  });

  it('does not show bid instruction on CPU bid turn', async () => {
    mockExec.mockResolvedValue(bidPhaseCpuTurnState);
    renderWithProviders(<BatakPage />);
    await waitFor(() => expect(screen.getAllByText(/CPU 1/).length).toBeGreaterThan(0));
    expect(screen.queryByText(/ビッド宣言/)).not.toBeInTheDocument();
  });

  it('calls bid command when bid button is clicked', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<BatakPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ビッド' })).toBeInTheDocument());

    // Select bid 7 from the button group.
    fireEvent.click(screen.getByTestId('bid-option-7'));
    expect(screen.getByTestId('bid-option-7')).toHaveAttribute('aria-pressed', 'true');
    // The live region reflects the current selection for screen readers.
    expect(screen.getByTestId('batak-bid-selected')).toHaveTextContent('7');

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ビッド' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 7));
  });

  it('play button disabled when not 1 card selected', async () => {
    renderWithProviders(<BatakPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeDisabled());
  });

  it('play button enabled when 1 card selected', async () => {
    renderWithProviders(<BatakPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: '出す' })).not.toBeDisabled();
  });

  it('marks cards outside validPlayIndices as aria-disabled with tooltip on human play turn', async () => {
    const must = makeBatakState({ validPlayIndices: [0] });
    mockExec.mockResolvedValue(must);
    renderWithProviders(<BatakPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    const allowed = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    const blocked = screen.getByAltText('♥ J').closest('button') as HTMLButtonElement;

    expect(allowed).not.toHaveAttribute('aria-disabled');
    expect(blocked).toHaveAttribute('aria-disabled', 'true');
    // Critical accessibility constraint: aria-disabled cards must stay focusable
    // so keyboard / screen-reader users can reach the tooltip explaining the rule.
    expect(blocked).not.toBeDisabled();
    expect(blocked).toHaveAttribute(
      'title',
      'このカードは出せません (リードスートに従うか、ボイドならスペードで切ってください)',
    );

    fireEvent.click(blocked);
    expect(screen.getByRole('button', { name: '出す' })).toBeDisabled();
  });

  it('does not show play button when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<BatakPage />);
    await waitFor(() => expect(screen.getAllByText(/CPU 1/).length).toBeGreaterThan(0));
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('shows next trick button on trick end', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<BatakPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('shows next round button on round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<BatakPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('shows game end with action log button (phase 4)', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<BatakPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows game end via gameEndFlag with non-4 phase', async () => {
    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<BatakPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('renders spades-broken status text', async () => {
    mockExec.mockResolvedValue(spadesBrokenState);
    renderWithProviders(<BatakPage />);
    await waitFor(() => expect(screen.getAllByText('スペードブレイク済').length).toBeGreaterThan(0));
  });

  it('shows a persistent spades-break indicator in the footer hand area', async () => {
    mockExec.mockResolvedValue(spadesBrokenState);
    renderWithProviders(<BatakPage />);
    await waitFor(() => {
      expect(screen.getByTestId('batak-spades-break-footer')).toHaveTextContent('スペードブレイク済');
    });
  });

  it('shows raw integer scores in the score table', async () => {
    const scoreState = makeBatakState({
      players: [
        { ...makeBatakState().players[0], roundScore: 5, cumulativeScore: 12 },
        { ...makeBatakState().players[1], roundScore: 2, cumulativeScore: 4 },
        { ...makeBatakState().players[2], roundScore: 0, cumulativeScore: -5 },
        { ...makeBatakState().players[3], roundScore: 3, cumulativeScore: 8 },
      ],
    });
    mockExec.mockResolvedValue(scoreState);
    renderWithProviders(<BatakPage />);
    await waitFor(() => {
      expect(screen.getByText('12')).toBeInTheDocument();
      expect(screen.getByText('-5')).toBeInTheDocument();
    });
    expect(screen.queryByText('1.2')).not.toBeInTheDocument();
  });

  it('renders declarer information on the page', async () => {
    const declarerState = makeBatakState({
      declarerIdx: 0,
      highBid: 7,
    });
    mockExec.mockResolvedValue(declarerState);
    renderWithProviders(<BatakPage />);
    await waitFor(() => expect(screen.getByTestId('batak-declarer')).toBeInTheDocument());
    expect(screen.getByTestId('batak-declarer')).toHaveTextContent('親: あなた');
  });

  it('displays pass as パス rather than 0 in player rows and score table', async () => {
    const passedState = makeBatakState({
      players: [
        { ...makeBatakState().players[0], bid: 6 },
        { ...makeBatakState().players[1], bid: 0 },
        { ...makeBatakState().players[2], bid: -1 },
        { ...makeBatakState().players[3], bid: 0 },
      ],
      declarerIdx: 0,
    });
    mockExec.mockResolvedValue(passedState);
    renderWithProviders(<BatakPage />);
    await waitFor(() => {
      expect(screen.getAllByText('パス').length).toBeGreaterThan(0);
    });
  });

  it('renders BidProgressBar only for the declarer', async () => {
    const declarerState = makeBatakState({
      declarerIdx: 0,
      players: [
        { ...makeBatakState().players[0], id: 0, bid: 6, trickCount: 2 },
        { ...makeBatakState().players[1], id: 1, bid: 0, trickCount: 1 },
      ],
    });
    mockExec.mockResolvedValue(declarerState);
    renderWithProviders(<BatakPage />);
    await waitFor(() => expect(screen.getByTestId('bid-progress-0')).toBeInTheDocument());
    expect(screen.queryByTestId('bid-progress-1')).not.toBeInTheDocument();
  });

  it('renders BidProgressBar for CPU declarer and never for defender players even with positive bids', async () => {
    const base = makeBatakState();
    const cpuDeclarerState = makeBatakState({
      declarerIdx: 1,
      players: base.players.map((p) =>
        p.id === 1 ? { ...p, bid: 6, trickCount: 2 } : { ...p, bid: 5, trickCount: 1 },
      ),
    });
    mockExec.mockResolvedValue(cpuDeclarerState);
    renderWithProviders(<BatakPage />);

    await waitFor(() => expect(screen.getByTestId('bid-progress-1')).toBeInTheDocument());
    expect(screen.queryByTestId('bid-progress-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-progress-2')).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-progress-3')).not.toBeInTheDocument();
  });

  it('shows error alert when API rejects', async () => {
    renderWithProviders(<BatakPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('settings panel changes cpuDifficulty', async () => {
    renderWithProviders(<BatakPage />);
    await waitFor(() => expect(screen.getByText('バタック')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: '2' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 2,
        maxRounds: 5,
      }),
    );
  });

  it('settings panel changes maxRounds', async () => {
    renderWithProviders(<BatakPage />);
    await waitFor(() => expect(screen.getByText('バタック')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[1], { target: { value: '10' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        maxRounds: 10,
      }),
    );
  });

  // このページのヒントは `hintAvailable` を使わないので、読み上げガードが
  // 一度も見ておらず aria-live が無いまま出荷されていた (#6663)。領域は
  // **常設**で、分岐ごとにその分岐だけが出す語を見る。
  describe('BatakPage hint live region', () => {
    it('is mounted and empty before any hint arrives', async () => {
      mockExec.mockResolvedValue(playPhaseState);
      renderWithProviders(<BatakPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());

      const region = screen.getByTestId('batak-hint-live');
      expect(region).toHaveAttribute('role', 'status');
      expect(region).toHaveAttribute('aria-live', 'polite');
      expect(region).toBeEmptyDOMElement();
    });

    it('names a bid recommendation', async () => {
      mockExec.mockResolvedValue(bidPhaseState);
      renderWithProviders(<BatakPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

      mockExec.mockResolvedValue({
        ...bidPhaseState,
        hint: { bid: 5, reason: 'strategic_bid' },
      } as unknown as BatakResponse);
      fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

      const region = await screen.findByTestId('batak-hint-live');
      await waitFor(() => expect(region).toHaveTextContent('推奨ビッド: 5 (戦略的なビッド)'));
      expect(region).not.toHaveTextContent('{{');
    });

    it('names pass recommendation when hint.bid is 0', async () => {
      mockExec.mockResolvedValue(bidPhaseState);
      renderWithProviders(<BatakPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

      mockExec.mockResolvedValue({
        ...bidPhaseState,
        hint: { bid: 0, reason: 'pass_weak_hand' },
      } as unknown as BatakResponse);
      fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

      const region = await screen.findByTestId('batak-hint-live');
      await waitFor(() => expect(region).toHaveTextContent('推奨ビッド: パス (手札が弱いためパス)'));
      expect(region).not.toHaveTextContent('0');
      expect(region).not.toHaveTextContent('{{');
    });

    it('names the card to play with suit and rank', async () => {
      mockExec.mockResolvedValue(playPhaseState);
      renderWithProviders(<BatakPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

      mockExec.mockResolvedValue({
        ...playPhaseState,
        hint: { cardIndex: 0, reason: 'lead_strong' },
      } as unknown as BatakResponse);
      fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

      const region = await screen.findByTestId('batak-hint-live');
      await waitFor(() => expect(region).toHaveTextContent('推奨カード: ♠ A [0] (強いカードでリード)'));
      expect(region).toHaveTextContent('♠ A');
      expect(region).toHaveTextContent('[0]');
      expect(region).not.toHaveTextContent('推奨ビッド');
      expect(region).not.toHaveTextContent('{{');
      expect(region).not.toHaveTextContent('{{card}}');
      expect(region).not.toHaveTextContent('{{idx}}');
    });

    it('falls back to dash when hint.cardIndex exceeds hand length', async () => {
      mockExec.mockResolvedValue(playPhaseState);
      renderWithProviders(<BatakPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

      mockExec.mockResolvedValue({
        ...playPhaseState,
        hint: { cardIndex: 99, reason: 'lead_strong' },
      } as unknown as BatakResponse);
      fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

      const region = await screen.findByTestId('batak-hint-live');
      await waitFor(() => expect(region).toHaveTextContent('推奨カード: - [99] (強いカードでリード)'));
      expect(region).toHaveTextContent('[99]');
      expect(region).toHaveTextContent('-');
      expect(region).not.toHaveTextContent('{{');
    });
  });
});
