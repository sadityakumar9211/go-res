package bytepacketbuffer

import (
	"errors"
)

// BytePacketBuffer is a buffer for working with binary data.
type BytePacketBuffer struct {
	Buf [512]byte
	Pos int // buffer pointer to track current position.
}

// NewBytePacketBuffer creates and returns a new BytePacketBuffer with default values.
func NewBytePacketBuffer() BytePacketBuffer {
	return BytePacketBuffer{}
}

// GetPos returns the current buffer pointer.
func (b *BytePacketBuffer) GetPos() int {
	return b.Pos
}

// Step moves the buffer pointer by steps amount.
func (b *BytePacketBuffer) Step(steps int) error {
	b.Pos += steps
	return nil
}

// Seek sets the buffer pointer to pos.
func (b *BytePacketBuffer) Seek(pos int) error {
	b.Pos = pos
	return nil
}

// Read reads a single byte from buffer and moves buffer pointer by same amount.
func (b *BytePacketBuffer) Read() (byte, error) {
	if b.Pos >= 512 {
		return 0, errors.New("end of buffer")
	}
	res := b.Buf[b.Pos]
	b.Pos++
	return res, nil
}

// Get returns a buffer byte at pos without changing buffer pointer.
func (b *BytePacketBuffer) Get(pos int) (byte, error) {
	if pos >= 512 {
		return 0, errors.New("end of buffer")
	}
	return b.Buf[pos], nil
}

// GetRange returns buffer bits from start with specified length without moving buffer pointer. 
func (b *BytePacketBuffer) GetRange(start int, length int) ([]byte, error) {
	if start+length >= 512 {
		return nil, errors.New("end of buffer")
	}
	return b.Buf[start : start+length], nil
}

// ReadU16 reads 2 buffer bytes and moves buffer pointer.
func (b *BytePacketBuffer) ReadU16() (uint16, error) {
	val1, err := b.Read()
	if err != nil {
		return 0, err
	}
	val2, err := b.Read()
	if err != nil {
		return 0, err
	}
	return (uint16(val1) << 8) | uint16(val2), nil
}

// ReadU32 reads 4 buffer bytes and moves buffer pointer. 
func (b *BytePacketBuffer) ReadU32() (uint32, error) {
	val1, err := b.Read()
	if err != nil {
		return 0, err
	}
	val2, err := b.Read()
	if err != nil {
		return 0, err
	}
	val3, err := b.Read()
	if err != nil {
		return 0, err
	}
	val4, err := b.Read()
	if err != nil {
		return 0, err
	}
	return (uint32(val1) << 24) | (uint32(val2) << 16) | (uint32(val3) << 8) | uint32(val4), nil
}

// ReadQName reads DNS question name and moves buffer pointer.
func (b *BytePacketBuffer) ReadQName(outstr *string) error {
	pos := b.GetPos()
	jumped := false
	delim := ""
	maxJumps := 5
	jumpsPerformed := 0

	for {
		if jumpsPerformed > maxJumps {
			return errors.New("limit of 5 jumps exceeded")
		}

		lenVal, err := b.Get(pos)
		if err != nil {
			return err
		}

		if (lenVal & 0xC0) == 0xC0 {
			if !jumped {
				b.Seek(pos + 2)
			}
			b2, err := b.Get(pos + 1)
			if err != nil {
				return err
			}
			offset := ((uint16(lenVal) ^ 0xC0) << 8) | uint16(b2)
			pos = int(offset)
			jumped = true
			jumpsPerformed++
			continue
		}

		pos++
		if lenVal == 0 {
			break
		}

		*outstr += delim
		strBuffer, err := b.GetRange(pos, int(lenVal))
		if err != nil {
			return err
		}
		*outstr += string(strBuffer)
		delim = "."
		pos += int(lenVal)
	}

	if !jumped {
		b.Seek(pos)
	}

	return nil
}

// Write writes a byte to buffer and moves buffer pointer.
func (b *BytePacketBuffer) Write(val byte) error {
	if b.Pos >= 512 {
		return errors.New("end of buffer")
	}
	b.Buf[b.Pos] = val
	b.Pos++
	return nil
}

// WriteU8 writes 1 byte to buffer and moves buffer pointer.
func (b *BytePacketBuffer) WriteU8(val byte) error {
	return b.Write(val)
}

// WriteU16 writes 2 bytes to buffer and moves buffer pointer.
func (b *BytePacketBuffer) WriteU16(val uint16) error {
	b.Write(byte(val >> 8))
	b.Write(byte(val & 0xFF))
	return nil
}

// WriteU32 writes 4 bytes to buffer and moves buffer pointer.
func (b *BytePacketBuffer) WriteU32(val uint32) error {
	b.Write(byte((val >> 24) & 0xFF))
	b.Write(byte((val >> 16) & 0xFF))
	b.Write(byte((val >> 8) & 0xFF))
	b.Write(byte(val & 0xFF))
	return nil
}

// WriteQName writes Question name to the buffer and moves buffer pointer.
func (b *BytePacketBuffer) WriteQName(qname string) error {
	for _, label := range SplitDNSName(qname) {
		lenVal := byte(len(label))
		if lenVal > 0x34 {
			return errors.New("single label exceeds 63 characters of length")
		}
		b.WriteU8(lenVal)
		for _, ch := range label {
			b.Write(byte(ch))
		}
	}
	b.WriteU8(0)
	return nil
}

// Set overwrite a byte from the given position.
func (b *BytePacketBuffer) Set(pos int, val byte) error {
	b.Buf[pos] = val
	return nil
}

// SetU16 overwrites 2 bytes from the given position.
func (b *BytePacketBuffer) SetU16(pos int, val uint16) error {
	b.Set(pos, byte(val>>8))
	b.Set(pos+1, byte(val&0xFF))
	return nil
}

// SplitDNSName splits question name string to multiple labels and returns an slice of labels.
func SplitDNSName(qname string) []string {
	labels := make([]string, 0)
	labelStart := 0

	for i, ch := range qname {
		if ch == '.' {
			labels = append(labels, qname[labelStart:i])
			labelStart = i + 1
		}
	}

	labels = append(labels, qname[labelStart:])
	return labels
}
