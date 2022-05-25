package tx

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/cosmos/cosmos-sdk/client"
)

// PrivKey - wrapper to expose interface
type PrivKey = cryptotypes.PrivKey

// NewTxBuilder - create TxBuilder
func NewTxBuilder(txConfig client.TxConfig) Builder {
	return Builder{
		TxBuilder: txConfig.NewTxBuilder(),
		TxConfig:  txConfig,
	}
}

// Sign - generate signatures of the tx with given armored private key
// Only support Secp256k1 uses the Bitcoin secp256k1 ECDSA parameters.
func (txBuilder Builder) Sign(
	signMode signing.SignMode, signerData SignerData,
	privKey PrivKey, overwriteSig bool) error {

	// For SIGN_MODE_DIRECT, calling SetSignatures calls setSignerInfos on
	// TxBuilder under the hood, and SignerInfos is needed to generated the
	// sign bytes. This is the reason for setting SetSignatures here, with a
	// nil signature.
	//
	// Note: this line is not needed for SIGN_MODE_LEGACY_AMINO, but putting it
	// also doesn't affect its generated sign bytes, so for code's simplicity
	// sake, we put it here.
	sigData := signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: nil,
	}
	sig := signing.SignatureV2{
		PubKey:   privKey.PubKey(),
		Data:     &sigData,
		Sequence: signerData.Sequence,
	}

	var err error
	var prevSignatures []signing.SignatureV2
	if !overwriteSig {
		prevSignatures, err = txBuilder.GetTx().GetSignaturesV2()
		if err != nil {
			return err
		}
	}

	if err := txBuilder.SetSignatures(sig); err != nil {
		return err
	}

	signature, err := SignWithPrivKey(
		signing.SignMode(signMode),
		authsigning.SignerData(signerData),
		client.TxBuilder(txBuilder.TxBuilder),
		cryptotypes.PrivKey(privKey),
		client.TxConfig(txBuilder.TxConfig),
		signerData.Sequence,
	)

	if err != nil {
		return err
	}

	if overwriteSig {
		return txBuilder.SetSignatures(signature)
	}
	prevSignatures = append(prevSignatures, signature)
	return txBuilder.SetSignatures(prevSignatures...)
}

// GetTxBytes return tx bytes for broadcast
func (txBuilder Builder) GetTxBytes() ([]byte, error) {
	return txBuilder.TxConfig.TxEncoder()(txBuilder.GetTx())
}
