import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { cassinoApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CassinoResponse } from '../types/card';
import { CassinoPage } from './CassinoPage';

vi.mock('../api/gameApi', () => ({
  cassinoApi: { exec: vi.fn() },
  actionLogApi: { cassino: vi.fn() },
}));

const mockExec = vi.mocked(cassinoApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<CassinoResponse> = {}): CassinoResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 3,
        cards: [card('SPADE', 3), card('HEART', 5), card('DIAMOND', 7)],
        capturedCount: 0,
        sweepCount: 0,
        totalScore: 0,
      },
      { id: 1, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
      { id: 2, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
      { id: 3, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
    ],
    currentTurn: 0,
    tableCards: [card('SPADE', 2), card('HEART', 5)],
    builds: [],
    lastCaptureIdx: -1,
    gameEndFlag: false,
    phase: 'playerTurn',
    config: { targetScore: 21, multiBuildEnabled: true, sweepBonusEnabled: true, cpuDifficulty: 1 },
    cpuActions: [],
    humanAction: null,
    remainingDeck: 32,
    packsDealt: 1,
    roundWinners: [],
    lastRoundDetail: null,
    message: '',
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('CassinoPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders the human hand', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('hand-card-1')).toBeInTheDocument();
    expect(screen.getByTestId('hand-card-2')).toBeInTheDocument();
  });

  it('renders the human and CPU stat lines via i18n (no hardcoded 枚/pt)', async () => {
    const players = [
      { id: 0, isHuman: true, cardCount: 3, cards: [], capturedCount: 5, sweepCount: 2, totalScore: 7 },
      { id: 1, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 3 },
      { id: 2, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
      { id: 3, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
    ];
    mockExec.mockResolvedValue(makeState({ players }));
    renderWithProviders(<CassinoPage />);
    // ja label.humanStats: "手札3枚 / 捕獲5枚 / スイープ2 / 累計7pt"
    await waitFor(() => expect(screen.getByText(/手札3枚.*捕獲5枚.*スイープ2.*累計7pt/)).toBeInTheDocument());
    // ja label.cpuStats for CPU 1: "4枚 / 3pt"
    expect(screen.getByText(/4枚 \/ 3pt/)).toBeInTheDocument();
  });

  it('renders table cards', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('table-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('table-card-1')).toBeInTheDocument();
  });

  it('labels table cards with content, take-candidate, and selected state', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('table-card-0')).toBeInTheDocument());
    // Base label = card content only.
    expect(screen.getByTestId('table-card-0')).toHaveAttribute('aria-label', '♠ 2');
    // Selecting the ♥5 hand card makes the matching ♥5 table card a take candidate.
    fireEvent.click(screen.getByTestId('hand-card-1'));
    await waitFor(() => expect(screen.getByTestId('table-card-1')).toHaveAttribute('aria-label', '♥ 5 テイク候補'));
    // The non-matching ♠2 keeps its plain label.
    expect(screen.getByTestId('table-card-0')).toHaveAttribute('aria-label', '♠ 2');
    // Selecting the candidate flips its label to "selected".
    fireEvent.click(screen.getByTestId('table-card-1'));
    await waitFor(() => expect(screen.getByTestId('table-card-1')).toHaveAttribute('aria-label', '♥ 5 選択中'));
  });

  it('take button is disabled until both hand and table are selected', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('take-button')).toBeInTheDocument());
    expect(screen.getByTestId('take-button')).toBeDisabled();

    fireEvent.click(screen.getByTestId('hand-card-0'));
    expect(screen.getByTestId('take-button')).toBeDisabled();

    fireEvent.click(screen.getByTestId('table-card-0'));
    await waitFor(() => expect(screen.getByTestId('take-button')).not.toBeDisabled());
  });

  it('calls take when Take is clicked with selections', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('table-card-0'));
    fireEvent.click(screen.getByTestId('take-button'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('take', {
        handIndex: 0,
        tableIndices: [0],
        buildIndices: [],
      }),
    );
  });

  it('calls trail when Trail is clicked with a hand selection', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('trail-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trail', { handIndex: 0 }));
  });

  it('calls build when Build is clicked with selection + declared value', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('table-card-0'));
    fireEvent.change(screen.getByTestId('build-value-select'), { target: { value: '5' } });
    fireEvent.click(screen.getByTestId('build-button'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('build', {
        handIndex: 0,
        tableIndices: [0],
        declaredValue: 5,
      }),
    );
  });

  it('disables actions when it is not human turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentTurn: 1 }));
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('take-button')).toBeDisabled());
    expect(screen.getByTestId('trail-button')).toBeDisabled();
  });

  it('resets with config and passes it to the API', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: {
          targetScore: 21,
          multiBuildEnabled: true,
          sweepBonusEnabled: true,
          cpuDifficulty: 1,
        },
      }),
    );
  });

  it('toggles multi-build setting', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    const multi = screen.getByRole('checkbox', { name: /複合ビルド|Multi-Builds/ });
    expect(multi).toBeChecked();
    fireEvent.click(multi);
    await waitFor(() => expect(multi).not.toBeChecked());
  });

  it('changes CPU difficulty', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    const select = screen.getByLabelText(/CPU難易度|CPU Difficulty/);
    fireEvent.change(select, { target: { value: '2' } });
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'reset',
        expect.objectContaining({ config: expect.objectContaining({ cpuDifficulty: 2 }) }),
      ),
    );
  });

  it('shows loading state when state has fewer than 4 players', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [{ id: 0, isHuman: true, cardCount: 0, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 }],
      }),
    );
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.queryByTestId('hand-card-0')).not.toBeInTheDocument());
  });

  it('renders builds area', async () => {
    mockExec.mockResolvedValue(
      makeState({
        builds: [{ ownerIdx: 1, value: 8, groups: [[card('SPADE', 3), card('HEART', 5)]], isMulti: false }],
      }),
    );
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('build-0')).toBeInTheDocument());
  });

  it('renders the build label fully localized (no mixed English literals)', async () => {
    mockExec.mockResolvedValue(
      makeState({
        builds: [
          { ownerIdx: 0, value: 7, groups: [[card('SPADE', 3), card('HEART', 4)]], isMulti: true },
          { ownerIdx: 2, value: 5, groups: [[card('CLUB', 5)]], isMulti: false },
        ],
      }),
    );
    renderWithProviders(<CassinoPage />);
    const b0 = await screen.findByTestId('build-0');
    const b1 = await screen.findByTestId('build-1');
    // No raw English literals from the old hardcoded string.
    for (const el of [b0, b1]) {
      expect(el.textContent).not.toMatch(/multi|single|owner:/i);
    }
    // Multi build owned by the human, single build owned by CPU 2.
    expect(b0.textContent).toContain('複合');
    expect(b0.textContent).toContain('あなた');
    expect(b1.textContent).toContain('単一');
    expect(b1.textContent).toContain('CPU 2');
  });

  it('toggles a build selection and includes it in take', async () => {
    mockExec.mockResolvedValue(
      makeState({
        builds: [{ ownerIdx: 1, value: 8, groups: [[card('SPADE', 3), card('HEART', 5)]], isMulti: false }],
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 1,
            cards: [card('HEART', 8)],
            capturedCount: 0,
            sweepCount: 0,
            totalScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
          { id: 2, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
          { id: 3, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
        ],
        tableCards: [],
      }),
    );
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('build-0')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('build-0'));
    fireEvent.click(screen.getByTestId('take-button'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('take', {
        handIndex: 0,
        tableIndices: [],
        buildIndices: [0],
      }),
    );
  });

  it('toggles sweepBonusEnabled setting', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    const sweep = screen.getByRole('checkbox', { name: /スイープボーナス|Sweep Bonus/ });
    expect(sweep).toBeChecked();
    fireEvent.click(sweep);
    await waitFor(() => expect(sweep).not.toBeChecked());
  });

  it('shows a Take suggestion when hand value equals table sum', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 2,
            cards: [card('SPADE', 9), card('CLOVER', 2)],
            capturedCount: 0,
            sweepCount: 0,
            totalScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
          { id: 2, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
          { id: 3, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
        ],
        tableCards: [card('SPADE', 4), card('HEART', 5)],
      }),
    );
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('table-card-0'));
    fireEvent.click(screen.getByTestId('table-card-1'));
    expect(screen.getByTestId('cs-suggest-button')).toHaveTextContent('取る (9)');

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('cs-suggest-button'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('take', {
        handIndex: 0,
        tableIndices: [0, 1],
        buildIndices: [],
      }),
    );
  });

  it('shows a Build suggestion and dispatches build with the declared value', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 2,
            cards: [card('SPADE', 4), card('CLOVER', 9)],
            capturedCount: 0,
            sweepCount: 0,
            totalScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
          { id: 2, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
          { id: 3, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
        ],
        tableCards: [card('SPADE', 5)],
      }),
    );
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('table-card-0'));
    expect(screen.getByTestId('cs-suggest-button')).toHaveTextContent('ビルド (9)');

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('cs-suggest-button'));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('build', {
        handIndex: 0,
        tableIndices: [0],
        declaredValue: 9,
      }),
    );
  });

  it('does not render a suggestion when there is no inferred action', async () => {
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('hand-card-0'));
    expect(screen.queryByTestId('cs-suggest-button')).not.toBeInTheDocument();
  });

  it('shows the advisory hint tooltip when hints are enabled and no selection-based suggestion is active', async () => {
    localStorage.setItem('hint_enabled_cassino', 'true');
    renderWithProviders(<CassinoPage />);
    // No hand card selected → no concrete suggestion → the single advisory hint shows.
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
    expect(screen.queryByTestId('cs-suggest-button')).not.toBeInTheDocument();
    localStorage.removeItem('hint_enabled_cassino');
  });

  it('consolidates guidance to a single source: the suggestion supersedes the advisory hint tooltip', async () => {
    localStorage.setItem('hint_enabled_cassino', 'true');
    mockExec.mockResolvedValue(
      makeState({
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 2,
            cards: [card('SPADE', 9), card('CLOVER', 2)],
            capturedCount: 0,
            sweepCount: 0,
            totalScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
          { id: 2, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
          { id: 3, isHuman: false, cardCount: 4, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
        ],
        tableCards: [card('SPADE', 4), card('HEART', 5)],
      }),
    );
    renderWithProviders(<CassinoPage />);
    // With hints enabled and nothing selected, the advisory tooltip is the sole hint.
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());

    // Selecting cards that sum to the played card produces a concrete take suggestion.
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('table-card-0'));
    fireEvent.click(screen.getByTestId('table-card-1'));

    // The actionable suggestion now supersedes the advisory tooltip: exactly one hint shows.
    await waitFor(() => expect(screen.getByTestId('cs-suggest-button')).toBeInTheDocument());
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();
    localStorage.removeItem('hint_enabled_cassino');
  });

  it('renders CLI terminal when CLI mode is enabled via localStorage', async () => {
    localStorage.setItem('cli-mode-cassino', 'true');
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    localStorage.removeItem('cli-mode-cassino');
  });
});

