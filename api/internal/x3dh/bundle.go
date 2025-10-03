package x3dh

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"time"
	"encoding/base64"

	"go.mongodb.org/mongo-driver/bson"
)

// Bundle represents the public keys a user needs to initiate an X3DH handshake with another user
type Bundle struct {
	IdentityKey   string  `json:"identity_key"`    // IK.pub (base64 or hex)
	SignedPreKey  string  `json:"signed_prekey"`   // SPK.pub
	SPKSignature  string  `json:"spk_signature"`   // Signature of SPK signed by IK
	OneTimePreKey *string `json:"one_time_prekey"` // Optional OPK.pub
}

// BundleService defines the logic for preparing a user's bundle
type BundleService struct {
	store   KeyStore
	crypto  CryptoService
	auditor *MongoAuditLogger
}

type RegisterBundleRequest struct {
	UserID         string                `json:"user_id"`
	IdentityKey    string                `json:"identity_key"`
	SignedPreKey   string                `json:"signed_prekey"`
	SPKSignature   string                `json:"spk_signature"`
	OneTimePreKeys []OneTimePreKeyUpload `json:"one_time_prekeys"`
}

type OneTimePreKeyUpload struct {
	PublicKey string `json:"public_key"`
}

type RefillOPKRequest struct {
	UserID string                `json:"user_id"`
	OPKs   []OneTimePreKeyUpload `json:"opks"`
}
type BundleResponse struct {
	IdentityKey   string `json:"identity_key"`
	SignedPreKey  string `json:"signed_prekey"`
	SPKSignature  string `json:"spk_signature"`
	OneTimePreKey string `json:"one_time_prekey,omitempty"` // optional
}

// NewBundleService returns a new instance of BundleService
func NewBundleService(store KeyStore, crypto CryptoService, auditor *MongoAuditLogger) *BundleService {
	return &BundleService{store: store, crypto: crypto, auditor: auditor}
}
func (s *BundleService) StoreBundle(ctx context.Context, req RegisterBundleRequest) error {
	// Step 1: Verify SPK signature
	err := verifySPKSignature(req.IdentityKey, req.SignedPreKey, req.SPKSignature)
	if err != nil {
		s.logAudit(ctx, req.UserID, "REGISTER_BUNDLE", "failure", "Invalid SPK signature", bson.M{
			"reason": "SPK signature verification failed",
			"route":  "/x3dh/register-bundle",
			"method": "POST",
		})
		return fmt.Errorf("invalid SPK signature: %w", err)
	}

	// Step 2: Store the bundle
	err = s.store.StoreBundle(ctx, req)
	status := "success"
	description := "User uploaded X3DH key bundle"
	if err != nil {
		status = "failure"
		description = "Failed to store X3DH key bundle"
	}

	s.logAudit(ctx, req.UserID, "REGISTER_BUNDLE", status, description, bson.M{
		"num_opks": len(req.OneTimePreKeys),
		"route":    "/x3dh/register-bundle",
		"method":   "POST",
	})

	return err
}
func (s *BundleService) logAudit(ctx context.Context, userID, action, status, description string, metadata bson.M) {
	userAgent, _ := ctx.Value("user-agent").(string)

	log := AuditLog{
		Action:      action,
		Service:     "x3dh",
		Status:      status,
		Description: description,
		Timestamp:   time.Now().UTC(),
		Actor: ActorInfo{
			ID:        userID,
			Role:      "user", // optional
			UserAgent: userAgent,
		},
		Target: TargetInfo{
			Type: "key_bundle",
			ID:   userID,
		},
		Metadata: metadata,
	}

	_ = s.auditor.Log(ctx, log) // log but don’t block
}

// GetBundle prepares a public key bundle for a given user
func (s *BundleService) GetBundle(ctx context.Context, userID string) (*BundleResponse, error) {
	ik, err := s.store.GetIdentityKey(ctx, userID)
	if err != nil {
		s.logAudit(ctx, userID, "GET_BUNDLE", "failure", "Failed to fetch IK", bson.M{
			"step":   "identity_key",
			"route":  "/x3dh/bundle/:user_id",
			"method": "GET",
		})
		return nil, fmt.Errorf("get IK: %w", err)
	}

	spk, err := s.store.GetSignedPreKey(ctx, userID)
	if err != nil {
		s.logAudit(ctx, userID, "GET_BUNDLE", "failure", "Failed to fetch SPK", bson.M{
			"step":   "signed_prekey",
			"route":  "/x3dh/bundle/:user_id",
			"method": "GET",
		})
		return nil, fmt.Errorf("get SPK: %w", err)
	}

	opk, err := s.store.ConsumeOneTimePreKey(ctx, userID)
	if err != nil && err.Error() != "sql: no rows in result set" {
		s.logAudit(ctx, userID, "GET_BUNDLE", "failure", "Failed to consume OPK", bson.M{
			"step":   "one_time_prekey",
			"route":  "/x3dh/bundle/:user_id",
			"method": "GET",
		})
		return nil, fmt.Errorf("get OPK: %w", err)
	}

	s.logAudit(ctx, userID, "GET_BUNDLE", "success", "Fetched X3DH bundle", bson.M{
		"has_opk": opk != nil,
		"route":   "/x3dh/bundle/:user_id",
		"method":  "GET",
	})

	return &BundleResponse{
		IdentityKey:   ik.PublicKey,
		SignedPreKey:  spk.PublicKey,
		SPKSignature:  spk.Signature,
		OneTimePreKey: opk.PublicKey,
	}, nil
}

func (s *BundleService) RefillOPKs(ctx context.Context, userID string, opks []OneTimePreKeyUpload) error {
	err := s.store.InsertOPKs(ctx, userID, opks)

	status := "success"
	description := "Refilled OPKs"
	if err != nil {
		status = "failure"
		description = "Failed to refill OPKs"
	}

	s.logAudit(ctx, userID, "REFILL_OPKS", status, description, bson.M{
		"num_opks": len(opks),
		"method":   "POST",
		"route":    "/x3dh/refill-opks",
	})

	return err
}
func decodeAnyB64(s string) ([]byte, error) {
	// Try URL-safe without padding first
	if decoded, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return decoded, nil
	}
	// Try URL-safe with padding
	if decoded, err := base64.URLEncoding.DecodeString(s); err == nil {
		return decoded, nil
	}
	// Fallback to standard base64
	return base64.StdEncoding.DecodeString(s)
}

func verifySPKSignature(identityKeyHex, spkHex, sigHex string) error {
	identityKey, err := hex.DecodeString(identityKeyHex)
	if err != nil {
		return fmt.Errorf("invalid identity key: %w", err)
	}
	if len(identityKey) != ed25519.PublicKeySize {
		return fmt.Errorf("identity key length invalid")
	}

	spk, err := hex.DecodeString(spkHex)
	if err != nil {
		return fmt.Errorf("invalid signed prekey: %w", err)
	}

	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		return fmt.Errorf("invalid signature: %w", err)
	}
	if len(sig) != ed25519.SignatureSize {
		return fmt.Errorf("signature length invalid")
	}

	if !ed25519.Verify(identityKey, spk, sig) {
		return fmt.Errorf("signature verification failed")
	}
	return nil
}

func (s *BundleService) RotateSPK(ctx context.Context, userID, newSPK, signature string, expiresAt *time.Time) error {
	// 🔐 (Optional) Verify the signature before rotating
	return s.store.RotateSignedPreKey(ctx, userID, newSPK, signature, expiresAt)
}
