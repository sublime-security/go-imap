package server_test

import (
	"bufio"
	"net"
	"testing"
	"time"

	"github.com/emersion/go-imap/backend/memory"
	"github.com/emersion/go-imap/server"
)

// Extnesions that are always advertised by go-imap server.
const builtinExtensions = "LITERAL+ SASL-IR CHILDREN UNSELECT MOVE IDLE APPENDLIMIT"

func newTestBackend(t *testing.T) *memory.Backend {
	bkd := memory.New()

	user, err := bkd.AddUser("username", "password")
	if err != nil {
		t.Fatal("Cannot add user:", err)
	}
	if err := user.CreateMailbox("INBOX"); err != nil {
		t.Fatal("Cannot create INBOX:", err)
	}
	mbox, err := user.GetMailbox("INBOX")
	if err != nil {
		t.Fatal("Cannot get INBOX:", err)
	}

	body := "From: contact@example.org\r\n" +
		"To: contact@example.org\r\n" +
		"Subject: A little message, just for you\r\n" +
		"Date: Wed, 11 May 2016 14:31:59 +0000\r\n" +
		"Message-ID: <0000000@localhost/>\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"Hi there :)"
	mbox.(*memory.Mailbox).Messages = []*memory.Message{{
		Uid:   6,
		Date:  time.Now(),
		Flags: []string{"\\Seen"},
		Size:  uint32(len(body)),
		Body:  []byte(body),
	}}

	return bkd
}

func testServer(t *testing.T) (s *server.Server, conn net.Conn) {
	bkd := newTestBackend(t)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal("Cannot listen:", err)
	}

	s = server.New(bkd)
	s.AllowInsecureAuth = true

	go s.Serve(l)

	conn, err = net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal("Cannot connect to server:", err)
	}

	return
}

func TestServer_greeting(t *testing.T) {
	s, conn := testServer(t)
	defer s.Close()
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	scanner.Scan() // Wait for greeting
	greeting := scanner.Text()

	if greeting != "* OK [CAPABILITY IMAP4rev1 "+builtinExtensions+" AUTH=PLAIN] IMAP4rev1 Service Ready" {
		t.Fatal("Bad greeting:", greeting)
	}
}
