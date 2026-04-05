import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { TutorialStep } from '../../types/tutorial';
import { TutorialWrapper } from './TutorialWrapper';

vi.mock('react-i18next', () => ({
  useTranslation: (ns: string) => ({ t: (key: string) => `${ns}:${key}`, i18n: { language: 'ja' } }),
}));

vi.mock('../../providers/TutorialProvider', () => ({
  TutorialProvider: ({
    children,
    config,
    translateMessage,
  }: {
    children: React.ReactNode;
    config: { gameName: string; steps: TutorialStep[] };
    translateMessage: (key: string) => string;
  }) => (
    <div data-testid={`tutorial-provider-${config.gameName}`}>
      <div data-testid="config-steps">{JSON.stringify(config.steps)}</div>
      <div data-testid="translated-message">{translateMessage('some.key')}</div>
      {children}
    </div>
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
  it('renders children inside TutorialProvider with correct props', () => {
    render(
      <TutorialWrapper gameName="testgame" steps={STEPS}>
        <div data-testid="child">hello</div>
      </TutorialWrapper>,
    );
    expect(screen.getByTestId('tutorial-provider-testgame')).toBeInTheDocument();
    expect(screen.getByTestId('child')).toBeInTheDocument();
    expect(screen.getByTestId('config-steps')).toHaveTextContent(JSON.stringify(STEPS));
    expect(screen.getByTestId('translated-message')).toHaveTextContent('testgame:some.key');
  });
});
