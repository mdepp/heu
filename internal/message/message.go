package message

import (
	"encoding/binary"
	"fmt"

	"github.com/lucasb-eyer/go-colorful"
)

type ColorSpace = uint8

const (
	COLOR_SPACE_RGB ColorSpace = 0
	COLOR_SPACE_XY  ColorSpace = 1
)

const (
	MAX_CHANNELS                 = 20
	PREAMBLE_SIZE                = 52
	ENTERTAINMENT_CONFIG_ID_SIZE = 36
	CHANNEL_MESSAGE_SIZE         = 7
	MAX_MESSAGE_SIZE             = PREAMBLE_SIZE + MAX_CHANNELS*CHANNEL_MESSAGE_SIZE
)

var MESSAGE_BYTE_ORDER = binary.BigEndian

type MessageBuilder struct {
	buffer              []byte
	colorSpace          ColorSpace
	channelMessageCount int
}

func NewBuilder(colorSpace ColorSpace) *MessageBuilder {
	return &MessageBuilder{buffer: make([]byte, MAX_MESSAGE_SIZE), colorSpace: colorSpace}
}

func (b *MessageBuilder) WritePreamble(entertainmentConfigId []byte) *MessageBuilder {
	if len(entertainmentConfigId) != ENTERTAINMENT_CONFIG_ID_SIZE {
		panic(fmt.Sprintf("entertainment config id is incorrect size: %d != %d", len(entertainmentConfigId), ENTERTAINMENT_CONFIG_ID_SIZE))
	}
	position := 0
	position += copy(b.buffer[position:], []byte{
		// Protocol name
		'H', 'u', 'e', 'S', 't', 'r', 'e', 'a', 'm',
		// Version
		0x02, 0x00,
		// Sequence ID
		0x00,
		// Reserved
		0x00, 0x00,
		// Color space
		b.colorSpace,
		// Reserved
		0x00,
	})
	// Entertainment configuration id
	position += copy(b.buffer[position:], entertainmentConfigId)

	return b
}

func (b *MessageBuilder) ResetBody() *MessageBuilder {
	b.channelMessageCount = 0
	return b
}

func (b *MessageBuilder) WriteChannelColor(channelId uint8, col colorful.Color) *MessageBuilder {
	if b.channelMessageCount += 1; b.channelMessageCount > MAX_CHANNELS {
		panic(fmt.Sprintf("attempted to add too many channel messages, a maximum of %d are allowed", MAX_CHANNELS))
	}

	start := PREAMBLE_SIZE

	b.buffer[start] = channelId

	var rx, gy, bz uint16
	if b.colorSpace == COLOR_SPACE_RGB {
		rx = uint16(col.R * 0xFFFF)
		gy = uint16(col.G * 0xFFFF)
		bz = uint16(col.B * 0xFFFF)
	} else {
		x, y, z := col.Xyz()
		rx = uint16(x * 0xFFFF)
		gy = uint16(y * 0xFFFF)
		bz = uint16(z * 0xFFFF)
	}
	MESSAGE_BYTE_ORDER.PutUint16(b.buffer[start+1:], rx)
	MESSAGE_BYTE_ORDER.PutUint16(b.buffer[start+3:], gy)
	MESSAGE_BYTE_ORDER.PutUint16(b.buffer[start+5:], bz)

	return b
}

func (b *MessageBuilder) Build() []byte {
	size := PREAMBLE_SIZE + b.channelMessageCount*CHANNEL_MESSAGE_SIZE
	return b.buffer[:size]
}
