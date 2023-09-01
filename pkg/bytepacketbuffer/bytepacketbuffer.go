package bytepacketbuffer

import (
	"errors"
)

// BytePacketBuffer is a buffer for working with binary data.
type BytePacketBuffer struct {
	Buf [512]byte
	Pos int
}

func NewBytePacketBuffer() BytePacketBuffer {
	return BytePacketBuffer{}
}

func (b *BytePacketBuffer) GetPos() int {
	return b.Pos
}

func (b *BytePacketBuffer) Step(steps int) error {
	b.Pos += steps
	return nil
}

func (b *BytePacketBuffer) Seek(pos int) error {
	b.Pos = pos
	return nil
}

func (b *BytePacketBuffer) Read() (byte, error) {
	if b.Pos >= 512 {
		return 0, errors.New("End of buffer")
	}
	res := b.Buf[b.Pos]
	b.Pos++
	return res, nil
}

func (b *BytePacketBuffer) Get(pos int) (byte, error) {
	if pos >= 512 {
		return 0, errors.New("End of buffer")
	}
	return b.Buf[pos], nil
}

func (b *BytePacketBuffer) GetRange(start int, length int) ([]byte, error) {
	if start+length >= 512 {
		return nil, errors.New("End of buffer")
	}
	return b.Buf[start : start+length], nil
}

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

func (b *BytePacketBuffer) ReadQName(outstr *string) error {
	pos := b.GetPos()
	jumped := false
	delim := ""
	maxJumps := 5
	jumpsPerformed := 0

	for {
		if jumpsPerformed > maxJumps {
			return errors.New("Limit of 5 jumps exceeded")
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

func (b *BytePacketBuffer) Write(val byte) error {
	if b.Pos >= 512 {
		return errors.New("End of buffer")
	}
	b.Buf[b.Pos] = val
	b.Pos++
	return nil
}

func (b *BytePacketBuffer) WriteU8(val byte) error {
	return b.Write(val)
}

func (b *BytePacketBuffer) WriteU16(val uint16) error {
	b.Write(byte(val >> 8))
	b.Write(byte(val & 0xFF))
	return nil
}

func (b *BytePacketBuffer) WriteU32(val uint32) error {
	b.Write(byte((val >> 24) & 0xFF))
	b.Write(byte((val >> 16) & 0xFF))
	b.Write(byte((val >> 8) & 0xFF))
	b.Write(byte(val & 0xFF))
	return nil
}

func (b *BytePacketBuffer) WriteQName(qname string) error {
	for _, label := range SplitDNSName(qname) {
		lenVal := byte(len(label))
		if lenVal > 0x34 {
			return errors.New("Single label exceeds 63 characters of length")
		}
		b.WriteU8(lenVal)
		for _, ch := range label {
			b.Write(byte(ch))
		}
	}
	b.WriteU8(0)
	return nil
}

func (b *BytePacketBuffer) Set(pos int, val byte) error {
	b.Buf[pos] = val
	return nil
}

func (b *BytePacketBuffer) SetU16(pos int, val uint16) error {
	b.Set(pos, byte(val>>8))
	b.Set(pos+1, byte(val&0xFF))
	return nil
}

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
