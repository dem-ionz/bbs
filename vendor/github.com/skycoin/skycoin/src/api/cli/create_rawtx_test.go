package cli

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/testutil"
	"github.com/skycoin/skycoin/src/util/fee"
	"github.com/skycoin/skycoin/src/visor"
	"github.com/skycoin/skycoin/src/wallet"
)

func TestMakeChangeOut(t *testing.T) {
	uxOuts := []wallet.UxBalance{
		{
			Hash:    cipher.MustSHA256FromHex("f569461182b0efe9a5c666e9a35c6602b351021c1803cc740aca548cf6db4cb2"),
			Address: cipher.MustDecodeBase58Address("k3rmz3PGbTxd7KL8AL5CeHrWy35C1UcWND"),
			BkSeq:   10,
			Coins:   400e6,
			Hours:   200,
		},
		{
			Hash:    cipher.MustSHA256FromHex("bddf0aaf80f96c144f33ac8a27764a868d37e1c11e568063ebeb1367de859566"),
			Address: cipher.MustDecodeBase58Address("A2h4iWC1SDGmS6UPezatFzEUwirLJtjFUe"),
			BkSeq:   11,
			Coins:   300e6,
			Hours:   100,
		},
	}

	spendAmt := []SendAmount{{
		Addr:  "2PBmUva7J8WFsyWg979cREZkU3z2pkYjNkE",
		Coins: 600e6,
	}}

	chgAddr := "2konv5no3DZvSMxf2GPVtAfZinfwqCGhfVQ"
	_, err := cipher.DecodeBase58Address(chgAddr)
	require.NoError(t, err)

	txOuts, err := makeChangeOut(uxOuts, chgAddr, spendAmt)
	require.NoError(t, err)
	require.NotEmpty(t, txOuts)

	// Should have a change output and an output to the destination in toAddrs
	require.Len(t, txOuts, 2)

	chgOut := txOuts[0]
	t.Logf("chgOut:%+v\n", chgOut)
	require.Equal(t, chgAddr, chgOut.Address.String())
	require.Exactly(t, uint64(100e6), chgOut.Coins)

	spendOut := txOuts[1]
	require.Equal(t, spendAmt[0].Addr, spendOut.Address.String())
	require.Exactly(t, spendAmt[0].Coins, spendOut.Coins)

	var totalHours uint64
	for _, ux := range uxOuts {
		totalHours += ux.Hours
	}
	remainingHours := totalHours - fee.RequiredFee(totalHours)
	chgHours := remainingHours / 2
	if remainingHours%2 != 0 {
		chgHours++
	}
	spendHours := remainingHours - chgHours

	require.Exactly(t, chgHours, chgOut.Hours)
	require.Exactly(t, spendHours, spendOut.Hours)
}

func TestMakeChangeOutOneCoinHour(t *testing.T) {
	// As long as there is at least one coin hour left, creating a transaction
	// will still succeed
	uxOuts := []wallet.UxBalance{
		{
			Hash:    cipher.MustSHA256FromHex("f569461182b0efe9a5c666e9a35c6602b351021c1803cc740aca548cf6db4cb2"),
			BkSeq:   10,
			Address: cipher.MustDecodeBase58Address("k3rmz3PGbTxd7KL8AL5CeHrWy35C1UcWND"),
			Coins:   400e6,
			Hours:   0,
		},
		{
			Hash:    cipher.MustSHA256FromHex("bddf0aaf80f96c144f33ac8a27764a868d37e1c11e568063ebeb1367de859566"),
			BkSeq:   11,
			Address: cipher.MustDecodeBase58Address("A2h4iWC1SDGmS6UPezatFzEUwirLJtjFUe"),
			Coins:   300e6,
			Hours:   1,
		},
	}

	spendAmt := []SendAmount{{
		Addr:  "2PBmUva7J8WFsyWg979cREZkU3z2pkYjNkE",
		Coins: 600e6,
	}}

	chgAddr := "2konv5no3DZvSMxf2GPVtAfZinfwqCGhfVQ"
	_, err := cipher.DecodeBase58Address(chgAddr)
	require.NoError(t, err)

	txOuts, err := makeChangeOut(uxOuts, chgAddr, spendAmt)
	require.NoError(t, err)
	require.NotEmpty(t, txOuts)

	// Should have a change output and an output to the destination in toAddrs
	require.Len(t, txOuts, 2)

	chgOut := txOuts[0]
	t.Logf("chgOut:%+v\n", chgOut)
	require.Equal(t, chgAddr, chgOut.Address.String())
	require.Exactly(t, uint64(100e6), chgOut.Coins)
	require.Exactly(t, uint64(0), chgOut.Hours)

	spendOut := txOuts[1]
	require.Equal(t, spendAmt[0].Addr, spendOut.Address.String())
	require.Exactly(t, spendAmt[0].Coins, spendOut.Coins)
	require.Exactly(t, uint64(0), spendOut.Hours)
}

