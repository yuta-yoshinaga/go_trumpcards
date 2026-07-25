import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { HintResult } from '../../types/hint';
import { FrontendHintTooltip } from './FrontendHintTooltip';

const hint: HintResult = { reason: 'hint.reason.key', confidence: 'strong' };
const t = (key: string) => (key === 'hint.reason.key' ? '最善手です' : key);

describe('FrontendHintTooltip', () => {
  it('renders the tooltip with the translated reason when enabled and a hint exists', () => {
    render(<FrontendHintTooltip hint={hint} enabled={true} t={t} />);
    expect(screen.getByTestId('hint-tooltip')).toHaveTextContent('最善手です');
  });

  it('renders nothing when disabled', () => {
    const { container } = render(<FrontendHintTooltip hint={hint} enabled={false} t={t} />);
    expect(container).toBeEmptyDOMElement();
  });

  it('renders nothing when there is no hint', () => {
    const { container } = render(<FrontendHintTooltip hint={null} enabled={true} t={t} />);
    expect(container).toBeEmptyDOMElement();
  });
});
