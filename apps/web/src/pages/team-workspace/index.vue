<template>
  <div class="flex h-full flex-col bg-background">
    <header class="border-b px-6 py-4">
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div class="min-w-0">
          <div class="text-xs text-muted-foreground">
            {{ t('teams.title') }}
          </div>
          <div class="mt-1 flex flex-wrap items-center gap-3">
            <h1 class="truncate text-lg font-semibold">
              {{ team?.name ?? t('teams.loading') }}
            </h1>
            <Badge
              v-if="team?.shared_dir_name"
              variant="outline"
              class="font-mono"
            >
              /team/{{ team.shared_dir_name }}
            </Badge>
            <span class="text-xs text-muted-foreground">
              {{ t('teams.issueCount', { count: issues.length }) }}
            </span>
          </div>
          <p
            v-if="team?.description"
            class="mt-1 max-w-3xl truncate text-sm text-muted-foreground"
          >
            {{ team.description }}
          </p>
        </div>
        <div class="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            @click="router.push({ name: 'team-detail', params: { teamId } })"
          >
            <Settings class="mr-1.5 size-4" /> {{ t('teams.settings') }}
          </Button>
          <Button
            size="sm"
            @click="openCreate()"
          >
            <Plus class="mr-1.5 size-4" /> {{ t('teams.newIssue') }}
          </Button>
        </div>
      </div>

      <div class="mt-4 flex flex-wrap items-center gap-3">
        <div class="relative w-full sm:w-80">
          <Search class="absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input
            v-model="searchText"
            :placeholder="t('teams.searchIssues')"
            class="pl-9"
          />
        </div>
        <Badge variant="secondary">
          {{ t('teams.openIssues', { count: openIssueCount }) }}
        </Badge>
      </div>
    </header>

    <div class="flex-1 overflow-auto px-6 py-4">
      <div class="flex min-h-full min-w-max gap-4 pb-4">
        <div
          v-if="rawIssues.length > 0 && issues.length === 0"
          class="flex w-72 shrink-0 items-center justify-center rounded-lg border border-dashed bg-muted/20 p-6 text-center text-sm text-muted-foreground"
        >
          {{ t('teams.noMatchingIssues') }}
        </div>
        <section
          v-for="column in ISSUE_COLUMNS"
          :key="column.value"
          class="flex w-72 shrink-0 flex-col rounded-lg border bg-muted/20"
        >
          <div class="flex items-center justify-between border-b px-3 py-2.5">
            <div class="flex min-w-0 items-center gap-2">
              <span
                class="size-2 rounded-full"
                :class="column.dotClass"
              />
              <div class="min-w-0">
                <div class="truncate text-sm font-medium">
                  {{ t(column.labelKey) }}
                </div>
                <div class="text-xs text-muted-foreground">
                  {{ t('teams.columnIssueCount', { count: columnIssues(column.value).length }) }}
                </div>
              </div>
            </div>
            <Button
              variant="ghost"
              size="icon"
              class="size-7"
              :title="t('teams.newIssueInStatus', { status: t(column.labelKey) })"
              @click="openCreate(column.value)"
            >
              <Plus class="size-4" />
            </Button>
          </div>

          <div
            :ref="(el) => setColumnRef(column.value, el)"
            :data-status="column.value"
            class="min-h-32 flex-1 space-y-2 p-2"
          >
            <div
              v-if="columnIssues(column.value).length === 0"
              class="rounded-md border border-dashed px-3 py-6 text-center text-xs text-muted-foreground"
            >
              {{ t('teams.emptyColumn') }}
            </div>

            <article
              v-for="issue in columnIssues(column.value)"
              :key="issue.id"
              :data-issue-id="issue.id"
              class="group cursor-pointer rounded-md border bg-background p-3 shadow-xs transition hover:border-primary/40 hover:bg-accent/30"
              @click="openIssue(issue.id)"
            >
              <div class="flex items-start justify-between gap-2">
                <div class="min-w-0">
                  <div class="text-xs text-muted-foreground">
                    #{{ issue.number }}
                  </div>
                  <h2 class="mt-1 line-clamp-2 text-sm font-medium leading-5">
                    {{ issue.title }}
                  </h2>
                </div>
                <Badge
                  v-if="isIssueUpdating(issue.id)"
                  variant="outline"
                  class="shrink-0"
                >
                  {{ t('common.saving') }}
                </Badge>
              </div>

              <p
                v-if="issue.description"
                class="mt-2 line-clamp-2 text-xs leading-5 text-muted-foreground"
              >
                {{ issue.description }}
              </p>

              <div class="mt-3 flex items-center justify-between gap-2">
                <button
                  type="button"
                  class="min-w-0 rounded text-left"
                  @click.stop="openAssignee(issue)"
                >
                  <div
                    v-if="assigneeMember(issue)"
                    class="flex min-w-0 items-center gap-1.5 text-xs text-muted-foreground"
                  >
                    <Avatar class="size-5">
                      <AvatarImage
                        v-if="assigneeAvatar(issue)"
                        :src="assigneeAvatar(issue)"
                        :alt="assigneeLabel(issue)"
                      />
                      <AvatarFallback class="text-[9px]">
                        {{ initials(assigneeLabel(issue)) }}
                      </AvatarFallback>
                    </Avatar>
                    <span class="truncate">{{ assigneeLabel(issue) }}</span>
                  </div>
                  <div
                    v-else
                    class="flex items-center gap-1.5 text-xs text-muted-foreground"
                  >
                    <UserCircle class="size-4" />
                    {{ t('teams.unassigned') }}
                  </div>
                </button>

                <div
                  class="w-28"
                  @click.stop
                >
                  <Select
                    :model-value="issueStatus(issue)"
                    @update:model-value="(value) => changeIssueStatus(issue, String(value) as IssueStatus)"
                  >
                    <SelectTrigger class="h-7 text-xs">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem
                        v-for="status in ISSUE_COLUMNS"
                        :key="status.value"
                        :value="status.value"
                      >
                        {{ t(status.labelKey) }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </article>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type {
  BotsBot,
  HandlersIssueResponse,
  HandlersMemberResponse,
} from '@memohai/sdk'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useMutation, useQuery, useQueryCache } from '@pinia/colada'
import { toast } from 'vue-sonner'
import { Plus, Search, Settings, UserCircle } from 'lucide-vue-next'
import Sortable from 'sortablejs'
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
  Badge,
  Button,
  Input,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@memohai/ui'
