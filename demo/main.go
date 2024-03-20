package main

type RedfishAction interface {
	Redfishing()
}

type X86Action struct{}

type Amd64Action struct{}

func (x *X86Action) Redfishing() {}

func (x *Amd64Action) Redfishing() {}

// 采用 TCP + TLS3.0 + protobuf version
// 传输层以上无法保证可靠性，需要在应用层保证
type Message struct {
	fixedHeader *FixedHeader
	// 可变头内容
	// 消息内容
}

func (m *Message) Encode() {

}

func (m *Message) Decode() {

}

type FixedHeader struct {
	version    byte   // 版本号，兼容性
	msgType    byte   // 消息类型，信令
	msgLen     uint32 // 消息长度
	varHeadLen uint32 // 可变头部长度
	crc32Sum   uint32 // CRC32校验
}

func main() {
	_ := new(Message)
}
