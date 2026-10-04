import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import i18n from '../i18n';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { FiftyOneResponse } from '../types/card';

vi.mock('../hooks/useCliMode', async () => {
  const { useState } = await import('react');
  return {
    useCliMode: () => {
      const [cliEnabled, setCliEnabled] = useState(false);
      const [logEntries, setLogEntries] = useState<{ type: 'input' | 'output' | 'error'; text: string; id: number }[]>(
        [],
      );
      const add = (type: 'input' | 'output' | 'error') => (text: string) =>
        setLogEntries((entries) => [...entries, { type, text, id: entries.length }]);
      return {
        cliEnabled,
        toggleCli: () => setCliEnabled((enabled) => !enabled),
        logEntries,
        addInput: add('input'),
        addOutput: add('output'),
        addError: add('error'),
        clearLog: () => setLogEntries([]),
      };
    },
  };
});

const mockExec = vi.fn();
vi.mock('../api/gameApi', () => ({
  fiftyoneApi: { exec: (...args: unknown[]) => mockExec(...args) },
}));

const mockPlaySound = vi.fn();
const mockSoundValue = { playSound: mockPlaySound, muted: false, toggleMute: vi.fn() };
vi.mock('../providers/SoundProvider', () => ({
  SoundProvider: ({ children }: { children: React.ReactNode }) => children,
  useSound: () => mockSoundValue,
  useOptionalSound: () => mockSoundValue,
}));

const baseState: FiftyOneResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 5,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'SPADE', value: 10 },
        { design: 'HEART', value: 5 },
        { design: 'DIAMOND', value: 3 },
        { design: 'CLOVER', value: 2 },
      ] as never[],
      score: 21,
    },
    { id: 1, isHuman: false, cardCount: 5, cards: [], score: 0 },
    { id: 2, isHuman: false, cardCount: 5, cards: [], score: 0 },
    { id: 3, isHuman: false, cardCount: 5, cards: [], score: 0 },
  ],
  tableCards: [
    { design: 'SPADE', value: 13 },
    { design: 'HEART', value: 9 },
    { design: 'DIAMOND', value: 12 },
    { design: 'CLOVER', value: 6 },
    { design: 'SPADE', value: 8 },
  ] as never[],
  phase: 0,
  currentTurn: 0,
  gameEndFlag: false,
  winnerIdx: -1,
  turnNumber: 1,
  stopCallerIdx: -1,
  lastAction: '',
  lastHandIdx: -1,
  lastTableIdx: -1,
  message: '',
  config: { cpuDifficulty: 1 },
};

const gameEndState: FiftyOneResponse = {
  ...baseState,
  phase: 1,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'You win!',
  messageCode: 'fiftyone.result.humanWin',
};

