import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { TutorialStep } from '../../types/tutorial';
import { TutorialWrapper } from './TutorialWrapper';

vi.mock('react-i18next', () => ({
  useTranslation: (ns: string) => ({ t: (key: string) => `${ns}:${key}`, i18n: { language: 'ja' } }),
}));

vi.mock('../../providers/TutorialProvider', () => ({
  TutorialProvider: ({ children, config }: { children: React.ReactNode; config: { gameName: string } }) => (
    <div data-testid={`tutorial-provider-${config.gameName}`}>{children}</div>
  ),
}));

const STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="test"]',
    messageKey: 'tutorial.test',
    placement: 'top',
    advanceOn: 'next',
  },
];

describe('TutorialWrapper', () => {
  it('renders children inside TutorialProvider', () => {
    render(
      <TutorialWrapper gameName="testgame" steps={STEPS}>
        <div data-testid="child">hello</div>
      </TutorialWrapper>,
    );
    expect(screen.getByTestId('tutorial-provider-testgame')).toBeInTheDocument();
    expect(screen.getByTestId('child')).toBeInTheDocument();
  });
});
