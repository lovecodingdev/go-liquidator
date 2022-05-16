package protobuf

import (
	"fmt"
	// "sync"
	// "context"
	// "math"
	// "math/big"
  "bytes"
	"encoding/binary"

	// . "go-liquidator/global"

	// "github.com/portto/solana-go-sdk/client"
	// "github.com/portto/solana-go-sdk/common"
	"google.golang.org/protobuf/encoding/protowire"
)

// Type represents the wire type.
type Type int8

const (
	VarintType     Type = 0
	Fixed32Type    Type = 5
	Fixed64Type    Type = 1
	BytesType      Type = 2
	StartGroupType Type = 3
	EndGroupType   Type = 4
)

const (
	_ = -iota
	errCodeTruncated
	errCodeFieldNumber
	errCodeOverflow
	errCodeReserved
	errCodeEndGroup
	errCodeRecursionDepth
)


type Reader struct {
	buf []byte
	Pos uint64
	Len uint64
	reader *bytes.Reader
}

func NewReader(buf []byte) Reader {
	reader := bytes.NewReader(buf)

	return Reader {
		buf: buf,
		Pos: 0,
		Len: uint64(len(buf)),
		reader: reader,
	}
}

func (r *Reader) Uint32() uint32 {
	v, n := protowire.ConsumeVarint(r.buf[r.Pos:])
	r.Pos += uint64(n)
	return uint32(v)
}

func (r *Reader) Double() float64 {
	/* istanbul ignore if */
	if (r.Pos + 8 > r.Len){
		fmt.Println("throw indexOutOfRange(this, 4)")
	}

	var d float64
	r.reader.Seek(int64(r.Pos), 0)
	binary.Read(r.reader, binary.LittleEndian, &d)
	r.Pos += 8;

	return d;
}

func (r *Reader) Uint64() uint64 {
	v, n := protowire.ConsumeVarint(r.buf[r.Pos:])
	r.Pos += uint64(n)
	return uint64(v)
}

func (r *Reader) Int64() int64 {
	v, n := protowire.ConsumeVarint(r.buf[r.Pos:])
	r.Pos += uint64(n)
	return int64(v)
}

func (r *Reader) Bytes () []byte {
	length := r.Uint32()
	start  := r.Pos
	end    := r.Pos + uint64(length)

	/* istanbul ignore if */
	if end > r.Len {
		fmt.Println("throw indexOutOfRange(this, length)", length)
		return r.buf[start:r.Len]
	}

	r.Pos += uint64(length)
	return r.buf[start:end]
};

func (r *Reader) String() string {
	var bytes = r.Bytes()
	return string(bytes)
};

func (r *Reader) Skip(length int64) {
	if (length >= 0) {
		/* istanbul ignore if */
		if (r.Pos + uint64(length) > r.Len){
			fmt.Println("throw indexOutOfRange(this, length)");
			r.Pos = r.Len
			return
		}
		r.Pos += uint64(length)
	} else {
		for {
			/* istanbul ignore if */
			if (r.Pos >= r.Len){
				fmt.Println("throw indexOutOfRange(this, length)", length);
				return
			}
			if r.buf[r.Pos] & 128 == 0{
				r.Pos++
				break
			}
			r.Pos++
		}
	}
}

func (r *Reader) SkipType(wireType int) {
	switch (wireType) {
		case 0: //VarintType
			r.Skip(-1);
			break
		case 1: //Fixed64Type
			r.Skip(8);
			break
		case 2: //BytesType
			r.Skip(int64(r.Uint32()));
			break
		case 3: //StartGroupType
			wireType = int(r.Uint32()) & 7
			for (wireType != 4) {
				r.SkipType(wireType)
				wireType = int(r.Uint32()) & 7
			}
			break
		case 4: //EndGroupType 
			break
		case 5: //Fixed32Type 
			r.Skip(4);
			break

		/* istanbul ignore next */
		default:
			fmt.Printf("invalid wire type %d at offset %d \n", wireType, r.Pos)
	}
};
