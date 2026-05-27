<template>
  <div class="max-w-2xl mx-auto pb-6 space-y-5">
    <div class="flex items-start justify-between pb-4 border-b border-border/50">
      <div class="space-y-1">
        <h3 class="text-sm font-semibold text-foreground">
          {{ t('teams.tabs.instructions') }}
        </h3>
        <p class="text-[11px] text-muted-foreground">
          {{ t('teams.instructionsSettingsHint') }}
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

    <div class="space-y-4 rounded-md border p-4">
      <div class="space-y-1">
        <h4 class="text-xs font-medium">
          {{ t('teams.instructions') }}
        </h4>
        <p class="text-[11px] text-muted-foreground">
          {{ t('teams.tabs.instructionsSubtitle') }}
        </p>
      </div>
      <Textarea
        v-model="draft"
        :placeholder="t('teams.instructionsPlaceholder')"
        rows="12"
        class="text-xs"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Button, Spinner, Textarea } from '@memohai/ui'
import type { HandlersTeamResponse, HandlersUpdateTeamRequest } from '@memohai/sdk'

const props = defineProps<{
  team?: HandlersTeamResponse
  saving: boolean
}>()

const emit = defineEmits<{
  save: [body: HandlersUpdateTeamRequest]
}>()

const { t } = useI18n()

const draft = ref('')

watch(
  () => props.team,
  (next) => {
    draft.value = next?.instructions ?? ''
  },
  { immediate: true },
)

const hasChanges = computed(() =>
  draft.value.trim() !== (props.team?.instructions ?? ''),
)

function handleSave() {
  if (!hasChanges.value || props.saving) return
  emit('save', { instructions: draft.value.trim() })
}
</script>
