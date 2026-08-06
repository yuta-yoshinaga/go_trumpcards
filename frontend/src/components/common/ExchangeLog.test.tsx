import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { ExchangeLogAction } from './ExchangeLog';
import { ExchangeLog } from './ExchangeLog';

const players = [
  { id: 0, isHuman: false },
  { id: 1, isHuman: true },
  { id: 2, isHuman: false },
  { id: 3, isHuman: false },
];

describe('ExchangeLog', () => {
  it('renders exchange title and entries', () => {
    const actions: ExchangeLogAction[] = [
      {
        fromPlayerIdx: 1,
        toPlayerIdx: 0,
        cards: [{ design: 'SPADE', value: 1 }],
      },
    ];
    render(<ExchangeLog ns="daifugo" players={players} actions={actions} />);
    expect(screen.getByText(/\[カード交換\]/)).toBeInTheDocument();
    expect(screen.getByText(/あなた → CPU 0: SPADE 1/)).toBeInTheDocument();
  });

  it('renders multiple exchange actions', () => {
    const actions: ExchangeLogAction[] = [
      {
        fromPlayerIdx: 1,
        toPlayerIdx: 0,
        cards: [{ design: 'SPADE', value: 1 }],
      },
      {
        fromPlayerIdx: 2,
        toPlayerIdx: 3,
        cards: [
          { design: 'HEART', value: 13 },
          { design: 'DIAMOND', value: 10 },
        ],
      },
    ];
    render(<ExchangeLog ns="daifugo" players={players} actions={actions} />);
    expect(screen.getByText(/あなた → CPU 0: SPADE 1/)).toBeInTheDocument();
    expect(screen.getByText(/CPU 2 → CPU 3: HEART 13, DIAMOND 10/)).toBeInTheDocument();
  });

  // **ns が本当に使われていること。**daifugo と president は文言が同一なので、
  // 「president を渡して同じ文字列が出る」ことを確かめても、ns を無視して
  // daifugo 固定に退化した実装で通ってしまう (実際に試して通ることを確認済み)。
  // 翻訳を持たない名前空間を渡し、i18next がキーをそのまま返すことで踏む。
  it('resolves the copy from the namespace it was given', () => {
    const actions: ExchangeLogAction[] = [{ fromPlayerIdx: 1, toPlayerIdx: 0, cards: [{ design: 'SPADE', value: 1 }] }];
    render(<ExchangeLog ns="__no_such_namespace__" players={players} actions={actions} />);
    // 訳が無ければ i18next はキー文字列を返す。daifugo 固定ならここは
    // 「[カード交換]」になり、この期待は外れる。
    expect(screen.getByTestId('exchange-log')).toHaveTextContent('exchange.title');
  });
});