func TestMakeChangeOutInsufficientCoinHours(t *testing.T) {
	// If there are no coin hours in the inputs, creating the txn will fail
	// because it will not be accepted by the network
	uxOuts := []wallet.UxBalance{
		{
			Hash:    cipher.MustSHA256FromHex("f569461182b0efe9a5c666e9a35c6602b351021c1803cc740aca548cf6db4cb2"),
			BkSeq:   10,
			Address: cipher.MustDecodeBase58Address("k3rmz3PGbTxd7KL8AL5CeHrWy35C1UcWND"),
			Coins:   400e6,
			Hours:   0,
		},
		{
			Hash:    cipher.MustSHA256FromHex("bddf0aaf80f96c144f33ac8a27764a868d37e1c11e568063ebeb1367de859566"),
			BkSeq:   11,
			Address: cipher.MustDecodeBase58Address("A2h4iWC1SDGmS6UPezatFzEUwirLJtjFUe"),
			Coins:   300e6,
			Hours:   0,
		},
	}

	spendAmt := []SendAmount{{
		Addr:  "2PBmUva7J8WFsyWg979cREZkU3z2pkYjNkE",
		Coins: 600e6,
	}}

	chgAddr := "2konv5no3DZvSMxf2GPVtAfZinfwqCGhfVQ"
	_, err := cipher.DecodeBase58Address(chgAddr)
	require.NoError(t, err)

	_, err = makeChangeOut(uxOuts, chgAddr, spendAmt)
	testutil.RequireError(t, err, fee.ErrTxnNoFee.Error())
}

