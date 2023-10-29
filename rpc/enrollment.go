package rpc

import (
	"context"
	"fmt"
	"log"
	"rahnit-rmm/config"
	"rahnit-rmm/pki"
	"rahnit-rmm/util"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
)

type enrollmentManager struct {
	waitingEnrollments *util.ObservableMap[string, *enrollmentConnection]
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
		waitingEnrollments: util.NewObservableMap[string, *enrollmentConnection](),
		mutex:              sync.Mutex{},
	}
}

func (m *enrollmentManager) cleanup() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for key, econn := range m.waitingEnrollments.GetAll() {
		if econn.mutex.TryLock() && time.Since(econn.enrollment.RequestTime) > maxEnrollmentTime {
			econn.connection.Close(408, "enrollment timed out")
			econn.mutex.Unlock()
			m.waitingEnrollments.Delete(key)
		}
	}
}

func (m *enrollmentManager) startEnrollment(conn *RpcConnection) error {
	session, err := conn.AcceptSession(context.Background())
	if err != nil {
		conn.Close(500, "error accepting session")
		return fmt.Errorf("error accepting QUIC session: %w", err)
	}

	err = session.mutateState(RpcSessionCreated, RpcSessionOpen)
	if err != nil {
		conn.Close(500, "error setting session state")
		return fmt.Errorf("error setting session state: %w", err)
	}

	err = exchangeKeys(session)
	if err != nil {
		return fmt.Errorf("error exchanging keys: %w", err)
	}

	encodedKey, err := session.partner.Base64Encode()
	if err != nil {
		return fmt.Errorf("error encoding key: %w", err)
	}

	m.mutex.Lock()
	if m.waitingEnrollments.Has(encodedKey) {
		return fmt.Errorf("enrollment already in progress")
	}

	m.waitingEnrollments.Set(encodedKey,
		&enrollmentConnection{
			connection: conn,
			session:    session,
			enrollment: Enrollment{
				PublicKey:   session.partner,
				Addr:        conn.connection.RemoteAddr().String(),
				RequestTime: time.Now(),
			},
		},
	)
	m.mutex.Unlock()

	log.Printf("enrollment started for %s", encodedKey)

	return nil
}

func (m *enrollmentManager) acceptEnrollment(cert *pki.Certificate) error {
	m.cleanup()
	encodedKey, err := cert.GetPublicKey().Base64Encode()
	if err != nil {
		return fmt.Errorf("error encoding key: %w", err)
	}

	econn, ok := m.waitingEnrollments.Get(encodedKey)

	if !ok {
		return fmt.Errorf("enrollment not in progress")
	}

	econn.mutex.Lock()
	defer econn.mutex.Unlock()

	m.waitingEnrollments.Delete(encodedKey)

	err = WriteMessage[*pki.Certificate](econn.session, cert)
	if err != nil {
		econn.connection.Close(500, "error writing certificate")
		return fmt.Errorf("error writing certificate: %w", err)
	}

	econn.session.Close()
	econn.connection.Close(200, "enrollment complete")

	return nil
}

func (m *enrollmentManager) subscribe(onSet func(string, Enrollment), onRemove func(string)) func() {
	return m.waitingEnrollments.Subscribe(
		func(key string, conn *enrollmentConnection) {
			onSet(key, conn.enrollment)
		},
		onRemove,
	)
}

func (m *enrollmentManager) getAll() map[string]Enrollment {
	allConns := m.waitingEnrollments.GetAll()
	copy := make(map[string]Enrollment)
	for key, conn := range allConns {
		copy[key] = conn.enrollment
	}

	return copy
}

func EnrollWithUpstream() (*pki.PermanentCredentials, error) {
	addr := config.V().GetString("upstream.address")
	if addr == "" {
		return nil, fmt.Errorf("upstream address is missing")
	}

	tlsConf := getTlsTempClientConfig([]TlsConnectionProto{ProtoAgentEnroll})

	quicConf := &quic.Config{
		KeepAlivePeriod: 30 * time.Second,
	}

	quicConn, err := quic.DialAddr(context.Background(), addr, tlsConf, quicConf)
	if err != nil {
		qErr, ok := err.(*quic.TransportError)
		if ok && uint8(qErr.ErrorCode) == 120 {
			return nil, fmt.Errorf("server not ready for login: %w", err)
		}
		return nil, fmt.Errorf("error creating QUIC connection: %w", err)
	}

	initNonceStorage = NewNonceStorage()

	tempCredentials, err := pki.GenerateCredentials()
	if err != nil {
		return nil, fmt.Errorf("error generating temp credentials: %w", err)
	}

	conn := newRpcConnection(quicConn, nil, RpcRoleInit, initNonceStorage, nil, ProtoAgentEnroll, tempCredentials)

	session, err := conn.OpenSession(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error opening session: %w", err)
	}

	err = session.mutateState(RpcSessionCreated, RpcSessionOpen)
	if err != nil {
		return nil, fmt.Errorf("error mutating session state: %w", err)
	}

	err = exchangeKeys(session)
	if err != nil {
		return nil, fmt.Errorf("error exchanging keys: %w", err)
	}

	var cert pki.Certificate
	err = ReadMessage[*pki.Certificate](session, &cert)
	if err != nil {
		return nil, fmt.Errorf("error reading certificate: %w", err)
	}

	credentials, err := tempCredentials.UpgradeToHostCredentials(&cert)
	if err != nil {
		return nil, fmt.Errorf("error upgrading to host credentials: %w", err)
	}

	return credentials, nil
}
