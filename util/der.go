package util

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
)

func ReadSingleDer(reader io.Reader) ([]byte, error) {
	derStart := make([]byte, 2)
	_, err := io.ReadFull(reader, derStart)
	if err != nil {
		return nil, fmt.Errorf("failed to first two asn1 bytes: %w", err)
	}

	isMultiByteLength := derStart[1]&0b1000_0000 != 0
	firstByteValue := derStart[1] & 0b0111_1111
	var lengthBytes []byte
	if isMultiByteLength {
		lengthBytes = make([]byte, uint(firstByteValue))
		_, err := io.ReadFull(reader, lengthBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to read extended asn1 length: %w", err)
		}
	} else {
		lengthBytes = []byte{firstByteValue}
	}

	length := &big.Int{}
	length.SetBytes(lengthBytes)

	derBody := make([]byte, length.Int64())
	_, err = io.ReadFull(reader, derBody)
	if err != nil {
		return nil, fmt.Errorf("failed to read asn1 body: %w", err)
	}

	toJoin := [][]byte{
		derStart,
	}

	if isMultiByteLength {
		toJoin = append(toJoin, lengthBytes)
	}
	toJoin = append(toJoin, derBody)

	return bytes.Join(toJoin, []byte{}), nil
}

func TryReadSingleDer(reader io.Reader, buffer []byte) (int, error) {
	_, err := io.ReadFull(reader, buffer[:2])
	if err != nil {
		return 0, fmt.Errorf("failed to first two asn1 bytes: %w", err)
	}

	var offset int64 = 2

	isMultiByteLength := buffer[1]&0b1000_0000 != 0
	firstByteValue := buffer[1] & 0b0111_1111
	var lengthBytes []byte
	if isMultiByteLength {
		fbv := int64(firstByteValue)
		newOffset := offset + fbv
		_, err := io.ReadFull(reader, buffer[offset:newOffset])
		if err != nil {
			return 0, fmt.Errorf("failed to read extended asn1 length: %w", err)
		}
		offset = newOffset
	} else {
		lengthBytes = []byte{firstByteValue}
	}

	length := &big.Int{}
	length.SetBytes(lengthBytes)
	lengthInt := length.Int64()
	if lengthInt > int64(len(buffer)) {
		return 0, fmt.Errorf("buffer too small")
	}

	n, err := io.ReadFull(reader, buffer[offset:offset+lengthInt])
	if err != nil {
		return 0, fmt.Errorf("failed to read asn1 body: %w", err)
	}

	return n + int(offset), nil
}
