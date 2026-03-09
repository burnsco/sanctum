import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Gamepad2, Trophy, UserCircle, Users } from 'lucide-react'
import { useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { apiClient } from '@/api/client'
import battleshipImg from '@/assets/images/battleship.webp'
import checkersImg from '@/assets/images/checkers.png'
import connect4Img from '@/assets/images/connect4.webp'
import othelloImg from '@/assets/images/Othello.jpg'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { gameKeys, getCurrentUser } from '@/hooks'
import {
  getResumableGameRooms,
  removeResumableGameRoom,
} from '@/lib/game-room-presence'

const GAMES = [
  {
    id: 'connect4',
    name: 'Connect Four',
    description: '7x6 Gravity Match',
    reward: '+15 VP',
    color: 'blue',
    image: connect4Img,
  },
  {
    id: 'checkers',
    name: 'Checkers',
    description: 'Classic Jumps',
    reward: '+20 VP',
    image: checkersImg,
  },
  {
    id: 'othello',
    name: 'Othello',
    description: 'Reversi Strategy',
    reward: '+25 VP',
    image: othelloImg,
  },
  {
    id: 'battleship',
    name: 'Battleship',
    description: 'Naval Warfare',
    reward: '+30 VP',
    image: battleshipImg,
  },
]

interface GameRoom {
  id: number
  type: string
  status: string
  creator_id?: number
  opponent_id?: number | null
  creator?: {
    username: string
  }
}

export default function Games() {
  const navigate = useNavigate()
  const currentUser = getCurrentUser()
  const queryClient = useQueryClient()

  const { data: activeRooms, isLoading } = useQuery({
    queryKey: gameKeys.roomsActive(),
    queryFn: () => apiClient.getActiveGameRooms(),
    refetchInterval: 2000,
    refetchIntervalInBackground: true,
    refetchOnWindowFocus: true,
  })

  const closeRoom = async (roomId: number) => {
    try {
      await apiClient.leaveGameRoom(roomId)
      removeResumableGameRoom(currentUser?.id, roomId)
      await queryClient.invalidateQueries({ queryKey: gameKeys.roomsActive() })
      toast.success('Room closed')
    } catch (error) {
      console.error('Failed to close room', error)
      toast.error('Failed to close room')
    }
  }

  const handlePlayNow = async (type: string) => {
    if (
      type === 'connect4' ||
      type === 'othello' ||
      type === 'battleship' ||
      type === 'checkers'
    ) {
      try {
        // Fetch fresh rooms at click time to avoid stale cache races.
        const freshRooms = await apiClient.getActiveGameRooms(type)
        const openRoom = freshRooms.find(
          room =>
            room.status === 'pending' &&
            room.creator_id &&
            room.creator_id !== currentUser?.id &&
            !room.opponent_id
        )

        if (openRoom) {
          navigate(`/games/${type}/${openRoom.id}`)
          return
        } else {
          const myPendingRoom = freshRooms.find(
            room =>
              room.status === 'pending' &&
              room.creator_id &&
              room.creator_id === currentUser?.id
          )
          if (myPendingRoom) {
            navigate(`/games/${type}/${myPendingRoom.id}`)
            return
          }

          const room = await apiClient.createGameRoom(type)
          navigate(`/games/${type}/${room.id}`)
        }
      } catch (err) {
        console.error('Failed to handle play now', err)
      }
    } else {
      const routeMap: Record<string, string> = {
        chess: '/games/chess',
        trivia: '/games/trivia',
        blackjack: '/games/blackjack',
        poker: '/games/poker',
        'crazy-eights': '/games/crazy-eights',
        hearts: '/games/hearts',
        president: '/games/president',
        'draw-guess': '/games/draw-guess',
        snake: '/games/snake',
      }
      navigate(routeMap[type] || '/games')
    }
  }

  const joinableRooms =
    (activeRooms as GameRoom[] | undefined)?.filter(
      room =>
        room.status === 'pending' &&
        room.creator_id &&
        room.creator_id !== currentUser?.id &&
        !room.opponent_id
    ) ?? []

  const myPendingRooms =
    (activeRooms as GameRoom[] | undefined)?.filter(
      room =>
        room.status === 'pending' &&
        room.creator_id &&
        room.creator_id === currentUser?.id &&
        !room.opponent_id
    ) ?? []

  const resumableRooms = useMemo(() => {
    const pendingRoomIds = new Set(myPendingRooms.map(room => room.id))
    return getResumableGameRooms(currentUser?.id).filter(
      room => !pendingRoomIds.has(room.roomId)
    )
  }, [currentUser?.id, myPendingRooms])

  return (
    <div className='flex-1 overflow-y-auto bg-background'>
      <main className='max-w-7xl mx-auto px-4 py-12'>
        <div className='flex flex-col md:flex-row justify-between items-start md:items-center mb-12 gap-6'>
          <div>
            <h1 className='text-4xl font-black italic uppercase tracking-tighter text-foreground mb-2 flex items-center gap-3'>
              <Gamepad2 className='w-10 h-10 text-primary' />
              Arcade
            </h1>
            <p className='text-muted-foreground font-medium'>
              Global Multiplayer & Competitive Leaderboards
            </p>
          </div>
          <div className='flex gap-4'>
            <Button
              size='lg'
              variant='outline'
              className='gap-2 border-primary text-primary font-bold'
            >
              <Trophy className='w-4 h-4' /> Leaderboards
            </Button>
          </div>
        </div>

        <div className='grid lg:grid-cols-4 gap-8'>
          <div className='lg:col-span-3 space-y-8'>
            <div className='flex items-center gap-2 border-b pb-2'>
              <Gamepad2 className='w-5 h-5 text-primary' />
              <h2 className='text-xl font-bold uppercase tracking-tight'>
                Live Games
              </h2>
            </div>
            <div className='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6'>
              {GAMES.map(game => (
                <Card
                  key={game.id}
                  className='group overflow-hidden border-2 transition-all hover:scale-[1.02] hover:shadow-2xl border-primary/20 hover:border-primary bg-card'
                >
                  <CardContent className='p-0 flex flex-col h-full'>
                    <div className='relative aspect-4/5 overflow-hidden'>
                      <img
                        src={game.image}
                        alt={game.name}
                        className='w-full h-full object-cover transition-transform duration-500 group-hover:scale-110'
                      />
                      <div className='absolute inset-0 bg-linear-to-t from-black/90 via-black/20 to-transparent' />
                      <div className='absolute top-4 right-4 px-3 py-1 bg-primary text-primary-foreground text-[10px] font-black uppercase rounded-full shadow-lg'>
                        Live
                      </div>

                      <div className='absolute bottom-0 left-0 right-0 p-6'>
                        <h3 className='font-black text-2xl text-white italic uppercase tracking-tighter mb-2 group-hover:text-primary transition-colors'>
                          {game.name}
                        </h3>
                        <div className='flex items-center gap-3 mb-4'>
                          <span className='text-[10px] font-black text-primary-foreground px-2 py-0.5 bg-primary rounded-md'>
                            {game.reward}
                          </span>
                          <p className='text-xs text-slate-300 font-bold uppercase tracking-wider'>
                            {game.description}
                          </p>
                        </div>
                        <Button
                          size='lg'
                          className='w-full font-black uppercase italic tracking-widest text-sm h-12 shadow-xl hover:scale-105 transition-transform'
                          onClick={() => handlePlayNow(game.id)}
                        >
                          Play Now
                        </Button>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>

          <div className='space-y-8'>
            <section className='bg-card border-2 rounded-2xl p-6 shadow-sm'>
              <h3 className='text-xl font-black uppercase italic tracking-tighter mb-6 flex items-center gap-2'>
                <Users className='w-5 h-5 text-primary' /> Active Hubs
              </h3>
              {isLoading ? (
                <div className='space-y-4'>
                  {[1, 2, 3].map(i => (
                    <div
                      key={i}
                      className='h-12 bg-muted animate-pulse rounded-xl'
                    />
                  ))}
                </div>
              ) : (
                <div className='space-y-4'>
                  {resumableRooms.length > 0 && (
                    <div className='space-y-3'>
                      <p className='text-[10px] font-black uppercase tracking-wider text-muted-foreground'>
                        Resume Matches
                      </p>
                      {resumableRooms.map(room => (
                        <div
                          key={`${room.type}-${room.roomId}`}
                          className='flex items-center justify-between rounded-xl border border-emerald-300/30 bg-emerald-500/5 p-3'
                        >
                          <div className='flex flex-col'>
                            <span className='text-xs font-black uppercase tracking-tighter text-emerald-600'>
                              {room.type === 'connect4'
                                ? 'Connect Four'
                                : room.type === 'othello'
                                  ? 'Othello'
                                  : room.type === 'checkers'
                                    ? 'Checkers'
                                    : 'Battleship'}
                            </span>
                            <span className='text-sm font-bold truncate max-w-30'>
                              {room.status === 'active'
                                ? 'In progress'
                                : 'Waiting for opponent'}
                            </span>
                          </div>
                          <Button
                            size='sm'
                            variant='ghost'
                            className='h-8 text-[10px] font-black uppercase'
                            onClick={() =>
                              navigate(`/games/${room.type}/${room.roomId}`)
                            }
                          >
                            Resume
                          </Button>
                        </div>
                      ))}
                    </div>
                  )}

                  {myPendingRooms.length > 0 && (
                    <div className='space-y-3'>
                      <p className='text-[10px] font-black uppercase tracking-wider text-muted-foreground'>
                        Your Open Rooms
                      </p>
                      {myPendingRooms.map((room: GameRoom) => (
                        <div
                          key={room.id}
                          className='flex items-center justify-between p-3 bg-primary/5 border border-primary/30 rounded-xl'
                        >
                          <div className='flex flex-col'>
                            <span className='text-xs font-black uppercase tracking-tighter text-primary'>
                              {room.type}
                            </span>
                            <span className='text-sm font-bold truncate max-w-30'>
                              Waiting for opponent
                            </span>
                          </div>
                          <div className='flex gap-2'>
                            <Button
                              size='sm'
                              variant='ghost'
                              className='h-8 text-[10px] font-black uppercase'
                              onClick={() =>
                                navigate(`/games/${room.type}/${room.id}`)
                              }
                            >
                              Open
                            </Button>
                            <Button
                              size='sm'
                              variant='destructive'
                              className='h-8 text-[10px] font-black uppercase'
                              onClick={() => void closeRoom(room.id)}
                            >
                              Close
                            </Button>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}

                  {joinableRooms.length > 0 && (
                    <div className='space-y-3'>
                      <p className='text-[10px] font-black uppercase tracking-wider text-muted-foreground'>
                        Joinable Rooms
                      </p>
                      {joinableRooms.map((room: GameRoom) => (
                        <div
                          key={room.id}
                          className='flex items-center justify-between p-3 bg-muted/20 border rounded-xl hover:bg-muted/30 transition-colors'
                        >
                          <div className='flex flex-col'>
                            <span className='text-xs font-black uppercase tracking-tighter text-primary'>
                              {room.type}
                            </span>
                            <span className='text-sm font-bold truncate max-w-30'>
                              {room.creator?.username ?? 'Deleted User'}'s Room
                            </span>
                          </div>
                          <Button
                            size='sm'
                            variant='ghost'
                            className='h-8 text-[10px] font-black uppercase'
                            onClick={() =>
                              navigate(`/games/${room.type}/${room.id}`)
                            }
                          >
                            Join
                          </Button>
                        </div>
                      ))}
                    </div>
                  )}

                  {resumableRooms.length === 0 &&
                    myPendingRooms.length === 0 &&
                    joinableRooms.length === 0 && (
                      <div
                        key='no-active-rooms'
                        className='bg-card/50 p-2 rounded-lg border flex flex-col items-center'
                      >
                        <p className='text-[10px] font-bold uppercase text-muted-foreground'>
                          No Active Rooms
                        </p>
                        <p className='text-xs font-bold'>Start a new game!</p>
                      </div>
                    )}
                </div>
              )}
            </section>

            <section className='bg-primary text-primary-foreground rounded-2xl p-6 shadow-lg shadow-primary/20 relative overflow-hidden'>
              <div className='relative z-10'>
                <h3 className='text-lg font-black uppercase italic tracking-tighter mb-4 flex items-center gap-2'>
                  <UserCircle className='w-5 h-5' /> Your Identity
                </h3>
                <div className='space-y-4'>
                  <div className='flex justify-between items-center bg-white/10 p-3 rounded-xl'>
                    <span className='text-xs font-bold uppercase opacity-80'>
                      Rank
                    </span>
                    <span className='font-black italic text-xl'>#42</span>
                  </div>
                  <div className='flex justify-between items-center bg-white/10 p-3 rounded-xl'>
                    <span className='text-xs font-bold uppercase opacity-80'>
                      Points
                    </span>
                    <span className='font-black italic text-xl'>1,250</span>
                  </div>
                </div>
              </div>
              <div className='absolute -bottom-6 -right-6 w-24 h-24 bg-white/5 rounded-full' />
            </section>
          </div>
        </div>
      </main>
    </div>
  )
}
