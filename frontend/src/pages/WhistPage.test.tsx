import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { whistApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, WhistResponse } from '../types/card';
import { WhistPage } from './WhistPage';

vi.mock('../api/gameApi', () => ({
  whistApi: { exec: vi.fn() },
  actionLogApi: { whist: vi.fn() },
}));

const mockExec = vi.mocked(whistApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<WhistResponse> = {}): WhistResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 3,
        cards: [card('SPADE', 1), card('HEART', 5), card('DIAMOND', 9)],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        team: 0,
      },
      { id: 1, isHuman: false, cardCount: 3, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0, team: 1 },
      { id: 2, isHuman: false, cardCount: 3, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0, team: 0 },
      { id: 3, isHuman: false, cardCount: 3, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0, team: 1 },
    ],
    phase: 0,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    currentTrick: [],
    trumpSuit: 0,
    dealerIdx: 0,
    teamScores: [0, 0],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 1, pointLimit: 5 },
    ...overrides,
  };
}

const playState = makeState();
const gameEndState = makeState({ phase: 3, gameEndFlag: true, winnerTeam: 0 });

beforeEach(() => {
  mockExec.mockResolvedValue(playState);
});

describe('WhistPage', () => {
  it('calls reset on mount with default config', async () => {
    renderWithProviders(<WhistPage />);
    // useTrickGameBase fires the mount reset with four positional args.
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 1, pointLimit: 5 }),
    );
  });

  // The phase key map must hold bare keys; usePhaseNames adds the `phase.`
  // prefix itself, so a prefixed key resolved to the literal
  // "phase.phase.play" on screen. See issue #4374.
  it('renders the translated phase name, not the raw i18n key', async () => {
    renderWithProviders(<WhistPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('プレイ'));
    expect(screen.getByTestId('phase-indicator')).not.toHaveTextContent('phase.');
  });

  it('lists the keyboard shortcuts in a collapsible panel', async () => {
    renderWithProviders(<WhistPage />);
    const panel = await screen.findByTestId('wh-kbd-shortcuts');
    // Closed by default so it stays discreet.
    expect(panel).not.toHaveAttribute('open');
    expect(screen.getByText('キーボードショートカット')).toBeInTheDocument();
    // While collapsed the rows are not mounted, so they add no text to the page
    // — the shortcut labels name the same actions as the buttons around them.
    // See KeyboardShortcutsPanel and issue #4369.
    expect(screen.queryByText('次のトリック / ラウンドへ進む')).not.toBeInTheDocument();

    fireEvent.click(screen.getByText('キーボードショートカット'));
    // The 'n' advance shortcut and card-selection keys are advertised.
    expect(screen.getByText('次のトリック / ラウンドへ進む')).toBeInTheDocument();
    expect(screen.getByText('数字キーで手札のカードを選択')).toBeInTheDocument();
  });

  it('shows a known hint reason translated', async () => {
    renderWithProviders(<WhistPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockResolvedValueOnce(makeState({ hint: { cardIndex: 0, reason: 'trump_cut' } }));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(/トランプでカット/)).toBeInTheDocument());
  });

  it('names the recommended card (suit + rank) in the hint text, not just its index', async () => {
    renderWithProviders(<WhistPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    // cardIndex 0 in the human hand is ♠ A.
    mockExec.mockResolvedValueOnce(makeState({ hint: { cardIndex: 0, reason: 'trump_cut' } }));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(/♠ A/)).toBeInTheDocument());
  });

  it('falls back to generic text for an unknown hint reason', async () => {
    renderWithProviders(<WhistPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockResolvedValueOnce(makeState({ hint: { cardIndex: 1, reason: 'brand_new_reason' } }));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    // Unknown reason -> hintReason.fallback, not the raw "brand_new_reason" key.
    await waitFor(() => expect(screen.getByText(/最善手/)).toBeInTheDocument());
    expect(screen.queryByText(/brand_new_reason/)).not.toBeInTheDocument();
  });

  it('advances to the next trick when pressing n at trick end', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1 }));
    renderWithProviders(<WhistPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'n' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('advances to the next round when pressing n at round end', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<WhistPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'n' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('shows mid-game リセット button that opens confirm dialog', async () => {
    renderWithProviders(<WhistPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('executes reset after confirm dialog is accepted', async () => {
    renderWithProviders(<WhistPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1, pointLimit: 5 }));
  });

  it('shows 次のゲーム and fires reset directly at game end (no confirm)', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<WhistPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, expect.any(Object)));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('color-codes the human player team badge (team 0 → info)', async () => {
    renderWithProviders(<WhistPage />);
    const humanTeam = await screen.findByTestId('whist-human-team');
    expect(humanTeam.querySelector('span')).toHaveClass('text-ds-info');
  });

  it('color-codes the score-table team rows (team 0 info, team 1 error)', async () => {
    renderWithProviders(<WhistPage />);
    await waitFor(() => expect(screen.getAllByText('チーム 0').length).toBeGreaterThan(0));
    // Score-table cells render the team label inside a colored chip span.
    const team0Chips = screen.getAllByText('チーム 0').filter((el) => el.className.includes('text-ds-info'));
    const team1Chips = screen.getAllByText('チーム 1').filter((el) => el.className.includes('text-ds-error'));
    expect(team0Chips.length).toBeGreaterThan(0);
    expect(team1Chips.length).toBeGreaterThan(0);
  });
});
