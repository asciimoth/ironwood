package network_test

import (
	"crypto/ed25519"
	"net"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/Arceliar/ironwood/network"
)

func TestGetPeersDuringSignalExchange(t *testing.T) {
	pubA, keyA, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	pubB, keyB, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	a, err := network.NewPacketConn(keyA)
	if err != nil {
		t.Fatal(err)
	}
	b, err := network.NewPacketConn(keyB)
	if err != nil {
		_ = a.Close()
		t.Fatal(err)
	}
	connA, connB := net.Pipe()
	var handlers sync.WaitGroup
	handlers.Add(2)
	go func() {
		defer handlers.Done()
		_ = a.HandleConn(pubB, connA, 0)
	}()
	go func() {
		defer handlers.Done()
		_ = b.HandleConn(pubA, connB, 0)
	}()
	defer func() {
		_ = a.Close()
		_ = b.Close()
		handlers.Wait()
	}()

	// Read peer status while the two peers exchange routing signatures.
	seenA, seenB := false, false
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		for _, peer := range a.Debug.GetPeers() {
			seenA = seenA || peer.Latency > 0
		}
		for _, peer := range b.Debug.GetPeers() {
			seenB = seenB || peer.Latency > 0
		}
		runtime.Gosched()
	}
	if !seenA || !seenB {
		t.Fatalf("signal response latency was not reported: a=%t b=%t", seenA, seenB)
	}
}
