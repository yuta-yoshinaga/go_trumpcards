import type { TFunction } from 'i18next';
import { useEffect, useMemo, useRef, useState } from 'react';
import type { klaverjasApi } from '../api/gameApi';
import { CpuAccordion } from '../components/CpuAccordion';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useKlaverjasGame } from '../hooks/useKlaverjasGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { KlaverjasResponse } from '../types/card';
import { KlaverjasPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { suitSymbolAt } from '../utils/cardAlt';
import { KLAVERJAS_HELP, parseKlaverjasCommand } from '../utils/cli/commands/klaverjasCommands';
import { formatKlaverjasState } from '../utils/cli/formatters/klaverjasFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

const playRejectReasonKeys: Record<string, string> = {
  'shared.errFollowLeadSuit': 'followSuit',
  'klaverjas.errMustPlayTrump': 'mustTrump',
  'klaverjas.errMustOvertrump': 'mustOvertrump',
};

/**
 * Card strength and points, strongest first (sync: `Klaverjas.trumpStrength` /
 * `klaverjasPlainStrength` / `cardPoints`).
 *
 * **切り札と非切り札で順序が完全に違う。**切り札は J > 9 > A > 10 …、非切り札は
 * A > 10 > K …。姉妹ゲームの Manille には早見表があるのに、より覚えにくいこちら
 * には無かった (#4757)。
 */
const KLAVERJAS_TRUMP_ROWS: ReadonlyArray<{ face: string; points: number; nameKey?: string }> = [
  { face: 'J', points: 20, nameKey: 'strengthLegend.jas' },
  { face: '9', points: 14, nameKey: 'strengthLegend.nel' },
  { face: 'A', points: 11 },
  { face: '10', points: 10 },
  { face: 'K', points: 4 },
  { face: 'Q', points: 3 },
  { face: '8', points: 0 },
  { face: '7', points: 0 },
];

/** Non-trump strength and points, strongest first. */
const KLAVERJAS_PLAIN_ROWS: ReadonlyArray<{ face: string; points: number }> = [
  { face: 'A', points: 11 },
  { face: '10', points: 10 },
  { face: 'K', points: 4 },
  { face: 'Q', points: 3 },
  { face: 'J', points: 2 },
  { face: '9', points: 0 },
  { face: '8', points: 0 },
  { face: '7', points: 0 },
];

/** Klaverjas tutorial step definitions. */
const KLAVERJAS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="klaverjas-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="klaverjas-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="klaverjas-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="klaverjas-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="klaverjas-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const KLAVERJAS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [KlaverjasPhase.PLAY]: 'play',
  [KlaverjasPhase.TRICK_END]: 'trickEnd',
  [KlaverjasPhase.ROUND_END]: 'roundEnd',
  [KlaverjasPhase.GAME_END]: 'gameEnd',
};

type KlaverjasGame = ReturnType<typeof useKlaverjasGame>;
type KlaverjasView = ReturnType<typeof useKlaverjasView>;
type KlaverjasContext = GamePageContext<KlaverjasResponse, KlaverjasGame, KlaverjasView>;

function useKlaverjasView({ game, t }: { game: KlaverjasGame; t: TFunction }) {
  const { state } = game;

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('klaverjas', KLAVERJAS_PHASE_KEYS);

  // Transient pulse whenever the combined Roem (bonus) total grows during a hand —
  // a scoring-relevant event that is otherwise easy to miss both visually and for SR users.
  const [roemPulse, setRoemPulse] = useState(false);
  const prevRoemRef = useRef<number | null>(null);
  useEffect(() => {
    const roemTotal = state?.roundRoem ? (state.roundRoem[0] ?? 0) + (state.roundRoem[1] ?? 0) : null;
    if (roemTotal == null) return;
    const prev = prevRoemRef.current;
    prevRoemRef.current = roemTotal;
    if (prev != null && roemTotal > prev) {
      setRoemPulse(true);
      const id = setTimeout(() => setRoemPulse(false), 2500);
      return () => clearTimeout(id);
    }
  }, [state?.roundRoem]);

  const roemBreakdown = useMemo(() => {
    if (!state?.roundPlayerRoem) return null;
    const entries = state.roundPlayerRoem
      .map((pts, idx) => {
        if (pts <= 0) return null;
        const p = state.players[idx];
        const name = p ? playerName(p.id, p.isHuman) : '';
        return t('roem.entry', { name, points: pts });
      })
      .filter((e): e is string => Boolean(e));
    if (entries.length === 0) return null;
    return t('roem.breakdown', { entries: entries.join(', ') });
  }, [state?.roundPlayerRoem, state?.players, t]);

  return {
    cardWidth,
    isMobile,
    phaseNames,
    roemPulse,
    roemBreakdown,
  };
}

