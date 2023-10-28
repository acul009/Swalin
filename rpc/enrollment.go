package rpc

import (
	"context"
	"fmt"
	"rahnit-rmm/pki"
	"sync"
	"time"
)

type enrollmentManager struct {
	waitingEnrollments map[string]*enrollmentConnection
	mutex              sync.Mutex
}

type enrollmentConnection struct {
	connection *RpcConnection
	session    *RpcSession
	enrollment Enrollment
	mutex      sync.Mutex
}

type Enrollment struct {
	PublicKey   *pki.PublicKey
	Addr        string
	RequestTime time.Time
}

const maxEnrollmentTime = 5 * time.Minute

func newEnrollmentManager() *enrollmentManager {
	return &enrollmentManager{
		waitingEnrollments: make(map[string]*enrollmentConnection),
		mutex:              sync.Mutex{},
	}
}

func (m *enrollmentManager) cleanup() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for _, econn := range m.waitingEnrollments {
		if econn.mutex.TryLock() && time.Since(econn.enrollment.RequestTime) > maxEnrollmentTime {
			econn.connection.Close(408, "enrollment timed out")
			econn.mutex.Unlock()
		}
	}
}

func (m *enrollmentManager) startEnrollment(conn *RpcConnection) error {
	session, err := conn.AcceptSession(context.Background())
	if err != nil {
		conn.Close(500, "error accepting session")
		return fmt.Errorf("error accepting QUIC session: %w", err)
	}

	conn.connection.ConnectionState()

	err = exchangeKeys(session)
	if err != nil {
		return fmt.Errorf("error exchanging keys: %w", err)
	}

	encodedKey, err := session.partner.Base64Encode()
	if err != nil {
		return fmt.Errorf("error encoding key: %w", err)
	}

	m.mutex.Lock()
	_, ok := m.waitingEnrollments[encodedKey]
	if ok {
		return fmt.Errorf("enrollment already in progress")
	}

	m.waitingEnrollments[encodedKey] = &enrollmentConnection{
		connection: conn,
		session:    session,
		enrollment: Enrollment{
			PublicKey:   session.partner,
			Addr:        conn.connection.RemoteAddr().String(),
			RequestTime: time.Now(),
		},
	}
	m.mutex.Unlock()

	return nil
}

func (m *enrollmentManager) acceptEnrollment(cert *pki.Certificate) error {
	m.cleanup()
	encodedKey, err := cert.GetPublicKey().Base64Encode()
	if err != nil {
		return fmt.Errorf("error encoding key: %w", err)
	}

	m.mutex.Lock()
	econn, ok := m.waitingEnrollments[encodedKey]
	m.mutex.Unlock()

	econn.mutex.Lock()
	defer econn.mutex.Unlock()

	if !ok {
		return fmt.Errorf("enrollment not in progress")
	}

	defer func() {
		m.mutex.Lock()
		delete(m.waitingEnrollments, encodedKey)
		m.mutex.Unlock()
	}()

	err = WriteMessage[*pki.Certificate](econn.session, cert)
	if err != nil {
		econn.connection.Close(500, "error writing certificate")
		return fmt.Errorf("error writing certificate: %w", err)
	}

	econn.session.Close()
	econn.connection.Close(200, "enrollment complete")

	return nil
}

func (m *enrollmentManager) list() []Enrollment {
	m.cleanup()
	m.mutex.Lock()
	defer m.mutex.Unlock()

	list := make([]Enrollment, 0, len(m.waitingEnrollments))
	for _, econn := range m.waitingEnrollments {
		if econn.mutex.TryLock() {
			list = append(list, econn.enrollment)
			econn.mutex.Unlock()
		}
	}

	return list
}
