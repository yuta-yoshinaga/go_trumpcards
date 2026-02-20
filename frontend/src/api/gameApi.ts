import type { BlackJackResponse, PokerResponse, OldMaidResponse, DaifugoResponse, SevensResponse } from '../types/card'

export const sessionId: string = crypto.randomUUID()

async function postJson<T>(url: string, body: unknown): Promise<T> {
  const res = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`HTTP error: ${res.status}`)
  return res.json() as Promise<T>
}

export const blackjackApi = {
  exec: (command: 'reset' | 'hit' | 'stand') =>
    postJson<BlackJackResponse>('/blackjac/exec', { command, sessionId }),
}

export const pokerApi = {
  exec: (command: 'reset' | 'exchange' | 'stand', indices?: number[]) =>
    postJson<PokerResponse>('/poker/exec', { command, indices, sessionId }),
}

export const oldmaidApi = {
  exec: (command: 'reset' | 'draw', drawIdx?: number) =>
    postJson<OldMaidResponse>('/oldmaid/exec', { command, drawIdx, sessionId }),
}

export const daifugoApi = {
  exec: (command: 'reset' | 'play', indices?: number[]) =>
    postJson<DaifugoResponse>('/daifugo/exec', { command, indices, sessionId }),
}

export const sevensApi = {
  exec: (command: 'reset' | 'play', index = -1) =>
    postJson<SevensResponse>('/sevens/exec', { command, index, sessionId }),
}
