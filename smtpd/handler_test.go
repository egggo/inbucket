package smtpd

import (
	"bytes"
	"fmt"
	"io"

	"github.com/egggo/inbucket/config"
	//"io/ioutil"
	"log"
	"net"
	"net/textproto"
	"os"
	"testing"
	"time"
)

type scriptStep struct {
	send   string
	expect int
}

// Test commands in GREET state
func TestGreetState(t *testing.T) {
	// Setup mock objects
	mds := &MockDataStore{}
	mb1 := &MockMailbox{}
	mds.On("MailboxFor").Return(mb1, nil)

	server, logbuf := setupSmtpServer(mds)
	defer teardownSmtpServer(server)

	var script []scriptStep

	// Test out some mangled HELOs
	script = []scriptStep{
		{"HELO", 501},
		{"EHLO", 501},
		{"HELLO", 500},
		{"HELL", 500},
		{"hello", 500},
		{"Outlook", 500},
	}
	if err := playSession(t, server, script); err != nil {
		t.Error(err)
	}

	// Valid HELOs
	if err := playSession(t, server, []scriptStep{{"HELO mydomain", 250}}); err != nil {
		t.Error(err)
	}
	if err := playSession(t, server, []scriptStep{{"HELO mydom.com", 250}}); err != nil {
		t.Error(err)
	}
	if err := playSession(t, server, []scriptStep{{"HelO mydom.com", 250}}); err != nil {
		t.Error(err)
	}
	if err := playSession(t, server, []scriptStep{{"helo 127.0.0.1", 250}}); err != nil {
		t.Error(err)
	}

	// Valid EHLOs
	if err := playSession(t, server, []scriptStep{{"EHLO mydomain", 250}}); err != nil {
		t.Error(err)
	}
	if err := playSession(t, server, []scriptStep{{"EHLO mydom.com", 250}}); err != nil {
		t.Error(err)
	}
	if err := playSession(t, server, []scriptStep{{"EhlO mydom.com", 250}}); err != nil {
		t.Error(err)
	}
	if err := playSession(t, server, []scriptStep{{"ehlo 127.0.0.1", 250}}); err != nil {
		t.Error(err)
	}

	if t.Failed() {
		// Wait for handler to finish logging
		time.Sleep(2 * time.Second)
		// Dump buffered log data if there was a failure
		io.Copy(os.Stderr, logbuf)
	}
}

// Test commands in READY state
func TestReadyState(t *testing.T) {
	// Setup mock objects
	mds := &MockDataStore{}
	mb1 := &MockMailbox{}
	mds.On("MailboxFor").Return(mb1, nil)

	server, logbuf := setupSmtpServer(mds)
	defer teardownSmtpServer(server)

	var script []scriptStep

	// Test out some mangled READY commands
	script = []scriptStep{
		{"HELO localhost", 250},
		{"FOOB", 500},
		{"HELO", 503},
		{"DATA", 503},
		{"MAIL", 501},
		{"MAIL FROM john@gmail.com", 501},
		{"MAIL FROM:john@gmail.com", 501},
		{"MAIL FROM:<john@gmail.com> SIZE=147KB", 501},
		{"MAIL FROM: <john@gmail.com> SIZE147", 501},
		{"MAIL FROM:<first@last@gmail.com>", 501},
		{"MAIL FROM:<first last@gmail.com>", 501},
	}
	if err := playSession(t, server, script); err != nil {
		t.Error(err)
	}

	// Test out some valid MAIL commands
	script = []scriptStep{
		{"HELO localhost", 250},
		{"MAIL FROM:<john@gmail.com>", 250},
		{"RSET", 250},
		{"MAIL FROM: <john@gmail.com>", 250},
		{"RSET", 250},
		{"MAIL FROM: <john@gmail.com> BODY=8BITMIME", 250},
		{"RSET", 250},
		{"MAIL FROM:<john@gmail.com> SIZE=1024", 250},
		{"RSET", 250},
		{"MAIL FROM:<host!host!user/data@foo.com>", 250},
		{"RSET", 250},
		{"MAIL FROM:<\"first last\"@space.com>", 250},
		{"RSET", 250},
		{"MAIL FROM:<user\\@internal@external.com>", 250},
		{"RSET", 250},
		{"MAIL FROM:<user\\>name@host.com>", 250},
		{"RSET", 250},
		{"MAIL FROM:<\"user>name\"@host.com>", 250},
		{"RSET", 250},
		{"MAIL FROM:<\"user@internal\"@external.com>", 250},
	}
	if err := playSession(t, server, script); err != nil {
		t.Error(err)
	}

	if t.Failed() {
		// Wait for handler to finish logging
		time.Sleep(2 * time.Second)
		// Dump buffered log data if there was a failure
		io.Copy(os.Stderr, logbuf)
	}
}

