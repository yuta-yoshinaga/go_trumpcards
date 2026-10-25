import { useEffect, useState } from 'react';
import type { musApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { CPU_DIFFICULTY_OPTIONS, TARGET_AMARRAKOS_OPTIONS, useMusGame } from '../hooks/useMusGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import type { MusResponse } from '../types/card';
import { MusBetAction, MusPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { MUS_HELP, parseMusCommand } from '../utils/cli/commands/musCommands';
import { formatMusState } from '../utils/cli/formatters/musFormatter';
import { evalMusHand } from '../utils/musHandEval';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** i18n value keys for each Pares category (0=none, 1=par, 2=medias, 3=duples). */
const PARES_VALUE_KEYS: readonly string[] = ['paresNone', 'paresPar', 'paresMedias', 'paresDuples'];

/** Mus tutorial step definitions. */
const MUS_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="mus-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="mus-amarrakos"]',
    messageKey: 'tutorial.amarrakos',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="mus-player-hand"]', messageKey: 'tutorial.mus', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="mus-action-buttons"]', messageKey: 'tutorial.bet', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="mus-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const MUS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [MusPhase.MUS]: 'mus',
  [MusPhase.DISCARD]: 'discard',
  [MusPhase.GRANDE]: 'grande',
  [MusPhase.CHICA]: 'chica',
  [MusPhase.PARES]: 'pares',
  [MusPhase.JUEGO]: 'juego',
  [MusPhase.SHOWDOWN]: 'showdown',
  [MusPhase.ROUND_END]: 'roundEnd',
  [MusPhase.GAME_END]: 'gameEnd',
};

/** Maps a betting-round phase to its i18n round-name key. */
const ROUND_NAME_KEYS: Readonly<Record<number, string>> = {
  [MusPhase.GRANDE]: 'grande',
  [MusPhase.CHICA]: 'chica',
  [MusPhase.PARES]: 'pares',
  [MusPhase.JUEGO]: 'juego',
};

/** Maps a round index (0..3) in `results` to its i18n round-name key. */
const ROUND_INDEX_KEYS: Readonly<Record<number, string>> = {
  0: 'grande',
  1: 'chica',
  2: 'pares',
  3: 'juego',
};

/** Maps a result kind to its localized settlement method. */
const RESULT_KIND_KEYS: Readonly<Record<number, string>> = {
  1: 'deferred',
  2: 'accepted',
  3: 'awarded',
  4: 'ordago',
};

type MusGame = ReturnType<typeof useMusGame>;
type MusView = ReturnType<typeof useMusView>;
type MusContext = GamePageContext<MusResponse, MusGame, MusView>;

function useMusView({ game }: { game: MusGame }) {
  // Envido stepper amount (local UI state).
  const [envidoAmount, setEnvidoAmount] = useState(2);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('mus', MUS_PHASE_KEYS);

  return { envidoAmount, setEnvidoAmount, cardWidth, isMobile, phaseNames };
}

