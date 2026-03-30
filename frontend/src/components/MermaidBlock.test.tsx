import { render, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

const mockRender = vi.fn().mockResolvedValue({ svg: '<svg data-testid="mermaid-svg">diagram</svg>' });

vi.mock('mermaid', () => ({
  default: {
    initialize: vi.fn(),
    render: mockRender,
  },
}));

import { MermaidBlock } from './MermaidBlock';

describe('MermaidBlock', () => {
  it('renders SVG from mermaid code', async () => {
    const { container } = render(<MermaidBlock code="flowchart TD\n    A-->B" />);
    await waitFor(() => {
      expect(container.querySelector('svg')).toBeTruthy();
    });
  });

  it('calls mermaid.render with the provided code', async () => {
    const code = 'graph LR\n    X-->Y';
    render(<MermaidBlock code={code} />);
    await waitFor(() => {
      const calls = mockRender.mock.calls;
      const lastCall = calls[calls.length - 1];
      expect(lastCall[1]).toBe(code);
    });
  });

  it('renders fallback code block on error', async () => {
    mockRender.mockRejectedValueOnce(new Error('parse error'));
    const { container } = render(<MermaidBlock code="invalid" />);
    await waitFor(() => {
      const code = container.querySelector('code');
      expect(code?.textContent).toBe('invalid');
    });
  });
});
