package p2p

import (
	"encoding/binary"
	"go-chain/block"
	"go-chain/util"
	"net"
)

const (
	MSG_BLOCK      = 1
	MSG_CHAIN_SYNC = 2
)

type Peer struct {
	C                net.Conn
	BlockChannel     chan block.Block
	ChainSyncChannel chan net.Conn
}

func SendBlock(conn net.Conn, bl block.Block) error {
	messageBuf := []byte{MSG_BLOCK}
	blockBuf := bl.Serialize()
	messageBuf = append(messageBuf, blockBuf...)
	if _, err := conn.Write(messageBuf); err != nil {
		return err
	}
	return nil
}

func SendChainSync(conn net.Conn) error {
	messageBuf := []byte{MSG_CHAIN_SYNC}
	if _, err := conn.Write(messageBuf); err != nil {
		return err
	}
	return nil
}

func handleBlock(c net.Conn, blockChannel chan block.Block) error {
	blockHeaderBuf := make([]byte, block.BLOCK_HEADER_LEN)
	if _, err := c.Read(blockHeaderBuf); err != nil {
		return err
	}
	dataLenBuf := make([]byte, 2)
	if _, err := c.Read(dataLenBuf); err != nil {
		return err
	}
	dataLen := binary.BigEndian.Uint16(dataLenBuf)
	dataBuf := make([]byte, dataLen)
	if _, err := c.Read(dataBuf); err != nil {
		return err
	}
	block := block.Deserialize(blockHeaderBuf, dataBuf)
	blockChannel <- block
	return nil
}

func HandlePeer(peer Peer) {
	// TODO: defer removal from peer collection
	defer peer.C.Close()
	for {
		messageTypeBuf := make([]byte, 1)
		if _, err := peer.C.Read(messageTypeBuf); err != nil {
			return
		}
		switch messageTypeBuf[0] {
		case MSG_BLOCK:
			if err := handleBlock(peer.C, peer.BlockChannel); err != nil {
				util.GoChainLogger.Println("Failed to read from peer: " + err.Error())
				return
			}
		case MSG_CHAIN_SYNC:
			peer.ChainSyncChannel <- peer.C
		default:
		}
	}
}