package main

import (
	"fmt"
	"fractus/pkg/dbflat"
)

type transaction struct {
	sender  string // hotfield
	recever string // hotfield
	amount  uint64 // hotfield
	nonce   []byte // coldfield
}

type transactionv2 struct {
	sender  string // hotfield
	recever string // hotfield
	amount  uint64 // hotfield
	nonce   []byte // coldfield
	id      string // new field
}

type format struct {
	build   *dbflat.Builder
	inspect *dbflat.Inspector
	buf     [][]byte
}

func NewFormat() *format {
	return &format{build: dbflat.NewBuilder(nil), inspect: dbflat.NewInspect(nil)}
}

func (f *format) Save(data []byte) {
	if f.buf != nil {
		f.buf = append(f.buf, data)
	}
	f.buf = make([][]byte, 0)
	f.buf = append(f.buf, data)
}

func (f *format) Return() [][]byte {
	return f.buf
}
func Clone(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

func (f *format) FromTx(txs []transaction) [][]byte {
	var encoded [][]byte
	for _, tx := range txs {
		f.build.AddField(uint16(1), uint16(0x8000), dbflat.ToBytes(tx.sender), true)
		f.build.AddField(uint16(2), uint16(0x8000), dbflat.ToBytes(tx.recever), true)
		f.build.AddField(uint16(3), uint16(0x8000), dbflat.ToBytes(tx.amount), true)
		//f.build.AddField(uint16(4), uint16(0x8000), dbflat.ToBytes(tx.sender), false)
		f.build.AddField(uint16(10), uint16(0x8000), dbflat.ToBytes(tx.nonce), false) // []byte so it's array
		f.build.Commit(uint64(15026), dbflat.FlagPadding|dbflat.FlagNoSchemaID)
		tmp := f.build.Output()
		// encoded get overwrite and old transactions are replaced with new ones
		// seems like []byte use pointers so create a helper and use it
		encoded = append(encoded, Clone(tmp))
		//f.Save(tmp)
		f.build.Reset()
	}
	return encoded
}

func (f *format) ToTx(enctx [][]byte) []transaction {
	var decoded []transaction

	for _, tx := range enctx {
		f.inspect.Insert(tx)
		var tmp uint64
		tmpamount, _ := dbflat.ReadAny(f.inspect.GetField(3), dbflat.TypeUint64)
		switch v := tmpamount.(type) {
		case uint64:
			tmp = v
		default:
			tmp = 0
		}
		// data will still be parsed as version 1 and unknow fields will be ignored
		decoded = append(decoded, transaction{
			sender:  string(f.inspect.GetField(1)),
			recever: string(f.inspect.GetField(2)),
			nonce:   f.inspect.GetField(10),
			amount:  tmp,
		})
	}
	f.inspect.Reset()
	return decoded
}

func (f *format) cnv(tx transactionv2) []byte {
	var encoded []byte
	f.build.AddField(uint16(1), uint16(0x8000), dbflat.ToBytes(tx.sender), true)
	f.build.AddField(uint16(2), uint16(0x8000), dbflat.ToBytes(tx.recever), true)
	f.build.AddField(uint16(3), uint16(0x8000), dbflat.ToBytes(tx.amount), true)
	//f.build.AddField(uint16(4), uint16(0x8000), dbflat.ToBytes(tx.sender), false)
	f.build.AddField(uint16(10), uint16(0x8000), dbflat.ToBytes(tx.nonce), false)
	f.build.AddField(uint16(4), uint16(0x8000), dbflat.ToBytes(tx.id), true)
	f.build.Commit(uint64(15026), dbflat.FlagPadding|dbflat.FlagNoSchemaID)
	tmp := f.build.Output()
	encoded = append(encoded, tmp...)
	//f.Save(tmp)
	f.build.Reset()
	return encoded
}
func case1() {
	f := NewFormat()
	txs := MakeTxs("small")
	encoded := f.FromTx(txs)
	//fmt.Print(encoded)
	fmt.Print(f.ToTx(encoded))

}

func case2() {
	f := NewFormat()
	txs := MakeTxs("small")
	txs2 := transactionv2{
		sender:  "Aren12",
		recever: "Hub20",
		amount:  uint64(158),
		nonce:   []byte{0x00, 0x01, 0x02},
		id:      "World"}
	enc := f.FromTx(txs)
	enc = append(enc, Clone(f.cnv(txs2)))
	fmt.Print(f.ToTx(enc))
}

func main() {
	fmt.Print("Case 1: same version: \n")
	case1() // same version

	fmt.Print("\n New version + Old version: \n")
	case2() // shows forward/backway compatibility
}

func MakeTxs(size string) []transaction {
	switch size {
	case "small":
		return []transaction{
			{sender: "Aren12", recever: "Hub20", amount: uint64(158), nonce: []byte{0x00, 0x01, 0x02}},
			{sender: "Gita", recever: "Subd", amount: uint64(200), nonce: []byte{0x00, 0x01, 0x02}},
			{sender: "Sub", recever: "Hub", amount: uint64(452), nonce: []byte{0x00, 0x01, 0x02}},
			{sender: "HubsGit", recever: "GitHubs", amount: uint64(456), nonce: []byte{0x00, 0x01, 0x02}},
			{sender: "Golang", recever: "Python", amount: uint64(560), nonce: []byte{0x00, 0x01, 0x02}},
			{sender: "C", recever: "C#", amount: uint64(800), nonce: []byte{0x00, 0x01, 0x02}},
		}
	default:
		return nil
	}

}