import {
  getBots,
  getTeamsByTeamId,
  getTeamsByTeamIdIssues,
  getTeamsByTeamIdMembers,
  putTeamsByTeamIdIssuesByIssueId,
} from '@memohai/sdk'
import { resolveApiErrorMessage } from '@/utils/api-error'

type IssueStatus = 'backlog' | 'todo' | 'in_progress' | 'blocked' | 'review' | 'done' | 'cancelled'

const ISSUE_COLUMNS: Array<{
  value: IssueStatus
  labelKey: string
  dotClass: string
}> = [
  { value: 'backlog', labelKey: 'teams.status.backlog', dotClass: 'bg-slate-400' },
  { value: 'todo', labelKey: 'teams.status.todo', dotClass: 'bg-blue-500' },
  { value: 'in_progress', labelKey: 'teams.status.inProgress', dotClass: 'bg-amber-500' },
  { value: 'blocked', labelKey: 'teams.status.blocked', dotClass: 'bg-red-500' },
  { value: 'review', labelKey: 'teams.status.review', dotClass: 'bg-violet-500' },
  { value: 'done', labelKey: 'teams.status.done', dotClass: 'bg-emerald-500' },
  { value: 'cancelled', labelKey: 'teams.status.cancelled', dotClass: 'bg-zinc-400' },
]

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const queryCache = useQueryCache()

const teamId = computed(() => String(route.params.teamId ?? ''))

const { data: teamData } = useQuery({
  key: () => ['team', teamId.value],
  query: async () => {
    const { data, error } = await getTeamsByTeamId({ path: { team_id: teamId.value } })
    if (error) throw error
    return data
  },
  enabled: () => !!teamId.value,
})

const { data: issuesData } = useQuery({
  key: () => ['team', teamId.value, 'issues'],
  query: async () => {
    const { data, error } = await getTeamsByTeamIdIssues({ path: { team_id: teamId.value } })
    if (error) throw error
    return data ?? []
  },
  enabled: () => !!teamId.value,
})

const { data: membersData } = useQuery({
  key: () => ['team', teamId.value, 'members'],
  query: async () => {
    const { data, error } = await getTeamsByTeamIdMembers({ path: { team_id: teamId.value } })
    if (error) throw error
    return data ?? []
  },
  enabled: () => !!teamId.value,
})

const { data: botListData } = useQuery({
  key: () => ['bots-for-team-board', teamId.value],
  query: async () => {
    const { data, error } = await getBots()
    if (error) throw error
    return data?.items ?? []
  },
})

const team = computed(() => teamData.value)
const searchText = ref('')
const optimisticStatuses = ref<Record<string, IssueStatus>>({})
const updatingIssueIDs = ref<Set<string>>(new Set())

const rawIssues = computed(() => issuesData.value ?? [])
const issues = computed(() => {
  const keyword = searchText.value.trim().toLowerCase()
  return rawIssues.value
    .map((issue) => ({
      ...issue,
      status: issueStatus(issue),
    }))
    .filter((issue) => {
      if (!keyword) return true
      return `${issue.number ?? ''} ${issue.title ?? ''} ${issue.description ?? ''}`.toLowerCase().includes(keyword)
    })
})

const openIssueCount = computed(() =>
  rawIssues.value.filter((issue) => !['done', 'cancelled'].includes(issue.status ?? '')).length,
)

const members = computed(() => membersData.value ?? [])
const bots = computed<BotsBot[]>(() => botListData.value ?? [])

const columnElements = new Map<IssueStatus, HTMLElement>()
let sortables: Sortable[] = []

