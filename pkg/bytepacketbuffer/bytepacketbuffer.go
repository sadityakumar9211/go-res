package bytepacketbuffer

import (
	"errors"
	"strings"
)

// pub struct BytePacketBuffer {
//     pub buf: [u8; 512],
//     pub pos: usize,
// }

type BytePacketBuffer struct {
	Buf []uint8
	Pos uint
}

func (buf *BytePacketBuffer) GetPos() uint {
	return buf.Pos
}

func (buf *BytePacketBuffer) Step(steps uint) {
	buf.Pos += steps
}

func (buf *BytePacketBuffer) Seek(pos uint) {
	buf.Pos = pos
}

func (buf *BytePacketBuffer) Read() (uint8, error) {
	if buf.Pos >= 512 {
		return 0, errors.New("end of buffer")
	}
	res := buf.Buf[buf.Pos]
	buf.Pos += 1
	return res, nil
}

func (buf *BytePacketBuffer) Get(pos uint) (uint8, error) {
	if pos >= 512 {
		return 0, errors.New("end of buffer")
	}
	return buf.Buf[pos], nil
}

func (buf *BytePacketBuffer) GetRange(start uint, len uint) ([]uint8, error) {
	if start+len >= 512 {
		return nil, errors.New("end of buffer")
	}
	return buf.Buf[start : start+len], nil
}

func (buf *BytePacketBuffer) ReadU16() (uint16, error) {
	byte1, err := buf.Read()
	if err != nil {
		return 0, err
	}

	byte2, err := buf.Read()
	if err != nil {
		return 0, err
	}

	res := (uint16(byte1) << 8) | uint16(byte2)
	return res, nil
}

func (buf *BytePacketBuffer) ReadU32() (uint32, error) {
	byte1, err := buf.Read()
	if err != nil {
		return 0, err
	}

	byte2, err := buf.Read()
	if err != nil {
		return 0, err
	}

	byte3, err := buf.Read()
	if err != nil {
		return 0, err
	}

	byte4, err := buf.Read()
	if err != nil {
		return 0, err
	}

	res := (uint32(byte1) << 24) | (uint32(byte2) << 16) | (uint32(byte3) << 8) | uint32(byte4)
	return res, nil
}

func (buf *BytePacketBuffer) ReadQName(outstr *string) error {
	var pos = buf.Pos
	var jumped = false

	var delim = ""
	const maxJumps = 5
	var jumpsPerformed = 0

	for {
		// DNS packets are untrusted data, so we need to be cautious. Someone
		// can craft a packet with a cycle in the jump instructions. This guards
		// against such packets.
		if jumpsPerformed > maxJumps {
			return errors.New("limit of jumps exceeded")
		}

		len, err := buf.Get(pos)
		if err != nil {
			return err
		}

		// A two-byte sequence, where the two highest bits of the first byte are
		// set, represents an offset relative to the start of the buffer. We
		// handle this by jumping to the offset, setting a flag to indicate
		// that we shouldn't update the shared buffer position once done.
		if (len & 0xC0) == 0xC0 {
			// When a jump is performed, we only modify the shared buffer
			// position once, and avoid making the change later on.
			if !jumped {
				buf.Seek(pos + 2)
			}

			b2, err := buf.Get(pos + 1)
			if err != nil {
				return err
			}
			offset := (((uint16(len) ^ 0xC0) << 8) | uint16(b2))
			pos = uint(offset)
			jumped = true
			jumpsPerformed++
			continue
		}

		pos++

		// Names are terminated by an empty label of length 0
		if len == 0 {
			break
		}

		*outstr += delim

		strBuffer, err := buf.GetRange(pos, uint(len))
		if err != nil {
			return err
		}
		*outstr += strings.ToLower(string(strBuffer))

		delim = "."

		pos += uint(len)
	}

	if !jumped {
		buf.Seek(pos)
	}

	return nil
}

// general write function
func (buf *BytePacketBuffer) Write(val uint8) error {
	if buf.Pos >= 512 {
		return errors.New("end of buffer")
	}

	buf.Buf[buf.Pos] = val
	buf.Pos += 1
	return nil
}

func (buf *BytePacketBuffer) WriteU8(val uint8) error {
	if err := buf.Write(val); err != nil {
		return err
	}
	return nil
}

func (buf *BytePacketBuffer) WriteU16(val uint16) error {
	if err := buf.Write(uint8(val >> 8)); err != nil {
		return err
	}

	if err := buf.Write(uint8(val & 0xFF)); err != nil {
		return err
	}
	return nil
}

func (buf *BytePacketBuffer) WriteU32(val uint32) error {
	if err := buf.Write(uint8((val >> 24) & 0xFF)); err != nil {
		return err
	}
	if err := buf.Write(uint8((val >> 16) & 0xFF)); err != nil {
		return err
	}
	if err := buf.Write(uint8((val >> 8) & 0xFF)); err != nil {
		return err
	}
	if err := buf.Write(uint8((val >> 0) & 0xFF)); err != nil {
		return err
	}

	return nil
}

func (buf *BytePacketBuffer) WriteQName(qname string) error {
	labels := strings.Split(qname, ".")
	for _, label := range labels {
		len := len(label)
		if len > 0x34 {
			return errors.New("single label exceeds 63 characters of length")
		}

		if err := buf.WriteU8(uint8(len)); err != nil {
			return err
		}

		for _, b := range []byte(label) {
			if err := buf.WriteU8(b); err != nil {
				return err
			}
		}
	}

	if err := buf.WriteU8(0); err != nil {
		return err
	}

	return nil // Return nil to indicate success
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
