package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"connectrpc.com/connect"
	servicepb "github.com/shortlink-org/shop/admin-graphql/pkg/generated/service/v1"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestQueryGoodsPropagatesHeaders(t *testing.T) {
	t.Parallel()

	var receivedAuth string
	var receivedCookie string
	var receivedUserID string

	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			receivedAuth = req.Header.Get(authHeader)
			receivedCookie = req.Header.Get(cookieHeader)
			receivedUserID = req.Header.Get(userIDHeader)
			body := `{"count":1,"next":null,"previous":null,"results":[{"id":"1","name":"A","price":"1.00","description":"D","created_at":"2026-03-08T00:00:00Z","updated_at":"2026-03-08T00:00:00Z"}]}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	svc := New(nil, "http://admin.example", client)
	req := connect.NewRequest(&servicepb.QueryGoodsRequest{})
	req.Header().Set(authHeader, "Bearer token")
	req.Header().Set(cookieHeader, "sid=1")
	req.Header().Set(userIDHeader, "user-1")

	res, err := svc.QueryGoods(context.Background(), req)
	if err != nil {
		t.Fatalf("QueryGoods returned error: %v", err)
	}

	if receivedAuth != "Bearer token" || receivedCookie != "sid=1" || receivedUserID != "user-1" {
		t.Fatalf("headers were not propagated correctly: auth=%q cookie=%q user=%q", receivedAuth, receivedCookie, receivedUserID)
	}

	if res.Msg.GetGoods().GetCount() != 1 {
		t.Fatalf("expected count 1, got %d", res.Msg.GetGoods().GetCount())
	}
}

func TestQueryGoodRequiresID(t *testing.T) {
	t.Parallel()

	svc := New(nil, "http://admin.example", &http.Client{})
	_, err := svc.QueryGood(context.Background(), connect.NewRequest(&servicepb.QueryGoodRequest{}))
	if err == nil {
		t.Fatal("expected invalid argument error")
	}
	if code := connect.CodeOf(err); code != connect.CodeInvalidArgument {
		t.Fatalf("expected invalid argument, got %v", code)
	}
}
