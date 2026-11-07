package terminal

import (
	"testing"
	"time"
)

func TestTicketSingleUse(t *testing.T) {
	store := NewTicketStore()
	spec := SessionSpec{UserID: "u1", ServiceID: "s1", ContainerID: "c1", Mode: "shell"}
	token, _, err := store.Issue(spec)
	if err != nil {
		t.Fatal(err)
	}
	got, err := store.Redeem(token)
	if err != nil {
		t.Fatal(err)
	}
	if got.ContainerID != "c1" {
		t.Fatalf("got %+v", got)
	}
	if _, err := store.Redeem(token); err != ErrTicketNotFound {
		t.Fatalf("want not found after consume, got %v", err)
	}
}

func TestTicketPeekBeforeRedeem(t *testing.T) {
	store := NewTicketStore()
	spec := SessionSpec{UserID: "u1", ServiceID: "s1", ContainerID: "c1"}
	token, _, err := store.Issue(spec)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Peek(token); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Peek(token); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Redeem(token); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Peek(token); err != ErrTicketNotFound {
		t.Fatalf("want not found after redeem, got %v", err)
	}
}

func TestTicketExpiry(t *testing.T) {
	store := NewTicketStore()
	store.mu.Lock()
	expired := time.Now().Add(-time.Second)
	store.records["dead"] = ticketRecord{
		spec:      SessionSpec{UserID: "u1"},
		expiresAt: expired,
	}
	store.mu.Unlock()
	if _, err := store.Redeem("dead"); err != ErrTicketExpired {
		t.Fatalf("got %v", err)
	}
}
