package erg

import (
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	retryClient *retryablehttp.Client
	transport *http.Transport
)

func initConfigs(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	
	// uses the parent directory of os.Getwd()
	viper.AddConfigPath(filepath.Dir(dir))
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	err = viper.ReadInConfig()
	assert.NoError(t, err)

	transport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 3 * time.Second,
		}).Dial,
		MaxIdleConns:        100,
		MaxConnsPerHost:     100,
		MaxIdleConnsPerHost: 100,
		TLSHandshakeTimeout: 3 * time.Second,
	}

	retryClient = retryablehttp.NewClient()
	retryClient.HTTPClient.Transport = transport
	retryClient.HTTPClient.Timeout = time.Second * 10
	retryClient.Logger = nil
	retryClient.RetryWaitMin = 200 * time.Millisecond
	retryClient.RetryWaitMax = 250 * time.Millisecond
	retryClient.RetryMax = 2
}

func TestGetTxFee(t *testing.T) {
	initConfigs(t)

	ergNodeClient, err := NewErgNode(retryClient)
	require.NoError(t, err)

	testCases := []struct {
		name string
		input int
		want int
	}{
		{
			"TestGoodInput",
			2776,
			1000000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			txFee, err := ergNodeClient.GetTxFee(2776)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, txFee, tc.want, "unexpected fee value.")
		})
	}
}

func TestGetUnconfirmedOutputsByErgoTree(t *testing.T) {
	initConfigs(t)

	ergNodeClient, err := NewErgNode(retryClient)
	require.NoError(t, err)

	// get 1 unconfirmed tx from the blockchain
	tx, err := ergNodeClient.GetUnconfirmedTxs(1, 0)
	assert.NoError(t, err)
	assert.Len(t, tx, 1)

	// pull out an ergoTree value to use
	ergoTree := tx[0].Outputs[0].ErgoTree

	boxes, err := ergNodeClient.GetUnconfirmedOutputsByErgoTree(ergoTree, 1, 0)
	assert.NoError(t, err)
	assert.Len(t, boxes, 1, "erg utxo output box not filtered.")
}