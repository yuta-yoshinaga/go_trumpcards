import { Link, useLocation } from 'react-router-dom'

const games = [
  { path: '/', label: 'ブラックジャック' },
  { path: '/poker', label: 'ポーカー' },
  { path: '/oldmaid', label: 'ババ抜き' },
]

export function NavBar() {
  const { pathname } = useLocation()

  return (
    <nav style={{ textAlign: 'right', margin: '8px 10px' }}>
      {games.map(({ path, label }) => (
        <Link
          key={path}
          to={path}
          className={`btn btn-default btn-xs${pathname === path ? ' active' : ''}`}
          style={{ marginLeft: 4 }}
        >
          {label}
        </Link>
      ))}
    </nav>
  )
}
