import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { presidentApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, PresidentResponse } from '../types/card';
import { PresidentPage } from './PresidentPage';

vi.mock('../api/gameApi', () => ({
  presidentApi: { exec: vi.fn() },
  actionLogApi: { president: vi.fn() },
}));

const mockExec = vi.mocked(presidentApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<PresidentResponse> = {}): PresidentResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        isFinished: false,
        rank: -1,
        cardCount: 3,
        cards: [card('SPADE', 3), card('HEART', 5), card('DIAMOND', 7)],
      },
      { id: 1, isHuman: false, isFinished: false, rank: -1, cardCount: 13, cards: [] },
      { id: 2, isHuman: false, isFinished: false, rank: -1, cardCount: 13, cards: [] },
      { id: 3, isHuman: false, isFinished: false, rank: -1, cardCount: 13, cards: [] },
    ],
    currentTurn: 0,
    tableCards: [],
    lastPlayPlayerIdx: -1,
    gameEndFlag: false,
    revolutionActive: false,
    config: {
      revolutionEnabled: true,
      cardExchangeEnabled: true,
      passFieldFlushEnabled: true,
      cpuDifficulty: 1,
    },
    exchangeActions: [],
    cpuActions: [],
    humanAction: null,
    message: '',
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('PresidentPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders the human hand', async () => {
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('hand-card-1')).toBeInTheDocument();
    expect(screen.getByTestId('hand-card-2')).toBeInTheDocument();
  });

  it('gives each display-only table card an accessible name', async () => {
    // Table cards not present in the human hand (S3/H5/D7) so the labels are unambiguous.
    mockExec.mockResolvedValue(makeState({ tableCards: [card('SPADE', 12), card('CLOVER', 1)], lastPlayPlayerIdx: 1 }));
    renderWithProviders(<PresidentPage />);
    // aria-label on the role="img" wrapper (getByLabelText matches aria-label, not the inner <img alt>).
    await waitFor(() => expect(screen.getByLabelText('♠ Q')).toBeInTheDocument());
    expect(screen.getByLabelText('♣ A')).toBeInTheDocument();
    expect(screen.getByLabelText('♠ Q')).toHaveAttribute('role', 'img');
  });

  it('enables Play button when a card is selected', async () => {
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    const playButton = screen.getByTestId('play-button');
    expect(playButton).toBeDisabled();

    fireEvent.click(screen.getByTestId('hand-card-0'));
    await waitFor(() => expect(screen.getByTestId('play-button')).not.toBeDisabled());
  });

  it('calls play on Play click', async () => {
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('play-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', [0]));
  });

  it('calls pass on Pass click', async () => {
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByTestId('pass-button')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('pass-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', []));
  });

  it('shows revolution banner when active', async () => {
    mockExec.mockResolvedValue(makeState({ revolutionActive: true }));
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByText(/革命中/)).toBeInTheDocument());
  });

  it('flashes a full-screen overlay when revolution turns on', async () => {
    // revolutionActive arrives true on mount → false→true transition fires the flash.
    mockExec.mockResolvedValue(makeState({ revolutionActive: true }));
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByTestId('president-revolution-flash')).toBeInTheDocument());
    expect(screen.getByTestId('president-revolution-flash')).toHaveClass('bg-ds-warning/40');
  });

  it('shows a distinct status banner when revolution reverts (true → false)', async () => {
    mockExec.mockReset();
    // Mount fetches an active-revolution state; the follow-up pass returns a
    // state where revolution has ended, driving the true → false transition.
    mockExec
      .mockResolvedValueOnce(makeState({ revolutionActive: true }))
      .mockResolvedValue(makeState({ revolutionActive: false }));
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByTestId('president-revolution-flash')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('pass-button'));
    await waitFor(() => expect(screen.getByTestId('president-revolution-end-flash')).toBeInTheDocument());
    const endFlash = screen.getByTestId('president-revolution-end-flash');
    expect(endFlash).toHaveAttribute('role', 'status');
    expect(endFlash).toHaveTextContent('革命終了');
    // Distinct from the activation flash (info-coloured banner vs warning scrim).
    expect(endFlash.querySelector('span')).toHaveClass('bg-ds-info');
  });

  it('disables action buttons when it is not human turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentTurn: 1 }));
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByTestId('pass-button')).toBeDisabled());
    expect(screen.getByTestId('play-button')).toBeDisabled();
  });

  it('resets with config and passes it to the API', async () => {
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        revolutionEnabled: true,
        cardExchangeEnabled: true,
        passFieldFlushEnabled: true,
        cpuDifficulty: 1,
      }),
    );
  });

  it('toggles revolutionEnabled setting', async () => {
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    const revCheckbox = screen.getByRole('checkbox', { name: /革命|Revolution/ });
    expect(revCheckbox).toBeChecked();
    fireEvent.click(revCheckbox);
    await waitFor(() => expect(revCheckbox).not.toBeChecked());
  });

  it('toggles cardExchangeEnabled setting', async () => {
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    const exchange = screen.getByRole('checkbox', { name: /カード交換|Card Exchange/ });
    fireEvent.click(exchange);
    await waitFor(() => expect(exchange).not.toBeChecked());
  });

  it('toggles passFieldFlushEnabled setting', async () => {
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    const flush = screen.getByRole('checkbox', { name: /パス即場流れ|Pass Flushes Field/ });
    fireEvent.click(flush);
    await waitFor(() => expect(flush).not.toBeChecked());
  });

  it('changes CPU difficulty', async () => {
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    const select = screen.getByLabelText(/CPU難易度|CPU Difficulty/);
    fireEvent.change(select, { target: { value: '2' } });
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, expect.objectContaining({ cpuDifficulty: 2 })),
    );
  });

  it('renders CPU difficulty options with localized (Japanese) labels', async () => {
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    const select = screen.getByLabelText(/CPU難易度|CPU Difficulty/) as HTMLSelectElement;
    const labels = Array.from(select.options).map((o) => o.textContent);
    expect(labels).toEqual(['やさしい', 'ふつう', 'むずかしい']);
  });

  it('renders game-end state with rank', async () => {
    mockExec.mockResolvedValue(
      makeState({
        gameEndFlag: true,
        players: [
          { id: 0, isHuman: true, isFinished: true, rank: 1, cardCount: 0, cards: [] },
          { id: 1, isHuman: false, isFinished: true, rank: 2, cardCount: 0, cards: [] },
          { id: 2, isHuman: false, isFinished: true, rank: 3, cardCount: 0, cards: [] },
          { id: 3, isHuman: false, isFinished: true, rank: 4, cardCount: 0, cards: [] },
        ],
      }),
    );
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    // Each finished player gets a uniquely-identified rank stamp badge.
    expect(screen.getByTestId('rank-stamp-1')).toBeInTheDocument();
    expect(screen.getByTestId('rank-stamp-2')).toBeInTheDocument();
    expect(screen.getByTestId('rank-stamp-3')).toBeInTheDocument();
    expect(screen.getByTestId('rank-stamp-4')).toBeInTheDocument();
    // Crown icon for rank 1 (president).
    expect(screen.getByTestId('rank-stamp-1').textContent).toContain('👑');
    // Trash icon for rank 4 (scum).
    expect(screen.getByTestId('rank-stamp-4').textContent).toContain('🗑️');
  });

  it('falls back to neutral styling and an empty icon for an unknown rank value', async () => {
    mockExec.mockResolvedValue(
      makeState({
        gameEndFlag: true,
        players: [
          // Out-of-range rank exercises the ?? fallbacks in PRESIDENT_RANK_BG / ICON / KEYS.
          { id: 0, isHuman: true, isFinished: true, rank: 99, cardCount: 0, cards: [] },
          { id: 1, isHuman: false, isFinished: true, rank: 2, cardCount: 0, cards: [] },
          { id: 2, isHuman: false, isFinished: true, rank: 3, cardCount: 0, cards: [] },
          { id: 3, isHuman: false, isFinished: true, rank: 4, cardCount: 0, cards: [] },
        ],
      }),
    );
    renderWithProviders(<PresidentPage />);
    const fallback = await screen.findByTestId('rank-stamp-99');
    // Fallback uses the neutral design tokens (no warning/info/error background).
    expect(fallback.className).toContain('bg-ds-surface');
    expect(fallback.className).toContain('text-ds-text-primary');
    // No matching icon for rank 99 → the emoji prefix is absent, leaving only a space + label text.
    expect(fallback.textContent?.trim().startsWith('👑')).toBe(false);
    expect(fallback.textContent?.trim().startsWith('🗑️')).toBe(false);
  });

  it('shows loading state when state has fewer than 4 players', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          { id: 0, isHuman: true, isFinished: false, rank: -1, cardCount: 0, cards: [] },
          { id: 1, isHuman: false, isFinished: false, rank: -1, cardCount: 0, cards: [] },
        ],
      }),
    );
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.queryByTestId('hand-card-0')).not.toBeInTheDocument());
  });

  it('sorts card indices when playing multiple cards', async () => {
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('hand-card-2'));
    fireEvent.click(screen.getByTestId('hand-card-1'));
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('play-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', [0, 1, 2]));
  });

  it('handles humanAction in state without crashing', async () => {
    mockExec.mockResolvedValue(
      makeState({
        humanAction: {
          playerIdx: 0,
          playedCards: [card('SPADE', 3)],
        },
      }),
    );
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
  });

  it('renders CPU actions panel when present', async () => {
    mockExec.mockResolvedValue(
      makeState({
        cpuActions: [
          { playerIdx: 1, playedCards: [card('HEART', 5)] },
          { playerIdx: 2, playedCards: null },
        ],
      }),
    );
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
  });

  it('runs the replay pipeline with human + cpu actions', async () => {
    // First call returns state with both human action and cpu actions (triggers replay)
    mockExec.mockResolvedValueOnce(makeState());
    mockExec.mockResolvedValueOnce(
      makeState({
        humanAction: { playerIdx: 0, playedCards: [card('SPADE', 3)] },
        cpuActions: [
          { playerIdx: 1, playedCards: [card('HEART', 5)] },
          { playerIdx: 2, playedCards: null },
          { playerIdx: 3, playedCards: [card('DIAMOND', 8)] },
        ],
      }),
    );
    renderWithProviders(<PresidentPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('play-button'));
    // Wait for the replay to complete and state to settle
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument(), { timeout: 4000 });
  });

  it('renders CLI terminal when CLI mode is enabled via localStorage', async () => {
    localStorage.setItem('cli-mode-president', 'true');
    renderWithProviders(<PresidentPage />);
    // CLI terminal surfaces a command input; presence of a textbox suggests CLI mode
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    localStorage.removeItem('cli-mode-president');
  });

  // **CUI と Daifugo は以前から交換内容を出していたのに、President の Web だけ
  // state.exchangeActions を描画していなかった (#4745)。**誰が誰に何を渡したかが
  // 分からないと、ラウンド開始時に手札が変わった理由を追えない。
  it('renders the round-start card exchange log', async () => {
    mockExec.mockResolvedValue(
      makeState({
        exchangeActions: [{ fromPlayerIdx: 1, toPlayerIdx: 0, cards: [{ design: 'SPADE', value: 1 }] }],
      }),
    );
    renderWithProviders(<PresidentPage />);

    const log = await screen.findByTestId('exchange-log');
    expect(log).toHaveTextContent('SPADE 1');
  });

  // 空のときは出さない。常時表示に退化すると、交換が無いラウンドでも見出しだけの
  // 空パネルが残る。
  it('renders no exchange log when there was no exchange', async () => {
    mockExec.mockResolvedValue(makeState({ exchangeActions: [] }));
    renderWithProviders(<PresidentPage />);

    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('exchange-log')).not.toBeInTheDocument();
  });
});
