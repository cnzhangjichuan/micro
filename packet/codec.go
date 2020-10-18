package packet

// Decoder 解码器
type Decoder interface {
	Decode(*Packet)
}

// Encoder 编码器
type Encoder interface {
	Encode(*Packet)
}

// Serializable 可序列化包
type Serializable interface {
	Decoder
	Encoder
}

// Identifier 带用ID标识的可序列化包
type Identifier interface {
	Serializable

	// GetUID 获取标识符
	GetUID() string
}
