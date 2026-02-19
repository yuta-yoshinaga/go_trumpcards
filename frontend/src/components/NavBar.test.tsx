import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { NavBar } from './NavBar'

function renderNavBar(initialPath = '/') {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <NavBar />
    </MemoryRouter>
  )
}

describe('NavBar', () => {
  it('renders three navigation links', () => {
    renderNavBar()
    const links = screen.getAllByRole('link')
    expect(links).toHaveLength(3)
  })

  it('renders BlackJack link', () => {
    renderNavBar()
    expect(screen.getByText('ブラックジャック')).toBeInTheDocument()
  })

  it('renders Poker link', () => {
    renderNavBar()
    expect(screen.getByText('ポーカー')).toBeInTheDocument()
  })

  it('renders OldMaid link', () => {
    renderNavBar()
    expect(screen.getByText('ババ抜き')).toBeInTheDocument()
  })

  it('marks BlackJack link as active when on root path', () => {
    renderNavBar('/')
    const blackjackLink = screen.getByText('ブラックジャック')
    expect(blackjackLink).toHaveClass('active')
    expect(screen.getByText('ポーカー')).not.toHaveClass('active')
    expect(screen.getByText('ババ抜き')).not.toHaveClass('active')
  })

  it('marks Poker link as active when on /poker path', () => {
    renderNavBar('/poker')
    expect(screen.getByText('ポーカー')).toHaveClass('active')
    expect(screen.getByText('ブラックジャック')).not.toHaveClass('active')
    expect(screen.getByText('ババ抜き')).not.toHaveClass('active')
  })

  it('marks OldMaid link as active when on /oldmaid path', () => {
    renderNavBar('/oldmaid')
    expect(screen.getByText('ババ抜き')).toHaveClass('active')
    expect(screen.getByText('ブラックジャック')).not.toHaveClass('active')
    expect(screen.getByText('ポーカー')).not.toHaveClass('active')
  })

  it('links point to correct hrefs', () => {
    renderNavBar()
    const links = screen.getAllByRole('link')
    expect(links[0]).toHaveAttribute('href', '/')
    expect(links[1]).toHaveAttribute('href', '/poker')
    expect(links[2]).toHaveAttribute('href', '/oldmaid')
  })
})
