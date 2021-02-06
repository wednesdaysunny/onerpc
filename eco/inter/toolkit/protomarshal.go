package toolkit

import (
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	std "github.com/wednesdaysunny/onerpc/eco/inter"
)

func MarshalResp(v interface{}) ([]byte, error) {
	if msg, ok := v.(proto.Message); !ok {
		return nil, std.ErrRpcCacheMarshal
	} else if b, err := proto.Marshal(msg); err != nil {
		return nil, std.ErrRpcCacheMarshal
	} else {
		return b, nil
	}
}

func UnmarshalResp(b []byte, v interface{}) error {
	if msg, ok := v.(proto.Message); !ok {
		return std.ErrRpcCacheUnmarshal
	} else if err := proto.Unmarshal(b, msg); err != nil {
		return std.ErrRpcCacheUnmarshal
	} else {
		return nil
	}
}

func T2Pb(t time.Time) *timestamp.Timestamp {
	pb, _ := ptypes.TimestampProto(t)
	return pb
}

func Pb2t(pb *timestamp.Timestamp) time.Time {
	t, _ := ptypes.Timestamp(pb)
	t = t.UTC()
	return t
}
