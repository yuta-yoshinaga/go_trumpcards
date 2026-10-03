import { renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import i18n from '../i18n';
import type { Card } from '../types/common';
import { useCommunityCardAnnouncement } from './useCommunityCardAnnouncement';

const cards: Card[] = [
  { design: 'SPADE', value: 10 },
  { design: 'HEART', value: 5 },
  { design: 'DIAMOND', value: 8 },
  { design: 'CLOVER', value: 2 },
  { design: 'SPADE', value: 4 },
];

const t = i18n.getFixedT('ja', 'holdem');

describe('useCommunityCardAnnouncement', () => {
  it('announces only newly revealed cards and resets for a new hand', () => {
    const { result, rerender } = renderHook(({ communityCards }) => useCommunityCardAnnouncement(communityCards, t), {
      initialProps: { communityCards: [] as Card[] },
    });

    expect(result.current).toBe('');
    rerender({ communityCards: cards.slice(0, 3) });
    expect(result.current).toBe('コミュニティカードが公開されました: ♠ 10、♥ 5、♦ 8');
    rerender({ communityCards: cards.slice(0, 4) });
    expect(result.current).toBe('コミュニティカードが公開されました: ♣ 2');
    rerender({ communityCards: cards });
    expect(result.current).toBe('コミュニティカードが公開されました: ♠ 4');
    rerender({ communityCards: [] });
    expect(result.current).toBe('');
    rerender({ communityCards: cards.slice(0, 3) });
    expect(result.current).toBe('コミュニティカードが公開されました: ♠ 10、♥ 5、♦ 8');
  });
});
