/**
 * @vitest-environment jsdom
 */
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { AXES } from '../../constants/discoverAxes';
import { MoodQuestion } from './MoodQuestion';

describe('MoodQuestion', () => {
  it('renders all options for the mood axis', () => {
    render(
      <MoodQuestion
        axis={AXES.mood}
        questionIndex={0}
        selected={null}
        onSelect={() => {}}
        onSkip={() => {}}
        questionNumber={1}
        totalQuestions={8}
      />,
    );
    expect(screen.getAllByRole('button').length).toBeGreaterThanOrEqual(AXES.mood.options.length);
  });

  it('marks the selected option with aria-pressed=true', () => {
    render(
      <MoodQuestion
        axis={AXES.mood}
        questionIndex={0}
        selected={2}
        onSelect={() => {}}
        onSkip={() => {}}
        questionNumber={1}
        totalQuestions={8}
      />,
    );
    const pressed = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') === 'true');
    expect(pressed).toHaveLength(1);
  });

  it('fires onSelect with the option index when clicked', () => {
    const onSelect = vi.fn();
    render(
      <MoodQuestion
        axis={AXES.mood}
        questionIndex={0}
        selected={null}
        onSelect={onSelect}
        onSkip={() => {}}
        questionNumber={1}
        totalQuestions={8}
      />,
    );
    const buttons = screen.getAllByRole('button');
    // First button is option 0 (skip is rendered after options).
    fireEvent.click(buttons[0]);
    expect(onSelect).toHaveBeenCalledWith(0);
  });

  it('fires onSkip when the skip button is clicked', () => {
    const onSkip = vi.fn();
    render(
      <MoodQuestion
        axis={AXES.mood}
        questionIndex={0}
        selected={null}
        onSelect={() => {}}
        onSkip={onSkip}
        questionNumber={1}
        totalQuestions={8}
      />,
    );
    // Last button is the skip control.
    const buttons = screen.getAllByRole('button');
    fireEvent.click(buttons[buttons.length - 1]);
    expect(onSkip).toHaveBeenCalled();
  });
});