const { mutateAsync: updateIssueStatus } = useMutation({
  mutation: async (payload: { issueID: string, status: IssueStatus }) => {
    const { data, error } = await putTeamsByTeamIdIssuesByIssueId({
      path: { team_id: teamId.value, issue_id: payload.issueID },
      body: { status: payload.status },
    })
    if (error) throw error
    return data
  },
})

watch(
  () => [issues.value.length, teamId.value],
  () => {
    void nextTick(mountSortables)
  },
  { immediate: true },
)

onMounted(() => {
  void nextTick(mountSortables)
})

onBeforeUnmount(() => {
  destroySortables()
})

function setColumnRef(status: IssueStatus, el: Element | null) {
  if (el instanceof HTMLElement) {
    columnElements.set(status, el)
  }
}

function mountSortables() {
  destroySortables()
  for (const column of ISSUE_COLUMNS) {
    const el = columnElements.get(column.value)
    if (!el) continue
    sortables.push(Sortable.create(el, {
      group: 'team-issue-board',
      animation: 150,
      sort: false,
      draggable: '[data-issue-id]',
      ghostClass: 'opacity-50',
      chosenClass: 'ring-2',
      onEnd: (event) => {
        const issueID = event.item.getAttribute('data-issue-id')
        const nextStatus = event.to.getAttribute('data-status') as IssueStatus | null
        const issue = rawIssues.value.find((item) => item.id === issueID)
        if (!issue || !nextStatus) return
        void changeIssueStatus(issue, nextStatus)
      },
    }))
  }
}

function destroySortables() {
  for (const sortable of sortables) {
    sortable.destroy()
  }
  sortables = []
}

function columnIssues(status: IssueStatus) {
  return issues.value
    .filter((issue) => issue.status === status)
    .sort((a, b) => (a.number ?? 0) - (b.number ?? 0))
}

function issueStatus(issue: HandlersIssueResponse): IssueStatus {
  if (issue.id && optimisticStatuses.value[issue.id]) {
    return optimisticStatuses.value[issue.id]
  }
  return normalizeStatus(issue.status)
}

function normalizeStatus(status: string | undefined): IssueStatus {
  if (ISSUE_COLUMNS.some((column) => column.value === status)) {
    return status as IssueStatus
  }
  return 'todo'
}

function isIssueUpdating(issueID: string | undefined) {
  return !!issueID && updatingIssueIDs.value.has(issueID)
}

function setIssueUpdating(issueID: string | undefined, value: boolean) {
  if (!issueID) return
  const next = new Set(updatingIssueIDs.value)
  if (value) {
    next.add(issueID)
  } else {
    next.delete(issueID)
  }
  updatingIssueIDs.value = next
}

async function changeIssueStatus(issue: HandlersIssueResponse, nextStatus: IssueStatus) {
  if (!issue.id) return
  const previousStatus = issueStatus(issue)
  if (previousStatus === nextStatus) return
  optimisticStatuses.value = { ...optimisticStatuses.value, [issue.id]: nextStatus }
  setIssueUpdating(issue.id, true)
  try {
    await updateIssueStatus({ issueID: issue.id, status: nextStatus })
    void queryCache.invalidateQueries({ key: ['team', teamId.value, 'issues'] })
  }
  catch (err) {
    const { [issue.id]: _removed, ...rest } = optimisticStatuses.value
    optimisticStatuses.value = rest
    toast.error(resolveApiErrorMessage(err, t('teams.issueUpdateFailed')))
  }
  finally {
    setIssueUpdating(issue.id, false)
  }
}

function openCreate(status: IssueStatus = 'todo') {
  router.push({ name: 'team-issue-new', params: { teamId: teamId.value }, query: { status } })
}

function openIssue(id?: string) {
  if (!id) return
  router.push({ name: 'team-issue', params: { teamId: teamId.value, issueId: id } })
}

function openAssignee(issue: HandlersIssueResponse) {
  if (!issue.id) return
  router.push({ name: 'team-issue', params: { teamId: teamId.value, issueId: issue.id } })
}

function assigneeMember(issue: HandlersIssueResponse) {
  if (issue.assignee_type === 'bot' && issue.assignee_bot_id) {
    return members.value.find((member) => member.member_type === 'bot' && member.bot_id === issue.assignee_bot_id)
  }
  if (issue.assignee_type === 'user' && issue.assignee_user_id) {
    return members.value.find((member) => member.member_type === 'user' && member.user_id === issue.assignee_user_id)
  }
  return undefined
}

function assigneeLabel(issue: HandlersIssueResponse) {
  const member = assigneeMember(issue)
  if (member) return memberLabel(member)
  return issue.assignee_bot_id || issue.assignee_user_id || t('teams.unassigned')
}

function assigneeAvatar(issue: HandlersIssueResponse) {
  if (!issue.assignee_bot_id) return ''
  return bots.value.find((bot) => bot.id === issue.assignee_bot_id)?.avatar_url ?? ''
}

function memberLabel(member: HandlersMemberResponse) {
  return member.display_name || member.bot_id || member.user_id || t('teams.unknownMember')
}

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
