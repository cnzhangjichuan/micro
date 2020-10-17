package packet

// Decoder 解码器
type Decoder interface {
	Decode(*Packet)
}

// Encoder 编码器
type Encoder interface {
	Encode(*Packet)
}

// Packable 编/解码器
type Packable interface {
	Decoder
	Encoder
}

// PackIdentifier 带用ID标识的编/解码器
type PackIdentifier interface {
	Packable

	// GetUID 获取标识符
	GetUID() string
}