// #5549: CUI は「誰が何のカードでテイク(何枚・スイープか)/ビルド/トレイルしたか」を
// 毎ターン出しているのに、Web はどちらのフィールドも読んでいなかった。
describe('CassinoPage action history', () => {
  const log = () => screen.getByTestId('cs-action-log');
  const card = (design: string, value: number) => ({ design, value }) as never;

  it('shows each action with its kind, and the capture count on a take', async () => {
    mockExec.mockResolvedValue(
      makeState({
        humanAction: {
          playerIdx: 0,
          type: 'take',
          playedCard: card('SPADE', 7),
          capturedCards: [card('HEART', 3), card('CLOVER', 4)],
          buildValue: 0,
          isSweep: true,
        },
        cpuActions: [
          {
            playerIdx: 1,
            type: 'build',
            playedCard: card('DIAMOND', 5),
            capturedCards: [],
            buildValue: 9,
            isSweep: false,
          },
          {
            playerIdx: 2,
            type: 'trail',
            playedCard: card('HEART', 2),
            capturedCards: [],
            buildValue: 0,
            isSweep: false,
          },
        ],
      }),
    );
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(log()).toBeInTheDocument());

    expect(log()).toHaveTextContent('あなた');
    expect(log()).toHaveTextContent('CPU 1');
    expect(log()).toHaveTextContent('CPU 2');
    // テイクは捕獲枚数とスイープを出す。
    expect(log()).toHaveTextContent('2');
    expect(log()).toHaveTextContent('スイープ');
    // ビルドは値、トレイルはその旨。
    expect(log()).toHaveTextContent('9');
    expect(log()).toHaveTextContent('場に置');
  });

  // **何も起きていないうちは出さない。**
  it('renders nothing before anyone has acted', async () => {
    mockExec.mockResolvedValue(makeState({ cpuActions: [], humanAction: null }));
    renderWithProviders(<CassinoPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(screen.queryByTestId('cs-action-log')).not.toBeInTheDocument();
  });
});
