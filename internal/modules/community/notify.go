package community

import (
	"context"
	"strings"

	"treeblog/internal/modules/notification"
)

// SetNotifier wires the notification service for fan-out. Notifications are
// best-effort: a nil notifier disables fan-out without affecting actions.
func (s *Service) SetNotifier(n *notification.Service) *Service {
	s.notifier = n
	return s
}

func truncateRunes(text string, max int) string {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= max {
		return string(runes)
	}
	return string(runes[:max]) + "…"
}

func (s *Service) postPublicPath(ctx context.Context, postID string) (authorID, handle, slug, title string, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT p.author_user_id, b.handle, p.slug, p.title
		FROM posts p JOIN blogs b ON b.id = p.blog_id
		WHERE p.id = $1`, postID).Scan(&authorID, &handle, &slug, &title)
	return authorID, handle, slug, title, err
}

// notifyCommentCreated fans out to the post author and, for replies, the
// parent commenter. Pending comments notify the author as review requests.
func (s *Service) notifyCommentCreated(ctx context.Context, comment *Comment, commenterID, commenterName, status string) {
	if s.notifier == nil || comment == nil {
		return
	}
	authorID, handle, slug, title, err := s.postPublicPath(ctx, comment.PostID)
	if err != nil {
		return
	}
	if commenterName == "" {
		commenterName = "누군가"
	}
	action := "/b/" + handle + "/" + slug + "#comment-" + comment.ID
	if authorID != commenterID {
		subject := "새 댓글: " + truncateRunes(title, 60)
		body := commenterName + ": " + truncateRunes(comment.ContentText, 120)
		if status == "PENDING_APPROVAL" {
			subject = "승인 대기 댓글: " + truncateRunes(title, 60)
		}
		_, _ = s.notifier.Create(ctx, authorID, "COMMENT",
			"COMMENT:"+comment.PostID, "COMMENT:"+comment.ID,
			subject, body, action)
	}
	if comment.ParentID != nil && *comment.ParentID != "" {
		var parentAuthor string
		if err := s.pool.QueryRow(ctx, `SELECT author_user_id FROM comments WHERE id = $1`, *comment.ParentID).Scan(&parentAuthor); err != nil {
			return
		}
		if parentAuthor != commenterID && parentAuthor != authorID {
			_, _ = s.notifier.Create(ctx, parentAuthor, "COMMENT",
				"COMMENT:"+comment.PostID, "REPLY:"+comment.ID,
				commenterName+"님의 답글",
				truncateRunes(comment.ContentText, 120), action)
		}
	}
}

// notifyReacted tells the post author about a first-time reaction.
func (s *Service) notifyReacted(ctx context.Context, postID, reactorID, reactorName, reactionType string) {
	if s.notifier == nil {
		return
	}
	authorID, handle, slug, title, err := s.postPublicPath(ctx, postID)
	if err != nil || authorID == reactorID {
		return
	}
	if reactorName == "" {
		reactorName = "누군가"
	}
	_, _ = s.notifier.Create(ctx, authorID, "REACTION",
		"REACTION:"+postID, "REACTION:"+postID+":"+reactorID,
		reactorName+"님의 공감: "+truncateRunes(title, 60),
		reactorName+"님이 회원님의 글에 "+reactionType+" 공감을 남겼습니다.",
		"/b/"+handle+"/"+slug)
}

// notifyBlogFollowed tells the blog owner about a new follower/subscriber.
func (s *Service) notifyBlogFollowed(ctx context.Context, blogID, actorID, actorName, category, verb string) {
	if s.notifier == nil {
		return
	}
	var ownerID, handle, title string
	if err := s.pool.QueryRow(ctx, `SELECT owner_user_id, handle, title FROM blogs WHERE id = $1`, blogID).Scan(&ownerID, &handle, &title); err != nil {
		return
	}
	if ownerID == actorID {
		return
	}
	if actorName == "" {
		actorName = "누군가"
	}
	_, _ = s.notifier.Create(ctx, ownerID, category,
		category+":"+blogID, category+":"+blogID+":"+actorID,
		actorName+"님이 "+truncateRunes(title, 60)+" 블로그를 "+verb+"했습니다.",
		actorName+"님이 새 "+verb+" 알림을 보냈습니다.",
		"/b/"+handle)
}
