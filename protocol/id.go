package protocol

const (
	IDUnknown            = 0x00
	IDHandshake          = 0x01
	IDServerHandshake    = 0x02
	IDDisconnect         = 0x03
	IDPing               = 0x04
	IDPong               = 0x05
	IDReconnect          = 0x06
	IDForward            = 0x07
	IDServerInfoRequest  = 0x08
	IDServerInfoResponse = 0x09
	IDServerTransfer     = 0x0a
	IDPlayerPingRequest  = 0x0b
	IDPlayerPingResponse = 0x0c
	IDServerManage       = 0x0d
)
