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
	upstream           *pki.Certificate
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

func newEnrollmentManager(upstream *pki.Certificate) *enrollmentManager {
	return &enrollmentManager{
		waitingEnrollments: util.NewObservableMap[string, *enrollmentConnection](),
		upstream:           upstream,
		mutex:              sync.Mutex{},
	}
}

func (m *enrollmentManager) cleanup() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for key, econn := range m.waitingEnrollments.GetAll() {
		if econn.mutex.TryLock() && time.Since(econn.enrollment.RequestTime) > maxEnrollmentTime {
			log.Printf("enrollment timed out for %s", key)
			econn.connection.Close(408, "enrollment timed out")
			m.waitingEnrollments.Delete(key)
		}
		econn.mutex.Unlock()
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

	encodedKey := session.partner.Base64Encode()

	m.mutex.Lock()
	if m.waitingEnrollments.Has(encodedKey) {
		return fmt.Errorf("enrollment already in progress")
	}

	m.waitingEnrollments.Set(encodedKey,
		&enrollmentConnection{
			connection: conn,
			session:    session,
			mutex:      sync.Mutex{},
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
	encodedKey := cert.GetPublicKey().Base64Encode()

	log.Printf("enrollment accepted for %s", encodedKey)

	econn, ok := m.waitingEnrollments.Get(encodedKey)

	if !ok {
		return fmt.Errorf("enrollment not in progress")
	}

	econn.mutex.Lock()
	defer econn.mutex.Unlock()

	m.waitingEnrollments.Delete(encodedKey)

	root, err := pki.Root.Get()
	if err != nil {
		econn.connection.Close(500, "error getting root certificate")
		return fmt.Errorf("error getting root certificate: %w", err)
	}

	reponse := &enrollmentResponse{
		Cert:     cert,
		Root:     root,
		Upstream: m.upstream,
	}

	err = WriteMessage[*enrollmentResponse](econn.session, reponse)
	if err != nil {
		econn.connection.Close(500, "error writing response")
		return fmt.Errorf("error writing response: %w", err)
	}

	econn.session.Close()

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

type enrollmentResponse struct {
	Cert     *pki.Certificate
	Root     *pki.Certificate
	Upstream *pki.Certificate
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

	initNonceStorage = util.NewNonceStorage()

	tempCredentials, err := pki.GenerateCredentials()
	if err != nil {
		return nil, fmt.Errorf("error generating temp credentials: %w", err)
	}

	conn := newRpcConnection(quicConn, nil, RpcRoleInit, initNonceStorage, nil, ProtoAgentEnroll, tempCredentials, pki.NewNilVerifier())

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

	response := &enrollmentResponse{}

	err = ReadMessage[*enrollmentResponse](session, response)
	if err != nil {
		return nil, fmt.Errorf("error reading certificate: %w", err)
	}

	err = pki.Root.Set(response.Root)
	if err != nil {
		return nil, fmt.Errorf("error setting root certificate: %w", err)
	}

	err = pki.Upstream.Set(response.Upstream)
	if err != nil {
		return nil, fmt.Errorf("error setting upstream certificate: %w", err)
	}

	credentials, err := tempCredentials.UpgradeToHostCredentials(response.Cert)
	if err != nil {
		return nil, fmt.Errorf("error upgrading to host credentials: %w", err)
	}

	err = session.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing session: %w", err)
	}

	err = conn.Close(200, "enrollment complete")
	if err != nil {
		return nil, fmt.Errorf("error closing connection: %w", err)
	}

	return credentials, nil
}
