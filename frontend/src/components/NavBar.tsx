import { Link, useLocation } from 'react-router-dom';

const games = [
  { path: '/', label: 'ブラックジャック' },
  { path: '/poker', label: 'ポーカー' },
  { path: '/oldmaid', label: 'ババ抜き' },
  { path: '/daifugo', label: '大富豪' },
  { path: '/sevens', label: '7並べ' },
  { path: '/doubt', label: 'ダウト' },
  { path: '/holdem', label: 'テキサスホールデム' },
];

export function NavBar() {
  const { pathname } = useLocation();

  return (
    <nav style={{ textAlign: 'right', margin: '8px 10px' }}>
      {games.map(({ path, label }) => (
        <Link
          key={path}
          to={path}
          aria-current={pathname === path ? 'page' : undefined}
          className={`inline-block px-2 py-0.5 text-xs font-medium rounded transition-colors ml-1${pathname === path ? ' bg-gray-400 text-white' : ' bg-gray-600 text-gray-200 hover:bg-gray-500'}`}
        >
          {label}
        </Link>
      ))}
    </nav>
  );
}
