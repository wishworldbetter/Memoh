<template>
  <MasterDetailSidebarLayout class="[&_td:last-child]:w-45">
    <template #sidebar-header>
      <div class="flex flex-col p-4 pb-3">
        <div class="flex items-center gap-3">
          <div class="group/avatar relative size-12 shrink-0 overflow-hidden rounded-full bg-muted">
            <Avatar class="size-12 rounded-full">
              <AvatarImage
                v-if="bot?.avatar_url"
                :src="bot.avatar_url"
                :alt="bot.display_name"
              />
              <AvatarFallback class="text-lg">
                {{ avatarFallback }}
              </AvatarFallback>
            </Avatar>
            <button
              type="button"
              class="absolute inset-0 flex items-center justify-center rounded-full bg-black/40 opacity-0 transition-opacity group-hover/avatar:opacity-100"
              :title="$t('common.edit')"
              :aria-label="$t('common.edit')"
              :disabled="!bot || botLifecyclePending"
              @click="$emit('editAvatar')"
            >
              <SquarePen class="size-4 text-white" />
            </button>
          </div>

          <div class="flex min-w-0 flex-1 flex-col justify-center">
            <div class="group/name relative flex min-w-0 items-center gap-1">
              <template v-if="isEditingBotName && bot">
                <Input
                  ref="editNameInputRef"
                  :model-value="botNameDraft"
                  class="h-7 w-full px-2 pr-6 text-xs shadow-none"
                  :placeholder="$t('bots.displayNamePlaceholder')"
                  :disabled="isSavingBotName"
                  @update:model-value="$emit('update:botNameDraft', String($event))"
                  @keydown.enter.prevent="$emit('confirmBotName')"
                  @keydown.esc.prevent="$emit('cancelBotName')"
                  @blur="$emit('confirmBotName')"
                />
                <div class="pointer-events-none absolute right-1.5 top-1/2 -translate-y-1/2 opacity-50">
                  <Check class="size-3" />
                </div>
              </template>
              <template v-else>
                <h2 class="truncate text-sm font-semibold text-foreground">
                  {{ botNameDraft.trim() || bot?.display_name || botId }}
                </h2>
                <button
                  v-if="bot"
                  type="button"
                  class="shrink-0 p-1 opacity-0 group-hover/name:opacity-100"
                  :disabled="botLifecyclePending"
                  @click="$emit('startEditBotName')"
                >
                  <SquarePen class="size-3 text-muted-foreground" />
                </button>
              </template>
            </div>

            <div class="mt-0.5">
              <div
                v-if="bot"
                class="inline-flex h-5 items-center gap-1.5 rounded-full bg-[#27272a] px-2"
                :title="hasIssue ? issueTitle : undefined"
              >
                <LoaderCircle
                  v-if="bot.status === 'creating' || bot.status === 'deleting'"
                  class="size-2.5 animate-spin text-[#d0d0d4]"
                />
                <div
                  v-else
                  class="size-1.5 rounded-full"
                  :class="statusVariant === 'destructive' ? 'bg-destructive' : 'bg-success'"
                />
                <span class="text-[10px] font-medium text-[#d0d0d4]">{{ statusLabel }}</span>
              </div>
              <span
                v-if="bot?.type"
                class="ml-1.5 text-[10px] text-muted-foreground"
              >
                {{ botTypeLabel }}
              </span>
            </div>
          </div>
        </div>

        <div class="relative mt-4">
          <Search class="absolute left-2.5 top-1/2 size-3 -translate-y-1/2 text-muted-foreground" />
          <Input
            :model-value="searchQuery"
            type="text"
            class="h-8 bg-transparent pl-8 text-xs shadow-none focus-visible:ring-0"
            :placeholder="$t('common.search')"
            @update:model-value="$emit('update:searchQuery', String($event))"
          />
          <button
            v-if="searchQuery"
            type="button"
            class="absolute right-2 top-1/2 flex size-4 shrink-0 -translate-y-1/2 items-center justify-center rounded-full text-muted-foreground hover:bg-muted"
            @click="$emit('update:searchQuery', '')"
          >
            <X class="size-2.5" />
          </button>
        </div>
      </div>
    </template>

    <template #sidebar-content>
      <div
        v-if="searchQuery"
        class="flex flex-col gap-1"
      >
        <div
          v-if="searchResults.length === 0"
          class="px-3 py-4 text-center text-xs text-muted-foreground"
        >
          {{ $t('common.noData') }}
        </div>
        <SidebarMenu
          v-else
          class="m-0 gap-1 p-0"
        >
          <SidebarMenuItem
            v-for="(result, idx) in searchResults"
            :key="idx"
          >
            <SidebarMenuButton
              as-child
              class="h-11 justify-start px-0 py-0! before:hidden"
            >
              <button
                class="group/result flex w-full flex-col items-start justify-center rounded-md border border-transparent px-3 py-2 text-left transition-colors hover:bg-accent hover:text-accent-foreground"
                @click="selectSearchResult(result.tab)"
              >
                <span class="text-xs font-medium text-foreground group-hover/result:text-accent-foreground">{{ result.translatedTitle }}</span>
                <span class="mt-1 flex items-center gap-1 text-[10px] text-muted-foreground group-hover/result:text-accent-foreground/70">
                  <component
                    :is="tabList.find(t => t.value === result.tab)?.icon"
                    class="size-3 opacity-70"
                  />
                  {{ $t(`bots.tabs.${result.tab}`) }}
                </span>
              </button>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </div>

      <template v-else>
        <div
          v-for="(group, idx) in groupedTabs"
          :key="group.key"
          :class="idx > 0 ? 'mt-4' : ''"
          class="flex flex-col gap-0.5"
        >
          <SidebarMenu
            v-for="tab in group.items"
            :key="tab.value"
            class="m-0 p-0"
          >
            <SidebarMenuItem>
              <SidebarMenuButton
                as-child
                :is-active="activeTab === tab.value"
                class="h-10 justify-start px-0 py-0! before:hidden"
              >
                <Toggle
                  class="h-10 w-full justify-start gap-3 border-0 bg-transparent! px-3 text-xs font-medium transition-colors"
                  :model-value="activeTab === tab.value"
                  @update:model-value="(isSelect: boolean) => selectTab(tab.value, isSelect)"
                >
                  <component
                    :is="tab.icon"
                    v-if="tab.icon"
                    class="size-4 shrink-0"
                  />
                  <span class="whitespace-nowrap">{{ $t(tab.label) }}</span>
                </Toggle>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </div>
      </template>
    </template>

    <template #sidebar-footer />

    <template #detail>
      <slot name="detail" />
    </template>
  </MasterDetailSidebarLayout>