function Board({ state, t, view }: KlaverjasContext) {
  const { cardWidth, isMobile, roemPulse, roemBreakdown } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);

  const isPlayPhase = state.phase === KlaverjasPhase.PLAY;
  const isTrickEnd = state.phase === KlaverjasPhase.TRICK_END;
  const isRoundEnd = state.phase === KlaverjasPhase.ROUND_END;
  const showLastTrickResult = isTrickEnd || isRoundEnd;
  const isGameEnd = state.phase === KlaverjasPhase.GAME_END || state.gameEndFlag;

  const humanTeam = humanIdx % 2;
  const playersWithTeams = state.players.map((p) => ({ ...p, team: p.id % 2 }));
  const trumpSymbol = suitSymbolAt(state.trumpSuit, '');

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span>{t('trump', { suit: trumpSymbol })}</span>
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={playersWithTeams}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            dataTutorial="klaverjas-trick-display"
          />
          {isPlayPhase && state.currentTrick.length > 0 && (
            <div className="my-2 text-center text-sm text-ds-text-primary" data-testid="klaverjas-current-trick-points">
              {t('currentTrickPoints', { points: state.currentTrickPoints })}
            </div>
          )}
          {/* Show the resolved trick result until the next trick starts. */}
          {showLastTrickResult && state.lastTrickTeam >= 0 && (
            <div
              className="my-2 p-2 rounded bg-ds-accent/15 text-center text-sm font-semibold text-ds-accent"
              role="status"
              aria-live="polite"
              data-testid="klaverjas-trick-winner"
            >
              {t(state.lastTrickBonus > 0 ? 'trickResultWithBonus' : 'trickResult', {
                team: state.lastTrickTeam === 0 ? t('team.a') : t('team.b'),
                points: state.lastTrickPoints,
                bonus: state.lastTrickBonus,
              })}
            </div>
          )}
        </div>

        {/* Right: info sidebar */}
        <div data-tutorial="klaverjas-info">
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            <div className="flex flex-wrap gap-x-3">
              <span>{t('teamScore', { team: t('team.a'), score: state.teamScores[0] })}</span>
              <span>{t('teamScore', { team: t('team.b'), score: state.teamScores[1] })}</span>
            </div>
            <details open={!isMobile || undefined} data-testid="klaverjas-match-details">
              <summary className="cursor-pointer select-none text-xs">{t('matchDetails')}</summary>
              <div className="mt-1 text-xs">
                <div>{t('targetPoints', { target: state.config.targetPoints })}</div>
                <div>
                  {t('yourTeam')}: {humanTeam === 0 ? t('team.a') : t('team.b')}
                </div>
                {!(isRoundEnd || isGameEnd) && (
                  <div
                    className="mt-1 flex flex-wrap gap-x-2"
                    role="status"
                    aria-live="polite"
                    data-testid="klaverjas-round-card-points"
                    aria-atomic="true"
                  >
                    <span>{t('roundResult.teamAShort', { points: state.roundCardPoints[0] })}</span>
                    <span>{t('roundResult.teamBShort', { points: state.roundCardPoints[1] })}</span>
                  </div>
                )}

                {!(isRoundEnd || isGameEnd) && (
                  <div
                    className={`mt-1 p-1 rounded bg-black/30${roemPulse ? ' motion-safe:animate-pulse ring-1 ring-ds-warning' : ''}`}
                    data-testid="klaverjas-roem"
                    role="status"
                    aria-live="polite"
                  >
                    <div>
                      {t('roundResult.roemShort', {
                        roemA: state.roundRoem[0] ?? 0,
                        roemB: state.roundRoem[1] ?? 0,
                      })}
                    </div>
                    {roemBreakdown && <div data-testid="klaverjas-roem-breakdown">{roemBreakdown}</div>}
                  </div>
                )}

                {/* Players: cards / tricks */}
                <div className="mt-1 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  {humanPlayer && (
                    <div className="py-0.5">
                      {playerName(humanPlayer.id, true)}: {t('cards', { count: humanPlayer.cardCount })} |{' '}
                      {t('tricks', { count: humanPlayer.trickCount })}
                    </div>
                  )}
                  <CpuAccordion playerCount={state.players.filter((p) => !p.isHuman).length}>
                    <div className="flex flex-wrap gap-x-3">
                      {state.players
                        .filter((p) => !p.isHuman)
                        .map((p) => (
                          <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                            {playerName(p.id, false)}: {t('cards', { count: p.cardCount })} |{' '}
                            {t('tricks', { count: p.trickCount })}
                          </div>
                        ))}
                    </div>
                  </CpuAccordion>
                </div>

                {/* **切り札と非切り札で強さの順序が完全に違う (#4757)。**切り札は
              J > 9 > A > 10、非切り札は A > 10 > K。姉妹ゲームの Manille には
              早見表があるのに、より覚えにくいこちらには無かった。 */}
                <details className="mt-1" data-testid="klaverjas-strength-legend">
                  <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                    {t('strengthLegend.title')}
                  </summary>
                  <div className="mt-1 text-ds-text-muted text-xs">
                    {(
                      [
                        ['trumpCol', KLAVERJAS_TRUMP_ROWS],
                        ['plainCol', KLAVERJAS_PLAIN_ROWS],
                      ] as const
                    ).map(([headKey, rows]) => (
                      <table className="w-full mb-1" key={headKey}>
                        <caption className="text-left">{t(`strengthLegend.${headKey}`)}</caption>
                        <thead>
                          <tr>
                            <th scope="col" className="text-left font-normal">
                              {t('strengthLegend.rankCol')}
                            </th>
                            <th scope="col" className="text-left font-normal">
                              {t('strengthLegend.cardCol')}
                            </th>
                            <th scope="col" className="text-right font-normal">
                              {t('strengthLegend.pointCol')}
                            </th>
                          </tr>
                        </thead>
                        <tbody>
                          {rows.map((row, i) => (
                            <tr key={row.face}>
                              <td>{i + 1}</td>
                              <td>{'nameKey' in row && row.nameKey ? `${row.face} (${t(row.nameKey)})` : row.face}</td>
                              <td className="text-right">{row.points}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    ))}
                    <div className="mt-1">{t('strengthLegend.note')}</div>
                  </div>
                </details>
              </div>
            </details>
          </div>

          {/* Round result */}
          {(isRoundEnd || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              <div>{t('roundResult.teamA', { points: state.roundCardPoints[0] ?? 0 })}</div>
              <div>{t('roundResult.teamB', { points: state.roundCardPoints[1] ?? 0 })}</div>
              <div className="mt-1">
                {t('roundResult.roem', { roemA: state.roundRoem[0] ?? 0, roemB: state.roundRoem[1] ?? 0 })}
              </div>
              {/* **内訳は結果でこそ要る。**進行中のパネルにだけ出すと、点を突き合わせ
                  たい瞬間に消える。`roundPlayerRoem` は `startRound` でしか
                  リセットされないので、ここでもまだ有効 (#6441 レビュー指摘)。 */}
              {roemBreakdown && <div data-testid="klaverjas-roem-breakdown-result">{roemBreakdown}</div>}
            </div>
          )}
        </div>
      </div>

      {/* Message */}
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterHand({ state, game, t, view }: KlaverjasContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (!humanPlayer) return null;
  const isPlayPhase = state.phase === KlaverjasPhase.PLAY;
  const isHumanTurn = state.isHumanTurn;
  const canPlay = isPlayPhase && isHumanTurn;
  return (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="klaverjas"
      validIndices={canPlay ? state.playableIndices : undefined}
      restrictedTooltip={t('playButton')}
      cardTitleFor={
        canPlay
          ? (idx) => {
              if (state.playableIndices.includes(idx)) return undefined;
              const code = state.playRejectCodes[idx];
              const reason = code ? playRejectReasonKeys[code] : undefined;
              return reason ? t(`playRestriction.${reason}`) : undefined;
            }
          : undefined
      }
    />
  );
}

function FooterStatus({ state, t, frontendHint }: KlaverjasContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="klaverjas-hint-live" role="status" aria-live="polite">
        {state.hint && isRequestedHint(state) && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
            {state.hint.cardIndices &&
              state.hint.cardIndices.length > 0 &&
              ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterActions({ state, game, t, busy }: KlaverjasContext) {
  const isPlayPhase = state.phase === KlaverjasPhase.PLAY;
  const isHumanTurn = state.isHumanTurn;
  const canPlay = isPlayPhase && isHumanTurn;
  const isTrickEnd = state.phase === KlaverjasPhase.TRICK_END;
  const isRoundEnd = state.phase === KlaverjasPhase.ROUND_END;

  return (
    <>
      {canPlay && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handlePlay}
          disabled={busy || game.selectedCardIndices.length !== 1}
        >
          {t('playButton')}
        </button>
      )}
      {isTrickEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextTrick} disabled={busy}>
          {t('nextTrick')}
        </button>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={busy}>
          {t('nextRound')}
        </button>
      )}
    </>
  );
}

const klaverjasDefinition: GameDefinition<
  KlaverjasResponse,
  KlaverjasGame,
  Parameters<typeof klaverjasApi.exec>,
  KlaverjasView
> = {
  key: 'klaverjas',
  useGame: useKlaverjasGame,
  useView: useKlaverjasView,
  tutorial: KLAVERJAS_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 8 },
  cli: () => ({
    parseCommand: parseKlaverjasCommand,
    formatResponse: formatKlaverjasState,
    helpText: KLAVERJAS_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === KlaverjasPhase.GAME_END || state.gameEndFlag;
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    const humanTeam = humanIdx % 2;
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.winnerTeam === humanTeam,
    };
  },
  shell: ({ state }) => ({
    isHumanTurn: state.isHumanTurn,
  }),
  reset: ({ game }) => {
    game.reset();
  },
  settingsPlacement: 'top',
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.klaverjasConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetPoints',
          label: t('settings.targetPoints'),
          value: game.klaverjasConfig.targetPoints,
          options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetPoints', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  resetTutorialId: 'klaverjas-reset-button',
  actionsTutorialId: 'klaverjas-action-buttons',
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Klaverjas game page: a Dutch 4-player (2 vs 2) trump trick-taker with Roem bonuses. */
export const KlaverjasPage = createGamePage(klaverjasDefinition);
