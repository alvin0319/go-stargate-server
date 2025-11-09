package util

import (
	"encoding/binary"
	"io"
	"math"
)

// WriteString writes a string with a 4-byte big-endian length prefix.
// This matches the Java ByteBuf string encoding used in StarGate.
func WriteString(w io.Writer, s string) error {
	data := []byte(s)
	if err := binary.Write(w, binary.BigEndian, int32(len(data))); err != nil {
		return err
	}
	_, err := w.Write(data)
	return err
}

// ReadString reads a string with a 4-byte big-endian length prefix.
// This matches the Java ByteBuf string encoding used in StarGate.
func ReadString(r io.Reader) (string, error) {
	var length int32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return "", err
	}
	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteInt64 writes an int64 value in big-endian format.
func WriteInt64(w io.Writer, v int64) error {
	return binary.Write(w, binary.BigEndian, v)
}

// ReadInt64 reads an int64 value in big-endian format.
func ReadInt64(r io.Reader) (int64, error) {
	var v int64
	err := binary.Read(r, binary.BigEndian, &v)
	return v, err
}

// WriteInt32 writes an int32 value in big-endian format.
func WriteInt32(w io.Writer, v int32) error {
	return binary.Write(w, binary.BigEndian, v)
}

// ReadInt32 reads an int32 value in big-endian format.
func ReadInt32(r io.Reader) (int32, error) {
	var v int32
	err := binary.Read(r, binary.BigEndian, &v)
	return v, err
}

// WriteInt16 writes an int16 value in big-endian format.
func WriteInt16(w io.Writer, v int16) error {
	return binary.Write(w, binary.BigEndian, v)
}

// ReadInt16 reads an int16 value in big-endian format.
func ReadInt16(r io.Reader) (int16, error) {
	var v int16
	err := binary.Read(r, binary.BigEndian, &v)
	return v, err
}

// WriteUint64 writes a uint64 value in big-endian format.
func WriteUint64(w io.Writer, v uint64) error {
	return binary.Write(w, binary.BigEndian, v)
}

// ReadUint64 reads a uint64 value in big-endian format.
func ReadUint64(r io.Reader) (uint64, error) {
	var v uint64
	err := binary.Read(r, binary.BigEndian, &v)
	return v, err
}

// WriteUint32 writes a uint32 value in big-endian format.
func WriteUint32(w io.Writer, v uint32) error {
	return binary.Write(w, binary.BigEndian, v)
}

// ReadUint32 reads a uint32 value in big-endian format.
func ReadUint32(r io.Reader) (uint32, error) {
	var v uint32
	err := binary.Read(r, binary.BigEndian, &v)
	return v, err
}

// WriteUint16 writes a uint16 value in big-endian format.
func WriteUint16(w io.Writer, v uint16) error {
	return binary.Write(w, binary.BigEndian, v)
}

// ReadUint16 reads a uint16 value in big-endian format.
func ReadUint16(r io.Reader) (uint16, error) {
	var v uint16
	err := binary.Read(r, binary.BigEndian, &v)
	return v, err
}

// WriteFloat32 writes a float32 value in big-endian format.
func WriteFloat32(w io.Writer, v float32) error {
	return binary.Write(w, binary.BigEndian, math.Float32bits(v))
}

// ReadFloat32 reads a float32 value in big-endian format.
func ReadFloat32(r io.Reader) (float32, error) {
	var bits uint32
	if err := binary.Read(r, binary.BigEndian, &bits); err != nil {
		return 0, err
	}
	return math.Float32frombits(bits), nil
}

// WriteFloat64 writes a float64 value in big-endian format.
func WriteFloat64(w io.Writer, v float64) error {
	return binary.Write(w, binary.BigEndian, math.Float64bits(v))
}

// ReadFloat64 reads a float64 value in big-endian format.
func ReadFloat64(r io.Reader) (float64, error) {
	var bits uint64
	if err := binary.Read(r, binary.BigEndian, &bits); err != nil {
		return 0, err
	}
	return math.Float64frombits(bits), nil
}

