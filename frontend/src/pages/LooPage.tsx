import { useEffect } from 'react';
import type { looApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { ANTE_OPTIONS, CPU_DIFFICULTY_OPTIONS, useLooGame } from '../hooks/useLooGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { LooResponse } from '../types/card';
import { LooPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt, suitSymbolAt } from '../utils/cardAlt';
import { LOO_HELP, parseLooCommand } from '../utils/cli/commands/looCommands';
import { formatLooState } from '../utils/cli/formatters/looFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { computeLooPotRisk } from '../utils/looPotRisk';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Loo tutorial step definitions. */
const LOO_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="loo-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="loo-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="loo-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="loo-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="loo-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const LOO_PHASE_KEYS: Readonly<Record<number, string>> = {
  [LooPhase.DECIDE]: 'decide',
  [LooPhase.PLAY]: 'play',
  [LooPhase.TRICK_END]: 'trickEnd',
  [LooPhase.ROUND_END]: 'roundEnd',
};

const SUIT_I18N_KEYS: Readonly<Record<string, string>> = {
  SPADE: 'common.suit.spade',
  CLOVER: 'common.suit.club',
  HEART: 'common.suit.heart',
  DIAMOND: 'common.suit.diamond',
};

type LooGame = ReturnType<typeof useLooGame>;
type LooView = ReturnType<typeof useLooView>;
type LooContext = GamePageContext<LooResponse, LooGame, LooView>;

