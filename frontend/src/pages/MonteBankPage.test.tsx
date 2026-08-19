import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { montebankApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, MonteBankResponse } from '../types/card';
import { MONTE_BANK_RESULT } from '../types/games/montebank';
import { MonteBankPhase } from '../types/phases';
import { MonteBankPage } from './MonteBankPage';

vi.mock('../api/gameApi', () => ({
  montebankApi: { exec: vi.fn() },
  actionLogApi: { montebank: vi.fn() },
}));

vi.mock('../hooks/useCliMode', () => ({
  useCliMode: vi.fn(() => ({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  })),
}));

const mockApi = vi.mocked(montebankApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

const card = (design: string, value: number): Card => ({ design, value }) as Card;

const entry = (over: Partial<MonteBankResponse['layout'][number]> = {}) =>
  ({
    card: card('SPADE', 1),
    suitCount: 1,
    remainingOfSuit: 9,
    isEven: true,
    isPicked: false,
    ...over,
  }) as MonteBankResponse['layout'][number];

const base: MonteBankResponse = {
  phase: MonteBankPhase.BET,
  layout: [
    entry({ card: card('SPADE', 1), suitCount: 2, isEven: false }),
    entry({ card: card('SPADE', 7), suitCount: 2, isEven: false }),
    entry({ card: card('HEART', 3) }),
    entry({ card: card('CLOVER', 13) }),
  ],
  pick: -1,
  bet: 0,
  result: MONTE_BANK_RESULT.none,
  payout: 0,
  chips: 1000,
  roundNumber: 1,
  remainingCards: 36,
  gameEndFlag: false,
  payoutMultiplier: 3,
  message: '',
};

const withState = (over: Partial<MonteBankResponse>): MonteBankResponse => ({ ...base, ...over });

beforeEach(() => {
  vi.clearAllMocks();
  mockUseCliMode.mockReturnValue({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  } as unknown as ReturnType<typeof useCliMode>);
});

describe('MonteBankPage', () => {
  it('マウント時に reset を呼ぶ', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<MonteBankPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('場札4枚を出す', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<MonteBankPage />);
    await waitFor(() => expect(screen.getByTestId('mb-layout-0')).toBeInTheDocument());
    for (let i = 0; i < 4; i++) {
      expect(screen.getByTestId(`mb-layout-${i}`)).toBeInTheDocument();
    }
  });

  // **賭けの良し悪しはサーバの値をそのまま出す。** ページが数え直すと、
  // 控除率を決めている唯一の規則が 2 か所に分かれる。
  it('互角と不利をサーバの値どおりに書き分ける', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<MonteBankPage />);
    await waitFor(() => expect(screen.getByTestId('mb-note-0')).toBeInTheDocument());

    expect(screen.getByTestId('mb-note-0')).toHaveTextContent('不利');
    expect(screen.getByTestId('mb-note-1')).toHaveTextContent('不利');
    expect(screen.getByTestId('mb-note-2')).toHaveTextContent('互角');
    expect(screen.getByTestId('mb-note-3')).toHaveTextContent('互角');
  });

  // **フラグが値と食い違っていても、フラグに従う。** 一方向だけの検査は
  // 壊れているときに通ってしまうので、両側を踏む。
  it('isEven が false なら枚数に関わらず不利と出す', async () => {
    mockApi.mockResolvedValue(
      withState({ layout: [entry({ suitCount: 1, isEven: false }), entry(), entry(), entry()] }),
    );
    renderWithProviders(<MonteBankPage />);
    await waitFor(() => expect(screen.getByTestId('mb-note-0')).toBeInTheDocument());
    expect(screen.getByTestId('mb-note-0')).toHaveTextContent('不利');
    expect(screen.getByTestId('mb-note-1')).toHaveTextContent('互角');
  });

  it('スートの枚数を出す', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<MonteBankPage />);
    await waitFor(() => expect(screen.getByTestId('mb-count-0')).toHaveTextContent('2'));
  });

  // **賭けの良し悪しは場の枚数と山の残りの両方で決まる** (#5779)。
  it('山に残る同スート枚数も出す', async () => {
    mockApi.mockResolvedValue({
      ...base,
      layout: [entry({ suitCount: 2, remainingOfSuit: 7 }), entry({ card: card('HEART', 5), remainingOfSuit: 11 })],
    });
    renderWithProviders(<MonteBankPage />);

    await waitFor(() => expect(screen.getByTestId('mb-remaining-0')).toHaveTextContent('7'));
    expect(screen.getByTestId('mb-remaining-1')).toHaveTextContent('11');
    // 場の枚数の表示は残っている（受け入れ条件3）。
    expect(screen.getByTestId('mb-count-0')).toHaveTextContent('2');
  });

  // **選んだ位置は 0 始まりでそのまま送る。** 0 は正当な値。
  it('既定では場札0に賭ける', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<MonteBankPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '賭ける' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '賭ける' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', { idx: 0, bet: 50 }));
  });

  it('場札を選ぶとその位置を送る', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<MonteBankPage />);
    await waitFor(() => expect(screen.getByTestId('mb-layout-2')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('mb-layout-2'));
    expect(screen.getByTestId('mb-layout-2')).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByTestId('mb-layout-0')).toHaveAttribute('aria-pressed', 'false');

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '賭ける' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', { idx: 2, bet: 50 }));
  });

  it('賭ける前はゲートを出さない', async () => {
    mockApi.mockResolvedValue(base);
    renderWithProviders(<MonteBankPage />);
    await waitFor(() => expect(screen.getByTestId('mb-layout-0')).toBeInTheDocument());
    expect(screen.queryByTestId('mb-gate')).not.toBeInTheDocument();
  });

  it('決着でゲートと収支を出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: MonteBankPhase.RESULT,
        gate: card('HEART', 5),
        pick: 2,
        bet: 50,
        result: MONTE_BANK_RESULT.win,
        payout: 200,
      }),
    );
    renderWithProviders(<MonteBankPage />);
    await waitFor(() => expect(screen.getByTestId('mb-gate')).toBeInTheDocument());
    expect(screen.getByTestId('mb-result')).toHaveTextContent('的中');
    expect(screen.getByTestId('mb-result')).toHaveTextContent('150');
  });

  it('外れの収支は負で出す', async () => {
    mockApi.mockResolvedValue(
      withState({
        phase: MonteBankPhase.RESULT,
        gate: card('CLOVER', 5),
        pick: 2,
        bet: 50,
        result: MONTE_BANK_RESULT.lose,
      }),
    );
    renderWithProviders(<MonteBankPage />);
    await waitFor(() => expect(screen.getByTestId('mb-result')).toHaveTextContent('外れ'));
    expect(screen.getByTestId('mb-result')).toHaveTextContent('-50');
  });

  it('決着後は次のラウンドを送る', async () => {
    mockApi.mockResolvedValue(withState({ phase: MonteBankPhase.RESULT, result: MONTE_BANK_RESULT.lose, bet: 50 }));
    renderWithProviders(<MonteBankPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンドへ' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のラウンドへ' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('next'));
  });

  it('終局では次のラウンドを出さない', async () => {
    mockApi.mockResolvedValue(
      withState({ phase: MonteBankPhase.GAME_END, gameEndFlag: true, result: MONTE_BANK_RESULT.lose, bet: 50 }),
    );
    renderWithProviders(<MonteBankPage />);
    await waitFor(() => expect(screen.getByTestId('mb-round-line')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '次のラウンドへ' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '賭ける' })).not.toBeInTheDocument();
  });

  it('チップ・ラウンド・残り枚数・倍率を出す', async () => {
    mockApi.mockResolvedValue(withState({ chips: 850, roundNumber: 4, remainingCards: 20 }));
    renderWithProviders(<MonteBankPage />);
    await waitFor(() => expect(screen.getByTestId('mb-chips')).toHaveTextContent('850'));
    expect(screen.getByTestId('mb-round-line')).toHaveTextContent('4');
    expect(screen.getByTestId('mb-round-line')).toHaveTextContent('20');
    expect(screen.getByTestId('mb-round-line')).toHaveTextContent('3');
  });

  it('CLIモードでは端末を出す', async () => {
    mockUseCliMode.mockReturnValue({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    } as unknown as ReturnType<typeof useCliMode>);
    mockApi.mockResolvedValue(base);
    renderWithProviders(<MonteBankPage />);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '賭ける' })).not.toBeInTheDocument();
  });
});
