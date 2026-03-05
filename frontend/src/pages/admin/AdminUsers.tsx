import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useAdminUsers } from '@/hooks/useAdminModeration'

export default function AdminUsers() {
  const [query, setQuery] = useState('')
  const [bannedOnly, setBannedOnly] = useState(false)
  const normalizedQuery = useMemo(() => query.trim(), [query])

  const {
    data: users = [],
    isLoading,
    isError,
    refetch,
  } = useAdminUsers({
    q: normalizedQuery || undefined,
    limit: 200,
  })

  const filtered = useMemo(
    () => (bannedOnly ? users.filter(u => u.is_banned) : users),
    [users, bannedOnly]
  )

  return (
    <div className='space-y-4'>
      <header>
        <h1 className='text-2xl font-bold'>Users</h1>
        <p className='text-sm text-muted-foreground'>
          Directory with moderation state and quick access to user detail.
        </p>
      </header>

      <div className='flex flex-wrap items-center gap-2'>
        <Input
          value={query}
          onChange={event => setQuery(event.target.value)}
          className='max-w-md'
          placeholder='Search username or email'
        />
        <Button
          size='sm'
          variant={bannedOnly ? 'destructive' : 'outline'}
          onClick={() => setBannedOnly(v => !v)}
        >
          {bannedOnly ? 'Showing banned only' : 'Show banned only'}
        </Button>
      </div>

      {isLoading && (
        <p className='text-sm text-muted-foreground'>Loading users…</p>
      )}
      {isError && (
        <div className='rounded-lg border border-destructive/40 bg-destructive/10 p-3'>
          <p className='text-sm text-destructive'>Failed to load users.</p>
          <Button
            size='sm'
            className='mt-2'
            variant='outline'
            onClick={() => refetch()}
          >
            Retry
          </Button>
        </div>
      )}

      {!isLoading && !isError && (
        <div className='overflow-hidden rounded-xl border border-border/70'>
          <table className='w-full text-left text-sm'>
            <thead className='bg-muted/50'>
              <tr>
                <th className='px-3 py-2 font-semibold'>ID</th>
                <th className='px-3 py-2 font-semibold'>Username</th>
                <th className='px-3 py-2 font-semibold'>Email</th>
                <th className='px-3 py-2 font-semibold'>Strikes</th>
                <th className='px-3 py-2 font-semibold'>State</th>
                <th className='px-3 py-2 font-semibold'>Action</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map(user => (
                <tr key={user.id} className='border-t border-border/60'>
                  <td className='px-3 py-2'>#{user.id}</td>
                  <td className='px-3 py-2 font-medium'>{user.username}</td>
                  <td className='px-3 py-2 text-muted-foreground'>
                    {user.email}
                  </td>
                  <td className='px-3 py-2'>
                    {(user.moderation_strikes ?? 0) > 0 ? (
                      <span className='font-semibold text-amber-500'>
                        {user.moderation_strikes}
                      </span>
                    ) : (
                      <span className='text-muted-foreground'>0</span>
                    )}
                  </td>
                  <td className='px-3 py-2'>
                    {user.is_banned ? (
                      <Badge variant='destructive'>Banned</Badge>
                    ) : (
                      <Badge variant='outline'>Active</Badge>
                    )}
                  </td>
                  <td className='px-3 py-2'>
                    <Button asChild size='sm' variant='outline'>
                      <Link to={`/admin/users/${user.id}`}>View</Link>
                    </Button>
                  </td>
                </tr>
              ))}
              {filtered.length === 0 && (
                <tr>
                  <td
                    colSpan={6}
                    className='px-3 py-4 text-center text-sm text-muted-foreground'
                  >
                    No users match the current filter.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