// Test commands in MAIL state
func TestMailState(t *testing.T) {
	// Setup mock objects
	mds := &MockDataStore{}
	mb1 := &MockMailbox{}
	msg1 := &MockMessage{}
	mds.On("MailboxFor").Return(mb1, nil)
	mb1.On("NewMessage").Return(msg1, nil)
	msg1.On("Close").Return(nil)

	server, logbuf := setupSmtpServer(mds)
	defer teardownSmtpServer(server)

	var script []scriptStep

	// Test out some mangled READY commands
	script = []scriptStep{
		{"HELO localhost", 250},
		{"MAIL FROM:<john@gmail.com>", 250},
		{"FOOB", 500},
		{"HELO", 503},
		{"DATA", 503},
		{"MAIL", 503},
		{"RCPT", 501},
		{"RCPT TO", 501},
		{"RCPT TO james@gmail.com", 501},
		{"RCPT TO:<first last@host.com>", 501},
		{"RCPT TO:<fred@fish@host.com", 501},
	}
	if err := playSession(t, server, script); err != nil {
		t.Error(err)
	}

	// Test out some good RCPT commands
	script = []scriptStep{
		{"HELO localhost", 250},
		{"MAIL FROM:<john@gmail.com>", 250},
		{"RCPT TO:<u1@gmail.com>", 250},
		{"RCPT TO: <u2@gmail.com>", 250},
		{"RCPT TO:u3@gmail.com", 250},
		{"RCPT TO: u4@gmail.com", 250},
		{"RSET", 250},
		{"MAIL FROM:<john@gmail.com>", 250},
		{"RCPT TO:<user\\@internal@external.com", 250},
		{"RCPT TO:<\"first last\"@host.com", 250},
		{"RCPT TO:<user\\>name@host.com>", 250},
		{"RCPT TO:<\"user>name\"@host.com>", 250},
	}
	if err := playSession(t, server, script); err != nil {
		t.Error(err)
	}

	// Test out recipient limit
	script = []scriptStep{
		{"HELO localhost", 250},
		{"MAIL FROM:<john@gmail.com>", 250},
		{"RCPT TO:<u1@gmail.com>", 250},
		{"RCPT TO:<u2@gmail.com>", 250},
		{"RCPT TO:<u3@gmail.com>", 250},
		{"RCPT TO:<u4@gmail.com>", 250},
		{"RCPT TO:<u5@gmail.com>", 250},
		{"RCPT TO:<u6@gmail.com>", 552},
	}
	if err := playSession(t, server, script); err != nil {
		t.Error(err)
	}

	// Test DATA
	script = []scriptStep{
		{"HELO localhost", 250},
		{"MAIL FROM:<john@gmail.com>", 250},
		{"RCPT TO:<u1@gmail.com>", 250},
		{"DATA", 354},
		{".", 250},
	}
	if err := playSession(t, server, script); err != nil {
		t.Error(err)
	}

	// Test RSET
	script = []scriptStep{
		{"HELO localhost", 250},
		{"MAIL FROM:<john@gmail.com>", 250},
		{"RCPT TO:<u1@gmail.com>", 250},
		{"RSET", 250},
		{"MAIL FROM:<john@gmail.com>", 250},
	}
	if err := playSession(t, server, script); err != nil {
		t.Error(err)
	}

	// Test QUIT
	script = []scriptStep{
		{"HELO localhost", 250},
		{"MAIL FROM:<john@gmail.com>", 250},
		{"RCPT TO:<u1@gmail.com>", 250},
		{"QUIT", 221},
	}
	if err := playSession(t, server, script); err != nil {
		t.Error(err)
	}

	if t.Failed() {
		// Wait for handler to finish logging
		time.Sleep(2 * time.Second)
		// Dump buffered log data if there was a failure
		io.Copy(os.Stderr, logbuf)
	}
}

