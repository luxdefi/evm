// (c) 2023, Lux Partners Limited. All rights reserved.
// See the file LICENSE for licensing terms.

package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/luxdefi/node/database/memdb"
	"github.com/luxdefi/node/ids"
	"github.com/luxdefi/node/snow"
	"github.com/luxdefi/node/snow/choices"
	"github.com/luxdefi/node/snow/consensus/snowman"
	"github.com/luxdefi/node/snow/engine/common"
	"github.com/luxdefi/node/snow/engine/snowman/block"
	"github.com/luxdefi/node/utils/crypto/bls"
	luxWarp "github.com/luxdefi/node/vms/platformvm/warp"
	"github.com/luxdefi/evm/plugin/evm/message"
	"github.com/luxdefi/evm/warp"
	"github.com/stretchr/testify/require"
)

func TestMessageSignatureHandler(t *testing.T) {
	database := memdb.New()
	snowCtx := snow.DefaultContextTest()
	blsSecretKey, err := bls.NewSecretKey()
	require.NoError(t, err)

	warpSigner := luxWarp.NewSigner(blsSecretKey, snowCtx.NetworkID, snowCtx.ChainID)
	backend := warp.NewBackend(snowCtx.NetworkID, snowCtx.ChainID, warpSigner, &block.TestVM{TestVM: common.TestVM{T: t}}, database, 100)

	msg, err := luxWarp.NewUnsignedMessage(snowCtx.NetworkID, snowCtx.ChainID, []byte("test"))
	require.NoError(t, err)

	messageID := msg.ID()
	require.NoError(t, backend.AddMessage(msg))
	signature, err := backend.GetMessageSignature(messageID)
	require.NoError(t, err)
	unknownMessageID := ids.GenerateTestID()

	emptySignature := [bls.SignatureLen]byte{}

	tests := map[string]struct {
		setup       func() (request message.MessageSignatureRequest, expectedResponse []byte)
		verifyStats func(t *testing.T, stats *handlerStats)
	}{
		"known message": {
			setup: func() (request message.MessageSignatureRequest, expectedResponse []byte) {
				return message.MessageSignatureRequest{
					MessageID: messageID,
				}, signature[:]
			},
			verifyStats: func(t *testing.T, stats *handlerStats) {
				require.EqualValues(t, 1, stats.messageSignatureRequest.Count())
				require.EqualValues(t, 1, stats.messageSignatureHit.Count())
				require.EqualValues(t, 0, stats.messageSignatureMiss.Count())
				require.EqualValues(t, 0, stats.blockSignatureRequest.Count())
				require.EqualValues(t, 0, stats.blockSignatureHit.Count())
				require.EqualValues(t, 0, stats.blockSignatureMiss.Count())
			},
		},
		"unknown message": {
			setup: func() (request message.MessageSignatureRequest, expectedResponse []byte) {
				return message.MessageSignatureRequest{
					MessageID: unknownMessageID,
				}, emptySignature[:]
			},
			verifyStats: func(t *testing.T, stats *handlerStats) {
				require.EqualValues(t, 1, stats.messageSignatureRequest.Count())
				require.EqualValues(t, 0, stats.messageSignatureHit.Count())
				require.EqualValues(t, 1, stats.messageSignatureMiss.Count())
				require.EqualValues(t, 0, stats.blockSignatureRequest.Count())
				require.EqualValues(t, 0, stats.blockSignatureHit.Count())
				require.EqualValues(t, 0, stats.blockSignatureMiss.Count())
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			handler := NewSignatureRequestHandler(backend, message.Codec)
			handler.stats.Clear()

			request, expectedResponse := test.setup()
			responseBytes, err := handler.OnMessageSignatureRequest(context.Background(), ids.GenerateTestNodeID(), 1, request)
			require.NoError(t, err)

			test.verifyStats(t, handler.stats)

			// If the expected response is empty, assert that the handler returns an empty response and return early.
			if len(expectedResponse) == 0 {
				require.Len(t, responseBytes, 0, "expected response to be empty")
				return
			}
			var response message.SignatureResponse
			_, err = message.Codec.Unmarshal(responseBytes, &response)
			require.NoError(t, err, "error unmarshalling SignatureResponse")

			require.Equal(t, expectedResponse, response.Signature[:])
		})
	}
}

