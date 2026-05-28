import type {
  AccountsAccount,
  BotsBot,
  HandlersMemberResponse,
} from '@memohai/sdk'
import type { Ref } from 'vue'
import { computed } from 'vue'

/**
 * Build the mention candidate pool by combining team members with
 * system-wide bots/users so the picker can offer both kinds even
 * when the team only has one member type.
 */
export function useMentionPool(
  members: Ref<HandlersMemberResponse[]>,
  bots: Ref<BotsBot[]>,
  users: Ref<AccountsAccount[]>,
) {
  return computed<HandlersMemberResponse[]>(() => {
    const seenBots = new Set<string>()
    const seenUsers = new Set<string>()
    const out: HandlersMemberResponse[] = []

    for (const m of members.value) {
      if (m.member_type === 'bot' && m.bot_id) seenBots.add(m.bot_id)
      if (m.member_type === 'user' && m.user_id) seenUsers.add(m.user_id)
      out.push(m)
    }

    for (const bot of bots.value) {
      if (!bot.id || seenBots.has(bot.id)) continue
      out.push({
        id: `__bot__${bot.id}`,
        member_type: 'bot',
        bot_id: bot.id,
        display_name: bot.name || bot.id,
      } as HandlersMemberResponse)
    }

    for (const user of users.value) {
      if (!user.id || seenUsers.has(user.id)) continue
      out.push({
        id: `__user__${user.id}`,
        member_type: 'user',
        user_id: user.id,
        display_name: user.display_name || user.id,
      } as HandlersMemberResponse)
    }

    return out
  })
}
