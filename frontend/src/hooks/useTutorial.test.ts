import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { TutorialConfig } from '../types/tutorial';
import { useTutorial } from './useTutorial';

const config: TutorialConfig = {
  gameName: 'testgame',
  steps: [
    { target: '[data-tutorial="step1"]', messageKey: 'step1', placement: 'bottom', advanceOn: 'next' },
    { target: '[data-tutorial="step2"]', messageKey: 'step2', placement: 'top', advanceOn: 'click' },
    { target: '[data-tutorial="step3"]', messageKey: 'step3', placement: 'right', advanceOn: 'next' },
  ],
};

describe('useTutorial', () => {
  afterEach(() => {
    localStorage.clear();
  });

  it('initializes as inactive', () => {
    const { result } = renderHook(() => useTutorial(config));
    expect(result.current.isActive).toBe(false);
    expect(result.current.currentStepIndex).toBe(0);
    expect(result.current.currentStep).toBeNull();
    expect(result.current.isCompleted).toBe(false);
  });

  it('starts the tutorial', () => {
    const { result } = renderHook(() => useTutorial(config));
    act(() => result.current.start());
    expect(result.current.isActive).toBe(true);
    expect(result.current.currentStepIndex).toBe(0);
    expect(result.current.currentStep).toEqual(config.steps[0]);
    expect(result.current.totalSteps).toBe(3);
  });

  it('advances to the next step', () => {
    const { result } = renderHook(() => useTutorial(config));
    act(() => result.current.start());
    act(() => result.current.next());
    expect(result.current.currentStepIndex).toBe(1);
    expect(result.current.currentStep).toEqual(config.steps[1]);
  });

  it('completes when advancing past the last step', () => {
    const { result } = renderHook(() => useTutorial(config));
    act(() => result.current.start());
    act(() => result.current.next());
    act(() => result.current.next());
    act(() => result.current.next());
    expect(result.current.isActive).toBe(false);
    expect(result.current.isCompleted).toBe(true);
    expect(localStorage.getItem('tutorial_completed_testgame')).toBe('true');
  });

  it('skips the tutorial', () => {
    const { result } = renderHook(() => useTutorial(config));
    act(() => result.current.start());
    act(() => result.current.skip());
    expect(result.current.isActive).toBe(false);
    expect(result.current.currentStepIndex).toBe(0);
  });

  it('skip does not mark as completed', () => {
    const { result } = renderHook(() => useTutorial(config));
    act(() => result.current.start());
    act(() => result.current.skip());
    expect(result.current.isCompleted).toBe(false);
    expect(localStorage.getItem('tutorial_completed_testgame')).toBeNull();
  });

  it('reads completed state from localStorage on mount', () => {
    localStorage.setItem('tutorial_completed_testgame', 'true');
    const { result } = renderHook(() => useTutorial(config));
    expect(result.current.isCompleted).toBe(true);
  });

  it('calls onEnter when a step becomes active', () => {
    const onEnter = vi.fn();
    const configWithEnter: TutorialConfig = {
      gameName: 'testgame',
      steps: [
        { target: '[data-tutorial="s1"]', messageKey: 's1', placement: 'bottom', advanceOn: 'next' },
        { target: '[data-tutorial="s2"]', messageKey: 's2', placement: 'top', advanceOn: 'next', onEnter },
      ],
    };
    const { result } = renderHook(() => useTutorial(configWithEnter));
    act(() => result.current.start());
    expect(onEnter).not.toHaveBeenCalled();
    act(() => result.current.next());
    expect(onEnter).toHaveBeenCalledTimes(1);
  });

  it('calls onEnter for the first step on start', () => {
    const onEnter = vi.fn();
    const configWithEnter: TutorialConfig = {
      gameName: 'testgame',
      steps: [{ target: '[data-tutorial="s1"]', messageKey: 's1', placement: 'bottom', advanceOn: 'next', onEnter }],
    };
    const { result } = renderHook(() => useTutorial(configWithEnter));
    act(() => result.current.start());
    expect(onEnter).toHaveBeenCalledTimes(1);
  });

  it('next does nothing when not active', () => {
    const { result } = renderHook(() => useTutorial(config));
    act(() => result.current.next());
    expect(result.current.currentStepIndex).toBe(0);
    expect(result.current.isActive).toBe(false);
  });

  it('skip does nothing when not active', () => {
    const { result } = renderHook(() => useTutorial(config));
    act(() => result.current.skip());
    expect(result.current.isActive).toBe(false);
  });

  it('handles empty steps config', () => {
    const emptyConfig: TutorialConfig = { gameName: 'empty', steps: [] };
    const { result } = renderHook(() => useTutorial(emptyConfig));
    act(() => result.current.start());
    expect(result.current.isActive).toBe(false);
    expect(result.current.isCompleted).toBe(true);
  });

  it('preserves completed state on restart', () => {
    localStorage.setItem('tutorial_completed_testgame', 'true');
    const { result } = renderHook(() => useTutorial(config));
    expect(result.current.isCompleted).toBe(true);
    act(() => result.current.start());
    expect(result.current.isActive).toBe(true);
    expect(result.current.isCompleted).toBe(true);
  });
});
