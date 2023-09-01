package bytepacketbuffer

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)


// BytePacketBuffer is a buffer for working with binary data.
type BytePacketBuffer struct {
	Buf []byte
	Pos uint
}

func newBytePacketBuffer() *BytePacketBuffer {
	return &BytePacketBuffer{
		Buf: make([]byte, 512),
		Pos: 0,
	}
}

func (b *BytePacketBuffer) Read() (byte, error) {
	if b.Pos >= 512 {
		return 0, errors.New("End of buffer")
	}
	res := b.Buf[b.Pos]
	b.Pos++
	return res, nil
}

func (b *BytePacketBuffer) ReadU16() (uint16, error) {
	highByte, err := b.Read()
	if err != nil {
		return 0, err
	}
	lowByte, err := b.Read()
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16([]byte{highByte, lowByte}), nil
}

func (b *BytePacketBuffer) ReadU32() (uint32, error) {
	byte1, err := b.Read()
	if err != nil {
		return 0, err
	}
	byte2, err := b.Read()
	if err != nil {
		return 0, err
	}
	byte3, err := b.Read()
	if err != nil {
		return 0, err
	}
	byte4, err := b.Read()
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32([]byte{byte1, byte2, byte3, byte4}), nil
}

func (b *BytePacketBuffer) ReadQName() (string, error) {
	var parts []string
	jumped := false
	maxJumps := 5
	jumpsPerformed := 0

	for {
		if jumpsPerformed > maxJumps {
			return "", fmt.Errorf("Limit of %d jumps exceeded", maxJumps)
		}

		lenByte, err := b.Read()
		if err != nil {
			return "", err
		}
		if (lenByte & 0xC0) == 0xC0 {
			if !jumped {
				_, _ = b.Read()
			}
			offset, err := b.ReadU16()
			if err != nil {
				return "", err
			}
			offset &= 0x3FFF
			b.Pos = uint(offset)
			jumped = true
			jumpsPerformed++
			continue
		}
		if lenByte == 0 {
			break
		}
		partBytes := make([]byte, lenByte)
		for i := byte(0); i < lenByte; i++ {
			partBytes[i], err = b.Read()
			if err != nil {
				return "", err
			}
		}
		part := string(partBytes)
		parts = append(parts, part)
	}
	return strings.Join(parts, "."), nil
}




func (b *BytePacketBuffer) Write(val byte) error {
    if b.Pos >= 512 {
        return errors.New("End of buffer")
    }
    b.Buf[b.Pos] = val
    b.Pos++
    return nil
}

func (b *BytePacketBuffer) WriteU16(val uint16) error {
    buf := make([]byte, 2)
    binary.BigEndian.PutUint16(buf, val)
    if err := b.Write(buf[0]); err != nil {
        return err
    }
    if err := b.Write(buf[1]); err != nil {
        return err
    }
    return nil
}

func (b *BytePacketBuffer) WriteU32(val uint32) error {
    buf := make([]byte, 4)
    binary.BigEndian.PutUint32(buf, val)
    for _, byteVal := range buf {
        if err := b.Write(byteVal); err != nil {
            return err
        }
    }
    return nil
}

func (b *BytePacketBuffer) WriteQName(qname string) error {
    labels := strings.Split(qname, ".")
    for _, label := range labels {
        lenByte := byte(len(label))
        if err := b.Write(lenByte); err != nil {
            return err
        }
        for _, char := range label {
            if err := b.Write(byte(char)); err != nil {
                return err
            }
        }
    }
    if err := b.Write(0); err != nil {
        return err
    }
    return nil
}


func (buf *BytePacketBuffer) GetPos() uint {
	return uint(buf.Pos)
}

func (buf *BytePacketBuffer) Step(steps uint) {
	buf.Pos += steps
}

func (buf *BytePacketBuffer) Seek(pos uint) {
	buf.Pos = pos
}

func (b *BytePacketBuffer) Get(pos int) (byte, error) {
    if pos >= 512 {
        return 0, errors.New("End of buffer")
    }
    return b.Buf[pos], nil
}

func (b *BytePacketBuffer) GetRange(start, length int) ([]byte, error) {
    if start+length >= 512 {
        return nil, errors.New("End of buffer")
    }
    return b.Buf[start : start+length], nil
}

func (buf *BytePacketBuffer) WriteU8(val uint8) error {
	if err := buf.Write(val); err != nil {
		return err
	}
	return nil
}

func (buf *BytePacketBuffer) Set(pos int, val uint8) error {
	if pos < 0 || pos >= len(buf.Buf) {
		return errors.New("position out of bounds")
	}
	buf.Buf[pos] = val
	return nil // Return nil to indicate success
}

func (buf *BytePacketBuffer) SetU16(pos int, val uint16) error {
	if pos < 0 || pos+1 >= len(buf.Buf) {
		return errors.New("position out of bounds")
	}

	buf.Buf[pos] = uint8(val >> 8)
	buf.Buf[pos+1] = uint8(val)

	return nil // Return nil to indicate success
}
