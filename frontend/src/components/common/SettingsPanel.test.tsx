import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { SettingsGroup } from './SettingsPanel';
import { SettingsPanel } from './SettingsPanel';

describe('SettingsPanel', () => {
  it('renders summary title', () => {
    render(<SettingsPanel title="Settings" groups={[]} />);
    expect(screen.getByText('Settings')).toBeInTheDocument();
  });

  it('renders checkbox item and toggles it', () => {
    const onToggle = vi.fn();
    render(
      <SettingsPanel
        title="Settings"
        groups={[{ items: [{ type: 'checkbox', id: 'cb1', label: 'Check me', checked: false, onToggle }] }]}
      />,
    );
    const checkbox = screen.getByLabelText('Check me') as HTMLInputElement;
    expect(checkbox.checked).toBe(false);
    fireEvent.click(checkbox);
    expect(onToggle).toHaveBeenCalledWith(true);
  });

  it('renders checkbox with checked=true', () => {
    render(
      <SettingsPanel
        title="Settings"
        groups={[{ items: [{ type: 'checkbox', id: 'cb1', label: 'Check me', checked: true }] }]}
      />,
    );
    const checkbox = screen.getByLabelText('Check me') as HTMLInputElement;
    expect(checkbox.checked).toBe(true);
  });

  it('renders checkbox with default checked when checked is omitted', () => {
    render(
      <SettingsPanel title="Settings" groups={[{ items: [{ type: 'checkbox', id: 'cb1', label: 'No checked' }] }]} />,
    );
    const checkbox = screen.getByLabelText('No checked') as HTMLInputElement;
    expect(checkbox.checked).toBe(false);
  });

  it('renders select item and fires onSelect', () => {
    const onSelect = vi.fn();
    render(
      <SettingsPanel
        title="Settings"
        groups={[
          {
            items: [
              {
                type: 'select',
                id: 'sel1',
                label: 'Pick one',
                value: 'a',
                options: [
                  { value: 'a', label: 'Alpha' },
                  { value: 'b', label: 'Beta' },
                ],
                onSelect,
              },
            ],
          },
        ]}
      />,
    );
    const select = screen.getByLabelText('Pick one') as HTMLSelectElement;
    expect(select.value).toBe('a');
    fireEvent.change(select, { target: { value: 'b' } });
    expect(onSelect).toHaveBeenCalledWith('b');
  });

  it('renders disabled checkbox', () => {
    render(
      <SettingsPanel
        title="Settings"
        groups={[{ items: [{ type: 'checkbox', id: 'cb1', label: 'Disabled Check', checked: false, disabled: true }] }]}
      />,
    );
    expect(screen.getByLabelText('Disabled Check')).toBeDisabled();
  });

  it('renders enabled checkbox when disabled is false', () => {
    render(
      <SettingsPanel
        title="Settings"
        groups={[{ items: [{ type: 'checkbox', id: 'cb1', label: 'Enabled Check', checked: false, disabled: false }] }]}
      />,
    );
    expect(screen.getByLabelText('Enabled Check')).not.toBeDisabled();
  });

  it('renders disabled select', () => {
    render(
      <SettingsPanel
        title="Settings"
        groups={[
          {
            items: [
              {
                type: 'select',
                id: 'sel1',
                label: 'Disabled',
                value: '1',
                options: [{ value: '1', label: 'One' }],
                disabled: true,
              },
            ],
          },
        ]}
      />,
    );
    expect(screen.getByLabelText('Disabled')).toBeDisabled();
  });

  it('disabled select has opacity-70 class for readability', () => {
    render(
      <SettingsPanel
        title="Settings"
        groups={[
          {
            items: [
              {
                type: 'select',
                id: 'sel1',
                label: 'Disabled',
                value: '1',
                options: [{ value: '1', label: 'One' }],
                disabled: true,
              },
            ],
          },
        ]}
      />,
    );
    const select = screen.getByLabelText('Disabled');
    expect(select.className).toContain('disabled:opacity-70');
    expect(select.className).toContain('disabled:text-gray-300');
  });

  it('renders enabled select when disabled is false', () => {
    render(
      <SettingsPanel
        title="Settings"
        groups={[
          {
            items: [
              {
                type: 'select',
                id: 'sel1',
                label: 'Enabled',
                value: '1',
                options: [{ value: '1', label: 'One' }],
                disabled: false,
              },
            ],
          },
        ]}
      />,
    );
    expect(screen.getByLabelText('Enabled')).not.toBeDisabled();
  });

  it('renders group with title as fieldset with legend', () => {
    render(
      <SettingsPanel
        title="Settings"
        groups={[
          {
            title: 'Group A',
            items: [{ type: 'checkbox', id: 'cb1', label: 'Item1', checked: false }],
          },
        ]}
      />,
    );
    expect(screen.getByText('Group A')).toBeInTheDocument();
    expect(screen.getByText('Group A').tagName).toBe('LEGEND');
  });

  it('renders group without title as div (no fieldset)', () => {
    const { container } = render(
      <SettingsPanel
        title="Settings"
        groups={[{ items: [{ type: 'checkbox', id: 'cb1', label: 'Item1', checked: false }] }]}
      />,
    );
    expect(container.querySelector('fieldset')).toBeNull();
  });

  it('renders multiple groups with correct spacing', () => {
    const groups: SettingsGroup[] = [
      { title: 'G1', items: [{ type: 'checkbox', id: 'c1', label: 'A', checked: false }] },
      { title: 'G2', items: [{ type: 'checkbox', id: 'c2', label: 'B', checked: false }] },
    ];
    render(<SettingsPanel title="Settings" groups={groups} />);
    expect(screen.getByText('G1')).toBeInTheDocument();
    expect(screen.getByText('G2')).toBeInTheDocument();
  });

  it('renders tooltip on checkbox hover with aria-describedby', () => {
    const { container } = render(
      <SettingsPanel
        title="Settings"
        groups={[
          {
            items: [{ type: 'checkbox', id: 'cb1', label: 'With tip', checked: false, tooltip: 'Help text' }],
          },
        ]}
      />,
    );
    const tooltip = screen.getByText('Help text');
    expect(tooltip).toBeInTheDocument();
    expect(tooltip).toHaveAttribute('id', 'cb1-tooltip');
    expect(tooltip).toHaveAttribute('role', 'tooltip');
    const checkbox = container.querySelector('#cb1') as HTMLInputElement;
    expect(checkbox).toHaveAttribute('aria-describedby', 'cb1-tooltip');
  });

  it('renders tooltip on select hover with aria-describedby', () => {
    render(
      <SettingsPanel
        title="Settings"
        groups={[
          {
            items: [
              {
                type: 'select',
                id: 'sel1',
                label: 'Sel',
                value: '1',
                options: [{ value: '1', label: 'One' }],
                tooltip: 'Select help',
              },
            ],
          },
        ]}
      />,
    );
    const tooltip = screen.getByText('Select help');
    expect(tooltip).toBeInTheDocument();
    expect(tooltip).toHaveAttribute('id', 'sel1-tooltip');
    expect(tooltip).toHaveAttribute('role', 'tooltip');
    const select = screen.getByLabelText('Sel');
    expect(select).toHaveAttribute('aria-describedby', 'sel1-tooltip');
  });

  it('does not set aria-describedby when no tooltip on checkbox', () => {
    render(
      <SettingsPanel
        title="Settings"
        groups={[{ items: [{ type: 'checkbox', id: 'cb1', label: 'No tip', checked: false }] }]}
      />,
    );
    expect(screen.getByLabelText('No tip')).not.toHaveAttribute('aria-describedby');
  });

  it('does not set aria-describedby when no tooltip on select', () => {
    render(
      <SettingsPanel
        title="Settings"
        groups={[
          {
            items: [
              { type: 'select', id: 'sel1', label: 'No tip', value: '1', options: [{ value: '1', label: 'One' }] },
            ],
          },
        ]}
      />,
    );
    expect(screen.getByLabelText('No tip')).not.toHaveAttribute('aria-describedby');
  });

  it('does not render tooltip when not provided on checkbox', () => {
    const { container } = render(
      <SettingsPanel
        title="Settings"
        groups={[{ items: [{ type: 'checkbox', id: 'cb1', label: 'No tip', checked: false }] }]}
      />,
    );
    const label = container.querySelector('label[for="cb1"]');
    const tooltipSpans = label?.querySelectorAll('span.hidden');
    expect(tooltipSpans?.length ?? 0).toBe(0);
  });

  it('does not render tooltip when not provided on select', () => {
    const { container } = render(
      <SettingsPanel
        title="Settings"
        groups={[
          {
            items: [
              { type: 'select', id: 'sel1', label: 'No tip', value: '1', options: [{ value: '1', label: 'One' }] },
            ],
          },
        ]}
      />,
    );
    const wrapper = container.querySelector('span.flex.items-center');
    const tooltipSpans = wrapper?.querySelectorAll('span.hidden');
    expect(tooltipSpans?.length ?? 0).toBe(0);
  });

  it('handles checkbox toggle without onToggle (no-op)', () => {
    render(
      <SettingsPanel
        title="Settings"
        groups={[{ items: [{ type: 'checkbox', id: 'cb1', label: 'No handler', checked: false }] }]}
      />,
    );
    const checkbox = screen.getByLabelText('No handler');
    expect(() => fireEvent.click(checkbox)).not.toThrow();
  });

  it('handles select change without onSelect (no-op)', () => {
    render(
      <SettingsPanel
        title="Settings"
        groups={[
          {
            items: [
              {
                type: 'select',
                id: 'sel1',
                label: 'No handler',
                value: 'a',
                options: [
                  { value: 'a', label: 'A' },
                  { value: 'b', label: 'B' },
                ],
              },
            ],
          },
        ]}
      />,
    );
    const select = screen.getByLabelText('No handler');
    expect(() => fireEvent.change(select, { target: { value: 'b' } })).not.toThrow();
  });

  it('renders select without options (empty)', () => {
    render(
      <SettingsPanel
        title="Settings"
        groups={[{ items: [{ type: 'select', id: 'sel1', label: 'Empty', value: '' }] }]}
      />,
    );
    const select = screen.getByLabelText('Empty') as HTMLSelectElement;
    expect(select.options).toHaveLength(0);
  });

  it('renders multiple groups with mixed titled and untitled', () => {
    const groups: SettingsGroup[] = [
      { items: [{ type: 'checkbox', id: 'c1', label: 'Untitled item', checked: false }] },
      { title: 'Titled', items: [{ type: 'checkbox', id: 'c2', label: 'Titled item', checked: false }] },
    ];
    render(<SettingsPanel title="Settings" groups={groups} />);
    expect(screen.getByText('Untitled item')).toBeInTheDocument();
    expect(screen.getByText('Titled')).toBeInTheDocument();
    expect(screen.getByText('Titled item')).toBeInTheDocument();
  });
});
