package pki

import (
	"encoding/json"
	"fmt"
	"log"
)

type ArtifactPayload interface {
	MayPublish(*Certificate) bool
}

type SignedArtifact[T ArtifactPayload] struct {
	blob     SignedBlob
	artifact T
}

func NewSignedArtifact[T ArtifactPayload](credentials *PermanentCredentials, artifact T) (*SignedArtifact[T], error) {

	cert, err := credentials.GetCertificate()
	if err != nil {
		return nil, fmt.Errorf("failed to get current cert: %w", err)
	}

	if !artifact.MayPublish(cert) {
		return nil, fmt.Errorf("not authorized to publish this artifact")
	}

	marshalled, err := json.Marshal(artifact)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	blob, err := NewSignedBlob(credentials, marshalled)
	if err != nil {
		return nil, fmt.Errorf("failed to sign payload: %w", err)
	}

	return &SignedArtifact[T]{
		blob:     *blob,
		artifact: artifact,
	}, nil
}

func LoadSignedArtifact[T ArtifactPayload](raw []byte, verifier Verifier, target T) (*SignedArtifact[T], error) {
	blob, err := LoadSignedBlob(raw, verifier)
	if err != nil {
		return nil, fmt.Errorf("failed to load blob: %w", err)
	}

	err = json.Unmarshal(blob.Payload(), target)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if !target.MayPublish(blob.Creator()) {
		errDangerous := fmt.Errorf("the creator of this artifact was not allowed to publish it")
		log.Print(errDangerous)
		return nil, errDangerous
	}

	return &SignedArtifact[T]{
		blob:     *blob,
		artifact: target,
	}, nil
}

func (s *SignedArtifact[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.blob.Raw())
}

func (s *SignedArtifact[T]) Artifact() T {
	return s.artifact
}
