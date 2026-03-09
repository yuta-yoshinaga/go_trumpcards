import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { actionLogApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { SevensBoard } from '../components/sevens/SevensBoard';
import { SevensCpuArea } from '../components/sevens/SevensCpuArea';
import { SevensHumanArea } from '../components/sevens/SevensHumanArea';
import { useSevensGame } from '../hooks/useSevensGame';
import { btnPrimary, btnSecondary, btnWarning } from '../styles/buttonStyles';
import type { ActionLogEntry } from '../types/card';
import { playerName } from '../utils/playerUtils';
import { actionDesc } from '../utils/sevensUtils';

export function SevensPage() {
  const { t } = useTranslation('sevens');
  const { t: tc } = useTranslation('common');
  const {
    state,
    loading,
    error,
    exec,
    jokerCardIdx,
    setJokerCardIdx,
    cfgTunnel,
    setCfgTunnel,
    cfgTunnelSkipWidth,
    setCfgTunnelSkipWidth,
    cfgJokerCount,
    setCfgJokerCount,
    cfgCpuStrategy,
    setCfgCpuStrategy,
    cfgMaxPasses,
    setCfgMaxPasses,
    cfgNoJokerFinish,
    setCfgNoJokerFinish,
    cfgJokerReclaim,
    setCfgJokerReclaim,
    cfgEndStop,
    setCfgEndStop,
    cfgJokerConsBan,
    setCfgJokerConsBan,
    handleCardPlay,
    handleJokerPlace,
  } = useSevensGame();

  const [actionLog, setActionLog] = useState<ActionLogEntry[] | null>(null);

  if (!state) return null;

  const tablePlaced = state.tablePlaced;
  const tunnelEnabled = state.config.tunnelEnabled;
  const isHumanTurn = !state.gameEndFlag && !!state.players[state.currentTurn]?.isHuman;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const canPass =
    isHumanTurn &&
    humanPlayer != null &&
    (humanPlayer.maxPasses === 0 || humanPlayer.passesUsed < humanPlayer.maxPasses);

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#1a5c1a]" aria-busy={loading} aria-live="polite">
      {loading && <span className="sr-only">{tc('status.loading')}</span>}
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        {state.config &&
          (state.config.tunnelEnabled ||
            state.config.tunnelSkipWidth >= 2 ||
            state.config.jokerCount > 0 ||
            state.config.cpuStrategy !== 0 ||
            state.config.maxPasses !== 5 ||
            state.config.noJokerFinish ||
            state.config.jokerReclaimEnabled ||
            state.config.endStopEnabled ||
            state.config.jokerConsecutiveBanned) && (
            <div className="bg-black/30 rounded-lg text-yellow-300 py-1.5 px-3 mb-2 text-[0.85em]">
              {t('rules.title')}
              {state.config.tunnelEnabled && ` ${t('rules.tunnelTag')}`}
              {state.config.tunnelSkipWidth >= 2 &&
                ` ${t('rules.tunnelSkipTag', { width: state.config.tunnelSkipWidth })}`}
              {state.config.jokerCount > 0 && ` ${t('rules.jokerTag', { count: state.config.jokerCount })}`}
              {state.config.cpuStrategy === 1 && ` ${t('rules.cpuStrategy')}`}
              {state.config.cpuStrategy === 2 && ` ${t('rules.cpuHarassment')}`}
              {state.config.maxPasses === 0 && ` ${t('rules.passUnlimited')}`}
              {state.config.maxPasses !== 5 &&
                state.config.maxPasses !== 0 &&
                ` ${t('rules.passCount', { count: state.config.maxPasses })}`}
              {state.config.noJokerFinish && ` ${t('rules.noJokerFinish')}`}
              {state.config.jokerReclaimEnabled && ` ${t('rules.jokerReclaim')}`}
              {state.config.endStopEnabled && ` ${t('rules.endStop')}`}
              {state.config.jokerConsecutiveBanned && ` ${t('rules.jokerConsecutiveBanned')}`}
            </div>
          )}

        <div className="flex gap-2.5 flex-wrap mb-2.5">
          {cpuPlayers.map((player) => (
            <SevensCpuArea key={player.id} player={player} isCurrentTurn={state.currentTurn === player.id} />
          ))}
        </div>

        <SevensBoard
          tablePlaced={tablePlaced}
          tunnelEnabled={tunnelEnabled}
          tunnelSkipWidth={state.config.tunnelSkipWidth}
          endStopEnabled={state.config.endStopEnabled}
          jokerSelecting={jokerCardIdx !== null}
          onJokerPlace={handleJokerPlace}
        />

        {state.humanAction && (
          <div
            data-testid={state.humanAction.forcedPass ? 'human-action-forced-pass' : 'human-action'}
            className={`rounded-lg py-2 px-3.5 my-2 text-[0.85em] ${state.humanAction.forcedPass ? 'bg-red-900/50 text-[#fca] border border-red-500/50' : 'bg-black/40 text-[#cfc]'}`}
          >
            {actionDesc(state.players, state.humanAction, t)}
          </div>
        )}

        {state.cpuActions && state.cpuActions.length > 0 && (
          <div className="bg-black/40 rounded-lg py-2 px-3.5 my-2 text-[0.85em]">
            <span className="text-[#ccc]">{tc('label.cpuActions')}</span>
            {state.cpuActions.map((a, i) => (
              <div
                key={`cpu-action-${a.playerIdx}-${i}`}
                data-testid={a.forcedPass ? `cpu-action-forced-pass-${i}` : `cpu-action-${i}`}
                className={a.forcedPass ? 'text-[#fca]' : 'text-[#ccc]'}
              >
                {actionDesc(state.players, a, t)}
              </div>
            ))}
          </div>
        )}

        <GameMessageBox
          message={
            state.gameEndFlag
              ? `${t('resultPrefix')} ${state.players
                  .filter((p) => p.rank > 0)
                  .sort((a, b) => a.rank - b.rank)
                  .map((p) =>
                    t('resultEntry', { name: playerName(p.id, p.isHuman), rank: t('rankLabel', { rank: p.rank }) }),
                  )
                  .join(' ')}`
              : state.message
          }
          messageCode={state.gameEndFlag ? undefined : state.messageCode}
          messageParams={state.gameEndFlag ? undefined : state.messageParams}
        />

        {state.gameEndFlag && !actionLog && (
          <div className="text-center my-2">
            <button
              type="button"
              className={btnSecondary}
              onClick={async () => {
                const res = await actionLogApi.sevens();
                setActionLog(res.entries);
              }}
            >
              {tc('actionLog.view')}
            </button>
          </div>
        )}
        {actionLog && <ActionLogPanel entries={actionLog} onClose={() => setActionLog(null)} />}
      </div>

      <GameFooter className="bg-[#163e16] border-white/20 px-4 py-2.5">
        {humanPlayer && (
          <div className="mb-2">
            <SevensHumanArea
              player={humanPlayer}
              isCurrentTurn={isHumanTurn}
              tablePlaced={tablePlaced}
              tunnelEnabled={tunnelEnabled}
              tunnelSkipWidth={state.config.tunnelSkipWidth}
              noJokerFinish={state.config.noJokerFinish}
              endStopEnabled={state.config.endStopEnabled}
              jokerConsecutiveBanned={state.config.jokerConsecutiveBanned}
              loading={loading}
              onPlay={handleCardPlay}
            />
          </div>
        )}

        <div className="bg-black/30 rounded-lg py-1.5 px-3 mb-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-[0.85em] text-white/80">
          <span className="text-yellow-300 font-bold">{t('config.title')}</span>
          <label className="flex items-center gap-1 cursor-pointer">
            <input type="checkbox" checked={cfgTunnel} onChange={(e) => setCfgTunnel(e.target.checked)} />
            {t('config.tunnel')}
          </label>
          <label className="flex items-center gap-1 cursor-pointer">
            {t('config.tunnelSkip')}
            <select
              value={cfgTunnelSkipWidth}
              onChange={(e) => setCfgTunnelSkipWidth(Number(e.target.value))}
              className="bg-black/50 text-white rounded px-1 py-0.5"
            >
              <option value={0}>{t('config.tunnelSkipOff')}</option>
              <option value={2}>2</option>
              <option value={3}>3</option>
              <option value={4}>4</option>
              <option value={5}>5</option>
              <option value={6}>6</option>
            </select>
          </label>
          <label className="flex items-center gap-1 cursor-pointer">
            {t('config.joker')}
            <select
              value={cfgJokerCount}
              onChange={(e) => setCfgJokerCount(Number(e.target.value))}
              className="bg-black/50 text-white rounded px-1 py-0.5"
            >
              <option value={0}>0</option>
              <option value={1}>1</option>
              <option value={2}>2</option>
            </select>
          </label>
          <label className="flex items-center gap-1 cursor-pointer">
            {t('config.cpuStrategy')}
            <select
              value={cfgCpuStrategy}
              onChange={(e) => setCfgCpuStrategy(Number(e.target.value))}
              className="bg-black/50 text-white rounded px-1 py-0.5"
            >
              <option value={0}>{t('config.cpuStrategyOff')}</option>
              <option value={1}>{t('config.cpuStrategyStrategic')}</option>
              <option value={2}>{t('config.cpuStrategyHarassment')}</option>
            </select>
          </label>
          <label className="flex items-center gap-1 cursor-pointer">
            {t('config.passCount')}
            <select
              value={cfgMaxPasses}
              onChange={(e) => setCfgMaxPasses(Number(e.target.value))}
              className="bg-black/50 text-white rounded px-1 py-0.5"
            >
              <option value={3}>3</option>
              <option value={5}>5</option>
              <option value={10}>10</option>
              <option value={0}>{t('config.passUnlimited')}</option>
            </select>
          </label>
          <label className="flex items-center gap-1 cursor-pointer">
            <input type="checkbox" checked={cfgNoJokerFinish} onChange={(e) => setCfgNoJokerFinish(e.target.checked)} />
            {t('config.noJokerFinish')}
          </label>
          <label className="flex items-center gap-1 cursor-pointer">
            <input type="checkbox" checked={cfgJokerReclaim} onChange={(e) => setCfgJokerReclaim(e.target.checked)} />
            {t('config.jokerReclaim')}
          </label>
          <label className="flex items-center gap-1 cursor-pointer">
            <input type="checkbox" checked={cfgEndStop} onChange={(e) => setCfgEndStop(e.target.checked)} />
            {t('config.endStop')}
          </label>
          <label className="flex items-center gap-1 cursor-pointer">
            <input type="checkbox" checked={cfgJokerConsBan} onChange={(e) => setCfgJokerConsBan(e.target.checked)} />
            {t('config.jokerConsecutiveBanned')}
          </label>
        </div>

        <ErrorAlert message={error} />

        <div className="text-center">
          <button
            type="button"
            className={`${btnPrimary} min-w-[90px]`}
            disabled={loading}
            onClick={() => {
              setActionLog(null);
              exec('reset', -1, 0, 0, {
                tunnelEnabled: cfgTunnel,
                tunnelSkipWidth: cfgTunnelSkipWidth,
                jokerCount: cfgJokerCount,
                cpuStrategy: cfgCpuStrategy,
                maxPasses: cfgMaxPasses,
                noJokerFinish: cfgNoJokerFinish,
                jokerReclaim: cfgJokerReclaim,
                endStop: cfgEndStop,
                jokerConsecutiveBanned: cfgJokerConsBan,
              });
            }}
          >
            {tc('button.reset')}
          </button>
          <button
            type="button"
            className={`${btnWarning} min-w-[90px]`}
            disabled={loading || !canPass}
            onClick={() => exec('play', -1)}
          >
            {tc('button.pass')}
          </button>
          {jokerCardIdx !== null && (
            <button type="button" className={`${btnWarning} min-w-[90px]`} onClick={() => setJokerCardIdx(null)}>
              {tc('button.cancel')}
            </button>
          )}
        </div>
      </GameFooter>
    </div>
  );
}
