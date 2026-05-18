/**
 * @vitest-environment jsdom
 */
import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { TOTAL_QUESTIONS } from '../../constants/discoverAxes';
import { SurveyProgress } from './SurveyProgress';

describe('SurveyProgress', () => {
  it('renders TOTAL_QUESTIONS cards', () => {
    render(<SurveyProgress current={1} />);
    expect(screen.getAllByTestId('survey-progress-card')).toHaveLength(TOTAL_QUESTIONS);
  });

  it('marks exactly one card as current', () => {
    render(<SurveyProgress current={3} />);
    const cards = screen.getAllByTestId('survey-progress-card');
    const current = cards.filter((c) => c.dataset.state === 'current');
    expect(current).toHaveLength(1);
    expect(cards.indexOf(current[0])).toBe(2);
  });

  it('marks previous cards as done', () => {
    render(<SurveyProgress current={5} />);
    const cards = screen.getAllByTestId('survey-progress-card');
    const done = cards.filter((c) => c.dataset.state === 'done');
    expect(done).toHaveLength(4);
  });
});
