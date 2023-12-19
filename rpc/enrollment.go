package rpc

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/util"

	"github.com/quic-go/quic-go"
)

type EnrollmentManager interface {
	util.ObservableMap[string, *Enrollment]
	AcceptEnrollment(cert *pki.Certificate) error
}

type EndPointInitInfo struct {
	Root        *pki.Certificate
	Upstream    *pki.Certificate
	Credentials *pki.PermanentCredentials
}

type enrollmentManager struct {
	waitingEnrollments util.UpdateableMap[string, *enrollmentConnection]
	upstream           *pki.Certificate
	root               *pki.Certificate
	mutex              sync.Mutex
}

type enrollmentConnection struct {
	connection *RpcConnection
	session    *RpcSession
	enrollment *Enrollment
	mutex      sync.Mutex
}

type Enrollment struct {
	PublicKey   *pki.PublicKey
	Addr        string
	RequestTime time.Time
}

const maxEnrollmentTime = 5 * time.Minute

func newEnrollmentManager(upstream *pki.Certificate, root *pki.Certificate) *enrollmentManager {
	return &enrollmentManager{
		waitingEnrollments: util.NewObservableMap[string, *enrollmentConnection](),
		upstream:           upstream,
		root:               root,
		mutex:              sync.Mutex{},
	}
}

func (m *enrollmentManager) cleanup() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	timeout := make([]string, 0)
	m.waitingEnrollments.ForEach(func(key string, econn *enrollmentConnection) error {
		if econn.mutex.TryLock() {
			if time.Since(econn.enrollment.RequestTime) > maxEnrollmentTime {
				timeout = append(timeout, key)
			}
			econn.mutex.Unlock()
		}
		return nil
	})

	for _, key := range timeout {
		log.Printf("enrollment timed out: %s", key)
		m.waitingEnrollments.Delete(key)
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

	encodedKey := session.partnerKey.Base64Encode()

	m.mutex.Lock()
	defer m.mutex.Unlock()
	_, ok := m.waitingEnrollments.Get(encodedKey)
	if ok {
		return fmt.Errorf("enrollment already in progress")
	}

	m.waitingEnrollments.Set(encodedKey,
		&enrollmentConnection{
			connection: conn,
			session:    session,
			mutex:      sync.Mutex{},
			enrollment: &Enrollment{
				PublicKey:   session.partnerKey,
				Addr:        conn.connection.RemoteAddr().String(),
				RequestTime: time.Now(),
			},
		},
	)

	log.Printf("enrollment started for %s", encodedKey)

	return nil
}

func (m *enrollmentManager) AcceptEnrollment(cert *pki.Certificate) error {
	m.cleanup()
	encodedKey := cert.PublicKey().Base64Encode()

	log.Printf("enrollment accepted for %s", encodedKey)

	econn, ok := m.waitingEnrollments.Get(encodedKey)

	if !ok {
		return fmt.Errorf("enrollment not in progress")
	}

	log.Printf("trying to aquire lock")

	econn.mutex.Lock()
	defer econn.mutex.Unlock()

	log.Printf("enrollment lock aquired")

	m.waitingEnrollments.Delete(encodedKey)

	reponse := &enrollmentResponse{
		Cert:     cert,
		Root:     m.root,
		Upstream: m.upstream,
	}

	err := WriteMessage[*enrollmentResponse](econn.session, reponse)
	if err != nil {
		econn.connection.Close(500, "error writing response")
		return fmt.Errorf("error writing response: %w", err)
	}

	time.Sleep(5 * time.Second)

	econn.session.Close()

	return nil
}

func (m *enrollmentManager) Subscribe(onSet func(string, *Enrollment), onRemove func(string, *Enrollment)) func() {
	return m.waitingEnrollments.Subscribe(
		func(key string, conn *enrollmentConnection) {
			log.Printf("enrollment in progress: %s", key)
			onSet(key, conn.enrollment)
		},
		func(key string, conn *enrollmentConnection) {
			log.Printf("enrollment completed: %s", key)
			onRemove(key, conn.enrollment)
		},
	)
}

func (m *enrollmentManager) ForEach(fn func(string, *Enrollment) error) error {
	m.cleanup()
	return m.waitingEnrollments.ForEach(func(key string, conn *enrollmentConnection) error {
		return fn(key, conn.enrollment)
	})
}

type enrollmentResponse struct {
	Cert     *pki.Certificate
	Root     *pki.Certificate
	Upstream *pki.Certificate
}

func EnrollWithUpstream(addr string) (*EndPointInitInfo, error) {

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
	defer conn.Close(0, "")

	session, err := conn.OpenSession(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error opening session: %w", err)
	}
	defer session.Close()

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

	credentials, err := tempCredentials.ToPermanentCredentials(response.Cert)
	if err != nil {
		return nil, fmt.Errorf("error upgrading to host credentials: %w", err)
	}

	initInfo := &EndPointInitInfo{
		Root:        response.Root,
		Upstream:    response.Upstream,
		Credentials: credentials,
	}
	return initInfo, nil
}
