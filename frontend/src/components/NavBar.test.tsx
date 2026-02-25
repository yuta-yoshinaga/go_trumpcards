import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { NavBar } from './NavBar';

function renderNavBar(initialPath = '/') {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <NavBar />
    </MemoryRouter>,
  );
}

describe('NavBar', () => {
  it('renders six navigation links', () => {
    renderNavBar();
    const links = screen.getAllByRole('link');
    expect(links).toHaveLength(6);
  });

  it('renders BlackJack link', () => {
    renderNavBar();
    expect(screen.getByText('ブラックジャック')).toBeInTheDocument();
  });

  it('renders Poker link', () => {
    renderNavBar();
    expect(screen.getByText('ポーカー')).toBeInTheDocument();
  });

  it('renders OldMaid link', () => {
    renderNavBar();
    expect(screen.getByText('ババ抜き')).toBeInTheDocument();
  });

  it('renders Daifugo link', () => {
    renderNavBar();
    expect(screen.getByText('大富豪')).toBeInTheDocument();
  });

  it('renders Sevens link', () => {
    renderNavBar();
    expect(screen.getByText('7並べ')).toBeInTheDocument();
  });

  it('renders Doubt link', () => {
    renderNavBar();
    expect(screen.getByText('ダウト')).toBeInTheDocument();
  });

  it('marks BlackJack link as active when on root path', () => {
    renderNavBar('/');
    const blackjackLink = screen.getByText('ブラックジャック');
    expect(blackjackLink).toHaveAttribute('aria-current', 'page');
    expect(screen.getByText('ポーカー')).not.toHaveAttribute('aria-current');
    expect(screen.getByText('ババ抜き')).not.toHaveAttribute('aria-current');
    expect(screen.getByText('大富豪')).not.toHaveAttribute('aria-current');
  });

  it('marks Poker link as active when on /poker path', () => {
    renderNavBar('/poker');
    expect(screen.getByText('ポーカー')).toHaveAttribute('aria-current', 'page');
    expect(screen.getByText('ブラックジャック')).not.toHaveAttribute('aria-current');
    expect(screen.getByText('ババ抜き')).not.toHaveAttribute('aria-current');
    expect(screen.getByText('大富豪')).not.toHaveAttribute('aria-current');
  });

  it('marks OldMaid link as active when on /oldmaid path', () => {
    renderNavBar('/oldmaid');
    expect(screen.getByText('ババ抜き')).toHaveAttribute('aria-current', 'page');
    expect(screen.getByText('ブラックジャック')).not.toHaveAttribute('aria-current');
    expect(screen.getByText('ポーカー')).not.toHaveAttribute('aria-current');
    expect(screen.getByText('大富豪')).not.toHaveAttribute('aria-current');
  });

  it('marks Daifugo link as active when on /daifugo path', () => {
    renderNavBar('/daifugo');
    expect(screen.getByText('大富豪')).toHaveAttribute('aria-current', 'page');
    expect(screen.getByText('ブラックジャック')).not.toHaveAttribute('aria-current');
    expect(screen.getByText('ポーカー')).not.toHaveAttribute('aria-current');
    expect(screen.getByText('ババ抜き')).not.toHaveAttribute('aria-current');
  });

  it('marks Sevens link as active when on /sevens path', () => {
    renderNavBar('/sevens');
    expect(screen.getByText('7並べ')).toHaveAttribute('aria-current', 'page');
    expect(screen.getByText('ブラックジャック')).not.toHaveAttribute('aria-current');
    expect(screen.getByText('ポーカー')).not.toHaveAttribute('aria-current');
    expect(screen.getByText('ババ抜き')).not.toHaveAttribute('aria-current');
    expect(screen.getByText('大富豪')).not.toHaveAttribute('aria-current');
  });

  it('marks Doubt link as active when on /doubt path', () => {
    renderNavBar('/doubt');
    expect(screen.getByText('ダウト')).toHaveAttribute('aria-current', 'page');
    expect(screen.getByText('ブラックジャック')).not.toHaveAttribute('aria-current');
    expect(screen.getByText('ポーカー')).not.toHaveAttribute('aria-current');
    expect(screen.getByText('ババ抜き')).not.toHaveAttribute('aria-current');
    expect(screen.getByText('大富豪')).not.toHaveAttribute('aria-current');
  });

  it('links point to correct hrefs', () => {
    renderNavBar();
    const links = screen.getAllByRole('link');
    expect(links[0]).toHaveAttribute('href', '/');
    expect(links[1]).toHaveAttribute('href', '/poker');
    expect(links[2]).toHaveAttribute('href', '/oldmaid');
    expect(links[3]).toHaveAttribute('href', '/daifugo');
    expect(links[4]).toHaveAttribute('href', '/sevens');
    expect(links[5]).toHaveAttribute('href', '/doubt');
  });
});
