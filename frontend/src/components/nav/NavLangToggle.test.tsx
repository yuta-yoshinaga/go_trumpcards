import { fireEvent, render, screen } from '@testing-library/react';
import i18n from 'i18next';
import { afterEach, describe, expect, it } from 'vitest';
import { NavLangToggle } from './NavLangToggle';

afterEach(() => {
  i18n.changeLanguage('ja');
});

describe('NavLangToggle', () => {
  it('renders JA and EN buttons with the project i18n labels', () => {
    render(<NavLangToggle />);
    expect(screen.getByRole('button', { name: i18n.t('nav.switchToJa') })).toHaveTextContent('JA');
    expect(screen.getByRole('button', { name: i18n.t('nav.switchToEn') })).toHaveTextContent('EN');
  });

  it('marks JA as the pressed button when current language is ja (default test setup)', () => {
    render(<NavLangToggle />);
    expect(screen.getByRole('button', { name: i18n.t('nav.switchToJa') })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: i18n.t('nav.switchToEn') })).toHaveAttribute('aria-pressed', 'false');
  });

  it('switches language to en and flips aria-pressed when EN is clicked', () => {
    render(<NavLangToggle />);
    fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.switchToEn') }));
    expect(i18n.language).toBe('en');
    expect(screen.getByRole('button', { name: i18n.t('nav.switchToEn') })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: i18n.t('nav.switchToJa') })).toHaveAttribute('aria-pressed', 'false');
  });

  it('switches back to ja and flips aria-pressed when JA is clicked', () => {
    render(<NavLangToggle />);
    fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.switchToEn') }));
    fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.switchToJa') }));
    expect(i18n.language).toBe('ja');
    expect(screen.getByRole('button', { name: i18n.t('nav.switchToJa') })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: i18n.t('nav.switchToEn') })).toHaveAttribute('aria-pressed', 'false');
  });

  it('applies the md size classes by default (44 px touch target)', () => {
    render(<NavLangToggle />);
    const ja = screen.getByRole('button', { name: i18n.t('nav.switchToJa') });
    expect(ja.className).toContain('px-3');
    expect(ja.className).toContain('py-2');
    expect(ja.className).toContain('min-h-[44px]');
  });

  it('applies the sm size classes when size="sm"', () => {
    render(<NavLangToggle size="sm" />);
    const ja = screen.getByRole('button', { name: i18n.t('nav.switchToJa') });
    expect(ja.className).toContain('px-2');
    expect(ja.className).toContain('py-1');
    expect(ja.className).not.toContain('min-h-[44px]');
  });
});