func TestBlockSignatureHandler(t *testing.T) {
	database := memdb.New()
	snowCtx := snow.DefaultContextTest()
	blsSecretKey, err := bls.NewSecretKey()
	require.NoError(t, err)

	warpSigner := luxWarp.NewSigner(blsSecretKey, snowCtx.NetworkID, snowCtx.ChainID)
	blkID := ids.GenerateTestID()
	testVM := &block.TestVM{
		TestVM: common.TestVM{T: t},
		GetBlockF: func(ctx context.Context, i ids.ID) (snowman.Block, error) {
			if i == blkID {
				return &snowman.TestBlock{
					TestDecidable: choices.TestDecidable{
						IDV:     blkID,
						StatusV: choices.Accepted,
					},
				}, nil
			}
			return nil, errors.New("invalid blockID")
		},
	}
	backend := warp.NewBackend(
		snowCtx.NetworkID,
		snowCtx.ChainID,
		warpSigner,
		testVM,
		database,
		100,
	)

	signature, err := backend.GetBlockSignature(blkID)
	require.NoError(t, err)
	unknownMessageID := ids.GenerateTestID()

	emptySignature := [bls.SignatureLen]byte{}

	tests := map[string]struct {
		setup       func() (request message.BlockSignatureRequest, expectedResponse []byte)
		verifyStats func(t *testing.T, stats *handlerStats)
	}{
		"known block": {
			setup: func() (request message.BlockSignatureRequest, expectedResponse []byte) {
				return message.BlockSignatureRequest{
					BlockID: blkID,
				}, signature[:]
			},
			verifyStats: func(t *testing.T, stats *handlerStats) {
				require.EqualValues(t, 0, stats.messageSignatureRequest.Count())
				require.EqualValues(t, 0, stats.messageSignatureHit.Count())
				require.EqualValues(t, 0, stats.messageSignatureMiss.Count())
				require.EqualValues(t, 1, stats.blockSignatureRequest.Count())
				require.EqualValues(t, 1, stats.blockSignatureHit.Count())
				require.EqualValues(t, 0, stats.blockSignatureMiss.Count())
			},
		},
		"unknown block": {
			setup: func() (request message.BlockSignatureRequest, expectedResponse []byte) {
				return message.BlockSignatureRequest{
					BlockID: unknownMessageID,
				}, emptySignature[:]
			},
			verifyStats: func(t *testing.T, stats *handlerStats) {
				require.EqualValues(t, 0, stats.messageSignatureRequest.Count())
				require.EqualValues(t, 0, stats.messageSignatureHit.Count())
				require.EqualValues(t, 0, stats.messageSignatureMiss.Count())
				require.EqualValues(t, 1, stats.blockSignatureRequest.Count())
				require.EqualValues(t, 0, stats.blockSignatureHit.Count())
				require.EqualValues(t, 1, stats.blockSignatureMiss.Count())
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			handler := NewSignatureRequestHandler(backend, message.Codec)
			handler.stats.Clear()

			request, expectedResponse := test.setup()
			responseBytes, err := handler.OnBlockSignatureRequest(context.Background(), ids.GenerateTestNodeID(), 1, request)
			require.NoError(t, err)

			test.verifyStats(t, handler.stats)

			// If the expected response is empty, assert that the handler returns an empty response and return early.
			if len(expectedResponse) == 0 {
				require.Len(t, responseBytes, 0, "expected response to be empty")
				return
			}
			var response message.SignatureResponse
			_, err = message.Codec.Unmarshal(responseBytes, &response)
			require.NoError(t, err, "error unmarshalling SignatureResponse")

			require.Equal(t, expectedResponse, response.Signature[:])
		})
	}
}
