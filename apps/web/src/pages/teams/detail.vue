<template>
  <section class="absolute inset-0 flex flex-col bg-background">
    <div class="relative flex-1">
      <TeamDetailSidebar
        v-model:active-tab="activeTab"
        v-model:name-draft="nameDraft"
        v-model:search-query="searchQuery"
        :avatar-fallback="avatarFallback"
        :team="team"
        :team-id="teamId"
        :is-editing-name="isEditingName"
        :is-saving-name="isSavingName"
        :member-count="memberCount"
        :grouped-tabs="groupedTabs"
        :search-results="searchResults"
        :tab-list="tabList"
        @back="router.push({ name: 'teams' })"
        @cancel-name="handleCancelName"
        @confirm-name="handleConfirmName"
        @edit-avatar="handleEditAvatar"
        @open-workspace="router.push({ name: 'team-workspace', params: { teamId } })"
        @start-edit-name="handleStartEditName"
      >
        <template #detail>
          <div class="absolute inset-0 overflow-y-auto bg-background">
            <div class="px-6 pt-4 pb-4">
              <KeepAlive>
                <component
                  :is="activeComponent?.component"
                  v-bind="activeComponent?.params"
                  v-on="activeComponent?.events ?? {}"
                />
              </KeepAlive>
            </div>
          </div>
        </template>
      </TeamDetailSidebar>
    </div>

    <AvatarEditDialog
      v-model:open="avatarDialogOpen"
      v-model:avatar-url="avatarUrlModel"
      :fallback-text="avatarFallback"
      :title="t('teams.editAvatar')"
      :description="t('teams.editAvatarDescription')"
      :placeholder="t('teams.avatarUrlPlaceholder')"
    />
  </section>
</template>

<script setup lang="ts">
import type { Component } from 'vue'
import type {
  HandlersTeamResponse,
  HandlersUpdateTeamRequest,
} from '@memohai/sdk'
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useMutation, useQuery, useQueryCache } from '@pinia/colada'
import { toast } from 'vue-sonner'
import {
  FileText,
  LayoutDashboard,
  Settings,
  Users,
} from 'lucide-vue-next'
import {
  deleteTeamsByTeamId,
  getTeamsByTeamId,
  getTeamsByTeamIdMembers,
  putTeamsByTeamId,
} from '@memohai/sdk'
import AvatarEditDialog from '@/components/avatar-edit-dialog/index.vue'
import TeamDetailSidebar from './components/team-detail-sidebar.vue'
import TeamOverview from './components/team-overview.vue'
import TeamGeneral from './components/team-general.vue'
import TeamInstructions from './components/team-instructions.vue'
import TeamMembers from './components/team-members.vue'
import { useSyncedQueryParam } from '@/composables/useSyncedQueryParam'
import { useAvatarInitials } from '@/composables/useAvatarInitials'
import { resolveApiErrorMessage } from '@/utils/api-error'

type TeamTab = 'overview' | 'general' | 'instructions' | 'members'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const queryCache = useQueryCache()

const teamId = computed(() => String(route.params.teamId ?? ''))
const activeTab = useSyncedQueryParam('tab', 'overview')

const { data: teamData } = useQuery({
  key: () => ['team', teamId.value],
  query: async () => {
    const { data, error } = await getTeamsByTeamId({ path: { team_id: teamId.value } })
    if (error) throw error
    return data
  },
  enabled: () => !!teamId.value,
})

// Lightweight members query in the parent so the sidebar can show member count
// without forcing TeamMembers to mount. Pinia Colada will dedupe against the
// query owned by TeamMembers since both share the same key.
const { data: membersData } = useQuery({
  key: () => ['team', teamId.value, 'members'],
  query: async () => {
    const { data, error } = await getTeamsByTeamIdMembers({ path: { team_id: teamId.value } })
    if (error) throw error
    return data ?? []
  },
  enabled: () => !!teamId.value,
})

const team = computed<HandlersTeamResponse | undefined>(() => teamData.value)
const memberCount = computed(() => membersData.value?.length ?? 0)

const avatarFallback = useAvatarInitials(() => team.value?.name || teamId.value || '')

const isEditingName = ref(false)
const nameDraft = ref('')

watch(team, (next) => {
  if (!isEditingName.value) {
    nameDraft.value = next?.name ?? ''
  }
}, { immediate: true })

watch(teamId, () => {
  isEditingName.value = false
  nameDraft.value = ''
})

const { mutateAsync: updateTeamMutation, isLoading: savingTeam } = useMutation({
  mutation: async (body: HandlersUpdateTeamRequest) => {
    const { data, error } = await putTeamsByTeamId({
      path: { team_id: teamId.value },
      body,
    })
    if (error) throw error
    return data
  },
  onSettled: () => {
    void queryCache.invalidateQueries({ key: ['teams'] })
    void queryCache.invalidateQueries({ key: ['team', teamId.value] })
  },
})

async function saveTeamPatch(body: HandlersUpdateTeamRequest, opts: { silent?: boolean } = {}) {
  try {
    await updateTeamMutation(body)
    if (!opts.silent) toast.success(t('teams.updateSuccess'))
  }
  catch (err) {
    toast.error(resolveApiErrorMessage(err, t('teams.updateFailed')))
    throw err
  }
}