function Board({ state, t }: MusContext) {
  const isRoundEnd = state.phase === MusPhase.ROUND_END || state.phase === MusPhase.SHOWDOWN;
  const isGameEnd = state.phase === MusPhase.GAME_END || state.gameEndFlag;
  const roundNameKey = ROUND_NAME_KEYS[state.phase];

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2" data-tutorial="mus-info">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        {roundNameKey && <span>{t(`phase.${roundNameKey}`)}</span>}
      </div>

      {/* Amarrakos (team scores) */}
      <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-tutorial="mus-amarrakos">
        <div className="mb-1 text-ds-text-primary">{t('amarrakos')}</div>
        {state.amarrakos.map((score, team) => (
          <div key={`team-${team}`} className={team === state.humanTeam ? 'text-ds-text-primary' : ''}>
            {t('amarrakosProgress', {
              team,
              score,
              target: state.config.targetAmarrakos,
              remaining: Math.max(0, state.config.targetAmarrakos - score),
            })}
            {team === state.humanTeam ? ` (${t('yourTeam')})` : ''}
          </div>
        ))}
        {/* **賭け金は CPU の側でも動く。**envido / ordago で吊り上げられた瞬間を
            能動的に知る手立てが無かった (#6436)。ライブ領域は**常設**にし、
            中身だけを差し替える ── 領域と中身が同じコミットで DOM に入ると、
            変化として扱われず読み上げられないことがある。 */}
        <div className="mt-1 text-ds-warning" role="status" aria-live="polite" data-testid="mus-pending-stake">
          {state.pendingStake !== 0 && (
            <>
              <span>
                {t('pendingStake', { amount: state.pendingStake === -1 ? t('ordagoLabel') : state.pendingStake })}
              </span>
              <span className="block">
                {t('betTeamLabel', { team: state.betTeam })}
                {state.betTeam === state.humanTeam ? ` (${t('yourTeam')})` : ''}
              </span>
              <span className="block">
                {t('lastBettorTeamLabel', { team: state.lastBettorTeam })}
                {state.lastBettorTeam === state.humanTeam ? ` (${t('yourTeam')})` : ''}
              </span>
            </>
          )}
        </div>
      </div>

      {/* Players */}
      <div className="mb-2 p-2 rounded bg-black/30">
        {state.players.map((p) => {
          // This rule pairs with Go's domain.MusTeamOf: player index modulo the team count.
          const teamCount = state.amarrakos.length;
          const playerTeam = teamCount === 0 ? 0 : p.id % teamCount;
          const isTeammate = playerTeam === state.humanTeam;
          return (
            <div
              key={p.id}
              className={`text-sm py-0.5 ${!p.isHuman && isTeammate ? 'text-ds-accent' : 'text-ds-text-muted'}`}
              data-testid={`mus-player-${p.id}`}
            >
              {p.isHuman
                ? `${playerName(p.id, p.isHuman)}: ${t('cards', { count: p.cardCount })}`
                : t('playerLine', {
                    name: playerName(p.id, p.isHuman),
                    team: playerTeam,
                    role: t(isTeammate ? 'partner' : 'opponent'),
                    cards: p.cardCount,
                  })}
              {p.id === state.manoIdx && <span className="ml-2 text-ds-accent">{t('mano')}</span>}
            </div>
          );
        })}
      </div>

      {/* Round (betting) results */}
      {(isRoundEnd || isGameEnd) && state.results.some((r) => r.team >= 0) && (
        <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
          <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
          {state.results.map((r, i) =>
            r.team >= 0 ? (
              <div key={`result-${i}`}>
                {t('roundResult.line', {
                  round: t(`roundResult.${ROUND_INDEX_KEYS[i] ?? i}`),
                  method: t(`roundResult.${RESULT_KIND_KEYS[r.kind]}`),
                  team: r.team,
                  stake: r.stake,
                })}
              </div>
            ) : null,
          )}
        </div>
      )}

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterLead({ state, t }: MusContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (humanPlayer?.cards.length !== 4) return null;

  const evalResult = evalMusHand(humanPlayer.cards);
  const paresValue = t(`handSummary.${PARES_VALUE_KEYS[evalResult.paresCategory]}`);
  const juegoValue =
    evalResult.points === 31
      ? t('handSummary.juegoBest')
      : evalResult.hasJuego
        ? t('handSummary.juegoYes', { points: evalResult.points })
        : t('handSummary.juegoPunto', { points: evalResult.points });
  const paresActive = state.phase === MusPhase.PARES;
  const juegoActive = state.phase === MusPhase.JUEGO;

  return (
    <div className="mt-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="mus-hand-summary">
      <div className="mb-1 text-ds-text-primary">{t('handSummary.title')}</div>
      {/* Grande and Chica are two of the four betting rounds, and the
          panel said nothing during either of them (#4721). */}
      <div
        className={state.phase === MusPhase.GRANDE ? 'text-ds-warning font-semibold' : ''}
        data-testid="mus-summary-grande"
      >
        {/* The panel only renders on a full four-card hand, so the ranks
            are never null here. */}
        {t('handSummary.grande')}: {t('handSummary.rankValue', { rank: evalResult.highestRank })}
      </div>
      <div
        className={state.phase === MusPhase.CHICA ? 'text-ds-warning font-semibold' : ''}
        data-testid="mus-summary-chica"
      >
        {t('handSummary.chica')}: {t('handSummary.rankValue', { rank: evalResult.lowestRank })}
      </div>
      <div className={paresActive ? 'text-ds-warning font-semibold' : ''} data-testid="mus-summary-pares">
        {t('handSummary.pares')}: {paresValue}
      </div>
      <div className={juegoActive ? 'text-ds-warning font-semibold' : ''} data-testid="mus-summary-juego">
        {t('handSummary.juego')}: {juegoValue}
      </div>
    </div>
  );
}

function FooterStatus({ frontendHint, t }: MusContext) {
  return <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />;
}

function FooterHand({ game, state, view }: MusContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (!humanPlayer) return null;

  return (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="mus"
    />
  );
}

function primaryActions({ game, state, t, busy, view }: MusContext) {
  const loading = busy;
  const isHumanTurn = state.isHumanTurn;
  const isMusPhase = state.phase === MusPhase.MUS;
  const isDiscardPhase = state.phase === MusPhase.DISCARD;
  const isBetPhase =
    state.phase === MusPhase.GRANDE ||
    state.phase === MusPhase.CHICA ||
    state.phase === MusPhase.PARES ||
    state.phase === MusPhase.JUEGO;
  const isRoundEnd = state.phase === MusPhase.ROUND_END || state.phase === MusPhase.SHOWDOWN;
  const isGameEnd = state.phase === MusPhase.GAME_END || state.gameEndFlag;

  if (
    !(
      (isMusPhase && isHumanTurn) ||
      (isDiscardPhase && isHumanTurn) ||
      (isBetPhase && isHumanTurn) ||
      (isRoundEnd && !isGameEnd)
    )
  ) {
    return null;
  }

  return (
    <div className="flex flex-wrap gap-2 items-center" data-tutorial="mus-action-buttons">
      {isMusPhase && isHumanTurn && (
        <>
          <button type="button" className={btnPrimary} onClick={() => game.handleMus(true)} disabled={loading}>
            {t('musButton')}
          </button>
          <button type="button" className={btnSecondary} onClick={() => game.handleMus(false)} disabled={loading}>
            {t('corteButton')}
          </button>
        </>
      )}

      {isDiscardPhase && isHumanTurn && (
        <>
          <button
            type="button"
            className={btnPrimary}
            onClick={game.handleDiscard}
            disabled={loading || game.selectedCardIndices.length === 0}
          >
            {t('discardButton')} ({t('discardSelected', { count: game.selectedCardIndices.length })})
          </button>
          <span className="text-ds-text-muted text-sm" data-testid="mus-discard-guide">
            {game.selectedCardIndices.length === 0
              ? t('discardGuide')
              : t('discardCount', { count: game.selectedCardIndices.length })}
          </span>
        </>
      )}

      {isBetPhase && isHumanTurn && (
        <>
          {state.canPaso && (
            <button
              type="button"
              className={btnSecondary}
              onClick={() => game.handleBet(MusBetAction.PASO)}
              disabled={loading}
            >
              {t('bet.paso')}
            </button>
          )}
          {state.canEnvido && (
            <fieldset className="flex items-center gap-1 border-0 p-0 m-0">
              <legend className="sr-only">{t('envidoStepperLabel')}</legend>
              <button
                type="button"
                className={btnSecondary}
                onClick={() => view.setEnvidoAmount((a) => Math.max(2, a - 1))}
                disabled={loading}
                aria-label={t('envidoDecrease')}
              >
                −
              </button>
              <span
                className="text-ds-text-primary text-sm min-w-[3rem] text-center"
                aria-live="polite"
                aria-atomic="true"
              >
                {t('envidoAmount', { amount: view.envidoAmount })}
              </span>
              <button
                type="button"
                className={btnSecondary}
                onClick={() => view.setEnvidoAmount((a) => a + 1)}
                disabled={loading}
                aria-label={t('envidoIncrease')}
              >
                ＋
              </button>
              <button
                type="button"
                className={btnPrimary}
                onClick={() => game.handleBet(MusBetAction.ENVIDO, view.envidoAmount)}
                disabled={loading}
              >
                {t('bet.envido')}
              </button>
            </fieldset>
          )}
          {state.canOrdago && (
            <button
              type="button"
              className={btnPrimary}
              onClick={() => game.handleBet(MusBetAction.ORDAGO)}
              disabled={loading}
            >
              {t('bet.ordago')}
            </button>
          )}
          {state.canQuiero && (
            <button
              type="button"
              className={btnSuccess}
              onClick={() => game.handleBet(MusBetAction.QUIERO)}
              disabled={loading}
            >
              {t('bet.quiero')}
            </button>
          )}
          {state.canNoQuiero && (
            <button
              type="button"
              className={btnSecondary}
              onClick={() => game.handleBet(MusBetAction.NO_QUIERO)}
              disabled={loading}
            >
              {t('bet.noquiero')}
            </button>
          )}
        </>
      )}

      {isRoundEnd && !isGameEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={loading}>
          {t('nextRound')}
        </button>
      )}
    </div>
  );
}

const musDefinition: GameDefinition<MusResponse, MusGame, Parameters<typeof musApi.exec>, MusView> = {
  key: 'mus',
  useGame: useMusGame,
  useView: useMusView,
  tutorial: MUS_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 4 },
  cli: () => ({
    parseCommand: parseMusCommand,
    formatResponse: formatMusState,
    helpText: MUS_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === MusPhase.GAME_END || state.gameEndFlag;
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.winnerTeam >= 0 && state.winnerTeam === state.humanTeam,
    };
  },
  shell: ({ state }) => ({
    isHumanTurn: state.isHumanTurn,
  }),
  reset: ({ game }) => {
    game.reset();
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.musConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetAmarrakos',
          label: t('settings.targetAmarrakos'),
          value: game.musConfig.targetAmarrakos,
          options: TARGET_AMARRAKOS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetAmarrakos', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  actionsClassName: 'flex items-center',
  resetTutorialId: 'mus-reset-button',
  Board,
  FooterHand,
  FooterLead,
  FooterStatus,
};

/** Renders the Mus game page: a 4-player Basque vying (betting) game. */
export const MusPage = createGamePage(musDefinition);
