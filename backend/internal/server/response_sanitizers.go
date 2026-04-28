package server

import "sanctum/internal/models"

func sanitizeSharedUser(user *models.User) {
	if user == nil {
		return
	}

	posts := user.Posts
	sanitizeSharedPostValues(posts)
	*user = models.User{
		ID:       user.ID,
		Username: user.Username,
		Avatar:   user.Avatar,
		Posts:    posts,
	}
}

func sanitizeSharedUsers(users []models.User) {
	for i := range users {
		sanitizeSharedUser(&users[i])
	}
}

func sanitizeSharedPost(post *models.Post) {
	if post == nil {
		return
	}

	sanitizeSharedUser(&post.User)
}

func sanitizeSharedPosts(posts []*models.Post) {
	for _, post := range posts {
		sanitizeSharedPost(post)
	}
}

func sanitizeSharedPostValues(posts []models.Post) {
	for i := range posts {
		sanitizeSharedPost(&posts[i])
	}
}

func sanitizeSharedComment(comment *models.Comment) {
	if comment == nil {
		return
	}

	sanitizeSharedUser(&comment.User)
	sanitizeSharedPost(&comment.Post)
}

func sanitizeSharedComments(comments []*models.Comment) {
	for _, comment := range comments {
		sanitizeSharedComment(comment)
	}
}

func sanitizeSharedFriendship(friendship *models.Friendship) {
	if friendship == nil {
		return
	}

	sanitizeSharedUser(&friendship.Requester)
	sanitizeSharedUser(&friendship.Addressee)
}

func sanitizeSharedFriendships(friendships []models.Friendship) {
	for i := range friendships {
		sanitizeSharedFriendship(&friendships[i])
	}
}

func sanitizeSharedConversation(conversation *models.Conversation) {
	if conversation == nil {
		return
	}

	sanitizeSharedUsers(conversation.Participants)
	sanitizeSharedMessageValues(conversation.Messages)
}

func sanitizeSharedConversations(conversations []*models.Conversation) {
	for _, conversation := range conversations {
		sanitizeSharedConversation(conversation)
	}
}

func sanitizeSharedMessage(message *models.Message) {
	if message == nil {
		return
	}

	sanitizeSharedUser(message.Sender)
}

func sanitizeSharedMessages(messages []*models.Message) {
	for _, message := range messages {
		sanitizeSharedMessage(message)
	}
}

func sanitizeSharedMessageValues(messages []models.Message) {
	for i := range messages {
		sanitizeSharedMessage(&messages[i])
	}
}

func sanitizeSharedUserBlocks(blocks []models.UserBlock) {
	for i := range blocks {
		sanitizeSharedUser(blocks[i].Blocker)
		sanitizeSharedUser(blocks[i].Blocked)
	}
}

func sanitizeSharedMentions(mentions []models.MessageMention) {
	for i := range mentions {
		sanitizeSharedMessage(mentions[i].Message)
		sanitizeSharedUser(mentions[i].MentionedUser)
		sanitizeSharedUser(mentions[i].MentionedByUser)
		sanitizeSharedConversation(mentions[i].Conversation)
	}
}

func sanitizeSharedChatroomBans(bans []models.ChatroomBan) {
	for i := range bans {
		sanitizeSharedUser(bans[i].User)
		sanitizeSharedUser(bans[i].BannedByUser)
	}
}

func sanitizeSharedChatroomMutes(mutes []models.ChatroomMute) {
	for i := range mutes {
		sanitizeSharedUser(mutes[i].User)
		sanitizeSharedUser(mutes[i].MutedByUser)
		sanitizeSharedConversation(mutes[i].Conversation)
	}
}

func sanitizeSharedSanctumRequests(requests []models.SanctumRequest) {
	for i := range requests {
		sanitizeSharedUser(requests[i].RequestedByUser)
		sanitizeSharedUser(requests[i].ReviewedByUser)
	}
}

func sanitizeSharedGameRoom(room *models.GameRoom) {
	if room == nil {
		return
	}

	sanitizeSharedUser(&room.Creator)
	sanitizeSharedUser(&room.Opponent)
	sanitizeSharedUser(&room.Winner)
}

func sanitizeSharedGameRooms(rooms []models.GameRoom) {
	for i := range rooms {
		sanitizeSharedGameRoom(&rooms[i])
	}
}
