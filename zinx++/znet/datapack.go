package znet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"zinxplusplus/config"
	"zinxplusplus/ziface"

	"github.com/cloudwego/netpoll"
)

var (
	ErrReadHeaderTimeout = errors.New("read header timeout")
	ErrReadHeaderEOF     = errors.New("read header EOF")
	ErrDataTooLarge      = errors.New("received msg data too large")
)

type DataPack struct{}

func NewDataPack() ziface.IDataPack {
	return &DataPack{}
}

func (dp *DataPack) GetHeadLen() uint32 {

	return 8
}

func (dp *DataPack) Pack(msg ziface.IMessage) ([]byte, error) {

	dataBuff := bytes.NewBuffer([]byte{})

	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetDataLen()); err != nil {
		return nil, fmt.Errorf("pack datalen error: %w", err)
	}

	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgID()); err != nil {
		return nil, fmt.Errorf("pack msgid error: %w", err)
	}

	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetData()); err != nil {
		return nil, fmt.Errorf("pack data error: %w", err)
	}

	return dataBuff.Bytes(), nil
}

func (dp *DataPack) Unpack(reader netpoll.Reader) (ziface.IMessage, error) {

	headLen := int(dp.GetHeadLen())
	headData, err := reader.Next(headLen)
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, netpoll.ErrEOF) {
			return nil, fmt.Errorf("unpack read header error(EOF): %w", ErrReadHeaderEOF)
		}

		return nil, fmt.Errorf("unpack read header error: %w", err)
	}

	if len(headData) < headLen {

		reader.Release()
		return nil, fmt.Errorf("unpack read header error: read %d bytes, expect %d", len(headData), headLen)
	}

	headBuf := bytes.NewReader(headData)

	msg := &Message{}

	if err := binary.Read(headBuf, binary.LittleEndian, &msg.DataLen); err != nil {
		reader.Release()
		return nil, fmt.Errorf("unpack read datalen error: %w", err)
	}

	if err := binary.Read(headBuf, binary.LittleEndian, &msg.Id); err != nil {
		reader.Release()
		return nil, fmt.Errorf("unpack read msgid error: %w", err)
	}

	maxPacketSize := config.GlobalConfig.Server.MaxPacketSize
	if maxPacketSize > 0 && msg.DataLen > maxPacketSize {

		reader.Release()
		return nil, fmt.Errorf("%w: dataLen=%d, maxPacketSize=%d", ErrDataTooLarge, msg.DataLen, maxPacketSize)
	}

	reader.Release()

	return msg, nil
}