</template>

<script setup lang="ts">
import type { Component } from 'vue'
import type { BotsBot } from '@memohai/sdk'
import { nextTick, ref, watch } from 'vue'
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
  Input,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  Toggle,
} from '@memohai/ui'
import { Check, LoaderCircle, Search, SquarePen, X } from 'lucide-vue-next'
import MasterDetailSidebarLayout from '@/components/master-detail-sidebar-layout/index.vue'

type BotDetailSidebarTab = {
  value: string
  label: string
  icon?: Component
}

type BotDetailSidebarGroup = {
  key: string
  items: BotDetailSidebarTab[]
}

type BotDetailSearchResult = {
  tab: string
  translatedTitle: string
}

const props = defineProps<{
  bot?: BotsBot
  botId: string
  avatarFallback: string
  activeTab: string
  botNameDraft: string
  botTypeLabel: string
  botLifecyclePending: boolean
  groupedTabs: BotDetailSidebarGroup[]
  hasIssue: boolean
  isEditingBotName: boolean
  isSavingBotName: boolean
  issueTitle?: string
  searchQuery: string
  searchResults: BotDetailSearchResult[]
  statusLabel: string
  statusVariant: string
  tabList: BotDetailSidebarTab[]
}>()

const emit = defineEmits<{
  'update:activeTab': [value: string]
  'update:botNameDraft': [value: string]
  'update:searchQuery': [value: string]
  cancelBotName: []
  confirmBotName: []
  editAvatar: []
  startEditBotName: []
}>()

const editNameInputRef = ref<InstanceType<typeof Input> | null>(null)

watch(() => props.isEditingBotName, (isEditing) => {
  if (!isEditing) return
  nextTick(() => {
    const el = editNameInputRef.value?.$el
    if (!el) return
    const input = el instanceof HTMLInputElement ? el : el.querySelector('input')
    input?.focus()
  })
})

function selectSearchResult(tab: string) {
  emit('update:activeTab', tab)
  emit('update:searchQuery', '')
}

function selectTab(tab: string, isSelect: boolean) {
  if (isSelect) {
    emit('update:activeTab', tab)
  }
}
</script>
