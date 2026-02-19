import type { BlackJackResponse, PokerResponse, OldMaidResponse, DaifugoResponse } from '../types/card'

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
    postJson<BlackJackResponse>('/blackjac/exec', { command }),
}

export const pokerApi = {
  exec: (command: 'reset' | 'exchange' | 'stand', indices?: number[]) =>
    postJson<PokerResponse>('/poker/exec', { command, indices }),
}

export const oldmaidApi = {
  exec: (command: 'reset' | 'draw', drawIdx?: number) =>
    postJson<OldMaidResponse>('/oldmaid/exec', { command, drawIdx }),
}

export const daifugoApi = {
  exec: (command: 'reset' | 'play', indices?: number[]) =>
    postJson<DaifugoResponse>('/daifugo/exec', { command, indices }),
}
