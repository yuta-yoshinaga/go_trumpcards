import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { DaifugoExchangeAction } from '../../types/card';
import { DaifugoExchangeLog } from './DaifugoExchangeLog';

const players = [
  { id: 0, isHuman: false },
  { id: 1, isHuman: true },
  { id: 2, isHuman: false },
  { id: 3, isHuman: false },
];

describe('DaifugoExchangeLog', () => {
  it('renders exchange title and entries', () => {
    const actions: DaifugoExchangeAction[] = [
      {
        fromPlayerIdx: 1,
        toPlayerIdx: 0,
        cards: [{ design: 'SPADE', value: 1 }],
      },
    ];
    render(<DaifugoExchangeLog players={players} actions={actions} />);
    expect(screen.getByText(/\[カード交換\]/)).toBeInTheDocument();
    expect(screen.getByText(/あなた → CPU 0: SPADE 1/)).toBeInTheDocument();
  });

  it('renders multiple exchange actions', () => {
    const actions: DaifugoExchangeAction[] = [
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
    render(<DaifugoExchangeLog players={players} actions={actions} />);
    expect(screen.getByText(/あなた → CPU 0: SPADE 1/)).toBeInTheDocument();
    expect(
      screen.getByText(/CPU 2 → CPU 3: HEART 13, DIAMOND 10/),
    ).toBeInTheDocument();
  });
});
