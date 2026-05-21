/// <reference types="vitest/globals" />
import { render, screen } from '@testing-library/react'
import { PlayerList } from './PlayerList'
import type { Player } from '@/types'

const players: Player[] = [
  { id: '1', name: 'Alice', score: 3, isCreator: true },
  { id: '2', name: 'Bob', score: 1, isCreator: false },
]

describe('PlayerList', () => {
  it('renders all player names', () => {
    render(<PlayerList players={players} />)
    expect(screen.getByText('Alice')).toBeInTheDocument()
    expect(screen.getByText('Bob')).toBeInTheDocument()
  })

  it('shows Host badge for creator', () => {
    render(<PlayerList players={players} />)
    expect(screen.getByText('Host')).toBeInTheDocument()
  })

  it('shows You badge for current player', () => {
    render(<PlayerList players={players} currentPlayerId="2" />)
    expect(screen.getByText('You')).toBeInTheDocument()
  })

  it('displays scores', () => {
    render(<PlayerList players={players} />)
    expect(screen.getByText('3 pts')).toBeInTheDocument()
    expect(screen.getByText('1 pts')).toBeInTheDocument()
  })
})
