import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { useFirstVisit } from './useFirstVisit';

describe('useFirstVisit', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  afterEach(() => {
    localStorage.clear();
  });

  it('shows dialog on first visit', () => {
    const { result } = renderHook(() => useFirstVisit('blackjack'));
    expect(result.current.shouldShowDialog).toBe(true);
  });

  it('does not show dialog if game was already visited', () => {
    localStorage.setItem('game_visited_blackjack', 'true');
    const { result } = renderHook(() => useFirstVisit('blackjack'));
    expect(result.current.shouldShowDialog).toBe(false);
  });

  it('does not show dialog if tutorial is already completed', () => {
    localStorage.setItem('tutorial_completed_blackjack', 'true');
    const { result } = renderHook(() => useFirstVisit('blackjack'));
    expect(result.current.shouldShowDialog).toBe(false);
  });

  it('does not show dialog if global no-suggest flag is set', () => {
    localStorage.setItem('tutorial_no_suggest', 'true');
    const { result } = renderHook(() => useFirstVisit('blackjack'));
    expect(result.current.shouldShowDialog).toBe(false);
  });

  it('dismiss marks game as visited and hides dialog', () => {
    const { result } = renderHook(() => useFirstVisit('blackjack'));
    act(() => result.current.dismiss());
    expect(result.current.shouldShowDialog).toBe(false);
    expect(localStorage.getItem('game_visited_blackjack')).toBe('true');
  });

  it('dismissPermanently sets global no-suggest flag', () => {
    const { result } = renderHook(() => useFirstVisit('blackjack'));
    act(() => result.current.dismissPermanently());
    expect(result.current.shouldShowDialog).toBe(false);
    expect(localStorage.getItem('game_visited_blackjack')).toBe('true');
    expect(localStorage.getItem('tutorial_no_suggest')).toBe('true');
  });

  it('dismissPermanently suppresses dialog for other games too', () => {
    const { result: r1 } = renderHook(() => useFirstVisit('blackjack'));
    act(() => r1.current.dismissPermanently());
    const { result: r2 } = renderHook(() => useFirstVisit('poker'));
    expect(r2.current.shouldShowDialog).toBe(false);
  });

  it('different games have independent visit state', () => {
    localStorage.setItem('game_visited_blackjack', 'true');
    const { result: r1 } = renderHook(() => useFirstVisit('blackjack'));
    const { result: r2 } = renderHook(() => useFirstVisit('poker'));
    expect(r1.current.shouldShowDialog).toBe(false);
    expect(r2.current.shouldShowDialog).toBe(true);
  });
});
