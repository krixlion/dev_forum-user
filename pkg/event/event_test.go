package event

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-user/pkg/helpers/gentest"
	"github.com/krixlion/dev_forum-user/pkg/tracing"
)

func TestMakeEvent(t *testing.T) {
	randString := gentest.RandomString(5)
	randUser := gentest.RandomUser(1, 2, 5)
	type args struct {
		t    EventType
		data interface{}
	}
	testCases := []struct {
		name string
		args args
		want Event
	}{
		{
			name: "Test is correctly serializes event with simple string data",
			args: args{
				t:    ArticleCreated,
				data: randString,
			},
			want: Event{
				AggregateId: tracing.ServiceName,
				Type:        ArticleCreated,
				Body: func() []byte {
					data, err := json.Marshal(randString)
					if err != nil {
						panic(err)
					}
					return data
				}(),
				Timestamp: time.Now(),
			},
		},
		{
			name: "",
			args: args{
				t:    UserCreated,
				data: randUser,
			},
			want: Event{
				AggregateId: tracing.ServiceName,
				Type:        UserCreated,
				Body: func() []byte {
					data, err := json.Marshal(randUser)
					if err != nil {
						panic(err)
					}
					return data
				}(),
				Timestamp: time.Now(),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			if got := MakeEvent(tC.args.t, tC.args.data); !cmp.Equal(got, tC.want, cmpopts.EquateApproxTime(time.Millisecond)) {
				t.Errorf("MakeEvent() = %+v\n want = %+v\n diff = %+v\n", got, tC.want, cmp.Diff(got, tC.want))
				return
			}
		})
	}
}
