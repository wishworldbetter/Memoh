<template>
  <SearchableSelectPopover
    v-model="selectedValue"
    :options="options"
    :placeholder="resolvedPlaceholder"
    :search-placeholder="resolvedSearchPlaceholder"
    :empty-text="resolvedEmptyText"
    :show-group-headers="kinds.length > 1"
  >
    <template #trigger="{ open, selectedOption: triggerOption }">
      <Button
        type="button"
        variant="outline"
        role="combobox"
        :aria-expanded="open"
        :aria-label="resolvedPlaceholder"
        :class="['justify-between font-normal', triggerClass]"
      >
        <div
          v-if="triggerOption"
          class="flex min-w-0 items-center gap-2"
        >
          <Avatar class="size-5 shrink-0">
            <AvatarImage
              v-if="optionAvatar(triggerOption)"
              :src="optionAvatar(triggerOption)"
              :alt="triggerOption.label"
            />
            <AvatarFallback class="text-[9px]">
              {{ initials(triggerOption.label) }}
            </AvatarFallback>
          </Avatar>
          <span
            class="truncate text-xs"
            :title="triggerOption.label"
          >{{ triggerOption.label }}</span>
        </div>
        <span
          v-else
          class="truncate text-xs text-muted-foreground"
        >{{ resolvedPlaceholder }}</span>
        <Search class="ml-2 size-3.5 shrink-0 text-muted-foreground" />
      </Button>
    </template>

    <template #group-label="{ group }">
      {{ groupLabel(group.key) }}
    </template>

    <template #option-label="{ option }">
      <span class="flex min-w-0 flex-1 items-center gap-2">
        <Avatar class="size-5 shrink-0">
          <AvatarImage
            v-if="optionAvatar(option)"
            :src="optionAvatar(option)"
            :alt="option.label"
          />
          <AvatarFallback class="text-[9px]">
            {{ initials(option.label) }}
          </AvatarFallback>
        </Avatar>
        <span
          class="truncate text-xs"
          :title="option.label"
        >{{ option.label }}</span>
      </span>
    </template>

    <template #option-suffix="{ option }">
      <Badge
        variant="outline"
        class="ml-auto h-4 shrink-0 px-1.5 text-[10px] font-normal"
      >
        {{ kindLabel(optionKind(option)) }}
      </Badge>
    </template>
  </SearchableSelectPopover>
</template>

<script setup lang="ts">
import type {
  AccountsAccount,
  BotsBot,
} from '@memohai/sdk'
import { computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useQuery } from '@pinia/colada'
import { getBotsQuery, getUsersQuery } from '@memohai/sdk/colada'
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
  Badge,
  Button,
} from '@memohai/ui'
import { Search } from 'lucide-vue-next'
import SearchableSelectPopover from '@/components/searchable-select-popover/index.vue'
import type { SearchableSelectOption } from '@/components/searchable-select-popover/index.vue'

export type PrincipalKind = 'bot' | 'user'

interface PrincipalOptionMeta {
  kind: PrincipalKind
  avatarUrl?: string
}

const props = withDefaults(defineProps<{
  modelValue?: string
  principalType?: PrincipalKind
  kinds?: PrincipalKind[]
  placeholder?: string
  searchPlaceholder?: string
  emptyText?: string
  triggerClass?: string
}>(), {
  modelValue: '',
  principalType: undefined,
  kinds: () => ['bot'],
  placeholder: '',
  searchPlaceholder: '',
  emptyText: '',
  triggerClass: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'update:principalType': [value: PrincipalKind | undefined]
}>()

const { t } = useI18n()

const includesBot = computed(() => props.kinds.includes('bot'))
const includesUser = computed(() => props.kinds.includes('user'))

const { data: botsData } = useQuery({
  ...getBotsQuery(),
  enabled: () => includesBot.value,
})
const { data: usersData } = useQuery({
  ...getUsersQuery(),
  enabled: () => includesUser.value,
})

const bots = computed<BotsBot[]>(() => botsData.value?.items ?? [])
const users = computed<AccountsAccount[]>(() => usersData.value?.items ?? [])

const options = computed<SearchableSelectOption[]>(() => {
  const out: SearchableSelectOption[] = []
  if (includesBot.value) {
    for (const bot of bots.value) {
      if (!bot.id) continue
      const label = bot.display_name || bot.id
      out.push({
        value: bot.id,
        label,
        group: 'bot',
        groupLabel: t('principalSelect.botsGroup'),
        keywords: [bot.display_name ?? '', bot.id].filter(Boolean) as string[],
        meta: { kind: 'bot', avatarUrl: bot.avatar_url } satisfies PrincipalOptionMeta,
      })
    }
  }
  if (includesUser.value) {
    for (const user of users.value) {
      if (!user.id) continue
      const label = user.display_name || user.username || user.id
      out.push({
        value: user.id,
        label,
        group: 'user',
        groupLabel: t('principalSelect.usersGroup'),
        keywords: [
          user.display_name ?? '',
          user.username ?? '',
          user.email ?? '',
          user.id,
        ].filter(Boolean) as string[],
        meta: { kind: 'user', avatarUrl: user.avatar_url } satisfies PrincipalOptionMeta,
      })
    }
  }
  return out
})

const selectedValue = computed({
  get: () => props.modelValue,
  set: (next: string) => {
    emit('update:modelValue', next)
    const option = options.value.find((o) => o.value === next)
    emit('update:principalType', option ? optionKind(option) : undefined)
  },
})

// When the kinds prop changes such that the current selection is no longer
// in the option set, clear the model so the trigger does not show a stale
// label. We only clear when there's truly no match — protects against
// transient empty option lists during loading.
watch(
  () => [props.kinds, options.value.length] as const,
  () => {
    if (!props.modelValue) return
    if (options.value.length === 0) return
    if (!options.value.some((o) => o.value === props.modelValue)) {
      emit('update:modelValue', '')
      emit('update:principalType', undefined)
    }
  },
  { flush: 'post' },
)

function optionKind(option: SearchableSelectOption): PrincipalKind {
  return (option.meta as PrincipalOptionMeta | undefined)?.kind ?? 'bot'
}

function optionAvatar(option: SearchableSelectOption): string {
  return (option.meta as PrincipalOptionMeta | undefined)?.avatarUrl ?? ''
}

function kindLabel(kind: PrincipalKind): string {
  return kind === 'user' ? t('principalSelect.kindUser') : t('principalSelect.kindBot')
}

function groupLabel(key: string): string {
  if (key === 'user') return t('principalSelect.usersGroup')
  if (key === 'bot') return t('principalSelect.botsGroup')
  return key
}

const resolvedPlaceholder = computed(() => {
  if (props.placeholder) return props.placeholder
  if (props.kinds.length === 1 && props.kinds[0] === 'user') return t('principalSelect.selectUserPlaceholder')
  if (props.kinds.length === 1 && props.kinds[0] === 'bot') return t('principalSelect.selectBotPlaceholder')
  return t('principalSelect.selectPrincipalPlaceholder')
})
const resolvedSearchPlaceholder = computed(() => props.searchPlaceholder || t('principalSelect.searchPlaceholder'))
const resolvedEmptyText = computed(() => props.emptyText || t('principalSelect.empty'))

function initials(name: string): string {
  return name
    .split(/[\s_-]+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((word) => word[0])
    .join('')
    .toUpperCase() || '?'
}
</script>