// WriteByte writes a single byte.
func WriteByte(w io.Writer, v byte) error {
	if bw, ok := w.(io.ByteWriter); ok {
		return bw.WriteByte(v)
	}
	_, err := w.Write([]byte{v})
	return err
}

// ReadByte reads a single byte.
func ReadByte(r io.Reader) (byte, error) {
	if br, ok := r.(io.ByteReader); ok {
		return br.ReadByte()
	}
	var b [1]byte
	_, err := io.ReadFull(r, b[:])
	return b[0], err
}

// WriteBool writes a boolean as a byte (0 or 1).
func WriteBool(w io.Writer, v bool) error {
	if v {
		return WriteByte(w, 1)
	}
	return WriteByte(w, 0)
}

// ReadBool reads a boolean from a byte (0 or 1).
func ReadBool(r io.Reader) (bool, error) {
	b, err := ReadByte(r)
	return b != 0, err
}

// WriteBytes writes a byte array with a 4-byte big-endian length prefix.
func WriteBytes(w io.Writer, data []byte) error {
	if err := WriteInt32(w, int32(len(data))); err != nil {
		return err
	}
	_, err := w.Write(data)
	return err
}

// ReadBytes reads a byte array with a 4-byte big-endian length prefix.
func ReadBytes(r io.Reader) ([]byte, error) {
	length, err := ReadInt32(r)
	if err != nil {
		return nil, err
	}
	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}
	return data, nil
}

// WriteStringArray writes a string array with a 4-byte length prefix followed by each string.
func WriteStringArray(w io.Writer, arr []string) error {
	if err := WriteInt32(w, int32(len(arr))); err != nil {
		return err
	}
	for _, s := range arr {
		if err := WriteString(w, s); err != nil {
			return err
		}
	}
	return nil
}

// ReadStringArray reads a string array with a 4-byte length prefix followed by each string.
func ReadStringArray(r io.Reader) ([]string, error) {
	length, err := ReadInt32(r)
	if err != nil {
		return nil, err
	}
	arr := make([]string, length)
	for i := int32(0); i < length; i++ {
		s, err := ReadString(r)
		if err != nil {
			return nil, err
		}
		arr[i] = s
	}
	return arr, nil
}

// WriteInt32Array writes an int32 array with a 4-byte length prefix.
func WriteInt32Array(w io.Writer, arr []int32) error {
	if err := WriteInt32(w, int32(len(arr))); err != nil {
		return err
	}
	for _, v := range arr {
		if err := WriteInt32(w, v); err != nil {
			return err
		}
	}
	return nil
}

// ReadInt32Array reads an int32 array with a 4-byte length prefix.
func ReadInt32Array(r io.Reader) ([]int32, error) {
	length, err := ReadInt32(r)
	if err != nil {
		return nil, err
	}
	arr := make([]int32, length)
	for i := int32(0); i < length; i++ {
		v, err := ReadInt32(r)
		if err != nil {
			return nil, err
		}
		arr[i] = v
	}
	return arr, nil
}

// WriteInt64Array writes an int64 array with a 4-byte length prefix.
func WriteInt64Array(w io.Writer, arr []int64) error {
	if err := WriteInt32(w, int32(len(arr))); err != nil {
		return err
	}
	for _, v := range arr {
		if err := WriteInt64(w, v); err != nil {
			return err
		}
	}
	return nil
}

// ReadInt64Array reads an int64 array with a 4-byte length prefix.
func ReadInt64Array(r io.Reader) ([]int64, error) {
	length, err := ReadInt32(r)
	if err != nil {
		return nil, err
	}
	arr := make([]int64, length)
	for i := int32(0); i < length; i++ {
		v, err := ReadInt64(r)
		if err != nil {
			return nil, err
		}
		arr[i] = v
	}
	return arr, nil
}
