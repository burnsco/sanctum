import {
  useAdminDeletedComments,
  useAdminDeletedPosts,
} from '@/hooks/useAdminModeration'

function formatDate(iso: string) {
  return new Date(iso).toLocaleString()
}

export default function AdminDeletedContent() {
  const { data: posts = [], isLoading: postsLoading } = useAdminDeletedPosts({
    limit: 200,
  })
  const { data: comments = [], isLoading: commentsLoading } =
    useAdminDeletedComments({ limit: 200 })

  return (
    <div className='space-y-6'>
      <header>
        <h1 className='text-2xl font-bold'>Deleted Content</h1>
        <p className='text-sm text-muted-foreground'>
          Soft-deleted posts and comments — metadata only, no content shown.
        </p>
      </header>

      {/* Deleted Posts */}
      <section className='space-y-2'>
        <h2 className='text-base font-semibold'>
          Deleted Posts{' '}
          <span className='text-muted-foreground font-normal'>
            ({posts.length})
          </span>
        </h2>
        {postsLoading ? (
          <p className='text-sm text-muted-foreground'>Loading…</p>
        ) : posts.length === 0 ? (
          <p className='text-sm text-muted-foreground'>No deleted posts.</p>
        ) : (
          <div className='overflow-hidden rounded-xl border border-border/70'>
            <table className='w-full text-left text-sm'>
              <thead className='bg-muted/50'>
                <tr>
                  <th className='px-3 py-2 font-semibold'>Post ID</th>
                  <th className='px-3 py-2 font-semibold'>User</th>
                  <th className='px-3 py-2 font-semibold'>Type</th>
                  <th className='px-3 py-2 font-semibold'>Sanctum ID</th>
                  <th className='px-3 py-2 font-semibold'>Created</th>
                  <th className='px-3 py-2 font-semibold'>Deleted</th>
                </tr>
              </thead>
              <tbody>
                {posts.map(post => (
                  <tr key={post.id} className='border-t border-border/60'>
                    <td className='px-3 py-2'>#{post.id}</td>
                    <td className='px-3 py-2'>
                      {post.username}{' '}
                      <span className='text-muted-foreground'>
                        (#{post.user_id})
                      </span>
                    </td>
                    <td className='px-3 py-2 text-muted-foreground'>
                      {post.post_type}
                    </td>
                    <td className='px-3 py-2 text-muted-foreground'>
                      {post.sanctum_id ?? '—'}
                    </td>
                    <td className='px-3 py-2 text-muted-foreground'>
                      {formatDate(post.created_at)}
                    </td>
                    <td className='px-3 py-2 text-destructive'>
                      {formatDate(post.deleted_at)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      {/* Deleted Comments */}
      <section className='space-y-2'>
        <h2 className='text-base font-semibold'>
          Deleted Comments{' '}
          <span className='text-muted-foreground font-normal'>
            ({comments.length})
          </span>
        </h2>
        {commentsLoading ? (
          <p className='text-sm text-muted-foreground'>Loading…</p>
        ) : comments.length === 0 ? (
          <p className='text-sm text-muted-foreground'>No deleted comments.</p>
        ) : (
          <div className='overflow-hidden rounded-xl border border-border/70'>
            <table className='w-full text-left text-sm'>
              <thead className='bg-muted/50'>
                <tr>
                  <th className='px-3 py-2 font-semibold'>Comment ID</th>
                  <th className='px-3 py-2 font-semibold'>User</th>
                  <th className='px-3 py-2 font-semibold'>Post ID</th>
                  <th className='px-3 py-2 font-semibold'>Created</th>
                  <th className='px-3 py-2 font-semibold'>Deleted</th>
                </tr>
              </thead>
              <tbody>
                {comments.map(comment => (
                  <tr key={comment.id} className='border-t border-border/60'>
                    <td className='px-3 py-2'>#{comment.id}</td>
                    <td className='px-3 py-2'>
                      {comment.username}{' '}
                      <span className='text-muted-foreground'>
                        (#{comment.user_id})
                      </span>
                    </td>
                    <td className='px-3 py-2 text-muted-foreground'>
                      #{comment.post_id}
                    </td>
                    <td className='px-3 py-2 text-muted-foreground'>
                      {formatDate(comment.created_at)}
                    </td>
                    <td className='px-3 py-2 text-destructive'>
                      {formatDate(comment.deleted_at)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </div>
  )
}
