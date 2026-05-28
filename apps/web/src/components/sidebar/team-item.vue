<template>
  <SidebarMenuButton
    :tooltip="team.name || team.id"
    as-child
  >
    <button
      :class="[
        'group/team flex items-center gap-2.5 w-full h-9.5 px-2.5 rounded-lg transition-colors',
        'group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:px-0',
        isActive
          ? 'bg-sidebar-accent'
          : 'hover:bg-sidebar-accent/60',
      ]"
      @click="handleSelect"
    >
      <div class="size-6.5 shrink-0 rounded-full border border-border bg-accent overflow-hidden p-px group-data-[collapsible=icon]:mx-auto">
        <img
          v-if="team.avatar_url && !imageError"
          :src="team.avatar_url"
          :alt="team.name || team.id"
          class="size-full rounded-full object-cover"
          @error="imageError = true"
          @load="imageError = false"
        >
        <span
          v-else
          class="size-full flex items-center justify-center text-[8px] font-medium text-muted-foreground"
        >
          <Users
            v-if="!avatarFallback"
            class="size-3"
          />
          <template v-else>
            {{ avatarFallback }}
          </template>
        </span>
      </div>
      <span class="truncate text-xs font-medium text-foreground leading-4.5 flex-1 text-left group-data-[collapsible=icon]:hidden">
        {{ team.name || team.id }}
      </span>

      <div class="group-data-[collapsible=icon]:hidden">
        <DropdownMenu>
          <DropdownMenuTrigger
            as-child
            @click.stop
          >
            <span
              class="shrink-0 size-6 flex items-center justify-center rounded text-muted-foreground opacity-0 group-hover/team:opacity-100 hover:text-foreground hover:bg-accent transition-opacity"
            >
              <Ellipsis class="size-3" />
            </span>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            align="start"
            side="bottom"
            @click.stop
          >
            <DropdownMenuItem @click.stop="handleTogglePin">
              <Pin class="size-3 mr-2" />
              {{ pinned ? $t('common.unpin') : $t('common.pin') }}
            </DropdownMenuItem>
            <DropdownMenuItem @click.stop="handleDetails">
              <Settings class="size-3 mr-2" />
              {{ $t('common.details') }}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </button>
  </SidebarMenuButton>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { HandlersTeamResponse } from '@memohai/sdk'
import { Ellipsis, Pin, Settings, Users } from 'lucide-vue-next'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  SidebarMenuButton,
} from '@memohai/ui'
import { useAvatarInitials } from '@/composables/useAvatarInitials'
import { usePinnedTeams } from '@/composables/usePinnedTeams'

const props = defineProps<{ team: HandlersTeamResponse }>()

const router = useRouter()
const route = useRoute()
const { isPinned, togglePin } = usePinnedTeams()

const displayName = computed(() => props.team.name || props.team.id || '')
const avatarFallback = useAvatarInitials(() => displayName.value, '')
const pinned = computed(() => isPinned(props.team.id ?? ''))
const imageError = ref(false)

const isActive = computed(() => {
  const teamId = props.team.id
  if (!teamId) return false
  return route.path.startsWith(`/teams/${teamId}`)
})

function handleSelect() {
  if (!props.team.id) return
  router.push({ name: 'team-workspace', params: { teamId: props.team.id } })
}

function handleDetails() {
  if (!props.team.id) return
  router.push({ name: 'team-detail', params: { teamId: props.team.id } })
}

function handleTogglePin() {
  togglePin(props.team.id ?? '')
}
</script>
