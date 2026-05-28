import { useStorage } from '@vueuse/core'
import { onAuthSessionCleared } from '@/lib/auth-session'

const pinnedTeamIds = useStorage<string[]>('pinned-team-ids', [])

onAuthSessionCleared(() => {
  pinnedTeamIds.value = []
})

export function usePinnedTeams() {
  function isPinned(teamId: string) {
    return pinnedTeamIds.value.includes(teamId)
  }

  function togglePin(teamId: string) {
    const idx = pinnedTeamIds.value.indexOf(teamId)
    if (idx >= 0) {
      pinnedTeamIds.value.splice(idx, 1)
    }
    else {
      pinnedTeamIds.value.push(teamId)
    }
  }

  function sortTeams<T extends { id?: string }>(teams: T[]): T[] {
    return [...teams].sort((a, b) => {
      const aPinned = isPinned(a.id ?? '')
      const bPinned = isPinned(b.id ?? '')
      if (aPinned && !bPinned) return -1
      if (!aPinned && bPinned) return 1
      return 0
    })
  }

  return { pinnedTeamIds, isPinned, togglePin, sortTeams }
}