const { mutateAsync: deleteTeamMutation, isLoading: deletingTeam } = useMutation({
  mutation: async () => {
    const { error } = await deleteTeamsByTeamId({
      path: { team_id: teamId.value },
    })
    if (error) throw error
  },
})

async function handleDeleteTeam() {
  try {
    await deleteTeamMutation()
    toast.success(t('teams.teamDeleted'))
    void queryCache.invalidateQueries({ key: ['teams'] })
    router.push({ name: 'teams' })
  }
  catch (err) {
    toast.error(resolveApiErrorMessage(err, t('teams.teamDeleteFailed')))
  }
}

type TabConfig = {
  value: TeamTab
  label: string
  icon: Component
  component: Component
  params: Record<string, unknown>
  events?: Record<string, (...args: unknown[]) => unknown>
}

const tabList = computed<TabConfig[]>(() => [
  {
    value: 'overview',
    label: 'teams.tabs.overview',
    icon: LayoutDashboard,
    component: TeamOverview,
    params: { team: team.value, memberCount: memberCount.value },
  },
  {
    value: 'general',
    label: 'teams.tabs.general',
    icon: Settings,
    component: TeamGeneral,
    params: { team: team.value, saving: savingTeam.value, deleting: deletingTeam.value },
    events: {
      save: (body) => { void saveTeamPatch(body as HandlersUpdateTeamRequest) },
      delete: () => { void handleDeleteTeam() },
    },
  },
  {
    value: 'instructions',
    label: 'teams.tabs.instructions',
    icon: FileText,
    component: TeamInstructions,
    params: { team: team.value, saving: savingTeam.value },
    events: {
      save: (body) => { void saveTeamPatch(body as HandlersUpdateTeamRequest) },
    },
  },
  {
    value: 'members',
    label: 'teams.tabs.members',
    icon: Users,
    component: TeamMembers,
    params: { teamId: teamId.value },
  },
])

const groupedTabs = computed(() => {
  const coreKeys: TeamTab[] = ['overview', 'general']
  const capabilityKeys: TeamTab[] = ['instructions', 'members']
  return [
    { key: 'core', items: tabList.value.filter((tab) => coreKeys.includes(tab.value)) },
    { key: 'capabilities', items: tabList.value.filter((tab) => capabilityKeys.includes(tab.value)) },
  ].filter((g) => g.items.length > 0)
})

watch([tabList, activeTab], ([tabs, tab]) => {
  if (!tabs.some((item) => item.value === tab)) {
    activeTab.value = 'overview'
  }
}, { immediate: true })

const activeComponent = computed(() =>
  tabList.value.find((tab) => tab.value === activeTab.value),
)

const searchQuery = ref('')

const searchIndex = computed(() => [
  { tab: 'overview', key: 'teams.teamSummary', keywords: ['summary', 'metadata', 'id'] },
  { tab: 'general', key: 'teams.generalSettings', keywords: ['name', 'description', 'shared', 'directory'] },
  { tab: 'general', key: 'common.dangerZone', keywords: ['delete', 'remove', 'danger'] },
  { tab: 'instructions', key: 'teams.instructions', keywords: ['instructions', 'context', 'prompt'] },
  { tab: 'members', key: 'teams.members', keywords: ['members', 'bots', 'users', 'roles'] },
].map((item) => ({
  ...item,
  translatedTitle: t(item.key),
})))

const searchResults = computed(() => {
  const query = searchQuery.value.toLowerCase().trim()
  if (!query) return []
  return searchIndex.value.filter((item) =>
    item.translatedTitle.toLowerCase().includes(query)
    || item.keywords.some((k) => k.toLowerCase().includes(query))
    || t(`teams.tabs.${item.tab}`).toLowerCase().includes(query)
    || item.tab.toLowerCase().includes(query),
  )
})

const isSavingName = computed(() => savingTeam.value)

const canConfirmName = computed(() => {
  if (!team.value) return false
  const next = nameDraft.value.trim()
  if (!next) return false
  return next !== (team.value.name || '').trim()
})

function handleStartEditName() {
  if (!team.value) return
  isEditingName.value = true
  nameDraft.value = team.value.name || ''
}

function handleCancelName() {
  isEditingName.value = false
  nameDraft.value = team.value?.name || ''
}

async function handleConfirmName() {
  if (!team.value || !canConfirmName.value) {
    handleCancelName()
    return
  }
  const nextName = nameDraft.value.trim()
  try {
    await saveTeamPatch({ name: nextName }, { silent: true })
    isEditingName.value = false
    toast.success(t('teams.updateSuccess'))
  }
  catch {
    // toast already surfaced inside saveTeamPatch
  }
}

const avatarDialogOpen = ref(false)

const avatarUrlModel = computed<string>({
  get: () => team.value?.avatar_url ?? '',
  set: (next) => {
    const trimmed = (next ?? '').trim()
    if (trimmed === (team.value?.avatar_url ?? '')) return
    void saveTeamPatch({ avatar_url: trimmed })
  },
})

function handleEditAvatar() {
  if (!team.value) return
  avatarDialogOpen.value = true
}
</script>
