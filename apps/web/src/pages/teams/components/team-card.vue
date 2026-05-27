<template>
  <Card
    class="group relative transition-shadow h-full flex flex-col hover:shadow-md cursor-pointer"
    role="button"
    tabindex="0"
    :aria-label="`Open team ${(team.name || team.id)}`"
    @click="onOpenDetail"
    @keydown.enter.prevent="onOpenDetail"
    @keydown.space.prevent="onOpenDetail"
  >
    <CardHeader class="flex flex-row items-start gap-3 space-y-0 pb-2">
      <Avatar class="size-11 shrink-0">
        <AvatarImage
          v-if="team.avatar_url"
          :src="team.avatar_url"
          :alt="team.name"
        />
        <AvatarFallback class="text-sm">
          <Users
            v-if="!avatarFallback"
            class="size-4"
          />
          <template v-else>
            {{ avatarFallback }}
          </template>
        </AvatarFallback>
      </Avatar>
      <div class="flex-1 min-w-0 flex flex-col gap-1.5">
        <div class="flex items-center justify-between gap-2">
          <CardTitle class="text-sm truncate">
            {{ team.name || team.id }}
          </CardTitle>
          <Badge
            variant="outline"
            class="shrink-0 text-xs"
          >
            {{ t('sidebar.teams') }}
          </Badge>
        </div>
        <div class="flex flex-wrap items-center gap-x-2 gap-y-0.5 text-xs text-muted-foreground">
          <span v-if="formattedDate">
            {{ t('common.createdAt') }} {{ formattedDate }}
          </span>
          <span
            v-if="team.shared_dir_name"
            class="truncate font-mono"
          >
            /team/{{ team.shared_dir_name }}
          </span>
        </div>
      </div>
    </CardHeader>
  </Card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
  Badge,
  Card,
  CardHeader,
  CardTitle,
} from '@memohai/ui'
import { Users } from 'lucide-vue-next'
import type { HandlersTeamResponse } from '@memohai/sdk'
import { formatDate } from '@/utils/date-time'
import { useAvatarInitials } from '@/composables/useAvatarInitials'

const router = useRouter()
const { t } = useI18n()

const props = defineProps<{
  team: HandlersTeamResponse
}>()

const avatarFallback = useAvatarInitials(() => props.team.name || props.team.id)

const formattedDate = computed(() => {
  if (!props.team.created_at) return ''
  return formatDate(props.team.created_at)
})

function onOpenDetail() {
  if (!props.team.id) return
  router.push({ name: 'team-detail', params: { teamId: props.team.id } })
}
</script>
