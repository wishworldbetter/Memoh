<template>
  <div class="max-w-2xl mx-auto pb-6 space-y-5">
    <div class="flex items-start justify-between pb-4 border-b border-border/50">
      <div class="space-y-1">
        <h3 class="text-sm font-semibold text-foreground">
          {{ t('teams.tabs.overview') }}
        </h3>
        <p class="text-[11px] text-muted-foreground">
          {{ t('teams.tabs.overviewSubtitle') }}
        </p>
      </div>
    </div>

    <div class="space-y-4 rounded-md border p-4">
      <div class="space-y-1">
        <h4 class="text-xs font-medium">
          {{ t('teams.teamSummary') }}
        </h4>
        <p class="text-[11px] text-muted-foreground">
          {{ t('teams.settingsDescription') }}
        </p>
      </div>

      <div class="grid gap-3 sm:grid-cols-2">
        <div class="rounded-md border bg-background/70 p-3 flex flex-col justify-between">
          <p class="text-xs text-muted-foreground">
            {{ t('teams.memberCount') }}
          </p>
          <p class="mt-2 text-2xl font-semibold">
            {{ memberCount }}
          </p>
        </div>
        <div class="rounded-md border bg-background/70 p-3 flex flex-col justify-between">
          <p class="text-xs text-muted-foreground">
            {{ t('teams.workspacePath') }}
          </p>
          <p class="mt-2 text-sm font-mono break-all">
            {{ team?.shared_dir_name ? `/team/${team.shared_dir_name}` : '-' }}
          </p>
        </div>
        <div class="rounded-md border bg-background/70 p-3 flex flex-col justify-between">
          <p class="text-xs text-muted-foreground">
            {{ t('teams.createdAt') }}
          </p>
          <p class="mt-2 text-xs">
            {{ formatDateValue(team?.created_at) }}
          </p>
        </div>
        <div class="rounded-md border bg-background/70 p-3 flex flex-col justify-between">
          <p class="text-xs text-muted-foreground">
            {{ t('common.updatedAt') }}
          </p>
          <p class="mt-2 text-xs">
            {{ formatDateValue(team?.updated_at) }}
          </p>
        </div>
        <div class="rounded-md border bg-background/70 p-3 flex flex-col justify-between sm:col-span-2">
          <p class="text-xs text-muted-foreground">
            {{ t('teams.id') }}
          </p>
          <p class="mt-2 break-all font-mono text-xs">
            {{ team?.id || '-' }}
          </p>
        </div>
      </div>
    </div>

    <div
      v-if="team?.description"
      class="space-y-4 rounded-md border p-4"
    >
      <div class="space-y-1">
        <h4 class="text-xs font-medium">
          {{ t('teams.description') }}
        </h4>
      </div>
      <p class="text-xs leading-relaxed text-muted-foreground whitespace-pre-wrap">
        {{ team.description }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { HandlersTeamResponse } from '@memohai/sdk'
import { formatDateTime } from '@/utils/date-time'

const { t } = useI18n()

defineProps<{
  team?: HandlersTeamResponse
  memberCount: number
}>()

function formatDateValue(value?: string): string {
  if (!value) return '-'
  return formatDateTime(value) || '-'
}
</script>
