package unit_tests

import (
	"aegis-api/internal/x3dh"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// AuditLogger interface for audit logging operations
type AuditLogger interface {
	Log(ctx context.Context, log x3dh.AuditLog) error
}

// Bundle represents the public keys a user needs to initiate an X3DH handshake with another user
type Bundle struct {
	IdentityKey   string  `json:"identity_key"`    // IK.pub (base64 or hex)
	SignedPreKey  string  `json:"signed_prekey"`   // SPK.pub
	SPKSignature  string  `json:"spk_signature"`   // Signature of SPK signed by IK
	OneTimePreKey *string `json:"one_time_prekey"` // Optional OPK.pub
}

// BundleService defines the logic for preparing a user's bundle
type BundleService struct {
	store   x3dh.KeyStore
	crypto  x3dh.CryptoService
	auditor AuditLogger // Changed from *MongoAuditLogger to AuditLogger interface
}

type RegisterBundleRequest struct {
	UserID         string                `json:"user_id"`
	IdentityKey    string                `json:"identity_key"`
	SignedPreKey   string                `json:"signed_prekey"`
	SPKSignature   string                `json:"spk_signature"`
	OneTimePreKeys []OneTimePreKeyUpload `json:"one_time_prekeys"`
}

type OneTimePreKeyUpload struct {
	KeyID     string `json:"key_id"`
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
	OneTimePreKey string `json:"one_time_prekey,omitempty"`
	OPKID         string `json:"opk_id,omitempty"`
}

type RotateSPKRequest struct {
	UserID    string  `json:"user_id"`
	NewSPK    string  `json:"new_spk"`
	Signature string  `json:"signature"`
	ExpiresAt *string `json:"expires_at,omitempty"`
}

// NewBundleService returns a new instance of BundleService
func NewBundleService(store x3dh.KeyStore, crypto x3dh.CryptoService, auditor AuditLogger) *BundleService {
	return &BundleService{store: store, crypto: crypto, auditor: auditor}
}

// logAudit is a helper method to log audit events
func (s *BundleService) logAudit(ctx context.Context, userID, action, status, description string, metadata bson.M) {
	userAgent, _ := ctx.Value("user-agent").(string)

	log := x3dh.AuditLog{
		Action:      action,
		Service:     "x3dh",
		Status:      status,
		Description: description,
		Timestamp:   time.Now().UTC(),
		Actor: x3dh.ActorInfo{
			ID:        userID,
			Role:      "user", // optional
			UserAgent: userAgent,
		},
		Target: x3dh.TargetInfo{
			Type: "key_bundle",
			ID:   userID,
		},
		Metadata: metadata,
	}

	_ = s.auditor.Log(ctx, log) // log but don't block
}

// decodeAnyB64 tries multiple base64 decoding methods
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

// verifySPKSignatureBase64 verifies the signature of a signed pre-key
func verifySPKSignatureBase64(identityKeyB64, spkB64, sigB64 string) error {
	log.Printf("[x3dh] IK base64 len=%d", len(identityKeyB64))

	ik, err := decodeAnyB64(identityKeyB64)
	if err != nil || len(ik) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid IK: %w", err)
	}

	spk, err := decodeAnyB64(spkB64)
	if err != nil || len(spk) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid SPK: %w", err)
	}

	sig, err := decodeAnyB64(sigB64)
	if err != nil || len(sig) != ed25519.SignatureSize {
		return fmt.Errorf("invalid signature: %w", err)
	}

	if !ed25519.Verify(ed25519.PublicKey(ik), spk, sig) {
		return fmt.Errorf("signature verification failed")
	}
	return nil
}

// StoreBundle stores a user's key bundle after verifying the SPK signature
func (s *BundleService) StoreBundle(ctx context.Context, req x3dh.RegisterBundleRequest) error {
	// Step 1: verify SPK signature
	if err := verifySPKSignatureBase64(req.IdentityKey, req.SignedPreKey, req.SPKSignature); err != nil {
		s.logAudit(ctx, req.UserID, "REGISTER_BUNDLE", "failure", "Invalid SPK signature", bson.M{
			"method": "POST",
			"route":  "/x3dh/register-bundle",
			"step":   "signature_verification",
		})
		return fmt.Errorf("invalid SPK signature: %w", err)
	}

	// Step 2: store bundle
	err := s.store.StoreBundle(ctx, req, s.crypto)

	status := "success"
	description := "Bundle stored successfully"
	if err != nil {
		status = "failure"
		description = "Failed to store bundle"
	}

	s.logAudit(ctx, req.UserID, "REGISTER_BUNDLE", status, description, bson.M{
		"method":           "POST",
		"route":            "/x3dh/register-bundle",
		"num_opks":         len(req.OneTimePreKeys),
		"has_identity_key": req.IdentityKey != "",
		"has_spk":          req.SignedPreKey != "",
	})

	return err
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

	var opkPub, opkID string
	if opk != nil {
		opkPub, opkID = opk.PublicKey, opk.KeyID // Use KeyID instead of ID
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
		OneTimePreKey: opkPub,
		OPKID:         opkID,
	}, nil
}

// RefillOPKs adds new one-time pre-keys for a user
func (s *BundleService) RefillOPKs(ctx context.Context, userID string, opks []x3dh.OneTimePreKeyUpload) error {
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

// RotateSPK rotates a user's signed pre-key
func (s *BundleService) RotateSPK(ctx context.Context, userID, newSPK, signature string, expiresAt *time.Time) error {
	// Get the user's identity key to verify the signature
	ik, err := s.store.GetIdentityKey(ctx, userID)
	if err != nil {
		s.logAudit(ctx, userID, "ROTATE_SPK", "failure", "Failed to fetch IK", bson.M{
			"method": "POST",
			"route":  "/x3dh/rotate-spk",
			"step":   "identity_key",
		})
		return fmt.Errorf("fetch IK: %w", err)
	}

	// Verify the new SPK signature
	if err := verifySPKSignatureBase64(ik.PublicKey, newSPK, signature); err != nil {
		s.logAudit(ctx, userID, "ROTATE_SPK", "failure", "Invalid SPK signature", bson.M{
			"method": "POST",
			"route":  "/x3dh/rotate-spk",
			"step":   "signature_verification",
		})
		return fmt.Errorf("invalid SPK signature: %w", err)
	}

	// Store the new SPK
	err = s.store.RotateSignedPreKey(ctx, userID, newSPK, signature, expiresAt)

	status := "success"
	description := "SPK rotated successfully"
	if err != nil {
		status = "failure"
		description = "Failed to rotate SPK"
	}

	s.logAudit(ctx, userID, "ROTATE_SPK", status, description, bson.M{
		"method":     "POST",
		"route":      "/x3dh/rotate-spk",
		"has_expiry": expiresAt != nil,
	})

	return err
}

// CountAvailableOPKs returns the count of available one-time pre-keys for a user
func (s *BundleService) CountAvailableOPKs(ctx context.Context, userID string) (int, error) {
	count, err := s.store.CountAvailableOPKs(ctx, userID)

	status := "success"
	description := "Counted available OPKs"
	if err != nil {
		status = "failure"
		description = "Failed to count OPKs"
	}

	s.logAudit(ctx, userID, "COUNT_OPKS", status, description, bson.M{
		"method": "GET",
		"route":  "/x3dh/opk-count/:user_id",
		"count":  count,
	})

	return count, err
}
