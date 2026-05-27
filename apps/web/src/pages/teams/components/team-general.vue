<template>
  <div class="max-w-2xl mx-auto pb-6 space-y-5">
    <div class="flex items-start justify-between pb-4 border-b border-border/50">
      <div class="space-y-1">
        <h3 class="text-sm font-semibold text-foreground">
          {{ t('teams.tabs.general') }}
        </h3>
        <p class="text-[11px] text-muted-foreground">
          {{ t('teams.generalSettingsHint') }}
        </p>
      </div>
      <div class="flex items-center gap-3 shrink-0">
        <Transition name="fade">
          <div
            v-if="hasChanges"
            class="flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-muted/40 border border-border/50"
          >
            <div class="size-1 rounded-full bg-muted-foreground/40" />
            <span class="text-[10px] text-muted-foreground font-medium whitespace-nowrap">{{ t('common.unsaved') }}</span>
          </div>
        </Transition>
        <Button
          size="sm"
          :disabled="!hasChanges || saving"
          class="h-8 text-xs font-medium min-w-24 shadow-none"
          @click="handleSave"
        >
          <Spinner
            v-if="saving"
            class="mr-1.5 size-3"
          />
          {{ t('common.save') }}
        </Button>
      </div>
    </div>

    <div class="space-y-4">
      <div class="space-y-4 rounded-md border p-4">
        <div class="space-y-1">
          <h4 class="text-xs font-medium">
            {{ t('teams.generalSettings') }}
          </h4>
          <p class="text-[11px] text-muted-foreground">
            {{ t('teams.teamSettingsHint') }}
          </p>
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <div class="space-y-1.5">
            <Label class="text-xs">{{ t('teams.name') }}</Label>
            <Input
              v-model="form.name"
              required
            />
          </div>
          <div class="space-y-1.5">
            <Label class="text-xs">{{ t('teams.sharedDir') }}</Label>
            <Input
              v-model="form.shared_dir_name"
              :placeholder="t('teams.sharedDirPlaceholder')"
            />
          </div>
        </div>
        <div class="space-y-1.5">
          <Label class="text-xs">{{ t('teams.description') }}</Label>
          <Textarea
            v-model="form.description"
            :placeholder="t('teams.descriptionPlaceholder')"
            rows="3"
            class="text-xs"
          />
        </div>
      </div>

      <div class="space-y-4 rounded-md border border-border bg-background p-4 shadow-none">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div class="space-y-0.5">
            <h4 class="text-xs font-medium text-destructive">
              {{ t('common.dangerZone') }}
            </h4>
            <p class="text-[11px] text-muted-foreground">
              {{ t('teams.deleteTeamHint') }}
            </p>
          </div>
          <div class="flex justify-end shrink-0">
            <ConfirmPopover
              :message="t('teams.deleteTeamConfirm')"
              :loading="deleting"
              :confirm-text="t('common.delete')"
              @confirm="$emit('delete')"
            >
              <template #trigger>
                <Button
                  variant="destructive"
                  size="sm"
                  :disabled="deleting"
                  class="min-w-28 h-8 text-xs font-medium shadow-none"
                >
                  <Spinner
                    v-if="deleting"
                    class="mr-1.5 size-3"
                  />
                  {{ t('teams.deleteTeam') }}
                </Button>
              </template>
            </ConfirmPopover>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Button,
  Input,
  Label,
  Spinner,
  Textarea,
} from '@memohai/ui'
import type { HandlersTeamResponse, HandlersUpdateTeamRequest } from '@memohai/sdk'
import ConfirmPopover from '@/components/confirm-popover/index.vue'

const props = defineProps<{
  team?: HandlersTeamResponse
  saving: boolean
  deleting: boolean
}>()

const emit = defineEmits<{
  save: [body: HandlersUpdateTeamRequest]
  delete: []
}>()

const { t } = useI18n()

const form = reactive({
  name: '',
  description: '',
  shared_dir_name: '',
})

watch(
  () => props.team,
  (next) => {
    form.name = next?.name ?? ''
    form.description = next?.description ?? ''
    form.shared_dir_name = next?.shared_dir_name ?? ''
  },
  { immediate: true },
)

const hasChanges = computed(() => {
  if (!props.team) return false
  return (
    form.name.trim() !== (props.team.name ?? '')
    || form.description.trim() !== (props.team.description ?? '')
    || form.shared_dir_name.trim() !== (props.team.shared_dir_name ?? '')
  )
})

function handleSave() {
  if (!hasChanges.value || props.saving) return
  if (!form.name.trim()) return
  emit('save', {
    name: form.name.trim(),
    description: form.description.trim(),
    shared_dir_name: form.shared_dir_name.trim(),
  })
}
</script>
