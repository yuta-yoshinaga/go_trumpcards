import { useEffect, useState } from 'react';
import { tehonbikiApi } from '../api/games/tehonbiki';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnPrimary } from '../styles/buttonStyles';

/** Renders the Tehonbiki number-guessing game. */
export const TehonbikiPage = withTutorial(TehonbikiPageContent, 'tehonbiki', []);
function TehonbikiPageContent() {
  const { t, tc, confirmOpen, requestConfirm, confirmReset, cancelReset } = useGamePageSetup('tehonbiki');
  const { state, loading, error, exec, retry } = useGameApi(tehonbikiApi.exec);
  const [betType, setBetType] = useState('single');
  const [numbers, setNumbers] = useState<number[]>([1]);
  const [bet, setBet] = useState(50);
  useEffect(() => {
    exec('reset');
  }, [exec]);
  if (!state) return <GameSkeleton gameKey="tehonbiki" layout={{ kind: 'casino-table', sections: [1] }} />;
  const toggle = (n: number) => setNumbers((v) => (v.includes(n) ? v.filter((x) => x !== n) : [...v, n]));
  return (
    <GamePageShell
      title={tc('nav.tehonbiki')}
      gameThemeBg="bg-ds-surface"
      phaseName={String(state.phase)}
      gamePath="/tehonbiki"
      gameEndFlag={state.gameEndFlag}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
    >
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
      <div className="text-center text-ds-text-primary p-6">
        <p>
          {t('label.chips')}: {state.chips}
        </p>
        {state.parentCard && (
          <p>
            {t('label.parentCard')}: {state.parentCard}
          </p>
        )}
        <div className="flex justify-center gap-2 p-4">
          {[1, 2, 3, 4, 5, 6].map((n) => (
            <button
              className={btnPrimary}
              type="button"
              key={n}
              onClick={() => toggle(n)}
              disabled={state.phase !== 0 || loading}
              aria-pressed={numbers.includes(n)}
            >
              {n}
            </button>
          ))}
        </div>
        {state.phase === 0 ? (
          <>
            <select value={betType} onChange={(e) => setBetType(e.target.value)}>
              <option value="single">single</option>
              <option value="double">double</option>
              <option value="triple">triple</option>
              <option value="half">half</option>
            </select>
            <input type="number" value={bet} onChange={(e) => setBet(Number(e.target.value))} />
            <button className={btnPrimary} type="button" onClick={() => exec('bet', { numbers, betType, bet })}>
              {t('button.bet')}
            </button>
          </>
        ) : (
          !state.gameEndFlag && (
            <button className={btnPrimary} type="button" onClick={() => exec('next')}>
              {t('button.next')}
            </button>
          )
        )}
      </div>
      <GameFooter>
        <ErrorAlert message={error} onRetry={retry} />
        <GameResetButton
          isGameEnd={state.gameEndFlag}
          onReset={() => exec('reset')}
          requestConfirm={requestConfirm}
          loading={loading}
        />
      </GameFooter>
    </GamePageShell>
  );
}