func TestChooseSpends(t *testing.T) {
	// Start with visor.ReadableOutputSet
	// Spends should be minimized

	// Insufficient HeadOutputs
	// Sufficient HeadOutputs, but insufficient after adjusting for OutgoingOutputs
	// Insufficient HeadOutputs, but sufficient after adjusting for IncomingOutputs
	// Sufficient HeadOutputs after adjusting for OutgoingOutputs

	var coins uint64 = 100e6

	hashA := testutil.RandSHA256(t).Hex()
	hashB := testutil.RandSHA256(t).Hex()
	hashC := testutil.RandSHA256(t).Hex()
	hashD := testutil.RandSHA256(t).Hex()

	addrA := testutil.MakeAddress().String()
	addrB := testutil.MakeAddress().String()
	addrC := testutil.MakeAddress().String()
	addrD := testutil.MakeAddress().String()

	cases := []struct {
		name     string
		err      error
		spendLen int
		ros      visor.ReadableOutputSet
	}{
		{
			"Insufficient HeadOutputs",
			wallet.ErrInsufficientBalance,
			0,
			visor.ReadableOutputSet{
				HeadOutputs: visor.ReadableOutputs{
					{
						Hash:    hashA,
						Address: addrA,
						BkSeq:   22,
						Coins:   "75.000000",
						Hours:   100,
					},
					{
						Hash:    hashB,
						Address: addrB,
						BkSeq:   19,
						Coins:   "13.000000",
						Hours:   0,
					},
				},
			},
		},

		{
			"Sufficient HeadOutputs, but insufficient after subtracting OutgoingOutputs",
			wallet.ErrInsufficientBalance,
			0,
			visor.ReadableOutputSet{
				HeadOutputs: visor.ReadableOutputs{
					{
						Hash:    hashA,
						Address: addrA,
						BkSeq:   22,
						Coins:   "75.000000",
						Hours:   100,
					},
					{
						Hash:    hashB,
						Address: addrB,
						BkSeq:   19,
						Coins:   "50.000000",
						Hours:   0,
					},
				},
				OutgoingOutputs: visor.ReadableOutputs{
					{
						Hash:    hashB,
						Address: addrB,
						BkSeq:   19,
						Coins:   "50.000000",
						Hours:   0,
					},
				},
			},
		},

		{
			"Insufficient HeadOutputs, but sufficient after adding IncomingOutputs",
			ErrTemporaryInsufficientBalance,
			0,
			visor.ReadableOutputSet{
				HeadOutputs: visor.ReadableOutputs{
					{
						Hash:    hashA,
						Address: addrA,
						BkSeq:   22,
						Coins:   "20.000000",
						Hours:   100,
					},
					{
						Hash:    hashB,
						Address: addrB,
						BkSeq:   19,
						Coins:   "30.000000",
						Hours:   0,
					},
				},
				IncomingOutputs: visor.ReadableOutputs{
					{
						Hash:    hashC,
						Address: addrC,
						BkSeq:   134,
						Coins:   "40.000000",
						Hours:   200,
					},
					{
						Hash:    hashD,
						Address: addrD,
						BkSeq:   29,
						Coins:   "11.000000",
						Hours:   0,
					},
				},
			},
		},

		{
			"Sufficient HeadOutputs and still sufficient after subtracting OutgoingOutputs",
			nil,
			2,
			visor.ReadableOutputSet{
				HeadOutputs: visor.ReadableOutputs{
					{
						Hash:    hashA,
						Address: addrA,
						BkSeq:   22,
						Coins:   "15.000000",
						Hours:   100,
					},
					{
						Hash:    hashB,
						Address: addrB,
						BkSeq:   19,
						Coins:   "90.000000",
						Hours:   0,
					},
					{
						Hash:    hashC,
						Address: addrC,
						BkSeq:   19,
						Coins:   "20.000000",
						Hours:   1,
					},
				},
				OutgoingOutputs: visor.ReadableOutputs{
					{
						Hash:    hashA,
						Address: addrA,
						BkSeq:   22,
						Coins:   "15.000000",
						Hours:   100,
					},
				},
			},
		},

		{
			"Sufficient HeadOutputs and still sufficient after subtracting OutgoingOutputs but will have no coinhours",
			fee.ErrTxnNoFee,
			0,
			visor.ReadableOutputSet{
				HeadOutputs: visor.ReadableOutputs{
					{
						Hash:    hashA,
						Address: addrA,
						BkSeq:   22,
						Coins:   "15.000000",
						Hours:   100,
					},
					{
						Hash:    hashB,
						Address: addrB,
						BkSeq:   19,
						Coins:   "90.000000",
						Hours:   0,
					},
					{
						Hash:    hashC,
						Address: addrC,
						BkSeq:   19,
						Coins:   "30.000000",
						Hours:   0,
					},
				},
				OutgoingOutputs: visor.ReadableOutputs{
					{
						Hash:    hashA,
						Address: addrA,
						BkSeq:   22,
						Coins:   "15.000000",
						Hours:   100,
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spends, err := chooseSpends(tc.ros, coins)

			if tc.err != nil {
				testutil.RequireError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.spendLen, len(spends))

				var totalCoins uint64
				for _, ux := range spends {
					totalCoins += ux.Coins
				}

				require.True(t, coins <= totalCoins)
			}
		})
	}
}
