package btc

import (
)

type Block struct {
	hash string
	previousHash string
	nextHash string
	height uint32
	timestamp int64
	txCount uint16
}

func NewBlock (hash string, previous string, next string, height uint32, timestamp int64, txCount uint16) Block {
	return Block { hash: hash, previousHash: previous, nextHash: next, height: height, timestamp: timestamp, txCount: txCount }
}

func (b *Block) IsNil () bool {
	return len (b.hash) == 0
}

func (b *Block) GetHash () string {
	return b.hash
}

func (b *Block) GetPreviousHash () string {
	return b.previousHash
}

func (b *Block) GetNextHash () string {
	return b.nextHash
}

func (b *Block) GetHeight () uint32 {
	return b.height
}

func (b *Block) GetTxCount () uint16 {
	return b.txCount
}

func (b *Block) GetTimestamp () int64 {
	return b.timestamp
}