describe('FiftyOnePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  it('labels hand and table cards and toggles aria-pressed on selection', async () => {
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    // Hand ♠A and table ♠K are distinct, labeled, selectable buttons.
    const handAce = await screen.findByRole('button', { name: '♠ A、手札の位置0' });
    const tableKing = screen.getByRole('button', { name: '♠ K、場札の位置0' });
    expect(handAce).toHaveAttribute('aria-pressed', 'false');
    expect(tableKing).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(handAce);
    expect(handAce).toHaveAttribute('aria-pressed', 'true');
    fireEvent.click(tableKing);
    expect(tableKing).toHaveAttribute('aria-pressed', 'true');

    // Clicking the same hand card again deselects it.
    fireEvent.click(handAce);
    expect(handAce).toHaveAttribute('aria-pressed', 'false');
  });

  it('calls reset on mount', async () => {
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders player score after load', async () => {
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(screen.getByText(/スコア: 21/)).toBeInTheDocument());
  });

  it('renders CPU difficulty options with localized labels', async () => {
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // Difficulty options are localized (ja), not the hardcoded Easy/Normal/Hard.
    expect(screen.getByRole('option', { name: '簡単' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: '普通' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: '難しい' })).toBeInTheDocument();
  });

  it('uses the selected CPU difficulty on the next reset', async () => {
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.click(screen.getByText('設定'));
    fireEvent.change(screen.getAllByRole('combobox')[0], { target: { value: '2' } });
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 2 } }));
  });

  it('updates CPU difficulty from CLI and rejects invalid values', async () => {
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.click(screen.getByRole('button', { name: 'CLIモードに切り替え' }));
    const prompt = screen.getByRole('textbox', { name: 'コマンドを入力...' });
    fireEvent.change(prompt, { target: { value: 'sd 2' } });
    fireEvent.keyDown(prompt, { key: 'Enter' });
    expect(await screen.findByRole('log')).toHaveTextContent(/難しいに設定しました/);
    fireEvent.change(prompt, { target: { value: 'sd 3' } });
    fireEvent.keyDown(prompt, { key: 'Enter' });
    expect(await screen.findByRole('log')).toHaveTextContent(/使い方: sd <0-2>/);
    fireEvent.change(prompt, { target: { value: 'unknown' } });
    fireEvent.keyDown(prompt, { key: 'Enter' });
    expect(await screen.findByRole('log')).toHaveTextContent('Unknown command: unknown');
    fireEvent.change(prompt, { target: { value: 'help' } });
    fireEvent.keyDown(prompt, { key: 'Enter' });
    expect(await screen.findByRole('log')).toHaveTextContent(/sd <0-2> - CPU難易度を設定/);
    fireEvent.click(screen.getByRole('button', { name: 'GUIモードに切り替え' }));
    fireEvent.click(screen.getByText('設定'));
    expect(screen.getAllByRole('combobox')[0]).toHaveValue('2');
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 2 } }));
  });

  it('explains exchange choices and shows the updated score after exchange', async () => {
    const updatedState: FiftyOneResponse = {
      ...baseState,
      players: baseState.players.map((player) =>
        player.isHuman
          ? {
              ...player,
              score: 28,
              cards: [
                { design: 'HEART', value: 13 },
                { design: 'HEART', value: 10 },
                { design: 'SPADE', value: 1 },
                { design: 'DIAMOND', value: 3 },
                { design: 'CLOVER', value: 2 },
              ] as never[],
            }
          : player,
      ),
    };
    mockExec.mockImplementation((command: string) => Promise.resolve(command === 'reset' ? baseState : updatedState));

    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    expect(screen.getByTestId('exchange-guidance')).toHaveTextContent(
      '個別交換: 手札と場札を1枚ずつ選択。全交換: 手札5枚を交換。交換後は新しい手札とスート別得点を確認して、次のカードを選び直してください。',
    );

    fireEvent.click(screen.getByTestId('exchange-all-button'));
    await waitFor(() => expect(screen.getByText('あなた — スコア: 28')).toBeInTheDocument());
    expect(screen.getByTestId('suit-score-badges')).toBeInTheDocument();
    expect(screen.getByTestId('suit-badge-HEART')).toHaveAccessibleName('♥、20/51、最高得点');
    expect(screen.getByTestId('suit-badge-SPADE')).toHaveAccessibleName('♠、11/51');
  });

  it('exchange all button calls exchangeall', async () => {
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const exchangeAllBtn = screen.getByTestId('exchange-all-button');
    fireEvent.click(exchangeAllBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('exchangeall'));
  });

  it('stop button calls stop', async () => {
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const stopBtn = screen.getByTestId('stop-button');
    fireEvent.click(stopBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('stop'));
  });

  it('disables buttons on game end', async () => {
    mockExec.mockResolvedValue(gameEndState);
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(screen.getByTestId('exchange-all-button')).toBeDisabled());
    expect(screen.getByTestId('stop-button')).toBeDisabled();
  });

  it('exchange button is disabled until both hand and table cards are selected', async () => {
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const exchangeBtn = screen.getByTestId('exchange-button');
    expect(exchangeBtn).toBeDisabled();
  });

  it('guides the next selection while the exchange button is disabled', async () => {
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    // Nothing selected → prompt for a hand card.
    expect(screen.getByText('手札を選択してください')).toBeInTheDocument();

    // Hand selected → prompt for a table card.
    fireEvent.click(screen.getByRole('button', { name: '♠ A、手札の位置0' }));
    expect(screen.getByText('場札を選択してください')).toBeInTheDocument();
    expect(screen.queryByText('手札を選択してください')).not.toBeInTheDocument();

    // Both selected → guide disappears, button enabled.
    fireEvent.click(screen.getByRole('button', { name: '♠ K、場札の位置0' }));
    expect(screen.queryByText(/を選択してください/)).not.toBeInTheDocument();
    expect(screen.getByTestId('exchange-button')).not.toBeDisabled();
  });

  it('hides the selection guide when it is not the human turn', async () => {
    mockExec.mockResolvedValue({ ...baseState, currentTurn: 1 });
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByText(/を選択してください/)).not.toBeInTheDocument();
  });

  it('plays a card-place sound when exchanging all cards', async () => {
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    fireEvent.click(screen.getByTestId('exchange-all-button'));
    // The central tap plays after the exec resolves, so await it.
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('cardPlace'));
  });

  it('plays a card-place sound when exchanging a single selected card', async () => {
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    fireEvent.click(screen.getByRole('button', { name: '♠ A、手札の位置0' }));
    fireEvent.click(screen.getByRole('button', { name: '♠ K、場札の位置0' }));
    mockPlaySound.mockClear();
    fireEvent.click(screen.getByTestId('exchange-button'));
    // The central tap plays after the exec resolves, so await it.
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('cardPlace'));
  });

  it('plays a chip-click sound when calling stop', async () => {
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    fireEvent.click(screen.getByTestId('stop-button'));
    expect(mockPlaySound).toHaveBeenCalledWith('chipClick');
  });

  it('plays an error buzz when the api call fails', async () => {
    mockExec.mockRejectedValueOnce(new Error('boom'));
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('errorBuzz'));
  });

  it('renders suit score badges with score/51 format and highlights the leading suit', async () => {
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(screen.getByTestId('suit-score-badges')).toBeInTheDocument());
    // SPADE=11+10=21, CLOVER=2, HEART=5, DIAMOND=3 → SPADE leads.
    const spade = screen.getByTestId('suit-badge-SPADE');
    expect(spade).toHaveTextContent('21/51');
    expect(spade.className).toContain('bg-ds-accent');
    expect(spade).toHaveAccessibleName('♠、21/51、最高得点');
    const heart = screen.getByTestId('suit-badge-HEART');
    expect(heart).toHaveTextContent('5/51');
    expect(heart.className).not.toContain('bg-ds-accent');
    expect(heart).toHaveAccessibleName('♥、5/51');

    // i18n キー名が生で画面に出ていないこと (i18n が解決している証拠)
    expect(screen.queryByText(/suitBadge/)).not.toBeInTheDocument();
  });

  it('announces every suit tied for the highest score', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      players: [
        {
          ...baseState.players[0],
          cards: [
            { design: 'SPADE', value: 7 },
            { design: 'CLOVER', value: 7 },
            { design: 'HEART', value: 2 },
            { design: 'DIAMOND', value: 3 },
          ],
        },
        ...baseState.players.slice(1),
      ],
    });
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(screen.getByTestId('suit-score-badges')).toBeInTheDocument());
    expect(screen.getByTestId('suit-badge-SPADE')).toHaveAccessibleName('♠、7/51、最高得点');
    expect(screen.getByTestId('suit-badge-CLOVER')).toHaveAccessibleName('♣、7/51、最高得点');
    expect(screen.getByTestId('suit-badge-HEART')).toHaveAccessibleName('♥、2/51');
  });

  it('uses English punctuation in suit badge accessible names', async () => {
    const previousLanguage = i18n.language;
    try {
      await i18n.changeLanguage('en');
      mockExec.mockResolvedValue(baseState);
      const { FiftyOnePage } = await import('./FiftyOnePage');
      renderWithProviders(<FiftyOnePage />);
      await waitFor(() => expect(screen.getByTestId('suit-score-badges')).toBeInTheDocument());
      const ariaLabel = screen.getByTestId('suit-badge-SPADE').getAttribute('aria-label');
      expect(ariaLabel).toBe('♠, 21/51, highest score');
      expect(ariaLabel).not.toContain('、');
    } finally {
      await i18n.changeLanguage(previousLanguage);
    }
  });

  it('renders suit score badges with custom hand and verifies score/51 format', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      players: [
        {
          ...baseState.players[0],
          cards: [
            { design: 'CLOVER', value: 1 },
            { design: 'CLOVER', value: 13 },
            { design: 'CLOVER', value: 12 },
            { design: 'SPADE', value: 2 },
            { design: 'HEART', value: 3 },
          ],
        },
        ...baseState.players.slice(1),
      ],
    });
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(screen.getByTestId('suit-score-badges')).toBeInTheDocument());
    const clover = screen.getByTestId('suit-badge-CLOVER');
    expect(clover).toHaveTextContent('31/51');
    const spade = screen.getByTestId('suit-badge-SPADE');
    expect(spade).toHaveTextContent('2/51');
    expect(screen.queryByText(/suitBadge/)).not.toBeInTheDocument();
  });
});

