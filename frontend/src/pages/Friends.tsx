import { useQuery } from "@tanstack/react-query";
import { Loader2, UserPlus, Users as UsersIcon } from "lucide-react";
import { useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { apiClient } from "@/api/client";
import type { User } from "@/api/types";
import { type FriendAction, FriendCard } from "@/components/friends/FriendCard";
import { FriendSidebar } from "@/components/friends/FriendSidebar";
import {
  useAcceptFriendRequest,
  useFriends,
  usePendingRequests,
  useRejectFriendRequest,
  useRemoveFriend,
  useSendFriendRequest,
  useSentRequests,
} from "@/hooks/useFriends";
import { getCurrentUser } from "@/hooks/useUsers";

type ViewType = "suggestions" | "requests" | "friends";

export default function Friends() {
  const [searchParams] = useSearchParams();
  const initialView: ViewType = searchParams.get("tab") === "find" ? "suggestions" : "suggestions";
  const [activeView, setActiveView] = useState<ViewType>(initialView);
  const currentUser = getCurrentUser();

  const { data: allUsers = [], isLoading: isLoadingUsers } = useQuery({
    queryKey: ["users"],
    queryFn: () => apiClient.getUsers({ limit: 100 }),
  });

  const { data: friends = [], isLoading: isLoadingFriends } = useFriends();
  const { data: incomingRequests = [], isLoading: isLoadingIncoming } = usePendingRequests();
  const { data: outgoingRequests = [], isLoading: isLoadingOutgoing } = useSentRequests();

  const sendRequest = useSendFriendRequest();
  const acceptRequest = useAcceptFriendRequest();
  const rejectRequest = useRejectFriendRequest();
  const removeFriend = useRemoveFriend();

  const suggestions = useMemo(() => {
    if (!currentUser) return [];
    const friendIds = new Set(friends.map((f) => f.id));
    const incomingIds = new Set(incomingRequests.map((r) => r.requester_id));
    const outgoingIds = new Set(outgoingRequests.map((r) => r.addressee_id));
    return allUsers.filter(
      (user) =>
        user.id !== currentUser.id &&
        !friendIds.has(user.id) &&
        !incomingIds.has(user.id) &&
        !outgoingIds.has(user.id),
    );
  }, [allUsers, friends, incomingRequests, outgoingRequests, currentUser]);

  const handleAction = (action: FriendAction, user: User) => {
    switch (action) {
      case "add":
        sendRequest.mutate(user.id);
        break;
      case "accept": {
        const req = incomingRequests.find((r) => r.requester_id === user.id);
        if (req) acceptRequest.mutate(req.id);
        break;
      }
      case "reject": {
        const req = incomingRequests.find((r) => r.requester_id === user.id);
        if (req) rejectRequest.mutate(req.id);
        break;
      }
      case "remove":
        removeFriend.mutate(user.id);
        break;
      case "cancel":
        removeFriend.mutate(user.id);
        break;
    }
  };

  const isLoading = isLoadingUsers || isLoadingFriends || isLoadingIncoming || isLoadingOutgoing;

  return (
    <div className="flex flex-col md:flex-row flex-1 overflow-hidden h-full">
      <FriendSidebar
        activeView={activeView}
        onViewChange={setActiveView}
        requestCount={incomingRequests.length}
      />

      <div className="flex-1 overflow-y-auto p-4 md:p-8">
        <div className="mb-6">
          <h1 className="text-2xl font-bold mb-1">
            {activeView === "requests"
              ? "Friend Requests"
              : activeView === "friends"
                ? "All Friends"
                : "People You May Know"}
          </h1>
          {activeView === "requests" && incomingRequests.length > 0 && (
            <p className="text-destructive font-medium text-sm">
              {incomingRequests.length} pending request
              {incomingRequests.length !== 1 ? "s" : ""}
            </p>
          )}
        </div>

        {isLoading ? (
          <div className="flex justify-center py-12">
            <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
          </div>
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
            {activeView === "suggestions" &&
              suggestions.map((user) => (
                <FriendCard key={user.id} user={user} actionType="add" onAction={handleAction} />
              ))}
            {activeView === "suggestions" && suggestions.length === 0 && (
              <div className="col-span-full py-12 text-center text-muted-foreground">
                No suggestions available.
              </div>
            )}

            {activeView === "requests" &&
              incomingRequests.map((req) =>
                req.requester ? (
                  <FriendCard
                    key={req.id}
                    user={req.requester}
                    actionType="accept_reject"
                    onAction={handleAction}
                  />
                ) : null,
              )}
            {activeView === "requests" && incomingRequests.length === 0 && (
              <div className="col-span-full py-12 text-center">
                <UsersIcon className="mx-auto mb-4 h-16 w-16 text-muted-foreground/30" />
                <p className="font-medium text-muted-foreground">No new friend requests</p>
              </div>
            )}

            {activeView === "friends" &&
              friends.map((user) => (
                <FriendCard key={user.id} user={user} actionType="remove" onAction={handleAction} />
              ))}
            {activeView === "friends" && friends.length === 0 && (
              <div className="col-span-full py-12 text-center">
                <UserPlus className="mx-auto mb-4 h-16 w-16 text-muted-foreground/30" />
                <p className="font-medium text-muted-foreground">
                  No friends yet. Check Suggestions to add some!
                </p>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