// Test commands in DATA state
func TestDataState(t *testing.T) {
	// Setup mock objects
	mds := &MockDataStore{}
	mb1 := &MockMailbox{}
	msg1 := &MockMessage{}
	mds.On("MailboxFor").Return(mb1, nil)
	mb1.On("NewMessage").Return(msg1, nil)
	msg1.On("Close").Return(nil)

	server, logbuf := setupSmtpServer(mds)
	defer teardownSmtpServer(server)

	var script []scriptStep
	pipe := setupSmtpSession(server)
	c := textproto.NewConn(pipe)

	// Get us into DATA state
	if code, _, err := c.ReadCodeLine(220); err != nil {
		t.Errorf("Expected a 220 greeting, got %v", code)
	}
	script = []scriptStep{
		{"HELO localhost", 250},
		{"MAIL FROM:<john@gmail.com>", 250},
		{"RCPT TO:<u1@gmail.com>", 250},
		{"DATA", 354},
	}
	if err := playScriptAgainst(t, c, script); err != nil {
		t.Error(err)
	}
	// Send a message
	body := `To: u1@gmail.com
From: john@gmail.com
Subject: test

Hi!
`
	dw := c.DotWriter()
	io.WriteString(dw, body)
	dw.Close()
	if code, _, err := c.ReadCodeLine(250); err != nil {
		t.Errorf("Expected a 250 greeting, got %v", code)
	}

	if t.Failed() {
		// Wait for handler to finish logging
		time.Sleep(2 * time.Second)
		// Dump buffered log data if there was a failure
		io.Copy(os.Stderr, logbuf)
	}
}

// playSession creates a new session, reads the greeting and then plays the script
func playSession(t *testing.T, server *Server, script []scriptStep) error {
	pipe := setupSmtpSession(server)
	c := textproto.NewConn(pipe)

	if code, _, err := c.ReadCodeLine(220); err != nil {
		return fmt.Errorf("Expected a 220 greeting, got %v", code)
	}

	err := playScriptAgainst(t, c, script)

	c.Cmd("QUIT")
	c.ReadCodeLine(221)

	return err
}

// playScriptAgainst an existing connection, does not handle server greeting
func playScriptAgainst(t *testing.T, c *textproto.Conn, script []scriptStep) error {
	for i, step := range script {
		id, err := c.Cmd(step.send)
		if err != nil {
			return fmt.Errorf("Step %d, failed to send %q: %v", i, step.send, err)
		}

		c.StartResponse(id)
		code, msg, err := c.ReadResponse(step.expect)
		if err != nil {
			err = fmt.Errorf("Step %d, sent %q, expected %v, got %v: %q",
				i, step.send, step.expect, code, msg)
		}
		c.EndResponse(id)

		if err != nil {
			// Return after c.EndResponse so we don't hang the connection
			return err
		}
	}
	return nil
}

// net.Pipe does not implement deadlines
type mockConn struct {
	net.Conn
}

func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func setupSmtpServer(ds DataStore) (*Server, *bytes.Buffer) {
	// Test Server Config
	cfg := config.SmtpConfig{
		Ip4address:      net.IPv4(127, 0, 0, 1),
		Ip4port:         2500,
		Domain:          "inbucket.local",
		DomainNoStore:   "bitbucket.local",
		MaxRecipients:   5,
		MaxIdleSeconds:  5,
		MaxMessageBytes: 5000,
		StoreMessages:   true,
	}

	// Capture log output
	buf := new(bytes.Buffer)
	log.SetOutput(buf)

	// Create a server, don't start it
	return NewSmtpServer(cfg, ds), buf
}

var sessionNum int

func setupSmtpSession(server *Server) net.Conn {
	// Pair of pipes to communicate
	serverConn, clientConn := net.Pipe()
	// Start the session
	server.waitgroup.Add(1)
	sessionNum++
	go server.startSession(sessionNum, &mockConn{serverConn})

	return clientConn
}

func teardownSmtpServer(server *Server) {
	//log.SetOutput(os.Stderr)
}