// #5532: CUI は「誰が宣言したか」と「残り1巡で終わる」を出しているのに、
// Web は「ストップ宣言済み」の一文だけで、どちらも分からなかった。
describe('FiftyOnePage stop indicator', () => {
  const stopBanner = () => screen.getByTestId('fo-stop-called');

  it('names the CPU that called stop and says it is the last round', async () => {
    mockExec.mockResolvedValue({ ...baseState, stopCallerIdx: 2 });
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(stopBanner()).toBeInTheDocument());
    expect(stopBanner()).toHaveTextContent('CPU 2');
    expect(stopBanner()).toHaveTextContent('最終');
  });

  it('names the human when the human called it', async () => {
    mockExec.mockResolvedValue({ ...baseState, stopCallerIdx: 0 });
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(stopBanner()).toBeInTheDocument());
    expect(stopBanner()).toHaveTextContent('あなた');
    // **CPU 0 と読ませない。**席0は人間。
    expect(stopBanner()).not.toHaveTextContent('CPU 0');
  });

  it('shows nothing before anyone calls stop', async () => {
    mockExec.mockResolvedValue(baseState);
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(screen.queryByTestId('fo-stop-called')).not.toBeInTheDocument();
  });
});

describe('FiftyOnePage keyboard shortcuts', () => {
  it('pressing p dispatches exchange when a hand and table card are selected', async () => {
    mockExec.mockResolvedValue(baseState);
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    // Select hand card index 0 and table card index 0
    fireEvent.click(screen.getByRole('button', { name: '♠ A、手札の位置0' }));
    fireEvent.click(screen.getByRole('button', { name: '♠ K、場札の位置0' }));

    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'p' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { handIdx: 0, tableIdx: 0 }));
  });

  it('pressing a dispatches exchangeall', async () => {
    mockExec.mockResolvedValue(baseState);
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'a' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('exchangeall'));
  });

  it('pressing s dispatches stop', async () => {
    mockExec.mockResolvedValue(baseState);
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 's' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('stop'));
  });

  it('pressing r dispatches reset when game is ended', async () => {
    mockExec.mockResolvedValue(gameEndState);
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'r' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 1 } }));
  });

  it('ignores p key when cards are not selected', async () => {
    mockExec.mockResolvedValue(baseState);
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'p' });
    // Need to flush pending dispatches, wait a tick.
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores s key when stop has already been called', async () => {
    mockExec.mockResolvedValue({ ...baseState, stopCallerIdx: 2 });
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 's' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('renders ActionShortcutsPanel with translated labels', async () => {
    mockExec.mockResolvedValue(baseState);
    const { FiftyOnePage } = await import('./FiftyOnePage');
    renderWithProviders(<FiftyOnePage />);
    await waitFor(() => expect(screen.getByTestId('fifty-one-kbd-shortcuts')).toBeInTheDocument());
    const panel = screen.getByTestId('fifty-one-kbd-shortcuts');
    // Open the panel so its contents are mounted
    fireEvent.click(screen.getByText('キーボードショートカット'));

    // Ensure the translated keys from common.json appear.
    // 'a' -> exchangeAll
    expect(panel).toHaveTextContent('全交換する');
    // 's' -> stop
    expect(panel).toHaveTextContent('ストップをかける');

    // Select hand card index 0 and table card index 0 to enable exchange
    fireEvent.click(screen.getByRole('button', { name: '♠ A、手札の位置0' }));
    fireEvent.click(screen.getByRole('button', { name: '♠ K、場札の位置0' }));

    // 'p' -> exchange
    expect(panel).toHaveTextContent('カードを交換する');
  });
});
