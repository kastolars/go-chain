package p2p

import (
	"encoding/binary"
	"go-chain/block"
	"log"
	"net"
	"os"
)

var logger *log.Logger = log.New(os.Stdout, "LOG: ", log.Lmicroseconds|log.Lshortfile)

const (
	NEW_BLOCK = 0
	CHAIN_TIP = 1
)

func HandlePeer(c net.Conn, newBlockChannel chan block.Block) {
	// TODO: defer removal from peer collection
	defer c.Close()
	for {
		// messageTypeBuf := make([]byte, 1)
		// if _, err := c.Read(messageTypeBuf); err != nil {
		// 	logger.Println("Failed to read from peer: " + err.Error())
		// 	return
		// }
		// switch messageTypeBuf[0] {
		// case NEW_BLOCK:
		blockHeaderBuf := make([]byte, block.BLOCK_HEADER_LEN)
		if _, err := c.Read(blockHeaderBuf); err != nil {
			logger.Println("Failed to read from peer: " + err.Error())
			return
		}
		dataLenBuf := make([]byte, 2)
		if _, err := c.Read(dataLenBuf); err != nil {
			logger.Println("Failed to read from peer: " + err.Error())
			return
		}
		dataLen := binary.BigEndian.Uint16(dataLenBuf)
		dataBuf := make([]byte, dataLen)
		if _, err := c.Read(dataBuf); err != nil {
			logger.Println("Failed to read from peer: " + err.Error())
			return
		}
		block := block.Deserialize(blockHeaderBuf, dataBuf)
		newBlockChannel <- block
		// default:
		// }

	}
}
