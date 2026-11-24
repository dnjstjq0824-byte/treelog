package community

import (
	"context"
	"testing"
)

// TestNeighborRelationsFlow covers the G-17 state machine: requests,
// accept/reject, mutual promotion, groups, favorites, subscription pause,
// channel toggles, and the following feed.
func TestNeighborRelationsFlow(t *testing.T) {
	svc, cleanup := setupIntegration(t)
	defer cleanup()
	ctx := context.Background()
	authorAID, blogAID, _ := seedAuthor(t, svc, ctx, "nrauthora")
	authorBID, blogBID, _ := seedAuthor(t, svc, ctx, "nrauthorb")
	readerID := seedReader(t, svc, ctx, "nrreader")

	// Neighbor request lifecycle.
	if err := svc.RequestNeighbor(ctx, readerID, blogAID, "please"); err != nil {
		t.Fatalf("request: %v", err)
	}
	if err := svc.RequestNeighbor(ctx, authorAID, blogAID, ""); faultCode(err) != "CANNOT_NEIGHBOR_OWN_BLOG" {
		t.Fatalf("expected CANNOT_NEIGHBOR_OWN_BLOG, got %v", err)
	}
	requests, err := svc.ListNeighborRequests(ctx, authorAID, blogAID)
	if err != nil || len(requests) != 1 {
		t.Fatalf("requests: %+v / %v", requests, err)
	}
	if _, err := svc.ListNeighborRequests(ctx, readerID, blogAID); faultCode(err) != "NOT_BLOG_MODERATOR" {
		t.Fatalf("expected NOT_BLOG_MODERATOR, got %v", err)
	}
	if err := svc.RespondNeighborRequest(ctx, authorAID, requests[0].FollowID, true); err != nil {
		t.Fatalf("accept: %v", err)
	}
	var relation, status string
	if err := svc.pool.QueryRow(ctx, `SELECT relation_type, status FROM follows
		WHERE follower_user_id = $1 AND target_blog_id = $2`, readerID, blogAID).Scan(&relation, &status); err != nil {
		t.Fatal(err)
	}
	if relation != "NEIGHBOR" || status != "ACCEPTED" {
		t.Fatalf("relation = %s/%s", relation, status)
	}

	// Mutual promotion across both owners' blogs.
	if err := svc.RequestNeighbor(ctx, authorAID, blogBID, "hi"); err != nil {
		t.Fatalf("request b: %v", err)
	}
	requestsB, err := svc.ListNeighborRequests(ctx, authorBID, blogBID)
	if err != nil || len(requestsB) != 1 {
		t.Fatalf("requests b: %+v / %v", requestsB, err)
	}
	if err := svc.RequestNeighbor(ctx, authorBID, blogAID, "back"); err != nil {
		t.Fatalf("request a: %v", err)
	}
	requestsA, err := svc.ListNeighborRequests(ctx, authorAID, blogAID)
	if err != nil || len(requestsA) != 1 {
		t.Fatalf("requests a: %+v / %v", requestsA, err)
	}
	if err := svc.RespondNeighborRequest(ctx, authorBID, requestsB[0].FollowID, true); err != nil {
		t.Fatalf("accept b: %v", err)
	}
	if err := svc.RespondNeighborRequest(ctx, authorAID, requestsA[0].FollowID, true); err != nil {
		t.Fatalf("accept a: %v", err)
	}
	var mutual int
	if err := svc.pool.QueryRow(ctx, `SELECT COUNT(*) FROM follows
		WHERE relation_type = 'MUTUAL_NEIGHBOR' AND status = 'ACCEPTED'
		AND ((follower_user_id = $1 AND target_blog_id = $2) OR (follower_user_id = $3 AND target_blog_id = $4))`,
		authorAID, blogBID, authorBID, blogAID).Scan(&mutual); err != nil {
		t.Fatal(err)
	}
	if mutual != 2 {
		t.Fatalf("expected 2 mutual relations, got %d", mutual)
	}

	// Reject + re-request reopens.
	rejecterID := seedReader(t, svc, ctx, "nrreject")
	if err := svc.RequestNeighbor(ctx, rejecterID, blogAID, "x"); err != nil {
		t.Fatalf("request reject: %v", err)
	}
	pending, _ := svc.ListNeighborRequests(ctx, authorAID, blogAID)
	if len(pending) != 1 {
		t.Fatalf("pending: %+v", pending)
	}
	if err := svc.RespondNeighborRequest(ctx, authorAID, pending[0].FollowID, false); err != nil {
		t.Fatalf("reject: %v", err)
	}
	if err := svc.RequestNeighbor(ctx, rejecterID, blogAID, "again"); err != nil {
		t.Fatalf("re-request: %v", err)
	}

	// Groups + favorites.
	group, err := svc.CreateNeighborGroup(ctx, readerID, "친한 이웃")
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	if err := svc.MoveFollowToGroup(ctx, readerID, blogAID, &group.ID); err != nil {
		t.Fatalf("move: %v", err)
	}
	otherGroup, err := svc.CreateNeighborGroup(ctx, authorAID, "타인 폴더")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.MoveFollowToGroup(ctx, readerID, blogAID, &otherGroup.ID); faultCode(err) != "GROUP_NOT_FOUND" {
		t.Fatalf("expected GROUP_NOT_FOUND, got %v", err)
	}
	if err := svc.SetFollowFavorite(ctx, readerID, blogAID, true); err != nil {
		t.Fatalf("favorite: %v", err)
	}
	if err := svc.RenameNeighborGroup(ctx, readerID, group.ID, "절친"); err != nil {
		t.Fatalf("rename: %v", err)
	}
	if err := svc.DeleteNeighborGroup(ctx, readerID, group.ID); err != nil {
		t.Fatalf("delete group: %v", err)
	}

	// Subscriptions: pause, resume, channels.
	if err := svc.Subscribe(ctx, readerID, blogAID); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := svc.PauseSubscription(ctx, readerID, blogAID); err != nil {
		t.Fatalf("pause: %v", err)
	}
	if err := svc.PauseSubscription(ctx, readerID, blogAID); faultCode(err) != "SUBSCRIPTION_STATE_MISMATCH" {
		t.Fatalf("expected SUBSCRIPTION_STATE_MISMATCH, got %v", err)
	}
	emailOff := false
	if err := svc.UpdateSubscriptionChannels(ctx, readerID, blogAID, nil, nil, &emailOff); err != nil {
		t.Fatalf("channels: %v", err)
	}
	if err := svc.ResumeSubscription(ctx, readerID, blogAID); err != nil {
		t.Fatalf("resume: %v", err)
	}

	// Following feed spans follows + subscriptions, newest first.
	if err := svc.Follow(ctx, readerID, blogBID); err != nil {
		t.Fatalf("follow b: %v", err)
	}
	items, _, err := svc.FollowingFeed(ctx, readerID, 20, "")
	if err != nil {
		t.Fatalf("feed: %v", err)
	}
	if len(items) < 2 {
		t.Fatalf("feed must include both blogs' posts: %+v", items)
	}
	page, cursor, err := svc.FollowingFeed(ctx, readerID, 1, "")
	if err != nil || len(page) != 1 || cursor == "" {
		t.Fatalf("feed page: %+v %q / %v", page, cursor, err)
	}
	rest, _, err := svc.FollowingFeed(ctx, readerID, 10, cursor)
	if err != nil || len(rest) == 0 || rest[0].PostID == page[0].PostID {
		t.Fatalf("feed cursor broken: %+v / %v", rest, err)
	}
	if _, _, err := svc.FollowingFeed(ctx, readerID, 10, "bogus"); faultCode(err) != "INVALID_CURSOR" {
		t.Fatalf("expected INVALID_CURSOR, got %v", err)
	}
}
