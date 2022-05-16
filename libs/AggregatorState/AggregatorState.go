package AggregatorState

import (
	"go-liquidator/libs/protobuf"
)

type AggregatorState struct {
	Version uint32
	Configs struct {}
	FulfillmentManagerPubkey []byte
	JobDefinitionPubkeys [][]byte
	Agreement  struct {}
	CurrentRoundResult RoundResult
	LastRoundResult RoundResult
	ParseOptimizedResultAddress []byte
	BundleAuthAddresses [][]byte
}

type RoundResult struct {
	NumSuccess uint32
	NumError uint32
	Result float64
	RoundOpenSlot uint64
	RoundOpenTimestamp int64
	MinResponse float64
	MaxResponse float64
	Medians []float64
}

func (rr *RoundResult) Decode (reader *protobuf.Reader, length uint64) {
	end := reader.Pos + length
	for reader.Pos < end {
		var tag = reader.Uint32();
		switch (tag >> 3) {
		case 1:
			rr.NumSuccess = reader.Uint32()
			break
		case 2:
			rr.NumError = reader.Uint32()
			break
		case 3:
			rr.Result = reader.Double()
			break
		case 4:
			rr.RoundOpenSlot = reader.Uint64();
			break
		case 5:
			rr.RoundOpenTimestamp = reader.Int64();
			break
		case 6:
			rr.MinResponse = reader.Double();
			break
		case 7:
			rr.MaxResponse = reader.Double();
			break
		case 8:
			if (tag & 7) == 2 {
				end2 := uint64(reader.Uint32()) + reader.Pos;
				for (reader.Pos < end2){
					rr.Medians = append(rr.Medians, reader.Double());
				}
			} else {
				rr.Medians = append(rr.Medians, reader.Double());
			}
			break
		default:
			reader.SkipType(int(tag & 7));
			break
		}
	}
}

func (agg *AggregatorState) Decode (reader *protobuf.Reader, length uint64) {
	end := reader.Pos + length
	for reader.Pos < end {
		var tag = reader.Uint32();
		switch (tag >> 3) {
		case 1:
			agg.Version = reader.Uint32();
			break
		case 2:
			_len := reader.Uint32();
			// todo: decode config
			reader.Skip(int64(_len))
			break;
		case 3:
			agg.FulfillmentManagerPubkey = reader.Bytes();
			break
		case 4:
			agg.JobDefinitionPubkeys = append(agg.JobDefinitionPubkeys, reader.Bytes())
			break
		case 5:
			_len := reader.Uint32();
			// todo: decode FulfillmentAgreement
			reader.Skip(int64(_len))
			break
		case 6:
			agg.CurrentRoundResult.Decode(reader, uint64(reader.Uint32()))
			break
		case 7:
			agg.LastRoundResult.Decode(reader, uint64(reader.Uint32()))
			break
		case 8:
			agg.ParseOptimizedResultAddress = reader.Bytes();
			break
		case 9:
			agg.BundleAuthAddresses = append(agg.BundleAuthAddresses, reader.Bytes())
			break
		default:
			reader.SkipType(int(tag & 7));
			break
		}
	}
}

func DecodeDelimited (buffer []byte) AggregatorState {
	reader := protobuf.NewReader(buffer)
	agg := AggregatorState {
		LastRoundResult: RoundResult {
			Result: 0,
		},
	}
	agg.Decode(&reader, uint64(reader.Uint32()))
	return agg
}