function useLooView({ game }: { game: LooGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('loo', LOO_PHASE_KEYS);

  return { cardWidth, isMobile, phaseNames };
}

function Board({ state, t, tc, view }: LooContext) {
  const isDecidePhase = state.phase === LooPhase.DECIDE;
  const isRoundEnd = state.phase === LooPhase.ROUND_END;

  const decideStartIdx = (state.dealerIdx + 1) % state.players.length;
  const decideCursor = (state.decidePlayerIdx - decideStartIdx + state.players.length) % state.players.length;
  const decisionCounts = state.players.reduce(
    (counts, player, idx) => {
      const order = (idx - decideStartIdx + state.players.length) % state.players.length;
      if (order < decideCursor) {
        if (player.playing) counts.play++;
        else counts.pass++;
      } else {
        counts.undecided++;
      }
      return counts;
    },
    { play: 0, pass: 0, undecided: 0 },
  );

  const trumpSymbol = state.trumpSuit >= 1 ? suitSymbolAt(state.trumpSuit, '-') : '-';
  const leadSuitKey = state.currentTrick[0]?.card.design
    ? SUIT_I18N_KEYS[state.currentTrick[0].card.design]
    : undefined;

  return (
    <>
      <div data-testid="loo-decision-counts-live" role="status" aria-live="polite">
        {isDecidePhase && (
          <div
            className="mb-2 mx-auto max-w-md text-center text-sm text-ds-text-muted"
            data-testid="loo-decision-counts"
          >
            {t('decisionCounts', decisionCounts)}
          </div>
        )}
      </div>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('deal', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber, total: state.totalTricks })}</span>
        <span className="mr-4">{t('pot', { pot: state.pot })}</span>
        <span>{t('trump', { suit: trumpSymbol })}</span>
        {state.turnUp && (
          <span className="ml-4 inline-flex items-center gap-2">
            <span>{t('turnUp')}</span>
            <CardImage card={state.turnUp} width={32} />
            <span aria-hidden="true">{cardAlt(state.turnUp)}</span>
          </span>
        )}
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={view.cardWidth}
            label={t('currentTrick')}
            leadSuit={leadSuitKey ? tc(leadSuitKey) : undefined}
            leadSuitLabel={t('leadSuit')}
            dataTutorial="loo-trick-display"
          />
        </div>

        {/* Right: info sidebar */}
        <div data-tutorial="loo-info">
          {/* Per-player chip balances with a participation badge */}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            {state.players.map((p) => (
              <div key={p.id} className="py-0.5 flex items-center gap-2">
                {/* ルーの罰金には下限が無いのでチップは負に落ちる。色だけの
                    警告にならないよう ▼ と aria-label も添える。 */}
                <span
                  className={p.chips < 0 ? 'text-ds-error-text font-semibold' : undefined}
                  data-testid={`loo-chips-${p.id}`}
                >
                  {p.chips < 0 && (
                    // role="img" + aria-label なので、色を見ない読者にも「赤字」が届く
                    // (素の span は aria-label を取れない)。
                    <span role="img" aria-label={t('chipsNegativeAria')}>
                      ▼{' '}
                    </span>
                  )}
                  {playerName(p.id, p.isHuman)}
                  {p.id === state.dealerIdx && (
                    <span
                      className="ml-1 rounded px-1.5 py-0.5 text-xs bg-ds-accent/30 text-ds-accent"
                      role="img"
                      aria-label={t('dealerBadge')}
                      data-testid={`loo-dealer-${p.id}`}
                    >
                      {t('dealerBadge')}
                    </span>
                  )}
                  : {t('chips', { chips: p.chips })}
                </span>
                <span
                  className={`px-1.5 py-0.5 rounded text-xs ${
                    p.playing ? 'bg-ds-accent/30 text-ds-accent' : 'bg-black/40 text-ds-text-muted'
                  }`}
                  role="status"
                  aria-label={t('statusAria', {
                    name: playerName(p.id, p.isHuman),
                    status: p.playing ? t('statusPlay') : t('statusPass'),
                  })}
                  title={t('playRiskTooltip')}
                  data-testid={`loo-status-${p.id}`}
                >
                  {/* Colour-independent icon (● play / ○ pass) alongside the label. */}
                  <span aria-hidden="true">{p.playing ? '● ' : '○ '}</span>
                  {p.playing ? t('statusPlay') : t('statusPass')}
                </span>
              </div>
            ))}
          </div>

          {/* Players: cards / tricks */}
          {view.isMobile ? (
            <details className="mb-2 p-2 rounded bg-black/30">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
              <div className="mt-1">
                {state.players.map((p) => (
                  <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                    {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                    {t('tricks', { count: p.trickCount })}
                  </div>
                ))}
              </div>
            </details>
          ) : (
            <div className="mb-2 p-2 rounded bg-black/30">
              {state.players.map((p) => (
                <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                  {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                  {t('tricks', { count: p.trickCount })}
                </div>
              ))}
            </div>
          )}

          {/* Deal result: per-player gained chips + looed players */}
          {isRoundEnd && state.lastDealDetail && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="loo-deal-result">
              <div className="mb-1 text-ds-text-primary">{t('dealResult.title')}</div>
              {state.lastDealDetail.looed.length > 0 && (
                <div className="text-ds-error-text mb-1">
                  {t('dealResult.looed', {
                    names: state.lastDealDetail.looed
                      .map((i) => playerName(i, state.players[i]?.isHuman ?? false))
                      .join(', '),
                  })}
                </div>
              )}
              {state.players.map((p) => (
                <div key={p.id}>
                  {t('dealResult.gained', {
                    name: playerName(p.id, p.isHuman),
                    chips: state.lastDealDetail?.gained[p.id] ?? 0,
                  })}
                </div>
              ))}
              {/* 誰も Loo されなければポットは次ディールへ繰り越される。
                  次の pot が ante 分より大きくなる理由がこれ (#6489)。
                  0 のときは出さない ── looed と同じ扱い。 */}
              {state.lastDealDetail.potCarry > 0 && (
                <div className="mt-1 text-ds-warning" data-testid="loo-pot-carry">
                  {t('dealResult.potCarry', { chips: state.lastDealDetail.potCarry })}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Message */}
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterLead({ state, t }: LooContext) {
  const isDecidePhase = state.phase === LooPhase.DECIDE;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const canDecide = isDecidePhase && state.decidePlayerIdx === humanIdx && state.isHumanTurn;
  const { looPenalty, perTrick, maxWin } = computeLooPotRisk(state.pot, state.potStart);

  return (
    <div data-testid="loo-pot-risk-live" role="status" aria-live="polite" className="sr-only">
      {canDecide && (
        <div data-testid="loo-pot-risk">
          {t('potRisk.label')} {t('potRisk.win', { pot: maxWin, perTrick })}{' '}
          {t('potRisk.loss', { penalty: looPenalty })}
        </div>
      )}
    </div>
  );
}

function FooterStatus({ state, t, frontendHint }: LooContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="loo-hint-live" role="status" aria-live="polite">
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

function primaryActions({ game, state, t, busy }: LooContext) {
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;
  const isDecidePhase = state.phase === LooPhase.DECIDE;
  const isPlayPhase = state.phase === LooPhase.PLAY;
  const isRoundEnd = state.phase === LooPhase.ROUND_END;

  const canDecide = isDecidePhase && state.decidePlayerIdx === humanIdx && isHumanTurn;
  const canPlay = isPlayPhase && isHumanTurn;

  if (!canDecide && !canPlay && !isRoundEnd) return null;

  const { looPenalty, perTrick, maxWin } = computeLooPotRisk(state.pot, state.potStart);

  return (
    <>
      {canDecide && (
        <div className="mb-2 mx-auto max-w-md p-2 rounded bg-black/30 text-center text-sm">
          <div className="text-ds-text-muted mb-0.5">{t('potRisk.label')}</div>
          <div className="text-ds-accent">{t('potRisk.win', { pot: maxWin, perTrick })}</div>
          <div className="text-ds-error-text">{t('potRisk.loss', { penalty: looPenalty })}</div>
        </div>
      )}
      <div className="flex flex-wrap gap-2 items-center" data-tutorial="loo-action-buttons">
        {canDecide && (
          <div className="flex flex-wrap gap-2" data-testid="loo-decide-buttons">
            <button type="button" className={btnPrimary} onClick={() => game.handleDecide(true)} disabled={busy}>
              {t('decidePlay')}
            </button>
            <button type="button" className={btnSecondary} onClick={() => game.handleDecide(false)} disabled={busy}>
              {t('decidePass')}
            </button>
          </div>
        )}
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
        {isRoundEnd && (
          <button type="button" className={btnSuccess} onClick={game.handleNextDeal} disabled={busy}>
            {t('nextDeal')}
          </button>
        )}
      </div>
    </>
  );
}

function FooterHand({ game, state, t, view }: LooContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;
  const isDecidePhase = state.phase === LooPhase.DECIDE;
  const isPlayPhase = state.phase === LooPhase.PLAY;

  const canDecide = isDecidePhase && state.decidePlayerIdx === humanIdx && isHumanTurn;
  const canPlay = isPlayPhase && isHumanTurn;

  return (
    <>
      {isDecidePhase && !canDecide && (
        <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="loo-decide-cpu">
          {t('decideCpu', { id: state.decidePlayerIdx })}
        </div>
      )}
      {/* 領域は**常設**。中身だけ差し替える ── 出現と同時に付けた領域は
          変化として扱われず読み上げられない (#5955)。CalabresellaPage と同じ形 (#6880)。 */}
      <div data-testid="loo-prompt-live" role="status" aria-live="polite">
        {canDecide && (
          <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="loo-decide-prompt">
            {t('decidePrompt')}
          </div>
        )}
      </div>
      {humanPlayer && (
        <PlayerHandSection
          humanPlayer={humanPlayer}
          selectedCardIndices={game.selectedCardIndices}
          toggleCard={game.toggleCard}
          cardWidth={view.cardWidth}
          isMobile={view.isMobile}
          stackRows
          dataTutorialPrefix="loo"
          validIndices={canPlay ? state.playableIndices : undefined}
          restrictedTooltip={t('playButton')}
        />
      )}
    </>
  );
}

const looDefinition: GameDefinition<LooResponse, LooGame, Parameters<typeof looApi.exec>, LooView> = {
  key: 'loo',
  useGame: useLooGame,
  useView: useLooView,
  tutorial: LOO_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 5 },
  cli: () => ({
    parseCommand: parseLooCommand,
    formatResponse: formatLooState,
    helpText: LOO_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: false,
  }),
  // The hand-wired page offered the action log at the end of each deal. Loo has no
  // game end (ended stays false), so the log follows the round end instead.
  actionLogEndPhase: ({ state }) => state.phase === LooPhase.ROUND_END,
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
          value: game.looConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'ante',
          label: t('settings.ante'),
          value: game.looConfig.ante,
          options: ANTE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('ante', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  resetTutorialId: 'loo-reset-button',
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  Board,
  FooterHand,
  FooterLead,
  FooterStatus,
};

/** Renders the Loo (Lanterloo) game page: a 4-player 52-card pot-based gambling trick-taker. */
export const LooPage = createGamePage(looDefinition);
