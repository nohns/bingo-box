package mongo_test

import (
	"context"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/nohns/bingo-box/server/mongo"
	"github.com/stretchr/testify/require"
	"github.com/tryvium-travels/memongo"
	"github.com/tryvium-travels/memongo/mongobin"
)

var mongod *memongo.Server
var sharedDB *mongo.DB

func TestMain(m *testing.M) {
	opts := &memongo.Options{
		MongoVersion: "5.0.4",
	}
	// Workaround for macos with arm processors. The binary will run under rosetta 2, for the time being
	if runtime.GOARCH == "arm64" && runtime.GOOS == "darwin" {
		dlSpec := mongobin.DownloadSpec{
			Version:        opts.MongoVersion,
			Platform:       "osx",
			OSName:         "",
			SSLBuildNeeded: false,
			Arch:           "x86_64",
		}
		opts.DownloadURL = dlSpec.GetDownloadURL()
	}
	mongoServ, err := memongo.StartWithOptions(opts)
	if err != nil {
		log.Fatalf("failed to start mongod in memory:\n%v\n", err)
	}
	mongod = mongoServ
	defer mongod.Stop()

	ctx := context.Background()
	sharedDB, err = mongo.New(ctx, mongod.URIWithRandomDB())
	if err != nil {
		log.Fatalf("failed mongodb setup:\n%v\n", err)
	}
	defer sharedDB.Close(ctx)

	os.Exit(m.Run())
}

func TestMongoNew(t *testing.T) {

	cases := []struct {
		cn            string
		uri           string
		ctx           context.Context
		timeoutAfter  time.Duration
		expectErr     bool
		expectedErrIs error
	}{
		{
			cn:            "success",
			uri:           mongod.URIWithRandomDB(),
			ctx:           context.Background(),
			expectErr:     false,
			expectedErrIs: nil,
		},
		{
			cn:            "fail no database",
			uri:           mongod.URI(),
			ctx:           context.Background(),
			expectErr:     true,
			expectedErrIs: mongo.ErrNoDatabase,
		},
		{
			cn:            "fail invalid conn uri",
			uri:           "i am invalid conn uri",
			ctx:           context.Background(),
			expectErr:     true,
			expectedErrIs: mongo.ErrInvalidConnUri,
		},
		{
			cn:            "fail no server",
			uri:           strings.Replace(mongod.URIWithRandomDB(), "localhost:"+strconv.Itoa(mongod.Port()), "localhost:80", 1),
			ctx:           context.Background(),
			timeoutAfter:  2 * time.Second,
			expectErr:     true,
			expectedErrIs: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.cn, func(t *testing.T) {
			ctx := c.ctx
			if c.timeoutAfter > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, c.timeoutAfter)
				defer cancel()
			}

			db, err := mongo.New(ctx, c.uri)
			if c.expectErr {
				require.Error(t, err, "expected error")
				require.Nil(t, db, "expected returned db pointer to be nil")
				if c.expectedErrIs != nil {
					require.ErrorIs(t, err, c.expectedErrIs, "expected different error")
				}

			} else {
				require.NoError(t, err, "expected no error")
				require.NotNil(t, db, "expected non-nil pointer to db")

				err := db.Close(ctx)
				require.NoError(t, err, "expected no error from db.Close()")
			}
		})
	}
}
