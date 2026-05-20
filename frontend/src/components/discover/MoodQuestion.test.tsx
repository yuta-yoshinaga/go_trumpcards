/**
 * @vitest-environment jsdom
 */
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { AXES } from '../../constants/discoverAxes';
import { MoodQuestion } from './MoodQuestion';

describe('MoodQuestion', () => {
  it('renders all options for the selected sub-question', () => {
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
    // Buttons include the per-option buttons + the skip button.
    expect(screen.getAllByRole('button').length).toBeGreaterThanOrEqual(AXES.mood.questions[0].options.length);
  });

  it('marks the selected option with aria-pressed=true', () => {
    render(
      <MoodQuestion
        axis={AXES.mood}
        questionIndex={0}
        selected={1}
        onSelect={() => {}}
        onSkip={() => {}}
        questionNumber={1}
        totalQuestions={8}
      />,
    );
    const pressed = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') === 'true');
    expect(pressed).toHaveLength(1);
  });

  it('fires onSelect with the sub-question option index when clicked', () => {
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
    fireEvent.click(buttons[0]);
    expect(onSelect).toHaveBeenCalledWith(0);
  });

  it('renders the Q2 sub-question independently of Q1', () => {
    render(
      <MoodQuestion
        axis={AXES.skill}
        questionIndex={1}
        selected={null}
        onSelect={() => {}}
        onSkip={() => {}}
        questionNumber={4}
        totalQuestions={8}
      />,
    );
    // Skill Q2 has 2 options (learning_rules, prefer_familiar); Q1 has 3 (beginner/intermediate/advanced).
    expect(screen.getAllByRole('button').length).toBeGreaterThanOrEqual(AXES.skill.questions[1].options.length);
    expect(screen.getAllByRole('button').length).toBeLessThan(AXES.skill.questions[0].options.length + 2);
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
    const buttons = screen.getAllByRole('button');
    fireEvent.click(buttons[buttons.length - 1]);
    expect(onSkip).toHaveBeenCalled();
  });
});
